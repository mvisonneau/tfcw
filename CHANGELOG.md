# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [0ver](https://0ver.org).

## [Unreleased]

## Added

- Organization and workspace configuration can now either be set directly through respective flags `--organization` / `--workspace`
- TFC configuration (address, token, organization & workspace)will default to what is configured as a remote backend in the Terraform configuration

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

[Unreleased]: https://github.com/mvisonneau/tfcw/compare/0.0.4...HEAD
[0.0.4]: https://github.com/mvisonneau/tfcw/tree/0.0.4
[0.0.3]: https://github.com/mvisonneau/tfcw/tree/0.0.3
[0.0.2]: https://github.com/mvisonneau/tfcw/tree/0.0.2
[0.0.1]: https://github.com/mvisonneau/tfcw/tree/0.0.1
