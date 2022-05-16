package tfcli

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func must(t *testing.T, err error) {
	if err != nil {
		assert.FailNow(t, "Error: "+err.Error())
	}
}
func TestTerraformDownloadURL(t *testing.T) {
	url, err := terraformDownloadURL("1.1.6")
	if assert.NoError(t, err) {
		assert.Contains(t, url, "1.1.6")
		assert.Contains(t, url, runtime.GOOS)
		assert.Contains(t, url, runtime.GOARCH)
	}
}

func TestTerraformDownload(t *testing.T) {

	file, err := DownloadTerraform("1.1.6", false)
	must(t, err)

	tftmp := t.TempDir()
	tf := New(file, tftmp)
	tf.WithVars(map[string]string{
		"string_var": "string var",
	})
	ver, err := tf.Version()
	must(t, err)
	assert.Equal(t, "1.1.6", ver)

	err = tf.GetModule("weakpixel/test-module/tfcli", "~> 0.0.1")
	if assert.NoError(t, err) {
		must(t, tf.Init())
		must(t, tf.Apply())
		_, err := tf.Output()
		must(t, err)
		// t.Logf("Output: %v", res)
		must(t, tf.Destroy())
	}
}

func TestMergeStringArrays(t *testing.T) {
	val := mergeStringArrays([]string{"get", "--verbose"}, []string{"--var", "myvar=value"}, []string{"--backend", "yes"})
	expected := []string{"get", "--verbose", "--var", "myvar=value", "--backend", "yes"}
	assert.Equal(t, expected, val)

}

func TestMapToArgs(t *testing.T) {
	val := map[string]string{
		"var2": "value2",
		"var1": "value1",
	}
	expected := []string{
		"-var", "var1=value1",
		"-var", "var2=value2",
	}
	res := mapToArgs(val, "var")
	assert.Equal(t, len(expected), len(res))
	assert.ElementsMatch(t, expected, res)

	res = mapToArgs(nil, "var")
	assert.Equal(t, 0, len(res))
}
