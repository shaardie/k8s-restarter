package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/shaardie/k8s-restarter/pkg/config"
	"github.com/shaardie/k8s-restarter/pkg/server"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	restartedAtAnnotation = "k8s-restarter.kubernetes.io/restartedAt"
	minInterval           = 50 * time.Microsecond
)

// Controller is responsible for the reconcilation
type Controller struct {
	Logger    *zap.Logger
	Cfg       *config.Config
	Clientset *kubernetes.Clientset
	Server    *server.Server
	stop      chan struct{}
	done      chan struct{}
}

// reconcilationInfo holds information about a reconsilation loop
type reconcilationInfo struct {
	Excluded  int `json:"excluded"`
	Skipped   int `json:"skipped"`
	Restarted int `json:"restarted"`
}

func (c *Controller) Stop() {
	if c.stop == nil {
		return
	}
	close(c.stop)
	<-c.done
	c.Logger.Info("Stopped")
}

func (c *Controller) Run(ctx context.Context) {
	c.stop = make(chan struct{})
	c.done = make(chan struct{})
	var interval time.Duration
	shoudRun := func() bool {
		select {
		case <-c.stop:
			close(c.done)
			return false
		default:
			return true
		}
	}
	for shoudRun() {
		if interval >= 0 {
			time.Sleep(minInterval)
			interval -= minInterval
			continue
		}
		err := c.reconcile(ctx)
		if err == nil {
			c.Server.SetHealth("controller", true)
		} else {
			c.Logger.Sugar().Errorw("Failed to reconcile", "error", err)
			c.Server.SetHealth("controller", false)
		}
		interval = c.Cfg.ReconcilationInterval
	}
}

// reconcile runs the reconcilation loop on all apps/*
func (c *Controller) reconcile(ctx context.Context) error {
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

	info := reconcilationInfo{}
	for _, a := range apps {
		err := c.reconcileApp(ctx, a, &info)
		if err != nil {
			fmt.Printf("Failed to reconcile %v/%v, %v", a.GetNamespace(), a.GetName(), err)
		}
	}
	opsExcluded.Set(float64(info.Excluded))
	opsRestarts.Set(float64(info.Restarted))
	opsSkips.Set(float64(info.Skipped))
	opsExcludedHisto.Observe(float64(info.Excluded))
	opsRestartsHisto.Observe(float64(info.Restarted))
	opsSkipsHisto.Observe(float64(info.Skipped))
	c.Logger.Sugar().Infow("Reconciled", "info", info)
	return nil
}

// reconcileApp reconciles a single app
func (c *Controller) reconcileApp(ctx context.Context, app App, info *reconcilationInfo) error {
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

	// First check for exclusion and then for inclusion
	if shouldSelect(app, c.Cfg.Exclude, false) ||
		!shouldSelect(app, c.Cfg.Include, true) {
		info.Excluded++
		logger.Debug("Excluded")
		return nil
	}

	// Check for status
	if !app.StatusOK() {
		info.Skipped++
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
		info.Skipped++
		return nil
	}

	setTimeInPodTemplateSpec(app.GetPodTemplateSpec())
	err = app.Update(context.TODO(), c.Clientset)
	if err != nil {
		return fmt.Errorf("failed to set annotations on pod template from %v %v/%v, %w", kind, namespace, name, err)
	}

	logger.Debug("restarted")
	info.Restarted++
	return nil
}

type selectable interface {
	GetNamespace() string
	GetLabels() map[string]string
}

func shouldSelect(s selectable, matcher config.Matcher, defaultValue bool) bool {
	// Mather is not enabled, so everything matches
	if !matcher.Enabled {
		return defaultValue
	}

	for _, selector := range matcher.Selectors {
		// Only check further if namespace either is empty or matches
		if selector.Namespace != "" && s.GetNamespace() != selector.Namespace {
			continue
		}
		// No MatchLabels or Matchlabels match
		if len(selector.MatchLabels) == 0 ||
			matchLabels(s.GetLabels(), selector.MatchLabels) {
			return true
		}
	}
	return false
}

// matchLabels checks if match are in labels.
func matchLabels(labels, match map[string]string) bool {
	for k, v := range match {
		if l, ok := labels[k]; !ok || l != v {
			return false
		}
	}
	return true
}

// getTimePodTemplateSpec get the restartAtAnnotation from a PodTemplateSpec.
// If not set, returns nil
func getTimePodTemplateSpec(pts *v1.PodTemplateSpec) (*time.Time, error) {
	s, ok := pts.Annotations[restartedAtAnnotation]
	if !ok {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, fmt.Errorf("unable to parse time string %v, %w", s, err)
	}
	return &t, err
}

// setTimeInPodTemplateSpec sets the restartAtAnnotation on a PodTemplateSpec
func setTimeInPodTemplateSpec(pts *v1.PodTemplateSpec) {
	if pts.Annotations == nil {
		pts.Annotations = make(map[string]string)
	}
	pts.Annotations[restartedAtAnnotation] = time.Now().Format(time.RFC3339)
}
