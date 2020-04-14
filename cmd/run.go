package cmd

import (
	"fmt"

	tfcw "github.com/mvisonneau/tfcw/lib/client"
	"github.com/urfave/cli"

	log "github.com/sirupsen/logrus"
)

// RunCreate create a run on TFC
func RunCreate(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	w, err := c.ConfigureWorkspace(cfg, false)
	if err != nil {
		return 1, err
	}

	if !ctx.Bool("ignore-pending-runs") {
		if runID, _ := c.GetWorkspaceCurrentRunID(w); runID != "" {
			return 1, fmt.Errorf("there is already a run (%s) pending on your workspace (%s), exiting", runID, w.ID)
		}
	}

	switch ctx.String("render-type") {
	case "tfc":
		err = c.RenderVariablesOnTFC(cfg, w, false, ctx.Bool("ignore-ttls"))
		if err != nil {
			return 1, err
		}
	case "local":
		err = c.RenderVariablesLocally(cfg)
		if err != nil {
			return 1, err
		}
	case "disabled":
		log.Infof("render-type set to disabled, not rendering values")
		return 0, nil
	default:
		return 1, fmt.Errorf("invalid render-type '%s'", ctx.String("render-type"))
	}

	if err = c.CreateRun(cfg, w, &tfcw.TFCCreateRunOptions{
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

	w, err := c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return 1, err
	}

	runID := ctx.Args().Get(0)
	if ctx.Bool("current") {
		runID, err = c.GetWorkspaceCurrentRunID(w)
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

	w, err := c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return 1, err
	}

	runID := ctx.Args().Get(0)
	if ctx.Bool("current") {
		runID, err = c.GetWorkspaceCurrentRunID(w)
		if err != nil {
			return 1, err
		}
	}

	if err := c.DiscardRun(runID, ctx.String("message")); err != nil {
		return 1, err
	}

	return 0, nil
}

// RunCurrentID return the ID of the current run on TFC
func RunCurrentID(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	w, err := c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return 1, err
	}

	runID, err := c.GetWorkspaceCurrentRunID(w)
	if err != nil {
		return 1, err
	}

	fmt.Println(runID)

	return 0, nil
}
