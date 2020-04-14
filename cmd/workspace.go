package cmd

import (
	"fmt"

	"github.com/urfave/cli"
)

// WorkspaceConfigure configures the workspace
func WorkspaceConfigure(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	if _, err = c.ConfigureWorkspace(cfg, ctx.Bool("dry-run")); err != nil {
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

// WorkspaceEnableOperations enable operations on the workspace
func WorkspaceEnableOperations(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	w, err := c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return 1, err
	}

	err = c.SetWorkspaceOperations(w, true)
	if err != nil {
		return 1, err
	}

	fmt.Printf("enabled operations on '%s/%s'\n", w.Organization.Name, w.Name)

	return 0, nil
}

// WorkspaceDisableOperations disable operations on the workspace
func WorkspaceDisableOperations(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	w, err := c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return 1, err
	}

	err = c.SetWorkspaceOperations(w, false)
	if err != nil {
		return 1, err
	}

	fmt.Printf("disabled operations on '%s/%s'\n", w.Organization.Name, w.Name)

	return 0, nil
}

// WorkspaceDeleteVariables removes managed or all variables on TFC
func WorkspaceDeleteVariables(ctx *cli.Context) (int, error) {
	c, cfg, err := configure(ctx)
	if err != nil {
		return 1, err
	}

	w, err := c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return 1, err
	}

	if ctx.Bool("all") {
		if err = c.DeleteAllWorkspaceVariables(w); err != nil {
			return 1, err
		}
	} else {
		if err = c.DeleteWorkspaceVariables(w, cfg.GetVariables()); err != nil {
			return 1, err
		}
	}

	return 0, nil
}
