package main

import (
	"fmt"
	"strings"
)

type TableBuilder struct {
	headers      []string
	rows         [][]string
	columnWidths []int
	maxWidths    []int
}

func newTableBuilder(headers []string) TableBuilder {
	return newTableBuilderWithLimits(headers, map[int]int{})
}

func newTableBuilderWithLimits(headers []string, maximumWidths map[int]int) TableBuilder {
	n := len(headers)

	result := TableBuilder{
		headers:      headers,
		rows:         make([][]string, 0),
		columnWidths: make([]int, n),
		maxWidths:    make([]int, n),
	}

	for i, maximum := range maximumWidths {
		if i >= 0 && i < n {
			result.maxWidths[i] = maximum
		}
	}

	result.updateColumnWidths(headers)

	return result
}

func (t *TableBuilder) addRow(row []string) error {
	n := len(t.headers)

	if len(row) != n {
		return fmt.Errorf("column count must match")
	}

	t.updateColumnWidths(row)

	t.rows = append(t.rows, row)

	return nil
}

func (t *TableBuilder) updateColumnWidths(row []string) {
	for i, entry := range row {
		width := len(entry)

		maxWidth := t.maxWidths[i]

		if maxWidth > 0 && width > maxWidth {
			width = maxWidth
		}

		if width > t.columnWidths[i] {
			t.columnWidths[i] = width
		}
	}
}

func (t *TableBuilder) separatorRow(left rune, middle rune, right rune) string {
	var columns []string
	for _, width := range t.columnWidths {
		columns = append(columns, strings.Repeat("─", width))
	}

	return string(left) + "─" + strings.Join(columns, "─"+string(middle)+"─") + "─" + string(right) + "\n"
}

func (t *TableBuilder) formatRow(row []string) string {
	var result []string

	for i, entry := range row {
		result = append(result, fmt.Sprintf("%-*s", t.columnWidths[i], truncateText(entry, t.columnWidths[i])))
	}

	return "│ " + strings.Join(result, " │ ") + " │\n"
}

func (t *TableBuilder) build() string {
	var builder strings.Builder

	top := t.separatorRow('┌', '┬', '┐')
	middle := t.separatorRow('├', '┼', '┤')
	bottom := t.separatorRow('└', '┴', '┘')

	builder.WriteString(top)
	builder.WriteString(t.formatRow(t.headers))
	builder.WriteString(middle)

	for _, row := range t.rows {
		builder.WriteString(t.formatRow(row))
	}

	builder.WriteString(bottom)

	return builder.String()
}

func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	if maxLength <= 3 {
		return strings.Repeat(".", maxLength)
	}

	return text[:maxLength-3] + "..."
}
