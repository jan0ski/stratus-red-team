package runner

import (
	"context"
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/datadog/stratus-red-team/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

const TerraformVersion = "1.1.2"

type TerraformManager struct {
	terraformBinaryPath string
	terraformVersion    string
	directory           string
}

func NewTerraformManager(terraformBinaryPath string) *TerraformManager {
	manager := TerraformManager{
		terraformVersion:    TerraformVersion,
		terraformBinaryPath: terraformBinaryPath,
	}
	manager.Initialize()
	return &manager
}

func (m *TerraformManager) Initialize() {
	// Download the Terraform binary if it doesn't exist already
	if !utils.FileExists(m.terraformBinaryPath) {
		terraformInstaller := &releases.ExactVersion{
			Product:                  product.Terraform,
			Version:                  version.Must(version.NewVersion(TerraformVersion)),
			InstallDir:               filepath.Dir(m.terraformBinaryPath),
			SkipChecksumVerification: false,
		}
		log.Println("Installing Terraform in " + m.terraformBinaryPath)
		_, err := terraformInstaller.Install(context.Background())
		if err != nil {
			log.Fatalf("error installing Terraform: %s", err)
		}
	}
}

func (m *TerraformManager) InitAndApply() (map[string]string, error) {
	terraform, err := tfexec.NewTerraform(m.directory, m.terraformBinaryPath)
	if err != nil {
		return map[string]string{}, errors.New("unable to instantiate Terraform: " + err.Error())
	}

	err = terraform.SetAppendUserAgent("stratus-red-team")
	if err != nil {
		return map[string]string{}, errors.New("unable to configure Terraform: " + err.Error())
	}

	terraformInitializedFile := path.Join(m.directory, ".terraform-initialized")
	if !utils.FileExists(terraformInitializedFile) {
		log.Println("Initializing Terraform to spin up technique prerequisites")
		err = terraform.Init(context.Background())
		if err != nil {
			return nil, errors.New("unable to Initialize Terraform: " + err.Error())
		}

		_, err = os.Create(terraformInitializedFile)
		if err != nil {
			return nil, errors.New("unable to initialize Terraform: " + err.Error())
		}

	}

	log.Println("Applying Terraform to spin up technique prerequisites")
	err = terraform.Apply(context.Background(), tfexec.Refresh(false))
	if err != nil {
		return nil, errors.New("unable to apply Terraform: " + err.Error())
	}

	rawOutputs, _ := terraform.Output(context.Background())
	outputs := make(map[string]string, len(rawOutputs))
	for outputName, outputRawValue := range rawOutputs {
		outputValue := string(outputRawValue.Value)
		// Strip the first and last quote which gets added for some reason
		outputValue = outputValue[1 : len(outputValue)-1]
		outputs[outputName] = outputValue
	}
	return outputs, nil
}

func (m *TerraformManager) Destroy() error {
	terraform, err := tfexec.NewTerraform(m.directory, m.terraformBinaryPath)
	if err != nil {
		return err
	}

	return terraform.Destroy(context.Background())
}
