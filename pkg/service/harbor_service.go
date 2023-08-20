package service

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/member"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/user"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"jiayi.com/auth-center/pkg/client"
)

type HarborService struct {
	ctx context.Context
	log logr.Logger
}

func NewHarborService(ctx context.Context, log logr.Logger) *HarborService {
	return &HarborService{
		ctx: ctx,
		log: log,
	}
}

// CreateUser 创建用户
func (s *HarborService) CreateUser(user *user.CreateUserParams) (bool, error) {
	harborClient := client.GetHarborClient()
	createUser, err := harborClient.V2().User.CreateUser(s.ctx, user)
	if err != nil {
		s.log.V(1).Error(err, "create user error", err)
		return false, err
	}
	return createUser.IsSuccess(), nil
}

// GetUser 查询用户信息
func (s *HarborService) GetUser(userID int64) (*models.UserResp, error) {
	harborClient := client.GetHarborClient()
	getUser, err := harborClient.V2().User.GetUser(s.ctx, &user.GetUserParams{
		UserID: userID,
	})
	if err != nil {
		s.log.V(1).Error(err, "get user error", err)
		return nil, err
	}
	return getUser.Payload, nil
}

// DeleteUser 删除用户
func (s *HarborService) DeleteUser(userID int64) (bool, error) {
	harborClient := client.GetHarborClient()
	deleteUser, err := harborClient.V2().User.DeleteUser(s.ctx, &user.DeleteUserParams{
		UserID: userID,
	})
	if err != nil {
		s.log.V(1).Error(err, "delete user error", err)
		return false, err
	}
	return deleteUser.IsSuccess(), nil
}

func (s *HarborService) GetProject(ProjectNameOrID string) (*models.Project, error) {
	harborClient := client.GetHarborClient()
	getProject, err := harborClient.V2().Project.GetProject(s.ctx, &project.GetProjectParams{
		ProjectNameOrID: ProjectNameOrID,
	})
	if err != nil {
		s.log.V(1).Error(err, "get project error", err)
		return nil, err
	}
	return getProject.Payload, nil
}

// ListAllProjects 查询拥有者下所有的项目
func (s *HarborService) ListAllProjects(owner *string) ([]*models.Project, error) {
	harborClient := client.GetHarborClient()
	listProjects, err := harborClient.V2().Project.ListProjects(s.ctx, &project.ListProjectsParams{
		Owner: owner,
	})
	if err != nil {
		s.log.V(1).Error(err, "list project error", err)
		return nil, err
	}
	return listProjects.Payload, nil
}

// DeleteProjectMember 删除项目下的成员
func (s *HarborService) DeleteProjectMember(projectID string, memberID int64) error {
	harborClient := client.GetHarborClient()
	_, err := harborClient.V2().Member.DeleteProjectMember(s.ctx, &member.DeleteProjectMemberParams{
		ProjectNameOrID: projectID,
		Mid:             memberID,
		Context:         s.ctx,
		HTTPClient:      nil,
	})
	if err != nil {
		s.log.V(1).Error(err, "delete project member error", err)
		return err
	}
	return nil
}

// ListProjectMembers 查询项目下的成员
func (s *HarborService) ListProjectMembers(projectID string) ([]*models.ProjectMemberEntity, error) {
	harborClient := client.GetHarborClient()
	listProjectMembers, err := harborClient.V2().Member.ListProjectMembers(s.ctx, &member.ListProjectMembersParams{
		ProjectNameOrID: projectID,
		Context:         s.ctx,
		HTTPClient:      nil,
	})
	if err != nil {
		s.log.V(1).Error(err, "list project members error", err)
		return nil, err
	}
	return listProjectMembers.Payload, nil
}

// CreateProjectMember 添加项目成员
func (s *HarborService) CreateProjectMember(projectID string, members *models.ProjectMember) error {
	harborClient := client.GetHarborClient()
	_, err := harborClient.V2().Member.CreateProjectMember(s.ctx, &member.CreateProjectMemberParams{
		ProjectMember:   members,
		ProjectNameOrID: projectID,
	})
	if err != nil {
		s.log.V(1).Error(err, "add project member error", err)
		return err
	}
	return nil
}
