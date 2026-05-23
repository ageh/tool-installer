// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
)

type ToolInfo struct {
	Name        string
	Link        string
	Description string
	Version     string
}

type ToolVersionInfo struct {
	Name      string
	Installed string
	Available string
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

func (app *App) addTool(githubSlug string, entryName string) UserMessage {
	parts := strings.Split(githubSlug, "/")
	if len(parts) != 2 {
		return UserMessage{Type: Error, Tool: "user", Content: fmt.Sprintf("%q is an invalid Github slug, expected the form 'owner/repository'", githubSlug)}
	}

	owner := parts[0]
	repository := parts[1]

	name := entryName
	if name == "" {
		name = repository
	}

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

	var repoInfo RepositoryInfo
	var repoError error

	var release Release
	var releaseError error

	var wg sync.WaitGroup

	wg.Go(func() { repoInfo, repoError = app.downloader.downloadRepoInfo(owner, repository) })
	wg.Go(func() { release, releaseError = app.downloader.downloadRelease(owner, repository) })

	wg.Wait()

	if repoError != nil {
		return UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to fetch repository info: %v", repoError)}
	}

	if releaseError != nil {
		return UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to fetch latest release: %v", releaseError)}
	}

	var asset Asset
	var assetRegex AssetRegex

	assetCandidates := matchBestAssetName(release.Assets, runtime.GOOS, runtime.GOARCH)

	if len(assetCandidates) == 1 {
		asset = assetCandidates[0].asset
		assetRegex = assetCandidates[0].assetRegex

		fmt.Printf("Automatically determined asset regex from asset name: %q\n", asset.Name)
	} else {
		fmt.Println("Could not determine asset name automatically.")
		pattern := promptForUniqueAssetRegex(release.Assets)

		tmp, err := stringToAssetRegex(pattern)
		if err != nil {
			return UserMessage{Type: Error, Tool: "tooli", Content: fmt.Sprintf("unexpected error compiling asset regex: %v", err)}
		}

		assetRegex = tmp
	}

	assetContent, err := app.downloader.downloadAsset(asset.Url)
	if err != nil {
		return UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("error when trying to download asset: %v", err)}
	}

	fileNames, err := getBinaryFileNames(assetContent, asset.Name)
	if err != nil {
		return UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("error when trying to obtain binary file names in asset: %v", err)}
	}

	binaries := promptForBinaries(fileNames)

	app.config.Tools[name] = Tool{
		Binaries:    binaries,
		Owner:       owner,
		Repository:  repository,
		Asset:       assetRegex,
		Description: repoInfo.Description,
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
			} else if result.upToDate {
				messageChannel <- UserMessage{Type: Info, Tool: name, Content: "skipping download - already up to date"}
			} else {
				assetType, err := extractFiles(result.data, result.assetName, tool.Binaries, toolDirectory)
				if err != nil {
					messageChannel <- UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to extract files: %v", err)}
					return
				}

				var message string
				if assetType == RawBinary {
					message = fmt.Sprintf("successfully installed version '%s' from the downloaded raw binary", result.tagName)
				} else {
					message = fmt.Sprintf("successfully installed version '%s' from the downloaded archive", result.tagName)
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
	tmp := make([]ToolInfo, len(app.config.Tools))

	i := 0
	for k, v := range app.config.Tools {
		tmp[i] = ToolInfo{Name: k, Link: fmt.Sprintf("%s/%s", v.Owner, v.Repository), Description: v.Description, Version: ""}

		if version, found := app.cache.Tools[k]; found {
			tmp[i].Version = version
		}

		i++
	}

	slices.SortFunc(tmp, func(a ToolInfo, b ToolInfo) int {
		return strings.Compare(a.Name, b.Name)
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

		allBinariesExist, err := app.allBinariesExist(toolDirectory, tool)
		if err != nil {
			return results, err
		}

		if !allBinariesExist {
			results = append(results, UserMessage{Type: Info, Tool: name, Content: "tool exists in the cache but one or more binaries are missing, removing stale cache entry"})
			app.cache.remove(name)
			continue
		}

		for _, binary := range tool.Binaries {
			n := binary.getTargetName()

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

	installedTools, cacheOnly, staleCache, err := app.toolsFromCache()
	if err != nil {
		return err
	}

	installed := len(installedTools)
	cacheOnlyCount := len(cacheOnly)
	staleCacheCount := len(staleCache)

	versionStatus := "skipped (dev build)"
	if version != "dev" {
		release, err := app.downloader.downloadRelease("ageh", "tool-installer")
		if err != nil {
			versionStatus = fmt.Sprintf("check failed (%v)", err)
		} else if release.Name == version {
			versionStatus = "up to date"
		} else {
			versionStatus = fmt.Sprintf("new version %s available", release.Name)
		}
	}

	statusTable := newTableBuilder([]string{"Field", "Value"})
	statusTable.addRow([]string{"Configuration file path", app.configLocation})
	statusTable.addRow([]string{"Cache file path", cachePath})
	statusTable.addRow([]string{"Tool installation directory", installPath})
	statusTable.addRow([]string{"Configured tools", fmt.Sprintf("%d", configured)})
	statusTable.addRow([]string{"Installed tools", fmt.Sprintf("%d", installed)})
	statusTable.addRow([]string{"Cache-only tools", fmt.Sprintf("%d", cacheOnlyCount)})
	statusTable.addRow([]string{"Stale cache entries", fmt.Sprintf("%d", staleCacheCount)})
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
	if staleCacheCount != 0 {
		if verbose {
			fmt.Println("Stale cache entries:")
			slices.Sort(staleCache)
			for _, name := range staleCache {
				fmt.Printf("- %s\n", name)
			}
		} else {
			colorPrintln(HintBlue, "hint: run `status verbose` to see which tools are present in the cache but missing binaries on disk")
		}
	}

	return nil
}

func (app *App) allBinariesExist(toolDirectory string, tool Tool) (bool, error) {
	for _, binary := range tool.Binaries {
		path := filepath.Join(toolDirectory, binary.getTargetName())
		_, err := os.Stat(path)
		if err == nil {
			continue
		}

		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("failed to stat binary '%s': %w", path, err)
	}

	return true, nil
}

func (app *App) toolsFromCache() (map[string]Tool, []string, []string, error) {
	toolDirectory, err := app.config.getSanitizedInstallationDirectory()
	if err != nil {
		return nil, nil, nil, err
	}

	tools := make(map[string]Tool, len(app.cache.Tools))
	notFound := make([]string, 0)
	stale := make([]string, 0)
	for name := range app.cache.Tools {
		tool, found := app.config.Tools[name]
		if !found {
			notFound = append(notFound, name)
		} else {
			allBinariesExist, err := app.allBinariesExist(toolDirectory, tool)
			if err != nil {
				return nil, nil, nil, err
			}

			if !allBinariesExist {
				stale = append(stale, name)
				continue
			}

			tools[name] = tool
		}
	}

	return tools, notFound, stale, nil
}

func (app *App) getOutdatedTools(checkAll bool) ([]UserMessage, []ToolVersionInfo, error) {
	messages := make([]UserMessage, 0)

	var tools map[string]Tool
	if checkAll {
		tools = app.config.Tools
	} else {
		tmp, notFound, stale, err := app.toolsFromCache()
		if err != nil {
			return messages, nil, err
		}
		tools = tmp

		for _, name := range notFound {
			messages = append(messages, UserMessage{Type: Error, Tool: name, Content: "tool exists in cache but is not in configuration"})
		}
		for _, name := range stale {
			messages = append(messages, UserMessage{Type: Error, Tool: name, Content: "tool exists in cache but one or more binaries are missing on disk"})
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var result []ToolVersionInfo

	for name, tool := range tools {
		wg.Go(func() {
			release, err := app.downloader.downloadRelease(tool.Owner, tool.Repository)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				messages = append(messages, UserMessage{Type: Error, Tool: name, Content: fmt.Sprintf("failed to download release info: %v", err)})
			} else if app.cache.Tools[name] != release.TagName {
				result = append(result, ToolVersionInfo{Name: name, Installed: app.cache.Tools[name], Available: release.TagName})
			}
		})
	}

	wg.Wait()

	slices.SortFunc(result, func(a ToolVersionInfo, b ToolVersionInfo) int {
		return strings.Compare(a.Name, b.Name)
	})

	return messages, result, nil
}
