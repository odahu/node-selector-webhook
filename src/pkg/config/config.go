package config

import (
	"fmt"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

var (
	CfgFile string
	log     = logf.Log.WithName("node-selector-webhook").WithName("config")
)

type Config struct {
	// Kubernetes node selector for model deployments
	NodeSelector map[string]string `json:"nodeSelector"`
	// Kubernetes tolerations for model deployments
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// Directory where certificate for webhook server TLS stored
	CrtDirName string `json:"crtDirName"`
	// Certificate name
	CrtName string `json:"crtName"`
	// Key name
	KeyName string `json:"keyName"`
	// Webhook port
	Port string `json:"port"`
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigFile(CfgFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Info(fmt.Sprintf("Error during reading of the odahuflow config: %s", err.Error()))
	}

	config := &Config{
		Port: "8080",
	}

	err := viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

