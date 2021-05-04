package main

import (
	"errors"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {

	/*************************
	 *    Database info
	 *************************/
	DBDriverName string `yaml:"DB_DRIVER_NAME"`

	DBDSN string `yaml:"DB_DSN"`

	/*************************
	 *    OWID Repository
	 *************************/
	OWIDDataUrl string `yaml:"OWID_DATA_URL"`
}

// Parses YAML configuration.
func parseConfiguration(configFilePath string) (*Config, error) {
	c := new(Config)

	rawConfig, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(rawConfig, c)
	if err != nil {
		return nil, err
	}

	if len(c.DBDriverName) == 0 {
		log.Println("database driver name not specified in the configuration file, " +
			"defaulting to 'mysql'.")
		c.DBDriverName = "mysql"
	}

	if len(c.DBDSN) == 0 {
		return nil, errors.New("database dsn must not be null in the configuration file")
	}

	if len(c.OWIDDataUrl) == 0 {
		return nil, errors.New("the owid repository must not be null in the configuration file")
	}

	return c, nil
}
