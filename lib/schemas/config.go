package schemas

// Config handles all components that can be defined in
// a tfcw config file
type Config struct {
	TFC                  *TFC      `hcl:"tfc,block"`
	Defaults             *Defaults `hcl:"defaults,block"`
	TerraformVariables   Variables `hcl:"tfvar,block"`
	EnvironmentVariables Variables `hcl:"envvar,block"`

	Runtime Runtime
}

// Runtime is a struct used by the client in order
// to store values configured at runtime
type Runtime struct {
	WorkingDir string
	TFC        struct {
		Disabled     bool
		Address      string
		Token        string
		Organization string
		Workspace    string
	}
}
