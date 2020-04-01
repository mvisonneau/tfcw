# Example of a variable configuration using a value stored in a environment variable on the system

This one is as easy as:

```hcl
tfc {
  organization = "acme"
  workspace {
    name = "foo"
  }
}

tfvar "my_variable" {
  env {
    variable = "FOO"
  }
}
```

This will provision the value of the `FOO` environment variable into a **Terraform variable** named `my_variable`.
