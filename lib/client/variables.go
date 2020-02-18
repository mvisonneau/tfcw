package client

import (
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/go-tfe"
	"github.com/mvisonneau/tfcw/lib/schemas"

	log "github.com/sirupsen/logrus"
)

type RenderVariablesType string

const (
	RenderVariablesTypeTFC   RenderVariablesType = "tfc"
	RenderVariablesTypeLocal RenderVariablesType = "local"
)

func (c *Client) RenderVariables(cfg *schemas.Config, t RenderVariablesType, dryRun bool) error {
	variables := schemas.Variables{}

	for _, variable := range cfg.TerraformVariables {
		variable.Kind = schemas.VariableKindTerraform
		variables = append(variables, variable)
	}

	for _, variable := range cfg.EnvironmentVariables {
		variable.Kind = schemas.VariableKindEnvironment
		variables = append(variables, variable)
	}

	switch t {
	case RenderVariablesTypeTFC:
		log.Info("Processing variables and updating their values on TFC")
		return c.renderVariablesOnTFC(cfg, variables, dryRun)
	case RenderVariablesTypeLocal:
		log.Info("Processing variables and updating their values locally")
		return c.renderVariablesLocally(variables)
	default:
		return fmt.Errorf("undefined ProcessVaribleType '%s'", t)
	}
}

func (c *Client) renderVariablesOnTFC(cfg *schemas.Config, vars schemas.Variables, dryRun bool) error {
	w, err := c.getWorkspace(cfg.TFC.Organization, cfg.TFC.Workspace)
	if err != nil {
		return fmt.Errorf("terraform cloud: %s", err)
	}
	log.Debugf("workspace id for %s: %s", w.Name, w.ID)

	// Find existing variables on TFC
	e, err := c.listVariables(w)
	if err != nil {
		return fmt.Errorf("terraform cloud: %s", err)
	}

	ch := make(chan error)
	wg := sync.WaitGroup{}

	for _, v := range vars {
		wg.Add(1)
		go func(v *schemas.Variable) {
			defer wg.Done()
			ch <- c.renderVariableOnTFC(w, v, e, dryRun)
		}(v)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for err := range ch {
		if err != nil {
			return err
		}
	}

	if cfg.TFC.PurgeUnmanagedVariables != nil && *cfg.TFC.PurgeUnmanagedVariables {
		log.Debugf("Looking for unmanaged variables to remove")
		return c.purgeUnmanagedVariables(vars, e, dryRun)
	}

	return nil
}

func (c *Client) renderVariablesLocally(vars schemas.Variables) error {
	envFile, err := os.OpenFile("./tfcw.env", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer envFile.Close()

	tfFile, err := os.OpenFile("./tfcw.auth.tfvars", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer tfFile.Close()

	ch := make(chan error)
	wg := sync.WaitGroup{}

	for _, v := range vars {
		wg.Add(1)
		go func(v *schemas.Variable) {
			defer wg.Done()
			ch <- c.renderVariableLocally(v, envFile, tfFile)
		}(v)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for err := range ch {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) renderVariableOnTFC(w *tfe.Workspace, v *schemas.Variable, e TFEVariables, dryRun bool) error {
	c.fetchVariableValue(v)
	if !dryRun {
		if _, err := c.setVariableOnTFC(w, v, e); err != nil {
			return err
		}
	}

	logVariable(v, dryRun)
	return nil
}

func (c *Client) renderVariableLocally(v *schemas.Variable, envFile, tfFile *os.File) error {
	c.fetchVariableValue(v)
	switch v.Kind {
	case schemas.VariableKindEnvironment:
		if _, err := envFile.WriteString(fmt.Sprintf("export %s=%s\n", v.Name, *v.Value)); err != nil {
			return err
		}
	case schemas.VariableKindTerraform:
		s := ""
		if *v.HCL {
			s = fmt.Sprintf("%s = %s\n", v.Name, *v.Value)
		} else {
			s = fmt.Sprintf("%s = \"%s\"\n", v.Name, *v.Value)
		}

		if _, err := tfFile.WriteString(s); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unkown kind '%s' for variable %s", v.Kind, v.Name)
	}

	logVariable(v, false)
	return nil
}

func (c *Client) fetchVariableValue(v *schemas.Variable) error {
	if c.isVariableAlreadyProcessed(v.Name, v.Kind) {
		return fmt.Errorf("duplicate variable '%s' (%s)", v.Name, v.Kind)
	}

	configuredProviders := 0
	if v.Vault != nil && v.Vault.Path != nil {
		configuredProviders++
	}

	if v.S5 != nil {
		configuredProviders++
	}

	if v.Env != nil {
		configuredProviders++
	}

	if configuredProviders != 1 {
		return fmt.Errorf("you can't have more or less than one provider configured per variable. Found %d for '%s'", configuredProviders, v.Name)
	}

	// We can map several keys in a single API call
	if v.Vault != nil && v.Vault.Path != nil {
		if values, err := c.Vault.GetValues(v.Vault); err == nil {
			if v.Vault.Key == nil && (v.Vault.Keys == nil || len(*v.Vault.Keys) == 0) {
				return fmt.Errorf("you either need to set 'key' or 'keys' when using the Vault provider")
			}

			if v.Vault.Keys == nil {
				v.Vault.Keys = &map[string]string{}
			}

			if v.Vault.Keys == nil || len(*v.Vault.Keys) == 0 {
				(*v.Vault.Keys)[*v.Vault.Key] = v.Name
			}

			for vaultKey, variableName := range *v.Vault.Keys {
				if value, found := values[vaultKey]; found {
					v.Name = variableName
					v.Value = &value
					return nil
				}
				return fmt.Errorf("key '%s' was not found in secret '%s'", vaultKey, *v.Vault.Path)
			}
		}
	}

	if v.S5 != nil {
		value, err := c.S5.GetValue(v.S5)
		if err != nil {
			return fmt.Errorf("s5 error: %s", err)
		}

		v.Value = &value
		return nil
	}

	if v.Env != nil {
		value := c.Env.GetValue(v.Env)
		v.Value = &value
		return nil
	}

	return fmt.Errorf("no providers could be used to fetch the variable value")
}

func (c *Client) isVariableAlreadyProcessed(name string, kind schemas.VariableKind) bool {
	c.ProcessedVariablesMutex.Lock()
	defer c.ProcessedVariablesMutex.Unlock()
	k, exists := c.ProcessedVariables[name]

	if exists && k == kind {
		return true
	}

	c.ProcessedVariables[name] = kind
	return false
}

func logVariable(v *schemas.Variable, dryRun bool) error {
	if dryRun {
		log.Infof("[DRY-RUN] Set variable %s - (%s) : %s", v.Name, v.Kind, secureSensitiveString(*v.Value))
	} else {
		log.Infof("Set variable %s (%s)", v.Name, v.Kind)
	}
	return nil
}

func secureSensitiveString(sensitive string) string {
	if len(sensitive) < 4 {
		return "**********"
	}
	return fmt.Sprintf("%s********%s", string(sensitive[1]), string(sensitive[len(sensitive)-1]))
}
