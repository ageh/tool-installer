// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "1.0.0"
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

OPTIONS:
    -h, --help      Print this help information
    -v, --version   Print version information

For more information about a specific command, try 'tooli <command> --help'.
`

const defaultConfigLocation = "~/.config/tool-installer/config.json"

func printHelp() {
	fmt.Println(helpText)
}

func makeOutputDirectory(path *string) error {
	return os.MkdirAll(*path, 0755)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	installCommand := flag.NewFlagSet("install", flag.ExitOnError)
	configLocation := installCommand.String("config", defaultConfigLocation, "Location of the configuration file")
	installOnly := installCommand.String("only", "", "Install only the specified tool instead of all")
	downloadTimeout := installCommand.Int("timeout", 10, "Timeout limit for requests in seconds")

	configCommand := flag.NewFlagSet("create-config", flag.ExitOnError)
	defaultConfigFileName := configCommand.String("path", defaultConfigLocation, "Path of the created file")

	switch command {
	case "-v", "--version":
		fmt.Println(fullVersion)
		os.Exit(0)
	case "-h", "--help":
		printHelp()
		os.Exit(0)
	case "install":
		installCommand.Parse(os.Args[2:])
	case "create-config":
		configCommand.Parse(os.Args[2:])
		err := writeDefaultConfiguration(defaultConfigFileName)
		if err != nil {
			fmt.Println("Error:", err)
		}
		os.Exit(0)
	default:
		fmt.Printf("Error: Invalid command '%s'.\n\n", command)
		printHelp()
		os.Exit(1)
	}

	config, err := GetConfig(*configLocation)
	if err != nil {
		fmt.Printf("Error: Could not load configuration: %v\n", err)
		fmt.Println("Check if the configuration file is valid.")
		fmt.Println("You can generate a new configuration file with `tooli create-config`")
		os.Exit(1)
	}

	err = makeOutputDirectory(&config.InstallationDirectory)
	if err != nil {
		fmt.Printf("Error: Could not create output directory %v.\n", config.InstallationDirectory)
		os.Exit(1)
	}

	downloader := newDownloader(*downloadTimeout)

	if *installOnly != "" {
		err = downloader.downloadTool(*installOnly, &config)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	} else {
		for k, _ := range config.Tools {
			fmt.Printf("Installing tool %s\n", k)
			err = downloader.downloadTool(k, &config)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	}

	os.Exit(0)
}
