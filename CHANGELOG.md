# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org).

## [Unreleased]

### Added

- Release GitHub container registry based images: [ghcr.io/mvisonneau/tfcw](https://github.com/users/mvisonneau/packages/container/package/tfcw)
- Release `arm64v8` based container images as part of docker manifests in both **docker.io** and **ghcr.io**
- GPG sign released artifacts checksums

### Changed

- Updated all dependencies / Terraform to `0.14.2`
- Migrated CI from Drone to GitHub actions

## [v0.0.10] - 2020-10-30

### Changed

- Prefix new releases with `^v` to make pkg.go.dev happy
- Upgraded all dependencies
- Refactored codebase following golang file structure
- Fixed goreleaser config

## [0.0.9] - 2020-09-14

### Added

- gosec tests

### Changed

- default branch from `master` to `main`
- Bumped urfave/cli to v2
- upgraded all dependencies to their latest version, in particular:
  - go to **1.15**
  - terraform to **0.13.2**
  - go-tfe to **0.10.1**
- fixed some missed error handlings

## [0.0.8] - 2020-04-14

### Added

- New CLI command to be able to remove variables configured on the workspace: `workspace delete-variables`
- A couple commands to enable/disable the operations on the workspace from the CLI

### Changed

- Renamed some CLI commands & helpers

## [0.0.7] - 2020-04-09

### Changed

- **provider/vault** - **Actually** fixed a bug on the default vault token validation method
- Fixed a panic occurring when defaults.var block was not defined
- Fixed a bug on the default definition of the HCL flag for the variable
- Prevent variables from being removed whilst using var.ttl field and tfc.purge-unmanaged-variables at the same time
- Prevent variables from being removed when using tfc.purge-unmanaged-variables with a Vault multi-key variable
- Better testing for the variable ttl management and couple bugfixes
- Automated the variable refresh (& removal if empty) when their TTL is modified

## [0.0.6] - 2020-04-07

### Added

- **provider/vault** - Added support for kv-v2 secret engine
- Return an error on run creation if remote runs are not enabled on the workspace
- Do not prompt for approval and follow the apply logs if the workspace is configured with AutoApply
- Validate there is no pending run on the workspace before attempting to create a new one

### Changed

- **provider/vault** - Fixed a bug on the default vault token validation method
- Refactored part of the CLI around the render function
- Corrected the expected/actual variables ordering in tests
- Ordered commands, subcommands and flags alphabetically in the CLI helpers
- Fixed some regression/bug introduced on the ConfigureWorkspace function in `0.0.5` preventing it from working properly

## [0.0.5] - 2020-04-06

### Added

- Organization and workspace configuration can now either be set directly through respective flags `--organization` / `--workspace`
- TFC configuration (address, token, organization & workspace)will default to what is configured as a remote backend in the Terraform configuration
- Defaults configuration capabilities for variable `sensitive` & `hcl` fields
- A `ttl` field on variables which makes TFCW only update some variables when their Time To Live has been exceeded. This results in much faster syncs
- Speed improvements by reducing the amount of calls to fetch the workspace ID. Combined with the TTL option, when nothing is needed the overhead is now only about a second

## [0.0.4] - 2020-04-01

### Added

- Capability to manage the workspace SSH key
- Added a timeout flag to automatically exit if the run is too long to start

### Changed

- Reviewed the examples and configuration syntax documentation
- Bumped go modules versions

## [0.0.3] - 2020-03-21

### Added

- Support for managing some workspace configuration within the tfcw.hcl file

### Changed

- Bumped to go 1.14 / goreleaser 0.129.0
- Fixed a bug preventing the Vault provider from working properly when using multiple values
- Fixed a bug preventing errors from being returned on provider deciphering failures

## [0.0.2] - 2020-02-27

### Added

- Custom name for runs
- Workspace status and current-run-id commands
- Refactored the CLI for creating runs
- Added standalone commands for approving or discarding runs
- Covered ~40% of the codebase with unit tests
- Added the possibility to export the runID into a file when created

## [0.0.1] - 2020-02-18

### Added

- Read configuration form HCL (or json) file
- Fetch sensitive values from 3 providers : `vault`, `s5` and `environment variables`
- Plan and apply Terraform stacks
- dry-run feature on render function
- purge unmanaged variables

[Unreleased]: https://github.com/mvisonneau/tfcw/compare/v0.0.10...HEAD
[v0.0.10]: https://github.com/mvisonneau/tfcw/tree/v0.0.10
[0.0.9]: https://github.com/mvisonneau/tfcw/tree/0.0.9
[0.0.8]: https://github.com/mvisonneau/tfcw/tree/0.0.8
[0.0.7]: https://github.com/mvisonneau/tfcw/tree/0.0.7
[0.0.6]: https://github.com/mvisonneau/tfcw/tree/0.0.6
[0.0.5]: https://github.com/mvisonneau/tfcw/tree/0.0.5
[0.0.4]: https://github.com/mvisonneau/tfcw/tree/0.0.4
[0.0.3]: https://github.com/mvisonneau/tfcw/tree/0.0.3
[0.0.2]: https://github.com/mvisonneau/tfcw/tree/0.0.2
[0.0.1]: https://github.com/mvisonneau/tfcw/tree/0.0.1
