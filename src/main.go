package main

import (
	"context"
	"flag"
	"log"
	"time"

	rc "github.com/kubernetes-csi/external-resizer/pkg/controller"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	pc "sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

const (
	pluginName = "dothill"

	fsTypeConfigKey                   = "fsType"
	poolConfigKey                     = "pool"
	targetIQNConfigKey                = "iqn"
	portalsConfigKey                  = "portals"
	initiatorNameConfigKey            = "initiatorName"
	apiAddressConfigKey               = "apiAddress"
	credentialsSecretNameConfigKey    = "credentialsSecretName"
	uniqueInitiatorNameByPvcConfigKey = "uniqueInitiatorNameByPvc"
	usernameSecretKey                 = "username"
	passwordSecretKey                 = "password"
	storageClassAnnotationKey         = "storageClass"

	maximumLUN                    = 255
	hostDoesNotExistsErrorCode    = -10386
	hostMapDoesNotExistsErrorCode = -10074
)

func start(config *rest.Config) error {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "unable to get k8s client")
	}

	klog.V(1).Info("fetching API server version")
	serverVersion, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		return errors.Wrap(err, "failed to get Kubernetes API server version")
	}
	klog.V(1).Infof("server version is %s", serverVersion.GitVersion)

	factory := informers.NewSharedInformerFactory(kubeClient, time.Minute)
	controller := NewDothillController(pluginName, kubeClient)
	provisioner := pc.NewProvisionController(
		kubeClient,
		pluginName,
		controller,
		serverVersion.GitVersion,
	)
	resizer := rc.NewResizeController(
		pluginName,
		controller,
		kubeClient,
		time.Minute,
		factory,
	)

	klog.Info("starting controller")
	factory.Start(wait.NeverStop)
	go func() { resizer.Run(1, context.TODO()) }()
	provisioner.Run(wait.NeverStop)

	return nil
}

func loadConfiguration(kubeconfigPath string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if len(kubeconfigPath) > 0 {
		klog.V(1).Infof("reading config from %s", kubeconfigPath)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		klog.V(1).Info("fetching configuration from within the cluster")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, errors.Wrap(err, "unable to get kubernetes client config")
	}

	klog.V(1).Infof("loaded configuration, API server is at %s", config.Host)
	klog.V(2).Infof("loaded configuration: %+v", config)
	return config, nil
}

func main() {
	kubeconfigPath := flag.String("kubeconfig", "", "path to the kubeconfig file to use instead of in-cluster configuration")
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	klog.Infof("starting dothill storage controller v%s", version)
	config, err := loadConfiguration(*kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	err = start(config)
	if err != nil {
		klog.Fatal(err)
	}
}
