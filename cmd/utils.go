package cmd

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mvisonneau/go-helpers/logger"
	"github.com/urfave/cli"

	tfcw "github.com/mvisonneau/tfcw/lib/client"
	"github.com/mvisonneau/tfcw/lib/schemas"

	log "github.com/sirupsen/logrus"
)

var start time.Time

func configure(ctx *cli.Context) (c *tfcw.Client, cfg *schemas.Config, err error) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	lc := &logger.Config{
		Level:  ctx.GlobalString("log-level"),
		Format: ctx.GlobalString("log-format"),
	}

	if err = lc.Configure(); err != nil {
		return
	}

	cfg = &schemas.Config{}
	log.Debugf("Using config file at %s", ctx.GlobalString("config-path"))
	err = hclsimple.DecodeFile(ctx.GlobalString("config-path"), nil, cfg)
	if err != nil {
		return c, cfg, fmt.Errorf("tfcw config/hcl: %s", err.Error())
	}

	clientConfig := &tfcw.Config{
		Config: cfg,
	}

	// We do not need TFC access to render variables locally
	switch ctx.Command.Name {
	case "local":
		clientConfig.Runtime.TFE.Disabled = true
	default:
		clientConfig.Runtime.TFE.Disabled = false
		clientConfig.Runtime.TFE.Address = ctx.String("tfc-address")
		clientConfig.Runtime.TFE.Token = ctx.String("tfc-token")
	}

	c, err = tfcw.NewClient(clientConfig)
	return
}

func exit(err error, exitCode int) *cli.ExitError {
	defer log.Debugf("Executed in %s, exiting..", time.Since(start))
	if err != nil {
		log.Error(err.Error())
	}

	return cli.NewExitError("", exitCode)
}

func ExecWrapper(f func(ctx *cli.Context) (error, int)) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		return exit(f(ctx))
	}
}
