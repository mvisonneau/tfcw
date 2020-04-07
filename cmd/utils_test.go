package cmd

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
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

func TestExit(t *testing.T) {
	err := exit(20, fmt.Errorf("test"))
	assert.Equal(t, "", err.Error())
	assert.Equal(t, 20, err.ExitCode())
}

func TestComputeConfigFilePath(t *testing.T) {
	assert.Equal(t, "./tfcw.hcl", computeConfigFilePath(".", "<working-dir>/tfcw.hcl"))
	assert.Equal(t, "/foo/bar/tfcw.hcl", computeConfigFilePath(".", "/foo/bar/tfcw.hcl"))
}

func TestHTTPSPrefixedURL(t *testing.T) {
	assert.Equal(t, "https://app.terraform.io", returnHTTPSPrefixedURL("app.terraform.io"))
	assert.Equal(t, "http://app.terraform.io", returnHTTPSPrefixedURL("http://app.terraform.io"))
}
