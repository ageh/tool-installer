// SPDX-License-Identifier: Apache-2.0

package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type AssetType int

const (
	Archive AssetType = iota
	RawBinary
)

var ignoredExtensions = []string{".md", ".txt", ".sh", ".bash", ".fish", ".ps1", ".bat", ".1", ".json"}
var ignoredFiles = []string{"COPYING", "CONTRIBUTING", "CONTRIBUTORS", "NOTICE", "AUTHORS", "Makefile", "VERSION"}

func isKnownNonBinaryFile(fileName string) bool {
	if strings.HasPrefix(fileName, "_") || strings.Contains(fileName, "LICENSE") || strings.Contains(fileName, "LICENCE") {
		return true
	}

	if slices.Contains(ignoredFiles, fileName) {
		return true
	}

	return slices.Contains(ignoredExtensions, filepath.Ext(fileName))
}

func getRenameTarget(fullName string, binaries []Binary) string {
	if strings.HasSuffix(fullName, "/") {
		return ""
	}

	fileName := path.Base(fullName)

	for _, binary := range binaries {
		if fileName == binary.Name {
			if binary.RenameTo != "" {
				return filepath.Base(binary.RenameTo)
			} else {
				return fileName
			}
		}
	}

	return ""
}

func extractFilesZip(rawData []byte, binaries []Binary, outputPath string) error {
	byteReader := bytes.NewReader(rawData)

	zipReader, err := zip.NewReader(byteReader, int64(len(rawData)))
	if err != nil {
		return err
	}

	toExtract := len(binaries)
	extracted := 0

	for _, file := range zipReader.File {
		fileName := getRenameTarget(file.Name, binaries)
		if fileName == "" {
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		fileContent, err := io.ReadAll(fileReader)
		closeErr := fileReader.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}

		filePath := filepath.Join(outputPath, fileName)

		err = os.WriteFile(filePath, fileContent, 0755)
		if err != nil {
			return err
		}

		extracted++
		if extracted == toExtract {
			break
		}
	}

	if extracted != toExtract {
		return fmt.Errorf("only extracted %d of %d expected binaries", extracted, toExtract)
	}

	return nil
}

func extractFilesTarGz(rawData []byte, binaries []Binary, outputPath string) error {
	byteReader := bytes.NewReader(rawData)

	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	toExtract := len(binaries)
	extracted := 0

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fileName := getRenameTarget(header.Name, binaries)
		if fileName == "" {
			continue
		}

		filePath := filepath.Join(outputPath, fileName)

		file, err := os.Create(filePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, tarReader)
		closeErr := file.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}

		err = os.Chmod(filePath, 0755)
		if err != nil {
			return err
		}

		extracted++
		if extracted == toExtract {
			break
		}
	}

	if extracted != toExtract {
		return fmt.Errorf("only extracted %d of %d expected binaries", extracted, toExtract)
	}

	return nil
}

func extractFilesRaw(rawData []byte, binaries []Binary, outputPath string) error {
	if len(binaries) != 1 {
		return errors.New("invalid number of binaries provided. Non-archive type assets can only be one binary")
	}

	fileName := binaries[0].Name
	if binaries[0].RenameTo != "" {
		fileName = filepath.Base(binaries[0].RenameTo)
	}

	filePath := filepath.Join(outputPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	byteReader := bytes.NewReader(rawData)

	_, err = io.Copy(file, byteReader)
	if err != nil {
		return err
	}

	return os.Chmod(filePath, 0755)
}

func extractFiles(rawData []byte, assetName string, binaries []Binary, outputPath string) (AssetType, error) {
	if strings.HasSuffix(assetName, ".tar.gz") {
		return Archive, extractFilesTarGz(rawData, binaries, outputPath)
	} else if strings.HasSuffix(assetName, ".zip") {
		return Archive, extractFilesZip(rawData, binaries, outputPath)
	} else {
		return RawBinary, extractFilesRaw(rawData, binaries, outputPath)
	}
}

func getFilesNamesZip(rawData []byte) ([]string, error) {
	result := make([]string, 0)

	byteReader := bytes.NewReader(rawData)

	zipReader, err := zip.NewReader(byteReader, int64(len(rawData)))
	if err != nil {
		return result, err
	}

	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		result = append(result, path.Base(file.Name))
	}

	return result, nil
}

func getFilesNamesTarGz(rawData []byte) ([]string, error) {
	result := make([]string, 0)

	byteReader := bytes.NewReader(rawData)

	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return result, err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}

		if strings.HasSuffix(header.Name, "/") {
			continue
		}

		result = append(result, path.Base(header.Name))
	}

	return result, nil
}

func getBinaryFileNames(rawData []byte, assetName string) ([]string, error) {
	var files []string
	var err error

	if strings.HasSuffix(assetName, ".tar.gz") {
		files, err = getFilesNamesTarGz(rawData)
	} else if strings.HasSuffix(assetName, ".zip") {
		files, err = getFilesNamesZip(rawData)
	} else {
		files, err = []string{assetName}, nil
	}

	result := make([]string, 0, len(files))

	for _, file := range files {
		if isKnownNonBinaryFile(file) {
			continue
		}

		result = append(result, file)
	}

	return result, err
}
