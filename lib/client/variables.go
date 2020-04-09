package client

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	tfc "github.com/hashicorp/go-tfe"
	"github.com/mvisonneau/tfcw/lib/schemas"

	log "github.com/sirupsen/logrus"
)

// TFCVariables gives us an accessible fashion for managing all our
// variables independently of their kind
type TFCVariables map[tfc.CategoryType]map[string]*tfc.Variable

// RenderVariablesType defines possible rendering methods
type RenderVariablesType string

const (
	// RenderVariablesTypeTFC refers to a Terraform Cloud rendering
	RenderVariablesTypeTFC RenderVariablesType = "tfc"

	// RenderVariablesTypeLocal refers to a local rendering
	RenderVariablesTypeLocal RenderVariablesType = "local"

	// VariableExpirationsName is the name of the variable used for storing VariableExpirations in TFC
	VariableExpirationsName string = "__TFCW_VARIABLES_EXPIRATIONS"
)

// RenderVariablesOnTFC issues a rendering of all variables defined in a schemas.Config object on TFC
func (c *Client) RenderVariablesOnTFC(cfg *schemas.Config, w *tfc.Workspace, dryRun, forceUpdate bool) error {
	log.Info("Processing variables and updating their values on TFC")
	return c.renderVariablesOnTFC(cfg, w, dryRun, forceUpdate)
}

// RenderVariablesLocally issues a rendering of all variables defined in a schemas.Config object on TFC
func (c *Client) RenderVariablesLocally(cfg *schemas.Config) error {
	log.Info("Processing variables and updating their values locally")
	return c.renderVariablesLocally(cfg.GetVariables())
}

func (c *Client) setVariableOnTFC(cfg *schemas.Config, w *tfc.Workspace, v *schemas.VariableWithValue, e TFCVariables) (*tfc.Variable, error) {
	if v.Sensitive == nil {
		if cfg.Defaults == nil || cfg.Defaults.Variable == nil || cfg.Defaults.Variable.Sensitive == nil {
			v.Sensitive = tfc.Bool(true)
		} else {
			v.Sensitive = cfg.Defaults.Variable.Sensitive
		}
	}

	if v.HCL == nil {
		if cfg.Defaults == nil || cfg.Defaults.Variable == nil || cfg.Defaults.Variable.HCL == nil {
			v.HCL = tfc.Bool(false)
		} else {
			v.HCL = cfg.Defaults.Variable.HCL
		}
	}

	if existingVariable, ok := e[getCategoryType(v.Kind)][v.Name]; ok {
		updatedVariable, err := c.TFC.Variables.Update(c.Context, w.ID, existingVariable.ID, tfc.VariableUpdateOptions{
			Key:       &v.Name,
			Value:     &v.Value,
			Sensitive: v.Sensitive,
			HCL:       v.HCL,
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
		Category:  tfc.Category(getCategoryType(v.Kind)),
		Sensitive: v.Sensitive,
		HCL:       v.HCL,
	})
}

func (c *Client) purgeUnmanagedVariables(vars schemas.Variables, e TFCVariables, dryRun bool) error {
	for _, v := range vars {
		if _, ok := e[getCategoryType(v.Kind)][v.Name]; ok {
			delete(e[getCategoryType(v.Kind)], v.Name)
		}

		// Check if we can have inner values if we are using the Vault provider
		if p, _ := v.GetProvider(); *p == schemas.VariableProviderVault {
			if v.Vault.Keys != nil {
				for _, variableName := range *v.Vault.Keys {
					if _, ok := e[getCategoryType(v.Kind)][variableName]; ok {
						delete(e[getCategoryType(v.Kind)], variableName)
					}
				}
			}
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

func (c *Client) listVariables(w *tfc.Workspace) (variables TFCVariables, variableExpirations schemas.VariableExpirations, variablesExpirationsTFCVariableID string, err error) {
	variables = make(TFCVariables)

	listOptions := tfc.VariableListOptions{
		ListOptions: tfc.ListOptions{
			PageNumber: 1,
			PageSize:   20,
		},
	}

	for {
		var list *tfc.VariableList
		list, err = c.TFC.Variables.List(c.Context, w.ID, listOptions)
		if err != nil {
			err = fmt.Errorf("unable to list variables from the Terraform Cloud API : %s", err.Error())
			return
		}

		for _, v := range list.Items {
			if v.Key == VariableExpirationsName {
				variablesExpirationsTFCVariableID = v.ID
				if err = json.Unmarshal([]byte(v.Value), &variableExpirations); err != nil {
					err = fmt.Errorf("unable to parse the variable ttls currently set on TFC (%s) : %s", VariableExpirationsName, err.Error())
					// TODO: Remove the existing variable automatically?
					return
				}
				continue
			}

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

	return
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

func (c *Client) renderVariablesOnTFC(cfg *schemas.Config, w *tfc.Workspace, dryRun, forceUpdate bool) error {
	// Find existing variables on TFC
	existingVariables, variableExpirations, variableExpirationsTFCVariableID, err := c.listVariables(w)
	if err != nil {
		return fmt.Errorf("terraform cloud: %s", err)
	}

	variablesWithValues := schemas.VariablesWithValues{}
	values := make(chan *schemas.VariableWithValue)
	errors := make(chan error)
	wg := sync.WaitGroup{}

	variablesToUpdate := cfg.GetVariables()
	if !forceUpdate {
		variablesToUpdate, err = cfg.GetVariablesToUpdate(variableExpirations)
		if err != nil {
			return err
		}
	}

	for _, v := range variablesToUpdate {
		wg.Add(1)
		go func(v *schemas.Variable) {
			defer wg.Done()
			fetchedValues, err := c.fetchVariablesWithValues(v)
			errors <- err
			for _, value := range fetchedValues {
				wg.Add(1)
				values <- value
			}
		}(v)
	}

	go func() {
		for value := range values {
			variablesWithValues = append(variablesWithValues, value)
			errors <- c.renderVariableOnTFC(cfg, w, value, existingVariables, dryRun)
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

	// Update variable expirations on TFC
	newVariableExpirations, updateVariableExpirations, err := cfg.ComputeNewVariableExpirations(variablesToUpdate, variableExpirations)
	if err != nil {
		return err
	}

	if updateVariableExpirations {
		if err = c.updateVariableExpirations(w, newVariableExpirations, variableExpirationsTFCVariableID); err != nil {
			return err
		}
	}

	if cfg.TFC.PurgeUnmanagedVariables != nil && *cfg.TFC.PurgeUnmanagedVariables {
		log.Debugf("Looking for unmanaged variables to remove")
		return c.purgeUnmanagedVariables(cfg.GetVariables(), existingVariables, dryRun)
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

	variablesWithValues := schemas.VariablesWithValues{}
	values := make(chan *schemas.VariableWithValue)
	errors := make(chan error)
	wg := sync.WaitGroup{}

	for _, v := range vars {
		wg.Add(1)
		go func(v *schemas.Variable) {
			defer wg.Done()
			fetchedValues, err := c.fetchVariablesWithValues(v)
			errors <- err
			for _, value := range fetchedValues {
				wg.Add(1)
				values <- value
			}
		}(v)
	}

	go func() {
		for value := range values {
			variablesWithValues = append(variablesWithValues, value)
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

func (c *Client) renderVariableOnTFC(cfg *schemas.Config, w *tfc.Workspace, v *schemas.VariableWithValue, e TFCVariables, dryRun bool) (err error) {
	if !dryRun {
		if _, err = c.setVariableOnTFC(cfg, w, v, e); err != nil {
			return
		}
	}

	logVariableWithValue(v, dryRun)
	return
}

func (c *Client) renderVariableLocally(v *schemas.VariableWithValue, envFile, tfFile *os.File) error {
	switch v.Kind {
	case schemas.VariableKindEnvironment:
		if _, err := envFile.WriteString(fmt.Sprintf("export %s=%s\n", v.Name, v.Value)); err != nil {
			return err
		}
	case schemas.VariableKindTerraform:
		s := ""
		if v.HCL != nil && *v.HCL {
			s = fmt.Sprintf("%s = %s\n", v.Name, v.Value)
		} else {
			s = fmt.Sprintf("%s = \"%s\"\n", v.Name, v.Value)
		}

		if _, err := tfFile.WriteString(s); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown kind '%s' for variable %s", v.Kind, v.Name)
	}

	logVariableWithValue(v, false)
	return nil
}

func (c *Client) fetchVariablesWithValues(v *schemas.Variable) (schemas.VariablesWithValues, error) {
	if c.isVariableAlreadyProcessed(v.Name, v.Kind) {
		return nil, fmt.Errorf("duplicate variable '%s' (%s)", v.Name, v.Kind)
	}

	provider, err := v.GetProvider()
	if err != nil {
		return nil, err
	}

	switch *provider {
	case schemas.VariableProviderEnv:
		return schemas.VariablesWithValues{
			&schemas.VariableWithValue{
				Variable: *v,
				Value:    c.Env.GetValue(v.Env),
			},
		}, nil
	case schemas.VariableProviderS5:
		value, err := c.S5.GetValue(v.S5)
		if err != nil {
			return nil, fmt.Errorf("s5 error: %s", err)
		}

		return schemas.VariablesWithValues{
			&schemas.VariableWithValue{
				Variable: *v,
				Value:    value,
			},
		}, nil
	case schemas.VariableProviderVault:
		return c.getVaultValues(v)
	}

	return nil, fmt.Errorf("unknown provider '%s' for variable '%s'", *provider, v.Name)
}

// getVaultValues will return an empty value if multiple keys are set
func (c *Client) getVaultValues(v *schemas.Variable) (schemas.VariablesWithValues, error) {
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
			return schemas.VariablesWithValues{
				&schemas.VariableWithValue{
					Variable: *v,
					Value:    value,
				},
			}, nil
		}
		return nil, fmt.Errorf("key '%s' was not found in secret '%s'", *v.Vault.Key, *v.Vault.Path)
	}

	variablesWithValues := schemas.VariablesWithValues{}
	for vaultKey, variableName := range *v.Vault.Keys {
		if value, found := values[vaultKey]; found {
			vv := &schemas.VariableWithValue{
				Variable: *v,
				Value:    value,
			}
			vv.Name = variableName
			variablesWithValues = append(variablesWithValues, vv)
			continue
		}
		return nil, fmt.Errorf("key '%s' was not found in secret '%s'", vaultKey, *v.Vault.Path)
	}

	return variablesWithValues, nil
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

func logVariableWithValue(v *schemas.VariableWithValue, dryRun bool) {
	if dryRun {
		log.Infof("[DRY-RUN] Set variable '%s' (%s) : %s", v.Name, v.Kind, secureSensitiveString(v.Value))
	} else {
		log.Infof("Set variable '%s' (%s)", v.Name, v.Kind)
	}
}

func secureSensitiveString(sensitive string) string {
	if len(sensitive) < 4 {
		return "**********"
	}
	return fmt.Sprintf("%s********%s", string(sensitive[0]), string(sensitive[len(sensitive)-1]))
}

func (c *Client) updateVariableExpirations(w *tfc.Workspace, variableExpirations schemas.VariableExpirations, tfcVariableID string) error {
	variableExpirationsByte, err := json.Marshal(variableExpirations)
	if err != nil {
		return err
	}

	if tfcVariableID != "" {
		log.Debug("updating variable expirations on TFC")
		_, err = c.TFC.Variables.Update(c.Context, w.ID, tfcVariableID, tfc.VariableUpdateOptions{
			Value: tfc.String(string(variableExpirationsByte)),
		})
		return err
	}

	log.Debug("creating variable expirations on TFC")
	category := tfc.CategoryEnv
	_, err = c.TFC.Variables.Create(c.Context, w.ID, tfc.VariableCreateOptions{
		Key:       tfc.String(VariableExpirationsName),
		Value:     tfc.String(string(variableExpirationsByte)),
		Category:  &category,
		Sensitive: tfc.Bool(false),
		HCL:       tfc.Bool(false),
	})

	return err
}
