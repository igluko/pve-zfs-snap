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
	vmIDs := []string{"100", "101", "102"}
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
	qmList := []qm{
		{vmid: "100", status: "running"},
		{vmid: "101", status: "stopped"},
		{vmid: "102", status: "running"},
	}
	pctList := []pct{
		{vmid: "200", status: "running"},
		{vmid: "201", status: "running"},
		{vmid: "202", status: "stopped"},
	}
	expected := []string{"100", "102", "200", "201"}
	got := getRunningVMIDs(qmList, pctList)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("getRunningVMIDs() = %v, want %v", got, expected)
	}
}
func TestGetPendingStoppedZFS(t *testing.T) {
	allZFS := []zfs{
		{name: "zfs1", running: true},
		{name: "zfs2", running: true},
		{name: "zfs3", running: true},
		{name: "zfs4", running: false},
	}
	runningZFS := []zfs{
		{name: "zfs1", running: true},
		{name: "zfs3", running: true},
	}
	expected := []zfs{
		{name: "zfs2", running: true},
	}
	got := getPendingStopZFS(allZFS, runningZFS)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("getPendingStoppedZFS() = %v, want %v", got, expected)
	}
}
func TestGetPendingStartZFS(t *testing.T) {
	allZFS := []zfs{
		{name: "zfs1", running: true},
		{name: "zfs2", running: true},
		{name: "zfs3", running: false},
		{name: "zfs4", running: false},
	}
	runningZFS := []zfs{
		{name: "zfs1", running: true},
		{name: "zfs3", running: false},
	}
	expected := []zfs{
		{name: "zfs3", running: false},
	}
	got := getPendingStartZFS(allZFS, runningZFS)
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

func TestProcessSnapshots_yearly(t *testing.T) {
	var pending = &Pending{}
	pending.Pool = "rpool"
	pending.Snapshots = []string{
		"datasetA@autosnap_A",
	}
	pending.Destroys = []string{
		"datasetB@autosnap_B",
	}
	pending.SetRunning = []string{
		"datasetC",
	}
	pending.UnsetRunning = []string{
		"datasetD",
	}

	env := environment{
		time: struct {
			unix  int64
			human string
		}{
			unix:  1674553502,
			human: "2023-01-24_09:45:02",
		},
		policy: map[string]policy{
			"yearly":     {count: 1, interval: 3600 * 24 * 365},
			"monthly":    {count: 4, interval: 3600 * 24 * 30},
			"daily":      {count: 3, interval: 3600 * 24},
			"hourly":     {count: 3, interval: 3600},
			"frequently": {count: 5, interval: 0},
		},
	}

	snapshots := []snapshot{
		{"vm-100-disk-1@autosnap_2020-10-21_04:00:02_yearly", 1603252802},
		{"vm-100-disk-1@autosnap_2021-10-21_04:00:02_yearly", 1634788802},
	}

	zfsName := "vm-100-disk-1"
	snapshotType := "yearly"

	expectedPending := &Pending{
		Pool: "rpool",
		Snapshots: []string{
			"datasetA@autosnap_A",
			"vm-100-disk-1@autosnap_2023-01-24_09:45:02_yearly",
		},
		Destroys: []string{
			"datasetB@autosnap_B",
			"vm-100-disk-1@autosnap_2020-10-21_04:00:02_yearly",
			"vm-100-disk-1@autosnap_2021-10-21_04:00:02_yearly",
		},
		SetRunning: []string{
			"datasetC",
		},
		UnsetRunning: []string{
			"datasetD",
		},
	}

	processSnapshots(pending, snapshots, zfsName, snapshotType, env.policy[snapshotType],
		env.time.unix, env.time.human)

	if !reflect.DeepEqual(pending, expectedPending) {
		t.Errorf("processSnapshots() pendingSnapshots = %v, want %v", pending, expectedPending)
	}
}

func TestProcessSnapshots_frequently(t *testing.T) {
	var pending = &Pending{}
	pending.Pool = "rpool"
	pending.Snapshots = []string{
		"datasetA@autosnap_A",
	}
	pending.Destroys = []string{
		"datasetB@autosnap_B",
	}
	pending.SetRunning = []string{
		"datasetC",
	}
	pending.UnsetRunning = []string{
		"datasetD",
	}

	env := environment{
		time: struct {
			unix  int64
			human string
		}{
			unix:  1674553502,
			human: "2023-01-24_09:45:02",
		},
		policy: map[string]policy{
			"yearly":     {count: 1, interval: 3600 * 24 * 365},
			"monthly":    {count: 4, interval: 3600 * 24 * 30},
			"daily":      {count: 3, interval: 3600 * 24},
			"hourly":     {count: 3, interval: 3600},
			"frequently": {count: 5, interval: 0},
		},
	}

	snapshots := []snapshot{
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

	zfsName := "vm-100-disk-1"
	snapshotType := "frequently"

	expectedPending := &Pending{
		Pool: "rpool",
		Snapshots: []string{
			"datasetA@autosnap_A",
			"vm-100-disk-1@autosnap_2023-01-24_09:45:02_frequently",
		},
		Destroys: []string{
			"datasetB@autosnap_B",
			"vm-100-disk-1@autosnap_2023-01-24_07:15:02_frequently",
			"vm-100-disk-1@autosnap_2023-01-24_07:30:02_frequently",
			"vm-100-disk-1@autosnap_2023-01-24_07:45:02_frequently",
			"vm-100-disk-1@autosnap_2023-01-24_08:00:02_frequently",
			"vm-100-disk-1@autosnap_2023-01-24_08:15:02_frequently",
			"vm-100-disk-1@autosnap_2023-01-24_08:30:02_frequently",
		},
		SetRunning: []string{
			"datasetC",
		},
		UnsetRunning: []string{
			"datasetD",
		},
	}

	processSnapshots(pending, snapshots, zfsName, snapshotType, env.policy[snapshotType],
		env.time.unix, env.time.human)

	if !reflect.DeepEqual(pending, expectedPending) {
		t.Errorf("processSnapshots() pendingSnapshots = %v, want %v", pending, expectedPending)
	}
}

func TestProcessSnapshots_monthly(t *testing.T) {
	var pending = &Pending{}
	pending.Pool = "rpool"
	pending.Snapshots = []string{
		"datasetA@autosnap_A",
	}
	pending.Destroys = []string{
		"datasetB@autosnap_B",
	}
	pending.SetRunning = []string{
		"datasetC",
	}
	pending.UnsetRunning = []string{
		"datasetD",
	}

	env := environment{
		time: struct {
			unix  int64
			human string
		}{
			unix:  1674553502,
			human: "2023-01-24_09:45:02",
		},
		policy: map[string]policy{
			"yearly":     {count: 1, interval: 3600 * 24 * 365},
			"monthly":    {count: 4, interval: 3600 * 24 * 30},
			"daily":      {count: 3, interval: 3600 * 24},
			"hourly":     {count: 3, interval: 3600},
			"frequently": {count: 5, interval: 0},
		},
	}

	snapshots := []snapshot{
		{"vm-100-disk-1@autosnap_2022-10-21_04:00:02_monthly", 1666324802},
		{"vm-100-disk-1@autosnap_2022-11-21_04:00:02_monthly", 1669003202},
		{"vm-100-disk-1@autosnap_2022-12-21_04:00:02_monthly", 1671595202},
		{"vm-100-disk-1@autosnap_2023-01-21_04:00:02_monthly", 1674273602},
	}

	zfsName := "vm-100-disk-1"
	snapshotType := "monthly"

	expectedPending := &Pending{
		Pool: "rpool",
		Snapshots: []string{
			"datasetA@autosnap_A",
		},
		Destroys: []string{
			"datasetB@autosnap_B",
		},
		SetRunning: []string{
			"datasetC",
		},
		UnsetRunning: []string{
			"datasetD",
		},
	}

	processSnapshots(pending, snapshots, zfsName, snapshotType, env.policy[snapshotType],
		env.time.unix, env.time.human)

	if !reflect.DeepEqual(pending, expectedPending) {
		t.Errorf("processSnapshots() pendingSnapshots = %+v, want %+v", pending, expectedPending)
	}
}

// func TestMain(t *testing.T) {
// 	// arrange test zfs
// 	testZfsRoot := "rpool/pve-zfs-snap"
// 	//  zfs destroy -r rpool/pve-zfs-snap
// 	exec.Command("zfs", "destroy", "-r", testZfsRoot).Run()
// 	//  zfs create rpool/pve-zfs-snap
// 	err := exec.Command("zfs", "create", testZfsRoot).Run()
// 	if err != nil {
// 		t.Errorf("zfs create %s: %s", testZfsRoot, err)
// 	}

// 	// arrang os.Args
// 	fmt.Println(os.Args)
// 	os.Args = append(os.Args[:1], "lua_snapshot")

// 	// run main
// 	main()

// }
