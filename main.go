package main

import (
	"auth0-ingress-controller/pkg/config"
	log "github.com/sirupsen/logrus"
	"os"

	"auth0-ingress-controller/pkg/controller"
	"auth0-ingress-controller/pkg/kube"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	//log.SetFormatter(&log.TextFormatter{
	//	DisableColors: false,
	//	FullTimestamp: true,
	//})
	log.SetFormatter(&log.TextFormatter{
		ForceColors:               false,
		DisableColors:             false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           "2006-01-02T15:04:05.000",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	})
}

func main() {
	currentNamespace := os.Getenv("KUBERNETES_NAMESPACE")
	var resource = "ingresses"
	var restClient rest.Interface
	//var added, deleted int

	if len(currentNamespace) == 0 {
		currentNamespace = v1.NamespaceAll
		log.Warn("Warning: KUBERNETES_NAMESPACE is unset, will monitor ingresses in all namespaces.")

	}

	var kubeClient kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = kube.GetClientOutOfCluster()
	} else {
		kubeClient = kube.GetClient()
	}
	restClient = kubeClient.ExtensionsV1beta1().RESTClient()
	config := config.GetControllerConfig()
	controller := controller.NewAuth0Controller(currentNamespace, kubeClient, config, resource, restClient)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
