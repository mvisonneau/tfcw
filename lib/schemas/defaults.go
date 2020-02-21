package schemas

// Defaults can handle default values for some providers
type Defaults struct {
	Vault *Vault `hcl:"vault,block"`
	S5    *S5    `hcl:"s5,block"`
}
