package pkg

import (
	"context"
	"fmt"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deployment appv1.Deployment

func (*Deployment) GetKind() string {
	return "Deployment"
}

func (d *Deployment) StatusOK() bool {
	return d.Status.Replicas == d.Status.UpdatedReplicas &&
		d.Status.ReadyReplicas == d.Status.AvailableReplicas
}

func (d *Deployment) GetPodTemplateSpec() *v1.PodTemplateSpec {
	return &d.Spec.Template
}

func (d *Deployment) Update(ctx context.Context, clientset *kubernetes.Clientset) error {
	_, err := clientset.AppsV1().Deployments(d.Namespace).Update(context.TODO(), (*appv1.Deployment)(d), metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update %v %v/%v, %v", d.GetKind(), d.GetNamespace(), d.GetName(), err)
	}
	return nil
}
