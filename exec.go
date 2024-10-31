package main

import (
	"os/exec"
	"strconv"
	"strings"
)

type Exec interface {
	Command(name string, arg ...string) ([]byte, error)
}

type OSExec struct{}

func (e OSExec) Command(cmd string, arg ...string) ([]byte, error) {
	return exec.Command(cmd, arg...).Output()
}

type zfs struct {
	name    string
	nosnap  bool
	running string
}

type snapshot struct {
	name     string
	creation int64
}

// ZpoolList retrieves the list of ZFS pools
func ZpoolList(e Exec) ([]string, error) {
	bytes, err := e.Command("zpool", "list", "-H", "-o", "name")
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	trimmed := strings.TrimSpace((string(bytes)))
	lines := strings.Split(trimmed, "\n")
	return lines, nil
}

// ZFSlist retrieves ZFS datasets with specific properties
func ZFSlist(e Exec, pool string) ([]zfs, error) {
	bytes, err := e.Command("zfs", "list", "-o", "name,label:nosnap,label:running", "-r", pool)
	if err != nil {
		return nil, err
	}
	table := SplitTable(bytes)
	if len(table) == 0 {
		return []zfs{}, nil
	}
	zfsList := make([]zfs, len(table)-1)
	for i, line := range table[1:] {
		nosnap := false
		if len(line) > 1 && line[1] == "nosnap" {
			nosnap = true
		}
		running := "-"
		if len(line) > 2 {
			running = line[2]
		}
		zfsList[i] = zfs{
			name:    line[0],
			nosnap:  nosnap,
			running: running,
		}
	}
	return zfsList, nil
}

// ZfsListSnapshots retrieves snapshots of a ZFS dataset
func ZfsListSnapshots(e Exec, zfs string) ([]snapshot, error) {
	bytes, err := e.Command("zfs", "list", "-p", "-o", "name,creation", "-t", "snapshot", zfs)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	table := SplitTable(bytes)
	if len(table) <= 1 {
		return nil, nil
	}
	snapshots := make([]snapshot, len(table)-1)
	for i, line := range table[1:] {
		epoc, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			return nil, err
		}
		snapshots[i] = snapshot{
			name:     line[0],
			creation: epoc,
		}
	}
	return snapshots, nil
}
