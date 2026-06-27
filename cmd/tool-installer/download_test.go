// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

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
