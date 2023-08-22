package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/gookit/goutil"
	authv1 "jiayi.com/auth-center/api/v1"
	harborClient "jiayi.com/auth-center/pkg/client"
	"jiayi.com/auth-center/pkg/config"
	"jiayi.com/auth-center/pkg/util"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type HarborService struct {
	ctx          context.Context
	k8sClient    *k8sClient.Client
	harborClient *harborClient.HarborClient
	log          logr.Logger
}

func NewHarborService(ctx context.Context, log logr.Logger, k8sClient *k8sClient.Client) *HarborService {
	return &HarborService{
		ctx:          ctx,
		k8sClient:    k8sClient,
		harborClient: harborClient.GetHarborClient(ctx),
		log:          log,
	}
}

func (s *HarborService) InitHarborUser(auth *authv1.AuthCenter) (bool, error) { // bool isReturn
	// 判断用户，看是修改密码还是创建用户
	hb := auth.Spec.Harbor
	isReturn := false
	if hb.Name != "" && hb.Password != "" {
		isReturn = true
		if hb.HarborUid != "" {
			// 修改密码
			password, err := s.harborClient.UpdateUserPassword(goutil.Int64(hb.HarborUid), &models.PasswordReq{
				NewPassword: hb.Password,
				OldPassword: hb.Password,
			})
			if err != nil || !password {
				s.log.Error(err, "HarborService update password failed", "name", hb.Name, "password", hb.Password)
				return isReturn, errors.New("HarborService update password failed")
			}

		} else {
			users, err := s.harborClient.SearchUsers(hb.Name, true)
			if err != nil {
				s.log.Error(err, "HarborService search users failed", "name", hb.Name, "password", hb.Password)
				return isReturn, errors.New("HarborService search users failed")
			}
			if len(users) == 0 {
				usr := &models.UserCreationReq{
					Username: hb.Name,
					Password: hb.Password,
					Email:    fmt.Sprintf(config.HarborEmailFormat, hb.Name),
					Realname: hb.Name,
				}
				createUser, err := s.harborClient.CreateUser(usr)
				if err != nil || !createUser {
					s.log.Error(err, "HarborService create user failed", "name", hb.Name, "password", hb.Password)
					return isReturn, errors.New("HarborService create user failed")
				}

				users, err = s.harborClient.SearchUsers(hb.Name, true)
				if err != nil {
					s.log.Error(err, "HarborService search users failed", "name", hb.Name, "password", hb.Password)
					return isReturn, errors.New("HarborService search users failed")
				}
			}
			// 创建用户

			auth.Spec.Harbor.HarborUid = goutil.String(users[0].UserID)
		}
		// 创建和更新密码后需要 清除原本密码，并转成加密密码
		auth.Spec.Harbor.EncryptPwd, _ = util.EncryptByAes([]byte(hb.Password))
		auth.Spec.Harbor.Password = ""
	}
	return isReturn, nil
}

func (s *HarborService) ApplyHarborAuth(auth *authv1.AuthCenter) ([]authv1.HarborPermissionStatus, error) {
	harborPermissionStatus := make([]authv1.HarborPermissionStatus, 0)
	rollbackByDelete, err := s.DeleteHarborAuth(auth)
	if err != nil {
		if config.GetHarborErrOfRollback() {
			s.RollbackHarborByDelete(auth, rollbackByDelete)
		}
		s.log.Error(err, "HarborService.ApplyHarborAuth  DeleteHarborAuth error")
		return harborPermissionStatus, err
	}

	rollbackByApply, result, err := s.applyHarborAuth(auth)
	if err != nil {
		if config.GetHarborErrOfRollback() {
			s.RollbackHarborByApply(auth, rollbackByApply)
		}
		s.log.Error(err, "HarborService.ApplyHarborAuth  applyHarborAuth error")
		return harborPermissionStatus, err
	}
	return result, nil
}

func (s *HarborService) DeleteHarborAuth(auth *authv1.AuthCenter) ([]authv1.HarborPermission, error) {
	rollbackItems := make([]authv1.HarborPermission, 0)
	userName := auth.Spec.Harbor.Name
	pwd, _ := util.DecryptByAes(auth.Spec.Harbor.EncryptPwd)
	userClient := harborClient.GetHarborClientByUser(userName, string(pwd), s.ctx)
	if userClient == nil {
		return rollbackItems, errors.New("HarborService.DeleteHarborAuth  GetHarborClientByUser error")
	}
	pjAuthItemStr := make([]string, 0)
	for _, item := range auth.Spec.HarborItems {
		pjAuthItemStr = append(pjAuthItemStr, fmt.Sprintf("%d/%d", item.ProjectID, item.RoleID))
	}

	owner := config.AllCfg.Harbor.UserName
	projects, err := userClient.ListAllProjects(&owner)
	if err != nil {
		s.log.Error(err, "HarborService.DeleteHarborAuth  ListAllProjects error")
		return rollbackItems, err
	}
	for _, project := range projects {
		members, err := s.harborClient.ListProjectMembers(goutil.String(project.ProjectID), &userName, false)
		if err != nil {
			s.log.Error(err, "HarborService.DeleteHarborAuth  ListProjectMembers error")
			return rollbackItems, err
		}
		if len(project.CurrentUserRoleIds) == 0 || len(members) == 0 {
			continue
		}
		if goutil.Contains(pjAuthItemStr, fmt.Sprintf("%d/%d", project.ProjectID, project.CurrentUserRoleID)) {
			continue
		}
		err = s.harborClient.DeleteProjectMember(goutil.String(project.ProjectID), members[0].ID)
		if err != nil {
			s.log.Error(err, "HarborService.DeleteHarborAuth  DeleteProjectMember error")
			return rollbackItems, err
		}
		rollbackItems = append(rollbackItems, authv1.HarborPermission{
			ProjectID: goutil.Int64(project.ProjectID),
			RoleID:    project.CurrentUserRoleID,
		})
	}
	return rollbackItems, nil
}

func (s *HarborService) RollbackHarborByDelete(auth *authv1.AuthCenter, rollbackItems []authv1.HarborPermission) error {
	for _, item := range rollbackItems {
		err := s.harborClient.CreateProjectMember(goutil.String(item.ProjectID), &models.ProjectMember{
			RoleID: item.RoleID,
			MemberUser: &models.UserEntity{
				UserID:   goutil.Int64(auth.Spec.Harbor.HarborUid),
				Username: auth.Spec.Harbor.Name,
			},
		})
		if err != nil {
			s.log.Error(err, "HarborService.RollbackHarborByDelete  AddProjectMember error")
			return err
		}
	}
	return nil
}

func (s *HarborService) applyHarborAuth(auth *authv1.AuthCenter) ([]*models.ProjectMemberEntity, []authv1.HarborPermissionStatus, error) {
	result := make([]authv1.HarborPermissionStatus, 0)
	rollbackItem := make([]*models.ProjectMemberEntity, 0)
	for index, item := range auth.Spec.HarborItems {
		result = append(result, authv1.HarborPermissionStatus{
			HarborPermission: item,
		})
		project, err := s.harborClient.GetProject(goutil.String(item.ProjectID))
		if err != nil {
			s.log.Info("HarborService.applyHarborAuth  GetProject error", "projectID", item.ProjectID)
			continue
		}
		result[index].ProjectName = project.Name
		members, err := s.harborClient.ListProjectMembers(goutil.String(item.ProjectID), &auth.Spec.Harbor.Name, false)
		if err != nil {
			s.log.Error(err, "HarborService.applyHarborAuth  ListProjectMembers error")
			return rollbackItem, result, err
		}
		if len(members) == 0 {
			mb := &models.ProjectMember{
				MemberUser: &models.UserEntity{
					UserID:   goutil.Int64(auth.Spec.Harbor.HarborUid),
					Username: auth.Spec.Harbor.Name,
				},
				RoleID: item.RoleID,
			}
			err := s.harborClient.CreateProjectMember(goutil.String(item.ProjectID), mb)
			if err != nil {
				s.log.Error(err, "HarborService.applyHarborAuth  AddProjectMember error")
				return nil, nil, err
			}
			result[index].Status = true

			rollbackItem = append(rollbackItem, &models.ProjectMemberEntity{
				EntityID:   mb.MemberUser.UserID,
				EntityName: mb.MemberUser.Username,
				ProjectID:  item.ProjectID,
				RoleID:     item.RoleID,
			})
		} else {
			if members[0].RoleID == item.RoleID {
				result[index].Status = true
				continue
			} else {
				err = s.harborClient.UpdateProjectMember(goutil.String(item.ProjectID), members[0].ID, item.RoleID)
				if err != nil {
					s.log.Error(err, "HarborService.applyHarborAuth  UpdateProjectMember error")
					return rollbackItem, result, err
				}
				result[index].Status = true
				rollbackItem = append(rollbackItem, members[0])
			}
		}
	}
	return rollbackItem, result, nil
}

func (s *HarborService) RollbackHarborByApply(auth *authv1.AuthCenter, items []*models.ProjectMemberEntity) {
	if items == nil {
		return
	}
	for _, item := range items {
		if item.ID == 0 {
			members, err := s.harborClient.ListProjectMembers(goutil.String(item.ProjectID), &auth.Spec.Harbor.Name, false)
			if err != nil || len(members) == 0 {
				s.log.Info("HarborService.RollbackHarborByApply  ListProjectMembers error", "err", err)
				continue
			}
			if err := s.harborClient.DeleteProjectMember(goutil.String(item.ProjectID), members[0].ID); err != nil {
				s.log.Info("HarborService.RollbackHarborByApply  DeleteProjectMember error", "err", err)
				continue
			}
		} else {
			if err := s.harborClient.UpdateProjectMember(goutil.String(item.ProjectID), item.ID, item.RoleID); err != nil {
				s.log.Info("HarborService.RollbackHarborByApply  UpdateProjectMember error", "err", err)
				continue
			}
		}
	}
	return
}

// DeleteUser 删除用户附带会自动移除所有harbor权限
func (s *HarborService) DeleteUser(auth *authv1.AuthCenter) error {
	if auth.Spec.Harbor.HarborUid == "" {
		return nil
	}
	if _, err := s.harborClient.GetUser(goutil.Int64(auth.Spec.Harbor.HarborUid)); err != nil {
		if s.harborClient.IsGetUserNotFound(err) {
			return nil
		}
		return err
	}

	success, err := s.harborClient.DeleteUser(goutil.Int64(auth.Spec.Harbor.HarborUid))
	if err != nil {
		s.log.Error(err, "HarborService.DeleteUser  DeleteUser error")
		return err
	}
	s.log.Info("delete user Status", "user", success)
	return nil
}
