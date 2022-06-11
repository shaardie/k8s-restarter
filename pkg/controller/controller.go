package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/shaardie/k8s-restarter/pkg/config"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Controller is responsible for the reconciliation
type Controller struct {
	Logger    *zap.Logger
	Cfg       *config.Config
	Clientset *kubernetes.Clientset
}

// Reconcile runs the reconciliation loop on all apps/*
func (c *Controller) Reconcile(ctx context.Context) error {
	deployments, err := c.Clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployments, %w", err)
	}
	statefulsets, err := c.Clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulsets, %w", err)
	}
	daemonsets, err := c.Clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
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
		err := c.ReconcileApp(ctx, a)
		if err != nil {
			fmt.Printf("Failed to reconcile %v/%v, %v", a.GetNamespace(), a.GetName(), err)
		}
	}
	return nil
}

// ReconcileApp reconciles a single app
func (c *Controller) ReconcileApp(ctx context.Context, app App) error {
	name := app.GetName()
	namespace := app.GetNamespace()
	kind := app.GetKind()

	logger := c.Logger.With(
		zap.Any("app", map[string]string{
			"name":      name,
			"namespace": namespace,
			"kind":      kind,
		}),
	)

	// Check for exclusion
	if c.namespaceExcluded(namespace) {
		logger.Debug("namespace excluded")
		return nil
	}
	if c.Cfg.ExcludeAnnotation != "" && HasAnnotation(app, c.Cfg.ExcludeAnnotation) {
		logger.Debug("excluded due to configured annotation")
		return nil
	}
	if c.Cfg.IncludeAnnotation != "" && !HasAnnotation(app, c.Cfg.IncludeAnnotation) {
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
	if last.Add(c.Cfg.RestartInterval).After(now) {
		logger.Debug("not scheduled for a restart")
		return nil
	}

	setTimeInPodTemplateSpec(app.GetPodTemplateSpec())
	err = app.Update(context.TODO(), c.Clientset)
	if err != nil {
		return fmt.Errorf("failed to set annotations on pod template from %v %v/%v, %w", kind, namespace, name, err)
	}

	logger.Debug("restarted")
	return nil
}

func (c *Controller) namespaceExcluded(ns string) bool {
	for _, excludedNs := range c.Cfg.ExcludeNamespaces {
		if ns == excludedNs {
			return true
		}
	}
	return false
}

// HasAnnotation is a helper function to check, if the App has a specific
// Annotation
func HasAnnotation(a App, ann string) bool {
	_, ok := a.GetAnnotations()[ann]
	return ok
}
