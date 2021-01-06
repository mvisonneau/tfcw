package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var wd, _ = os.Getwd()

const (
	validConfig = `
tfc {
	organization = "foo"

	workspace {
		name = "bar"
	}
}
`
)

func createTestConfigFile(config string) (string, string, error) {
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, "tfcw-test-cfg-")

	if _, err = tmpFile.Write([]byte(config)); err != nil {
		return "", "", fmt.Errorf("Failed to write to temporary file : %s", err.Error())
	}

	if err = tmpFile.Close(); err != nil {
		return "", "", fmt.Errorf("Failed to close temporary file : %s", err.Error())
	}

	configPath := fmt.Sprint(tmpFile.Name(), ".hcl")
	if err = os.Rename(tmpFile.Name(), fmt.Sprint(tmpFile.Name(), ".hcl")); err != nil {
		return "", "", fmt.Errorf("Failed to rename temporary file with file extension : %s", err.Error())
	}

	return tmpDir, configPath, nil
}

func TestRenderWithDefaultValues(t *testing.T) {
	ctx, _, _ := NewTestContext()
	exitCode, err := Render(ctx)
	assert.Equal(t, "tfcw config/hcl: <nil>: Configuration file not found; The configuration file  does not exist.", err.Error())
	assert.Equal(t, 1, exitCode)
}

func TestRenderLocalWithValidConfig(t *testing.T) {
	tmpDir, tmpFilePath, err := createTestConfigFile(validConfig)
	if err != nil {
		t.Fatalf(fmt.Sprintf("error whilst creating temporary config file : %s", err.Error()))
	}
	defer os.Remove(tmpFilePath)

	ctx, flags, globalFlags := NewTestContext()
	flags.String("render-type", "local", "")
	globalFlags.String("working-dir", tmpDir, "")
	globalFlags.String("config-file", tmpFilePath, "")

	defer os.Remove(fmt.Sprint(wd, "/tfcw.auto.tfvars"))
	defer os.Remove(fmt.Sprint(wd, "/tfcw.env"))
	exitCode, err := Render(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, exitCode)
}

func TestRunCreateWithDefaultValues(t *testing.T) {
	ctx, _, _ := NewTestContext()
	exitCode, err := RunCreate(ctx)
	assert.Equal(t, "tfcw config/hcl: <nil>: Configuration file not found; The configuration file  does not exist.", err.Error())
	assert.Equal(t, 1, exitCode)
}
