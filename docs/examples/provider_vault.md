# Example of a variable configuration using a value stored in a Vault secret

We consider here that the [Vault token](https://learn.hashicorp.com/vault/getting-started/authentication)
has been made available either through the `VAULT_TOKEN` environment variable or at `~/.vault-token`

```hcl
tfc {
  organization = "acme"
  workspace {
    name = "foo"
  }
}

defaults {
  vault {
    address = "https://vault.acme.local"
  }
}

tfvar "my_variable" {
  vault {
    path = "secret/mysecret"
    key = "foo"
  }
}
```

You can also override all the parameters on a per secret basis

```hcl
envvar "my_other_variable" {
  vault {
    // In here you can optionally override all the defaults values
    address = "https://alt-vault.acme.local"
    token   = "alternative-token"
    method = "write"
    // ...

    path = "secret/mysecret"
    key = "bar"
  }
}
```
