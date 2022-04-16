package tfcli

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsimple"

	"github.com/stretchr/testify/assert"
)

func TestWriteTerraformConfig(t *testing.T) {
	// Note: hclsimple expects .hcl extension.
	file, err := ioutil.TempFile("", "TestWriteTerraformConfig*.hcl")
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	defer os.Remove(file.Name())

	creds := []RegistryCredential{
		{Type: "hello", Token: "123"},
	}

	err = writeTerraformConfig(file.Name(), creds)
	assert.NoErrorf(t, err, "writeTerraformConfig must not fail")

	config := &tfconfig{}

	err = hclsimple.DecodeFile(file.Name(), nil, config)

	assert.NoErrorf(t, err, "readHcl must not fail")

	assert.ElementsMatch(t, config.Credentials, creds)

}

func TestWriteModuleFile(t *testing.T) {
	// Note: hclsimple expects .hcl extension.
	file, err := ioutil.TempFile("", "TestWriteModuleFile*.hcl")
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	defer os.Remove(file.Name())
	err = writeModuleFile(file.Name(), "mymodule", "~> 1.0.0")
	assert.NoErrorf(t, err, "writeModuleFile must not fail")

	fd, err := os.Open(file.Name())
	assert.NoErrorf(t, err, "cannot open file")

	result := &tfjson{}
	raw, err := ioutil.ReadAll(fd)
	assert.NoErrorf(t, err, "cannot read file")
	err = json.Unmarshal(raw, result)
	assert.NoErrorf(t, err, "cannot unmarshal json")

	assert.Equal(t, 1, len(result.Module), "Must contain one module")

	m := result.Module["module"]

	assert.Equal(t, "mymodule", m.Source)
	assert.Equal(t, "~> 1.0.0", m.Version)

}
