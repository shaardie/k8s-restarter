package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/shaardie/k8s-restarter/pkg"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig string
	configFile string
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&configFile, "config", "", "path to the configuration file")
	flag.Parse()
}

func getK8sClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes config from kubeconfig %v, %w", kubeconfig, err)
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
