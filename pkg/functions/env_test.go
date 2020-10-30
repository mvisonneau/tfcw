package functions

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestFunctionEnv(t *testing.T) {
	tests := []struct {
		Variable cty.Value
		Value    cty.Value
	}{
		{
			cty.StringVal("FOO"),
			cty.StringVal("BAR"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("env(%#v)", test.Variable.AsString()), func(t *testing.T) {
			os.Setenv(test.Variable.AsString(), test.Value.AsString())

			v, _ := Env(test.Variable)
			assert.Equal(t, test.Value, v)
		})
	}
}
