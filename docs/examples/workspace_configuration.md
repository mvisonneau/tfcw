# Example of a complete workspace configuration

This example demonstrate how to fully configure a workspace using TFCW

```hcl
tfc {
  // Name of your organization on TFC (required)
  organization = "acme"

  // Workspace related configuration block (required)
  workspace {
    // Name of the workspace of your Terraform stack on TFC (required)
    name = "foo"

    // Whether to run terraform remotely or locally (optional, default: true (remotely))
    operations = false

    // Configure the workspace with the auto-apply flag (optional, default: <unmanaged>)
    auto-apply = true

    // Configure the workspace terraform version (optional, default: <unmanaged>)
    terraform-version = "0.12.24"

    // Configure the workspace working directory (optional, default: <unmanaged>)
    working-directory = "/foo"

    // Name of the SSH key to use (optional, default: <unmanaged>)
    ssh-key = "bar"
  }

  // This flag enables the creating of the workspace if TFCW cannot find it under
  // the organization (optional, default: true)
  workspace-auto-create = true

  // Whether to purge or leave the workspace variables which are
  // not configured within this file (optional, default: false)
  purge-unmanaged-variables = false
}
```
