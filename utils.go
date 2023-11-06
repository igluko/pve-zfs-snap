package main

import "strings"

func getPositions(header string) []int {
	var starPositions []int
	var firstColumnEndPosition int
	for i := 1; i < len(header); i++ {
		if header[i] == ' ' && header[i-1] != ' ' {
			firstColumnEndPosition = i
			break
		}
	}
	for i := firstColumnEndPosition; i < len(header); i++ {
		if header[i] != ' ' && header[i-1] == ' ' {
			starPositions = append(starPositions, i)
		}
	}
	return starPositions
}

func splitByPositions(line string, positions []int) []string {
	var parts []string
	prevPos := 0
	for _, pos := range positions {
		parts = append(parts, strings.TrimSpace(line[prevPos:pos]))
		prevPos = pos
	}
	parts = append(parts, strings.TrimSpace(line[prevPos:]))
	return parts
}

func SplitTable(bytes []byte) [][]string {
	if len(bytes) == 0 {
		return [][]string{}
	}
	trimmed := strings.Trim((string(bytes)), "\n")
	lines := strings.Split(trimmed, "\n")
	header := lines[0]
	positions := getPositions(header)
	var table [][]string
	for _, line := range lines {
		table = append(table, splitByPositions(line, positions))
	}
	return table
}

func snapshotsToNames(snapshots []snapshot) []string {
	snapshotNames := make([]string, len(snapshots))
	for i, snapshot := range snapshots {
		snapshotNames[i] = snapshot.name
	}
	return snapshotNames
}
