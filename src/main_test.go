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
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	stdlog "log"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

const (
	nsEnabled = "model-deployment-labeled"
	ns        = "model-deployment"
)

var (
	namespacedScopeV1 = admissionv1.NamespacedScope
	failedTypeV1      = admissionv1.Fail
	equivalentTypeV1  = admissionv1.Equivalent
	noSideEffectsV1   = admissionv1.SideEffectClassNone

	kubeConfig   *rest.Config
	mgr          manager.Manager

	appConfig Config
)

// Create namespaces one of which is labeled by `ActivationLabel` and therefore is activated
// to catch API requests for this webhook. Another one is not activated and therefore pod that
// are created in this one are skipped by the webhook servere
func setupNamespaces() error {

	client := mgr.GetClient()

	if err := client.Create(context.TODO(), &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   nsEnabled,
			Labels: map[string]string{
				ActivationLabel: "enabled",
			},
		},
		Spec:       corev1.NamespaceSpec{},
		Status:     corev1.NamespaceStatus{},
	}); err != nil {
		return err
	}

	if err := client.Create(context.TODO(), &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns,
		},
		Spec:       corev1.NamespaceSpec{},
		Status:     corev1.NamespaceStatus{},
	}); err != nil {
		return err
	}
	return nil

}

func setupTestEnv() *envtest.Environment {
	return &envtest.Environment{
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			MutatingWebhooks: []runtime.Object{
				&admissionv1.MutatingWebhookConfiguration{
					TypeMeta: metav1.TypeMeta{
						Kind:       "MutatingWebhookConfiguration",
						APIVersion: "admissionregistration.k8s.io/v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-selector-mutator",
					},
					Webhooks: []admissionv1.MutatingWebhook{
						{
							Name: "node-selector-mutator.odahu.org",
							ClientConfig: admissionv1.WebhookClientConfig{
								Service: &admissionv1.ServiceReference{Path: &WebhookV1Path},
							},
							Rules: []admissionv1.RuleWithOperations{
								{
									Operations: []admissionv1.OperationType{"CREATE", "UPDATE"},
									Rule: admissionv1.Rule{
										APIGroups:   []string{""},
										APIVersions: []string{"v1"},
										Resources:   []string{"pods"},
										Scope:       &namespacedScopeV1,
									},
								},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      ActivationLabel,
										Operator: metav1.LabelSelectorOpExists,
										Values:   []string{},
									},
								},
							},
							FailurePolicy: &failedTypeV1,
							MatchPolicy:   &equivalentTypeV1,
							SideEffects:   &noSideEffectsV1,
						},
					},
				},
			},
		},
	}
}

func setupConfig() Config {
	return Config {
		NodeSelector: map[string]string{"mode": "odahu-flow-deployment"},
		Tolerations:   []corev1.Toleration{
			{
				Key:      "dedicated",
				Operator: corev1.TolerationOpEqual,
				Value:    "deployment",
				Effect:   corev1.TaintEffectNoSchedule,
			},
		},
		CrtDirName:   "",
		CrtName:      "",
		KeyName:      "",
	}
}

func TestMain(m *testing.M) {

	var err error

	appConfig = setupConfig()

	env := setupTestEnv()

	if kubeConfig, err = env.Start(); err != nil {
		stdlog.Fatal(err)
	}

	if mgr, err = manager.New(kubeConfig, manager.Options {
		MetricsBindAddress: "0",
		Port:               env.WebhookInstallOptions.LocalServingPort,
		Host:               env.WebhookInstallOptions.LocalServingHost,
		CertDir:            env.WebhookInstallOptions.LocalServingCertDir,
	}); err != nil {
		stdlog.Fatal(err)
	}

	hookServer := mgr.GetWebhookServer()
	hookServer.Register(V1Path, &webhook.Admission{Handler: &NodeSelectorMutator{
		NodeSelector: appConfig.NodeSelector,
		Toleration:   appConfig.Tolerations,
	}})

	// Start Manager with Webhook server
	stop := make(chan struct{})
	go func() {
		if err := mgr.Start(stop); err != nil {
			stdlog.Fatal(err)
		}
	}()

	code := m.Run()

	// Stop Manager with Webhook server
	close(stop)

	// Stop Test Kube Env
	if err := env.Stop(); err != nil {
		stdlog.Fatal(err)
	}

	os.Exit(code)

}

func TestNodeSelectorMutator(t *testing.T) {

	g := NewGomegaWithT(t)

	client := mgr.GetClient()

	err := setupNamespaces()
	g.Expect(err).NotTo(HaveOccurred())

	podInEnabledNs := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: nsEnabled, Name: "model-pod"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx",
					Name:  "nginx",
				},
			},
		},
	}
	podInDisabledNs := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "model-pod"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx",
					Name:  "nginx",
				},
			},
		},
	}

	err = client.Create(context.Background(), podInEnabledNs)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(podInEnabledNs.Spec.NodeSelector).Should(Equal(appConfig.NodeSelector))
	for _, tol := range appConfig.Tolerations {
		g.Expect(podInEnabledNs.Spec.Tolerations).Should(ContainElement(tol))
	}

	err = client.Create(context.Background(), podInDisabledNs)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(podInDisabledNs.Spec.NodeSelector).Should(BeEmpty())
	for _, tol := range appConfig.Tolerations {
		g.Expect(podInDisabledNs.Spec.Tolerations).Should(Not(ContainElement(tol)))
	}
}
