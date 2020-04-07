package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigGetVariables(t *testing.T) {
	cfg := &Config{
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
	}

	variables := cfg.GetVariables()
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
