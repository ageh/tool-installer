// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLength int
		expected  string
	}{
		{"No truncation needed", "hello", 10, "hello"},
		{"Exact length", "hello", 5, "hello"},
		{"Truncated with ellipsis", "hello world", 8, "hello..."},
		{"MaxLength of 3", "hello", 3, "..."},
		{"MaxLength of 2", "hello", 2, ".."},
		{"MaxLength of 1", "hello", 1, "."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := truncateText(test.text, test.maxLength)
			if got != test.expected {
				t.Errorf("wrong truncation result: got %q but expected %q", got, test.expected)
			}
		})
	}
}

func TestTableBuilderAddRow(t *testing.T) {
	t.Run("Correct column count", func(t *testing.T) {
		tb := newTableBuilder([]string{"A", "B", "C"})
		tb.addRow([]string{"x", "y", "z"})
	})

	t.Run("Incorrect column count", func(t *testing.T) {
		tb := newTableBuilder([]string{"A", "B", "C"})
		defer func() {
			if recover() == nil {
				t.Errorf("expected panic for wrong column count")
			}
		}()
		tb.addRow([]string{"x", "y"})
	})
}

func TestTableBuilder(t *testing.T) {
	tb := newTableBuilder([]string{"Name", "Version"})
	tb.addRow([]string{"ripgrep", "14.0.0"})
	tb.addRow([]string{"fd", "10.0.0"})

	result := tb.build()

	for _, expected := range []string{"Name", "Version", "ripgrep", "14.0.0", "fd", "10.0.0"} {
		if !strings.Contains(result, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}

	if !strings.HasPrefix(result, "┌") {
		t.Error("expected table to start with top-left corner character '┌'")
	}
	if !strings.HasSuffix(strings.TrimRight(result, "\n"), "┘") {
		t.Error("expected table to end with bottom-right corner character '┘'")
	}
}

func TestTableBuilderWithLimits(t *testing.T) {
	tooLongDescription := "A very long description that exceeds the limit"
	tb := newTableBuilderWithLimits([]string{"Name", "Description"}, map[int]int{1: 10})
	tb.addRow([]string{"ripgrep", tooLongDescription})

	result := tb.build()

	if strings.Contains(result, tooLongDescription) {
		t.Error("expected long description to be truncated")
	}
	if !strings.Contains(result, "...") {
		t.Error("expected truncated text to contain ellipsis")
	}
}
