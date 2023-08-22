package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/user"
	"jiayi.com/auth-center/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var harborclient *HarborClient

func TestHarborClient_GetUser(t *testing.T) {
	user1, err := harborclient.GetUser(17)
	fmt.Println(reflect.TypeOf(err))
	var u *user.GetUserNotFound
	println(errors.As(err, &u))
	if client.IgnoreNotFound(err) != nil {
		fmt.Println(err.Error())
		t.Log(err)
	}
	t.Log(user1)
}

func init() {
	config.AllCfg.Harbor.Uri = "https://172.31.1.58:10043"
	harborclient = GetHarborClient(context.Background())
}

func TestHarborClient_ListAllProjects(t *testing.T) {
	got, err := harborclient.ListAllProjects(nil)
	if err != nil {
		t.Errorf("ListAllProjects() error = %v", err)
		return
	}
	for _, v := range got {
		t.Log(v)
	}
}
