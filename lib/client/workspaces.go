package client

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/mvisonneau/tfcw/lib/schemas"
	log "github.com/sirupsen/logrus"
)

func (c *Client) createWorkspace(cfg *schemas.Config) (*tfe.Workspace, error) {
	log.Debug("Creating workspace")
	opts := tfe.WorkspaceCreateOptions{
		AutoApply:        cfg.TFC.Workspace.AutoApply,
		Name:             &cfg.TFC.Workspace.Name,
		TerraformVersion: cfg.TFC.Workspace.TerraformVersion,
		WorkingDirectory: cfg.TFC.Workspace.WorkingDirectory,
	}
	w, err := c.TFE.Workspaces.Create(c.Context, cfg.TFC.Organization, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching TFC workspace: %s", err)
	}

	log.Debugf("Workspace %s' created with ID : %s", w.Name, w.ID)

	return w, nil
}

func (c *Client) getWorkspace(organization, workspace string) (*tfe.Workspace, error) {
	log.Debug("Fetching workspace")
	w, err := c.TFE.Workspaces.Read(c.Context, organization, workspace)
	if err != nil {
		return nil, fmt.Errorf("error fetching TFC workspace: %s", err)
	}

	log.Debugf("Found workspace id for '%s': %s", w.Name, w.ID)

	return w, nil
}

// ConfigureWorkspace ensures the configuration of the workspace
func (c *Client) ConfigureWorkspace(cfg *schemas.Config, dryRun bool) error {
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace.Name)
	if err != nil {
		if (cfg.TFC.WorkspaceAutoCreate == nil ||
			*cfg.TFC.WorkspaceAutoCreate) &&
			err.Error() == "error fetching TFC workspace: resource not found" {

			if !dryRun {
				// Create the workspace
				w, err = c.createWorkspace(cfg)
				if err != nil {
					return err
				}
			} else {
				log.Warnf("[DRY-RUN] - would have created the workspace as it does not currently exists")
				return fmt.Errorf("exiting as workspace does not exist so we won't be able to simulate the dry run further")
			}
		} else {
			return err
		}
	}

	workspaceNeedToBeUpdated := false
	trueVar := true

	// Check if we actually need to trigger an update or not

	// If not explicitly set to false, we enforce workspace operations to true (remote executions)
	if cfg.TFC.Workspace.Operations == nil {
		cfg.TFC.Workspace.Operations = &trueVar
	}

	if *cfg.TFC.Workspace.Operations != w.Operations {
		workspaceNeedToBeUpdated = true
		log.Infof("Workspace operations configured with '%v', wanted '%v', we will update", w.Operations, *cfg.TFC.Workspace.Operations)
	}

	if cfg.TFC.Workspace.AutoApply != nil {
		if *cfg.TFC.Workspace.AutoApply != w.AutoApply {
			workspaceNeedToBeUpdated = true
			log.Infof("Workspace auto-apply configured with '%v', wanted '%v', we will update", w.AutoApply, *cfg.TFC.Workspace.AutoApply)
		}
	}

	if cfg.TFC.Workspace.TerraformVersion != nil {
		if *cfg.TFC.Workspace.TerraformVersion != w.TerraformVersion {
			workspaceNeedToBeUpdated = true
			log.Infof("Workspace terraform version configured with '%s', wanted '%s', we will update", w.TerraformVersion, *cfg.TFC.Workspace.TerraformVersion)
		}
	}

	if cfg.TFC.Workspace.WorkingDirectory != nil {
		if *cfg.TFC.Workspace.WorkingDirectory != w.WorkingDirectory {
			workspaceNeedToBeUpdated = true
			log.Infof("Workspace working directory configured with '%s', wanted '%s', we will update", w.WorkingDirectory, *cfg.TFC.Workspace.WorkingDirectory)
		}
	}

	if cfg.TFC.Workspace.SSHKey != nil {
		shouldUpdateSSHKey, err := c.shouldUpdateSSHKey(w, *cfg.TFC.Workspace.SSHKey)
		if err != nil {
			return err
		}

		if shouldUpdateSSHKey {
			if !dryRun {
				err = c.updateSSHKey(w, *cfg.TFC.Workspace.SSHKey)
				if err != nil {
					return fmt.Errorf("error updating TFC workspace ssh key: %s", err)
				}
			} else {
				log.Infof("[DRY-RUN] not actually updating workspace's SSH key configuration as we dry-run mode")
			}
		}
	}

	if workspaceNeedToBeUpdated {
		if !dryRun {
			opts := tfe.WorkspaceUpdateOptions{
				Name:             &cfg.TFC.Workspace.Name,
				Operations:       cfg.TFC.Workspace.Operations,
				AutoApply:        cfg.TFC.Workspace.AutoApply,
				TerraformVersion: cfg.TFC.Workspace.TerraformVersion,
				WorkingDirectory: cfg.TFC.Workspace.WorkingDirectory,
			}

			_, err = c.TFE.Workspaces.UpdateByID(c.Context, w.ID, opts)
			if err != nil {
				return fmt.Errorf("error updating TFC workspace: %s", err)
			}

			if cfg.TFC.Workspace.SSHKey != nil {
				err = c.updateSSHKey(w, *cfg.TFC.Workspace.SSHKey)
				if err != nil {
					return fmt.Errorf("error updating TFC workspace ssh key: %s", err)
				}
			}
		} else {
			log.Infof("[DRY-RUN] not actually updating workspace configuration as we dry-run mode")
		}
	}

	return nil
}

// GetWorkspaceStatus returns the status of the configured workspace
func (c *Client) GetWorkspaceStatus(cfg *schemas.Config) error {
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace.Name)
	if err != nil {
		return err
	}

	if w.Locked {
		fmt.Printf("Workspace %s is currently locked by run ID '%s'\n", cfg.TFC.Workspace.Name, w.CurrentRun.ID)
		currentRun, err := c.TFE.Runs.Read(c.Context, w.CurrentRun.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Status: %v\n", currentRun.Status)
	} else {
		fmt.Printf("Workspace %s is idle\n", cfg.TFC.Workspace.Name)
	}

	return nil
}

// GetWorkspaceCurrentRunID returns the status of the configured workspace
func (c *Client) GetWorkspaceCurrentRunID(cfg *schemas.Config) (string, error) {
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace.Name)
	if err != nil {
		return "", err
	}

	if w.Locked {
		log.Debugf("Workspace %s is currently locked by run ID '%s'\n", cfg.TFC.Workspace.Name, w.CurrentRun.ID)
		return w.CurrentRun.ID, nil
	}

	return "", fmt.Errorf("workspace %s is currently idle", cfg.TFC.Workspace.Name)
}

func (c *Client) updateSSHKey(w *tfe.Workspace, sshKeyName string) error {
	if sshKeyName == "-" {
		log.Infof("Removing currently configured SSH key")
		_, err := c.TFE.Workspaces.UnassignSSHKey(c.Context, w.ID)
		return err
	}

	sshKeys, err := c.TFE.SSHKeys.List(c.Context, w.Organization.Name, tfe.SSHKeyListOptions{})
	if err != nil {
		return err
	}

	for _, sshKey := range sshKeys.Items {
		if sshKey.Name == sshKeyName {
			log.Infof("Updating configured SSH key to '%s'", sshKey.Name)
			_, err := c.TFE.Workspaces.AssignSSHKey(c.Context, w.ID, tfe.WorkspaceAssignSSHKeyOptions{
				SSHKeyID: &sshKey.ID,
			})
			return err
		}
	}

	return fmt.Errorf("could not find ssh key '%s'", sshKeyName)
}

func (c *Client) shouldUpdateSSHKey(w *tfe.Workspace, sshKeyName string) (bool, error) {
	var sshKey *tfe.SSHKey
	var err error
	if sshKeyName == "-" && w.SSHKey != nil {
		log.Infof("Workspace ssh key should not be configured, we will remove it")
		return true, nil
	}

	if w.SSHKey == nil {
		log.Infof("Workspace ssh key not configured, wanted '%s', we will update", sshKeyName)
		return true, nil
	}

	sshKey, err = c.TFE.SSHKeys.Read(c.Context, w.SSHKey.ID)
	if err != nil {
		return false, fmt.Errorf("could not fetch ssh key from API: %v", err)
	}

	if sshKeyName != sshKey.Name {
		log.Infof("Workspace ssh key configured with '%s', wanted '%s', we will update", sshKey.Name, sshKeyName)
		return true, nil
	}

	return false, nil
}
