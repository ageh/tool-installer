// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func newTestDownloader(t *testing.T, handler http.Handler) Downloader {
	t.Helper()

	server := httptest.NewTestServer(t, handler)

	return Downloader{client: *server.Client()}
}

func TestNewDownloaderGitHubTokenPrecedence(t *testing.T) {
	tests := []struct {
		name            string
		configuredToken string
		tooliToken      string
		githubToken     string
		expected        string
		expectsMessage  bool
	}{
		{name: "Highest precedence: Configuration", configuredToken: "configured", tooliToken: "tooli", githubToken: "github", expected: "configured"},
		{name: "Second highest precedence: TOOLI_GITHUB_TOKEN", tooliToken: "tooli", githubToken: "github", expected: "tooli"},
		{name: "Last precedence: GITHUB_TOKEN", githubToken: "github", expected: "github"},
		{name: "No token", expectsMessage: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("TOOLI_GITHUB_TOKEN", test.tooliToken)
			t.Setenv("GITHUB_TOKEN", test.githubToken)

			downloader, message := newDownloader(30, test.configuredToken)
			if downloader.githubToken != test.expected {
				t.Errorf("got token %q, expected %q", downloader.githubToken, test.expected)
			}
			if (message != nil) != test.expectsMessage {
				t.Errorf("got message: %t, expected: %t", message != nil, test.expectsMessage)
			}
		})
	}
}

func testDownloadableTool(owner string, repository string) Tool {
	return Tool{
		Owner:      owner,
		Repository: repository,
		Asset:      AssetRegex{Pattern: `\.bin$`, Regex: regexp.MustCompile(`\.bin$`)},
		Binaries:   []Binary{{Name: repository}},
	}
}

func TestDownloadReleaseSurfacesRateLimitError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/tool/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	downloader := newTestDownloader(t, mux)

	_, err := downloader.downloadRelease("owner", "tool")
	if err == nil {
		t.Fatal("expected an error for a 403 response")
	}

	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected a rate-limit error, got: %v", err)
	}
}

func TestDownloadToolSkipsAssetDownloadWhenUpToDate(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/tool/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.0.0",
			Assets:  []Asset{{Name: "tool.bin", Url: gitHubApiUrl + "/assets/tool"}},
		})
	})
	mux.HandleFunc("/assets/tool", func(w http.ResponseWriter, r *http.Request) {
		t.Error("asset should not be downloaded when the cached version is already up to date")
	})

	downloader := newTestDownloader(t, mux)

	result, err := downloader.downloadTool(testDownloadableTool("owner", "tool"), "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.upToDate {
		t.Error("expected the result to report the tool as up to date")
	}
}

func TestDownloadToolRejectsChecksumMismatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/tool/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Release{
			TagName: "v1.0.0",
			Assets: []Asset{{
				Name:   "tool.bin",
				Url:    gitHubApiUrl + "/assets/tool",
				Digest: "sha256:" + strings.Repeat("0", 64),
			}},
		})
	})
	mux.HandleFunc("/assets/tool", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("actual binary content"))
	})

	downloader := newTestDownloader(t, mux)

	_, err := downloader.downloadTool(testDownloadableTool("owner", "tool"), "")
	if err == nil {
		t.Fatal("expected an error for a mismatched checksum")
	}
}

func TestDownloadAssetRejectsOversizedResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/assets/huge", func(w http.ResponseWriter, r *http.Request) {
		chunk := make([]byte, 1024*1024)
		var written int
		for written <= maxAssetSize {
			n, err := w.Write(chunk)
			if err != nil {
				return
			}
			written += n
		}
	})

	downloader := newTestDownloader(t, mux)

	_, err := downloader.downloadAsset(gitHubApiUrl + "/assets/huge")
	if err == nil {
		t.Fatal("expected an error for an oversized asset")
	}

	if !strings.Contains(err.Error(), "exceeds maximum allowed size") {
		t.Errorf("unexpected error: %v", err)
	}
}
