// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func prompt(text string) string {
	fmt.Print(text)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return ""
	}

	return strings.TrimSpace(input)
}

func promptNonEmpty(text string) string {
	fmt.Print(text)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return ""
		}

		result := strings.TrimSpace(input)

		if result != "" {
			return result
		}

		fmt.Print("Input must not be empty. Please try again: ")
	}
}

func promptRegex(text string) (*regexp.Regexp, string) {
	fmt.Print(text)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return nil, ""
		}

		result := strings.TrimSpace(input)

		rx, err := regexp.Compile(result)
		if err == nil {
			return rx, result
		}

		fmt.Print("Input must be a valid regular expression. Please try again: ")
	}
}

func promptForBinary() (Binary, bool) {
	binary := prompt("Binary name: ")
	if binary == "" {
		return Binary{}, false
	}

	rename := prompt("Rename binary to (leave empty if no rename): ")

	return Binary{Name: strings.TrimSuffix(binary, ".exe"), RenameTo: strings.TrimSuffix(rename, ".exe")}, true
}

func promptForBinaries(fileNames []string) []Binary {
	result := make([]Binary, 0)

	if len(fileNames) != 0 {
		fmt.Println("Found some files inside the asset, please enter which ones are the binaries to add.")
		fmt.Println("For each file, input a name if it is a binary and should be renamed, press enter if it is a binary to add without a rename, and input 's' if you want to skip the file")

		for _, file := range fileNames {
			rename := prompt(fmt.Sprintf("For file %q:", file))

			if rename == "s" || rename == "S" {
				continue
			}

			result = append(result, Binary{Name: strings.TrimSuffix(file, ".exe"), RenameTo: strings.TrimSuffix(rename, ".exe")})
		}
	}

	if len(result) != 0 {
		return result
	}

	fmt.Println("Please provide the binaries to extract from the asset. Pressing enter (empty binary name) will finish the process:")

	binary := promptNonEmpty("Binary name: ")
	rename := prompt("Rename binary to (leave empty if no rename): ")

	result = append(result, Binary{Name: binary, RenameTo: rename})

	for {
		binary, ok := promptForBinary()

		if !ok {
			break
		}

		result = append(result, binary)
	}

	return result
}

func promptForUniqueAssetRegex(assets []Asset) string {
	assetNames := make([]string, len(assets))

	for i, asset := range assets {
		assetNames[i] = asset.Name
	}

	for {
		regex, pattern := promptRegex("Enter asset regex: ")

		matches := make([]string, 0)
		for _, name := range assetNames {
			if regex.MatchString(name) {
				matches = append(matches, name)
			}
		}

		if len(matches) == 1 {
			return pattern
		}

		fmt.Println("The provided pattern matches more than one asset name, please be more specific.")
		fmt.Println("The following asset names would be matched by this regex:")
		for _, name := range matches {
			fmt.Printf("  - %s", name)
		}
	}
}
