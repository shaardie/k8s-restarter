package controller

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// App is an interface generalize the access to Deployments, StatefulSets and
// DaemonSets
type App interface {
	metav1.Object

	// GetKind returns the actual Kind of the Kubernetes Resource
	GetKind() string

	// GetPodTemplateSpec retuns the inner PodTemplateSpec
	GetPodTemplateSpec() *v1.PodTemplateSpec

	// StatusOK indicated, if the Kubernetes Resource is properly running,
	// e.g. if a Deployment has a proper Number of Pods.
	StatusOK() bool

	// Update updates the app with the the new set values
	Update(context.Context, *kubernetes.Clientset) error
}
