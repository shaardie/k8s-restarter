# K8s Restarter

K8s Restarter is a small controller to restart Pods associated with Deployments, StateFulSets and DaemonSets.

It is meant to work the same way like `kubectl rollout restart` and adds an `k8s-restarter.kubernetes.io/restartedAt` annotation to `spec.template.metadata.annotations` to trigger the recreation of the Pods.

## Installation

This is

## Configuration

