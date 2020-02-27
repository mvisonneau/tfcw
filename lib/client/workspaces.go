package client

import (
	"fmt"

	"github.com/mvisonneau/tfcw/lib/schemas"
	"github.com/prometheus/common/log"
)

// GetWorkspaceStatus returns the status of the configured workspace
func (c *Client) GetWorkspaceStatus(cfg *schemas.Config) error {
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace)
	if err != nil {
		return err
	}

	if w.Locked {
		fmt.Printf("Workspace %s is currently locked by run ID '%s'\n", cfg.TFC.Workspace, w.CurrentRun.ID)
		currentRun, err := c.TFE.Runs.Read(c.Context, w.CurrentRun.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Status: %v\n", currentRun.Status)
	} else {
		fmt.Printf("Workspace %s is idle\n", cfg.TFC.Workspace)
	}

	return nil
}

// GetWorkspaceCurrentRunID returns the status of the configured workspace
func (c *Client) GetWorkspaceCurrentRunID(cfg *schemas.Config) (string, error) {
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace)
	if err != nil {
		return "", err
	}

	if w.Locked {
		log.Debugf("Workspace %s is currently locked by run ID '%s'\n", cfg.TFC.Workspace, w.CurrentRun.ID)
		return w.CurrentRun.ID, nil
	}

	return "", fmt.Errorf("workspace %s is currently idle", cfg.TFC.Workspace)
}
