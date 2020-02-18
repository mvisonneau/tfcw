# Complete example of the configuration syntax

```hcl
// tfcw.hcl

// Terraform Cloud related configuration
tfc {
  // Name of your organization on TFC (required)
  organization = "acme"

  // Name of the workspace of your Terraform stack on TFC (required)
  workspace = "foo"

  // Whether to purge or leave the workspace variables which are
  // not configured within this file (optional, default: false)
  purge-unmanaged-variables = false
}

// Default configuration of the storage engines
defaults {
  
  // Default Vault configuration
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

  // Default S5 configuration
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
}

//
// Variables definition
//

// Terraform variables can be defined using the `tfvar` resource type, eg:

tfvar "foo" {
  // Using the environment provider
  env {
    variable = "FOO"
  }
}

// Environment variables can be defined using the `envvar` resource type, eg:

envvar "FOO" {
  // Using the environment provider
  env {
    variable = "FOO"
  }
}

// Here is an example of the full list of options both tfvar and envvar resources can have

tfvar "all-options" {
  // Name can be used to override the label of the resource (optional, default: <label name>)
  // NB: This value has to be unique amongst all the definitions.
  // You can have a tfvar "foo" {} and a envvar "foo" {} defined at the same time
  name = "all-options-alt"

  // Whether to declare this variable sensitive in TFC (optional, default: true)
  // More information: https://www.terraform.io/docs/cloud/workspaces/variables.html#sensitive-values
  sensitive = true

  // Whether to interprete this variable content as HCL in TFC (optional, default: false)
  // More information: https://www.terraform.io/docs/cloud/workspaces/variables.html#hcl-values
  hcl = false

  // You then must define only ONE provider between vault{}, s5{} or env{}
  // eg:
  env {
    variable = "ALL_OPTIONS"
  }
}


//
// Environment provider (env)
//

tfvar "env" {
  // env only takes one option, the variable name
  env {
    variable = "ENV"
  }
}

//
// Vault provider
//

tfvar "vault_single_key" {
  vault {
    // In here you can optionally override all the default configuration
    address = "https://alt-vault.acme.local"
    token   = "alternative-token"
    method = "write"
    // ...

    path = "secret/mysecret"
    key = "FOO"
  }
}

envvar "vault_multi_key" {
  vault {
    method = "write"
    path = "aws/sts/foo"

    keys = {
      access_key = "AWS_ACCESS_KEY_ID",
      secret_key = "AWS_SECRET_ACCESS_KEY",
      security_token = "AWS_SECURITY_TOKEN",
    }

    params = {
      ttl = "15m"
    }
  }
}


//
// S5 provider
//

// It can be as simple if you are leveraging the default configuration
envvar "s5_variable_with_default"{
  s5 {
    // Ciphered value
    value = "{{s5:OGRmNTNmMzViZjA4Y2VkMjk5M2U3NDY4OTYwZWY4MzI3ZmU1Y=}}"
  }
}

// You can also override all the parameters on a per secret basis
envvar "s5_variable_with_override"{
  s5 {
     // In here you can optionally override all the default configuration
    engine = "aws"
    aws {
      kms-key-arn = "arn:aws:kms:*:111111111111:key/anotherkey"
    }
    // ...

    // Ciphered value
    value = "{{s5:OGRmNTNmMzViZjA4Y2VkMjk5M2U3NDY4OTYwZWY4MzI3ZmU1Y=}}"
  }
}
```
