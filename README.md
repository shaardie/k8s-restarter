# K8s Restarter

K8s Restarter is a small controller to restart Pods associated with Deployments, StateFulSets and DaemonSets.

It is meant to work the same way like `kubectl rollout restart` and adds an `k8s-restarter.kubernetes.io/restartedAt` annotation to `spec.template.metadata.annotations` to trigger the recreation of the Pods.

## Installation

This Controller comes with a helm chart and can be installed as easy as running:

```bash
$ helm repo add k8s-restarter https://k8s-restarter-chart.haardiek.org
$ helm install my-release k8s-restarter/k8s-restarter
```

For detailed information about the chart, take a look at the [Helm Chart Readme](./charts/k8s-restarter/README.md).

## Configuration

The configuration can be done via a configuration file in the YAML format.
You can exclude Namespace and specific Apps from being restarted as well as using whitelist Annotations.
The configuration is also explained in the [Helm Chart Readme](./charts/k8s-restarter/README.md) and the format can also be seen in the [values.yaml](./charts/k8s-restarter/values.yaml#78)


## Building and Testing

You can build this controller by running

```bash
$ make k8s-restarter
```

and you can also test this out of cluster by providing a `kubeconfig` and a proper `configuration` and run something like this on the command line:

```bash
$ cat << EOF > config.yaml
reconcilationInterval: 60s
restartInterval: 10m
includeAnnotation: ""
excludeAnnotation: ""
excludeNamespaces: []
EOF
$ ./k8s-restarter -kubeconfig ~/.kube/config -config config.yaml -lease-lock-namespace default -lease-lock-name test
```
