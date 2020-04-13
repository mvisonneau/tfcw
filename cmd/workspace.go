package cmd

import (
	"fmt"

	"github.com/urfave/cli"
)

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
