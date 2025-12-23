// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Downloader struct {
	client      http.Client
	githubToken string
}

type DownloadResult struct {
	data      []byte
	assetName string
	tagName   string
	updated   bool
}

type RequestFormat int

const (
	rtJson RequestFormat = iota
	rtBinary
)

const rateLimitText = `got non-OK status code '%v'.

This most likely means that you hit Github's API rate limit. To increase the number of requests you can make, set the 'GITHUB_TOKEN' environment variable`

func createUserAgent() string {
	return "ageh/tool-installer-" + version
}

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
		return nil, errors.New("invalid request type")
	}

	userAgent := createUserAgent()
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
		return result, fmt.Errorf(rateLimitText, resp.StatusCode)
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
		return result, fmt.Errorf(rateLimitText, resp.StatusCode)
	}

	result, err = io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (client *Downloader) downloadTool(tool Tool, currentVersion string) (DownloadResult, error) {
	var result DownloadResult
	release, err := client.downloadRelease(tool.Owner, tool.Repository)
	if err != nil {
		return result, err
	}

	if currentVersion == release.TagName {
		result.updated = true
		return result, nil
	}

	var assetName string
	switch os := runtime.GOOS; os {
	case "linux":
		assetName = tool.LinuxAsset
	case "windows":
		assetName = tool.WindowsAsset
	default:
		return result, fmt.Errorf("the platform '%s' is not supported", os)
	}

	if assetName == "" {
		return result, errors.New("no asset name provided for the current platform")
	}

	checksumRegex, err := regexp.Compile(`(?i)\.(sha(\d+)?(sum)?|md5(sum)?|checksums\.txt)$`)
	if err != nil {
		return result, fmt.Errorf("failed to compile checksum regex: %w", err)
	}
	assetRegex, err := regexp.Compile(assetName)
	if err != nil {
		return result, fmt.Errorf("failed to compile asset regex: %w", err)
	}

	var res []Asset
	for _, a := range release.Assets {
		if checksumRegex.MatchString(a.Name) {
			continue
		}

		if assetRegex.MatchString(a.Name) {
			res = append(res, a)
		}
	}

	if len(res) == 0 {
		return result, errors.New("could not find a matching asset. Did you forget to include one in the config?")
	}
	if len(res) > 1 {
		assets := make([]string, 0)
		for _, a := range res {
			assets = append(assets, a.Name)
		}
		return result, fmt.Errorf("found two or more matching assets (%v). Please be more specific", strings.Join(assets, ", "))
	}

	asset := res[0]

	assetUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/assets/%d", tool.Owner, tool.Repository, asset.Id)

	binaryContent, err := client.downloadAsset(assetUrl)
	if err != nil {
		return result, err
	}

	hash := fmt.Sprintf("sha256:%x", sha256.Sum256(binaryContent))
	if asset.Digest != "" && hash != asset.Digest {
		return result, errors.New("found non-matching sha256 hash. It is possible that the download got corrupted")
	}

	result.data = binaryContent
	result.assetName = asset.Name
	result.tagName = release.TagName

	return result, nil
}
