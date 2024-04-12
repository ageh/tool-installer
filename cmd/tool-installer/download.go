// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type Downloader struct {
	client      http.Client
	githubToken string
}

type RequestFormat int

const (
	rtJson RequestFormat = iota
	rtBinary
)

const rateLimitText = `Error: Got non-OK status code '%v'.

This most likely means that you hit Github's API rate limit. To increase the number of requests you can make, set the 'GITHUB_TOKEN' environment variable.
`

func newDownloader(timeoutSeconds int) Downloader {
	githubToken := os.Getenv("GITHUB_TOKEN")

	res := Downloader{client: http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}, githubToken: githubToken}

	return res
}

func (client *Downloader) newRequest(url string, requestFormat RequestFormat) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	switch requestFormat {
	case rtJson:
		req.Header.Add("Accept", "application/vnd.github+json")
	case rtBinary:
		req.Header.Add("Accept", "application/octet-stream")
	default:
		return nil, errors.New("Invalid request type")
	}

	req.Header.Add("User-Agent", userAgent)
	if client.githubToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", client.githubToken))
	}

	return req, nil
}

func (client *Downloader) downloadRelease(owner string, repository string) (Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repository)

	var result Release

	req, err := client.newRequest(url, rtJson)
	if err != nil {
		return result, err
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, errors.New(fmt.Sprintf(rateLimitText, resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (client *Downloader) downloadAsset(url string) ([]byte, error) {
	var result []byte

	req, err := client.newRequest(url, rtBinary)
	if err != nil {
		return result, err
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, errors.New(fmt.Sprintf(rateLimitText, resp.StatusCode))
	}

	result, err = io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (client *Downloader) downloadTool(name string, config *Configuration) error {

	tool, found := config.Tools[name]
	if !found {
		return errors.New(fmt.Sprintf("Tool '%s' not found in configuration.", name))
	}

	release, err := client.downloadRelease(tool.Owner, tool.Repository)
	if err != nil {
		return err
	}

	var asset string
	switch os := runtime.GOOS; os {
	case "linux":
		asset = tool.LinuxAsset
	case "windows":
		asset = tool.WindowsAsset
	default:
		return errors.New(fmt.Sprintf("The platform '%s' is not supported", os))
	}

	if asset == "" {
		return errors.New("No asset name provided for the current platform.")
	}

	var res []Asset
	for _, a := range release.Assets {
		if strings.HasSuffix(a.Name, asset) {
			res = append(res, a)
		}
	}

	if len(res) == 0 {
		return errors.New("Could not find a matching asset. Did you forget to include one in the config?")
	}
	if len(res) > 1 {
		return errors.New("Found two or more matching assets. Please be more specific.")
	}

	assetUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/assets/%d", tool.Owner, tool.Repository, res[0].Id)

	binaryContent, err := client.downloadAsset(assetUrl)
	if err != nil {
		return err
	}

	return extractFiles(binaryContent, &res[0], &tool, &config.InstallationDirectory)
}
