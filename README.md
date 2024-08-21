# tool-installer

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

tool-installer (executable name: `tooli`) is a tool to quickly download binaries from GitHub release pages and install them into a folder.

I wrote tool-installer to automate downloading a bunch of tools from GitHub release pages because obviously having to do that manually when setting up a new computer is tedious. It always installs the latest version and can therefore also update existing tools.

## Quickstart

1. Download tool-installer from the releases page
2. Create default config: `tooli create-config`
3. Edit the [configuration file](#configuration) if needed (add/remove tools)
4. Run tool-installer: `tooli install`
5. Wait for all tools to be installed

Please see the [commands section](#commands) for more details on the individual commands.

## Configuration

The configuration for tool-installer is a simple JSON file with the following structure:

```json
{
	"install_dir": "~/.local/bin",
	"tools": {
		"tool1": {
			"binaries": [
				{
					"name": "cool-binary",
					"rename_to": ""
				}
			],
			"owner": "owner1",
			"repository": "repo1",
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip",
			"description": "Very cool tool"
		},
		"tool2": {
			"binaries": [
				{
					"name": "awesome-tool",
					"rename_to": "atx"
				}
			],
			"owner": "owner2",
			"repository": "repo2",
			"linux_asset": "x86_64-unknown-linux-musl.tar.gz",
			"windows_asset": "x86_64-pc-windows-msvc.zip",
			"description": "A tool to do stuff"
		}
	}
}
```

To change the installation directory, set the value of `install_dir` to a different path. To add or remove tools, change the entries of `tools`. Each entry of `tools` should be a struct with the entries:

- `owner`: Name of the GitHub account under which the repository is located
- `repository`: Name of the repository
- `linux_asset`: The suffix of the name of the asset to download on Linux, leave empty if the tool does not support Linux
- `windows_asset`: The suffix of the name of the asset to download on Windows, leave empty if the tool does not support Windows
- `binaries`: A list of structs where each struct has these entries:
	- `name`: Name of the file to extract
	- `rename_to`: The name which the file should have after extraction, if left empty the file is not renamed. Do _not_ include the `.exe` file ending.
- `description`: A (short) description of what the tool does

Additionally, a tool can have an entry `"asset_prefix"`. You should only set this if the suffix is not sufficient to uniquely identify the asset, e.g. when putting tools that have multiple possible binaries, for example [Hugo](https://github.com/gohugoio/hugo), in your configuration.

### Default configuration

The default configuration, which contains some commonly used tools, can be generated with `tooli create-config --path /path/to/config.json`. The `--path` option defaults to `${XDG_CONFIG_HOME}/tool-installer/config.json`.

### Acess Token

Since GitHub's API is subject to rate limits, you should create a [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token) and set that as the `GITHUB_TOKEN` environment variable. This also allows you to download from (your own) private repositories.

## Commands

tool-installer has four commands:

1. `install` (`i`)
2. `create-config` (`cc`)
3. `list`  (`l`)
4. `check` (`c`)

### `install`

The `install` command is tool-installer's primary command and used to install tools. It has 3 options:

1. `--config PATH` to specify a given file to be used as the config file (default: `~/.config/tool-installer/config.json`)
2. `--only TOOLNAME` to only install/update the named tool
3. `--timeout AMOUNT` to set the timeout for the web requests in seconds (default 10)

The `timeout` parameter's default value should work fine for most tools on normal internet connection speeds. Increase it if you have a very large tool to download or a slow connection.

**Notes:**

- tool-installer will always get the latest release from GitHub, version fixing is intentionally not supported.
- The installed version is cached at `${XDG_CACHE_HOME}/tool-installer/tool-versions.json`. If no newer version is available on GitHub releases, tool-installer will skip the tool if an attempt to install it again is made. If you uninstall a tool by deleting the binary, make sure to also remove the entry from the cache file.

### `create-config`

The `create-config` command creates a valid configuration for tool-installer, containing some commonly used tools. It only takes a single parameter, `--path PATH` (default `~/.config/tool-installer/config.json`), which can be used to specify where tool-installer should write the generated configuration file to. If the specified path already exists, tool-installer will ask you if you want to overwrite that file.

### `list`

The `list` command lists the tools specified in the configuration, sorted by tool name.

A `--long` option is available to display everything, by default the description is limited to 50 characters and the repository is omitted.

### `check`

The `check` commands downloads the latest release information from GitHub and displays for which of the installed tools an update is available.

By default it only checks the installed tools from the cache, but with the `--all` flag it will also obtain the latest release information from all tools listed in the configuration file.

## FAQ

> Why Go?

I wanted to evaluate if Go is a usable language and this project happened to fit because it is basically just doing a bunch of things which Go has a standard library package for.

> Will there be support for downloading from other websites than just GitHub?

Maybe. Depends on how many useful single binary tools are being published by other means.

> Can you add X feature?

Feel free to suggest something but most likely no, tool-installer is by design very narrow in scope. It does what I need it to do and I have no plans of going beyond that.

## License

This project is licensed under the [Apache License 2.0](LICENSE).

Licenses for the third-party tools (only the Go compiler/standard library) used by tool-installer are listed in [LICENSE-THIRD-PARTY](LICENSE-THIRD-PARTY).
