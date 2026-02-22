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
		Repository:  k.Owner,
		Description: k.Description,
		Asset:       AssetRegex{Pattern: asset, Regex: re},
	}

	return result, nil
}

const defaultLinuxArm = "aarch64-unknown-linux-gnu\\.tar\\.gz$"
const defaultLinux64 = "x86_64-unknown-linux-gnu\\.tar\\.gz$"
const defaultWindowsArm = "aarch-pc-windows-msvc\\.zip$"
const defaultWindows64 = "x86_64-pc-windows-msvc\\.zip$"
const defaultAppleArm = "aarch64-apple-darwin\\.tar\\.gz$"
const defaultApple64 = "x86_64-apple-darwin\\.tar\\.gz$"

var knownTools = map[string]KnownTool{
	// Shells and common shell command/tool replacements/upgrades
	"bat": {
		Binaries:    []Binary{{Name: "bat", RenameTo: ""}},
		Owner:       "sharkdp",
		Repository:  "bat",
		Description: "A cat(1) clone with wings.",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"dua": {
		Binaries:    []Binary{{Name: "dua", RenameTo: ""}},
		Owner:       "Byron",
		Repository:  "dua-cli",
		Description: "View disk space usage and delete unwanted data, fast.",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"dust": {
		Binaries:    []Binary{{Name: "dust", RenameTo: ""}},
		Owner:       "bootandy",
		Repository:  "dust",
		Description: "A more intuitive version of du in rust",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"eza": {
		Binaries:    []Binary{{Name: "eza", RenameTo: ""}},
		Owner:       "eza-community",
		Repository:  "eza",
		Description: "A modern alternative to ls",
		AssetNames: map[string]string{
			"linux/amd64":   "eza_x86_64-unknown-linux-gnu\\.tar\\.gz$",
			"windows/amd64": "eza\\.exe$_x86_64-pc-windows-gnu\\.zip$",
		},
	},
	"fd": {
		Binaries:    []Binary{{Name: "fd", RenameTo: ""}},
		Owner:       "sharkdp",
		Repository:  "fd",
		Description: "A simple, fast and user-friendly alternative to 'find'",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"fzf": {
		Binaries:    []Binary{{Name: "fzf", RenameTo: ""}},
		Owner:       "junegunn",
		Repository:  "fzf",
		Description: "A command-line fuzzy finder",
		AssetNames: map[string]string{
			"linux/amd64":   "linux_amd64\\.tar\\.gz$",
			"windows/amd64": "windows_amd64\\.zip$",
		},
	},
	"numbat": {
		Binaries:    []Binary{{Name: "numbat", RenameTo: ""}},
		Owner:       "sharkdp",
		Repository:  "numbat",
		Description: "A statically typed programming language for scientific computations with first class support for physical dimensions and units",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"nushell": {
		Binaries: []Binary{
			{Name: "nu", RenameTo: ""},
			{Name: "nu_plugin_stress_internals", RenameTo: ""},
			{Name: "nu_plugin_query", RenameTo: ""},
			{Name: "nu_plugin_polars", RenameTo: ""},
			{Name: "nu_plugin_inc", RenameTo: ""},
			{Name: "nu_plugin_gstat", RenameTo: ""},
			{Name: "nu_plugin_formats", RenameTo: ""},
			{Name: "nu_plugin_example", RenameTo: ""},
			{Name: "nu_plugin_custom_values", RenameTo: ""},
			{Name: "less", RenameTo: ""},
		},
		Owner:       "nushell",
		Repository:  "nushell",
		Description: "Data oriented shell",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"ouch": {
		Binaries:    []Binary{{Name: "ouch", RenameTo: ""}},
		Owner:       "ouch-org",
		Repository:  "ouch",
		Description: "(De)compression for your terminal",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"ripgrep": {
		Binaries:    []Binary{{Name: "rg", RenameTo: ""}},
		Owner:       "burntsushi",
		Repository:  "ripgrep",
		Description: "Better grep",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"sd": {
		Binaries:    []Binary{{Name: "sd", RenameTo: ""}},
		Owner:       "chmln",
		Repository:  "sd",
		Description: "Better sed",
		AssetNames: map[string]string{
			"linux/amd64":   "x86_64-unknown-linux-gnu",
			"windows/amd64": defaultWindows64,
		},
	},
	"starship": {
		Binaries:    []Binary{{Name: "starship", RenameTo: ""}},
		Owner:       "starship",
		Repository:  "starship",
		Description: "Cross-shell custom prompt",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},

	// Linters, formatters, and LSPs
	"biome": {
		Binaries:    []Binary{{Name: "biome", RenameTo: "biome"}},
		Owner:       "biomejs",
		Repository:  "biome",
		Description: "Web Dev formatter and linter",
		AssetNames: map[string]string{
			"linux/amd64":   "biome-linux-x64-musl",
			"windows/amd64": "biome-win32-x64\\.exe$",
		},
	},
	"ruff": {
		Binaries:    []Binary{{Name: "ruff", RenameTo: ""}},
		Owner:       "astral-sh",
		Repository:  "ruff",
		Description: "An extremely fast Python linter and code formatter",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"ty": {
		Binaries:    []Binary{{Name: "ty", RenameTo: ""}},
		Owner:       "astral-sh",
		Repository:  "ty",
		Description: "An extremely fast Python type checker and language server, written in Rust.",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"uv": {
		Binaries:    []Binary{{Name: "uv", RenameTo: ""}},
		Owner:       "astral-sh",
		Repository:  "uv",
		Description: "An extremely fast Python package installer and resolver, written in Rust.",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"vale": {
		Binaries:    []Binary{{Name: "vale", RenameTo: ""}},
		Owner:       "errata-ai",
		Repository:  "vale",
		Description: "Text linter",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_64-bit\\.tar\\.gz$",
			"windows/amd64": "Windows_64-bit\\.zip$",
		},
	},

	// Compilers and interpreters
	"bun": {
		Binaries:    []Binary{{Name: "bun", RenameTo: ""}},
		Owner:       "oven-sh",
		Repository:  "bun",
		Description: "Javascript runtime written in Zig",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x64\\.zip$",
			"windows/amd64": "windows-x64\\.zip$",
		},
	},
	"deno": {
		Binaries:    []Binary{{Name: "deno", RenameTo: ""}},
		Owner:       "denoland",
		Repository:  "deno",
		Description: "A modern runtime for JavaScript and TypeScript",
		AssetNames: map[string]string{
			"linux/amd64":   "deno-x86_64-unknown-linux-gnu\\.zip$",
			"windows/amd64": "deno-x86_64-pc-windows-msvc\\.zip$",
		},
	},
	"luau": {
		Binaries: []Binary{
			{Name: "luau", RenameTo: ""},
			{Name: "luau-analyze", RenameTo: ""},
			{Name: "luau-compile", RenameTo: ""},
		},
		Owner:       "luau-lang",
		Repository:  "luau",
		Description: "A fast, small, safe, gradually typed embeddable scripting language derived from Lua",
		AssetNames: map[string]string{
			"linux/amd64":   "-ubuntu\\.zip$",
			"windows/amd64": "-windows\\.zip$",
		},
	},
	"lune": {
		Binaries:    []Binary{{Name: "lune", RenameTo: ""}},
		Owner:       "lune-org",
		Repository:  "lune",
		Description: "A standalone Luau runtime ",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x86_64\\.zip$",
			"windows/amd64": "windows-x86_64\\.zip$",
		},
	},
	"teal": {
		Binaries:    []Binary{{Name: "tl", RenameTo: ""}},
		Owner:       "teal-language",
		Repository:  "tl",
		Description: "The compiler for Teal, a typed dialect of Lua",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x86_64\\.tar\\.gz$",
			"windows/amd64": "windows-x86_64\\.zip$",
		},
	},

	// Build tools and command runners
	"just": {
		Binaries:    []Binary{{Name: "just", RenameTo: ""}},
		Owner:       "casey",
		Repository:  "just",
		Description: "Just a command runner",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"ninja": {
		Binaries:    []Binary{{Name: "ninja", RenameTo: ""}},
		Owner:       "ninja-build",
		Repository:  "ninja",
		Description: "A small build system with a focus on speed",
		AssetNames: map[string]string{
			"linux/amd64":   "ninja-linux\\.zip$",
			"windows/amd64": "ninja-win\\.zip$",
		},
	},

	// Text processors and static site generators
	"hugo": {
		Binaries:    []Binary{{Name: "hugo", RenameTo: ""}},
		Owner:       "gohugoio",
		Repository:  "hugo",
		Description: "Single binary static site generator",
		AssetNames: map[string]string{
			"linux/amd64":   "hugo_extended_[\\d\\.]+_linux-amd64\\.tar\\.gz$",
			"windows/amd64": "hugo_extended_[\\d\\.]+_windows-amd64\\.zip$",
		},
	},
	"mdbook": {
		Binaries:    []Binary{{Name: "mdbook", RenameTo: ""}},
		Owner:       "rust-lang",
		Repository:  "mdBook",
		Description: "Create a book from markdown files",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"mdbook-admonish": {
		Binaries:    []Binary{{Name: "mdbook-admonish", RenameTo: ""}},
		Owner:       "tommilligan",
		Repository:  "mdbook-admonish",
		Description: "A preprocessor for mdbook to add Material Design admonishments",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"pandoc": {
		Binaries:    []Binary{{Name: "pandoc", RenameTo: ""}},
		Owner:       "jgm",
		Repository:  "pandoc",
		Description: "Universal markup converter",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-amd64\\.tar\\.gz$",
			"windows/amd64": "windows-x86_64\\.zip$",
		},
	},
	"typst": {
		Binaries:    []Binary{{Name: "typst", RenameTo: ""}},
		Owner:       "typst",
		Repository:  "typst",
		Description: "A new markup-based typesetting system",
		AssetNames: map[string]string{
			"linux/amd64":   "x86_64-unknown-linux-gnu.tar.xz",
			"windows/amd64": defaultWindows64,
		},
	},
	"zola": {
		Binaries:    []Binary{{Name: "zola", RenameTo: ""}},
		Owner:       "getzola",
		Repository:  "zola",
		Description: "Fast single binary static site generator",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},

	// Git/VCS
	"delta": {
		Binaries:    []Binary{{Name: "delta", RenameTo: ""}},
		Owner:       "dandavison",
		Repository:  "delta",
		Description: "A syntax-highlighting pager for git, diff, grep, and blame output",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"github": {
		Binaries:    []Binary{{Name: "gh", RenameTo: ""}},
		Owner:       "cli",
		Repository:  "cli",
		Description: "GitHub's official command line tool",
		AssetNames: map[string]string{
			"linux/amd64":   "linux amd64$",
			"windows/amd64": "windows amd64$",
		},
	},
	"gitui": {
		Binaries:    []Binary{{Name: "gitui", RenameTo: ""}},
		Owner:       "extrawurst",
		Repository:  "gitui",
		Description: "Blazing fast terminal-ui for git written in rust",
		AssetNames: map[string]string{
			"linux/amd64":   "gitui-linux-gnu\\.tar\\.gz$",
			"windows/amd64": "gitui-win\\.tar\\.gz$",
		},
	},
	"jujutsu": {
		Binaries:    []Binary{{Name: "jj", RenameTo: ""}},
		Owner:       "martinvonz",
		Repository:  "jj",
		Description: "A Git-compatible VCS that is both simple and powerful",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"lazygit": {
		Binaries:    []Binary{{Name: "lg", RenameTo: ""}},
		Owner:       "jesseduffield",
		Repository:  "lazygit",
		Description: "Simple terminal UI for git commands",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_x86_64\\.tar\\.gz$",
			"windows/amd64": "Windows_x86_64\\.zip$",
		},
	},

	// Text editors
	"micro": {
		Binaries:    []Binary{{Name: "micro", RenameTo: ""}},
		Owner:       "zyedidia",
		Repository:  "micro",
		Description: "A modern and intuitive terminal-based text editor",
		AssetNames: map[string]string{
			"linux/amd64":   "linux64\\.tar\\.gz$",
			"windows/amd64": "win64\\.zip$",
		},
	},

	// Multimedia
	"freeze": {
		Binaries:    []Binary{{Name: "freeze", RenameTo: ""}},
		Owner:       "charmbracelet",
		Repository:  "freeze",
		Description: "Generate images of code and terminal output",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_x86_64\\.tar\\.gz$",
			"windows/amd64": "Windows_x86_64\\.zip$",
		},
	},
	"vhs": {
		Binaries:    []Binary{{Name: "vhs", RenameTo: ""}},
		Owner:       "charmbracelet",
		Repository:  "vhs",
		Description: "Your CLI home video recorder",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_x86_64\\.tar\\.gz$",
			"windows/amd64": "Windows_x86_64\\.zip$",
		},
	},
	"yt-dlp": {
		Binaries:    []Binary{{Name: "yt-dlp", RenameTo: ""}},
		Owner:       "yt-dlp",
		Repository:  "yt-dlp",
		Description: "A feature-rich command-line audio/video downloader",
		AssetNames: map[string]string{
			"linux/amd64":   "yt-dlp_linux$",
			"windows/amd64": "yt-dlp\\.exe$",
		},
	},

	// Cryptography
	"age": {
		Binaries: []Binary{
			{Name: "age", RenameTo: ""},
			{Name: "age-keygen", RenameTo: ""},
		},
		Owner:       "FiloSottile",
		Repository:  "age",
		Description: "A simple, modern and secure encryption tool (and Go library) with small explicit keys, no config options, and UNIX-style composability.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-amd64\\.tar\\.gz$",
			"windows/amd64": "windows-amd64\\.zip$",
		},
	},
	"minisign": {
		Binaries:    []Binary{{Name: "minisign", RenameTo: ""}},
		Owner:       "jedisct1",
		Repository:  "minisign",
		Description: "A dead simple tool to sign files and verify digital signatures.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux\\.tar\\.gz$",
			"windows/amd64": "win64\\.zip$",
		},
	},
	"mkcert": {
		Binaries:    []Binary{{Name: "mkcert", RenameTo: ""}},
		Owner:       "FiloSottile",
		Repository:  "mkcert",
		Description: "A simple zero-config tool to make locally trusted development certificates with any names you'd like.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-amd64",
			"windows/amd64": "windows-arm64\\.exe$",
		},
	},

	// AI
	"codex": {
		Binaries: []Binary{
			{Name: "codex-x86_64-pc-windows-msvc", RenameTo: "codex"},
			{Name: "codex-x86_64-unknown-linux-gnu", RenameTo: "codex"},
		},
		Owner:       "openai",
		Repository:  "codex",
		Description: "Lightweight coding agent that runs in your terminal",
		AssetNames: map[string]string{
			"linux/amd64":   "codex-x86_64-unknown-linux-gnu.tar\\.gz$",
			"windows/amd64": "codex-x86_64-pc-windows-msvc\\.exe$\\.zip$",
		},
	},
	"crush": {
		Binaries:    []Binary{{Name: "crush", RenameTo: ""}},
		Owner:       "charmbracelet",
		Repository:  "crush",
		Description: "The glamourous AI coding agent for your favourite terminal",
		AssetNames: map[string]string{
			"linux/amd64":   "Linux_x86_64\\.tar\\.gz$",
			"windows/amd64": "Windows_x86_64\\.zip$",
		},
	},
	"opencode": {
		Binaries:    []Binary{{Name: "opencode", RenameTo: ""}},
		Owner:       "sst",
		Repository:  "opencode",
		Description: "The open source coding agent.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x64\\.tar\\.gz$",
			"windows/amd64": "windows-x64\\.zip$",
		},
	},

	// Miscellaneous
	"hledger": {
		Binaries: []Binary{
			{Name: "hledger", RenameTo: ""},
			{Name: "hledger-ui", RenameTo: ""},
			{Name: "hledger-web", RenameTo: ""},
		},
		Owner:       "simonmichael",
		Repository:  "hledger",
		Description: "Robust, fast, intuitive plain text accounting tool with CLI, TUI and web interfaces.",
		AssetNames: map[string]string{
			"linux/amd64":   "linux-x64\\.zip$",
			"windows/amd64": "windows-x64\\.zip$",
		},
	},
	"hyperfine": {
		Binaries:    []Binary{{Name: "hyperfine", RenameTo: ""}},
		Owner:       "sharkdp",
		Repository:  "hyperfine",
		Description: "A command-line benchmarking tool",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": defaultWindows64,
		},
	},
	"miniserve": {
		Binaries:    []Binary{{Name: "miniserve", RenameTo: ""}},
		Owner:       "svenstaro",
		Repository:  "miniserve",
		Description: "For when you really just want to serve some files over HTTP right now!",
		AssetNames: map[string]string{
			"linux/amd64":   "x86_64-unknown-linux-gnu$",
			"windows/amd64": "x86_64-pc-windows-msvc\\.exe$",
		},
	},
	"tealdeer": {
		Binaries:    []Binary{{Name: "tealdeer", RenameTo: "tldr"}},
		Owner:       "dbrgn",
		Repository:  "tealdeer",
		Description: "A very fast implementation of tldr in Rust.",
		AssetNames: map[string]string{
			"linux/amd64":   "tealdeer-linux-x86_64-musl$",
			"windows/amd64": "windows-x86_64-msvc\\.exe$",
		},
	},
	"tokei": {
		Binaries:    []Binary{{Name: "tokei", RenameTo: ""}},
		Owner:       "XAMPPRocky",
		Repository:  "tokei",
		Description: "Count your code, quickly.",
		AssetNames: map[string]string{
			"linux/amd64":   defaultLinux64,
			"windows/amd64": "x86_64-pc-windows-msvc\\.exe$",
		},
	},
	"tailwind": {
		Binaries:    []Binary{{Name: "tailwind", RenameTo: ""}},
		Owner:       "tailwindlabs",
		Repository:  "tailwindcss",
		Description: "Tailwind CSS standalone CLI tool",
		AssetNames: map[string]string{
			"linux/amd64":   "tailwindcss-linux-x64$",
			"windows/amd64": "tailwindcss-windows-x64\\.exe$",
		},
	},
}
