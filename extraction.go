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
	"runtime"
	"strings"
)

func getRenameTarget(fullName string, binaries []Binary) string {
	fileName := path.Base(fullName)

	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(fileName, ".exe") {
			fileName = fileName + ".exe"
		}
	}

	res := ""

	for _, binary := range binaries {
		if fileName == binary.Name {
			if binary.RenameTo != "" {
				res = binary.RenameTo

				if runtime.GOOS == "windows" && !strings.HasSuffix(res, ".exe") {
					res = res + ".exe"
				}

				break
			} else {
				res = fileName
				break
			}
		}
	}

	return res
}

func extractFilesZip(rawData []byte, binaries []Binary, outputPath *string) error {
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
		defer fileReader.Close()
		if err != nil {
			return err
		}

		fileContent, err := io.ReadAll(fileReader)
		if err != nil {
			return err
		}

		filePath := filepath.Join(*outputPath, fileName)

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

func extractFilesTarGz(rawData []byte, binaries []Binary, outputPath *string) error {
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

		filePath := filepath.Join(*outputPath, fileName)

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

func extractFilesRaw(rawData []byte, binaries []Binary, outputPath *string) error {
	if len(binaries) != 1 {
		return errors.New("Invalid number of binaries provided")
	}

	fileName := binaries[0].Name
	if runtime.GOOS == "windows" {
		fileName = fileName + ".exe"
	}

	if fileName == "" {
		return errors.New("Invalid filename, cannot be blank")
	}

	filePath := filepath.Join(*outputPath, fileName)

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

func extractFiles(rawData []byte, asset *Asset, tool *Tool, outputPath *string) error {
	if strings.HasSuffix(asset.Name, ".tar.gz") {
		return extractFilesTarGz(rawData, tool.Binaries, outputPath)
	} else if strings.HasSuffix(asset.Name, ".zip") {
		return extractFilesZip(rawData, tool.Binaries, outputPath)
	} else {
		fmt.Println("WARNING: The asset does not have a file ending. While this can be legitimate, you should probably talk to the tool author to see if he is willing to change that.")
		return extractFilesRaw(rawData, tool.Binaries, outputPath)
	}

	return nil
}
