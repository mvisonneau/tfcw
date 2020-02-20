package cmd

import (
	"flag"
	"testing"
	"time"

	"github.com/mvisonneau/go-helpers/test"
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

func TestRenderWithDefaultValues(t *testing.T) {
	ctx, _, _ := NewTestContext()
	err, exitCode := Render(ctx)
	test.Expect(t, err.Error(), "tfcw config/hcl: <nil>: Configuration file not found; The configuration file  does not exist.")
	test.Expect(t, exitCode, 1)
}
