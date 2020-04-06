package functions

import (
	"fmt"
	"os"
	"testing"

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

			got, _ := Env(test.Variable)
			if !got.RawEquals(test.Value) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Value)
			}
		})
	}
}
