package runner

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/datadog/stratus-red-team/internal/providers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const StratusNamespaceName = "stratus-red-team"

var StratusNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{Name: StratusNamespaceName},
}

type K8sManager struct {
	client         *kubernetes.Clientset
	kubeConfigPath string
}

func NewK8sManager(kubeConfigPath string) *K8sManager {
	manager := K8sManager{kubeConfigPath: kubeConfigPath}
	manager.Initialize()
	return &manager
}

func (m *K8sManager) Initialize() {
	// Set up the kubernetes client
	log.Println("Using kubeconfig at path: " + m.kubeConfigPath)
	client, err := providers.K8sClient(m.kubeConfigPath)
	if err != nil {
		// No k8s config found, ask to setup kind
		var answer string
		fmt.Println("No kubeconfig found, create test cluster with Kind? (y/n)")
		fmt.Scan(&answer)
		if answer != "yes" && answer != "y" {
			log.Fatalf("error creating kubernetes client: %s", err)
		}

		// Start Kind cluster, setting up kubeconfig in the process
		err = providers.StartKindCluster()
		if err != nil {
			log.Fatalf("error creating kubernetes client: %s", err)
		}
	}

	// Client is now set up either with remote cluster or Kind
	m.client = client
}

func (m *K8sManager) InitAndApply() (map[string]string, error) {
	log.Println("Applying K8s resources to spin up technique prerequisites")
	namespace, err := m.client.CoreV1().Namespaces().Create(
		context.Background(),
		StratusNamespace,
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, errors.New("unable to create resources: " + err.Error())
	}
	outputs := map[string]string{"namespace": namespace.Name}

	return outputs, nil
}

func (m *K8sManager) Destroy() error {
	log.Println("Deleting K8s resources to cleanup technique prerequisites")
	return m.client.CoreV1().Namespaces().Delete(
		context.Background(),
		StratusNamespace.ObjectMeta.Name,
		metav1.DeleteOptions{},
	)
}
