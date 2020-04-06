package client

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/jpillora/backoff"
	"github.com/manifoldco/promptui"
	"github.com/mvisonneau/tfcw/lib/schemas"
	log "github.com/sirupsen/logrus"
)

// TFCRunType defines possible TFC run types
type TFCRunType string

const (
	// TFCRunTypePlan refers to a TFC `plan`
	TFCRunTypePlan TFCRunType = "plan"

	// TFCRunTypeApply refers to a TFC `apply`
	TFCRunTypeApply TFCRunType = "apply"
)

// TFCCreateRunOptions handles configuration variables for creating a new run on TFE
type TFCCreateRunOptions struct {
	AutoApprove  bool
	AutoDiscard  bool
	NoPrompt     bool
	OutputPath   string
	Message      string
	StartTimeout time.Duration
}

// CreateRun triggers a `run` over the TFC API
func (c *Client) CreateRun(cfg *schemas.Config, opts *TFCCreateRunOptions) error {
	log.Info("Preparing plan")
	w, err := c.getWorkspace(cfg.Runtime.TFC.Organization, cfg.Runtime.TFC.Workspace)
	if err != nil {
		return err
	}

	configVersion, err := c.createConfigurationVersion(w)
	if err != nil {
		return err
	}

	if err := c.uploadConfigurationVersion(w, configVersion, cfg.Runtime.WorkingDir); err != nil {
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

	plan, err := c.waitForTerraformPlan(planID, opts.StartTimeout)
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
	c.TFC.Runs.Apply(c.Context, runID, tfc.RunApplyOptions{
		Comment: &message,
	})

	// Refresh run object to fetch the Apply.ID
	run, err := c.TFC.Runs.Read(c.Context, runID)
	if err != nil {
		return err
	}

	return c.waitForTerraformApply(run.Apply.ID)
}

// DiscardRun given its ID
func (c *Client) DiscardRun(runID, message string) error {
	log.Infof("Discarding run ID: %s", runID)
	return c.TFC.Runs.Discard(c.Context, runID, tfc.RunDiscardOptions{
		Comment: &message,
	})
}

func (c *Client) createConfigurationVersion(w *tfc.Workspace) (*tfc.ConfigurationVersion, error) {
	log.Debug("Creating configuration version")
	configVersion, err := c.TFC.ConfigurationVersions.Create(c.Context, w.ID, tfc.ConfigurationVersionCreateOptions{
		AutoQueueRuns: tfc.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating TFC configuration version: %s", err)
	}

	log.Debugf("Configuration version ID: %s", configVersion.ID)
	return configVersion, nil
}

func (c *Client) uploadConfigurationVersion(w *tfc.Workspace, configVersion *tfc.ConfigurationVersion, uploadPath string) error {
	if len(w.WorkingDirectory) > 0 {
		absolutePath, err := filepath.Abs(uploadPath)
		if err != nil {
			return fmt.Errorf("unable to find absolute path for terraform configuration folder %s", err.Error())
		}
		uploadPath = strings.Replace(absolutePath, w.WorkingDirectory, "", 1)
		log.Debugf("Upload path set to %s", uploadPath)
	}

	log.Debug("Uploading configuration version..")
	if err := c.TFC.ConfigurationVersions.Upload(c.Context, configVersion.UploadURL, uploadPath); err != nil {
		return fmt.Errorf("error uploading configuration version: %s", err)
	}
	log.Debug("Uploaded configuration version!")
	return nil
}

func (c *Client) createRun(w *tfc.Workspace, configVersion *tfc.ConfigurationVersion, message string) (*tfc.Run, error) {
	log.Debugf("Creating run for workspace '%s' / configuration version '%s'", w.ID, configVersion.ID)
	run, err := c.TFC.Runs.Create(c.Context, tfc.RunCreateOptions{
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

func (c *Client) setVariableOnTFC(cfg *schemas.Config, w *tfc.Workspace, v *schemas.VariableValue, e TFCVariables) (*tfc.Variable, error) {
	if v.Variable.Sensitive == nil {
		if cfg.Defaults.Variable.Sensitive == nil {
			v.Variable.Sensitive = tfc.Bool(true)
		} else {
			v.Variable.Sensitive = cfg.Defaults.Variable.Sensitive
		}
	}

	if v.Variable.HCL == nil {
		if cfg.Defaults.Variable.Sensitive == nil {
			v.Variable.HCL = tfc.Bool(false)
		} else {
			v.Variable.HCL = cfg.Defaults.Variable.HCL
		}
	}

	if existingVariable, ok := e[getCategoryType(v.Variable.Kind)][v.Name]; ok {
		updatedVariable, err := c.TFC.Variables.Update(c.Context, w.ID, existingVariable.ID, tfc.VariableUpdateOptions{
			Key:       &v.Name,
			Value:     &v.Value,
			Sensitive: v.Variable.Sensitive,
			HCL:       v.Variable.HCL,
		})

		// In case we cannot update the fields, we delete the variable and recreate it
		if err != nil {
			log.Debugf("Could not update variable id %s, attempting to recreate it.", existingVariable.ID)
			err = c.TFC.Variables.Delete(c.Context, w.ID, existingVariable.ID)
			if err != nil {
				return nil, err
			}
		} else {
			return updatedVariable, nil
		}
	}

	return c.TFC.Variables.Create(c.Context, w.ID, tfc.VariableCreateOptions{
		Key:       &v.Name,
		Value:     &v.Value,
		Category:  tfc.Category(getCategoryType(v.Variable.Kind)),
		Sensitive: v.Variable.Sensitive,
		HCL:       v.Variable.HCL,
	})
}

func (c *Client) purgeUnmanagedVariables(vars schemas.VariableValues, e TFCVariables, dryRun bool) error {
	for _, v := range vars {
		if _, ok := e[getCategoryType(v.Variable.Kind)][v.Name]; ok {
			delete(e[getCategoryType(v.Variable.Kind)], v.Name)
		}
	}

	for _, tfeVars := range e {
		for _, v := range tfeVars {
			if !dryRun {
				log.Warnf("Deleting unmanaged variable %s (%s)", v.Key, v.Category)
				err := c.TFC.Variables.Delete(c.Context, v.Workspace.ID, v.ID)
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

func (c *Client) listVariables(w *tfc.Workspace) (TFCVariables, error) {
	variables := TFCVariables{}

	listOptions := tfc.VariableListOptions{
		ListOptions: tfc.ListOptions{
			PageNumber: 1,
			PageSize:   20,
		},
	}

	for {
		list, err := c.TFC.Variables.List(c.Context, w.ID, listOptions)
		if err != nil {
			return variables, fmt.Errorf("Unable to list variables from the Terraform Cloud API : %v", err.Error())
		}

		for _, v := range list.Items {
			if _, ok := variables[v.Category]; !ok {
				variables[v.Category] = map[string]*tfc.Variable{}
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

func getCategoryType(kind schemas.VariableKind) tfc.CategoryType {
	switch kind {
	case schemas.VariableKindEnvironment:
		return tfc.CategoryEnv
	case schemas.VariableKindTerraform:
		return tfc.CategoryTerraform
	}

	return tfc.CategoryType("")
}

func (c *Client) getTerraformPlanID(run *tfc.Run) (string, error) {
	var err error

	// Sometimes the plan ID is not immediately available when the run is created
	for {
		if run.Plan != nil {
			break
		}

		t := c.Backoff.Duration()
		log.Infof("Waiting %s for plan ID to be generated..", t.String())
		time.Sleep(t)

		run, err = c.TFC.Runs.Read(c.Context, run.ID)
		if err != nil {
			return "", err
		}
	}

	log.Debugf("Plan ID: %s", run.Plan.ID)
	return run.Plan.ID, nil
}

func (c *Client) waitForTerraformPlan(planID string, startTimeout time.Duration) (plan *tfc.Plan, err error) {
	time.Sleep(2 * time.Second)
	plan, err = c.TFC.Plans.Read(c.Context, planID)
	c.Backoff.Reset()

wait:
	for {
		plan, err = c.TFC.Plans.Read(c.Context, planID)
		if err != nil {
			return
		}

		switch plan.Status {
		case tfc.PlanCanceled:
			return plan, fmt.Errorf("plan has been cancelled")
		case tfc.PlanErrored:
			break wait
		case tfc.PlanFinished:
			break wait
		case tfc.PlanRunning:
			break wait
		case tfc.PlanUnreachable:
			return plan, fmt.Errorf("plan is unreachable from TFC API")
		default:
			t := c.Backoff.Duration()
			if timeoutExhausted(c.Backoff, startTimeout) {
				return nil, fmt.Errorf("timed out waiting for the plan to start, exiting now")
			}
			log.Infof("Waiting for plan to start, current status: %s, sleeping for %s", plan.Status, t.String())
			time.Sleep(t)
		}
	}

	logs, err := c.TFC.Plans.Logs(c.Context, planID)
	if err != nil {
		return
	}

	if err = readTerraformLogs(logs); err != nil {
		return
	}

	plan, err = c.TFC.Plans.Read(c.Context, planID)
	if err != nil {
		return
	}

	if plan.Status != tfc.PlanFinished {
		return plan, fmt.Errorf("plan status: %s", plan.Status)
	}

	return
}

func (c *Client) waitForTerraformApply(applyID string) error {
	var apply *tfc.Apply
	var err error

	// Reset the backoff in case it got incremented somewhere else beforehand
	c.Backoff.Reset()

	// Sleep for a second on init
	time.Sleep(time.Second)

wait:
	for {
		apply, err = c.TFC.Applies.Read(c.Context, applyID)
		if err != nil {
			return err
		}

		switch apply.Status {
		case tfc.ApplyCanceled:
			return fmt.Errorf("apply has been cancelled")
		case tfc.ApplyErrored:
			break wait
		case tfc.ApplyFinished:
			break wait
		case tfc.ApplyRunning:
			break wait
		case tfc.ApplyUnreachable:
			return fmt.Errorf("apply is unreachable from TFC API")
		default:
			t := c.Backoff.Duration()
			log.Infof("Waiting for apply to start, current status: %s, sleeping for %s", apply.Status, t.String())
			time.Sleep(t)
		}
	}

	logs, err := c.TFC.Applies.Logs(c.Context, applyID)
	if err != nil {
		return err
	}

	if err = readTerraformLogs(logs); err != nil {
		return err
	}

	apply, err = c.TFC.Applies.Read(c.Context, applyID)
	if err != nil {
		return err
	}

	if apply.Status != tfc.ApplyFinished {
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

func timeoutExhausted(b *backoff.Backoff, t time.Duration) bool {
	if t == 0 {
		return false
	}

	var totalDuration time.Duration
	a := float64(0)
	for a < b.Attempt() {
		totalDuration += b.ForAttempt(a)
		if totalDuration >= t {
			return true
		}
		a++
	}
	return false
}
