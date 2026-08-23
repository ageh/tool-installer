// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
)

func TestIsArgumentCountIn(t *testing.T) {
	tests := []struct {
		name           string
		minimum        int
		maximum        int
		arguments      []string
		expectedResult bool
	}{
		{"No arguments", 0, 0, []string{}, true},
		{"Too many arguments", 0, 0, []string{"foo"}, false},
		{"Too few arguments", 2, 0, []string{"foo"}, false},
		{"In bounds", 0, 3, []string{"foo"}, true},
		{"Out of bounds", 0, 3, []string{"foo", "bar", "baz", "qux"}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := Arguments{
				commandArguments: test.arguments,
				command:          "test",
				requestTimeout:   10,
				showHelp:         false,
				showVersion:      false,
			}

			result := args.isArgumentCountIn(test.minimum, test.maximum)
			if result != test.expectedResult {
				t.Errorf("wrong result of isArgumentCountIn: got %v, expected %v", result, test.expectedResult)
			}
		})
	}
}
