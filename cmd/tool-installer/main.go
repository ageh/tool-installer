// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "1.3.0"
const fullVersion = "tooli " + version
const userAgent = "ageh/tool-installer-" + version
const helpText = "tool-installer " + version + `

tool-installer (tooli) provides an easy way to download
all your favourite binaries from GitHub at once.

Project home page: https://github.com/ageh/tool-installer

USAGE:
    tooli [OPTIONS] <COMMAND>

COMMANDS:
    install         Installs the newest version of all tools
    create-config   Creates the default configuration
    list            Lists the tools in the configuration, sorted by name

OPTIONS:
    -h, --help      Print this help information
    -v, --version   Print version information

For more information about a specific command, try 'tooli <command> --help'.
`

const maxShortListDescriptionLength = 50

func printHelp() {
	fmt.Print(helpText)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	defaultConfigLocation, err := getConfigFilePath()
	if err != nil {
		fmt.Printf("Error obtaining default config file path: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	installCommand := flag.NewFlagSet("install", flag.ExitOnError)
	configLocation := installCommand.String("config", defaultConfigLocation, "Location of the configuration file")
	installOnly := installCommand.String("only", "", "Install only the specified tool instead of all")
	downloadTimeout := installCommand.Int("timeout", 10, "Timeout limit for requests in seconds")

	configCommand := flag.NewFlagSet("create-config", flag.ExitOnError)
	writeConfigPath := configCommand.String("path", defaultConfigLocation, "Path of the created file")

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	listConfigLocation := listCommand.String("config", defaultConfigLocation, "Location of the configuration file")
	listShort := listCommand.Bool("short", false, "Short listing only")

	switch command {
	case "-v", "--version":
		fmt.Println(fullVersion)
	case "-h", "--help":
		printHelp()
	case "install":
		installCommand.Parse(os.Args[2:])
		installTools(configLocation, installOnly, *downloadTimeout)
	case "list":
		listCommand.Parse(os.Args[2:])
		listTools(listConfigLocation, *listShort)
	case "create-config":
		configCommand.Parse(os.Args[2:])
		err := writeDefaultConfiguration(writeConfigPath)
		if err != nil {
			fmt.Println("Error:", err)
		}
	default:
		fmt.Printf("Error: Invalid command '%s'.\n\n", command)
		printHelp()
		os.Exit(1)
	}

	os.Exit(0)
}
