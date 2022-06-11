package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shaardie/k8s-restarter/pkg/config"
	"github.com/shaardie/k8s-restarter/pkg/controller"
	"github.com/shaardie/k8s-restarter/pkg/server"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	configFile string
	debug      bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
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
	// Create logger
	loggerCfg := zap.NewProductionConfig()
	if debug {
		loggerCfg.Level.SetLevel(zap.DebugLevel)
	}
	logger, err := loggerCfg.Build()
	if err != nil {
		log.Fatalf("Failed to create logger, %v\n", err)
	}

	clientset, err := getK8sClientset(kubeconfig)
	if err != nil {
		logger.Sugar().Fatalw("Failed to create kubernetes client set", "error", err)
	}

	cfg, err := config.GetConfig(configFile)
	if err != nil {
		logger.Sugar().Fatalw("Unable to read config file", "config file", configFile, "error", err)
	}

	ctrl := controller.Controller{
		Logger:    logger,
		Cfg:       cfg,
		Clientset: clientset,
	}

	server := server.New(logger, ":8080")
	go func() {
		err := server.Run()
		if err != nil && err != http.ErrServerClosed {
			logger.Sugar().Fatalw("Failure while running server", "error", err)
		}
	}()

	for {
		err = ctrl.Reconcile(context.Background())
		if err != nil {
			logger.Sugar().Errorw("Failed to reconcile", "error", err)
			server.SetHealth("controller", false)
		}
		server.SetHealth("controller", true)
		time.Sleep(time.Minute)
	}
}
