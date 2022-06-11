package controller

import (
	"context"
	"fmt"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StatefulSet fulfilling the App Interface
type StatefulSet appv1.StatefulSet

func (*StatefulSet) GetKind() string {
	return "StatefulSet"
}

func (s *StatefulSet) StatusOK() bool {
	return s.Status.Replicas != s.Status.UpdatedReplicas ||
		s.Status.ReadyReplicas != s.Status.AvailableReplicas
}

func (s *StatefulSet) GetPodTemplateSpec() *v1.PodTemplateSpec {
	return &s.Spec.Template
}

func (s *StatefulSet) Update(ctx context.Context, clientset *kubernetes.Clientset) error {
	_, err := clientset.AppsV1().StatefulSets(s.Namespace).Update(context.TODO(), (*appv1.StatefulSet)(s), metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update %v %v/%v, %v", s.GetKind(), s.GetNamespace(), s.GetName(), err)
	}
	return nil
}
