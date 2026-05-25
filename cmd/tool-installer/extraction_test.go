// SPDX-License-Identifier: Apache-2.0

package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ulikunitz/xz/v2"
)

const testBinaryContent = "definitely not an executable"

var testArchiveContent = map[string][]byte{
	"tooli":                []byte(testBinaryContent),
	"ignored":              []byte(testBinaryContent),
	"tooli-helper":         []byte(testBinaryContent),
	"README.md":            []byte(testBinaryContent),
	"LICENSE":              []byte(testBinaryContent),
	"completions/shell.sh": []byte(testBinaryContent),
}

var expectedBinaries = []Binary{{Name: "tooli"}}
var expectedBinariesRename = []Binary{{Name: "helpi", SourceNames: []string{"tooli-helper"}}}
var expectedBinariesMissing = []Binary{{Name: "tooli"}, {Name: "tooli2"}, {Name: "tooli3"}}

func makeZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	keys := slices.Sorted(maps.Keys(files))
	for _, name := range keys {
		content := files[name]
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("unexpected error creating zip entry: %v", err)
		}

		if _, err := f.Write(content); err != nil {
			t.Fatalf("unexpected error writing zip content: %v", err)
		}
	}

	w.Close()

	return buf.Bytes()
}

func makeZipWithDirectory(t *testing.T, name string) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	header := &zip.FileHeader{Name: name}
	header.SetMode(os.ModeDir | 0755)
	if _, err := w.CreateHeader(header); err != nil {
		t.Fatalf("unexpected error creating zip directory entry: %v", err)
	}

	w.Close()

	return buf.Bytes()
}

func makeTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	keys := slices.Sorted(maps.Keys(files))
	for _, name := range keys {
		content := files[name]
		err := tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(content)), Mode: 0755})
		if err != nil {
			t.Fatalf("unexpected error writing tar header: %v", err)
		}

		if _, err := tw.Write(content); err != nil {
			t.Fatalf("unexpected error writing tar content: %v", err)
		}
	}

	tw.Close()
	gw.Close()

	return buf.Bytes()
}

func makeEmptyTarGz(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	tw.Close()
	gw.Close()

	return buf.Bytes()
}

func makeTarGzWithDirectory(t *testing.T, name string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeDir, Mode: 0755}); err != nil {
		t.Fatalf("unexpected error writing tar directory header: %v", err)
	}

	tw.Close()
	gw.Close()

	return buf.Bytes()
}

func makeTarXz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	xw, err := xz.NewWriter(&buf)
	if err != nil {
		t.Fatalf("unexpected error creating xz writer: %v", err)
	}
	tw := tar.NewWriter(xw)

	keys := slices.Sorted(maps.Keys(files))
	for _, name := range keys {
		content := files[name]
		err := tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(content)), Mode: 0755})
		if err != nil {
			t.Fatalf("unexpected error writing tar header: %v", err)
		}

		if _, err := tw.Write(content); err != nil {
			t.Fatalf("unexpected error writing tar content: %v", err)
		}
	}

	tw.Close()
	xw.Close()

	return buf.Bytes()
}

func makeEmptyTarXz(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	xw, err := xz.NewWriter(&buf)
	if err != nil {
		t.Fatalf("unexpected error creating xz writer: %v", err)
	}
	tw := tar.NewWriter(xw)

	tw.Close()
	xw.Close()

	return buf.Bytes()
}

func makeTarXzWithDirectory(t *testing.T, name string) []byte {
	t.Helper()

	var buf bytes.Buffer
	xw, err := xz.NewWriter(&buf)
	if err != nil {
		t.Fatalf("unexpected error creating xz writer: %v", err)
	}
	tw := tar.NewWriter(xw)

	if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeDir, Mode: 0755}); err != nil {
		t.Fatalf("unexpected error writing tar directory header: %v", err)
	}

	tw.Close()
	xw.Close()

	return buf.Bytes()
}

func TestGetRenameTarget(t *testing.T) {
	tests := []struct {
		name           string
		fullBinaryName string
		binaries       []Binary
		expectedName   string
	}{
		{"Directory", "some_directory/", []Binary{{Name: "unrelated"}}, ""},
		{"Without rename", "tooli", []Binary{{Name: "unrelated"}, {Name: "tooli"}}, "tooli"},
		{"With rename", "tooli", []Binary{{Name: "unrelated"}, {Name: "tooli", SourceNames: []string{"tool-installer"}}}, "tooli"},
		{"No match", "tooli", []Binary{{Name: "unrelated"}, {Name: "tool-installer"}}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renamed := getRenameTarget(test.fullBinaryName, test.binaries)
			if renamed != test.expectedName {
				t.Errorf("wrong rename target: got %q, expected %q", renamed, test.expectedName)
			}
		})
	}
}

func TestExtractFilesZip(t *testing.T) {
	data := makeZip(t, testArchiveContent)

	t.Run("Extract expected binary", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesZip(data, expectedBinaries, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "tooli"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract expected binary with rename", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesZip(data, expectedBinariesRename, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "helpi"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Missing binary", func(t *testing.T) {
		data := makeZip(t, testArchiveContent)
		err := extractFilesZip(data, expectedBinariesMissing, t.TempDir())
		if err == nil {
			t.Error("expected error for missing binary, got nil")
		}
	})

	t.Run("Directory does not count as extracted binary", func(t *testing.T) {
		err := extractFilesZip(makeZipWithDirectory(t, "tooli"), expectedBinaries, t.TempDir())
		if err == nil {
			t.Error("expected error for directory matching binary name, got nil")
		}
	})
}

func TestExtractFilesTarGz(t *testing.T) {
	data := makeTarGz(t, testArchiveContent)

	t.Run("Extract expected binary", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesTarGz(data, expectedBinaries, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "tooli"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract expected binary with rename", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesTarGz(data, expectedBinariesRename, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "helpi"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Missing binary", func(t *testing.T) {
		err := extractFilesTarGz(data, expectedBinariesMissing, t.TempDir())
		if err == nil {
			t.Error("expected error for missing binary, got nil")
		}
	})

	t.Run("Directory does not count as extracted binary", func(t *testing.T) {
		err := extractFilesTarGz(makeTarGzWithDirectory(t, "tooli"), expectedBinaries, t.TempDir())
		if err == nil {
			t.Error("expected error for directory matching binary name, got nil")
		}
	})
}

func TestExtractFilesTarXz(t *testing.T) {
	data := makeTarXz(t, testArchiveContent)

	t.Run("Extract expected binary", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesTarXz(data, expectedBinaries, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "tooli"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract expected binary with rename", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesTarXz(data, expectedBinariesRename, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "helpi"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Missing binary", func(t *testing.T) {
		err := extractFilesTarXz(data, expectedBinariesMissing, t.TempDir())
		if err == nil {
			t.Error("expected error for missing binary, got nil")
		}
	})

	t.Run("Directory does not count as extracted binary", func(t *testing.T) {
		err := extractFilesTarXz(makeTarXzWithDirectory(t, "tooli"), expectedBinaries, t.TempDir())
		if err == nil {
			t.Error("expected error for directory matching binary name, got nil")
		}
	})
}

func TestExtractFilesRaw(t *testing.T) {
	data := testArchiveContent["tooli"]

	t.Run("Extract raw binary", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesRaw(data, expectedBinaries, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "tooli"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract raw binary with rename", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesRaw(data, expectedBinariesRename, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "helpi"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q, expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract raw binary: too many expected", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesRaw(data, expectedBinariesMissing, outDir)
		if err == nil {
			t.Error("expected error for extracting more than one raw binary but got nil")
		}
	})
}

func TestExtractFiles(t *testing.T) {
	tests := []struct {
		name         string
		assetName    string
		data         []byte
		expectedType AssetType
		shouldError  bool
	}{
		{"Zip", "tooli-windows-x86_64.zip", makeZip(t, testArchiveContent), ZipArchive, false},
		{"Zip content without suffix", "tooli-windows-x86_64", makeZip(t, testArchiveContent), ZipArchive, false},
		{"Tar.gz", "tooli-windows-x86_64.tar.gz", makeTarGz(t, testArchiveContent), TarGzArchive, false},
		{"Tar.gz content without suffix", "tooli-windows-x86_64", makeTarGz(t, testArchiveContent), TarGzArchive, false},
		{"Tar.xz", "tooli-windows-x86_64.tar.xz", makeTarXz(t, testArchiveContent), TarXzArchive, false},
		{"Tar.xz content without suffix", "tooli-windows-x86_64", makeTarXz(t, testArchiveContent), TarXzArchive, false},
		{"Raw exe", "tooli-windows-x86_64.exe", testArchiveContent["tooli"], RawBinary, false},
		{"Raw without suffix", "tooli-linux-x86_64", testArchiveContent["tooli"], RawBinary, false},
		{"Invalid zip suffix", "tooli-windows-x86_64.zip", testArchiveContent["tooli"], RawBinary, true},
		{"Invalid uppercase zip suffix", "tooli-windows-x86_64.ZIP", testArchiveContent["tooli"], RawBinary, true},
		{"Invalid tar.gz suffix", "tooli-windows-x86_64.tar.gz", testArchiveContent["tooli"], RawBinary, true},
		{"Invalid tar.xz suffix", "tooli-windows-x86_64.tar.xz", testArchiveContent["tooli"], RawBinary, true},
		{"Empty tar.gz", "tooli-windows-x86_64.tar.gz", makeEmptyTarGz(t), RawBinary, true},
		{"Empty tar.xz", "tooli-windows-x86_64.tar.xz", makeEmptyTarXz(t), RawBinary, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assetType, err := extractFiles(test.data, test.assetName, expectedBinaries, t.TempDir())
			if err != nil {
				if test.shouldError {
					return
				}
				t.Errorf("unexpected error extracting: %v", err)
			}

			if test.shouldError {
				t.Fatal("expected error but got nil")
			}

			if assetType != test.expectedType {
				t.Errorf("wrong detected asset type: got %d, expected %d", assetType, test.expectedType)
			}
		})
	}
}

func TestExtractBinaryFileNames(t *testing.T) {
	tests := []struct {
		name              string
		assetName         string
		data              []byte
		expectedFilenames []string
		shouldError       bool
	}{
		{"Zip", "tooli-windows-x86_64.zip", makeZip(t, testArchiveContent), []string{"tooli", "ignored", "tooli-helper"}, false},
		{"Zip content without suffix", "tooli-windows-x86_64", makeZip(t, testArchiveContent), []string{"tooli", "ignored", "tooli-helper"}, false},
		{"Tar.gz", "tooli-windows-x86_64.tar.gz", makeTarGz(t, testArchiveContent), []string{"tooli", "ignored", "tooli-helper"}, false},
		{"Tar.gz content without suffix", "tooli-windows-x86_64", makeTarGz(t, testArchiveContent), []string{"tooli", "ignored", "tooli-helper"}, false},
		{"Tar.xz", "tooli-windows-x86_64.tar.xz", makeTarXz(t, testArchiveContent), []string{"tooli", "ignored", "tooli-helper"}, false},
		{"Tar.xz content without suffix", "tooli-windows-x86_64", makeTarXz(t, testArchiveContent), []string{"tooli", "ignored", "tooli-helper"}, false},
		{"Raw without suffix", "tooli-windows-x86_64", testArchiveContent["tooli"], []string{"tooli-windows-x86_64"}, false},
		{"Raw exe", "tooli-windows-x86_64.exe", testArchiveContent["tooli"], []string{"tooli-windows-x86_64.exe"}, false},
		{"Invalid zip suffix", "tooli-windows-x86_64.zip", testArchiveContent["tooli"], nil, true},
		{"Invalid uppercase zip suffix", "tooli-windows-x86_64.ZIP", testArchiveContent["tooli"], nil, true},
		{"Invalid tar.gz suffix", "tooli-windows-x86_64.tar.gz", testArchiveContent["tooli"], nil, true},
		{"Invalid tar.xz suffix", "tooli-windows-x86_64.tar.xz", testArchiveContent["tooli"], nil, true},
		{"Empty tar.gz", "tooli-windows-x86_64.tar.gz", makeEmptyTarGz(t), nil, true},
		{"Empty tar.xz", "tooli-windows-x86_64.tar.xz", makeEmptyTarXz(t), nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files, err := getBinaryFileNames(test.data, test.assetName)
			if err != nil {
				if test.shouldError {
					return
				}
				t.Fatalf("unexpected error reading filenames: %v", err)
			}

			if test.shouldError {
				t.Fatal("expected error but got nil")
			}

			slices.Sort(files)
			slices.Sort(test.expectedFilenames)
			if !slices.Equal(files, test.expectedFilenames) {
				t.Errorf("mismatch in extracted filenames: got %v, expected %v", files, test.expectedFilenames)
			}
		})
	}
}
