package schemas

type VariableKind string

const (
	VariableKindTerraform   VariableKind = "terraform"
	VariableKindEnvironment VariableKind = "environment"
)

type Variable struct {
	Name  string `hcl:"name,label"`
	Vault *Vault `hcl:"vault,block"`
	S5    *S5    `hcl:"s5,block"`
	Env   *Env   `hcl:"env,block"`

	Kind  VariableKind
	Value *string
}

type Variables []*Variable
