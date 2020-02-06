package cmd

import (
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mvisonneau/tfcs/logger"
	"github.com/urfave/cli"

	tfcs "github.com/mvisonneau/tfcs/lib/client"
	"github.com/mvisonneau/tfcs/lib/schemas"

	log "github.com/sirupsen/logrus"
)

var start time.Time

func configure(ctx *cli.Context) (c *tfcs.Client, cfg *schemas.Config, err error) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	lc := &logger.Config{
		Level:  ctx.GlobalString("log-level"),
		Format: ctx.GlobalString("log-format"),
	}

	if err = lc.Configure(); err != nil {
		return
	}

	cfg = &schemas.Config{}
	err = hclsimple.DecodeFile(ctx.GlobalString("config-path"), nil, cfg)
	if err != nil {
		return
	}

	c, err = tfcs.NewClient(cfg)
	return
}

func exit(err error, exitCode int) *cli.ExitError {
	defer log.Debugf("Executed in %s, exiting..", time.Since(start))
	if err != nil {
		log.Error(err.Error())
		return cli.NewExitError("", exitCode)
	}

	return cli.NewExitError("", 0)
}
