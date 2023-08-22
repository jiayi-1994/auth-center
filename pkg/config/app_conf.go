package config

type AppConf struct {
	ErrOfRollback           bool   `yaml:"errOfRollback"`
	ContainerErrOfRollback  bool   `yaml:"containerErrOfRollback"`
	HarborErrOfRollback     bool   `yaml:"harborErrOfRollback"`
	AuthCenterConditionSize int    `yaml:"authCenterConditionSize"`
	AesKey                  string `yaml:"aesKey"`                  //16,24,32位字符串的话，分别对应AES-128，AES-192，AES-256 加密方法
	LeaderElectionNamespace string `yaml:"leaderElectionNamespace"` //leader选举的namespace
}

func newApp() AppConf {
	return AppConf{
		ErrOfRollback:           false,
		ContainerErrOfRollback:  false,
		HarborErrOfRollback:     false,
		AuthCenterConditionSize: 6,
		AesKey:                  "jiayi.AuthCenter.202308@",
		LeaderElectionNamespace: "auth-center-system",
	}
}

func GetContainerErrOfRollback() bool {
	return AllCfg.App.ErrOfRollback && AllCfg.App.ContainerErrOfRollback
}

func GetHarborErrOfRollback() bool {
	return AllCfg.App.ErrOfRollback && AllCfg.App.HarborErrOfRollback
}
