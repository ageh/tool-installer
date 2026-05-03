# tool-installer

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

tool-installer (executable name: `tooli`) is a tool to quickly download binaries from GitHub release pages and install them into a folder.

I wrote tool-installer to automate downloading a bunch of tools from GitHub release pages because obviously having to do that manually when setting up a new computer is tedious. It always installs the latest version and can therefore also update existing tools.

## Quickstart

1. Download tool-installer from the releases page
2. Set your [GitHub access token](#access-token)
3. Run `tooli list` to see which tools are configured by default and then edit the [configuration file](#configuration) if needed
4. Run `tooli install` to install all tools or provide the ones you want as arguments
5. Wait for all tools to be installed

Please see the [usage section](#usage) for more details.

## Configuration

The configuration for tool-installer is a simple JSON file with the following structure:

```json
{
	"version": 2,
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
			"asset": "x86_64-unknown-linux-gnu\\.tar\\.gz$",
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
			"asset": "x86_64-unknown-linux-musl\\.tar\\.gz$",
			"description": "A tool to do stuff"
		}
	}
}
```

The `rename_to` field is optional, if you do not need it, you can safely omit it.

To change the installation directory, set the value of `install_dir` to a different path. To add or remove tools, you can use the `add` and `remove` commands or directly change the entries in the configuration file. Each entry of `tools` should be a struct with the entries:

- `owner`: Name of the GitHub account under which the repository is located
- `repository`: Name of the repository
- `asset`: Regular expression to match the name of the asset to download
- `binaries`: A list of structs where each struct has these entries:
	- `name`: Name of the file to extract
	- `rename_to`: The name which the file should have after extraction, if left empty or omitted the file is not renamed. Do _not_ include the `.exe` file ending here, tool-installer handles this automatically on Windows.
- `description`: A (short) description of what the tool does

### Default configuration

On the first run without a configuration file present, tool-installer will use its default configuration which includes some widely used tools such as `ripgrep` and write this configuration to the disk.

### Access Token

Since GitHub's API is subject to rate limits, you should create a [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token) and set that as the `GITHUB_TOKEN` environment variable. This also allows you to download from (your own) private repositories.

## Usage

The general usage of tool-installer is `tooli [OPTIONS] COMMAND [COMMAND_ARGS]`.

tool-installer has a single global option you can use to change the timeout for the web requests in seconds (default 30): `--timeout AMOUNT`.

Additionally you can print tool-installer's version with `-v`/`--version` or the help with `-h`/`--help`.

tool-installer has the following commands (you can use the long or the short form):

1. `install` (`i`)
2. `add` (`a`)
3. `check` (`c`)
4. `delete` (`d`)
5. `help` (`h`)
6. `list`  (`l`)
7. `remove` (`r`)
8. `status` (`s`)
9. `update` (`u`)

### `install`

The `install` command is tool-installer's primary command and used to install tools. By default it installs all tools in the configuration but you can provide it with further arguments to only install the named tools.

For example `tooli install bat ripgrep` will only install `bat` and `ripgrep` (as long as both have entries in the configuration).

The `timeout` parameter's default value should work fine for most tools on normal internet connection speeds. Increase it if you have a very large tool to download or a slow connection.

**Notes:**

- tool-installer will always get the latest release from GitHub, version fixing is intentionally not supported.
- The installed versions are cached ([you can configure where exactly](#file-locations)). If no newer version is available on GitHub releases, tool-installer will skip the tool if an attempt to install it again is made. If you uninstall a tool by deleting the binary, make sure to also remove the entry from the cache file or just use the `delete` command which does both things for you.

### `add`

This command allows you to easily add a new tool to the configuration without needing to edit the configuration file directly. It takes the Github `owner/repository` slug as an argument and optionally the name the tool should have in the configuration instead of the `repository` part as a second argument.

If the name of the tool is one of the well-known tools baked into tool-installer, it will automatically be added to the configuration. If not, tool-installer will automatically try to deduce the correct asset name (regex) and binary names, prompting the user if more information is required.

### `check`

The `check` command downloads the latest release information from GitHub and shows which installed tools have updates available.

By default it only checks the installed tools from the cache, but if you pass `all` as an argument it will also obtain the latest release information from all tools listed in the configuration file.

### `delete`

By using the `delete` command, you can uninstall one or more installed tools. It will remove the binaries and the cache entries, but keeps the configuration entries so you can easily install the tools later again. If you also want the configuration entries to be deleted, use the `remove` command instead.

### `help`

This shows the help for the entire program or the specified command. Use `tooli help` to display the general help and `tooli help <COMMAND>` to show the more specific help for individual commands.

### `list`

The `list` command lists the tools specified in the configuration, sorted by tool name.

If you pass `long` as an argument it switches to long mode, by default the description is limited to 50 characters and the repository is omitted.

### `remove`

This command is the exact opposite of the `add` command and allows you to fully uninstall installed tools, including their configuration entries. If you only want to uninstall tools but keep their configuration entries, use the `delete` command instead.

### `status`

This command displays some information about tool-installer and the configuration in use. It shows the paths for the configuration file, the cache file, and the directory in which tools are installed. It also shows how many tools are listed in the configuration and how many are installed (based on cache entries). Finally it checks and displays whether a newer version of tool-installer is available.
Pass `verbose` to list tools that exist in the cache but not in the configuration.

### `update`

This command is basically a shorthand for `tooli check` followed by `tooli install` (with the tools in need of an update as arguments). It will update all currently installed tools to their latest version. Skips tools which are already up to date.

## File Locations

You can fully configure the directories in which tool-installer stores the cache, configuration, and installed tools. No other directories will ever be touched by tool-installer.

The directory for the installed tools is configured in the configuration, using the `install_dir` field. For the cache and configuration directories it can be configured via environment variables and fall back to platform-specific default values if those are not set or empty. Typically you should be fine with the defaults.

|**Priority**|**Cache**|**Configuration**|
|:-:|:-:|:-:|
|Highest|`$TOOLI_CACHE_DIRECTORY`|`$TOOLI_CONFIG_DIRECTORY`|
|Second|`$XDG_CACHE_HOME/tool-installer`|`$XDG_CONFIG_HOME/tool-installer`|
|Default (Linux)|`$HOME/.cache/tool-installer`|`$HOME/.config/tool-installer`|
|Default (Windows)|`%LOCALAPPDATA%/tool-installer`|`%APPDATA%/tool-installer`|

Please note that tool-installer supports the XDG variables on all platforms because this useful convention should be usable everywhere. There is no reason to exclude Windows from it even though most programs seem to think otherwise.

## FAQ

> Why Go?

I wanted to evaluate if Go is a usable language and this project happened to fit because it is basically just doing a bunch of things which Go has a standard library package for. Since then I have decided to use a single third party dependency for handling `tar.xz` assets because unfortunately some nice tools like typst use this and it is not supported by the standard library.

> Will there be support for downloading from other websites than just GitHub?

Maybe. Depends on how many useful single binary tools are being published by other means.

> Can you add X feature?

Feel free to suggest something but most likely no, tool-installer is by design very narrow in scope. It does what I need it to do and I have no plans of going beyond that.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
