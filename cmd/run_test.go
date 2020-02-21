package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/mvisonneau/go-helpers/test"
	"github.com/urfave/cli"
)

var (
	wd, _ = os.Getwd()
)

const (
	validConfig = `
tfc {
	organization = "foo"
	workspace = "bar"
}
`
)

func NewTestContext() (ctx *cli.Context, flags, globalFlags *flag.FlagSet) {
	app := cli.NewApp()
	app.Name = "tfcw"

	app.Metadata = map[string]interface{}{
		"startTime": time.Now(),
	}

	globalFlags = flag.NewFlagSet("test", flag.ContinueOnError)
	globalCtx := cli.NewContext(app, globalFlags, nil)

	flags = flag.NewFlagSet("test", flag.ContinueOnError)
	ctx = cli.NewContext(app, flags, globalCtx)

	globalFlags.String("log-level", "fatal", "")
	globalFlags.String("log-format", "text", "")

	return
}

func createTestConfigFile(config string) (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "tfcw-test-cfg-")

	if _, err = tmpFile.Write([]byte(config)); err != nil {
		return "", fmt.Errorf("Failed to write to temporary file : %s", err.Error())
	}

	if err = tmpFile.Close(); err != nil {
		return "", fmt.Errorf("Failed to close temporary file : %s", err.Error())
	}

	configPath := fmt.Sprint(tmpFile.Name(), ".hcl")
	if err = os.Rename(tmpFile.Name(), fmt.Sprint(tmpFile.Name(), ".hcl")); err != nil {
		return "", fmt.Errorf("Failed to rename temporary file with file extension : %s", err.Error())
	}

	return configPath, nil
}

func TestRenderWithDefaultValues(t *testing.T) {
	ctx, _, _ := NewTestContext()
	exitCode, err := Render(ctx)
	test.Expect(t, err.Error(), "tfcw config/hcl: <nil>: Configuration file not found; The configuration file  does not exist.")
	test.Expect(t, exitCode, 1)
}

func TestRenderLocalWithValidConfig(t *testing.T) {
	tmpFilePath, err := createTestConfigFile(validConfig)
	if err != nil {
		t.Fatalf(fmt.Sprintf("error whilst creating temporary config file : %s", err.Error()))
	}
	defer os.Remove(tmpFilePath)

	ctx, _, globalFlags := NewTestContext()
	ctx.Command.Name = "local"
	globalFlags.String("config-path", tmpFilePath, "")

	defer os.Remove(fmt.Sprint(wd, "/tfcw.auth.tfvars"))
	defer os.Remove(fmt.Sprint(wd, "/tfcw.env"))
	exitCode, err := Render(ctx)
	test.Expect(t, err, nil)
	test.Expect(t, exitCode, 0)
}

func TestTFERunWithDefaultValues(t *testing.T) {
	ctx, _, _ := NewTestContext()
	exitCode, err := TFERun(ctx)
	test.Expect(t, err.Error(), "tfcw config/hcl: <nil>: Configuration file not found; The configuration file  does not exist.")
	test.Expect(t, exitCode, 1)
}
