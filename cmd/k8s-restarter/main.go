package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/shaardie/k8s-restarter/pkg/config"
	"github.com/shaardie/k8s-restarter/pkg/controller"
	"github.com/shaardie/k8s-restarter/pkg/server"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var (
	kubeconfig         string
	configFile         string
	leaseLockName      string
	leaseLockNamespace string
	id                 string
	debug              bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	flag.StringVar(&leaseLockName, "lease-lock-name", "", "the lease lock resource name")
	flag.StringVar(&id, "id", uuid.New().String(), "the holder identity name")
	flag.StringVar(&leaseLockNamespace, "lease-lock-namespace", "", "the lease lock resource namespace")
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

	// Run Server
	server := server.New(logger, ":8080")
	go func() {
		err := server.Run()
		if err != nil && err != http.ErrServerClosed {
			logger.Sugar().Fatalw("Failure while running server", "error", err)
		}
	}()

	ctrl := controller.Controller{
		Logger:    logger,
		Cfg:       cfg,
		Clientset: clientset,
		Server:    server,
	}

	// Create Context with Cancel Option
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Stop controller and cancel context on shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		logger.Info("Received termination, signaling shutdown")
		ctrl.Stop()
		cancel()
	}()

	// Run Controller with Leaderelection
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      leaseLockName,
				Namespace: leaseLockNamespace,
			},
			Client: clientset.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: id,
			},
		},
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				logger.Info("Start leading")
				ctrl.Run(ctx)
			},
			OnStoppedLeading: func() {
				logger.Sugar().Infow("leader lost", "id", id)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				if identity == id {
					return
				}
				logger.Sugar().Infow("new leader elected", "identity", identity)
			},
		},
	})
}
