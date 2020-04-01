# Example of multiple variable configuration using a single Vault secret

In this usecase, we will provision [AWS STS credentials](https://docs.aws.amazon.com/STS/latest/APIReference/Welcome.html) onto TFC env variables using the [Vault AWS secret engine](https://www.vaultproject.io/docs/secrets/aws/index.html).

We consider here that the [Vault token](https://learn.hashicorp.com/vault/getting-started/authentication)
has been made available either through the `VAULT_TOKEN` environment variable or at `~/.vault-token`

We also consider that the AWS secret engine has been configured properly and policies are allowing use from requesting the credentials.

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

// Notice that in this context, the <name> provided has no impact on the outcome
// Therefore you can use anything you would like as this value, it doesn't matter.
envvar "_" {
  vault {
    method = "write"
    path = "aws/sts/foo"

    keys = {
      access_key = "AWS_ACCESS_KEY_ID",
      secret_key = "AWS_SECRET_ACCESS_KEY",
      security_token = "AWS_SESSION_TOKEN",
    }

    params = {
      ttl = "15m"
    }
  }
}
```
