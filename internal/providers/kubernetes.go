package providers

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

const (
	KubeconfigDefaultDirectory = ".kube"
	KubeconfigDefaultFile      = "config"
	StratusClusterName         = "stratus-k8s"
)

// cachedClient is package level state that will cache the Clienset
var cachedClient *kubernetes.Clientset

// kubeConfigPathValue is a very flexible string, so we have to handle it
// carefully. It can be either the environmental variable, or the CLI flag
//
// KUBECONFIG environmental variable
var KubeConfigPathValue string

// K8sClient is used to authenticate with Kubernetes and build the Kube client
// for the rest of the program given a specific kube config
func K8sClient(kubeConfigPath string) (*kubernetes.Clientset, error) {
	if cachedClient != nil {
		return cachedClient, nil
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to build kube config: %v", err)
	}
	cachedClient = client
	return cachedClient, nil
}

var clusterStarted bool = false

// StartKindCluster can be used to create a kind cluster to test stratus techniques
func StartKindCluster() error {
	if clusterStarted {
		return nil
	}
	provider := cluster.NewProvider(cluster.ProviderWithDocker(), cluster.ProviderWithLogger(cmd.NewLogger()))
	err := provider.Create(StratusClusterName)
	if err != nil {
		defer StopKindCluster()
		return fmt.Errorf("unable to create kind test cluster: %v", err)
	}
	err = provider.ExportKubeConfig(StratusClusterName, KubeConfigPathValue)
	if err != nil {
		return fmt.Errorf("unable to export test kube config: %v", err)
	}
	clusterStarted = true
	return nil
}

// StopKindCluster can be used to stop the test cluster
func StopKindCluster() error {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())
	err := provider.Delete(StratusClusterName, KubeConfigPathValue)
	if err != nil {
		return fmt.Errorf("unable to delete kind test cluster: %v", err)
	}
	return nil
}
