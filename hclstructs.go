package tfcli

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type tfjson struct {
	Module map[string]tfmodule `json:"module"`
}

type tfmodule struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

// Terraform CLI configuration:
// https://www.terraform.io/cli/config/config-file
type tfconfig struct {
	Credentials []RegistryCredential `hcl:"credentials,block"`
}

// RegistryCredential defines a registry credential pair
type RegistryCredential struct {
	Type  string `hcl:"type,label"`
	Token string `hcl:"token"`
}

// writeTerraformRC writes the terraform cli configuration file to given to filepath
func writeHcl(filepath string, val interface{}) error {
	out := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(val, out.Body())
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = out.WriteTo(f)
	if err != nil {
		return err
	}
	return nil
}

func writeTerraformConfig(filepath string, credentials []RegistryCredential) error {
	tfconfig := tfconfig{
		Credentials: credentials,
	}
	return writeHcl(filepath, tfconfig)
}

// writeModuleFile writes a temporary module file
func writeModuleFile(filename string, packet, version string) error {
	tfjs := tfjson{
		Module: map[string]tfmodule{},
	}
	tfjs.Module["module"] = tfmodule{
		Source:  packet,
		Version: version,
	}
	raw, err := json.MarshalIndent(tfjs, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, raw, 0644)
}
