package cmd

import (
	tfcw "github.com/mvisonneau/tfcw/lib/client"
	"github.com/urfave/cli"
)

func Render(ctx *cli.Context) (error, int) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return err, 1
	}

	err = c.RenderVariables(cfg, tfcw.RenderVariablesType(ctx.Command.Name), ctx.Bool("dry-run"))
	if err != nil {
		return err, 1
	}

	return nil, 0
}

func TFERun(ctx *cli.Context) (error, int) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return err, 1
	}

	if !ctx.Bool("no-render") {
		renderVariablesType := tfcw.RenderVariablesTypeTFC
		if ctx.Bool("render-local") {
			renderVariablesType = tfcw.RenderVariablesTypeLocal
		}

		err = c.RenderVariables(cfg, renderVariablesType, false)
		if err != nil {
			return err, 1
		}
	}

	err = c.Run(cfg, ctx.String("tf-config-folder"), tfcw.TFERunType(ctx.Command.Name))
	if err != nil {
		return err, 1
	}

	return nil, 0
}
