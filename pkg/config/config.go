package config

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	Client                 Auth0Client   `yaml:"client"`
	EnableCallbackDeletion bool          `yaml:"enableCallbackDeletion,omitempty"`
	ResyncPeriod           int           `yaml:"resyncPeriod,omitempty"`
	CreationDelay          time.Duration `yaml:"creationDelay,omitempty"`
}

type Auth0Client struct {
	AppID        string `yaml:"appID,omitempty"`
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
	Domain       string `yaml:"domain"`
	APIURL       string `yaml:"apiURL"`
}

func ReadConfig(filePath string) Config {
	var config Config
	// Read YML
	log.Println("Reading YAML Configuration", filePath)
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(err)
	}

	// Unmarshall
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		log.Panic(err)
	}

	return config
}

func GetControllerConfig() Config {
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if len(configFilePath) == 0 {
		configFilePath = "./Configs/config.yaml"
	}

	config := ReadConfig(configFilePath)
	return config
}
