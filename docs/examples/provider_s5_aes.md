# Example of a variable configuration using a value stored in a S5 payload ciphered with AES

You can find more details on [how to use this cipher engine here](https://github.com/mvisonneau/s5/blob/master/examples/aes-gcm.md).

We will consider here that the `S5_AES_KEY` environment variable has been configured accordingly.

```hcl
tfc {
  organization = "acme"
  workspace {
    name = "foo"
  }
}

defaults {
  s5 {
    engine = "aes"
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
