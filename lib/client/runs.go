package client

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/manifoldco/promptui"
	"github.com/mvisonneau/tfcw/lib/schemas"
	log "github.com/sirupsen/logrus"
)

// TFERunType defines possible TFC run types
type TFERunType string

const (
	// TFERunTypePlan refers to a TFC `plan`
	TFERunTypePlan TFERunType = "plan"

	// TFERunTypeApply refers to a TFC `apply`
	TFERunTypeApply TFERunType = "apply"
)

// TFECreateRunOptions handles configuration variables for creating a new run on TFE
type TFECreateRunOptions struct {
	AutoApprove bool
	AutoDiscard bool
	NoPrompt    bool
	ConfigPath  string
	OutputPath  string
	Message     string
}

// CreateRun triggers a `run` over the TFC API
func (c *Client) CreateRun(cfg *schemas.Config, opts *TFECreateRunOptions) error {
	log.Info("Preparing plan")
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace)
	if err != nil {
		return err
	}

	configVersion, err := c.createConfigurationVersion(w)
	if err != nil {
		return err
	}

	if err := c.uploadConfigurationVersion(w, configVersion, opts.ConfigPath); err != nil {
		return err
	}

	run, err := c.createRun(w, configVersion, opts.Message)
	if err != nil {
		return err
	}

	if len(opts.OutputPath) > 0 {
		log.Debugf("saving run ID on disk at '%s'", opts.OutputPath)
		if err = ioutil.WriteFile(opts.OutputPath, []byte(run.ID), 0644); err != nil {
			c.DiscardRun(run.ID, opts.Message)
			return err
		}
	}

	planID, err := c.getTerraformPlanID(run)
	if err != nil {
		c.DiscardRun(run.ID, opts.Message)
		return err
	}

	plan, err := c.waitForTerraformPlan(planID)
	if err != nil {
		c.DiscardRun(run.ID, opts.Message)
		return err
	}

	if plan.HasChanges {
		if opts.AutoDiscard {
			return c.DiscardRun(run.ID, opts.Message)
		}

		if opts.AutoApprove {
			return c.ApproveRun(run.ID, opts.Message)
		}

		if opts.NoPrompt {
			return nil
		}

		if promptApproveRun() {
			return c.ApproveRun(run.ID, opts.Message)
		}

		return c.DiscardRun(run.ID, opts.Message)
	}

	return nil
}

// ApproveRun given its ID
func (c *Client) ApproveRun(runID, message string) error {
	log.Infof("Approving run ID: %s", runID)
	c.TFE.Runs.Apply(c.Context, runID, tfe.RunApplyOptions{
		Comment: &message,
	})

	// Refresh run object to fetch the Apply.ID
	run, err := c.TFE.Runs.Read(c.Context, runID)
	if err != nil {
		return err
	}

	return c.waitForTerraformApply(run.Apply.ID)
}

// DiscardRun given its ID
func (c *Client) DiscardRun(runID, message string) error {
	log.Infof("Discarding run ID: %s", runID)
	return c.TFE.Runs.Discard(c.Context, runID, tfe.RunDiscardOptions{
		Comment: &message,
	})
}

func (c *Client) getWorkspace(organization, workspace string) (*tfe.Workspace, error) {
	log.Debug("Fetching workspace")
	w, err := c.TFE.Workspaces.Read(c.Context, organization, workspace)
	if err != nil {
		return nil, fmt.Errorf("error fetching TFC workspace: %s", err)
	}

	log.Debugf("Found workspace id for '%s': %s", w.Name, w.ID)
	log.Debugf("Configured working directory: %s", w.WorkingDirectory)

	return w, nil
}

func (c *Client) createConfigurationVersion(w *tfe.Workspace) (*tfe.ConfigurationVersion, error) {
	log.Debug("Creating configuration version")
	configVersion, err := c.TFE.ConfigurationVersions.Create(c.Context, w.ID, tfe.ConfigurationVersionCreateOptions{
		AutoQueueRuns: tfe.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating TFC configuration version: %s", err)
	}

	log.Debugf("Configuration version ID: %s", configVersion.ID)
	return configVersion, nil
}

func (c *Client) uploadConfigurationVersion(w *tfe.Workspace, configVersion *tfe.ConfigurationVersion, uploadPath string) error {
	if len(w.WorkingDirectory) > 0 {
		absolutePath, err := filepath.Abs(uploadPath)
		if err != nil {
			return fmt.Errorf("unable to find absolute path for terraform configuration folder %s", err.Error())
		}
		uploadPath = strings.Replace(absolutePath, w.WorkingDirectory, "", 1)
		log.Debugf("Upload path set to %s", uploadPath)
	}

	log.Debug("Uploading configuration version..")
	if err := c.TFE.ConfigurationVersions.Upload(c.Context, configVersion.UploadURL, uploadPath); err != nil {
		return fmt.Errorf("error uploading configuration version: %s", err)
	}
	log.Debug("Uploaded configuration version!")
	return nil
}

func (c *Client) createRun(w *tfe.Workspace, configVersion *tfe.ConfigurationVersion, message string) (*tfe.Run, error) {
	log.Debugf("Creating run for workspace '%s' / configuration version '%s'", w.ID, configVersion.ID)
	run, err := c.TFE.Runs.Create(c.Context, tfe.RunCreateOptions{
		Message:              &message,
		ConfigurationVersion: configVersion,
		Workspace:            w,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating run: %s", err)
	}

	log.Debugf("Run ID: %s", run.ID)
	return run, nil
}

func (c *Client) setVariableOnTFC(w *tfe.Workspace, v *schemas.Variable, e TFEVariables) (*tfe.Variable, error) {
	if v.Sensitive == nil {
		v.Sensitive = tfe.Bool(true)
	}

	if v.HCL == nil {
		v.HCL = tfe.Bool(false)
	}

	if existingVariable, ok := e[getCategoryType(v.Kind)][v.Name]; ok {
		updatedVariable, err := c.TFE.Variables.Update(c.Context, w.ID, existingVariable.ID, tfe.VariableUpdateOptions{
			Key:       &v.Name,
			Value:     &v.Value,
			Sensitive: v.Sensitive,
			HCL:       v.HCL,
		})

		// In case we cannot update the fields, we delete the variable and recreate it
		if err != nil {
			log.Debugf("Could not update variable id %s, attempting to recreate it.", existingVariable.ID)
			err = c.TFE.Variables.Delete(c.Context, w.ID, existingVariable.ID)
			if err != nil {
				return nil, err
			}
		} else {
			return updatedVariable, nil
		}
	}

	return c.TFE.Variables.Create(c.Context, w.ID, tfe.VariableCreateOptions{
		Key:       &v.Name,
		Value:     &v.Value,
		Category:  tfe.Category(getCategoryType(v.Kind)),
		Sensitive: v.Sensitive,
		HCL:       v.HCL,
	})
}

func (c *Client) purgeUnmanagedVariables(vars schemas.Variables, e TFEVariables, dryRun bool) error {
	for _, v := range vars {
		if _, ok := e[getCategoryType(v.Kind)][v.Name]; ok {
			delete(e[getCategoryType(v.Kind)], v.Name)
		}
	}

	for _, vars := range e {
		for _, v := range vars {
			if !dryRun {
				log.Warnf("Deleting unmanaged variable %s (%s)", v.Key, v.Category)
				err := c.TFE.Variables.Delete(c.Context, v.Workspace.ID, v.ID)
				if err != nil {
					return fmt.Errorf("error deleting variable %s (%s) on TFC: %s", v.Key, v.Category, err.Error())
				}
			} else {
				log.Warnf("[DRY-RUN] Deleting unmanaged variable %s (%s)", v.Key, v.Category)
			}
		}
	}

	return nil
}

func (c *Client) listVariables(w *tfe.Workspace) (TFEVariables, error) {
	variables := TFEVariables{}

	listOptions := tfe.VariableListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 1,
			PageSize:   20,
		},
	}

	for {
		list, err := c.TFE.Variables.List(c.Context, w.ID, listOptions)
		if err != nil {
			return variables, fmt.Errorf("Unable to list variables from the Terraform Cloud API : %v", err.Error())
		}

		for _, v := range list.Items {
			if _, ok := variables[v.Category]; !ok {
				variables[v.Category] = map[string]*tfe.Variable{}
			}
			variables[v.Category][v.Key] = v
		}

		if list.Pagination.CurrentPage >= list.Pagination.TotalPages {
			break
		}

		listOptions.PageNumber = list.Pagination.NextPage
	}
	return variables, nil
}

func getCategoryType(kind schemas.VariableKind) tfe.CategoryType {
	switch kind {
	case schemas.VariableKindEnvironment:
		return tfe.CategoryEnv
	case schemas.VariableKindTerraform:
		return tfe.CategoryTerraform
	}

	return tfe.CategoryType("")
}

func (c *Client) getTerraformPlanID(run *tfe.Run) (string, error) {
	var err error

	// Sometimes the plan ID is not immediately available when the run is created
	for {
		if run.Plan != nil {
			break
		}

		t := c.Backoff.Duration()
		log.Infof("Waiting %s for plan ID to be generated..", t.String())
		time.Sleep(t)

		run, err = c.TFE.Runs.Read(c.Context, run.ID)
		if err != nil {
			return "", err
		}
	}

	log.Debugf("Plan ID: %s", run.Plan.ID)
	return run.Plan.ID, nil
}

func (c *Client) waitForTerraformPlan(planID string) (plan *tfe.Plan, err error) {
	time.Sleep(2 * time.Second)
	plan, err = c.TFE.Plans.Read(c.Context, planID)
	c.Backoff.Reset()

wait:
	for {
		plan, err = c.TFE.Plans.Read(c.Context, planID)
		if err != nil {
			return
		}

		switch plan.Status {
		case tfe.PlanCanceled:
			return plan, fmt.Errorf("plan has been cancelled")
		case tfe.PlanErrored:
			break wait
		case tfe.PlanFinished:
			break wait
		case tfe.PlanRunning:
			break wait
		case tfe.PlanUnreachable:
			return plan, fmt.Errorf("plan is unreachable from TFC API")
		default:
			t := c.Backoff.Duration()
			log.Infof("Waiting for plan to start, current status: %s, sleeping for %s", plan.Status, t.String())
			time.Sleep(t)
		}
	}

	logs, err := c.TFE.Plans.Logs(c.Context, planID)
	if err != nil {
		return
	}

	if err = readTerraformLogs(logs); err != nil {
		return
	}

	plan, err = c.TFE.Plans.Read(c.Context, planID)
	if err != nil {
		return
	}

	if plan.Status != tfe.PlanFinished {
		return plan, fmt.Errorf("plan status: %s", plan.Status)
	}

	return
}

func (c *Client) waitForTerraformApply(applyID string) error {
	var apply *tfe.Apply
	var err error

	// Reset the backoff in case it got incremented somewhere else beforehand
	c.Backoff.Reset()

	// Sleep for a second on init
	time.Sleep(time.Second)

wait:
	for {
		apply, err = c.TFE.Applies.Read(c.Context, applyID)
		if err != nil {
			return err
		}

		switch apply.Status {
		case tfe.ApplyCanceled:
			return fmt.Errorf("apply has been cancelled")
		case tfe.ApplyErrored:
			break wait
		case tfe.ApplyFinished:
			break wait
		case tfe.ApplyRunning:
			break wait
		case tfe.ApplyUnreachable:
			return fmt.Errorf("apply is unreachable from TFC API")
		default:
			t := c.Backoff.Duration()
			log.Infof("Waiting for apply to start, current status: %s, sleeping for %s", apply.Status, t.String())
			time.Sleep(t)
		}
	}

	logs, err := c.TFE.Applies.Logs(c.Context, applyID)
	if err != nil {
		return err
	}

	if err = readTerraformLogs(logs); err != nil {
		return err
	}

	apply, err = c.TFE.Applies.Read(c.Context, applyID)
	if err != nil {
		return err
	}

	if apply.Status != tfe.ApplyFinished {
		return fmt.Errorf("apply status: %s", apply.Status)
	}

	return nil
}

func readTerraformLogs(l io.Reader) error {
	r := bufio.NewReader(l)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Print(line)
	}
	return nil
}

func promptApproveRun() bool {
	prompt := promptui.Prompt{
		Label:     "Apply",
		IsConfirm: true,
	}

	if _, err := prompt.Run(); err != nil {
		return false
	}

	return true
}

func saveRunID(runID, outputFile string) {
	if len(outputFile) > 0 {
		log.Debugf("Saving run ID '%s' onto file %s.", runID, outputFile)

	} else {
		log.Debugf("Output file not defined, not saving run ID on disk.")
	}
}
