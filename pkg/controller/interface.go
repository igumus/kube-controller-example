package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentController interface {
	Run(chan struct{})
	Shutdown(chan struct{})
}

func hasKeyValue(meta metav1.ObjectMeta, key, value string) bool {
	val, ok := meta.Labels[key]
	if !ok || val != value {
		return false
	}
	return true
}
