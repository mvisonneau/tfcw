package cmd

import (
	"fmt"

	"github.com/mvisonneau/tfcw/lib/client"
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

// RunCreate create a run on TFC
func RunCreate(ctx *cli.Context) (int, error) {
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

	if err = c.CreateRun(cfg, &client.TFECreateRunOptions{
		ConfigPath:   ctx.String("tf-config-folder"),
		AutoApprove:  ctx.Bool("auto-approve"),
		AutoDiscard:  ctx.Bool("auto-discard"),
		NoPrompt:     ctx.Bool("no-prompt"),
		OutputPath:   ctx.String("output"),
		Message:      ctx.String("message"),
		StartTimeout: ctx.Duration("start-timeout"),
	}); err != nil {
		return 1, err
	}

	return 0, nil
}

// RunApprove approve a run on TFC
func RunApprove(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	runID := ctx.Args().Get(0)
	if ctx.Bool("current") {
		runID, err = c.GetWorkspaceCurrentRunID(cfg)
		if err != nil {
			return 1, err
		}
	}

	if err := c.ApproveRun(runID, ctx.String("message")); err != nil {
		return 1, err
	}

	return 0, nil
}

// RunDiscard discard a run on TFC
func RunDiscard(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	runID := ctx.Args().Get(0)
	if ctx.Bool("current") {
		runID, err = c.GetWorkspaceCurrentRunID(cfg)
		if err != nil {
			return 1, err
		}
	}

	if err := c.DiscardRun(runID, ctx.String("message")); err != nil {
		return 1, err
	}

	return 0, nil
}

// WorkspaceStatus return status of the workspace on TFC
func WorkspaceStatus(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	if err := c.GetWorkspaceStatus(cfg); err != nil {
		return 1, err
	}

	return 0, nil
}

// WorkspaceCurrentRunID return the ID of the current run on TFC
func WorkspaceCurrentRunID(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	runID, err := c.GetWorkspaceCurrentRunID(cfg)
	if err != nil {
		return 1, err
	}

	fmt.Println(runID)

	return 0, nil
}
