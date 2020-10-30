package schemas

// Defaults can handle default values for some providers
type Defaults struct {
	Variable *VariableDefaults `hcl:"var,block"`
	Vault    *Vault            `hcl:"vault,block"`
	S5       *S5               `hcl:"s5,block"`
}

// VariableDefaults can handle default values for variables
type VariableDefaults struct {
	Sensitive *bool   `hcl:"sensitive"`
	HCL       *bool   `hcl:"hcl"`
	TTL       *string `hcl:"ttl"`
}
