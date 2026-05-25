// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
)

type KnownTool struct {
	Binaries    []Binary
	Owner       string
	Repository  string
	Description string
	AssetNames  map[string]string
}

func (k KnownTool) intoToolForPlatform() (Tool, error) {
	lookup := runtime.GOOS + "/" + runtime.GOARCH

	asset, found := k.AssetNames[lookup]
	if !found {
		return Tool{}, errors.New("no known asset name for current platform")
	}

	re, err := regexp.Compile(asset)
	if err != nil {
		return Tool{}, fmt.Errorf("failed to compile regex: %w", err)
	}

	var result = Tool{
		Binaries:    k.Binaries,
		Owner:       k.Owner,
		Repository:  k.Repository,
		Description: k.Description,
		Asset:       AssetRegex{Pattern: asset, Regex: re},
	}

	return result, nil
}

const standardAssetLinuxArm = "aarch64-unknown-linux-gnu\\.tar\\.gz$"
const standardAssetLinuxMuslArm = "aarch64-unknown-linux-musl\\.tar\\.gz$"
const standardAssetLinuxx64 = "x86_64-unknown-linux-gnu\\.tar\\.gz$"
const standardAssetLinuxMuslx64 = "x86_64-unknown-linux-musl\\.tar\\.gz$"
const standardAssetWindowsArm = "aarch64-pc-windows-msvc\\.zip$"
const standardAssetWindowsx64 = "x86_64-pc-windows-msvc\\.zip$"
const standardAssetWindowsGnux64 = "x86_64-pc-windows-gnu\\.zip$"
const standardAssetAppleArm = "aarch64-apple-darwin\\.tar\\.gz$"
const standardAssetApplex64 = "x86_64-apple-darwin\\.tar\\.gz$"

var knownTools = map[string]KnownTool{
	// Shells and common shell command/tool replacements/upgrades
	"bat": {
		Binaries:    []Binary{{Name: "bat"}},
		Owner:       "sharkdp",
		Repository:  "bat",
		Description: "A cat(1) clone with wings.",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"dua": {
		Binaries:    []Binary{{Name: "dua"}},
		Owner:       "Byron",
		Repository:  "dua-cli",
		Description: "View disk space usage and delete unwanted data, fast.",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
		},
	},
	"dust": {
		Binaries:    []Binary{{Name: "dust"}},
		Owner:       "bootandy",
		Repository:  "dust",
		Description: "A more intuitive version of du in rust",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
		},
	},
	"eza": {
		Binaries:    []Binary{{Name: "eza"}},
		Owner:       "eza-community",
		Repository:  "eza",
		Description: "A modern alternative to ls",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsGnux64,
		},
	},
	"fd": {
		Binaries:    []Binary{{Name: "fd"}},
		Owner:       "sharkdp",
		Repository:  "fd",
		Description: "A simple, fast and user-friendly alternative to 'find'",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
		},
	},
	"fzf": {
		Binaries:    []Binary{{Name: "fzf"}},
		Owner:       "junegunn",
		Repository:  "fzf",
		Description: "A command-line fuzzy finder",
		AssetNames: map[string]string{
			"linux/amd64":   "linux_amd64\\.tar\\.gz$",
			"linux/arm64":   "linux_arm64\\.tar\\.gz$",
			"windows/amd64": "windows_amd64\\.zip$",
			"windows/arm64": "windows_arm64\\.zip$",
		},
	},
	"numbat": {
		Binaries:    []Binary{{Name: "numbat"}},
		Owner:       "sharkdp",
		Repository:  "numbat",
		Description: "A statically typed programming language for scientific computations with first class support for physical dimensions and units",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"nushell": {
		Binaries: []Binary{
			{Name: "nu"},
			{Name: "nu_plugin_stress_internals"},
			{Name: "nu_plugin_query"},
			{Name: "nu_plugin_polars"},
			{Name: "nu_plugin_inc"},
			{Name: "nu_plugin_gstat"},
			{Name: "nu_plugin_formats"},
			{Name: "nu_plugin_example"},
			{Name: "nu_plugin_custom_values"},
			{Name: "less"},
		},
		Owner:       "nushell",
		Repository:  "nushell",
		Description: "Data oriented shell",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"ouch": {
		Binaries:    []Binary{{Name: "ouch"}},
		Owner:       "ouch-org",
		Repository:  "ouch",
		Description: "(De)compression for your terminal",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
		},
	},
	"ripgrep": {
		Binaries:    []Binary{{Name: "rg"}},
		Owner:       "burntsushi",
		Repository:  "ripgrep",
		Description: "Better grep",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"sd": {
		Binaries:    []Binary{{Name: "sd"}},
		Owner:       "chmln",
		Repository:  "sd",
		Description: "Better sed",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxMuslArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"starship": {
		Binaries:    []Binary{{Name: "starship"}},
		Owner:       "starship",
		Repository:  "starship",
		Description: "Cross-shell custom prompt",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxMuslArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},

	// Linters, formatters, and LSPs
	"biome": {
		Binaries:    []Binary{{Name: "biome"}},
		Owner:       "biomejs",
		Repository:  "biome",
		Description: "Web Dev formatter and linter",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x64-musl",
			"linux/arm64":   "linux-arm64$",
			"windows/amd64": "win32-x64\\.exe$",
			"windows/arm64": "win32-arm64\\.exe$",
			"darwin/amd64":  "darwin-x64$",
			"darwin/arm64":  "darwin-arm64$",
		},
	},
	"ruff": {
		Binaries:    []Binary{{Name: "ruff"}},
		Owner:       "astral-sh",
		Repository:  "ruff",
		Description: "An extremely fast Python linter and code formatter",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"ty": {
		Binaries:    []Binary{{Name: "ty"}},
		Owner:       "astral-sh",
		Repository:  "ty",
		Description: "An extremely fast Python type checker and language server, written in Rust.",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"uv": {
		Binaries:    []Binary{{Name: "uv"}},
		Owner:       "astral-sh",
		Repository:  "uv",
		Description: "An extremely fast Python package installer and resolver, written in Rust.",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"vale": {
		Binaries:    []Binary{{Name: "vale"}},
		Owner:       "errata-ai",
		Repository:  "vale",
		Description: "Text linter",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_64-bit\\.tar\\.gz$",
			"linux/arm64":   "Linux_arm64\\.tar\\.gz$",
			"windows/amd64": "Windows_64-bit\\.zip$",
			"darwin/amd64":  "macOS_64-bit\\.zip$",
			"darwin/arm64":  "macOS_arm64\\.zip$",
		},
	},

	// Compilers and interpreters
	"bun": {
		Binaries:    []Binary{{Name: "bun"}},
		Owner:       "oven-sh",
		Repository:  "bun",
		Description: "Javascript runtime written in Zig",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x64\\.zip$",
			"linux/arm64":   "linux-aarch64\\.zip$",
			"windows/amd64": "windows-x64\\.zip$",
			"darwin/amd64":  "darwin-x64\\.zip$",
			"darwin/arm64":  "darwin-aarch64\\.zip$",
		},
	},
	"deno": {
		Binaries:    []Binary{{Name: "deno"}},
		Owner:       "denoland",
		Repository:  "deno",
		Description: "A modern runtime for JavaScript and TypeScript",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"luau": {
		Binaries: []Binary{
			{Name: "luau"},
			{Name: "luau-analyze"},
			{Name: "luau-compile"},
		},
		Owner:       "luau-lang",
		Repository:  "luau",
		Description: "A fast, small, safe, gradually typed embeddable scripting language derived from Lua",
		AssetNames: map[string]string{
			"linux/amd64":   "-ubuntu\\.zip$",
			"windows/amd64": "-windows\\.zip$",
			"darwin/amd64":  "-macos\\.zip$",
		},
	},
	"lune": {
		Binaries:    []Binary{{Name: "lune"}},
		Owner:       "lune-org",
		Repository:  "lune",
		Description: "A standalone Luau runtime",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x86_64\\.zip$",
			"linux/arm64":   "linux-aarch64\\.zip$",
			"windows/amd64": "windows-x86_64\\.zip$",
			"windows/arm64": "windows-aarch64\\.zip$",
			"darwin/amd64":  "macos-x86_64\\.zip$",
			"darwin/arm64":  "macos-aarch64\\.zip$",
		},
	},
	"teal": {
		Binaries:    []Binary{{Name: "tl"}},
		Owner:       "teal-language",
		Repository:  "tl",
		Description: "The compiler for Teal, a typed dialect of Lua",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x86_64\\.tar\\.gz$",
			"windows/amd64": "windows-x86_64\\.zip$",
		},
	},

	// Build tools and command runners
	"goreleaser": {
		Binaries:    []Binary{{Name: "goreleaser"}},
		Owner:       "goreleaser",
		Repository:  "goreleaser",
		Description: "Release engineering, simplified",
		AssetNames: map[string]string{
			"linux/amd64":   "goreleaser_Linux_x86_64\\.tar\\.gz$",
			"linux/arm64":   "goreleaser_Linux_arm64\\.tar\\.gz$",
			"windows/amd64": "goreleaser_Windows_x86_64\\.zip$",
			"windows/arm64": "goreleaser_Windows_arm64\\.zip$",
			"darwin/amd64":  "goreleaser_Darwin_x86_64\\.tar\\.gz$",
			"darwin/arm64":  "goreleaser_Darwin_arm64\\.tar\\.gz$",
		},
	},
	"just": {
		Binaries:    []Binary{{Name: "just"}},
		Owner:       "casey",
		Repository:  "just",
		Description: "Just a command runner",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxMuslx64,
			"linux/arm64":   standardAssetLinuxMuslArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"ninja": {
		Binaries:    []Binary{{Name: "ninja"}},
		Owner:       "ninja-build",
		Repository:  "ninja",
		Description: "A small build system with a focus on speed",
		AssetNames: map[string]string{
			"linux/amd64":   "ninja-linux\\.zip$",
			"linux/arm64":   "ninja-linux-aarch64\\.zip$",
			"windows/amd64": "ninja-win\\.zip$",
			"windows/arm64": "ninja-winarm64\\.zip$",
			"darwin/amd64":  "ninja-mac\\.zip$",
		},
	},

	// Text processors and static site generators
	"hugo": {
		Binaries:    []Binary{{Name: "hugo"}},
		Owner:       "gohugoio",
		Repository:  "hugo",
		Description: "Single binary static site generator",
		AssetNames: map[string]string{
			"linux/amd64":   "hugo_extended_[\\d\\.]+_linux-amd64\\.tar\\.gz$",
			"windows/amd64": "hugo_extended_[\\d\\.]+_windows-amd64\\.zip$",
		},
	},
	"mdbook": {
		Binaries:    []Binary{{Name: "mdbook"}},
		Owner:       "rust-lang",
		Repository:  "mdBook",
		Description: "Create a book from markdown files",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxMuslArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"mdbook-admonish": {
		Binaries:    []Binary{{Name: "mdbook-admonish"}},
		Owner:       "tommilligan",
		Repository:  "mdbook-admonish",
		Description: "A preprocessor for mdbook to add Material Design admonishments",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxMuslArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
		},
	},
	"pandoc": {
		Binaries:    []Binary{{Name: "pandoc"}},
		Owner:       "jgm",
		Repository:  "pandoc",
		Description: "Universal markup converter",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-amd64\\.tar\\.gz$",
			"linux/arm64":   "linux-arm64\\.tar\\.gz$",
			"windows/amd64": "windows-x86_64\\.zip$",
			"darwin/amd64":  "x86_64-maxOS\\.zip$",
			"darwin/arm64":  "arm64-maxOS\\.zip$",
		},
	},
	"typst": {
		Binaries:    []Binary{{Name: "typst"}},
		Owner:       "typst",
		Repository:  "typst",
		Description: "A new markup-based typesetting system",
		AssetNames: map[string]string{
			"linux/amd64":   "x86_64-unknown-linux-gnu\\.tar\\.xz$",
			"linux/arm64":   "aarch64-unknown-linux-musl\\.tar\\.xz$",
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  "x86_64-apple-darwin\\.tar\\.xz$",
			"darwin/arm64":  "aarch64-apple-darwin\\.tar\\.xz$",
		},
	},
	"zola": {
		Binaries:    []Binary{{Name: "zola"}},
		Owner:       "getzola",
		Repository:  "zola",
		Description: "Fast single binary static site generator",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},

	// Git/VCS
	"delta": {
		Binaries:    []Binary{{Name: "delta"}},
		Owner:       "dandavison",
		Repository:  "delta",
		Description: "A syntax-highlighting pager for git, diff, grep, and blame output",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"github": {
		Binaries:    []Binary{{Name: "gh"}},
		Owner:       "cli",
		Repository:  "cli",
		Description: "GitHub's official command line tool",
		AssetNames: map[string]string{
			"linux/amd64":   "linux_amd64\\.tar\\.gz$",
			"linux/arm64":   "linux_arm64\\.tar\\.gz$",
			"windows/amd64": "windows_amd64\\.zip$",
			"windows/arm64": "windows_arm64\\.zip$",
			"darwin/amd64":  "macOS_amd64\\.zip$",
			"darwin/arm64":  "macOS_arm64\\.zip$",
		},
	},
	"gitui": {
		Binaries:    []Binary{{Name: "gitui"}},
		Owner:       "gitui-org",
		Repository:  "gitui",
		Description: "Blazing fast terminal-ui for git written in rust",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x86_64\\.tar\\.gz$",
			"linux/arm64":   "linux-aarch64\\.tar\\.gz$",
			"windows/amd64": "win\\.tar\\.gz$",
			"darwin/amd64":  "mac-x86\\.tar\\.gz$",
		},
	},
	"jujutsu": {
		Binaries:    []Binary{{Name: "jj"}},
		Owner:       "martinvonz",
		Repository:  "jj",
		Description: "A Git-compatible VCS that is both simple and powerful",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxMuslArm,
			"windows/amd64": standardAssetWindowsx64,
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"lazygit": {
		Binaries:    []Binary{{Name: "lg"}},
		Owner:       "jesseduffield",
		Repository:  "lazygit",
		Description: "Simple terminal UI for git commands",
		AssetNames: map[string]string{
			"linux/amd64":   "linux_x86_64\\.tar\\.gz$",
			"linux/arm64":   "linux_arm64\\.tar\\.gz$",
			"windows/amd64": "windows_x86_64\\.zip$",
			"windows/arm64": "windows_arm64\\.zip$",
			"darwin/amd64":  "darwin_x86_64\\.tar\\.gz$",
			"darwin/arm64":  "darwin_arm64\\.tar\\.gz$",
		},
	},

	// Text editors
	"micro": {
		Binaries:    []Binary{{Name: "micro"}},
		Owner:       "micro-editor",
		Repository:  "micro",
		Description: "A modern and intuitive terminal-based text editor",
		AssetNames: map[string]string{
			"linux/amd64":   "linux64\\.tar\\.gz$",
			"linux/arm64":   "linux-arm64\\.tar\\.gz$",
			"windows/amd64": "win64\\.zip$",
			"windows/arm64": "win-arm64\\.zip$",
			"darwin/amd64":  "osx\\.tar\\.gz$",
			"darwin/arm64":  "macos-arm64\\.tar\\.gz$",
		},
	},

	// Multimedia
	"freeze": {
		Binaries:    []Binary{{Name: "freeze"}},
		Owner:       "charmbracelet",
		Repository:  "freeze",
		Description: "Generate images of code and terminal output",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_x86_64\\.tar\\.gz$",
			"linux/arm64":   "Linux_arm64\\.tar\\.gz$",
			"windows/amd64": "Windows_x86_64\\.zip$",
			"darwin/amd64":  "Darwin_x86_64\\.tar\\.gz$",
			"darwin/arm64":  "Darwin_arm64\\.tar\\.gz$",
		},
	},
	"vhs": {
		Binaries:    []Binary{{Name: "vhs"}},
		Owner:       "charmbracelet",
		Repository:  "vhs",
		Description: "Your CLI home video recorder",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_x86_64\\.tar\\.gz$",
			"linux/arm64":   "Linux_arm64\\.tar\\.gz$",
			"windows/amd64": "Windows_x86_64\\.zip$",
			"darwin/amd64":  "Darwin_x86_64\\.tar\\.gz$",
			"darwin/arm64":  "Darwin_arm64\\.tar\\.gz$",
		},
	},
	"yt-dlp": {
		Binaries:    []Binary{{Name: "yt-dlp"}},
		Owner:       "yt-dlp",
		Repository:  "yt-dlp",
		Description: "A feature-rich command-line audio/video downloader",
		AssetNames: map[string]string{
			"linux/amd64":   "yt-dlp_linux$",
			"linux/arm64":   "yt-dlp_linux_aarch64$",
			"windows/amd64": "yt-dlp\\.exe$",
			"windows/arm64": "yt-dlp_arm64\\.exe$",
			"darwin/amd64":  "yt-dlp_macos$",
		},
	},

	// Cryptography
	"age": {
		Binaries: []Binary{
			{Name: "age"},
			{Name: "age-keygen"},
		},
		Owner:       "FiloSottile",
		Repository:  "age",
		Description: "A simple, modern and secure encryption tool (and Go library) with small explicit keys, no config options, and UNIX-style composability.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-amd64\\.tar\\.gz$",
			"linux/arm64":   "linux-arm64\\.tar\\.gz$",
			"windows/amd64": "windows-amd64\\.zip$",
			"darwin/amd64":  "darwin-amd64\\.tar\\.gz$",
			"darwin/arm64":  "darwin-arm64\\.tar\\.gz$",
		},
	},
	"minisign": {
		Binaries:    []Binary{{Name: "minisign"}},
		Owner:       "jedisct1",
		Repository:  "minisign",
		Description: "A dead simple tool to sign files and verify digital signatures.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux\\.tar\\.gz$",
			"windows/amd64": "win64\\.zip$",
			"darwin/amd64":  "macos\\.zip$",
		},
	},
	"mkcert": {
		Binaries:    []Binary{{Name: "mkcert"}},
		Owner:       "FiloSottile",
		Repository:  "mkcert",
		Description: "A simple zero-config tool to make locally trusted development certificates with any names you'd like.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-amd64",
			"linux/arm64":   "linux-arm64",
			"windows/amd64": "windows-amd64\\.exe$",
			"windows/arm64": "windows-arm64\\.exe$",
			"darwin/amd64":  "darwin-amd64$",
			"darwin/arm64":  "darwin-arm64$",
		},
	},

	// AI
	"codex": {
		Binaries: []Binary{
			{Name: "codex", SourceNames: []string{"codex-x86_64-pc-windows-msvc", "codex-x86_64-unknown-linux-gnu"}},
		},
		Owner:       "openai",
		Repository:  "codex",
		Description: "Lightweight coding agent that runs in your terminal",
		AssetNames: map[string]string{
			"linux/amd64":   "codex-x86_64-unknown-linux-gnu.tar\\.gz$",
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": "codex-x86_64-pc-windows-msvc\\.exe\\.zip$",
			"windows/arm64": standardAssetWindowsArm,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"crush": {
		Binaries:    []Binary{{Name: "crush"}},
		Owner:       "charmbracelet",
		Repository:  "crush",
		Description: "The glamourous AI coding agent for your favourite terminal",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_x86_64\\.tar\\.gz$",
			"linux/arm64":   "Linux_arm64\\.tar\\.gz$",
			"windows/amd64": "Windows_x86_64\\.zip$",
			"windows/arm64": "Windows_arm64\\.zip$",
			"darwin/amd64":  "Darwin_x86_64\\.zip$",
			"darwin/arm64":  "Darwin_arm64\\.zip$",
		},
	},
	"opencode": {
		Binaries:    []Binary{{Name: "opencode"}},
		Owner:       "sst",
		Repository:  "opencode",
		Description: "The open source coding agent.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x64\\.tar\\.gz$",
			"linux/arm64":   "linux-arm64\\.tar\\.gz$",
			"windows/amd64": "windows-x64\\.zip$",
			"darwin/amd64":  "darwin-x64\\.zip$",
			"darwin/arm64":  "darwin-arm64\\.zip$",
		},
	},

	// Miscellaneous
	"hledger": {
		Binaries: []Binary{
			{Name: "hledger"},
			{Name: "hledger-ui"},
			{Name: "hledger-web"},
		},
		Owner:       "simonmichael",
		Repository:  "hledger",
		Description: "Robust, fast, intuitive plain text accounting tool with CLI, TUI and web interfaces.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x64\\.zip$",
			"windows/amd64": "windows-x64\\.zip$",
			"darwin/amd64":  "mac-x64\\.zip$",
			"darwin/arm64":  "mac-arm64\\.zip$",
		},
	},
	"hyperfine": {
		Binaries:    []Binary{{Name: "hyperfine"}},
		Owner:       "sharkdp",
		Repository:  "hyperfine",
		Description: "A command-line benchmarking tool",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": standardAssetWindowsx64,
			"darwin/amd64":  standardAssetApplex64,
			"darwin/arm64":  standardAssetAppleArm,
		},
	},
	"miniserve": {
		Binaries:    []Binary{{Name: "miniserve"}},
		Owner:       "svenstaro",
		Repository:  "miniserve",
		Description: "For when you really just want to serve some files over HTTP right now!",
		AssetNames: map[string]string{
			"linux/amd64":   "x86_64-unknown-linux-gnu$",
			"linux/arm64":   "aarch64-unknown-linux-gnu$",
			"windows/amd64": "x86_64-pc-windows-msvc\\.exe$",
			"darwin/amd64":  "x86_64-apple-darwin$",
			"darwin/arm64":  "aarch64-apple-darwin$",
		},
	},
	"tealdeer": {
		Binaries:    []Binary{{Name: "tldr", SourceNames: []string{"tealdeer"}}},
		Owner:       "dbrgn",
		Repository:  "tealdeer",
		Description: "A very fast implementation of tldr in Rust.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x86_64-musl$",
			"linux/arm64":   "linux-aarch64-musl$",
			"windows/amd64": "windows-x86_64-msvc\\.exe$",
			"darwin/amd64":  "macos-x86_64$",
			"darwin/arm64":  "macos-aarch64$",
		},
	},
	"tokei": {
		Binaries:    []Binary{{Name: "tokei"}},
		Owner:       "XAMPPRocky",
		Repository:  "tokei",
		Description: "Count your code, quickly.",
		AssetNames: map[string]string{
			"linux/amd64":   standardAssetLinuxx64,
			"linux/arm64":   standardAssetLinuxArm,
			"windows/amd64": "x86_64-pc-windows-msvc\\.exe$",
			"darwin/amd64":  standardAssetApplex64,
		},
	},
	"tailwind": {
		Binaries:    []Binary{{Name: "tailwind"}},
		Owner:       "tailwindlabs",
		Repository:  "tailwindcss",
		Description: "Tailwind CSS standalone CLI tool",
		AssetNames: map[string]string{
			"linux/amd64":   "tailwindcss-linux-x64$",
			"linux/arm64":   "tailwindcss-linux-arm$",
			"windows/amd64": "tailwindcss-windows-x64\\.exe$",
			"darwin/amd64":  "macos-x64$",
			"darwin/arm64":  "macos-arm64$",
		},
	},
}
