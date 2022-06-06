package pkg

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type App interface {
	metav1.Object
	GetKind() string

	GetPodTemplateSpec() *v1.PodTemplateSpec

	StatusOK() bool
	Update(context.Context, *kubernetes.Clientset) error
}

func HasAnnotation(a App, ann string) bool {
	_, ok := a.GetAnnotations()[ann]
	return ok
}
