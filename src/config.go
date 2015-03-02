package azb

import (
	"errors"
	"github.com/BurntSushi/toml"
	"io/ioutil"
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

	if env.ManagementCertificatePath != "" {
		buf, err := ioutil.ReadFile(env.ManagementCertificatePath)
		if err != nil {
			return nil, err
		}
		return &AzbConfig{env.Name, env.AccessKey, buf}, nil
	}

	return &AzbConfig{env.Name, env.AccessKey, nil}, nil
}
