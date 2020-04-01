# TFCW - Configuration syntax

## Minimal configuration

```hcl
tfc {
  // Your organization name in Terraform Cloud
  organization = "foo"

  // A workspace block with the name of your workspace
  workspace {
    name = "bar"
  }
}
```

## Block types

There are **4 block types** supported by TFCW:

|**name**|**description**|**required**|**unique**|
|---|---|---|---|
|[tfc](#tfc)|configuration related to TFC and the workspace|`yes`|`yes`|
|[defaults](#defaults)|a block containing some default configuration for the variable providers|`no`|`yes`|
|[tfvar](#tfvar)|defines a [Terraform](https://www.terraform.io/docs/cloud/workspaces/variables.html#terraform-variables) variable in TFC|`no`|`no`|
|[envvar](#envvar)|defines an [Environment](https://www.terraform.io/docs/cloud/workspaces/variables.html#environment-variables) variable in TFC|`no`|`no`|

`tfvar` and `envvar` share exactly the same capabilities. They only differ in the sense of types of variables they provider on the TFC API.

### tfc

`tfc` is a required block that defines the configuration of your workspace

```hcl
tfc {
  // Name of your organization on TFC (required)
  organization = "acme"

  // Workspace related configuration block (required)
  workspace {
    // Name of the workspace of your Terraform stack on TFC (required)
    name = "foo"

    // Whether to run terraform remotely or locally (optional, default: true (remotely))
    operations = true

    // Configure the workspace with the auto-apply flag (optional, default: <unmanaged>)
    auto-apply = false

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

Here is a contextualized example: [docs/examples/workspace_configuration.md](docs/examples/workspace_configuration.md)

### defaults

`defaults` is an optional block that allows you to define default configuration for the variable providers you are planning on using.

```hcl
defaults {
  
  // You can define as many provider blocks as you want
  // Default Vault configuration
  vault {
    ...
  }

  // Default S5 configuration
  s5 {
    ...
  }
  
  // There is no default configuration support for the env provider though
}
```

### tfvar

`tfvar` defines a [Terraform](https://www.terraform.io/docs/cloud/workspaces/variables.html#terraform-variables) variable in TFC. You can only use **one** provider block in each `tfvar` block.

```hcl
tfvar "<name>" {
  // Name can be used to override the label of the resource (optional, default: <label name>)
  // NB: This value has to be unique amongst all the definitions.
  // You can have a tfvar "foo" {} and a envvar "foo" {} defined at the same time
  name = "<name_override>"

  // Whether to declare this variable sensitive in TFC (optional, default: true)
  // More information: https://www.terraform.io/docs/cloud/workspaces/variables.html#sensitive-values
  sensitive = true

  // Whether to interprete this variable content as HCL in TFC (optional, default: false)
  // More information: https://www.terraform.io/docs/cloud/workspaces/variables.html#hcl-values
  hcl = false

  // You have to define exactly ONE provider between vault{}, s5{} or env{}
  vault {
    ...
  }

  // or
  s5 {
    ...
  }

  // or
  env {
    ...
  }
}
```

### envvar

`envvar` defines a [Terraform](https://www.terraform.io/docs/cloud/workspaces/variables.html#terraform-variables) variable in TFC. You can only use **one** provider block in each `envvar` block.

```hcl
envvar "<name>" {
  // Name can be used to override the label of the resource (optional, default: <label name>)
  // NB: This value has to be unique amongst all the definitions.
  // You can have a tfvar "foo" {} and a envvar "foo" {} defined at the same time
  name = "<name_override>"

  // Whether to declare this variable sensitive in TFC (optional, default: true)
  // More information: https://www.terraform.io/docs/cloud/workspaces/variables.html#sensitive-values
  sensitive = true

  // Whether to interprete this variable content as HCL in TFC (optional, default: false)
  // More information: https://www.terraform.io/docs/cloud/workspaces/variables.html#hcl-values
  hcl = false

  // You have to define exactly ONE provider between vault{}, s5{} or env{}
  vault {
    ...
  }

  // or
  s5 {
    ...
  }

  // or
  env {
    ...
  }
}
```

### Provider block types (or subblocks 🤷‍♂️)

Provider blocks can be used under either `defaults`, `tfvar` or `envvar` blocks. They represent the necessary configuration to access the data from the provider. There is currently 3 kind of provider blocks:

- [vault](#vault) to fetch values from [Vault](https://www.vaultproject.io/)
- [s5](#s5) to fetch values through [s5](https://github.com/mvisonneau/s5)
- [env](#env) to fetch values from environment variables

#### vault

```hcl
vault {
  // Vault endpoint (required, can also be defined using the
  // VAULT_ADDR env variable)
  address = "https://vault.acme.local"

  // Vault token (required, can also be defined using the
  // VAULT_TOKEN env variable or at ~/.vault-token)
  token = "s.FCcSvkeZaCsIkddhdQ9Itn3g"

  // Following parameters can be also defined here but are more commonly defined
  // on a per secret basis
  //

  // Method to use for making requests (optional, default: read)
  method = "read"

  // Path to query for getting the value (required, default: <empty_string>)
  path = ""

  // Params to add to the query (optional, default: <empty_map>)
  params = {}

  // The following ones are mutually exclusive but required, you need to use one of them
  //

  // Key of the secret data to use as a value (required, default: <empty_string>)
  key = ""

  // Keys is a mapping of the keys in the secret to assign with variable names in TFC
  // Using this parameter will overide the `name` of the secret and actually iterate over this list
  // in order to create all the desired variables (required, default: <empty_map>)
  keys = {}
}
```

Here are contextualized examples:

- [docs/examples/provider_vault.md](docs/examples/provider_vault.md)
- [docs/examples/provider_vault_multi_keys.md](docs/examples/provider_vault_multi_keys.md)

#### s5

```hcl
s5 {
  // S5 engine to use (required)
  // Can either be "aes", "aws", "gcp", "pgp" or "vault"
  engine = "aes"

  // AES configuration
  // More details here: https://github.com/mvisonneau/s5/blob/master/examples/aes-gcm.md
  aes {
    // AES key to use (required, can also be defined using the S5_AES_KEY env variable)
    key = "3cf9d1b57c588f68bfd04b2e9644bd9e90c03cd18d15caba9d5b0b7162d52a69"
  }

  // AWS configuration
  // More details here: https://github.com/mvisonneau/s5/blob/master/examples/aws-kms.md
  aws {
    // ARN of the KMS key to use (required, can also be defined using the S5_AWS_KMS_KEY_ARN env variable)
    kms-key-arn = "arn:aws:kms:*:111111111111:key/mykey"
  }

  // GCP configuration
  // More details here: https://github.com/mvisonneau/s5/blob/master/examples/gcp-kms.md
  gcp {
    // Name of the KMS key to use (required, can also be defined using the S5_GCP_KMS_KEY_NAME env variable)
    kms-key-name = "foo"
  }

  // PGP configuration
  // More details here: https://github.com/mvisonneau/s5/blob/master/examples/pgp.md
  pgp {
    public-key-path  = "~/public-key.pem"
    private-key-path = "~/private-key.pem"
  }

  // Vault configuration
  // More details here: https://github.com/mvisonneau/s5/blob/master/examples/vault.md
  vault {
    transit-key = "default"
  }
}
```

Here are contextualized examples:

- [docs/examples/provider_s5_aes.md](docs/examples/provider_s5_aes.md)
- [docs/examples/provider_s5_aws_kms.md](docs/examples/provider_s5_aws_kms.md)
- [docs/examples/provider_s5_gcp_kms.md](docs/examples/provider_s5_gcp_kms.md)
- [docs/examples/provider_s5_pgp.md](docs/examples/provider_s5_pgp.md)
- [docs/examples/provider_s5_vault.md](docs/examples/provider_s5_vault.md)

#### env

`env` is the easiest to implement. You only need to specific which environment variable to fetch the value from.

```hcl
env {
  variable = "FOO"
}
```

Here is a contextualized example: [docs/examples/provider_env.md](docs/examples/provider_env.md)