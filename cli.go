package tfcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Terraform interface
type Terraform interface {
	Init() error
	Apply() error
	Destroy() error
	Output() (map[string]string, error)
	Dir() string
	WithRegistry(credentials []RegistryCredential)
	GetModule(moduleSource, version string) error
	WithBackendVars(backendVars map[string]string)
	WithVars(backendVars map[string]string)
	WithEnv(env map[string]string)
	ConfigFilePath() string
	Version() (string, error)
	SetStdout(stdout io.Writer) Terraform
	Stderr() io.Writer
	Stdout() io.Writer
	SetStderr(stderr io.Writer) Terraform

	SetDir(dir string) Terraform
}

// Version version
type Version struct {
	Version string `json:"terraform_version"`
}

// New creates a new Terraform cli instance.
// 		tfBin - File path to terraform binary to use
// 		dir - Working directory used for terraform execution
// 		stdout - default stdout for terraform execution
// 		stderr - default stderr for terraform execution
func New(tfBin, dir string) Terraform {
	logrus.Debugf("New Terraform Client. Executable: '%s', Working Dir: '%s'", tfBin, dir)
	return &terraform{
		command: filepath.FromSlash(tfBin),
		stdout:  io.Discard,
		stderr:  io.Discard,
		dir:     filepath.FromSlash(dir),
	}
}

type terraform struct {
	command     string
	stdout      io.Writer
	stderr      io.Writer
	dir         string
	backendVars map[string]string
	vars        map[string]string
	env         map[string]string
	credentials []RegistryCredential
}

func (t *terraform) Stderr() io.Writer {
	return t.stderr
}

func (t *terraform) Stdout() io.Writer {
	return t.stdout
}

func (t *terraform) SetStdout(stdout io.Writer) Terraform {
	t.stdout = stdout
	return t
}

func (t *terraform) SetStderr(stderr io.Writer) Terraform {
	t.stderr = stderr
	return t
}

func (t *terraform) SetDir(dir string) Terraform {
	t.dir = dir
	return t
}

func (t *terraform) Dir() string {
	return t.dir
}

// WithRegistry configures the terraform registry in the Terraform working directory
func (t *terraform) WithRegistry(credentials []RegistryCredential) {
	t.credentials = credentials
}

// GetModule downloads the given module and prepares the workspace
// Configure the terraform registry (WithRegistry) if module needs
// credentials to be accessed
func (t *terraform) GetModule(moduleSource, version string) error {
	logrus.Debugf("Terraform GetModule: %s (%s)", moduleSource, version)
	err := t.writeConfig()
	if err != nil {
		return err
	}
	err = t.downloadModule(moduleSource, version)
	if err != nil {
		return err
	}
	return t.copyModuleToWorkingDir()
}

// WithBackendVars configures the backend for all relevant commands.
func (t *terraform) WithBackendVars(backendVars map[string]string) {
	t.backendVars = backendVars
}

// WithVars sets terraform variables for apply/destroy
func (t *terraform) WithVars(vars map[string]string) {
	t.vars = vars
}

// WithEnv sets envrionment variables for terraform execution
func (t *terraform) WithEnv(env map[string]string) {
	t.env = env
}

func (t *terraform) Init() error {
	err := t.writeConfig()
	if err != nil {
		return err
	}
	backendArgs := mapToArgs(t.backendVars, "backend-config")
	cmd := t.newCommand([]string{"init", "-no-color", "-input=false", "-force-copy", "-get=true"}, backendArgs)
	return t.run(cmd)
}

func (t *terraform) Apply() error {
	varsArgs := mapToArgs(t.vars, "var")
	cmd := t.newCommand([]string{"apply", "-no-color", "-input=false", "-auto-approve"}, varsArgs)
	return t.run(cmd)
}

func (t *terraform) Destroy() error {
	varsArgs := mapToArgs(t.vars, "var")
	cmd := t.newCommand([]string{"destroy", "-no-color", "-input=false", "-auto-approve"}, varsArgs)
	// implementation of workaround, described in https://github.com/hashicorp/terraform/issues/18026
	// Note: Make sure to not overwrite default envs set by "newCommand"
	cmd.Env = append(cmd.Env, "TF_WARN_OUTPUT_ERRORS=1")
	return t.run(cmd)
}

func (t *terraform) Output() (map[string]string, error) {
	cmd := t.newCommand([]string{"output", "-json"})
	buffer := bytes.Buffer{}
	cmd.Stdout = &buffer
	err := t.run(cmd)
	if err != nil {
		return nil, err
	}
	return readOutVars(buffer.Bytes())
}

func (t *terraform) Version() (string, error) {
	cmd := t.newCommand([]string{"version", "-json"})
	buffer := bytes.Buffer{}
	cmd.Stdout = &buffer
	err := t.run(cmd)
	if err != nil {
		return "", err
	}
	v := &Version{}
	err = json.Unmarshal(buffer.Bytes(), v)
	if err != nil {
		return "", err
	}
	return v.Version, nil
}

// private

func (t *terraform) writeConfig() error {
	if t.credentials == nil || len(t.credentials) == 0 {
		return nil
	}
	err := writeTerraformConfig(t.ConfigFilePath(), t.credentials)
	if err != nil {
		return fmt.Errorf("cannot configure terraform registry credentials: %s", err)
	}
	return nil
}

func (t *terraform) copyModuleToWorkingDir() error {
	modulePath := filepath.Join(t.dir, ".terraform", "modules", "module")
	list, err := ioutil.ReadDir(modulePath)
	if err != nil {
		return fmt.Errorf("preparing terraform module failed, cannot read module source: %s", err)
	}
	for _, f := range list {
		err := os.Rename(filepath.Join(modulePath, f.Name()), filepath.Join(t.dir, f.Name()))
		if err != nil {
			return fmt.Errorf("preparing terraform module failed, can not move module source file: %s", err)
		}
	}
	return nil
}

func (t *terraform) downloadModule(moduleSource, version string) error {
	file := filepath.Join(t.dir, "main.tf.json")
	err := writeModuleFile(file, moduleSource, version)
	if err != nil {
		return fmt.Errorf("cannot prepare module file for '%s' version '%s': %s", moduleSource, version, err)
	}
	// Make sure to delete the temporary main file after downloading the module
	defer os.Remove(file)
	cmd := t.newCommand([]string{"get", "-no-color"})
	err = t.run(cmd)
	if err != nil {
		return fmt.Errorf("cannot download module '%s' version '%s': %s", moduleSource, version, err)
	}
	return nil
}

func (t *terraform) newCommand(args ...[]string) *exec.Cmd {
	cmd := exec.Command(t.command, mergeStringArrays(args...)...)
	cmd.Stdout = t.stdout
	cmd.Stderr = t.stderr
	cmd.Dir = t.dir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TF_IN_AUTOMATION=true")
	if t.env != nil {
		for k, v := range t.env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	if t.credentials != nil && len(t.credentials) > 0 {
		cmd.Env = append(cmd.Env, "TF_CLI_CONFIG_FILE="+t.ConfigFilePath())
	}
	return cmd
}

func (t *terraform) run(cmd *exec.Cmd) error {
	logrus.Debugf("Command Run: '%s'", cmd.String())
	logrus.Debugf("Command Env: %+v", cmd.Env)
	return cmd.Run()
}

func (t *terraform) ConfigFilePath() string {
	return filepath.Join(t.dir, ".terraformrc")
}
