package client

import (
	"fmt"
	"sync"

	"github.com/mvisonneau/tfcw/lib/schemas"

	log "github.com/sirupsen/logrus"
)

func (c *Client) ProcessAllVariables(cfg *schemas.Config) (err error) {
	variables := schemas.Variables{}

	for _, variable := range cfg.TerraformVariables {
		variable.Kind = schemas.VariableKindTerraform
		variables = append(variables, variable)
	}

	for _, variable := range cfg.EnvironmentVariables {
		variable.Kind = schemas.VariableKindEnvironment
		variables = append(variables, variable)
	}

	return c.processVariables(variables)
}

func (c *Client) isVariableAlreadyProcessed(name string, kind schemas.VariableKind) bool {
	c.ProcessedVariablesMutex.Lock()
	k, exists := c.ProcessedVariables[name]
	c.ProcessedVariablesMutex.Unlock()

	if exists && k == kind {
		return true
	}

	c.ProcessedVariables[name] = kind
	return false
}

func (c *Client) processVariables(vars schemas.Variables) error {
	ch := make(chan error)
	wg := sync.WaitGroup{}

	for _, v := range vars {
		wg.Add(1)
		go func(c *Client, v *schemas.Variable) {
			defer wg.Done()
			ch <- c.processVariable(v)
		}(c, v)
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

func (c *Client) processVariable(v *schemas.Variable) error {
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
		return fmt.Errorf("You can't have more or less than one provider configured per variable. Found %d for '%s'", configuredProviders, v.Name)
	}

	// We can map several keys in a single API call
	if v.Vault != nil && v.Vault.Path != nil {
		if values, err := c.Vault.GetValues(v.Vault); err == nil {
			if v.Vault.Key == nil && (v.Vault.Keys == nil || len(*v.Vault.Keys) == 0) {
				return fmt.Errorf("You either need to set 'key' or 'keys' when using the Vault provider")
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
					setVariable(v)
				} else {
					return fmt.Errorf("key '%s' was not found in secret '%s'", vaultKey, *v.Vault.Path)
				}
			}
		}
	}

	if v.S5 != nil {
		value, err := c.S5.GetValue(v.S5)
		if err != nil {
			return fmt.Errorf("s5 error: %s", err)
		}

		v.Value = &value
		setVariable(v)
	}

	if v.Env != nil {
		value := c.Env.GetValue(v.Env)
		v.Value = &value
		setVariable(v)
	}
	return nil
}

func setVariable(v *schemas.Variable) error {
	log.WithFields(log.Fields{
		"kind":  v.Kind,
		"name":  v.Name,
		"value": secureSensitiveString(*v.Value),
	}).Infof("set!")

	return nil
}

func secureSensitiveString(sensitive string) string {
	if len(sensitive) < 4 {
		return "**********"
	}
	return fmt.Sprintf("%s********%s", string(sensitive[1]), string(sensitive[len(sensitive)-1]))
}
