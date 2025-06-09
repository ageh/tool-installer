// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
	"runtime/debug"
)

const version = "2.0.0"

const helpText = `tool-installer (tooli) provides an easy way to download
all your favourite binaries from GitHub at once.

Project home page: https://github.com/ageh/tool-installer

USAGE:
    tooli [OPTIONS] <COMMAND> [COMMAND_ARGS...]

COMMANDS:
    i,  install         Installs the newest version of all or the selected tools
    a,  add             Adds a new tool to the configuration
    c,  check           Checks and displays available updates
    d,  delete          Uninstalls one or more tools but keeps the config entry
    cc, create-config   Creates the default configuration
    h,  help            Shows the help for the program or given command
    l,  list            Lists the tools in the configuration, sorted by name
    r,  remove          Uninstalls one or more tools and removes the config entries
    u,  update          Updates the installed tools to the latest version

OPTIONS:
    -h, --help      Print this help information
    -v, --version   Print the version of tool-installer
    -c, --config    Specify from where to read the configuration (default: ~/.config/tool-installer/config.json)
    -t, --timeout   Timeout for requests to GitHub in seconds (default: 10)

Use 'tooli help <command>' for more information on a specific command.
`

const addHelp = `Adds a tool to the configuration by prompting the necessary values from the user.

Examples:
tooli add ripgrep
tooli add bat`

const checkHelp = `Checks the configured tools for version updates.

By default only the currently installed tools are check, to change this pass 'all' as an argument to the command.

Examples:

tooli check
tooli check all`

const createConfigHelp = `Creates the default configuration.

By default the configuration is written to '~/.config/tool-installer/config.json',
but this can be changed by passing a different path as an argument.

Examples:

tooli create-config
tooli create-config test.json`

const deleteHelp = `Uninstalls one or more tools.

Removes the binaries but keeps the configuration entry. To also remove the entry, use the 'remove' command.


Examples:
tooli delete ripgrep
tooli delete ripgrep bat micro`

const helpHelp = `Shows the help for the program or command.

Examples:
tooli help
tooli help install`

const installHelp = `Installs tools. If no arguments are provided, it installs all tools in the configuration.
Installs only the named tools if provided with a space separated list of tools to install.

Examples

tooli install
tooli install ripgrep
tooli install ripgrep eza bat fd`

const listHelp = `Lists the tools present in the configuration.

Examples:

tooli list
tooli list long`

const removeHelp = `Uninstalls one or more tools.

WARNING: This command also removes the configuration entry.
To only uninstall the binaries but keep the configuration entry, use the 'delete' command.

Examples:
tooli remove ripgrep
tooli remove ripgrep bat micro`

const updateHelp = `Updates all installed tools to their latest version.

Examples:
tooli update`

func getCommandHelp(command string) string {
	switch command {
	case "a", "add":
		return addHelp
	case "c", "check":
		return checkHelp
	case "cc", "create-config":
		return createConfigHelp
	case "d", "delete":
		return deleteHelp
	case "h", "help":
		return helpHelp
	case "i", "install":
		return installHelp
	case "l", "list":
		return listHelp
	case "r", "remove":
		return removeHelp
	case "u", "update":
		return updateHelp
	default:
		return fmt.Sprintf("Error: '%s' is not a valid command", command)
	}
}

type CompileInfo struct {
	revision  string
	timeStamp string
}

func getCompileInfo() CompileInfo {
	revision := "No VCS info available"
	timeStamp := "No VCS info available"

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
			} else if setting.Key == "vcs.time" {
				timeStamp = setting.Value
			}
		}
	}

	return CompileInfo{revision: revision, timeStamp: timeStamp}
}

type Arguments struct {
	commandArguments []string
	command          string
	configPath       string
	requestTimeout   int
	showHelp         bool
	showVersion      bool
}

func (args *Arguments) hasCommandArguments() bool {
	return len(args.commandArguments) > 0
}

func printHelp() {
	info := getCompileInfo()
	fmt.Printf("tool-installer (tooli)\nVersion: %s\nCommit hash: %s\nCompiled at: %s\n\n%s", version, info.revision, info.timeStamp, helpText)
}

func parseArguments() (Arguments, error) {
	var result Arguments
	defaultConfigLocation, err := getConfigFilePath()
	if err != nil {
		return result, err
	}

	flag.StringVar(&result.configPath, "config", defaultConfigLocation, "Location of the configuration file")
	flag.StringVar(&result.configPath, "c", defaultConfigLocation, "Location of the configuration file")
	flag.BoolVar(&result.showHelp, "help", false, "Show program help")
	flag.BoolVar(&result.showVersion, "version", false, "Show program version")
	flag.BoolVar(&result.showVersion, "v", false, "Show program version")
	flag.IntVar(&result.requestTimeout, "timeout", 10, "Timeout for requests to GitHub")
	flag.IntVar(&result.requestTimeout, "t", 10, "Timeout for requests to GitHub")

	// Override by default existing -h to produce the same effect as --help
	flag.Usage = printHelp

	flag.Parse()

	if result.showHelp || result.showVersion {
		return result, nil
	}

	args := flag.Args()
	if len(args) < 1 {
		return result, errors.New("missing command")
	}

	result.command = args[0]
	result.commandArguments = args[1:]

	return result, nil
}

func run() error {
	args, err := parseArguments()
	if err != nil {
		return err
	}

	if args.showHelp {
		printHelp()
		return nil
	}

	if args.showVersion {
		info := getCompileInfo()
		fmt.Printf("tool-installer (tooli)\nVersion: %s\nCommit hash: %s\nCompiled at: %s", version, info.revision, info.timeStamp)
		return nil
	}

	hasArguments := args.hasCommandArguments()

	if args.command == "h" || args.command == "help" {
		if hasArguments {
			fmt.Println(getCommandHelp(args.commandArguments[0]))
		} else {
			printHelp()
		}

		return nil
	}

	if args.command == "cc" || args.command == "create-config" {
		configWritePath := args.configPath
		if hasArguments {
			configWritePath = args.commandArguments[0]
		}
		return writeDefaultConfiguration(configWritePath)
	}

	app, err := newApp(args.configPath, args.requestTimeout)
	if err != nil {
		return err
	}

	switch args.command {
	case "a", "add":
		err = app.addTool()
	case "c", "check":
		checkAll := hasArguments && args.commandArguments[0] == "all"
		err = app.checkToolVersions(checkAll)
	case "d", "delete":
		if !hasArguments {
			err = fmt.Errorf("you need to provide at least one tool name as argument to 'delete'")
		} else {
			err = app.removeTools(args.commandArguments, false)
		}
	case "i", "install":
		err = app.installTools(args.commandArguments)
	case "l", "list":
		listLong := hasArguments && args.commandArguments[0] == "long"
		err = app.listTools(listLong)
	case "r", "remove":
		if !hasArguments {
			err = fmt.Errorf("you need to provide at least one tool name as argument to 'remove'")
		} else {
			err = app.removeTools(args.commandArguments, true)
		}
	case "u", "update":
		err = app.updateTools()
	default:
		err = fmt.Errorf("invalid command '%s'", args.command)
	}

	return err
}
