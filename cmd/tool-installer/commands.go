// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"sort"
)

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

func printConfigError(err error) {
	fmt.Printf("Error: Could not load configuration: %v.\n", err)
	fmt.Println("Check if the configuration file is valid.")
	fmt.Println("You can generate a new configuration file with 'tooli create-config'.")
}

func checkToolVersions(configLocation *string, checkAll bool, downloadTimeout int) {
	config, err := getConfig(*configLocation)
	if err != nil {
		printConfigError(err)
		os.Exit(1)
	}

	cache, err := getCache()
	if err != nil {
		fmt.Printf("Error: Failed to obtain cache. Message: %v", err)
		os.Exit(1)
	}

	downloader := newDownloader(downloadTimeout)

	var nTools int
	if checkAll {
		nTools = len(config.Tools)
	} else {
		nTools = len(cache.Tools)
	}

	tmp := make([]VersionTableEntry, nTools)

	nameSize := 4
	installedSize := 9
	availableSize := 9

	if checkAll {
		i := 0
		for k, v := range config.Tools {
			release, err := downloader.downloadRelease(v.Owner, v.Repository)
			if err != nil {
				fmt.Printf("Error obtaining latest release of tool '%v'. Message: %v\n", k, err)
				continue
			}

			tmp[i] = VersionTableEntry{Name: k, Installed: "", Available: release.TagName}

			if current, found := cache.Tools[k]; found {
				tmp[i].Installed = current
			}

			nameSize = max(nameSize, len(k))
			installedSize = max(installedSize, len(tmp[i].Installed))
			availableSize = max(availableSize, len(tmp[i].Available))

			i++
		}
	} else {
		i := 0
		for name, version := range cache.Tools {
			tool := config.Tools[name]
			release, err := downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				fmt.Printf("Error obtaining latest release of tool '%v'. Message: %v\n", name, err)
				continue
			}

			tmp[i] = VersionTableEntry{Name: name, Installed: version, Available: release.TagName}

			nameSize = max(nameSize, len(name))
			installedSize = max(installedSize, len(tmp[i].Installed))
			availableSize = max(availableSize, len(tmp[i].Available))

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

	if len(results) > 0 {
		fmt.Printf("%-*s    %-*s    %-*s\n\n", nameSize, "Name", installedSize, "Installed", availableSize, "Available")

		for _, j := range results {
			fmt.Printf("%-*s    %-*s    %-*s\n", nameSize, j.Name, installedSize, j.Installed, availableSize, j.Available)
		}
	} else {
		fmt.Println("All tools are up to date.")
	}
}

func listTools(configLocation *string, longList bool) {
	config, err := getConfig(*configLocation)
	if err != nil {
		printConfigError(err)
		os.Exit(1)
	}

	cache, err := getCache()
	if err != nil {
		fmt.Printf("Error: Failed to obtain cache. Message: %v", err)
		os.Exit(1)
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
}

func makeOutputDirectory(path *string) error {
	return os.MkdirAll(*path, 0755)
}

func installTools(configLocation *string, installOnly *string, downloadTimeout int) {
	config, err := getConfig(*configLocation)
	if err != nil {
		printConfigError(err)
		os.Exit(1)
	}

	err = makeOutputDirectory(&config.InstallationDirectory)
	if err != nil {
		fmt.Printf("Error: Could not create output directory %v.\n", config.InstallationDirectory)
		os.Exit(1)
	}

	cache, err := getCache()
	if err != nil {
		fmt.Printf("Error: Could not obtain cache directory.\n")
		os.Exit(1)
	}

	downloader := newDownloader(downloadTimeout)

	if *installOnly != "" {
		fmt.Printf("Installing tool '%s'.\n", *installOnly)
		err = downloader.downloadTool(*installOnly, &config, &cache)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	} else {
		for k := range config.Tools {
			fmt.Printf("Installing tool '%s'.\n", k)
			err = downloader.downloadTool(k, &config, &cache)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	}

	cache.writeCache()
}
