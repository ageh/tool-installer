# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Asset integrity is checked via the provided SHA256, which [GitHub now provides automatically](https://github.blog/changelog/2025-06-03-releases-now-expose-digests-for-release-assets/)

## [2.0.0] - 2025-06-09

### Added

- `help` command instead of `--help` flag on every command
- `add` command to add a new tool without having to directly edit the configuration
- `delete` command to uninstall the binaries of tools but keep the configuration entries
- `remove` command to uninstall the binaries of tools and remove the configuration entries
- `update` command as a shortcut for `check` followed by `install` for the tools needing an update

### Changed

- `--config` and `--timeout` are now global flags instead of having to be passed for each command
- Downloads are now parallel
- Printing the version now also shows the commit hash and time of compilation
- Output tables now have borders
- Asset names are now regular expressions to allow for better selection

### Removed

- Removed `dust`, `fzf`, and `hexyl` from the default configuration

## [1.5.0] - 2024-08-21

### Added

- Windows binary for `sd` in default config
- Short forms for commands (e.g. `l` for `list`)
- An _optional_ `asset_prefix` entry in the config for tools that provide multiple possible binaries

### Changed

- `list` command now defaults to short form

## [1.4.0] - 2024-06-08

### Added

- The version of the currently installed tools is now cached
- Added a `check` command that checks for newer releases
- The default configuration file location now follows the XDG specification
- The `list` command displays the installed version (if available)

### Fixed

- The short option of the `list` command now properly truncates the description

## [1.3.0] - 2024-04-20

### Added

- The `list` command now has a `-short` option which only displays the name and description and limits the description to 50 characters
- `bat` was added to the default configuration

### Changed

- Replaced `exa` with `eza` in the default configuration

## [1.2.3] - 2023-06-18

### Added

- `fd` tool was added to the default configuration

### Changed

- Improved user-facing messages

### Fixed

- Assets that are not archives are now named correctly
- Description of `sd` in the default configuration

## [1.2.2] - 2023-06-03

### Fixed

- The binary names now correctly handle the `.exe` ending on Windows

## [1.2.1] - 2023-06-03

### Fixed

- The path to load the configuration now properly works with `~` paths

## [1.2.0] - 2023-06-03

### Added

- `description` field in the configuration, to help with remembering what each tool does
- `list` command now also displays the tool's description

## [1.1.0] - 2023-06-03

### Added

- `list` command to display tools listed in the configuration, sorted by name

## [1.0.0] - 2023-06-01

### Added

- First working version
