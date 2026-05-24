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

func promptForBinary(allowEmpty bool) (Binary, bool) {
	var binary string

	if allowEmpty {
		binary = prompt("Binary name (result): ")
		if binary == "" {
			return Binary{}, false
		}
	} else {
		binary = promptNonEmpty("Binary name (result): ")
	}

	binary = stripExeSuffix(binary)

	sourceNames := make([]string, 0)

	fmt.Println("Enter all possible source names of the binary (leave empty if the name is identical/to quit):")

	for {
		source := stripExeSuffix(prompt("Source name: "))
		if source == "" {
			break
		}

		if source != binary {
			sourceNames = append(sourceNames, source)
		}
	}

	return Binary{Name: binary, SourceNames: sourceNames}, true
}

func promptForBinaries(fileNames []string) []Binary {
	result := make([]Binary, 0)

	if len(fileNames) != 0 {
		fmt.Println("Found the following potential binary files in the archive:")
		for i, file := range fileNames {
			fmt.Printf("%d: %q\n", i, file)
		}
	}

	fmt.Println("Please provide the binaries to extract from the asset. Pressing enter (empty binary name) will finish the process:")

	binary, _ := promptForBinary(false)
	result = append(result, binary)

	for {
		binary, ok := promptForBinary(true)

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

		fmt.Println("The provided pattern does not match exactly one asset. Please be more specific.")
		fmt.Println("The following asset names would be matched by this regex:")
		for _, name := range matches {
			fmt.Printf("  - %s\n", name)
		}
	}
}
