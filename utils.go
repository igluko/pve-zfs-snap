package main

import (
	"strings"
)

func SplitTable(bytes []byte) [][]string {
	if len(bytes) == 0 {
		return [][]string{}
	}
	trimmed := strings.Trim(string(bytes), "\n")
	lines := strings.Split(trimmed, "\n")
	headerLine := strings.TrimSpace(lines[0])
	headers := strings.Fields(headerLine)
	numColumns := len(headers)
	var table [][]string
	table = append(table, headers)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < numColumns {
			// Skip lines that don't have enough fields
			continue
		}
		if len(fields) == numColumns {
			// Line has the correct number of fields
			table = append(table, fields)
			continue
		}
		// More fields than headers; assume extra fields are part of the NAME column
		var row []string
		row = append(row, fields[0]) // VMID
		nameFieldEnd := len(fields) - (numColumns - 1)
		name := strings.Join(fields[1:nameFieldEnd], " ")
		row = append(row, name)
		for i := nameFieldEnd; i < len(fields); i++ {
			row = append(row, fields[i])
		}
		table = append(table, row)
	}
	return table
}
