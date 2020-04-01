# Example of a variable configuration using a value stored in a S5 payload ciphered with a Vault transit key

You can find more details on [how to use this cipher engine here](https://github.com/mvisonneau/s5/blob/master/examples/vault.md).

We will consider here that the necessary accesses for Vault and the transit key have been configured accordingly.

```hcl
tfc {
  organization = "acme"
  workspace {
    name = "foo"
  }
}

defaults {
  s5 {
    engine = "vault"
    vault {
      transit-key = "foo"
    }
  }
}

tfvvar "my_variable"{
  s5 {
    // Ciphered value
    value = "{{s5:OGRmNTNmMzViZjA4Y2VkMjk5M2U3NDY4OTYwZWY4MzI3ZmU1Y=}}"
  }
}
```

You can also override all the parameters on a per secret basis

```hcl
envvar "my_other_variable"{
  s5 {
    // In here you can optionally override all the default configuration
    vault {
      transit-key = "bar"
    }
    // ...

    // Ciphered value
    value = "{{s5:OGRmNTNmMzViZjA4Y2VkMjk5M2U3NDY4OTYwZWY4MzI3ZmU2Y=}}"
  }
}
```
