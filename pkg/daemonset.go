package pkg

import (
	"context"
	"fmt"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DaemonSet appv1.DaemonSet

func (*DaemonSet) GetKind() string {
	return "DaemonSet"
}

func (d *DaemonSet) StatusOK() bool {
	return d.Status.NumberMisscheduled == 0 &&
		d.Status.NumberUnavailable == 0 &&
		d.Status.DesiredNumberScheduled == d.Status.NumberAvailable &&
		d.Status.DesiredNumberScheduled == d.Status.NumberReady
}

func (d *DaemonSet) GetPodTemplateSpec() *v1.PodTemplateSpec {
	return &d.Spec.Template
}

func (d *DaemonSet) Update(ctx context.Context, clientset *kubernetes.Clientset) error {
	_, err := clientset.AppsV1().DaemonSets(d.Namespace).Update(context.TODO(), (*appv1.DaemonSet)(d), metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update %v %v/%v, %v", d.GetKind(), d.GetNamespace(), d.GetName(), err)
	}
	return nil
}
