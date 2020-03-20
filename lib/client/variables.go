package client

import (
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/go-tfe"
	"github.com/mvisonneau/tfcw/lib/schemas"

	log "github.com/sirupsen/logrus"
)

// TFEVariables gives us an accessible fashion for managing all our
// variables independently of their kind
type TFEVariables map[tfe.CategoryType]map[string]*tfe.Variable

// RenderVariablesType defines possible rendering methods
type RenderVariablesType string

const (
	// RenderVariablesTypeTFC refers to a Terraform Cloud rendering
	RenderVariablesTypeTFC RenderVariablesType = "tfc"

	// RenderVariablesTypeLocal refers to a local rendering
	RenderVariablesTypeLocal RenderVariablesType = "local"
)

// RenderVariables issues a rendering of all variables defined in a schemas.Config object
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

	variableValues := schemas.VariableValues{}
	values := make(chan *schemas.VariableValue)
	errors := make(chan error)
	wg := sync.WaitGroup{}

	for _, v := range vars {
		wg.Add(1)
		go func(v *schemas.Variable) {
			defer wg.Done()
			fetchedValues, err := c.fetchVariableValues(v)
			errors <- err
			for _, value := range fetchedValues {
				wg.Add(1)
				values <- value
			}
		}(v)
	}

	go func() {
		for value := range values {
			variableValues = append(variableValues, value)
			errors <- c.renderVariableOnTFC(w, value, e, dryRun)
			wg.Done()
		}
	}()

	go func() {
		wg.Wait()
		close(values)
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return err
		}
	}

	if cfg.TFC.PurgeUnmanagedVariables != nil && *cfg.TFC.PurgeUnmanagedVariables {
		log.Debugf("Looking for unmanaged variables to remove")
		return c.purgeUnmanagedVariables(variableValues, e, dryRun)
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

	variableValues := schemas.VariableValues{}
	values := make(chan *schemas.VariableValue)
	errors := make(chan error)
	wg := sync.WaitGroup{}

	for _, v := range vars {
		wg.Add(1)
		go func(v *schemas.Variable) {
			defer wg.Done()
			fetchedValues, err := c.fetchVariableValues(v)
			errors <- err
			for _, value := range fetchedValues {
				wg.Add(1)
				values <- value
			}
		}(v)
	}

	go func() {
		for value := range values {
			variableValues = append(variableValues, value)
			errors <- c.renderVariableLocally(value, envFile, tfFile)
			wg.Done()
		}
	}()

	go func() {
		wg.Wait()
		close(values)
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) renderVariableOnTFC(w *tfe.Workspace, v *schemas.VariableValue, e TFEVariables, dryRun bool) error {
	if !dryRun {
		if _, err := c.setVariableOnTFC(w, v, e); err != nil {
			return err
		}
	}

	logVariableValue(v, dryRun)
	return nil
}

func (c *Client) renderVariableLocally(v *schemas.VariableValue, envFile, tfFile *os.File) error {
	switch v.Variable.Kind {
	case schemas.VariableKindEnvironment:
		if _, err := envFile.WriteString(fmt.Sprintf("export %s=%s\n", v.Name, v.Value)); err != nil {
			return err
		}
	case schemas.VariableKindTerraform:
		s := ""
		if v.Variable.HCL != nil && *v.Variable.HCL {
			s = fmt.Sprintf("%s = %s\n", v.Name, v.Value)
		} else {
			s = fmt.Sprintf("%s = \"%s\"\n", v.Name, v.Value)
		}

		if _, err := tfFile.WriteString(s); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown kind '%s' for variable %s", v.Variable.Kind, v.Name)
	}

	logVariableValue(v, false)
	return nil
}

func (c *Client) fetchVariableValues(v *schemas.Variable) (schemas.VariableValues, error) {
	if c.isVariableAlreadyProcessed(v.Name, v.Kind) {
		return nil, fmt.Errorf("duplicate variable '%s' (%s)", v.Name, v.Kind)
	}

	provider, err := v.GetProvider()
	if err != nil {
		return nil, err
	}

	switch *provider {
	case schemas.VariableProviderEnv:
		return schemas.VariableValues{
			&schemas.VariableValue{
				Variable: v,
				Name:     v.Name,
				Value:    c.Env.GetValue(v.Env),
			},
		}, nil
	case schemas.VariableProviderS5:
		value, err := c.S5.GetValue(v.S5)
		if err != nil {
			return nil, fmt.Errorf("s5 error: %s", err)
		}

		return schemas.VariableValues{
			&schemas.VariableValue{
				Variable: v,
				Name:     v.Name,
				Value:    value,
			},
		}, nil
	case schemas.VariableProviderVault:
		return c.getVaultValues(v)
	}

	return nil, fmt.Errorf("unknown provider '%s' for variable '%s'", *provider, v.Name)
}

// getVaultValues will return an empty value if multiple keys are set
func (c *Client) getVaultValues(v *schemas.Variable) (schemas.VariableValues, error) {
	values, err := c.Vault.GetValues(v.Vault)
	if err != nil {
		return nil, fmt.Errorf("error getting values from vault for variable '%s' : %s", v.Name, err)
	}

	// We can map several keys in a single API call
	if (v.Vault.Key == nil && (v.Vault.Keys == nil || len(*v.Vault.Keys) == 0)) ||
		(v.Vault.Key != nil && v.Vault.Keys != nil && len(*v.Vault.Keys) > 0) {
		return nil, fmt.Errorf("you either need to set 'key' or 'keys' when using the Vault provider")
	}

	if v.Vault.Key != nil {
		if value, found := values[*v.Vault.Key]; found {
			return schemas.VariableValues{
				&schemas.VariableValue{
					Variable: v,
					Name:     v.Name,
					Value:    value,
				},
			}, nil
		}
		return nil, fmt.Errorf("key '%s' was not found in secret '%s'", *v.Vault.Key, *v.Vault.Path)
	}

	variableValues := schemas.VariableValues{}
	for vaultKey, variableName := range *v.Vault.Keys {
		if value, found := values[vaultKey]; found {
			variableValues = append(variableValues, &schemas.VariableValue{
				Variable: v,
				Name:     variableName,
				Value:    value,
			})
			continue
		}
		return nil, fmt.Errorf("key '%s' was not found in secret '%s'", vaultKey, *v.Vault.Path)
	}

	return variableValues, nil
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

func logVariableValue(v *schemas.VariableValue, dryRun bool) {
	if dryRun {
		log.Infof("[DRY-RUN] Set variable '%s' (%s) : %s", v.Name, v.Variable.Kind, secureSensitiveString(v.Value))
	} else {
		log.Infof("Set variable '%s' (%s)", v.Name, v.Variable.Kind)
	}
}

func secureSensitiveString(sensitive string) string {
	if len(sensitive) < 4 {
		return "**********"
	}
	return fmt.Sprintf("%s********%s", string(sensitive[0]), string(sensitive[len(sensitive)-1]))
}
