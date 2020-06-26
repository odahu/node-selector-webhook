/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = logf.Log.WithName("node-selector-controller")

func AttachWebhookServer(mgr manager.Manager, cfg WebhookConfig) (*webhook.Server, error)  {


	// Setup webhooks
	log.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir = cfg.CrtDirName
	hookServer.CertName = cfg.CrtName
	hookServer.KeyName = cfg.KeyName

	log.Info("registering webhooks to the webhook server")
	hookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: &NodeSelectorMutator{}})

	log.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		return nil, err
	}

	return hookServer, nil

}

func main() {

	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	// Setup a Manager
	log.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	_, err = AttachWebhookServer(mgr, WebhookConfig{})
	if err != nil {
		panic("Unable to attach webhook server to manager")
	}
}
