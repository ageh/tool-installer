// SPDX-License-Identifier: Apache-2.0

package main

import (
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

type ToolVersionInfo struct {
	Name      string
	Installed string
	Available string
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
	downloader           Downloader
	config               Configuration
	cache                Cache
	configLocation       string
	createdDefaultConfig bool
}

func newApp(configPath string, timeout int) (App, error) {
	var result App

	config, defaulted, err := readConfigurationOrCreateDefault(configPath)
	if err != nil {
		return result, fmt.Errorf("could not obtain configuration: %w", err)
	}

	result.createdDefaultConfig = defaulted
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

func (app *App) addTool(name string) UserMessage {
	_, found := app.config.Tools[name]
	if found {
		return UserMessage{Type: Info, Tool: name, Content: "skipping addition to configuration - an entry already exists"}
	}

	tool, found := knownTools[name]
	if found {
		app.config.Tools[name] = tool
		err := app.config.save(app.configLocation, false)
		if err != nil {
			return UserMessage{Type: Error, Tool: name, Content: "failed to write configuration to disk"}
		} else {
			return UserMessage{Type: Success, Tool: name, Content: "successfully added to the configuration with values taken from well-known tools list"}
		}
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

	err := app.config.save(app.configLocation, false)
	if err != nil {
		return UserMessage{Type: Error, Tool: name, Content: "failed to write configuration to disk"}
	} else {
		return UserMessage{Type: Success, Tool: name, Content: "successfully added to the configuration"}
	}
}

func (app *App) checkToolVersions(checkAll bool) ([]UserMessage, error) {
	messages, results, err := app.getOutdatedTools(checkAll)
	if err != nil {
		return messages, fmt.Errorf("error during check for outdated versions: %w", err)
	}

	table := newTableBuilder([]string{"Name", "Installed", "Available"})

	if len(results) == 0 {
		fmt.Println("All tools are up to date.")
		return messages, nil
	}

	for _, e := range results {
		table.addRow([]string{e.Name, e.Installed, e.Available})
	}

	fmt.Print(table.build())

	return messages, nil
}

func (app *App) installTools(tools []string) ([]UserMessage, error) {
	toolDirectory, err := app.config.getSanitizedInstallationDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain installation path: %w", err)
	}

	err = makeOutputDirectory(toolDirectory)
	if err != nil {
		return nil, err
	}

	var toInstall map[string]Tool

	messages := make([]UserMessage, 0)

	if len(tools) > 0 {
		toInstall = make(map[string]Tool, len(tools))
		for _, name := range tools {
			tool, found := app.config.Tools[name]
			if !found {
				messages = append(messages, UserMessage{Type: Error, Tool: name, Content: "tool not found in the configuration"})
				continue
			}

			toInstall[name] = tool
		}
	} else {
		toInstall = app.config.Tools
	}

	var wg sync.WaitGroup

	messageChannel := make(chan UserMessage, len(toInstall))
	versionInfoChannel := make(chan ToolVersionInfo, len(toInstall))

	for name, tool := range toInstall {
		wg.Go(func() {
			currentVersion := app.cache.Tools[name]

			result, err := app.downloader.downloadTool(tool, currentVersion)
			if err != nil {
				messageChannel <- UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to download tool: %v\n", err)}
			} else if result.updated {
				messageChannel <- UserMessage{Type: Info, Tool: name, Content: "skipping download - already up to date"}
			} else {
				assetType, err := extractFiles(result.data, result.assetName, tool.Binaries, toolDirectory)
				if err != nil {
					messageChannel <- UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to extract files: %v", err)}
					return
				}

				var message string
				if assetType == Archive {
					message = fmt.Sprintf("successfully installed version '%s' from the downloaded archive", result.tagName)
				} else {
					message = fmt.Sprintf("successfully installed version '%s' from the downloaded raw binary", result.tagName)
				}

				messageChannel <- UserMessage{Type: Success, Tool: name, Content: message}
				versionInfoChannel <- ToolVersionInfo{Name: name, Installed: result.tagName}
			}
		})
	}

	go func() {
		wg.Wait()
		close(messageChannel)
		close(versionInfoChannel)
	}()

	for m := range messageChannel {
		messages = append(messages, m)
	}

	for info := range versionInfoChannel {
		app.cache.add(info.Name, info.Installed)
	}

	err = app.cache.writeCache()
	if err != nil {
		return messages, err
	}

	return messages, nil
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

func (app *App) removeTools(tools []string, removeFromConfig bool) ([]UserMessage, error) {
	toolDirectory, err := app.config.getSanitizedInstallationDirectory()
	if err != nil {
		return nil, err
	}

	results := make([]UserMessage, 0)

	for _, name := range tools {
		tool, found := app.config.Tools[name]
		if !found {
			results = append(results, UserMessage{Type: Error, Tool: name, Content: "tool not found in the configuration"})
			continue
		}

		isInstalled := app.cache.contains(name)
		if !isInstalled {
			results = append(results, UserMessage{Type: Info, Tool: name, Content: "skipping uninstall - tool exists in the configuration but is not installed"})
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
				results = append(results, UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to remove binary '%s'", n)})
			} else {
				results = append(results, UserMessage{Type: Success, Tool: name, Content: fmt.Sprintf("successfully removed binary '%s'", n)})
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

func (app *App) updateTools() ([]UserMessage, error) {
	messages, outdated, err := app.getOutdatedTools(false)
	if err != nil {
		return messages, err
	}

	tools := make([]string, len(outdated))
	for i, tmp := range outdated {
		tools[i] = tmp.Name
	}

	installMessages, err := app.installTools(tools)
	messages = append(messages, installMessages...)

	return messages, err
}

func (app *App) toolsFromCache() (map[string]Tool, []string) {
	tools := make(map[string]Tool, len(app.cache.Tools))
	notFound := make([]string, 0)
	for name := range app.cache.Tools {
		tool, found := app.config.Tools[name]
		if !found {
			notFound = append(notFound, name)
		} else {
			tools[name] = tool
		}
	}

	return tools, notFound
}

func (app *App) getOutdatedTools(checkAll bool) ([]UserMessage, []ToolVersionInfo, error) {
	messages := make([]UserMessage, 0)

	var tools map[string]Tool
	if checkAll {
		tools = app.config.Tools
	} else {
		tmp, notFound := app.toolsFromCache()
		tools = tmp

		for _, name := range notFound {
			messages = append(messages, UserMessage{Type: Error, Tool: name, Content: "tool exists in cache but is not in configuration"})
		}
	}

	var wg sync.WaitGroup

	results := make(chan ToolVersionInfo, len(tools))
	messageChannel := make(chan UserMessage, len(tools))

	for name, tool := range tools {
		wg.Go(func() {
			release, err := app.downloader.downloadRelease(tool.Owner, tool.Repository)
			if err != nil {
				messageChannel <- UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to download release info: %v", err)}
			} else {
				results <- ToolVersionInfo{Name: name, Installed: app.cache.Tools[name], Available: release.TagName}
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
		close(messageChannel)
	}()

	result := make([]ToolVersionInfo, 0)

	for r := range results {
		if r.Installed != r.Available {
			result = append(result, r)
		}
	}

	for m := range messageChannel {
		messages = append(messages, m)
	}

	sort.Sort(ByName[ToolVersionInfo]{result})

	return messages, result, nil
}
