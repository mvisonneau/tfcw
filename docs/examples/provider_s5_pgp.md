# Example of a variable configuration using a value stored in a S5 payload ciphered with a PGP keypair

You can find more details on [how to use this cipher engine here](https://github.com/mvisonneau/s5/blob/main/examples/pgp.md).

```hcl
tfc {
  organization = "acme"
  workspace {
    name = "foo"
  }
}

defaults {
  s5 {
    engine = "pgp"
    pgp {
      public-key-path  = "~/public-key.pem"
      private-key-path = "~/private-key.pem"
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
    pgp {
      public-key-path  = "~/other-public-key.pem"
      private-key-path = "~/other-private-key.pem"
    }
    // ...

    // Ciphered value
    value = "{{s5:OGRmNTNmMzViZjA4Y2VkMjk5M2U3NDY4OTYwZWY4MzI3ZmU2Y=}}"
  }
}
```
