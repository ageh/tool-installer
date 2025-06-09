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

func promptRegex(text string) string {
	fmt.Print(text)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return ""
		}

		result := strings.TrimSpace(input)

		_, err = regexp.Compile(result)
		if err == nil {
			return result
		}

		fmt.Print("Input must be a valid regular expression. Please try again: ")
	}
}

func promptForBinary() (Binary, bool) {
	binary := prompt("Binary name: ")
	rename := prompt("Rename binary to (leave empty if no rename): ")

	if binary == "" {
		return Binary{}, false
	}

	return Binary{Name: binary, RenameTo: rename}, true
}
