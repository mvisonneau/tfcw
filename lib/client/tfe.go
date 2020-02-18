package client

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/manifoldco/promptui"
	"github.com/mvisonneau/tfcw/lib/schemas"
	log "github.com/sirupsen/logrus"
)

type TFEVariables map[tfe.CategoryType]map[string]*tfe.Variable

type TFERunType string

const (
	TFERunTypePlan  TFERunType = "plan"
	TFERunTypeApply TFERunType = "apply"
)

func (c *Client) Run(cfg *schemas.Config, uploadPath string, t TFERunType) error {
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace)
	if err != nil {
		return fmt.Errorf("terraform cloud: %s", err)
	}

	log.Info("Preparing plan")
	log.Debugf("Workspace id for %s: %s", w.Name, w.ID)
	log.Debugf("Configured working directory: %s", w.WorkingDirectory)

	if len(w.WorkingDirectory) > 0 {
		absolutePath, err := filepath.Abs(uploadPath)
		if err != nil {
			return fmt.Errorf("unable to find absolute path for terraform configuration folder %s", err.Error())
		}
		uploadPath = strings.Replace(absolutePath, w.WorkingDirectory, "", 1)
		log.Debugf("Upload path set to ", uploadPath)
	}

	log.Debug("Creating configuration version..")
	configVersion, err := c.TFE.ConfigurationVersions.Create(c.Context, w.ID, tfe.ConfigurationVersionCreateOptions{
		AutoQueueRuns: tfe.Bool(false),
	})
	log.Debugf("Configuration version ID: %s", configVersion.ID)

	if err != nil {
		return fmt.Errorf("terraform cloud: %s", err)
	}

	log.Debug("Uploading configuration version..")
	err = c.TFE.ConfigurationVersions.Upload(c.Context, configVersion.UploadURL, uploadPath)
	if err != nil {
		return fmt.Errorf("terraform cloud: %s", err)
	}
	log.Debug("Uploaded configuration version!")

	run, err := c.TFE.Runs.Create(c.Context, tfe.RunCreateOptions{
		Message:              tfe.String("Triggered from TFCW"),
		ConfigurationVersion: configVersion,
		Workspace:            w,
	})

	if err != nil {
		return err
	}

	log.Info("Run ID: %s", run.ID)

	// Sometimes the plan ID is not immediately available
	for {
		if run.Plan != nil {
			log.Debugf("Plan ID: %s", run.Plan.ID)
			break
		}

		t := c.Backoff.Duration()
		log.Infof("Waiting %s for plan ID to be generated..", t.String())
		time.Sleep(t)

		run, err = c.TFE.Runs.Read(c.Context, run.ID)
		if err != nil {
			return nil
		}
	}

	plan, err := c.waitForTerraformPlan(run.Plan.ID)
	if err != nil {
		return err
	}

	if plan.HasChanges {
		if t == TFERunTypeApply && confirmApply() {
			log.Infof("Applying plan..")
			c.TFE.Runs.Apply(c.Context, run.ID, tfe.RunApplyOptions{
				Comment: tfe.String("Applied from TFCW"),
			})

			// Refresh run object to fetch the Apply.ID
			run, err = c.TFE.Runs.Read(c.Context, run.ID)
			if err != nil {
				return err
			}

			return c.waitForTerraformApply(run.Apply.ID)
		}

		// Discard the run
		log.Debugf("Discarding run ID: %s", run.ID)
		if err = c.TFE.Runs.Discard(c.Context, run.ID, tfe.RunDiscardOptions{}); err != nil {
			return err
		}
	}

	if t == TFERunTypeApply {
		return fmt.Errorf("not applying")
	}

	return nil
}

func (c *Client) getWorkspace(organization, workspace string) (*tfe.Workspace, error) {
	return c.TFE.Workspaces.Read(c.Context, organization, workspace)
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
			Value:     v.Value,
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
		Value:     v.Value,
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
	time.Sleep(2 * time.Second)
	apply, err := c.TFE.Applies.Read(c.Context, applyID)
	c.Backoff.Reset()

wait:
	for {
		apply, err := c.TFE.Applies.Read(c.Context, applyID)
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
			} else {
				return err
			}
		}
		fmt.Print(line)
	}
	return nil
}

func confirmApply() bool {
	prompt := promptui.Prompt{
		Label:     "Apply",
		IsConfirm: true,
	}

	if _, err := prompt.Run(); err != nil {
		return false
	}

	return true
}
