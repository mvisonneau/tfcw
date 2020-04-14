package client

import (
	"fmt"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/mvisonneau/tfcw/lib/schemas"
	log "github.com/sirupsen/logrus"
)

// GetWorkspace returns a workspace given its name and organization
func (c *Client) GetWorkspace(organization, workspace string) (*tfc.Workspace, error) {
	log.Debug("Fetching workspace")
	w, err := c.TFC.Workspaces.Read(c.Context, organization, workspace)
	if err != nil {
		return nil, fmt.Errorf("error fetching TFC workspace: %s", err)
	}

	log.Debugf("Found workspace id for '%s': %s", w.Name, w.ID)

	return w, nil
}

func (c *Client) createWorkspace(cfg *schemas.Config) (*tfc.Workspace, error) {
	log.Debug("Creating workspace")
	opts := tfc.WorkspaceCreateOptions{
		Name:             &cfg.Runtime.TFC.Workspace,
		AutoApply:        cfg.TFC.Workspace.AutoApply,
		TerraformVersion: cfg.TFC.Workspace.TerraformVersion,
		WorkingDirectory: cfg.TFC.Workspace.WorkingDirectory,
	}
	w, err := c.TFC.Workspaces.Create(c.Context, cfg.Runtime.TFC.Organization, opts)
	if err != nil {
		return nil, fmt.Errorf("error fetching TFC workspace: %s", err)
	}

	log.Debugf("Workspace %s' created with ID : %s", w.Name, w.ID)

	return w, nil
}

// ConfigureWorkspace check and remediate the configuration of the configured workspace
func (c *Client) ConfigureWorkspace(cfg *schemas.Config, dryRun bool) (w *tfc.Workspace, err error) {
	w, err = c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		if err.Error() == "error fetching TFC workspace: resource not found" {
			if cfg.TFC.WorkspaceAutoCreate == nil || *cfg.TFC.WorkspaceAutoCreate {
				if !dryRun {
					log.Infof("workspace '%s' does not exist on organization '%s', creating it..", cfg.Runtime.TFC.Workspace, cfg.Runtime.TFC.Organization)
					w, err = c.createWorkspace(cfg)
					if err != nil {
						return
					}
				} else {
					log.Warnf("[DRY-RUN] - would have created the workspace as it does not currently exists")
					return nil, fmt.Errorf("exiting as workspace does not exist so we won't be able to simulate the dry run further")
				}
			} else {
				return nil, fmt.Errorf("workspace does not exist and auto-create is set to false")
			}
		} else {
			return
		}
	}

	log.Info("Checking workspace configuration")

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
		var shouldUpdateSSHKey bool
		shouldUpdateSSHKey, err = c.shouldUpdateSSHKey(w, *cfg.TFC.Workspace.SSHKey)
		if err != nil {
			return
		}

		if shouldUpdateSSHKey {
			if !dryRun {
				err = c.updateSSHKey(w, *cfg.TFC.Workspace.SSHKey)
				if err != nil {
					return w, fmt.Errorf("error updating TFC workspace ssh key: %s", err)
				}
			} else {
				log.Infof("[DRY-RUN] not actually updating workspace's SSH key configuration as we dry-run mode")
			}
		}
	}

	if workspaceNeedToBeUpdated {
		if !dryRun {
			opts := tfc.WorkspaceUpdateOptions{
				Name:             cfg.TFC.Workspace.Name,
				Operations:       cfg.TFC.Workspace.Operations,
				AutoApply:        cfg.TFC.Workspace.AutoApply,
				TerraformVersion: cfg.TFC.Workspace.TerraformVersion,
				WorkingDirectory: cfg.TFC.Workspace.WorkingDirectory,
			}

			w, err = c.TFC.Workspaces.UpdateByID(c.Context, w.ID, opts)
			if err != nil {
				return w, fmt.Errorf("error updating TFC workspace: %s", err)
			}

			if cfg.TFC.Workspace.SSHKey != nil {
				err = c.updateSSHKey(w, *cfg.TFC.Workspace.SSHKey)
				if err != nil {
					return w, fmt.Errorf("error updating TFC workspace ssh key: %s", err)
				}
			}
		} else {
			log.Infof("[DRY-RUN] not actually updating workspace configuration as we dry-run mode")
		}
	}

	return
}

// GetWorkspaceStatus returns the status of the configured workspace
func (c *Client) GetWorkspaceStatus(cfg *schemas.Config) error {
	w, err := c.GetWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return err
	}

	if w.Locked {
		fmt.Printf("Workspace %s is currently locked by run ID '%s'\n", cfg.Runtime.TFC.Workspace, w.CurrentRun.ID)
		currentRun, err := c.TFC.Runs.Read(c.Context, w.CurrentRun.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Status: %v\n", currentRun.Status)
	} else {
		fmt.Printf("Workspace %s is idle\n", cfg.Runtime.TFC.Workspace)
	}

	return nil
}

// GetWorkspaceCurrentRunID returns the status of the configured workspace
func (c *Client) GetWorkspaceCurrentRunID(w *tfc.Workspace) (string, error) {
	if w.Locked {
		log.Debugf("Workspace %s is currently locked by run ID '%s'\n", w.ID, w.CurrentRun.ID)
		return w.CurrentRun.ID, nil
	}

	return "", fmt.Errorf("workspace %s is currently idle", w.ID)
}

// SetWorkspaceOperations update the workspace operations value
func (c *Client) SetWorkspaceOperations(w *tfc.Workspace, operations bool) (err error) {
	opts := tfc.WorkspaceUpdateOptions{
		Operations: tfc.Bool(operations),
	}

	_, err = c.TFC.Workspaces.UpdateByID(c.Context, w.ID, opts)
	return
}

// DeleteAllWorkspaceVariables delete the all the variables present on the workspace
func (c *Client) DeleteAllWorkspaceVariables(w *tfc.Workspace) (err error) {
	existingVariables, _, _, err := c.listVariables(w)

	for _, vars := range existingVariables {
		for _, v := range vars {
			if err = c.TFC.Variables.Delete(c.Context, w.ID, v.ID); err != nil {
				return err
			}
			log.Infof("deleted variable %s", v.Key)
		}
	}

	return
}

// DeleteWorkspaceVariables delete the variables list passed as an argument if they are present
// on the workspace
func (c *Client) DeleteWorkspaceVariables(w *tfc.Workspace, variables schemas.Variables) (err error) {
	existingVariables, _, _, err := c.listVariables(w)

	for _, v := range variables {
		kind := getCategoryType(v.Kind)
		if _, ok := existingVariables[kind]; ok {
			if _, ok := existingVariables[kind][v.Name]; ok {
				if err = c.TFC.Variables.Delete(c.Context, w.ID, existingVariables[kind][v.Name].ID); err != nil {
					return err
				}
				log.Infof("deleted variable %s", v.Name)
			}
		}
	}

	return
}

func (c *Client) updateSSHKey(w *tfc.Workspace, sshKeyName string) error {
	if sshKeyName == "-" {
		log.Infof("Removing currently configured SSH key")
		_, err := c.TFC.Workspaces.UnassignSSHKey(c.Context, w.ID)
		return err
	}

	sshKeys, err := c.TFC.SSHKeys.List(c.Context, w.Organization.Name, tfc.SSHKeyListOptions{})
	if err != nil {
		return err
	}

	for _, sshKey := range sshKeys.Items {
		if sshKey.Name == sshKeyName {
			log.Infof("Updating configured SSH key to '%s'", sshKey.Name)
			_, err := c.TFC.Workspaces.AssignSSHKey(c.Context, w.ID, tfc.WorkspaceAssignSSHKeyOptions{
				SSHKeyID: &sshKey.ID,
			})
			return err
		}
	}

	return fmt.Errorf("could not find ssh key '%s'", sshKeyName)
}

func (c *Client) shouldUpdateSSHKey(w *tfc.Workspace, sshKeyName string) (bool, error) {
	var sshKey *tfc.SSHKey
	var err error
	if sshKeyName == "-" && w.SSHKey != nil {
		log.Infof("Workspace ssh key should not be configured, we will remove it")
		return true, nil
	}

	if w.SSHKey == nil {
		log.Infof("Workspace ssh key not configured, wanted '%s', we will update", sshKeyName)
		return true, nil
	}

	sshKey, err = c.TFC.SSHKeys.Read(c.Context, w.SSHKey.ID)
	if err != nil {
		return false, fmt.Errorf("could not fetch ssh key from API: %v", err)
	}

	if sshKeyName != sshKey.Name {
		log.Infof("Workspace ssh key configured with '%s', wanted '%s', we will update", sshKey.Name, sshKeyName)
		return true, nil
	}

	return false, nil
}
