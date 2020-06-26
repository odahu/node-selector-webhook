package main

import v1 "k8s.io/api/core/v1"

type WebhookConfig struct {
	// Kubernetes node selector for model deployments
	NodeSelector map[string]string `json:"nodeSelector"`
	// Kubernetes tolerations for model deployments
	Toleration                *v1.Toleration                `json:"toleration,omitempty"`
	// Directory where certificate for webhook server TLS stored
	CrtDirName string `json:"crtDirName"`
	// Certificate name
	CrtName string `json:"crtName"`
	// Key name
	KeyName string `json:"keyName"`
}
