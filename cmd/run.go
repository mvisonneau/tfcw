package cmd

import (
	tfcw "github.com/mvisonneau/tfcw/lib/client"
	"github.com/urfave/cli"
)

// Render handles the processing of the variables and update of their values
// on supported providers (tfc or local)
func Render(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	err = c.RenderVariables(cfg, tfcw.RenderVariablesType(ctx.Command.Name), ctx.Bool("dry-run"))
	if err != nil {
		return 1, err
	}

	return 0, nil
}

// TFERun handles Terraform runs over TFC
func TFERun(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	if !ctx.Bool("no-render") {
		renderVariablesType := tfcw.RenderVariablesTypeTFC
		if ctx.Bool("render-local") {
			renderVariablesType = tfcw.RenderVariablesTypeLocal
		}

		err = c.RenderVariables(cfg, renderVariablesType, false)
		if err != nil {
			return 1, err
		}
	}

	err = c.Run(cfg, ctx.String("tf-config-folder"), tfcw.TFERunType(ctx.Command.Name))
	if err != nil {
		return 1, err
	}

	return 0, nil
}
