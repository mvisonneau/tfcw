package schemas

import (
	"fmt"
	"testing"
	"time"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
)

var testConfig = &Config{
	TerraformVariables: Variables{
		&Variable{
			Name: "foo",
		},
	},
	EnvironmentVariables: Variables{
		&Variable{
			Name: "bar",
		},
	},
	Defaults: &Defaults{
		Variable: &VariableDefaults{
			TTL: pointy.String("15m"),
		},
	},
}

const (
	fifteenMinuteDuration = time.Minute * 15
	thirtyMinuteDuration  = time.Minute * 30
)

func TestConfigGetVariables(t *testing.T) {
	variables := testConfig.GetVariables()
	assert.Len(t, variables, 2)

	assert.Equal(t, Variable{
		Kind: VariableKindTerraform,
		Name: "foo",
	}, *(variables[0]))

	assert.Equal(t, Variable{
		Kind: VariableKindEnvironment,
		Name: "bar",
	}, *(variables[1]))
}

func TestConfigGetVariableTTL(t *testing.T) {
	// Defining the TTL in the Variable
	ttl, err := testConfig.GetVariableTTL(&Variable{TTL: pointy.String("30m")})
	assert.NoError(t, err)
	assert.Equal(t, thirtyMinuteDuration, ttl)

	// With an incorrect value in the variable definition
	_, err = testConfig.GetVariableTTL(&Variable{TTL: pointy.String("foo")})
	assert.Error(t, err)

	// Using the default configuration value
	ttl, err = testConfig.GetVariableTTL(&Variable{})
	assert.NoError(t, err)
	assert.Equal(t, fifteenMinuteDuration, ttl)

	// With an incorrect value in the default config definition
	incorrectDefaultConfig := &Config{
		Defaults: &Defaults{
			Variable: &VariableDefaults{
				TTL: pointy.String("foo"),
			},
		},
	}

	_, err = incorrectDefaultConfig.GetVariableTTL(&Variable{})
	assert.Error(t, err)

	// Without any configuration
	emptyConfig := &Config{}
	ttl, err = emptyConfig.GetVariableTTL(&Variable{})
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ttl)
}

func TestComputeNewVariableExpirations(t *testing.T) {
	existingVariableExpirations := VariableExpirations{}
	updatedVariables := Variables{
		&Variable{
			Kind: VariableKindEnvironment,
			Name: "foo",
		},
	}

	variableExpirations, hasChanges, err := testConfig.ComputeNewVariableExpirations(updatedVariables, existingVariableExpirations)
	assert.NoError(t, err)
	assert.True(t, hasChanges)
	assert.NotNil(t, variableExpirations)
	assert.NotNil(t, variableExpirations[VariableKindEnvironment])
	assert.NotNil(t, variableExpirations[VariableKindEnvironment]["foo"])
	assert.Equal(t, fifteenMinuteDuration, variableExpirations[VariableKindEnvironment]["foo"].TTL)
	assert.True(t, variableExpirations[VariableKindEnvironment]["foo"].ExpireAt.After(time.Now()))
	assert.True(t, variableExpirations[VariableKindEnvironment]["foo"].ExpireAt.Before(time.Now().Add(thirtyMinuteDuration)))

	// Already existent expiration variable and no changes
	existingVariableExpirations = make(VariableExpirations)
	existingVariableExpirations[VariableKindTerraform] = make(map[string]*VariableExpiration)
	existingVariableExpirations[VariableKindTerraform]["foo"] = &VariableExpiration{
		TTL:      fifteenMinuteDuration,
		ExpireAt: time.Now().Add(fifteenMinuteDuration),
	}

	_, hasChanges, err = testConfig.ComputeNewVariableExpirations(Variables{}, existingVariableExpirations)
	assert.NoError(t, err)
	assert.False(t, hasChanges)

	// No TTL defined on an updatedVariable
	emptyConfig := &Config{}
	updatedVariables = Variables{
		&Variable{
			Kind: VariableKindTerraform,
			Name: "foo",
		},
	}

	variableExpirations, hasChanges, err = emptyConfig.ComputeNewVariableExpirations(updatedVariables, existingVariableExpirations)
	assert.NoError(t, err)
	assert.True(t, hasChanges)
	assert.Len(t, variableExpirations, 0)
	fmt.Println(variableExpirations)

	// Invalid TTL of an updated variable
	updatedVariables = Variables{
		&Variable{
			Name: "foo",
			TTL:  pointy.String("bar"),
		},
	}
	_, _, err = testConfig.ComputeNewVariableExpirations(updatedVariables, existingVariableExpirations)
	assert.Error(t, err)
}

func TestConfigGetVariablesToUpdate(t *testing.T) {
	// Empty expiration variables, we update all variables in the config
	variableExpirations := VariableExpirations{}
	variables, err := testConfig.GetVariablesToUpdate(variableExpirations)
	assert.NoError(t, err)
	assert.Equal(t, testConfig.GetVariables(), variables)

	// Expiration still in range
	variableExpirations = make(VariableExpirations)
	variableExpirations[VariableKindTerraform] = make(map[string]*VariableExpiration)
	variableExpirations[VariableKindTerraform]["foo"] = &VariableExpiration{
		TTL:      fifteenMinuteDuration,
		ExpireAt: time.Now().Add(fifteenMinuteDuration),
	}
	variables, err = testConfig.GetVariablesToUpdate(variableExpirations)
	assert.NoError(t, err)
	assert.Len(t, variables, 1)
	assert.Equal(t, "bar", variables[0].Name)

	// Invalid variable TTL
	invalidVariableTTLConfig := &Config{
		TerraformVariables: Variables{
			&Variable{
				Name: "foo",
				TTL:  pointy.String("bar"),
			},
		},
	}
	_, err = invalidVariableTTLConfig.GetVariablesToUpdate(variableExpirations)
	assert.Error(t, err)
}
