package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	log "github.com/sirupsen/logrus"
)

// Render handles the processing of the variables and update of their values
// on supported providers (tfc or local)
func Render(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	switch ctx.String("render-type") {
	case "tfc":
		w, err := c.ConfigureWorkspace(cfg, ctx.Bool("dry-run"))
		if err != nil {
			return 1, err
		}
		err = c.RenderVariablesOnTFC(cfg, w, ctx.Bool("dry-run"), ctx.Bool("ignore-ttls"))
		if err != nil {
			return 1, err
		}
	case "local":
		err = c.RenderVariablesLocally(cfg)
		if err != nil {
			return 1, err
		}
	case "disabled":
		log.Warningf("render-type set to disabled, not doing anything")
		return 0, nil
	default:
		return 1, fmt.Errorf("invalid render-type '%s'", ctx.String("render-type"))
	}

	return 0, nil
}
