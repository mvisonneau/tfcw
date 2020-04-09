package schemas

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// Config handles all components that can be defined in
// a tfcw config file
type Config struct {
	TFC                  *TFC      `hcl:"tfc,block"`
	Defaults             *Defaults `hcl:"defaults,block"`
	TerraformVariables   Variables `hcl:"tfvar,block"`
	EnvironmentVariables Variables `hcl:"envvar,block"`

	Runtime Runtime
}

// Runtime is a struct used by the client in order
// to store values configured at runtime
type Runtime struct {
	WorkingDir string
	TFC        struct {
		Address      string
		Token        string
		Organization string
		Workspace    string
	}
}

// GetVariables returns a Variables containing the configured variables
func (cfg *Config) GetVariables() (variables Variables) {
	for _, variable := range cfg.TerraformVariables {
		variable.Kind = VariableKindTerraform
		variables = append(variables, variable)
	}

	for _, variable := range cfg.EnvironmentVariables {
		variable.Kind = VariableKindEnvironment
		variables = append(variables, variable)
	}

	return
}

// GetVariableTTL returns the TTL of a variable
func (cfg *Config) GetVariableTTL(v *Variable) (ttl time.Duration, err error) {
	if v.TTL != nil {
		return time.ParseDuration(*v.TTL)
	}

	if cfg.Defaults != nil && cfg.Defaults.Variable != nil && cfg.Defaults.Variable.TTL != nil {
		return time.ParseDuration(*cfg.Defaults.Variable.TTL)
	}
	return time.Duration(0), nil
}

// ComputeNewVariableExpirations ...
func (cfg *Config) ComputeNewVariableExpirations(updatedVariables Variables, existingVariableExpirations VariableExpirations) (variableExpirations VariableExpirations, hasChanges bool, err error) {
	if len(existingVariableExpirations) > 0 {
		variableExpirations = existingVariableExpirations
	} else {
		variableExpirations = make(VariableExpirations)
	}

	if len(updatedVariables) > 0 {
		hasChanges = true
	}

	for _, v := range updatedVariables {
		if _, ok := variableExpirations[v.Kind]; !ok {
			variableExpirations[v.Kind] = map[string]*VariableExpiration{}
		}

		var ttl time.Duration
		ttl, err = cfg.GetVariableTTL(v)
		if err != nil {
			return
		}

		// If there is no TTL defined, we omit this variable from the expirations list
		if ttl == 0 {
			if _, ok := variableExpirations[v.Kind][v.Name]; ok {
				delete(variableExpirations[v.Kind], v.Name)
			}
			continue
		}

		variableExpirations[v.Kind][v.Name] = &VariableExpiration{
			TTL:      ttl,
			ExpireAt: time.Now().Add(ttl),
		}
	}

	// Cleanup maps which could turn out to be empty
	for k := range variableExpirations {
		if len(variableExpirations[k]) == 0 {
			delete(variableExpirations, k)
		}
	}

	return
}

// GetVariablesToUpdate returns the list of the variables to update based on the current configuration
// and the existing variables
func (cfg *Config) GetVariablesToUpdate(variableExpirations VariableExpirations) (variables Variables, err error) {
	for _, v := range cfg.GetVariables() {
		var ttl time.Duration
		ttl, err = cfg.GetVariableTTL(v)
		if err != nil {
			return
		}

		if ttl > 0 {
			// Check if there is a TTL flag set for this variable
			if variableExpiration, ok := variableExpirations[v.Kind][v.Name]; ok {
				// If the TTL hasn't changed and the expiration date is still in the future we do not update it
				if ttl == variableExpiration.TTL && variableExpiration.ExpireAt.After(time.Now()) {
					log.Debugf("variable %s (%s) is still valid for %s (ttl: %s), not updating", v.Name, v.Kind, variableExpiration.ExpireAt.Sub(time.Now()).String(), variableExpiration.TTL.String())
					continue
				}
			}
		}

		variables = append(variables, v)
	}
	return
}
