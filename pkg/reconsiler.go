package pkg

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Reconsiler struct {
	Logger    *zap.Logger
	Cfg       *Config
	Clientset *kubernetes.Clientset
}

func (r *Reconsiler) Resonsile(ctx context.Context) error {
	deployments, err := r.Clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployments, %w", err)
	}
	statefulsets, err := r.Clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulsets, %w", err)
	}
	daemonsets, err := r.Clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get daemonsets, %w", err)
	}

	apps := make([]App, 0, len(deployments.Items)+len(statefulsets.Items)+len(daemonsets.Items))
	for i := range deployments.Items {
		apps = append(apps, (*Deployment)(&deployments.Items[i]))
	}
	for i := range statefulsets.Items {
		apps = append(apps, (*StatefulSet)(&statefulsets.Items[i]))
	}
	for i := range daemonsets.Items {
		apps = append(apps, (*DaemonSet)(&daemonsets.Items[i]))
	}

	for _, a := range apps {
		err := r.ResonsileApp(ctx, a)
		if err != nil {
			fmt.Printf("Failed to reconsile %v/%v, %v", a.GetNamespace(), a.GetName(), err)
		}
	}
	return nil
}

func (r *Reconsiler) ResonsileApp(ctx context.Context, app App) error {
	name := app.GetName()
	namespace := app.GetNamespace()
	kind := app.GetKind()

	logger := r.Logger.With(
		zap.Any("app", map[string]string{
			"name":      name,
			"namespace": namespace,
			"kind":      kind,
		}),
	)

	// Check for exclusion
	if r.namespaceExcluded(namespace) {
		logger.Debug("namespace excluded")
		return nil
	}
	if r.Cfg.ExcludeAnnotation != "" && HasAnnotation(app, r.Cfg.ExcludeAnnotation) {
		logger.Debug("excluded due to configured annotation")
		return nil
	}
	if r.Cfg.IncludeAnnotation != "" && !HasAnnotation(app, r.Cfg.IncludeAnnotation) {
		logger.Debug("excluded due to missing annotation")
		return nil
	}

	// Check for status
	if !app.StatusOK() {
		logger.Debug("not ready...skipping")
		return nil
	}

	// Check for age
	now := time.Now()
	last, err := getTimePodTemplateSpec(app.GetPodTemplateSpec())
	if err != nil {
		return fmt.Errorf("failed to get time from pod template from %v %v/%v, %w", kind, namespace, name, err)
	}
	if last == nil {
		t := app.GetCreationTimestamp().Time
		last = &t
	}
	if last.Add(r.Cfg.RestartInterval).After(now) {
		logger.Debug("not scheduled for a restart")
		return nil
	}

	setTimeInPodTemplateSpec(app.GetPodTemplateSpec())
	err = app.Update(context.TODO(), r.Clientset)
	if err != nil {
		return fmt.Errorf("failed to set annotations on pod template from %v %v/%v, %w", kind, namespace, name, err)
	}

	logger.Debug("restarted")
	return nil
}

func (r *Reconsiler) namespaceExcluded(ns string) bool {
	for _, excludedNs := range r.Cfg.ExcludeNamespaces {
		if ns == excludedNs {
			return true
		}
	}
	return false
}
