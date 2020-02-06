package schemas

type Defaults struct {
	Vault *Vault `hcl:"vault,block"`
	S5    *S5    `hcl:"s5,block"`
}
