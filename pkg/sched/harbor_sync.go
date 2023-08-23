package sched

import (
	"context"
	"time"

	"github.com/gookit/goutil"
	authv1 "jiayi.com/auth-center/api/v1"
	harborClient "jiayi.com/auth-center/pkg/client"
	"jiayi.com/auth-center/pkg/config"
	"jiayi.com/auth-center/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HarborRunnable /*
type HarborRunnable struct {
	Client client.Client
}

func (r *HarborRunnable) Start(ctx context.Context) error {
	klog.Infof("HarborRunnable Start")
	r.Sync(ctx)
	for {
		select {
		case <-ctx.Done():
			klog.Infof("HarborRunnable Done")
			return nil
		case <-time.After(2 * time.Minute):
			if err := r.Sync(ctx); err != nil {
				klog.Errorf("HarborRunnable Sync error:%v", err)
			}
		}
	}
}
func (r *HarborRunnable) Sync(ctx context.Context) error {
	klog.Infof("HarborRunnable Sync")
	authList := new(authv1.AuthCenterList)
	err := r.Client.List(ctx, authList)
	if err != nil {
		klog.Errorf("HarborRunnable Sync error:%v", err)
		return err
	}

	hbCli := harborClient.GetHarborClient(ctx)
	users, err := hbCli.ListAllUsers()
	if err != nil {
		klog.Errorf("HarborRunnable Sync ListAllUsers error:%v", err)
		return err
	}
	usrIds := make([]string, 0)
	for _, user := range users {
		usrIds = append(usrIds, goutil.String(user.UserID))
	}
	//只处理进入稳态的数据 statu=success
	group := goutil.NewErrGroup(10)
	for _, item := range authList.Items {
		item := item
		if item.Status.Status != authv1.StatusTypeSuccess {
			continue
		}

		if item.Spec.Harbor.HarborUid != "" && !goutil.Contains(usrIds, item.Spec.Harbor.HarborUid) {
			item.Spec.Harbor.HarborUid = ""
			pwd, _ := util.DecryptByAes(item.Spec.Harbor.EncryptPwd)
			item.Spec.Harbor.Password = string(pwd)
			group.Go(func() error {
				return r.Client.Update(ctx, &item)
			})

		}
	}
	err = group.Wait()
	if err != nil {
		klog.Errorf("HarborRunnable Sync User error:%v", err)
		return err
	}

	// 权限同步
	err = r.Client.List(ctx, authList)
	if err != nil {
		klog.Errorf("HarborRunnable Sync List error:%v", err)
		return err
	}
	group = goutil.NewErrGroup(10)
	for _, item := range authList.Items {
		item := item
		if item.Status.Status != authv1.StatusTypeSuccess {
			continue
		}
		if item.Spec.Harbor.HarborUid != "" {
			group.Go(func() error {
				info := item.Spec.Harbor
				pwd, _ := util.DecryptByAes(info.EncryptPwd)
				usrCli := harborClient.GetHarborClientByUser(info.Name, string(pwd), ctx)
				pjs, err := usrCli.ListAllProjects(&config.AllCfg.Harbor.UserName)
				if err != nil {
					klog.Errorf("HarborRunnable Sync ListAllProjects error:%v", err)
					return err
				}
				pjRoleMap := make(map[int64]int64)
				for _, pj := range pjs {
					if pj.CurrentUserRoleIds != nil {
						pjRoleMap[int64(pj.ProjectID)] = pj.CurrentUserRoleID
					}
				}
				change := false
				newHarborItems := make([]authv1.HarborPermission, 0)
				if len(pjRoleMap) <= len(item.Spec.HarborItems) {
					// 删除或则权限被修改
					for _, harborItem := range item.Spec.HarborItems {
						role, ok := pjRoleMap[harborItem.ProjectID]
						// 项目不存在，则进行移除
						if !ok {
							change = true
							continue
						}
						// 权限被修改 则更新权限
						if role != harborItem.RoleID {
							change = true
						}
						newHarborItems = append(newHarborItems, authv1.HarborPermission{
							ProjectID: harborItem.ProjectID,
							RoleID:    role,
						})
					}
				} else {
					// 权限被添加和修改
					HarborItemMap := make(map[int64]authv1.HarborPermission)
					for _, harborItem := range item.Spec.HarborItems {
						HarborItemMap[harborItem.ProjectID] = harborItem
					}
					for project, role := range pjRoleMap {
						roleItem, ok := HarborItemMap[project]
						if !ok {
							change = true
						}
						if role != roleItem.RoleID {
							change = true
						}
						newHarborItems = append(newHarborItems, authv1.HarborPermission{
							ProjectID: project,
							RoleID:    role,
						})
					}
				}

				if change {
					item.Spec.HarborItems = newHarborItems
					return r.Client.Update(ctx, &item)
				}
				return nil
			})

		}
	}
	return group.Wait()

}
