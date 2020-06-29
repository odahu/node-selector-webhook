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
	"fmt"
	"github.com/odahu/node-selector-webhook/pkg/config"
	nswebhook "github.com/odahu/node-selector-webhook/pkg/webhook"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	k8s_config "sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)


var (
	log = logf.Log.WithName("node-selector-controller")
	WebhookV1Path = "/mutate-v1-pod"
)



var mainCmd = &cobra.Command{
	Use:   "node-selector-webhook",
	Short: "Node selector webhook server",
	Run: func(cmd *cobra.Command, args []string) {
		err := runManager()
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {

	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	mainCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file")
	if err := mainCmd.MarkPersistentFlagRequired("config"); err != nil {
		log.Info(fmt.Sprintf("%v", err))
	}
}

func runManager() error {

	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Error(err, "Unable load config")
		return err
	}

	// Setup a Manager
	log.Info("Setting up manager")
	mgr, err := manager.New(k8s_config.GetConfigOrDie(), manager.Options{
		Namespace: "model_deployment",
		Port: appConfig.Port,
	})
	if err != nil {
		log.Error(err, "Unable to create manager")
		return err
	}

	// Setup webhooks
	log.Info("Setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir  = appConfig.CrtDirName
	hookServer.CertName = appConfig.CrtName
	hookServer.KeyName =  appConfig.KeyName

	log.Info("Registering webhooks to the webhook server")
	hookServer.Register(WebhookV1Path, &webhook.Admission{Handler: &nswebhook.NodeSelectorMutator{
		NodeSelector: appConfig.NodeSelector,
		Toleration:   appConfig.Tolerations,
	}})

	log.Info("Starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Error in manager control loop")
		return err
	}
	return nil
}


func main() {

	if err := mainCmd.Execute(); err != nil {
		os.Exit(1)
	}




}
