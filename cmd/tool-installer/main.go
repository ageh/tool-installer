// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

const version = "1.2.3"
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

const defaultConfigLocation = "~/.config/tool-installer/config.json"

func printHelp() {
	fmt.Println(helpText)
}

func printConfigError(err error) {
	fmt.Printf("Error: Could not load configuration: %v.\n", err)
	fmt.Println("Check if the configuration file is valid.")
	fmt.Println("You can generate a new configuration file with 'tooli create-config'.")
}

func makeOutputDirectory(path *string) error {
	return os.MkdirAll(*path, 0755)
}

type TableEntry struct {
	Name        string
	Link        string
	Description string
}

type ByName []TableEntry

func (t ByName) Len() int {
	return len(t)
}

func (t ByName) Less(i int, j int) bool {
	return t[i].Name < t[j].Name
}

func (t ByName) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}

func max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

func listTools(configLocation *string) {
	config, err := GetConfig(*configLocation)
	if err != nil {
		printConfigError(err)
		os.Exit(1)
	}

	// Minimum sizes based on header line
	nameSize := 4
	linkSize := 16
	descriptionSize := 11

	tmp := make([]TableEntry, len(config.Tools))

	i := 0
	for k, v := range config.Tools {
		tmp[i] = TableEntry{Name: k, Link: fmt.Sprintf("%s/%s", v.Owner, v.Repository), Description: v.Description}

		nameSize = max(nameSize, len(k))
		linkSize = max(linkSize, len(tmp[i].Link))
		descriptionSize = max(descriptionSize, len(v.Description))

		i++
	}

	sort.Sort(ByName(tmp))

	fmt.Printf("%-*s    %-*s    %-*s\n\n", nameSize, "Name", linkSize, "Owner/Repository", descriptionSize, "Description")

	for _, j := range tmp {
		fmt.Printf("%-*s    %-*s    %-*s\n", nameSize, j.Name, linkSize, j.Link, descriptionSize, j.Description)
	}
}

func installTools(configLocation *string, installOnly *string, downloadTimeout int) {
	config, err := GetConfig(*configLocation)
	if err != nil {
		printConfigError(err)
		os.Exit(1)
	}

	err = makeOutputDirectory(&config.InstallationDirectory)
	if err != nil {
		fmt.Printf("Error: Could not create output directory %v.\n", config.InstallationDirectory)
		os.Exit(1)
	}

	downloader := newDownloader(downloadTimeout)

	if *installOnly != "" {
		fmt.Printf("Installing tool '%s'.\n", *installOnly)
		err = downloader.downloadTool(*installOnly, &config)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	} else {
		for k, _ := range config.Tools {
			fmt.Printf("Installing tool '%s'.\n", k)
			err = downloader.downloadTool(k, &config)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	}
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

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	listConfigLocation := listCommand.String("config", defaultConfigLocation, "Location of the configuration file")

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
		listTools(listConfigLocation)
	case "create-config":
		configCommand.Parse(os.Args[2:])
		err := writeDefaultConfiguration(defaultConfigFileName)
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
