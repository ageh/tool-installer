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
	"strings"
)

func getRenameTarget(fullName string, binaries []Binary) string {
	if strings.HasSuffix(fullName, "/") {
		return ""
	}

	fileName := path.Base(fullName)

	for _, binary := range binaries {
		if fileName == binary.Name {
			if binary.RenameTo != "" {
				return binary.RenameTo
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
		defer fileReader.Close()

		fileContent, err := io.ReadAll(fileReader)
		if err != nil {
			return err
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
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}

		os.Chmod(filePath, 0755)

		extracted++
		if extracted == toExtract {
			break
		}
	}

	return nil
}

func extractFilesRaw(rawData []byte, binaries []Binary, outputPath string) error {
	if len(binaries) != 1 {
		return errors.New("invalid number of binaries provided. Non-archive type assets can only be one binary")
	}

	fileName := binaries[0].Name
	if binaries[0].RenameTo != "" {
		fileName = binaries[0].RenameTo
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

	os.Chmod(filePath, 0755)

	return nil
}

func extractFiles(rawData []byte, assetName string, binaries []Binary, outputPath string) error {
	if strings.HasSuffix(assetName, ".tar.gz") {
		return extractFilesTarGz(rawData, binaries, outputPath)
	} else if strings.HasSuffix(assetName, ".zip") {
		return extractFilesZip(rawData, binaries, outputPath)
	} else {
		fmt.Println("Warning: The asset does not have a file ending. While this can be legitimate, you should probably talk to the tool author to see if he is willing to change that.")
		return extractFilesRaw(rawData, binaries, outputPath)
	}
}
