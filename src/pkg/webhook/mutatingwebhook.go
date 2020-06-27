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

package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	log = logf.Log.WithName("node-selector-controller").WithName("webhook")
)

const (
	ActivationLabel = "odahu/node-selector-webhook"
)

// NodeSelectorMutator annotates Pods
type NodeSelectorMutator struct {
	decoder *admission.Decoder
	// Kubernetes node selector for model deployments
	NodeSelector map[string]string `json:"nodeSelector"`
	// Kubernetes tolerations for model deployments
	Toleration                []corev1.Toleration                `json:"tolerations,omitempty"`
}

// NodeSelectorMutator adds an annotation to every incoming pods.
func (nsm *NodeSelectorMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := nsm.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	nsm.addNodeSelectors(pod)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// NodeSelectorMutator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (nsm *NodeSelectorMutator) InjectDecoder(d *admission.Decoder) error {
	nsm.decoder = d
	return nil
}


//Adds node selectors and tolerations from the deployment Config to knative pods
func (nsm *NodeSelectorMutator) addNodeSelectors(pod *corev1.Pod)  {
	nodeSelector := nsm.NodeSelector
	if len(nodeSelector) > 0 {
		pod.Spec.NodeSelector = nodeSelector
		log.Info("Assigning node selector to nsm pod", "nodeSelector", nodeSelector, "pod name", pod.Name)
	} else {
		log.Info("Got empty node selector from deployment Config, skipping", "pod name", pod.Name)
	}

	toleration := nsm.Toleration
	if toleration != nil {
		pod.Spec.Tolerations = append(pod.Spec.Tolerations, toleration...)
		log.Info("Assigning tolerations to nsm pod", "tolerations", toleration, "pod name", pod.Name)
	} else {
		log.Info("Got empty tolerations from deployment Config, skipping", "pod name", pod.Name)
	}
}