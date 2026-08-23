// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
)

const defaultTimeout = 30

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
    h,  help            Shows the help for the program or given command
    l,  list            Lists the tools in the configuration, sorted by name
    r,  remove          Uninstalls one or more tools and removes the config entries
    s,  status          Shows config/cache paths, tool counts, and version status
    u,  update          Updates the installed tools to the latest version

OPTIONS:
    -h, --help      Print this help information
    -v, --version   Print the version of tool-installer
    -t, --timeout   Timeout for requests to GitHub in seconds (default: 30)

Use 'tooli help <command>' for more information on a specific command.
`

const addHelp = `Adds a tool to the configuration.

Usable in two ways, the first one is adding a known tool by simply providing its name.

The second way expects the "owner/repository" GitHub slug as an argument,
optionally taking a second argument which names the tool in the configuration,
otherwise it defaults to using the repository name for the entry.
Takes as much information about the description, asset name, and binaries from GitHub as possible,
prompting the user if necessary.

Examples:
tooli add ripgrep
tooli add burntsushi/ripgrep
tooli add burntsushi/ripgrep rg`

const checkHelp = `Checks the configured tools for version updates.

By default only the currently installed tools are check, to change this pass 'all' as an argument to the command.

Examples:

tooli check
tooli check all`

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

const statusHelp = `Shows installation status.

Displays config file, cache file, and tool installation directory paths, configured and installed tool counts, and whether a new tool-installer version is available.
Pass 'verbose' to list tools that exist in the cache but not in the configuration.

Examples:
tooli status
tooli status verbose`

const updateHelp = `Updates all installed tools to their latest version.

Examples:
tooli update`

const shortHelpHelp = "Show program help"
const shortVersionHelp = "Show program version"
const shortTimeoutHelp = "Timeout for requests to GitHub"

func getCommandHelp(command string) string {
	switch command {
	case "a", "add":
		return addHelp
	case "c", "check":
		return checkHelp
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
	case "s", "status":
		return statusHelp
	case "u", "update":
		return updateHelp
	default:
		return fmt.Sprintf("Error: '%s' is not a valid command", command)
	}
}

type Arguments struct {
	commandArguments []string
	command          string
	requestTimeout   int
	showHelp         bool
	showVersion      bool
}

func (args *Arguments) argumentCount() int {
	return len(args.commandArguments)
}

func (args *Arguments) hasCommandArguments() bool {
	return args.argumentCount() > 0
}

func (args *Arguments) isArgumentCountIn(minimum int, maximum int) bool {
	n := args.argumentCount()

	if minimum > 0 {
		if maximum >= minimum {
			return minimum <= n && n <= maximum
		}

		return minimum <= n
	}

	if maximum >= 0 {
		return n <= maximum
	}

	return true
}

func versionInfo() string {
	return fmt.Sprintf("tool-installer (tooli)\nVersion: %s\nCommit hash: %s\nCompiled at: %s\nCompiled by: %s", version, commitHash, commitDate, builtBy)
}

func printHelp() {
	info := versionInfo()
	fmt.Printf("%s\n\n%s", info, helpText)
}

func parseArguments() (Arguments, error) {
	var result Arguments

	flag.BoolVar(&result.showHelp, "help", false, shortHelpHelp)
	flag.BoolVar(&result.showVersion, "version", false, shortVersionHelp)
	flag.BoolVar(&result.showVersion, "v", false, shortVersionHelp)
	flag.IntVar(&result.requestTimeout, "timeout", defaultTimeout, shortTimeoutHelp)
	flag.IntVar(&result.requestTimeout, "t", defaultTimeout, shortTimeoutHelp)

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

	if result.requestTimeout <= 0 {
		return result, errors.New("request timeout must be positive")
	}

	result.command = args[0]
	result.commandArguments = args[1:]

	return result, nil
}

func printMessages(messages []UserMessage) {
	for _, m := range messages {
		m.Print()
	}
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
		info := versionInfo()
		fmt.Println(info)
		return nil
	}

	hasArguments := args.hasCommandArguments()

	if args.command == "h" || args.command == "help" {
		if !args.isArgumentCountIn(0, 1) {
			UserMessage{Type: Warning, Tool: "tooli", Content: "'help' will only show the help for the first command provided as an argument"}.Print()
		}

		if hasArguments {
			fmt.Println(getCommandHelp(args.commandArguments[0]))
		} else {
			printHelp()
		}

		return nil
	}

	configPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to obtain the configuration file path: %w", err)
	}

	app, messages, err := newApp(configPath, args.requestTimeout)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		msg.Print()
	}

	var commandError error
	switch args.command {
	case "a", "add":
		if !args.isArgumentCountIn(1, 2) {
			return errors.New("'add' requires at least one and at most two arguments")
		}

		entryName := ""
		if args.argumentCount() > 1 {
			entryName = args.commandArguments[1]
		}
		message := app.addTool(args.commandArguments[0], entryName)
		message.Print()
	case "c", "check":
		if !args.isArgumentCountIn(0, 1) {
			return errors.New("'check' takes at most one argument")
		}

		checkAll := hasArguments && args.commandArguments[0] == "all"
		if hasArguments && !checkAll {
			return fmt.Errorf("unknown argument '%s' for 'check'", args.commandArguments[0])
		}
		messages, err := app.checkToolVersions(checkAll)
		printMessages(messages)
		commandError = err
	case "d", "delete":
		if !args.isArgumentCountIn(1, 0) {
			return errors.New("'delete' requires at least one argument")
		}

		messages, err := app.removeTools(args.commandArguments, false)
		printMessages(messages)
		commandError = err
	case "i", "install":
		commandError = app.installTools(args.commandArguments)
	case "l", "list":
		if !args.isArgumentCountIn(0, 1) {
			return errors.New("'list' takes at most one argument")
		}

		listLong := hasArguments && args.commandArguments[0] == "long"
		if hasArguments && !listLong {
			return fmt.Errorf("unknown argument '%s' for 'list'", args.commandArguments[0])
		}
		commandError = app.listTools(listLong)
	case "r", "remove":
		if !args.isArgumentCountIn(1, 0) {
			return errors.New("'remove' requires at least one argument")
		}

		messages, err := app.removeTools(args.commandArguments, true)
		printMessages(messages)
		commandError = err
	case "s", "status":
		if !args.isArgumentCountIn(0, 1) {
			return errors.New("'status' takes at most one argument")
		}

		statusVerbose := hasArguments && args.commandArguments[0] == "verbose"
		if hasArguments && !statusVerbose {
			return fmt.Errorf("unknown argument '%s' for 'status'", args.commandArguments[0])
		}
		commandError = app.showStatus(statusVerbose)
	case "u", "update":
		if !args.isArgumentCountIn(0, 0) {
			return errors.New("'update' takes no arguments")
		}

		messages, err := app.updateTools()
		printMessages(messages)
		commandError = err
	default:
		commandError = fmt.Errorf("invalid command '%s'", args.command)
	}

	return commandError
}
