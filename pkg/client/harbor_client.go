package client

import (
	"log"

	"github.com/goharbor/go-client/pkg/harbor"
	"jiayi.com/auth-center/pkg/config"
)

var harborClient *harbor.ClientSet

func GetHarborClient() *harbor.ClientSet {
	if harborClient == nil {
		client, err := InitHarborClient()
		if err != nil {
			log.Printf("init harbor client error %v", err)
			return nil
		}
		harborClient = client
	}
	return harborClient
}

func InitHarborClient() (*harbor.ClientSet, error) {
	return harbor.NewClientSet(&harbor.ClientSetConfig{
		URL:      config.AllCfg.Harbor.Uri,
		Insecure: true,
		Username: config.AllCfg.Harbor.UserName,
		Password: config.AllCfg.Harbor.Password,
	})
}
