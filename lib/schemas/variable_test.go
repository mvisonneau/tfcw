package schemas

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariableGetProviderEnv(t *testing.T) {
	v := &Variable{
		Env: &Env{},
	}

	p, err := v.GetProvider()
	assert.Equal(t, err, nil)
	assert.Equal(t, *p, VariableProviderEnv)
}

func TestVariableGetProviderS5(t *testing.T) {
	v := &Variable{
		S5: &S5{},
	}

	p, err := v.GetProvider()
	assert.Equal(t, err, nil)
	assert.Equal(t, *p, VariableProviderS5)
}

func TestVariableGetProviderVault(t *testing.T) {
	v := &Variable{
		Vault: &Vault{},
	}

	p, err := v.GetProvider()
	assert.Equal(t, err, nil)
	assert.Equal(t, *p, VariableProviderVault)
}

func TestVariableGetProviderInvalid(t *testing.T) {
	v := &Variable{
		Name: "foo",
	}
	p, err := v.GetProvider()
	var emptyProvider *VariableProvider
	assert.Equal(t, err, fmt.Errorf("you can't have more or less than one provider configured per variable. Found 0 for 'foo'"))
	assert.Equal(t, p, emptyProvider)
}
