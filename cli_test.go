package tfcli

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	tfbin               = ""
	tfTestFileV1orLater = `
		variable "myvar" {
			type = string
		}

		variable "myvar_from_env" { 
			type = string
		}

		output "myvar" {
			value = var.myvar
		}

		output "myvar_from_env" {
			value = var.myvar_from_env
		}
	`
)

func logBuffer(t *testing.T, out *bytes.Buffer) {
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		t.Log(scanner.Text())
	}
}

func TestSetters(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "")
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Cannot create tmp dir")
	}
	defer os.RemoveAll(tmpDir)

	tf := &terraform{
		command: "/path/to/terrform",
		dir:     tmpDir,
	}

	assert.Equal(t, tmpDir, tf.Dir())
	assert.Equal(t, filepath.Join(tmpDir, ".terraformrc"), tf.ConfigFilePath())

	envs := map[string]string{
		"hello": "world",
	}
	tf.WithEnv(envs)
	assert.Equal(t, envs, tf.env)

	backendVars := map[string]string{
		"backend-var": "value",
	}

	tf.WithBackendVars(backendVars)
	assert.Equal(t, backendVars, tf.backendVars)

	vars := map[string]string{
		"TF_VAR_hello": "world",
	}

	tf.WithVars(vars)
	assert.Equal(t, vars, tf.vars)

	creds := []RegistryCredential{
		{Type: "type123", Token: "token1234"},
	}

	tf.WithRegistry(creds)
	assert.Equal(t, creds, tf.credentials)

	err = tf.writeConfig()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	raw, err := ioutil.ReadFile(tf.ConfigFilePath())
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(raw), "type"), "Config must contain 'type'")
	assert.True(t, strings.Contains(string(raw), "token"), "Config must contain 'type'")

}

func TestGetModule(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	tfbin, err := downloadTerraform(t.TempDir(), "1.1.6", true)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Cannot download terraform")
	}
	tmpDir := t.TempDir()
	out := &bytes.Buffer{}
	tf := New(tfbin, tmpDir, out, out)
	err = tf.GetModule("weakpixel/test-module/tfcli", "~> 0.0.1")
	if !assert.NoError(t, err) {
		logBuffer(t, out)
		assert.FailNow(t, "init failed")
	}
}

// Note: Test is skipped if terraform executable cannot be found
func TestCliMock(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	tfbin, err := downloadTerraform(t.TempDir(), "1.1.6", true)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Cannot download terraform")
	}
	testfile := tfTestFileV1orLater
	tmpDir := t.TempDir()
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Cannot create tmp dir")
	}
	defer os.RemoveAll(tmpDir)
	err = ioutil.WriteFile(path.Join(tmpDir, "main.tf"), []byte(testfile), 0644)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "Cannot create main.tf")
	}
	out := &bytes.Buffer{}
	tf := New(tfbin, tmpDir, out, out)
	tf.WithEnv(map[string]string{
		"TF_VAR_myvar_from_env": "env_value",
	})
	tf.WithBackendVars(map[string]string{
		"backend": "backend_value",
	})
	tf.WithVars(map[string]string{
		"myvar": "var_value",
	})
	tf.WithRegistry([]RegistryCredential{
		{Type: "helloworld.com", Token: "assd"},
	})
	err = tf.Init()
	if !assert.NoError(t, err) {
		logBuffer(t, out)
		assert.FailNow(t, "init failed")
	}

	err = tf.Apply()
	if !assert.NoError(t, err) {
		logBuffer(t, out)
		assert.FailNow(t, "apply failed")
	}

	res, err := tf.Output()
	if !assert.NoError(t, err) {
		logBuffer(t, out)
		assert.FailNow(t, "output failed")
	}

	t.Logf("Output: %+v", res)
	assert.Equal(t, res["myvar"], "var_value")
	assert.Equal(t, res["myvar_from_env"], "env_value")

	err = tf.Destroy()
	if !assert.NoError(t, err) {
		logBuffer(t, out)
		assert.FailNow(t, "destroy failed")
	}
}
