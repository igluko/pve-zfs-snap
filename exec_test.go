package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type MockExec struct {
	Outputs map[string][]byte
	Errors  map[string]error
}

func (m *MockExec) Command(name string, arg ...string) ([]byte, error) {
	key := name + " " + strings.Join(arg, " ")
	if err, ok := m.Errors[key]; ok {
		return nil, err
	}
	if output, ok := m.Outputs[key]; ok {
		return output, nil
	}
	return nil, fmt.Errorf("command not found: %s", key)
}

func TestGetVMs(t *testing.T) {
	sampleOutput := `[
	   {
		  "cpu" : 0.00299916133057372,
		  "disk" : 0,
		  "diskread" : 472992073728,
		  "diskwrite" : 608784817664,
		  "id" : "qemu/100",
		  "maxcpu" : 4,
		  "maxdisk" : 53687091200,
		  "maxmem" : 5293211648,
		  "mem" : 2495614976,
		  "name" : "Terminal-Simbirsk",
		  "netin" : 10836356743,
		  "netout" : 525029614,
		  "node" : "AX101-Hels-03",
		  "status" : "running",
		  "template" : 0,
		  "type" : "qemu",
		  "uptime" : 20124577,
		  "vmid" : 100
	   },
	   {
		  "cpu" : 0,
		  "disk" : 0,
		  "diskread" : 0,
		  "diskwrite" : 0,
		  "id" : "lxc/952",
		  "maxcpu" : 1,
		  "maxdisk" : 8589934592,
		  "maxmem" : 536870912,
		  "mem" : 0,
		  "name" : "test-savelov-empty",
		  "netin" : 0,
		  "netout" : 0,
		  "node" : "AX101-Falk-01",
		  "status" : "stopped",
		  "template" : 0,
		  "type" : "lxc",
		  "uptime" : 0,
		  "vmid" : 952
	   }
	]`

	mockExec := &MockExec{
		Outputs: map[string][]byte{
			"pvesh get /cluster/resources --type vm --output-format json": []byte(sampleOutput),
		},
	}

	node := "AX101-Hels-03"

	vms, err := GetVMs(mockExec, node)
	if err != nil {
		t.Fatalf("GetVMs returned error: %v", err)
	}

	if len(vms) != 1 {
		t.Fatalf("Expected 1 VM, got %d", len(vms))
	}

	vm := vms[0]
	if vm.Node != node {
		t.Errorf("Expected node %s, got %s", node, vm.Node)
	}
	if vm.VMID != 100 {
		t.Errorf("Expected VMID 100, got %d", vm.VMID)
	}
	if vm.Name != "Terminal-Simbirsk" {
		t.Errorf("Expected Name 'Terminal-Simbirsk', got '%s'", vm.Name)
	}
}

func TestGetAllVMIDs(t *testing.T) {
	vms := []VM{
		{VMID: 100},
		{VMID: 952},
	}

	vmIDs := GetAllVMIDs(vms)
	expectedVMIDs := []int{100, 952}

	if !reflect.DeepEqual(vmIDs, expectedVMIDs) {
		t.Errorf("Expected VMIDs %v, got %v", expectedVMIDs, vmIDs)
	}
}

func TestZpoolList(t *testing.T) {
	mockExec := &MockExec{
		Outputs: map[string][]byte{
			"zpool list -H -o name": []byte("tank\nbackup\n"),
		},
	}

	zpools, err := ZpoolList(mockExec)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedZpools := []string{"tank", "backup"}
	if !reflect.DeepEqual(zpools, expectedZpools) {
		t.Errorf("unexpected zpools: got %v, want %v", zpools, expectedZpools)
	}
}

func TestZFSlist(t *testing.T) {
	pool := "rpool"
	mockExec := &MockExec{
		Outputs: map[string][]byte{
			"zfs list -o name,label:nosnap,label:running -r rpool": []byte(
				"NAME                          LABEL:NOSNAP  LABEL:RUNNING\n" +
					"rpool                         -             -\n" +
					"rpool/ROOT                    nosnap        stopped\n" +
					"rpool/data/subvol-952-disk-0  -             HOST-1\n"),
		},
	}

	zfsList, err := ZFSlist(mockExec, pool)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedZfsList := []zfs{
		{name: "rpool", nosnap: false, running: "-"},
		{name: "rpool/ROOT", nosnap: true, running: "stopped"},
		{name: "rpool/data/subvol-952-disk-0", nosnap: false, running: "HOST-1"},
	}

	if !reflect.DeepEqual(zfsList, expectedZfsList) {
		t.Errorf("unexpected datasets: got %v, want %v", zfsList, expectedZfsList)
	}
}

func TestZfsListSnapshots(t *testing.T) {
	zfs := "pool1/dataset1"
	mockExec := &MockExec{
		Outputs: map[string][]byte{
			"zfs list -p -o name,creation -t snapshot pool1/dataset1": []byte(
				"NAME                                                          CREATION\n" +
					"pool1/dataset1@autosnap_2023-10-19_10:00:00_hourly  1697709600\n" +
					"pool1/dataset1@autosnap_2023-10-19_11:00:03_hourly  1697713203\n"),
		},
	}

	snapshots, err := ZfsListSnapshots(mockExec, zfs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedSnapshots := []snapshot{
		{name: "pool1/dataset1@autosnap_2023-10-19_10:00:00_hourly", creation: 1697709600},
		{name: "pool1/dataset1@autosnap_2023-10-19_11:00:03_hourly", creation: 1697713203},
	}

	if !reflect.DeepEqual(snapshots, expectedSnapshots) {
		t.Errorf("unexpected snapshots: got %v, want %v", snapshots, expectedSnapshots)
	}
}
