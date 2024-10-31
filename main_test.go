package main

import (
	"reflect"
	"testing"
)

func TestFilterZfsInVms(t *testing.T) {
	zfsList := []zfs{
		{name: "vm-100-disk-1"},
		{name: "vm-101-disk-1"},
		{name: "subvol-102-disk-1"},
		{name: "vm-103-disk-1"},
	}
	vmIDs := []int{100, 101, 102}
	expected := []zfs{
		{name: "vm-100-disk-1"},
		{name: "vm-101-disk-1"},
		{name: "subvol-102-disk-1"},
	}
	got := filterZfsInVms(zfsList, vmIDs)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("filterZfsInVms() = %v, want %v", got, expected)
	}
}

func TestFilterNoSnap(t *testing.T) {
	zfsList := []zfs{
		{name: "zfs1", nosnap: false},
		{name: "zfs2", nosnap: true},
		{name: "zfs3", nosnap: false},
		{name: "zfs4", nosnap: true},
	}
	expected := []zfs{
		{name: "zfs1", nosnap: false},
		{name: "zfs3", nosnap: false},
	}
	got := filterNoSnap(zfsList)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("filterNoSnap() = %v, want %v", got, expected)
	}
}

func TestGetRunningVMIDs(t *testing.T) {
	vms := []VM{
		{VMID: 100, Status: "running"},
		{VMID: 101, Status: "stopped"},
		{VMID: 102, Status: "running"},
		{VMID: 200, Status: "running"},
		{VMID: 201, Status: "running"},
		{VMID: 202, Status: "stopped"},
	}
	expected := []int{100, 102, 200, 201}
	got := GetRunningVMIDs(vms)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetRunningVMIDs() = %v, want %v", got, expected)
	}
}

func TestGetPendingStoppedZFS(t *testing.T) {
	hostname := "HOST-1"
	allZFS := []zfs{
		{name: "zfs1", running: hostname},
		{name: "zfs2", running: "HOST-2"},
		{name: "zfs3", running: "-"},
		{name: "zfs4", running: "stopped"},
		{name: "zfs5", running: hostname},
	}
	runningZFS := []zfs{
		{name: "zfs5", running: hostname},
	}
	expected := []zfs{
		{name: "zfs1", running: hostname},
		{name: "zfs3", running: "-"},
	}
	got := getPendingStopZFS(allZFS, runningZFS, hostname)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("getPendingStoppedZFS() = %v, want %v", got, expected)
	}
}

func TestGetPendingStartZFS(t *testing.T) {
	hostname := "HOST-1"
	allZFS := []zfs{
		{name: "zfs1", running: hostname},
		{name: "zfs2", running: "HOST-2"},
		{name: "zfs3", running: "-"},
		{name: "zfs4", running: "stopped"},
		{name: "zfs5", running: hostname},
	}
	runningZFS := []zfs{
		{name: "zfs1", running: hostname},
		{name: "zfs2", running: "HOST-2"},
		{name: "zfs3", running: "-"},
		{name: "zfs4", running: "stopped"},
	}
	expected := []zfs{
		{name: "zfs2", running: "HOST-2"},
		{name: "zfs3", running: "-"},
		{name: "zfs4", running: "stopped"},
	}
	got := getPendingStartZFS(allZFS, runningZFS, hostname)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("getPendingStartZFS() = %v, want %v", got, expected)
	}
}

func TestSplitSnapshots(t *testing.T) {
	snapshots := []snapshot{
		{"vm-100-disk-1@autosnap_2020-10-21_04:00:02_yearly", 1603252802},
		{"vm-100-disk-1@autosnap_2021-10-21_04:00:02_yearly", 1634788802},
		{"vm-100-disk-1@autosnap_2022-10-21_04:00:02_yearly", 1666324802},
		{"vm-100-disk-1@autosnap_2022-10-21_04:00:02_monthly", 1666324802},
		{"vm-100-disk-1@autosnap_2022-11-21_04:00:02_monthly", 1669003202},
		{"vm-100-disk-1@autosnap_2022-12-21_04:00:02_monthly", 1671595202},
		{"vm-100-disk-1@autosnap_2023-01-21_04:00:02_monthly", 1674273602},
		{"vm-100-disk-1@autosnap_2023-01-22_04:00:02_daily", 1674360002},
		{"vm-100-disk-1@autosnap_2023-01-23_04:00:02_daily", 1674446402},
		{"vm-100-disk-1@autosnap_2023-01-24_04:00:02_daily", 1674532802},
		{"vm-100-disk-1@autosnap_2023-01-24_05:00:02_hourly", 1674536402},
		{"vm-100-disk-1@autosnap_2023-01-24_06:00:02_hourly", 1674540002},
		{"vm-100-disk-1@autosnap_2023-01-24_07:00:02_hourly", 1674543602},
		{"vm-100-disk-1@autosnap_2023-01-24_07:15:02_frequently", 1674544502},
		{"vm-100-disk-1@autosnap_2023-01-24_07:30:02_frequently", 1674545402},
		{"vm-100-disk-1@autosnap_2023-01-24_07:45:02_frequently", 1674546302},
		{"vm-100-disk-1@autosnap_2023-01-24_08:00:02_frequently", 1674547202},
		{"vm-100-disk-1@autosnap_2023-01-24_08:15:02_frequently", 1674548102},
		{"vm-100-disk-1@autosnap_2023-01-24_08:30:02_frequently", 1674549002},
		{"vm-100-disk-1@autosnap_2023-01-24_08:45:02_frequently", 1674549902},
		{"vm-100-disk-1@autosnap_2023-01-24_09:00:02_frequently", 1674550802},
		{"vm-100-disk-1@autosnap_2023-01-24_09:15:02_frequently", 1674551702},
		{"vm-100-disk-1@autosnap_2023-01-24_09:30:02_frequently", 1674552602},
	}
	expected := map[string][]snapshot{
		"yearly": {
			{"vm-100-disk-1@autosnap_2020-10-21_04:00:02_yearly", 1603252802},
			{"vm-100-disk-1@autosnap_2021-10-21_04:00:02_yearly", 1634788802},
			{"vm-100-disk-1@autosnap_2022-10-21_04:00:02_yearly", 1666324802},
		},
		"monthly": {
			{"vm-100-disk-1@autosnap_2022-10-21_04:00:02_monthly", 1666324802},
			{"vm-100-disk-1@autosnap_2022-11-21_04:00:02_monthly", 1669003202},
			{"vm-100-disk-1@autosnap_2022-12-21_04:00:02_monthly", 1671595202},
			{"vm-100-disk-1@autosnap_2023-01-21_04:00:02_monthly", 1674273602},
		},
		"daily": {
			{"vm-100-disk-1@autosnap_2023-01-22_04:00:02_daily", 1674360002},
			{"vm-100-disk-1@autosnap_2023-01-23_04:00:02_daily", 1674446402},
			{"vm-100-disk-1@autosnap_2023-01-24_04:00:02_daily", 1674532802},
		},
		"hourly": {
			{"vm-100-disk-1@autosnap_2023-01-24_05:00:02_hourly", 1674536402},
			{"vm-100-disk-1@autosnap_2023-01-24_06:00:02_hourly", 1674540002},
			{"vm-100-disk-1@autosnap_2023-01-24_07:00:02_hourly", 1674543602},
		},
		"frequently": {
			{"vm-100-disk-1@autosnap_2023-01-24_07:15:02_frequently", 1674544502},
			{"vm-100-disk-1@autosnap_2023-01-24_07:30:02_frequently", 1674545402},
			{"vm-100-disk-1@autosnap_2023-01-24_07:45:02_frequently", 1674546302},
			{"vm-100-disk-1@autosnap_2023-01-24_08:00:02_frequently", 1674547202},
			{"vm-100-disk-1@autosnap_2023-01-24_08:15:02_frequently", 1674548102},
			{"vm-100-disk-1@autosnap_2023-01-24_08:30:02_frequently", 1674549002},
			{"vm-100-disk-1@autosnap_2023-01-24_08:45:02_frequently", 1674549902},
			{"vm-100-disk-1@autosnap_2023-01-24_09:00:02_frequently", 1674550802},
			{"vm-100-disk-1@autosnap_2023-01-24_09:15:02_frequently", 1674551702},
			{"vm-100-disk-1@autosnap_2023-01-24_09:30:02_frequently", 1674552602},
		},
	}
	got := splitSnapshots(snapshots)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("splitSnapshots() = %v, want %v", got, expected)
	}
}
