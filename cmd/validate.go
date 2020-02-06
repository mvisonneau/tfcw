package cmd

import (
	"github.com/urfave/cli"
)

func Validate(ctx *cli.Context) error {
	c, cfg, err := configure(ctx)
	if err != nil {
		return exit(err, 1)
	}

	err = c.ProcessAllVariables(cfg)
	if err != nil {
		return exit(err, 1)
	}

	return exit(nil, 0)
}
