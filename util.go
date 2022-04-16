package tfcli

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	getter "github.com/hashicorp/go-getter"
)

// mergeStringArrays merges a list of string arrays into one string array
func mergeStringArrays(args ...[]string) []string {
	finalArgs := []string{}
	for _, argGroup := range args {
		finalArgs = append(finalArgs, argGroup...)
	}
	return finalArgs
}

// mapToArgs converts the given map into a list of arguments.
func mapToArgs(params map[string]string, optionName string) []string {
	args := []string{}
	if params == nil {
		return args
	}
	for key, value := range params {
		args = append(args, "-"+optionName, key+`=`+value)
	}
	return args
}

// DownloadTerraform downloads the terraform binary for the given version from the official source.
func DownloadTerraform(version string, force bool) (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	tfbaseDir := filepath.Join(userDir, ".tf", "cache", "terraform")
	return downloadTerraform(tfbaseDir, version, force)
}

func downloadTerraform(dir, version string, force bool) (string, error) {
	url, err := terraformDownloadURL(version)
	if err != nil {
		return "", err
	}
	file := filepath.Join(dir, version)
	tffile := filepath.Join(file, "terraform")

	if fileExists(tffile) && !force {
		return tffile, nil
	}
	opts := []getter.ClientOption{}
	client := &getter.Client{
		Ctx:     context.Background(),
		Src:     url,
		Dst:     file,
		Mode:    getter.ClientModeAny,
		Options: opts,
	}
	err = client.Get()
	if err != nil {
		return "", err
	}

	if !fileExists(tffile) {
		f, _ := ioutil.ReadDir(file)
		list := []string{}
		for _, e := range f {
			list = append(list, e.Name())
		}
		return "", fmt.Errorf("Terraform executable not found: %s, Content: %+v", tffile, list)
	}
	return tffile, nil
}

func terraformDownloadURL(version string) (string, error) {
	var b bytes.Buffer
	tpl, err := template.New("").Parse("https://releases.hashicorp.com/terraform/{{.version}}/terraform_{{.version}}_{{.goos}}_{{.goarch}}.zip")
	if err != nil {
		return "", err
	}
	err = tpl.Execute(&b, map[string]string{
		"version": version,
		"goos":    runtime.GOOS,
		"goarch":  runtime.GOARCH,
	})
	return b.String(), err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
