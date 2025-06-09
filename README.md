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

Please see the [usage section](#usage) for more details.

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
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
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
			"linux_asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"windows_asset": "x86_64-pc-windows-msvc\\.zip$",
			"description": "A tool to do stuff"
		}
	}
}
```

To change the installation directory, set the value of `install_dir` to a different path. To add or remove tools, you can use the `add` and `remove` commands or directly change the entries in the configuration file. Each entry of `tools` should be a struct with the entries:

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

The default configuration, which contains some commonly used tools, can be generated with `tooli create-config [/path/to/config.json]`. By default it writes to `${XDG_CONFIG_HOME}/tool-installer/config.json`.

### Acess Token

Since GitHub's API is subject to rate limits, you should create a [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token) and set that as the `GITHUB_TOKEN` environment variable. This also allows you to download from (your own) private repositories.

## Usage

The general usage of tool-installer is `tooli [OPTIONS] COMMAND [COMMAND_ARGS]

tool-installer comes with two options you can use to influence the behaviour of the commands:

1. `--config PATH` to specify a given file to be used as the config file (default: `${XDG_CONFIG_HOME}/tool-installer/config.json`)
2. `--timeout AMOUNT` to set the timeout for the web requests in seconds (default 10)

Additionally you can print tool-installer's version with `-v`/`--version` or the help with `-h`/`--help`

tool-installer has the following commands (you can use the long or the short form):

1. `install` (`i`)
2. `add` (`a`)
3. `check` (`c`)
4. `create-config` (`cc`)
5. `delete` (`d`)
6. `help` (`h`)
7. `list`  (`l`)
8. `remove` (`r`)
9. `update` (`u`)

### `install`

The `install` command is tool-installer's primary command and used to install tools. By default it installs all tools in the configuration but you can provide it with further arguments to only install the named tools.

For example `tooli install bat ripgrep` will only install `bat` and `ripgrep` (as long as both have entries in the configuration).

The `timeout` parameter's default value should work fine for most tools on normal internet connection speeds. Increase it if you have a very large tool to download or a slow connection.

**Notes:**

- tool-installer will always get the latest release from GitHub, version fixing is intentionally not supported.
- The installed version is cached at `${XDG_CACHE_HOME}/tool-installer/tool-versions.json`. If no newer version is available on GitHub releases, tool-installer will skip the tool if an attempt to install it again is made. If you uninstall a tool by deleting the binary, make sure to also remove the entry from the cache file or just use the `delete` command which does both things for you.

### `add`

This command opens a prompt which allows you to enter a new tool entry to the configuration in case you do not want to edit the configuration file directly.

### `check`

The `check` commands downloads the latest release information from GitHub and displays for which of the installed tools an update is available.

By default it only checks the installed tools from the cache, but if you pass `all` as an argument it will also obtain the latest release information from all tools listed in the configuration file.

### `create-config`

The `create-config` command creates a valid configuration for tool-installer, containing some commonly used tools. It only takes a single parameter, `--path PATH` (default `~/.config/tool-installer/config.json`), which can be used to specify where tool-installer should write the generated configuration file to. If the specified path already exists, tool-installer will ask you if you want to overwrite that file.

### `delete`

By using the `delete` command, you can uninstall one or more installed tools. It will remove the binaries and the cache entries, but keeps the configuration entries so you can easily install the tools later again. If you also want the configuration entries to be deleted, use the `remove` command instead.

### `help`

This shows the help for the entire program or the specified command. Use `tooli help` to display the general help and `tooli help <COMMAND>` to show the more specific help for individual commands.

### `list`

The `list` command lists the tools specified in the configuration, sorted by tool name.

If you pass `long` as an argument it switches to long mode, by default the description is limited to 50 characters and the repository is omitted.

### `remove`

This command is the exact opposite of the `add` command and allows you to fully uninstall installed tools, including their configuration entries. If you only want to uninstall tools but keep their configuration entries, use the `delete` command instead.

### `update`

This command is basically a shorthand for `tooli check` followed by `tooli install` (with the tools in need of an update as arguments). It will update all currently installed tools to their latest version. Skips tools which are already up to date.


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
