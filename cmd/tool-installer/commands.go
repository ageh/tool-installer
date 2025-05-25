// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"sort"
)

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
const updateHelp = `Updates all installed tools to their latest version.

Examples:
tooli update`

type TableEntry struct {
	Name        string
	Link        string
	Description string
	Version     string
}

func (t TableEntry) GetName() string {
	return t.Name
}

type VersionTableEntry struct {
	Name      string
	Installed string
	Available string
}

func (v VersionTableEntry) GetName() string {
	return v.Name
}

type Named interface {
	GetName() string
}

// Define a generic type that implements sort.Interface for any slice of Named
type ByName[T Named] struct {
	data []T
}

func (array ByName[T]) Len() int {
	return len(array.data)
}

func (array ByName[T]) Less(i int, j int) bool {
	return array.data[i].GetName() < array.data[j].GetName()
}

func (array ByName[T]) Swap(i int, j int) {
	array.data[i], array.data[j] = array.data[j], array.data[i]
}

func max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

func getCommandHelp(command string) string {
	switch command {
	case "c", "check":
		return checkHelp
	case "cc", "create-config":
		return createConfigHelp
	case "i", "install":
		return installHelp
	case "l", "list":
		return listHelp
	case "h", "help":
		return helpHelp
	case "u", "update":
		return updateHelp
	default:
		return fmt.Sprintf("Error: '%s' is not a valid command", command)
	}
}

func getOutdatedTools(config Configuration, checkAll bool, downloadTimeout int, cache Cache) ([]VersionTableEntry, error) {
	downloader := newDownloader(downloadTimeout)

	var nTools int
	if checkAll {
		nTools = len(config.Tools)
	} else {
		nTools = len(cache.Tools)
	}

	tmp := make([]VersionTableEntry, nTools)

	if checkAll {
		i := 0
		for name, tool := range config.Tools {
			release, err := downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				fmt.Printf("Error: failed to obtain latest release of tool '%s': %v\n", name, err)
				continue
			}

			tmp[i] = VersionTableEntry{Name: name, Installed: "", Available: release.TagName}

			if current, found := cache.Tools[name]; found {
				tmp[i].Installed = current
			}

			i++
		}
	} else {
		i := 0
		for name, version := range cache.Tools {
			tool := config.Tools[name]
			release, err := downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				fmt.Printf("Error: failed to obtain latest release of tool '%s': %v\n", name, err)
				continue
			}

			tmp[i] = VersionTableEntry{Name: name, Installed: version, Available: release.TagName}

			i++
		}
	}

	sort.Sort(ByName[VersionTableEntry]{tmp})

	results := make([]VersionTableEntry, 0)
	for _, entry := range tmp {
		if entry.Installed != entry.Available {
			results = append(results, entry)
		}
	}

	return results, nil
}

func checkToolVersions(config Configuration, checkAll bool, downloadTimeout int) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	results, err := getOutdatedTools(config, checkAll, downloadTimeout, cache)
	if err != nil {
		return err
	}

	nameSize := 4
	installedSize := 9
	availableSize := 9

	for _, r := range results {
		nameSize = max(nameSize, len(r.Name))
		installedSize = max(installedSize, len(r.Installed))
		availableSize = max(availableSize, len(r.Available))
	}

	if len(results) > 0 {
		fmt.Printf("%-*s    %-*s    %-*s\n\n", nameSize, "Name", installedSize, "Installed", availableSize, "Available")

		for _, e := range results {
			fmt.Printf("%-*s    %-*s    %-*s\n", nameSize, e.Name, installedSize, e.Installed, availableSize, e.Available)
		}
	} else {
		fmt.Println("All tools are up to date.")
	}

	return nil
}

func listTools(config Configuration, longList bool) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	// Minimum sizes based on header line
	nameSize := 4
	linkSize := 16
	descriptionSize := 11
	versionSize := 7

	tmp := make([]TableEntry, len(config.Tools))

	i := 0
	for k, v := range config.Tools {
		tmp[i] = TableEntry{Name: k, Link: fmt.Sprintf("%s/%s", v.Owner, v.Repository), Description: v.Description, Version: ""}

		if version, found := cache.Tools[k]; found {
			tmp[i].Version = version
		}

		nameSize = max(nameSize, len(k))
		linkSize = max(linkSize, len(tmp[i].Link))
		descriptionSize = max(descriptionSize, len(v.Description))
		versionSize = max(versionSize, len(tmp[i].Version))

		i++
	}

	sort.Sort(ByName[TableEntry]{tmp})

	if longList {
		fmt.Printf("%-*s    %-*s    %-*s    %-*s\n\n", nameSize, "Name", linkSize, "Owner/Repository", descriptionSize, "Description", versionSize, "Version")

		for _, j := range tmp {
			fmt.Printf("%-*s    %-*s    %-*s    %-*s\n", nameSize, j.Name, linkSize, j.Link, descriptionSize, j.Description, versionSize, j.Version)
		}
	} else {
		descriptionSize = min(descriptionSize, maxShortListDescriptionLength)
		fmt.Printf("%-*s    %-*s       %-*s\n\n", nameSize, "Name", descriptionSize, "Description", versionSize, "Version")

		for _, j := range tmp {
			extra := "   "
			if len(j.Description) > maxShortListDescriptionLength {
				extra = "..."
				j.Description = j.Description[:maxShortListDescriptionLength]
			}
			fmt.Printf("%-*s    %-*s%s    %-*s\n", nameSize, j.Name, descriptionSize, j.Description, extra, versionSize, j.Version)
		}
	}

	return nil
}

func makeOutputDirectory(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("error creating output directory ('%s'): %w", path, err)
	}

	return nil
}

func installFiles(name string, binaries []Binary, result DownloadResult, installationDirectory string, cache *Cache) error {
	err := extractFiles(result.data, result.assetName, binaries, installationDirectory)
	if err != nil {
		return fmt.Errorf("error during tool installation: %w", err)
	}

	cache.Tools[name] = result.tagName

	return nil
}

func installTools(config Configuration, tools []string, downloadTimeout int) error {
	err := makeOutputDirectory(config.InstallationDirectory)
	if err != nil {
		return err
	}

	cache, err := getCache()
	if err != nil {
		return err
	}

	downloader := newDownloader(downloadTimeout)

	if len(tools) > 0 {
		for _, name := range tools {
			fmt.Printf("Installing tool '%s'.\n", name)
			tool, found := config.Tools[name]
			if !found {
				fmt.Printf("Error: tool '%s' not found in configuration\n", name)
				continue
			}

			currentVersion := cache.Tools[name]

			result, err := downloader.downloadTool(tool, currentVersion)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			err = installFiles(name, tool.Binaries, result, config.InstallationDirectory, &cache)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	} else {
		for name, tool := range config.Tools {
			fmt.Printf("Installing tool '%s'.\n", name)

			currentVersion := cache.Tools[name]

			result, err := downloader.downloadTool(tool, currentVersion)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			if result.updated {
				fmt.Printf("Info: skipping download for '%v' because it is already installed and up to date", name)
				continue
			}

			err = installFiles(name, tool.Binaries, result, config.InstallationDirectory, &cache)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	}

	err = cache.writeCache()
	if err != nil {
		return err
	}

	return nil
}

func updateTools(config Configuration, downloadTimeout int) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	outdated, err := getOutdatedTools(config, false, downloadTimeout, cache)
	if err != nil {
		return err
	}

	err = makeOutputDirectory(config.InstallationDirectory)
	if err != nil {
		return err
	}

	downloader := newDownloader(downloadTimeout)
	for _, t := range outdated {
		name := t.Name

		fmt.Printf("Installing tool '%s'.\n", name)
		tool, found := config.Tools[name]
		if !found {
			fmt.Printf("Error: tool '%s' not found in configuration\n", name)
			continue
		}

		currentVersion := cache.Tools[name]

		result, err := downloader.downloadTool(tool, currentVersion)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		err = installFiles(name, tool.Binaries, result, config.InstallationDirectory, &cache)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}

	err = cache.writeCache()
	if err != nil {
		return err
	}

	return nil
}
