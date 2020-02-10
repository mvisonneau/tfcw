package schemas

type Config struct {
	TFC                  *TFC      `hcl:"tfc,block"`
	Defaults             *Defaults `hcl:"defaults,block"`
	TerraformVariables   Variables `hcl:"tfvar,block"`
	EnvironmentVariables Variables `hcl:"envvar,block"`
}
