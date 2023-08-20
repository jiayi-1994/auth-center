package config

type HarborConf struct {
	Enable   bool   `yaml:"enable"`   // 是否启用
	Uri      string `yaml:"uri"`      // 地址
	UserName string `yaml:"userName"` // 用户名
	Password string `yaml:"password"` // 密码
	Owner    string `yaml:"owner"`
}

func newHarbor() HarborConf {
	return HarborConf{
		Enable:   false,
		Uri:      "",
		UserName: "admin",
		Password: "admin@123",
		Owner:    "admin",
	}
}
