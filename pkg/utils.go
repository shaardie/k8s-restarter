package pkg

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
)

const restartedAtAnnotation = "k8s-restarter.kubernetes.io/restartedAt"

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

func setTimeInPodTemplateSpec(pts *v1.PodTemplateSpec) {
	if pts.Annotations == nil {
		pts.Annotations = make(map[string]string)
	}
	pts.Annotations[restartedAtAnnotation] = time.Now().Format(time.RFC3339)
}
