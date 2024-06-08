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

func printConfigError(err error) {
	fmt.Printf("Error: Could not load configuration: %v.\n", err)
	fmt.Println("Check if the configuration file is valid.")
	fmt.Println("You can generate a new configuration file with 'tooli create-config'.")
}

func listTools(configLocation *string, shortList bool) {
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

	sort.Sort(ByName(tmp))

	if shortList {
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
	} else {
		fmt.Printf("%-*s    %-*s    %-*s    %-*s\n\n", nameSize, "Name", linkSize, "Owner/Repository", descriptionSize, "Description", versionSize, "Version")

		for _, j := range tmp {
			fmt.Printf("%-*s    %-*s    %-*s    %-*s\n", nameSize, j.Name, linkSize, j.Link, descriptionSize, j.Description, versionSize, j.Version)
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
