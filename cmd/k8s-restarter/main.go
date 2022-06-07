package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/shaardie/k8s-restarter/pkg"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	configFile string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	flag.StringVar(&configFile, "config", "", "path to the configuration file")
	flag.Parse()
}

func getK8sClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	var k8sConfig *rest.Config
	var err error
	if kubeconfig == "" {
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to use in cluster kubernetes config, %w", err)
		}
	} else {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubernetes config from kubeconfig %v, %w", kubeconfig, err)
		}
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset from kubernetes config, %w", err)
	}

	return clientset, nil
}

func main() {
	clientset, err := getK8sClientset(kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	cfg, err := pkg.GetConfig(configFile)
	if err != nil {
		panic(err.Error())
	}

	reconsiler := pkg.Reconsiler{
		Cfg:       cfg,
		Clientset: clientset,
	}

	for {
		err = reconsiler.Resonsile(context.Background())
		if err != nil {
			fmt.Printf("Failed to reconsile, %v", err)
		}
		time.Sleep(time.Minute)
	}
}
