// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type ToolInfo struct {
	Name        string
	Link        string
	Description string
	Version     string
}

func (t ToolInfo) GetName() string {
	return t.Name
}

type MessageType int

const (
	Success MessageType = iota
	Info
	Error
)

type ToolVersionInfo struct {
	Name        string
	Installed   string
	Available   string
	MessageType MessageType
	Message     string
}

func (v ToolVersionInfo) GetName() string {
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

type App struct {
	downloader     Downloader
	config         Configuration
	cache          Cache
	configLocation string
}

func newApp(configPath string, timeout int) (App, error) {
	var result App

	config, err := readConfiguration(configPath)
	if err != nil {
		return result, fmt.Errorf("could not read configuration: %w", err)
	}

	result.config = config

	cache, err := getCache()
	if err != nil {
		return result, fmt.Errorf("could not obtain cache: %w", err)
	}

	result.cache = cache

	result.downloader = newDownloader(timeout)

	result.configLocation = configPath

	return result, nil
}

func (app *App) addTool() error {
	name := promptNonEmpty("Please enter the name of the tool: ")

	_, found := app.config.Tools[name]
	if found {
		return errors.New("an entry for this tool already exists. If you want to modify it, please edit the configuration file")
	}

	fmt.Printf("Creating configuration entry for %s:\n", name)

	description := promptNonEmpty("Short description: ")
	owner := promptNonEmpty("GitHub user/org: ")
	repo := promptNonEmpty("Repository name: ")

	windows := promptRegex("Windows asset name (regex): ")
	linux := promptRegex("Linux asset name (regex): ")

	binary := promptNonEmpty("Binary name: ")
	rename := prompt("Rename binary to (leave empty if no rename): ")

	furtherEntries := prompt("Does this tool have more binaries? [y/N]: ")

	binaries := []Binary{{Name: binary, RenameTo: rename}}

	if furtherEntries == "y" || furtherEntries == "Y" {
		for {
			binary, ok := promptForBinary()

			if !ok {
				break
			}

			binaries = append(binaries, binary)
		}
	}

	app.config.Tools[name] = Tool{
		Binaries:     binaries,
		Owner:        owner,
		Repository:   repo,
		LinuxAsset:   linux,
		WindowsAsset: windows,
		Description:  description,
	}

	return app.config.save(app.configLocation, false)
}

func (app *App) checkToolVersions(checkAll bool) error {
	results, err := app.getOutdatedTools(checkAll)
	if err != nil {
		return fmt.Errorf("error during check for outdated versions: %w", err)
	}

	table := newTableBuilder([]string{"Name", "Installed", "Available"})

	if len(results) == 0 {
		fmt.Println("All tools are up to date.")
		return nil
	}

	for _, e := range results {
		table.addRow([]string{e.Name, e.Installed, e.Available})
	}

	fmt.Print(table.build())

	return nil
}

func (app *App) installTools(tools []string) ([]ToolVersionInfo, error) {
	toolDirectory := replaceTildePath(app.config.InstallationDirectory)
	err := makeOutputDirectory(toolDirectory)
	if err != nil {
		return nil, err
	}

	var toInstall map[string]Tool

	if len(tools) > 0 {
		toInstall = make(map[string]Tool, len(tools))
		for _, name := range tools {
			tool, found := app.config.Tools[name]
			if !found {
				fmt.Printf("Error: tool '%s' not found in configuration\n", name)
				continue
			}

			toInstall[name] = tool
		}
	} else {
		toInstall = app.config.Tools
	}

	var wg sync.WaitGroup

	results := make(chan ToolVersionInfo, len(toInstall))

	for name, tool := range toInstall {
		wg.Add(1)
		go func() {
			defer wg.Done()

			currentVersion := app.cache.Tools[name]

			result, err := app.downloader.downloadTool(tool, currentVersion)
			if err != nil {
				results <- ToolVersionInfo{Name: name, MessageType: Error, Message: fmt.Sprintf("%s: failed to download: %v\n", name, err)}
			} else if result.updated {
				results <- ToolVersionInfo{Name: name, MessageType: Info, Message: fmt.Sprintf("%s: skipping download - already up to date", name)}
			} else {
				assetType, err := extractFiles(result.data, result.assetName, tool.Binaries, toolDirectory)
				if err != nil {
					results <- ToolVersionInfo{Name: name, MessageType: Error, Message: fmt.Sprintf("%s: failed to extract files: %v", name, err)}
					return
				}

				if assetType == Archive {
					results <- ToolVersionInfo{Name: name, Installed: result.tagName, MessageType: Success, Message: fmt.Sprintf("%s: successfully installed from the downloaded archive", name)}
				} else {
					results <- ToolVersionInfo{Name: name, Installed: result.tagName, MessageType: Success, Message: fmt.Sprintf("%s: successfully installed from the downloaded raw binary", name)}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	result := make([]ToolVersionInfo, 0, len(toInstall))

	for res := range results {
		if res.MessageType == Success {
			app.cache.add(res.Name, res.Installed)
		}
		result = append(result, res)
	}

	err = app.cache.writeCache()
	if err != nil {
		return result, err
	}

	return result, nil
}

func (app *App) listTools(longList bool) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	tmp := make([]ToolInfo, len(app.config.Tools))

	i := 0
	for k, v := range app.config.Tools {
		tmp[i] = ToolInfo{Name: k, Link: fmt.Sprintf("%s/%s", v.Owner, v.Repository), Description: v.Description, Version: ""}

		if version, found := cache.Tools[k]; found {
			tmp[i].Version = version
		}

		i++
	}

	sort.Sort(ByName[ToolInfo]{tmp})

	var builder TableBuilder

	if longList {
		builder = newTableBuilder([]string{"Name", "Source", "Version", "Description"})

		for _, row := range tmp {
			builder.addRow([]string{row.Name, row.Link, row.Version, row.Description})
		}
	} else {
		builder = newTableBuilderWithLimits([]string{"Name", "Version", "Description"}, map[int]int{2: 50})

		for _, row := range tmp {
			builder.addRow([]string{row.Name, row.Version, row.Description})
		}
	}

	fmt.Print(builder.build())

	return nil
}

func (app *App) removeTools(tools []string, removeFromConfig bool) ([]string, error) {
	toolDirectory := replaceTildePath(app.config.InstallationDirectory)

	results := make([]string, 0)

	for _, name := range tools {

		tool, found := app.config.Tools[name]
		if !found {
			continue
		}

		for _, binary := range tool.Binaries {
			n := binary.Name
			if binary.RenameTo != "" {
				n = binary.RenameTo
			}

			path := filepath.Join(toolDirectory, n)
			err := os.Remove(path)
			if err != nil {
				results = append(results, fmt.Sprintf("%s: Failed to remove binary '%s'\n", name, n))
			} else {
				results = append(results, fmt.Sprintf("%s: Removed binary '%s'\n", name, n))
			}
		}

		app.cache.remove(name)
	}

	if removeFromConfig {
		for _, name := range tools {
			delete(app.config.Tools, name)
		}

		err := app.config.save(app.configLocation, false)
		if err != nil {
			return results, err
		}
	}

	return results, app.cache.writeCache()
}

func (app *App) updateTools() ([]ToolVersionInfo, error) {
	outdated, err := app.getOutdatedTools(false)
	if err != nil {
		return nil, err
	}

	results := make([]ToolVersionInfo, 0)

	tools := make([]string, len(outdated))
	for i, tmp := range outdated {
		if tmp.MessageType == Success {
			tools[i] = tmp.Name
		} else {
			results = append(results, tmp)
		}
	}

	installResults, err := app.installTools(tools)
	results = append(results, installResults...)

	return results, err
}

func (app *App) toolsFromCache() map[string]Tool {
	tools := make(map[string]Tool, len(app.cache.Tools))
	for name := range app.cache.Tools {
		tools[name] = app.config.Tools[name]
	}

	return tools
}

func (app *App) getOutdatedTools(checkAll bool) ([]ToolVersionInfo, error) {
	var tools map[string]Tool
	if checkAll {
		tools = app.config.Tools
	} else {
		tools = app.toolsFromCache()
	}

	var wg sync.WaitGroup

	results := make(chan ToolVersionInfo, len(tools))

	for name, tool := range tools {
		wg.Add(1)

		go func() {
			defer wg.Done()

			release, err := app.downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				shortMessage := fmt.Sprintf("Error: %v\n", err)
				fullMessage := fmt.Sprintf("%s: failed to download release info: %v\n", name, err)
				results <- ToolVersionInfo{Name: name, Installed: app.cache.Tools[name], Available: shortMessage, MessageType: Error, Message: fullMessage}
			} else {
				results <- ToolVersionInfo{Name: name, Installed: app.cache.Tools[name], Available: release.TagName, MessageType: Success}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	result := make([]ToolVersionInfo, 0)

	for r := range results {
		if r.Installed != r.Available {
			result = append(result, r)
		}
	}

	sort.Sort(ByName[ToolVersionInfo]{result})

	return result, nil
}
