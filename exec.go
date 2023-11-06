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

type qm struct {
	vmid   string
	name   string
	status string
	mem    string
	disk   string
	pid    string
}

type pct struct {
	vmid   string
	status string
	lock   string
	name   string
}

type zfs struct {
	name    string
	nosnap  bool
	running bool
}

type snapshot struct {
	name     string
	creation int64
}

// qm list
// Empty output if no VMs
/*
   VMID NAME                 STATUS     MEM(MB)    BOOTDISK(GB) PID
    901 MikroTik-bot-test    running    1024               0.12 3678283
    951 repl-test2           running    2048              32.00 3679206
    953 test-savelov-emptyvm stopped    1024               0.12 0
*/
func QmList(e Exec) ([]qm, error) {
	bytes, err := e.Command("qm", "list")
	if err != nil {
		return nil, err
	}
	table := SplitTable(bytes)
	if len(table) == 0 {
		return []qm{}, nil
	}
	qms := make([]qm, len(table)-1)
	for i, row := range table[1:] {
		qms[i] = qm{
			vmid:   row[0],
			name:   row[1],
			status: row[2],
			mem:    row[3],
			disk:   row[4],
			pid:    row[5],
		}
	}
	return qms, nil
}

// pct list
// Empty output if no VMs
/*
	VMID       Status     Lock         Name
	952        running                 test-savelov-empty
*/
func PctList(e Exec) ([]pct, error) {
	bytes, err := e.Command("pct", "list")
	if err != nil {
		return nil, err
	}
	table := SplitTable(bytes)
	if len(table) == 0 {
		return []pct{}, nil
	}
	pcts := make([]pct, len(table)-1)
	for i, line := range table[1:] {
		pcts[i] = pct{
			vmid:   line[0],
			status: line[1],
			lock:   line[2],
			name:   line[3],
		}
	}
	return pcts, nil
}

// zpool list -H -o name
/*
	rpool
	hdd
*/
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

// zfs list -o name,label:nosnap,label:running $pool
/*
	rpool   -       -
	rpool/ROOT      -       -
	rpool/ROOT/pve-1        -       -
	rpool/data      -       -
	rpool/data/subvol-952-disk-0    -       running
	rpool/data/vm-106-disk-0        -       -
*/
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
		zfsList[i] = zfs{
			name:    line[0],
			nosnap:  line[1] == "nosnap",
			running: line[2] == "running",
		}
	}
	return zfsList, nil
}

// zfs list -p -o name,creation -t snapshot rpool/data/vm-950-disk-0
/*
NAME                                                          CREATION
rpool/data/vm-950-disk-0@autosnap_2023-10-19_10:00:00_hourly  1697709600
rpool/data/vm-950-disk-0@autosnap_2023-10-19_11:00:03_hourly  1697713203
rpool/data/vm-950-disk-0@autosnap_2023-10-19_12:00:00_hourly  1697716800
rpool/data/vm-950-disk-0@autosnap_2023-10-19_13:00:01_hourly  1697720401
*/
func ZfsListSnapshots(e Exec, zfs string) ([]snapshot, error) {
	bytes, err := e.Command("zfs", "list", "-p", "-o", "name,creation", "-t", "snapshot", zfs)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	table := SplitTable(bytes)
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
