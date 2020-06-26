package main_test

import (
	"context"
	. "github.com/odahu/node-selector-webhook"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	namespacedScopeV1 := admissionv1.NamespacedScope
	failedTypeV1 := admissionv1.Fail
	equivalentTypeV1 := admissionv1.Equivalent
	noSideEffectsV1 := admissionv1.SideEffectClassNone

	env := &envtest.Environment{
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			MutatingWebhooks: []runtime.Object{
				&admissionv1.MutatingWebhookConfiguration{
					TypeMeta:   metav1.TypeMeta{
						Kind:       "MutatingWebhookConfiguration",
						APIVersion: "admissionregistration.k8s.io/v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-selector-mutator",
					},
					Webhooks:   []admissionv1.MutatingWebhook{
						{
							Name:                    "node-selector-mutator.odahu.org",
							ClientConfig:            admissionv1.WebhookClientConfig{
								Service: &admissionv1.ServiceReference{Path: &WebhookV1Path},
							},
							Rules:                   []admissionv1.RuleWithOperations{
								{
									Operations: []admissionv1.OperationType{"CREATE", "UPDATE"},
									Rule:       admissionv1.Rule{
										APIGroups:   []string{"core"},
										APIVersions: []string{"v1"},
										Resources:   []string{"pods"},
										Scope:       &namespacedScopeV1,
									},
								},
							},
							FailurePolicy:           &failedTypeV1,
							MatchPolicy:             &equivalentTypeV1,
							SideEffects:             &noSideEffectsV1,
						},
					},
				},
			},
		},
	}

	if kubeConfig, err = env.Start(); err != nil {
		stdlog.Fatal(err)
	}

	mgr, err = manager.New(kubeConfig, manager.Options{
		MetricsBindAddress: "0",
		Namespace: "model-deployment",
		Port:    env.WebhookInstallOptions.LocalServingPort,
		Host:    env.WebhookInstallOptions.LocalServingHost,
		CertDir: env.WebhookInstallOptions.LocalServingCertDir,
	})
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
		NodeSelector: nodeSelector,
		Toleration:   toleration,
		CrtDirName: env.WebhookInstallOptions.LocalServingCertDir,
	})

	stop := make(chan struct{})

	go func() {
		if err := mgr.Start(stop); err != nil {
			stdlog.Fatal(err)
		}
	}()

	code := m.Run()

	stop <- struct{}{}

	if err := env.Stop(); err != nil {
		stdlog.Fatal(err)
	}
	os.Exit(code)

}

func TestNodeSelectorMutator(t *testing.T) {

	g := NewGomegaWithT(t)


	client := mgr.GetClient()

	err := client.Create(context.TODO(), &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: "model-deployment"},
		Spec:       corev1.NamespaceSpec{},
		Status:     corev1.NamespaceStatus{},
	})
	g.Expect(err).NotTo(HaveOccurred())

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "model-deployment", Name: "model-pod"},
		Spec:       corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx",
					Name: "nginx",
				},
			},
		},
		Status:     corev1.PodStatus{},
	}
	err = client.Create(context.Background(), pod)

	g.Expect(err).NotTo(HaveOccurred())

}
