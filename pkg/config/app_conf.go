package config

type AppConf struct {
	ErrOfRollback           bool `yaml:"errOfRollback"`
	ContainerErrOfRollback  bool `yaml:"containerErrOfRollback"`
	HarborErrOfRollback     bool `yaml:"harborErrOfRollback"`
	AuthCenterConditionSize int  `yaml:"authCenterConditionSize"`
}

func newApp() AppConf {
	return AppConf{
		ErrOfRollback:           false,
		ContainerErrOfRollback:  false,
		HarborErrOfRollback:     false,
		AuthCenterConditionSize: 6,
	}
}

func GetContainerErrOfRollback() bool {
	return AllCfg.App.ErrOfRollback && AllCfg.App.ContainerErrOfRollback
}

func GetHarborErrOfRollback() bool {
	return AllCfg.App.ErrOfRollback && AllCfg.App.HarborErrOfRollback
}
