package runner

import (
	"fmt"
	"path/filepath"

	"github.com/datadog/stratus-red-team/internal/providers"
	"github.com/datadog/stratus-red-team/internal/state"
	"github.com/datadog/stratus-red-team/pkg/stratus"
	"k8s.io/client-go/util/homedir"
)

type ConfigManager interface {
	Initialize()
	InitAndApply() (map[string]string, error)
	Destroy() error
}

func NewConfigManager(technique *stratus.AttackTechnique, stateManager state.StateManager) (ConfigManager, error) {
	var cm ConfigManager
	switch technique.Platform {
	case stratus.AWS:
		filepath := filepath.Join(stateManager.GetRootDirectory(), "terraform")
		cm = NewTerraformManager(filepath)
	case stratus.Kubernetes:
		filepath := filepath.Join(homedir.HomeDir(), providers.KubeconfigDefaultDirectory, providers.KubeconfigDefaultFile)
		cm = NewK8sManager(filepath)
	default:
		return nil, fmt.Errorf("platform %s for technique %s is not supported", technique.Platform, technique.FriendlyName)
	}

	return cm, nil
}
