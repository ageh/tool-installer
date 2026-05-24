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
	"strings"
	"time"
)

var checksumRegex = regexp.MustCompile(`(?i)\.(sha(\d+)?(sum)?|md5(sum)?|checksums\.txt)$`)

type Downloader struct {
	client      http.Client
	githubToken string
}

type DownloadResult struct {
	data      []byte
	assetName string
	tagName   string
	upToDate  bool
}

type RequestFormat int

const (
	rtJson RequestFormat = iota
	rtBinary
)

func createUserAgent() string {
	return "ageh/tool-installer-" + version
}

func httpError(statusCode int) error {
	if statusCode == http.StatusForbidden || statusCode == http.StatusTooManyRequests {
		return fmt.Errorf("HTTP status %d: GitHub API rate limit is likely hit, check if you have set the `GITHUB_TOKEN` environment variable", statusCode)
	}

	return fmt.Errorf("unexpected HTTP status: %d", statusCode)
}

func newDownloader(timeoutSeconds int) (Downloader, *UserMessage) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	res := Downloader{client: http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}, githubToken: githubToken}

	if githubToken == "" {
		return res, &UserMessage{Type: Info, Tool: "tooli", Content: "GITHUB_TOKEN is not set in the environment variables, consider setting it to avoid rate limiting"}
	}

	return res, nil
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
		req.Header.Add("Authorization", "Bearer "+client.githubToken)
	}

	return req, nil
}

func (client *Downloader) downloadRepoInfo(owner string, repository string) (RepositoryInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repository)

	var result RepositoryInfo

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
		return result, httpError(resp.StatusCode)
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
		return result, httpError(resp.StatusCode)
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
		return result, httpError(resp.StatusCode)
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
		result.upToDate = true
		return result, nil
	}

	var res []Asset
	for _, a := range release.Assets {
		if checksumRegex.MatchString(a.Name) {
			continue
		}

		if tool.Asset.Regex.MatchString(a.Name) {
			res = append(res, a)
		}
	}

	if len(res) == 0 {
		return result, errors.New("could not find a matching asset. Did you forget to include one in the config?")
	}

	if len(res) > 1 {
		assets := make([]string, len(res))
		for i, a := range res {
			assets[i] = a.Name
		}
		return result, fmt.Errorf("found two or more matching assets (%v). Please be more specific", strings.Join(assets, ", "))
	}

	asset := res[0]

	binaryContent, err := client.downloadAsset(asset.Url)
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
