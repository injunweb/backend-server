package kubernetes

import (
	"context"
	"fmt"
	"log"

	"github.com/injunweb/backend-server/internal/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func Init() error {
	var restConfig *rest.Config
	var err error

	if config.AppConfig.InCluster == "true" {
		restConfig, err = rest.InClusterConfig()
	} else {
		restConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(config.AppConfig.KubeConfig))
	}

	if err != nil {
		return fmt.Errorf("failed to create kubernetes config: %v", err)
	}

	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	log.Println("Kubernetes client initialized")
	return nil
}

func NamespaceExists(namespaceName string) bool {
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return false
	}

	return true
}

func DeleteNamespace(namespaceName string) error {
	err := clientset.CoreV1().Namespaces().Delete(context.TODO(), namespaceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace %s: %v", namespaceName, err)
	}

	log.Printf("Namespace %s deleted successfully\n", namespaceName)
	return nil
}
