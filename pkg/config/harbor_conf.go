package config

type HarborConf struct {
	Enable   bool   `yaml:"enable"`   // 是否启用
	Uri      string `yaml:"uri"`      // 地址
	UserName string `yaml:"userName"` // 用户名
	Password string `yaml:"password"` // 密码
}

func newHarbor() HarborConf {
	return HarborConf{
		Enable:   false,
		Uri:      "",
		UserName: "xjy",
		Password: "Harbor12345",
	}
}
