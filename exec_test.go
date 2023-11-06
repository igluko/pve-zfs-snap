package main

import (
	"reflect"
	"testing"
)

type MockExec struct {
	commandFunc func(name string, arg ...string) ([]byte, error)
}

func (m MockExec) Command(name string, arg ...string) ([]byte, error) {
	return m.commandFunc(name, arg...)
}

func TestQmList(t *testing.T) {
	// create a mock executor
	mockExecutor := &MockExec{
		commandFunc: func(name string, arg ...string) ([]byte, error) {
			expectedName := "qm"
			expectedArg := []string{"list"}
			if name != expectedName || !reflect.DeepEqual(arg, expectedArg) {
				t.Errorf("unexpected command: got %v %v, want %v %v", name, arg, expectedName, expectedArg)
			}
			return []byte("      VMID NAME                 STATUS     MEM(MB)    BOOTDISK(GB) PID\n" +
				"       901 MikroTik-bot-test    running    1024               0.12 3678283\n" +
				"       951 repl-test2           running    2048              32.00 3679206\n" +
				"       953 test-savelov-emptyvm stopped    1024               0.12 0\n"), nil
		},
	}

	// call the function being tested
	qms, err := QmList(mockExecutor)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check the result
	expectedQms := []qm{
		{vmid: "901", name: "MikroTik-bot-test", status: "running", mem: "1024", disk: "0.12", pid: "3678283"},
		{vmid: "951", name: "repl-test2", status: "running", mem: "2048", disk: "32.00", pid: "3679206"},
		{vmid: "953", name: "test-savelov-emptyvm", status: "stopped", mem: "1024", disk: "0.12", pid: "0"},
	}
	if !reflect.DeepEqual(qms, expectedQms) {
		t.Errorf("unexpected qms: got %v, want %v", qms, expectedQms)
	}
}
func TestPctList(t *testing.T) {
	// create a mock executor
	mockExecutor := &MockExec{
		commandFunc: func(name string, arg ...string) ([]byte, error) {
			expectedName := "pct"
			expectedArg := []string{"list"}
			if name != expectedName || !reflect.DeepEqual(arg, expectedArg) {
				t.Errorf("unexpected command: got %v %v, want %v %v", name, arg, expectedName, expectedArg)
			}
			return []byte("VMID       Status     Lock         Name\n" +
				"952        running                 test-vm-1\n" +
				"953        stopped                 test-vm-2\n"), nil
		},
	}

	// call the function being tested
	pcts, err := PctList(mockExecutor)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check the result
	expectedPcts := []pct{
		{vmid: "952", status: "running", lock: "", name: "test-vm-1"},
		{vmid: "953", status: "stopped", lock: "", name: "test-vm-2"},
	}
	if !reflect.DeepEqual(pcts, expectedPcts) {
		t.Errorf("unexpected pcts: got %v, want %v", pcts, expectedPcts)
	}
}
func TestZpoolList(t *testing.T) {
	// create a mock executor
	mockExecutor := &MockExec{
		commandFunc: func(name string, arg ...string) ([]byte, error) {
			expectedName := "zpool"
			expectedArg := []string{"list", "-H", "-o", "name"}
			if name != expectedName || !reflect.DeepEqual(arg, expectedArg) {
				t.Errorf("unexpected command: got %v %v, want %v %v", name, arg, expectedName, expectedArg)
			}
			return []byte("tank\n" +
				"backup\n"), nil
		},
	}

	// call the function being tested
	zpools, err := ZpoolList(mockExecutor)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check the result
	expectedZpools := []string{"tank", "backup"}
	if !reflect.DeepEqual(zpools, expectedZpools) {
		t.Errorf("unexpected zpools: got %v, want %v", zpools, expectedZpools)
	}
}
func TestZfsList(t *testing.T) {
	pool := "rpool"
	// create a mock executor
	mockExecutor := &MockExec{
		commandFunc: func(name string, arg ...string) ([]byte, error) {
			expectedName := "zfs"
			expectedArg := []string{"list", "-o", "name,label:nosnap,label:running", "-r", pool}
			if name != expectedName || !reflect.DeepEqual(arg, expectedArg) {
				t.Errorf("unexpected command: got %v %v, want %v %v", name, arg, expectedName, expectedArg)
			}
			return []byte("NAME                          LABEL:NOSNAP  LABEL:RUNNING\n" +
				"rpool                         -             -\n" +
				"rpool/ROOT                    nosnap        -\n" +
				"rpool/data/subvol-952-disk-0  -             running"), nil
		},
	}

	// call the function being tested
	zfsList, err := ZFSlist(mockExecutor, pool)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check the result
	expectesZfsList := []zfs{
		{name: "rpool", nosnap: false, running: false},
		{name: "rpool/ROOT", nosnap: true, running: false},
		{name: "rpool/data/subvol-952-disk-0", nosnap: false, running: true},
	}
	if !reflect.DeepEqual(zfsList, expectesZfsList) {
		t.Errorf("unexpected datasets: got %v, want %v", zfsList, expectesZfsList)
	}
}
func TestZfsListSnapshots(t *testing.T) {
	zfs := "pool1/dataset1"
	// create a mock executor
	mockExecutor := &MockExec{
		commandFunc: func(name string, arg ...string) ([]byte, error) {
			expectedName := "zfs"
			expectedArg := []string{"list", "-p", "-o", "name,creation", "-t", "snapshot", zfs}
			if name != expectedName || !reflect.DeepEqual(arg, expectedArg) {
				t.Errorf("unexpected command: got %v %v, want %v %v", name, arg, expectedName, expectedArg)
			}
			return []byte("NAME                                                          CREATION\n" +
				"rpool/data/vm-950-disk-0@autosnap_2023-10-19_10:00:00_hourly  1697709600\n" +
				"rpool/data/vm-950-disk-0@autosnap_2023-10-19_11:00:03_hourly  1697713203\n"), nil
		},
	}

	// call the function being tested
	snapshots, err := ZfsListSnapshots(mockExecutor, zfs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check the result
	expectedSnapshots := []snapshot{
		{name: "rpool/data/vm-950-disk-0@autosnap_2023-10-19_10:00:00_hourly", creation: 1697709600},
		{name: "rpool/data/vm-950-disk-0@autosnap_2023-10-19_11:00:03_hourly", creation: 1697713203},
	}
	if !reflect.DeepEqual(snapshots, expectedSnapshots) {
		t.Errorf("unexpected snapshots: got %v, want %v", snapshots, expectedSnapshots)
	}
}
