package main_test

import (
	"context"
	. "github.com/odahu/node-selector-webhook"
	corev1 "k8s.io/api/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	stdlog "log"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

var (
	kubeConfig *rest.Config
	mgr manager.Manager
)

func TestMain(m *testing.M) {
	var err error

	t := &envtest.Environment{
	}

	if kubeConfig, err = t.Start(); err != nil {
		stdlog.Fatal(err)
	}

	mgr, err = manager.New(kubeConfig, manager.Options{MetricsBindAddress: "0"})
	if err != nil {
		stdlog.Fatal(err)
	}

	intVal := 42
	int64Val := int64(intVal)
	nodeSelector := map[string]string{"label": "labelValue"}
	toleration := &corev1.Toleration{
		Key:               "key",
		Operator:          corev1.TolerationOpExists,
		Value:             "value",
		Effect:            corev1.TaintEffectNoSchedule,
		TolerationSeconds: &int64Val}

	AttachWebhookServer(mgr, WebhookConfig{
		Namespace:    "model-deployment",
		NodeSelector: nodeSelector,
		Toleration:   toleration,
	})

	code := m.Run()

	if err := t.Stop(); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func TestNodeSelectorMutator(t *testing.T) {

	g := NewGomegaWithT(t)


	client := mgr.GetClient()

	pod := &corev1.Pod{}
	err := client.Create(context.Background(), pod)

	g.Expect(err).NotTo(HaveOccurred())

}
