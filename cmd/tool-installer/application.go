// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

func compareNames(a string, b string) int {
	if a < b {
		return -1
	}

	if a > b {
		return 1
	}

	return 0
}

type App struct {
	downloader     Downloader
	config         Configuration
	cache          Cache
	configLocation string
}

func newApp(configPath string, timeout int) (App, []UserMessage, error) {
	var result App
	messages := make([]UserMessage, 0)

	config, message, err := readConfigurationOrCreateDefault(configPath)
	if err != nil {
		return result, messages, fmt.Errorf("could not obtain configuration: %w", err)
	}

	if message != nil {
		messages = append(messages, *message)
	}

	result.config = config

	cache, err := getCache()
	if err != nil {
		return result, messages, fmt.Errorf("could not obtain cache: %w", err)
	}

	result.cache = cache

	downloader, message := newDownloader(timeout)
	if message != nil {
		messages = append(messages, *message)
	}

	result.downloader = downloader

	result.configLocation = configPath

	return result, messages, nil
}

func (app *App) addTool(name string) UserMessage {
	_, found := app.config.Tools[name]
	if found {
		return UserMessage{Type: Info, Tool: name, Content: "skipping addition to configuration - an entry already exists"}
	}

	tool, found := knownTools[name]
	if found {
		tmp, err := tool.intoToolForPlatform()
		if err != nil {
			return UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("error obtaining known tool: %v", err)}
		}

		app.config.Tools[name] = tmp
		err = app.config.save(app.configLocation, false)
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

	assetName := promptRegex("Asset name (regex): ")
	asset, err := stringToAssetRegex(assetName)
	if err != nil {
		return UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to compile asset name regex: %v", err)}
	}

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
		Binaries:    binaries,
		Owner:       owner,
		Repository:  repo,
		Asset:       asset,
		Description: description,
	}

	err = app.config.save(app.configLocation, false)
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

	slices.SortFunc(tmp, func(a ToolInfo, b ToolInfo) int {
		return compareNames(a.Name, b.Name)
	})

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

func (app *App) showStatus(verbose bool) error {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return fmt.Errorf("failed to obtain the cache file path: %w", err)
	}

	installPath, err := app.config.getSanitizedInstallationDirectory()
	if err != nil {
		return fmt.Errorf("failed to obtain the installation path: %w", err)
	}

	configured := len(app.config.Tools)
	installed := len(app.cache.Tools)

	_, cacheOnly := app.toolsFromCache()
	cacheOnlyCount := len(cacheOnly)

	versionStatus := "skipped (dev build)"
	if version != "dev" {
		release, err := app.downloader.downloadRelease("ageh", "tool-installer")
		if err != nil {
			versionStatus = fmt.Sprintf("check failed (%v)", err)
		} else if release.TagName == version {
			versionStatus = "up to date"
		} else {
			versionStatus = fmt.Sprintf("new version %s available", release.TagName)
		}
	}

	statusTable := newTableBuilder([]string{"Field", "Value"})
	statusTable.addRow([]string{"Configuration file path", app.configLocation})
	statusTable.addRow([]string{"Cache file path", cachePath})
	statusTable.addRow([]string{"Tool installation directory", installPath})
	statusTable.addRow([]string{"Configured tools", fmt.Sprintf("%d", configured)})
	statusTable.addRow([]string{"Installed tools", fmt.Sprintf("%d", installed)})
	statusTable.addRow([]string{"Cache-only tools", fmt.Sprintf("%d", cacheOnlyCount)})
	statusTable.addRow([]string{"Current version", version})
	statusTable.addRow([]string{"Update status", versionStatus})

	fmt.Print(statusTable.build())
	if installed > configured {
		colorPrintln(WarningYellow, "warning: installed tool count exceeds configured tool count")
	}
	if cacheOnlyCount != 0 {
		if verbose {
			fmt.Println("Cache-only tools:")
			slices.Sort(cacheOnly)
			for _, name := range cacheOnly {
				fmt.Printf("- %s\n", name)
			}
		} else {
			colorPrintln(HintBlue, "hint: run `status verbose` to see which tools are present in the cache but not in the configuration")
		}
	}

	return nil
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

	slices.SortFunc(result, func(a ToolVersionInfo, b ToolVersionInfo) int {
		return compareNames(a.Name, b.Name)
	})

	return messages, result, nil
}
