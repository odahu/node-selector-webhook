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
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io

// NodeSelectorMutator annotates Pods
type NodeSelectorMutator struct {
	decoder *admission.Decoder
	config *WebhookConfig
}

// NodeSelectorMutator adds an annotation to every incoming pods.
func (nsm *NodeSelectorMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := nsm.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Namespace != "model-deployment" {
		return admission.Allowed("Not observed namespace")
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


//Adds node selectors and tolerations from the deployment config to knative pods
func (nsm *NodeSelectorMutator) addNodeSelectors(pod *corev1.Pod)  {
	nodeSelector := nsm.config.NodeSelector
	if len(nodeSelector) > 0 {
		pod.Spec.NodeSelector = nodeSelector
		log.Info("Assigning node selector to nsm pod", "nodeSelector", nodeSelector, "pod name", pod.Name)
	} else {
		log.Info("Got empty node selector from deployment config, skipping", "pod name", pod.Name)
	}

	toleration := nsm.config.Toleration
	if toleration != nil {
		pod.Spec.Tolerations = append(pod.Spec.Tolerations, *toleration)
		log.Info("Assigning toleration to nsm pod", "toleration", toleration, "pod name", pod.Name)
	} else {
		log.Info("Got empty toleration from deployment config, skipping", "pod name", pod.Name)
	}
}