package config

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var valuesFile string
var configFile string
var showConfig bool

var app = newApp()
var harbor = newHarbor()

// AllCfg include all configurations.
var AllCfg = &allCfg{App: &app, Harbor: &harbor}

type allCfg struct {
	App    *AppConf    `yaml:"app"`
	Harbor *HarborConf `yaml:"harbor"`
}

func InitFlag() {
	flag.StringVar(&valuesFile, "values", "", "path of the values file")
	flag.StringVar(&configFile, "conf", "resources/app-dev.yaml", "path of the config file")
	flag.BoolVar(&showConfig, "showConfig", false, "log config")
}

// InitConfig init configs.
func InitConfig() {
	if valuesFile != "" {
		log.Printf("using values file %v", valuesFile)

		err := loadValues(valuesFile)
		if err != nil {
			log.Printf("load values error %v", err)
		}
	} else {
		log.Printf("using conf file %v", configFile)

		err := LoadConfig(configFile)
		if err != nil {
			log.Printf("load allCfg error %v", err)
		}
	}

	if showConfig {
		configData, err := yaml.Marshal(AllCfg)
		if err != nil {
			log.Printf("log allCfg error:%v", err)
		}

		log.Println(">>>>>>>>>>allCfg:")
		log.Println(string(configData))
	}
}

// LoadConfig load configs.
func LoadConfig(file string) error {
	yamlFile, err := loadFile(file)
	if err != nil {
		return err
	}

	err = resetConfig(yamlFile)
	if err != nil {
		return err
	}

	return nil
}

type values struct {
	AllConf string `yaml:"ze.conf"`
}

func loadValues(file string) error {
	yamlFile, err := loadFile(file)
	if err != nil {
		return err
	}

	va := new(values)
	err = yaml.Unmarshal(yamlFile, va)
	if err != nil {
		return err
	}

	err = resetConfig([]byte(va.AllConf))
	if err != nil {
		return err
	}

	return nil
}

func resetConfig(yamlData []byte) error {
	err := yaml.Unmarshal(yamlData, AllCfg)
	if err != nil {
		return err
	}

	return nil
}

func loadFile(file string) ([]byte, error) {
	yamlAbsPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	log.Printf("yamlAbsPath:%v", yamlAbsPath)
	yamlFile, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		return nil, err
	}

	return yamlFile, nil
}
