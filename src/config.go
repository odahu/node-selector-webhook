package main

import v1 "k8s.io/api/core/v1"

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
}
