package backend

import (
	"io/ioutil"
	"log"
	"os"

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
	Browser       string     `yaml:"browser"`
	Log_file      string     `yaml:"log_file"`
	UpdateStartup bool       `yaml:"update_at_startup"`
	Feeds         []ConfFeed `yaml:"feeds"`
	Layout        layout     `yaml:"layout"`
}

// confDir should be already expanded
func LoadConfig(cfgDir string) Config {
	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		err := os.MkdirAll(cfgDir, os.ModePerm)
		if err != nil {
			log.Fatalln("Can not create config directory: ", err)
		}
	}

	cfgFile := cfgDir + "atrss.yml"
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			var conf Config
			d, err := yaml.Marshal(&conf)
			if err != nil {
				log.Fatalln("Can not create default config: ", err)
			}

			err = ioutil.WriteFile(cfgFile, d, os.ModePerm)
			if err != nil {
				log.Fatalln("Can not create default config file: ", err)
			}

		} else {
			log.Fatalln("Unknown error when reading config: ", err)
		}
	}

	var conf Config
	err = yaml.UnmarshalStrict([]byte(data), &conf)
	if err != nil {
		log.Fatalln("Can not get config: ", err)
	}

	return conf

}
