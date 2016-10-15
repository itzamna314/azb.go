package lib

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

var (
	ErrEnvironmentNotFound = errors.New("undefined environment")
)

type AzbConfig struct {
	Name                  string
	AccessKey             string
	ManagementCertificate []byte
}

func GetConfig(configFile, environment string) (*AzbConfig, error) {

	type envInfo struct {
		Name                      string `toml:"storage_account_name"`
		AccessKey                 string `toml:"storage_account_access_key"`
		ManagementCertificatePath string `toml:"management_certificate"`
	}

	var config map[string]envInfo
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, err
	}

	env, ok := config[environment]
	if !ok {
		return nil, ErrEnvironmentNotFound
	}

	if env.Name == "" || env.AccessKey == "" {
		return nil, fmt.Errorf("Missing storage_account_name and/or storage_account_access_key for environment %s in file %s", environment, configFile)
	}

	if env.ManagementCertificatePath != "" {
		buf, err := ioutil.ReadFile(env.ManagementCertificatePath)
		if err != nil {
			return nil, err
		}
		return &AzbConfig{env.Name, env.AccessKey, buf}, nil
	}

	return &AzbConfig{env.Name, env.AccessKey, nil}, nil
}
