package backend

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type layout struct {
	ColumnWidth int `yaml:"column_width"`
	ItemsMargin int `yaml:"items_margin"`
	BoxHeigh    int `yaml:"items_box_heigh"`
}

type ConfFeed struct {
	Url  string   `yaml:"url"`
	Tags []string `yaml:"tags"`
}

type Config struct {
	Browser        string        `yaml:"browser"`
	Log_file       string        `yaml:"log_file"`
	UpdateStartup  bool          `yaml:"update_at_startup"`
	UpdateInterval time.Duration `yaml:"update_interval"`
	Feeds          []ConfFeed    `yaml:"feeds"`
	Layout         layout        `yaml:"layout"`
}

var configFile string

var conf Config

// confDir must be already expanded
func LoadConfig(cfgDir string) *Config {
	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		err := os.MkdirAll(cfgDir, os.ModePerm)
		if err != nil {
			log.Fatalln("Can not create config directory: ", err)
		}
	}

	configFile = cfgDir + "atrss.yml"
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			var conf Config
			WriteConfig(conf)
		} else {
			log.Fatalln("Unknown error when reading config: ", err)
		}
	}

	err = yaml.UnmarshalStrict([]byte(data), &conf)
	if err != nil {
		log.Fatalln("Can not get config: ", err)
	}

	return &conf
}

func WriteConfig(c Config) {
	if configFile == "" {
		log.Println("Config must be loaded before any write")
		return
	}

	d, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalln("Can not create default config: ", err)
	}

	err = ioutil.WriteFile(configFile, d, os.ModePerm)
	if err != nil {
		log.Fatalln("Can not create default config file: ", err)
	}
	conf = c
}

func GetConfig() *Config {
	return &conf
}
