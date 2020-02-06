package schemas

type Config struct {
	Defaults             *Defaults `hcl:"defaults,block"`
	TerraformVariables   Variables `hcl:"tfvar,block"`
	EnvironmentVariables Variables `hcl:"envvar,block"`
}
