// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
)

const version = "1.5.0"
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

func getFullVersion() string {
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

	return fmt.Sprintf("tool-installer (tooli)\nVersion: %s\nCommit hash: %s\nCompiled at: %s", version, revision, timeStamp)
}

func printHelp() {
	fmt.Print(helpText)
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
		fmt.Println(getFullVersion())
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

	config, err := readConfiguration(args.configPath)
	if err != nil {
		return fmt.Errorf(`could not load configuration: '%w'
Check if the configuration file exists and is valid.
You can generate a new configuration file with 'tooli create-config'`, err)
	}

	switch args.command {
	case "i", "install":
		err = installTools(config, args.commandArguments, args.requestTimeout)
	case "l", "list":
		listLong := hasArguments && args.commandArguments[0] == "long"
		err = listTools(config, listLong)
	case "c", "check":
		checkAll := hasArguments && args.commandArguments[0] == "all"
		err = checkToolVersions(config, checkAll, args.requestTimeout)
	case "u", "update":
		err = updateTools(config, args.requestTimeout)
	default:
		err = fmt.Errorf("invalid command '%s'", args.command)
	}

	return err
}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("Error: %v.\n\n", err)
		fmt.Println("See `tooli help` for instructions.")
		os.Exit(1)
	}

	os.Exit(0)
}
