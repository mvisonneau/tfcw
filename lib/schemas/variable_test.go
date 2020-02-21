package schemas

import (
	"fmt"
	"testing"

	"github.com/mvisonneau/go-helpers/test"
)

func TestVariableGetProviderEnv(t *testing.T) {
	v := &Variable{
		Env: &Env{},
	}

	p, err := v.GetProvider()
	test.Expect(t, err, nil)
	test.Expect(t, *p, VariableProviderEnv)
}

func TestVariableGetProviderS5(t *testing.T) {
	v := &Variable{
		S5: &S5{},
	}

	p, err := v.GetProvider()
	test.Expect(t, err, nil)
	test.Expect(t, *p, VariableProviderS5)
}

func TestVariableGetProviderVault(t *testing.T) {
	v := &Variable{
		Vault: &Vault{},
	}

	p, err := v.GetProvider()
	test.Expect(t, err, nil)
	test.Expect(t, *p, VariableProviderVault)
}

func TestVariableGetProviderInvalid(t *testing.T) {
	v := &Variable{
		Name: "foo",
	}
	p, err := v.GetProvider()
	var emptyProvider *VariableProvider
	test.Expect(t, err, fmt.Errorf("you can't have more or less than one provider configured per variable. Found 0 for 'foo'"))
	test.Expect(t, p, emptyProvider)
}
