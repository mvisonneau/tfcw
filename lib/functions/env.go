package functions

import (
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// EnvFunction constructs a function that looks up an EnvironmentVariable given the
// variable name
var EnvFunction = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "envvar",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		envvar := args[0].AsString()
		return cty.StringVal(os.Getenv(envvar)), nil
	},
})

// Env performs a function call to EnvFunction, useful for testing
func Env(envvar cty.Value) (cty.Value, error) {
	return EnvFunction.Call([]cty.Value{envvar})
}
