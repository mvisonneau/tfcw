# 💝 TFCW - Terraform Cloud Wrapper

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mvisonneau/tfcw)](https://pkg.go.dev/mod/github.com/mvisonneau/tfcw)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvisonneau/tfcw)](https://goreportcard.com/report/github.com/mvisonneau/tfcw)
[![Docker Pulls](https://img.shields.io/docker/pulls/mvisonneau/tfcw.svg)](https://hub.docker.com/r/mvisonneau/tfcw/)
[![test](https://github.com/mvisonneau/tfcw/actions/workflows/test.yml/badge.svg)](https://github.com/mvisonneau/tfcw/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/mvisonneau/tfcw/badge.svg?branch=main)](https://coveralls.io/github/mvisonneau/tfcw?branch=main)
[![release](https://github.com/mvisonneau/tfcw/actions/workflows/release.yml/badge.svg)](https://github.com/mvisonneau/tfcw/actions/workflows/release.yml)
[![tfcw](https://snapcraft.io/tfcw/badge.svg)](https://snapcraft.io/tfcw)
[![Gitter](https://badges.gitter.im/terraform-cloud-wrapper/community.svg)](https://gitter.im/terraform-cloud-wrapper/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

`Terraform Cloud Wrapper (TFCW)` wraps the Terraform Cloud API. It provides an easy way to **dynamically** maintain configuration and particularily **sensitive [variables](https://www.terraform.io/docs/cloud/workspaces/variables.html)** of [Terraform Cloud (TFC) workspaces](https://www.terraform.io/docs/cloud/workspaces/index.html).

## TL;DR

> **Use case:** You need a token or API key for your terraform provider which is stored in Vault

First, you do not have to change any of your Terraform code, although you can eventually omit the remote backend block if you want to:

```hcl
// terraform.tf

terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "acme"

    workspaces {
      name = "foo"
    }
  }
}

provider "cloudflare" {
  version = "~> 2.0"
  email = "foo@bar.com"
}

resource "cloudflare_zone" "example" {
  zone = "example.com"
}
```

You need to add a new file within your Terraform folder (or anywhere you would like to store it) which can look like this:

```hcl
// tfcw.hcl

envvar "CLOUDFLARE_API_TOKEN" {
  vault {
    address = "https://vault.acme.local"
    path    = "secret/cloudflare"
    key     = "api-token"
  }
}
```

That's it, you now have a declarative way to ensure that your variable is picked up from **Vault** whenever you trigger a Terraform run, even if it is a [remote operation](https://www.terraform.io/docs/cloud/run/index.html#remote-operations)!

```bash
// Render the variables on TFC
~$ tfcw render
INFO[] Checking workspace configuration
INFO[] Processing variables and updating their values on TFC
INFO[] Set variable 'CLOUDFLARE_API_TOKEN' (environment)

// Run terraform
~$ terraform plan
...

// Or all-in-one
~$ tfcw run create
INFO[] Checking workspace configuration
INFO[] Processing variables and updating their values on TFC
INFO[] Set variable 'CLOUDFLARE_API_TOKEN' (environment)
INFO[] Preparing plan
Terraform v0.13.5
Configuring remote state backend...
Initializing Terraform configuration...
[...]
```

## Why should I use TFCW

It is particularily useful when you work with ephemeral secrets which need to be renewed fairly often. However, you can also get massive benefit from it if you want to have a declarative way to manage your workspaces definitions as code.

You will most likely do not need to learn a new configuration syntax as TFCW is configured using HCL files. 
TFCW allows you to do all that whilst continuing to **only write HCL files**.

### Without TFCW

Assuming you are using Terraform Cloud for managing this super simple stack:

```hcl
// terraform.tf

terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "acme"
    workspaces {
      name = "foo"
    }
  }
}

provider "local" {
   version = "~> 1.4.0"
}

variable "credentials" {}

resource "local_file" "credentials_file" {
  filename        = "./credentials"
  file_permissons = "0600"
  content         = var.credentials
}
```

You then need to manage the value of the `credentials` somehow 🤷‍♂️ either through a **.tfvars** file or straight into the [Terraform Cloud workspace definition](https://www.terraform.io/docs/cloud/workspaces/variables.html). There are many different ways of achieving this but this can however quickly become a burden to maintain, specially at scale or when using different ways to store the sensitive values of the variables.

Once done, you can trigger a `terraform` run, manually, through the Terraform Cloud API or using a GitOps approach. eg:

```shell
~$ terraform plan
...
```

## With TFCW

You do not have to change any of your Terraform code:

```hcl
// terraform.tf

terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "acme"

    workspaces {
      name = "foo"
    }
  }
}

provider "local" {
   version = "~> 1.4.0"
}

variable "credentials" {}

resource "local_file" "credentials_file" {
  filename        = "./credentials"
  file_permission = "0600"
  content         = var.credentials
}
```

You need to add a new file within your Terraform folder (or anywhere you would like to store it) which can look like this at a bare minimum:

```hcl
// tfcw.hcl

tfvar "credentials" {
  // If you do not want to update the value on each run, you can optionally set a TTL
  // to let TFCW aware that it should still be valid.
  ttl = "1h"

  vault {
    address = "https://vault.acme.local"
    path    = "secret/very_sensitive"
    key     = "data"
  }
}
```

That's it, you now have a declarative way to ensure that your variable is picked up from **Vault** whenever you trigger a Terraform run, even if it is a [remote operation](https://www.terraform.io/docs/cloud/run/index.html#remote-operations)!

```shell
~$ tfcw run create
INFO[] Checking workspace configuration
INFO[] Processing variables and updating their values on TFC
INFO[] Preparing plan
Terraform v0.13.5
Configuring remote state backend...
Initializing Terraform configuration...
2020/04/06 17:23:13 [DEBUG] Using modified User-Agent: Terraform/0.12.24 TFC/d310d4ebb1
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

local_file.foo: Refreshing state... [id=376ee705fb211bc1d753477ba5607e0c3754009b]

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # local_file.credentials_file will be created
  + resource "local_file" "credentials_file" {
      + content              = "secret_value"
      + directory_permission = "0777"
      + file_permission      = "0600"
      + filename             = "./credentials"
      + id                   = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

If you would prefer to keep your current way of triggering the Terraform runs, you can also simply use the `render` command which will _only_ update the variables in Terraform Cloud or even locally:

```shell
~$ tfcw render --help
NAME:
   tfcw render - render variables values

USAGE:
   tfcw render [command options] [arguments...]

OPTIONS:
   --render-type value, -r value  where to render to values - options are : tfc, local or disabled (default: "tfc")
   --ignore-ttls                  render all variables, unconditionnaly of their current expirations or configured TTLs
   --dry-run                      simulate what TFCW would do onto the TFC API
```

You can also do [dry runs](https://en.wikipedia.org/wiki/Dry_run_(testing)) if you want to get insights about what tfcw would actually do.

```shell
~$ tfcw render --dry-run
INFO[] Checking workspace configuration
INFO[] Processing variables and updating their values on TFC
INFO[] [DRY-RUN] Set variable credentials - (terraform) : x********x
```

## Configuration syntax

The [configuration syntax is maintained here](docs/configuration_syntax.md).

## Examples

Several examples are available in the [docs/examples](docs/examples) folder of this repository.

## Supported sources for storing variable values

We currently support **6 sources** as variable storage backends (2 natively and 4 others through [s5](https://github.com/mvisonneau/s5) payloads):

### Natively supported

- [Vault](https://www.vaultproject.io) (of course!)
- [Environment variables](https://en.wikipedia.org/wiki/Environment_variable)

### Through S5 payloads

- [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) - [GCM](https://en.wikipedia.org/wiki/Galois/Counter_Mode) (using hexadecimal keys >= 128b)
- [AWS](https://aws.amazon.com) - [KMS](https://aws.amazon.com/kms/)
- [GCP](https://cloud.google.com) - [KMS](https://cloud.google.com/kms/)
- [PGP](https://www.openpgp.org/)
- ([Vault](https://www.vaultproject.io) is also supported through S5)

## Usage

```bash
~$ tfcw --help
NAME:
   tfcw - Terraform Cloud Wrapper

USAGE:
   tfcw [global options] command [command options] [arguments...]

COMMANDS:
   render         render variables values
   run            manipulate runs
   workspace, ws  manipulate the workspace
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --address address, -a address                 address to access Terraform Cloud API [$TFCW_ADDRESS]
   --config-file path, -c path                   path of a readable TFCW configuration file (.hcl or .json) (default: "<working-dir>/tfcw.hcl") [$TFCW_CONFIG_FILE]
   --log-level level                             log level (debug,info,warn,fatal,panic) (default: "info") [$TFCW_LOG_LEVEL]
   --log-format format                           log format (json,text) (default: "text") [$TFCW_LOG_FORMAT]
   --organization organization, -o organization  organization to use on Terraform Cloud API [$TFCW_ORGANIZATION]
   --token token, -t token                       token to access Terraform Cloud API [$TFCW_TOKEN]
   --working-dir path, -d path                   path of the directory containing your Terraform files (default: ".") [$TFCW_WORKING_DIR]
   --workspace workspace, -w workspace           workspace to use on Terraform Cloud API [$TFCW_WORKSPACE]
   --help, -h                                    show help
```

## Install

Have a look onto the [latest release page](https://github.com/mvisonneau/tfcw/releases/latest) and pick your flavor.

Checksums are signed with the [following GPG key](https://keybase.io/mvisonneau/pgp_keys.asc): `C09C A9F7 1C5C 988E 65E3  E5FC ADEA 38ED C46F 25BE`


### Go

```bash
~$ go install github.com/mvisonneau/tfcw/cmd/tfcw@latest
```

### Homebrew

```bash
~$ brew install mvisonneau/tap/tfcw
```

### Snapcraft

```bash
~$ snap install tfcw
```

### Docker

```bash
~$ docker run -it --rm docker.io/mvisonneau/tfcw
~$ docker run -it --rm ghcr.io/mvisonneau/tfcw
~$ docker run -it --rm quay.io/mvisonneau/tfcw
```

### Scoop

```bash
~$ scoop bucket add https://github.com/mvisonneau/scoops
~$ scoop install tfcw
```

### Binaries, DEB and RPM packages

For the following ones, you need to know which version you want to install, to fetch the latest available :

```bash
~$ export TFCW_VERSION=$(curl -s "https://api.github.com/repos/mvisonneau/tfcw/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
```

```bash
# Binary (eg: freebsd/amd64)
~$ wget https://github.com/mvisonneau/tfcw/releases/download/${TFCW_VERSION}/tfcw_${TFCW_VERSION}_freebsd_amd64.tar.gz
~$ tar zxvf tfcw_${TFCW_VERSION}_freebsd_amd64.tar.gz -C /usr/local/bin

# DEB package (eg: linux/386)
~$ wget https://github.com/mvisonneau/tfcw/releases/download/${TFCW_VERSION}/tfcw_${TFCW_VERSION}_linux_386.deb
~$ dpkg -i tfcw_${TFCW_VERSION}_linux_386.deb

# RPM package (eg: linux/arm64)
~$ wget https://github.com/mvisonneau/tfcw/releases/download/${TFCW_VERSION}/tfcw_${TFCW_VERSION}_linux_arm64.rpm
~$ rpm -ivh tfcw_${TFCW_VERSION}_linux_arm64.rpm
```

## Troubleshoot

You can use the `--log-level debug` flag in order to troubleshoot

```bash
~$ tfcw --log-level debug run create
INFO[] Checking workspace configuration
INFO[] Processing variables and updating their values on TFC
DEBU[] workspace id for foo: ws-wzzmTai00qifQAxB
INFO[] Set variable credentials (terraform)
INFO[] Preparing plan
DEBU[] Workspace id for foo: ws-wzzmTai00qifQAxB
DEBU[] Configured working directory:
DEBU[] Creating configuration version..
DEBU[] Configuration version ID: cv-6qwJz000vLCx5ktH
DEBU[] Uploading configuration version..
DEBU[] Uploaded configuration version!
INFO[] Run ID: run-Uo1C0000uvMcacBg
DEBU[] Plan ID: plan-xF1C0000EiFatd65
Terraform v0.13.5
Configuring remote state backend...
Initializing Terraform configuration...
[DEBUG] Using modified User-Agent: Terraform/0.12.24 TFC/d310d4ebb1
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.
[...]

DEBU[2020-02-18T17:48:27Z] Discarding run ID: run-Uo1C0000uvMcacBg
DEBU[2020-02-18T17:48:27Z] Executed in 29.874019206s, exiting..
```

### Perform local Terraform runs

Sometimes you can find it useful to be able to run Terraform locally instead of TFC in order to troubleshoot or plan/apply
your changes in a quicker fashion. Adding the following function to your [bashrc](https://www.gnu.org/software/bash/manual/html_node/Bash-Startup-Files.html) or [zshrc](https://www.zsh.org/) could help you dynamically reprogramming
your workspace in order to try your changes locally, using your regular terraform binaries.

```bash
tfcw-local () {
  tfcw workspace operations disable > /dev/null
  tfcw render --render-type local > /dev/null
  [[ "$?" == "0" ]] && source tfcw.env
  terraform "$@"
}
```

Of course, this config is quite opinionated and tailored to specific needs so feel free to amend it as you need!

## Develop / Test

```bash
~$ make build-local
~$ ./tfcw
```

## Build / Release

If you want to build and/or release your own version of `tfcw`, you need the following prerequisites :

- [git](https://git-scm.com/)
- [golang](https://golang.org/)
- [make](https://www.gnu.org/software/make/)
- [goreleaser](https://goreleaser.com/)

```bash
~$ git clone git@github.com:mvisonneau/tfcw.git && cd tfcw

# Build the binaries locally
~$ make build-local

# Build the binaries and release them (you will need a GITHUB_TOKEN and to reconfigure .goreleaser.yml)
~$ make release
```

## Contribute

Contributions are more than welcome! Feel free to submit a [PR](https://github.com/mvisonneau/tfcw/pulls).
