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

	"github.com/ulikunitz/xz/v2"
)

type AssetType int

const (
	ZipArchive AssetType = iota
	TarGzArchive
	TarXzArchive
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
		if file.FileInfo().IsDir() {
			continue
		}

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

	return extractFromTar(gzipReader, binaries, outputPath)
}

func extractFilesTarXz(rawData []byte, binaries []Binary, outputPath string) error {
	byteReader := bytes.NewReader(rawData)

	xzReader, err := xz.NewReader(byteReader)
	if err != nil {
		return err
	}
	defer xzReader.Close()

	return extractFromTar(xzReader, binaries, outputPath)
}

func extractFromTar(uncompressReader io.Reader, binaries []Binary, outputPath string) error {
	tarReader := tar.NewReader(uncompressReader)

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

		if header.Typeflag != tar.TypeReg {
			continue
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

func hasArchiveSuffix(assetName string) bool {
	assetName = strings.ToLower(assetName)
	return strings.HasSuffix(assetName, ".zip") || strings.HasSuffix(assetName, ".tar.gz") || strings.HasSuffix(assetName, ".tar.xz")
}

func isZip(rawData []byte) bool {
	_, err := zip.NewReader(bytes.NewReader(rawData), int64(len(rawData)))
	return err == nil
}

func isTarGz(rawData []byte) bool {
	gzipReader, err := gzip.NewReader(bytes.NewReader(rawData))
	if err != nil {
		return false
	}
	defer gzipReader.Close()

	_, err = tar.NewReader(gzipReader).Next()
	return err == nil
}

func isTarXz(rawData []byte) bool {
	r, err := xz.NewReader(bytes.NewReader(rawData))
	if err != nil {
		return false
	}
	defer r.Close()

	_, err = tar.NewReader(r).Next()
	return err == nil
}

func detectAssetType(rawData []byte, assetName string) (AssetType, error) {
	if isZip(rawData) {
		return ZipArchive, nil
	}

	if isTarGz(rawData) {
		return TarGzArchive, nil
	}

	if isTarXz(rawData) {
		return TarXzArchive, nil
	}

	if hasArchiveSuffix(assetName) {
		return RawBinary, fmt.Errorf("asset %q has an archive suffix but is not a valid zip, tar.gz, or tar.xz archive", assetName)
	}

	return RawBinary, nil
}

func extractFiles(rawData []byte, assetName string, binaries []Binary, outputPath string) (AssetType, error) {
	detectedType, err := detectAssetType(rawData, assetName)
	if err != nil {
		return RawBinary, err
	}

	switch detectedType {
	case ZipArchive:
		return detectedType, extractFilesZip(rawData, binaries, outputPath)
	case TarGzArchive:
		return detectedType, extractFilesTarGz(rawData, binaries, outputPath)
	case TarXzArchive:
		return detectedType, extractFilesTarXz(rawData, binaries, outputPath)
	default:
		return detectedType, extractFilesRaw(rawData, binaries, outputPath)
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
		if file.FileInfo().IsDir() {
			continue
		}

		result = append(result, path.Base(file.Name))
	}

	return result, nil
}

func getFilesNamesTarGz(rawData []byte) ([]string, error) {
	byteReader := bytes.NewReader(rawData)

	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return make([]string, 0), err
	}
	defer gzipReader.Close()

	return getFilesNamesFromTar(gzipReader)
}

func getFilesNamesTarXz(rawData []byte) ([]string, error) {
	byteReader := bytes.NewReader(rawData)

	xzReader, err := xz.NewReader(byteReader)
	if err != nil {
		return make([]string, 0), err
	}
	defer xzReader.Close()

	return getFilesNamesFromTar(xzReader)
}

func getFilesNamesFromTar(uncompressReader io.Reader) ([]string, error) {
	result := make([]string, 0)

	tarReader := tar.NewReader(uncompressReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		result = append(result, path.Base(header.Name))
	}

	return result, nil
}

func getBinaryFileNames(rawData []byte, assetName string) ([]string, error) {
	var files []string

	detectedType, err := detectAssetType(rawData, assetName)
	if err != nil {
		return nil, err
	}

	switch detectedType {
	case ZipArchive:
		files, err = getFilesNamesZip(rawData)
	case TarGzArchive:
		files, err = getFilesNamesTarGz(rawData)
	case TarXzArchive:
		files, err = getFilesNamesTarXz(rawData)
	default:
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
