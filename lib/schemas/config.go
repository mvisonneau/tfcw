package schemas

// Config handles all components that can be defined in
// a tfcw config file
type Config struct {
	TFC                  *TFC      `hcl:"tfc,block"`
	Defaults             *Defaults `hcl:"defaults,block"`
	TerraformVariables   Variables `hcl:"tfvar,block"`
	EnvironmentVariables Variables `hcl:"envvar,block"`
}
