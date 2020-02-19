package cmd

import (
	"flag"
	"testing"
	"time"

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
	if err.Error() != "tfcw config/hcl: <nil>: Configuration file not found; The configuration file  does not exist." {
		t.Fatalf("expected to get following error 'tfcw config/hcl: <nil>: Configuration file not found; The configuration file  does not exist.', got '%s'", err.Error())
	}
	if exitCode != 1 {
		t.Fatalf("expected exitCode to be 1, got %d", exitCode)
	}
}
