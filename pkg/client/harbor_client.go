package client

import (
	"context"
	"errors"
	"log"

	"github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/member"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/user"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/gookit/goutil"
	"jiayi.com/auth-center/pkg/config"
)

type HarborClient struct {
	*harbor.ClientSet
	ctx context.Context
}

var pageSize = int64(100)
var page = int64(1)

var harborClient *HarborClient

func GetHarborClient(ctx context.Context) *HarborClient {
	if harborClient == nil {
		client, err := InitHarborClient()
		if err != nil {
			log.Printf("init harbor client error %v", err)
			return nil
		}
		harborClient = &HarborClient{
			ClientSet: client,
			ctx:       ctx,
		}
	}
	return harborClient
}
func GetHarborClientByUser(name, pwd string, ctx context.Context) *HarborClient {
	client, err := harbor.NewClientSet(&harbor.ClientSetConfig{
		URL:      config.AllCfg.Harbor.Uri,
		Insecure: true,
		Username: name,
		Password: pwd,
	})
	if err != nil {
		log.Printf("init harbor client error %v, Username: %s", err, name)
		return nil
	}

	return &HarborClient{
		ClientSet: client,
		ctx:       ctx,
	}
}

func InitHarborClient() (*harbor.ClientSet, error) {
	return harbor.NewClientSet(&harbor.ClientSetConfig{
		URL:      config.AllCfg.Harbor.Uri,
		Insecure: true,
		Username: config.AllCfg.Harbor.UserName,
		Password: config.AllCfg.Harbor.Password,
	})
}

func (c *HarborClient) name() {

}

// CreateUser 创建用户
func (c *HarborClient) CreateUser(UserReq *models.UserCreationReq) (bool, error) {
	createUser, err := c.V2().User.CreateUser(c.ctx, &user.CreateUserParams{
		UserReq: UserReq,
	})
	if err != nil {
		return false, err
	}
	return createUser.IsSuccess(), nil
}

// SearchUsers 根据名称搜索用户
func (c *HarborClient) SearchUsers(username string, isLike bool) ([]*models.UserSearchRespItem, error) {
	searchUsers, err := c.V2().User.SearchUsers(c.ctx, &user.SearchUsersParams{
		Page:     &page,
		PageSize: &pageSize,
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	if !isLike {
		for _, item := range searchUsers.Payload {
			if item.Username == username {
				return []*models.UserSearchRespItem{item}, nil
			}
		}
		return []*models.UserSearchRespItem{}, nil
	}
	return searchUsers.Payload, nil
}

// GetUser 查询用户信息
func (c *HarborClient) GetUser(userID int64) (*models.UserResp, error) {
	getUser, err := c.V2().User.GetUser(c.ctx, &user.GetUserParams{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return getUser.Payload, nil
}

// ListAllUsers 查询所有用户
func (c *HarborClient) ListAllUsers() ([]*models.UserResp, error) {
	listUsers, err := c.V2().User.ListUsers(c.ctx, &user.ListUsersParams{
		PageSize: &pageSize,
		Page:     &page,
	})
	if err != nil {
		return nil, err
	}
	if len(listUsers.Payload) == goutil.Int(listUsers.XTotalCount) {
		return listUsers.Payload, nil
	}
	for i := int64(2); i <= listUsers.XTotalCount/pageSize+1; i++ {
		listUsers2, err := c.V2().User.ListUsers(c.ctx, &user.ListUsersParams{
			PageSize: &pageSize,
			Page:     &i,
		})
		if err != nil {
			return nil, err
		}
		listUsers.Payload = append(listUsers.Payload, listUsers2.Payload...)
	}
	return listUsers.Payload, nil
}

// IsGetUserNotFound 判断是否是用户不存在的错误
func (c *HarborClient) IsGetUserNotFound(err error) bool {
	var u *user.GetUserNotFound
	return errors.As(err, &u)
}

// DeleteUser 删除用户
func (c *HarborClient) DeleteUser(userID int64) (bool, error) {
	deleteUser, err := c.V2().User.DeleteUser(c.ctx, &user.DeleteUserParams{
		UserID: userID,
	})
	if err != nil {
		return false, err
	}
	return deleteUser.IsSuccess(), nil
}

// UpdateUserPassword 更新用户密码
func (c *HarborClient) UpdateUserPassword(userID int64, password *models.PasswordReq) (bool, error) {
	updateUserPassword, err := c.V2().User.UpdateUserPassword(c.ctx, &user.UpdateUserPasswordParams{
		UserID:   userID,
		Password: password,
	})
	if err != nil {
		return false, err
	}
	return updateUserPassword.IsSuccess(), nil
}

func (c *HarborClient) GetProject(ProjectNameOrID string) (*models.Project, error) {
	getProject, err := c.V2().Project.GetProject(c.ctx, &project.GetProjectParams{
		ProjectNameOrID: ProjectNameOrID,
	})
	if err != nil {
		return nil, err
	}
	return getProject.Payload, nil
}

// ListAllProjects 查询拥有者下所有的项目
func (c *HarborClient) ListAllProjects(owner *string) ([]*models.Project, error) {
	listProjects, err := c.V2().Project.ListProjects(c.ctx, &project.ListProjectsParams{
		Owner:    owner,
		PageSize: &pageSize,
		Page:     &page,
	})
	if err != nil {
		return nil, err
	}
	if len(listProjects.Payload) == goutil.Int(listProjects.XTotalCount) {
		return listProjects.Payload, nil
	}
	for i := int64(2); i <= listProjects.XTotalCount/pageSize+1; i++ {
		listProjects2, err := c.V2().Project.ListProjects(c.ctx, &project.ListProjectsParams{
			Owner:    owner,
			PageSize: &pageSize,
			Page:     &i,
		})
		if err != nil {
			return nil, err
		}
		listProjects.Payload = append(listProjects.Payload, listProjects2.Payload...)
	}
	return listProjects.Payload, nil
}

// DeleteProjectMember 删除项目下的成员
func (c *HarborClient) DeleteProjectMember(projectID string, memberID int64) error {
	_, err := c.V2().Member.DeleteProjectMember(c.ctx, &member.DeleteProjectMemberParams{
		ProjectNameOrID: projectID,
		Mid:             memberID,
		Context:         c.ctx,
		HTTPClient:      nil,
	})
	if err != nil {
		return err
	}
	return nil
}

// ListProjectMembers 查询项目下的成员
func (c *HarborClient) ListProjectMembers(projectID string, userName *string, isLike bool) ([]*models.ProjectMemberEntity, error) {
	listProjectMembers, err := c.V2().Member.ListProjectMembers(c.ctx, &member.ListProjectMembersParams{
		Entityname:      userName,
		Page:            &page,
		PageSize:        &pageSize,
		ProjectNameOrID: projectID,
		Context:         c.ctx,
	})
	if err != nil {
		return nil, err
	}
	if !isLike && userName != nil {
		for _, entity := range listProjectMembers.Payload {
			if entity.EntityName == *userName {
				return []*models.ProjectMemberEntity{entity}, nil
			}
		}
		return []*models.ProjectMemberEntity{}, nil
	}
	return listProjectMembers.Payload, nil
}

// CreateProjectMember 添加项目成员
func (c *HarborClient) CreateProjectMember(projectID string, members *models.ProjectMember) error {
	_, err := c.V2().Member.CreateProjectMember(c.ctx, &member.CreateProjectMemberParams{
		ProjectMember:   members,
		ProjectNameOrID: projectID,
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateProjectMember 修改项目成员
func (c *HarborClient) UpdateProjectMember(projectID string, mid, roleID int64) error {
	_, err := c.V2().Member.UpdateProjectMember(c.ctx, &member.UpdateProjectMemberParams{
		Mid:             mid,
		ProjectNameOrID: projectID,
		Role:            &models.RoleRequest{RoleID: roleID},
	})
	if err != nil {
		return err
	}
	return nil
}
