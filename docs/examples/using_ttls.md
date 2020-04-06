# Example of a variable configuration with a defined Time To Live (TTL)

You can leverage the `ttl` field on your variables to make TFCW more efficient with the management of your variables.

```hcl
defaults {
  ttl = "15m"
}

tfvar "my_variable" {
  vault {
    path = "secret/mysecret"
    key = "foo"
  }
}

envvar "my_other_variable" {
  ttl = "1h"
  vault {
    path = "secret/mysecret"
    key = "bar"
  }
}
```

With this configuration, TFCW will only update `my_variable` every 15m and `my_other_variable` every hour.

As a rule of thumb, be cautious and use values lower than the actual expiration of the values in order to leave enough time to your Terraform run to execute successfully.
