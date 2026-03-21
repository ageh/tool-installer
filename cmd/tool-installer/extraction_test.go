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
)

const testBinaryContent = "definitely not an executable"

var testArchiveContent = map[string][]byte{
	"tooli":                []byte(testBinaryContent),
	"ignored":              []byte(testBinaryContent),
	"README.md":            []byte(testBinaryContent),
	"LICENSE":              []byte(testBinaryContent),
	"completions/shell.sh": []byte(testBinaryContent),
}

var expectedBinaries = []Binary{{Name: "tooli"}}
var expectedBinariesRename = []Binary{{Name: "tooli", RenameTo: "tool-installer"}}
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

func TestGetRenameTarget(t *testing.T) {
	tests := []struct {
		name           string
		fullBinaryName string
		binaries       []Binary
		expectedName   string
	}{
		{"Directory", "some_directory/", []Binary{{Name: "unrelated", RenameTo: "ignored"}}, ""},
		{"Without rename", "tooli", []Binary{{Name: "unrelated", RenameTo: "ignored"}, {Name: "tooli"}}, "tooli"},
		{"With rename", "tooli", []Binary{{Name: "unrelated", RenameTo: "ignored"}, {Name: "tooli", RenameTo: "tool-installer"}}, "tool-installer"},
		{"No match", "tooli", []Binary{{Name: "unrelated", RenameTo: "ignored"}, {Name: "tool-installer", RenameTo: ""}}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renamed := getRenameTarget(test.fullBinaryName, test.binaries)
			if renamed != test.expectedName {
				t.Errorf("wrong rename target: got %q but expected %q", renamed, test.expectedName)
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
			t.Errorf("wrong file content: got %q but expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract expected binary with rename", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesZip(data, expectedBinariesRename, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "tool-installer"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q but expected %q", result, testBinaryContent)
		}
	})

	t.Run("Missing binary", func(t *testing.T) {
		data := makeZip(t, testArchiveContent)
		err := extractFilesZip(data, expectedBinariesMissing, t.TempDir())
		if err == nil {
			t.Error("expected error for missing binary, got nil")
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
			t.Errorf("wrong file content: got %q but expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract expected binary with rename", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesTarGz(data, expectedBinariesRename, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "tool-installer"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q but expected %q", result, testBinaryContent)
		}
	})

	t.Run("Missing binary", func(t *testing.T) {
		err := extractFilesTarGz(data, expectedBinariesMissing, t.TempDir())
		if err == nil {
			t.Error("expected error for missing binary, got nil")
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
			t.Errorf("wrong file content: got %q but expected %q", result, testBinaryContent)
		}
	})

	t.Run("Extract raw binary with rename", func(t *testing.T) {
		outDir := t.TempDir()

		err := extractFilesRaw(data, expectedBinariesRename, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := os.ReadFile(filepath.Join(outDir, "tool-installer"))
		if err != nil {
			t.Fatalf("expected file not found: %v", err)
		}

		if string(result) != testBinaryContent {
			t.Errorf("wrong file content: got %q but expected %q", result, testBinaryContent)
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
		name            string
		assetNameSuffix string
		expectedType    AssetType
	}{
		{"Zip", ".zip", Archive},
		{"Tar.gz", ".tar.gz", Archive},
		{"Raw", "", RawBinary},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assetName := "tooli-windows-x86_64" + test.assetNameSuffix

			var data []byte
			switch test.assetNameSuffix {
			case ".zip":
				data = makeZip(t, testArchiveContent)
			case ".tar.gz":
				data = makeTarGz(t, testArchiveContent)
			default:
				data = testArchiveContent["tooli"]
			}

			assetType, err := extractFiles(data, assetName, expectedBinaries, t.TempDir())
			if err != nil {
				t.Errorf("unexpected error extracting: %v", err)
			}

			if assetType != test.expectedType {
				t.Errorf("wrong detected asset type: got %d but expected %d", assetType, test.expectedType)
			}
		})
	}
}

func TestExtractBinaryFileNames(t *testing.T) {
	tests := []struct {
		name              string
		assetNameSuffix   string
		expectedFilenames []string
	}{
		{"Zip", ".zip", []string{"tooli", "ignored"}},
		{"Tar.gz", ".tar.gz", []string{"tooli", "ignored"}},
		{"Raw", "", []string{"tooli-windows-x86_64"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assetName := "tooli-windows-x86_64" + test.assetNameSuffix

			var data []byte
			switch test.assetNameSuffix {
			case ".zip":
				data = makeZip(t, testArchiveContent)
			case ".tar.gz":
				data = makeTarGz(t, testArchiveContent)
			default:
				data = testArchiveContent["tooli"]
			}

			files, err := getBinaryFileNames(data, assetName)
			if err != nil {
				t.Fatalf("unexpected error reading fileNames: %v", err)
			}

			slices.Sort(test.expectedFilenames)
			if !slices.Equal(files, test.expectedFilenames) {
				t.Errorf("mismatch in extracted filenames, got %v but expected %v", files, test.expectedFilenames)
			}
		})
	}
}
