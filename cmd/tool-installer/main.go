// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

const version = "1.5.0"
const fullVersion = "tooli " + version
const userAgent = "ageh/tool-installer-" + version
const helpText = "tool-installer " + version + `

tool-installer (tooli) provides an easy way to download
all your favourite binaries from GitHub at once.

Project home page: https://github.com/ageh/tool-installer

USAGE:
    tooli [OPTIONS] <COMMAND> [COMMAND_ARGS...]

COMMANDS:
    i,  install         Installs the newest version of all or the selected tools
    c,  check           Checks and displays available updates
    cc, create-config   Creates the default configuration
    h,  help            Shows the help for the program or given command
    l,  list            Lists the tools in the configuration, sorted by name
    u,  update          Updates the installed tools to the latest version

OPTIONS:
    -h, --help      Print this help information
    -v, --version   Print the version of tool-installer
    -c, --config    Specify from where to read the configuration (default: ~/.config/tool-installer/config.json)
    -t, --timeout   Timeout for requests to GitHub in seconds (default: 10)

For more information about a specific command, try 'tooli help <command>'.
`

const maxShortListDescriptionLength = 50

func printHelp() {
	fmt.Print(helpText)
}

func printConfigError(err error) {
	fmt.Printf("Error: could not load configuration: '%v'\n", err)
	fmt.Println("Check if the configuration file is valid.")
	fmt.Println("You can generate a new configuration file with 'tooli create-config'.")
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

func parseArguments() (Arguments, error) {
	var result Arguments
	defaultConfigLocation, err := getConfigFilePath()
	if err != nil {
		return result, err
	}

	flag.StringVar(&result.configPath, "config", defaultConfigLocation, "Location of the configuration file")
	flag.BoolVar(&result.showHelp, "help", false, "Show program help")
	flag.BoolVar(&result.showVersion, "version", false, "Show program version")
	flag.IntVar(&result.requestTimeout, "timeout", 10, "Timeout for requests to GitHub")

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

func run() int {
	args, err := parseArguments()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return 1
	}

	if args.showHelp {
		printHelp()
		return 0
	}

	if args.showVersion {
		fmt.Println(fullVersion)
		return 0
	}

	config, err := getConfig(args.configPath)
	if err != nil {
		printConfigError(err)
		return 1
	}

	hasArguments := args.hasCommandArguments()

	switch args.command {
	case "h", "help":
		if hasArguments {
			fmt.Println(getCommandHelp(args.commandArguments[0]))
		} else {
			printHelp()
		}
	case "i", "install":
		err = installTools(config, args.commandArguments, args.requestTimeout)
	case "l", "list":
		listLong := hasArguments && args.commandArguments[0] == "long"
		err = listTools(config, listLong)
	case "cc", "create-config":
		configWritePath := args.configPath
		if hasArguments {
			configWritePath = args.commandArguments[0]
		}
		err = writeDefaultConfiguration(configWritePath)
	case "c", "check":
		checkAll := hasArguments && args.commandArguments[0] == "all"
		err = checkToolVersions(config, checkAll, args.requestTimeout)
	case "u", "update":
		err = updateTools(config, args.requestTimeout)
	default:
		fmt.Printf("Error: Invalid command '%s'.\n\n", args.command)
		printHelp()
		return 1
	}

	if err != nil {
		fmt.Println("Error:", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run())
}
