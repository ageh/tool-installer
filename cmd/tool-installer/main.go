// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
)

var (
	version    = "dev"
	commitHash = "none"
	commitDate = "unknown"
	builtBy    = "unknown"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("Error: %v.\n\n", err)
		fmt.Println("Run `tooli help` for instructions.")
		os.Exit(1)
	}

	os.Exit(0)
}
