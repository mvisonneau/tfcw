package schemas

// TFC handles Terraform Cloud related configuration
type TFC struct {
	Address      *string    `hcl:"address"`
	Token        *string    `hcl:"token"`
	Organization *string    `hcl:"organization"`
	Workspace    *Workspace `hcl:"workspace,block"`

	WorkspaceAutoCreate     *bool `hcl:"workspace-auto-create"`
	PurgeUnmanagedVariables *bool `hcl:"purge-unmanaged-variables"`
}

// Workspace is used to refer to and configure the workspace
type Workspace struct {
	Name             *string `hcl:"name"`
	Operations       *bool   `hcl:"operations"`
	AutoApply        *bool   `hcl:"auto-apply"`
	TerraformVersion *string `hcl:"terraform-version"`
	WorkingDirectory *string `hcl:"working-directory"`
	SSHKey           *string `hcl:"ssh-key"`
}
