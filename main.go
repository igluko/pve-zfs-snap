///usr/bin/env go run "$0" "$@"; exit

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	frequently = "frequently"
	hourly     = "hourly"
	daily      = "daily"
	monthly    = "monthly"
	yearly     = "yearly"
	stopped    = "stopped" // for stopped VMs
)

type policy struct {
	count    int
	interval int64
}

type environment struct {
	hostname string
	path     string
	time     struct {
		unix  int64
		human string
	}
	policy map[string]policy
}

func help() {
	fmt.Println("All parameters must have the format '<one_letter><int>'")
	fmt.Println("  Example usage: ./pve-zfs-snap f100000")
	fmt.Println("Possible keys and their descriptions:")
	fmt.Println("  f<int> - number of frequently snapshots")
	fmt.Println("  h<int> - number of hourly snapshots")
	fmt.Println("  d<int> - number of daily snapshots")
	fmt.Println("  m<int> - number of monthly snapshots")
	fmt.Println("  y<int> - number of yearly snapshots")
}

// Regular expression to match snapshot types
var snapshotTypeRE = regexp.MustCompile(`@autosnap_[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}:[0-9]{2}:[0-9]{2}_([a-z]+)`)

func init() {
	os.Setenv("PATH", os.Getenv("PATH")+":/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin")
}

func checkCallLuaCode(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("minimum number of parameters is 1")
	}

	switch args[1] {
	case "lua_hello":
		fmt.Println(lua_hello())
		os.Exit(0)
	case "lua_snapshot":
		fmt.Println(lua_snapshot())
		os.Exit(0)
	case "lua_destroy":
		fmt.Println(lua_destroy())
		os.Exit(0)
	case "lua_set_running":
		fmt.Println(lua_set_running())
		os.Exit(0)
	}
	return nil
}

func getEnvironment(args []string) (environment, error) {
	if len(args) < 2 {
		return environment{}, fmt.Errorf("minimum number of parameters is 1")
	}

	var env = environment{
		policy: make(map[string]policy),
	}

	for _, arg := range args[1:] {
		i, err := strconv.Atoi(arg[1:])
		if err != nil {
			return environment{}, fmt.Errorf("parameter '%s' is not a number", arg)
		}
		switch arg[0] {
		case 'f':
			env.policy[frequently] = policy{count: i, interval: 0}
		case 'h':
			env.policy[hourly] = policy{count: i, interval: 3600}
		case 'd':
			env.policy[daily] = policy{count: i, interval: 3600 * 24}
		case 'm':
			env.policy[monthly] = policy{count: i, interval: 3600 * 24 * 30}
		case 'y':
			env.policy[yearly] = policy{count: i, interval: 3600 * 24 * 365}
		default:
			return environment{}, fmt.Errorf("unknown parameter '%s'", arg)
		}
	}
	env.path = args[0]
	env.time.human = time.Now().Format("2006-01-02_15:04:05")
	env.time.unix = time.Now().Unix()
	env.hostname, _ = os.Hostname()
	return env, nil
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		help()
		os.Exit(1)
	}
}

// VM represents a virtual machine or container
type VM struct {
	CPU       float64 `json:"cpu"`
	Disk      int64   `json:"disk"`
	DiskRead  int64   `json:"diskread"`
	DiskWrite int64   `json:"diskwrite"`
	ID        string  `json:"id"`
	MaxCPU    int     `json:"maxcpu"`
	MaxDisk   int64   `json:"maxdisk"`
	MaxMem    int64   `json:"maxmem"`
	Mem       int64   `json:"mem"`
	Name      string  `json:"name"`
	NetIn     int64   `json:"netin"`
	NetOut    int64   `json:"netout"`
	Node      string  `json:"node"`
	Status    string  `json:"status"`
	Template  int     `json:"template"`
	Type      string  `json:"type"`
	Uptime    int64   `json:"uptime"`
	VMID      int     `json:"vmid"`
}

// GetVMs retrieves the list of VMs for the current node
func GetVMs(e Exec, node string) ([]VM, error) {
	output, err := e.Command("pvesh", "get", "/cluster/resources", "--type", "vm", "--output-format", "json")
	if err != nil {
		return nil, err
	}
	var allVMs []VM
	if err := json.Unmarshal(output, &allVMs); err != nil {
		return nil, err
	}
	var nodeVMs []VM
	for _, vm := range allVMs {
		if vm.Node == node {
			nodeVMs = append(nodeVMs, vm)
		}
	}
	return nodeVMs, nil
}

// GetAllVMIDs extracts VMIDs from the list of VMs
func GetAllVMIDs(vms []VM) []int {
	var vmIDs []int
	for _, vm := range vms {
		vmIDs = append(vmIDs, vm.VMID)
	}
	return vmIDs
}

// GetRunningVMIDs extracts VMIDs of running VMs
func GetRunningVMIDs(vms []VM) []int {
	var vmIDs []int
	for _, vm := range vms {
		if vm.Status == "running" {
			vmIDs = append(vmIDs, vm.VMID)
		}
	}
	return vmIDs
}

// Filter ZFS datasets related to a list of VMIDs
func filterZfsInVms(zfsList []zfs, vmIDs []int) []zfs {
	var filteredZfs []zfs
	vmidStrings := make([]string, len(vmIDs))
	for i, vmid := range vmIDs {
		vmidStrings[i] = strconv.Itoa(vmid)
	}
	vmidPattern := strings.Join(vmidStrings, "|")
	re := regexp.MustCompile(fmt.Sprintf("vm-(%s)-disk-|subvol-(%s)-disk-", vmidPattern, vmidPattern))
	for _, zfs := range zfsList {
		if re.MatchString(zfs.name) {
			filteredZfs = append(filteredZfs, zfs)
		}
	}
	return filteredZfs
}

// Check if a zfs is in a list of zfs
func containsZFS(zfsList []zfs, target zfs) bool {
	for _, zfs := range zfsList {
		if zfs.name == target.name {
			return true
		}
	}
	return false
}

// Get datasets pending stop action
func getPendingStopZFS(allZFS []zfs, runningZFS []zfs, hostname string) []zfs {
	var pendingStoppedZFS []zfs
	for _, zfs := range allZFS {
		if containsZFS(runningZFS, zfs) {
			continue
		}
		if zfs.running == hostname || zfs.running == "-" {
			pendingStoppedZFS = append(pendingStoppedZFS, zfs)
		}
	}
	return pendingStoppedZFS
}

// Get datasets pending start action
func getPendingStartZFS(allZFS []zfs, runningZFS []zfs, hostname string) []zfs {
	var pendingStartZFS []zfs
	for _, zfs := range allZFS {
		if !containsZFS(runningZFS, zfs) {
			continue
		}
		if zfs.running != hostname {
			pendingStartZFS = append(pendingStartZFS, zfs)
		}
	}
	return pendingStartZFS
}

func processPendingsZFS(pending *Pending, pendingStopZFS []zfs, pendingStartZFS []zfs, env environment) {
	for _, zfs := range pendingStopZFS {
		pending.Snapshots = append(pending.Snapshots, fmt.Sprintf("%s@autosnap_%s_stopped", zfs.name, env.time.human))
		pending.SetStopped = append(pending.SetStopped, zfs.name)
	}
	for _, zfs := range pendingStartZFS {
		pending.SetRunning = append(pending.SetRunning, zfs.name)
	}
}

// Split snapshots into groups by type
func splitSnapshots(snapshots []snapshot) map[string][]snapshot {
	group := make(map[string][]snapshot)
	for _, snapshot := range snapshots {
		submatch := snapshotTypeRE.FindStringSubmatch(snapshot.name)
		if len(submatch) < 2 {
			continue
		}
		snapshotType := submatch[1]
		group[snapshotType] = append(group[snapshotType], snapshot)
	}
	return group
}

// Filter out nosnap datasets
func filterNoSnap(zfsList []zfs) []zfs {
	var filteredZFS []zfs
	for _, zfs := range zfsList {
		if zfs.nosnap {
			continue
		}
		filteredZFS = append(filteredZFS, zfs)
	}
	return filteredZFS
}

func snapshotsToNames(snapshots []snapshot) []string {
	snapshotNames := make([]string, len(snapshots))
	for i, snapshot := range snapshots {
		snapshotNames[i] = snapshot.name
	}
	return snapshotNames
}

// Process snapshots based on policy
func processSnapshots(
	pending *Pending,
	snapshots []snapshot,
	zfsName string,
	snapshotType string,
	policy policy,
	timeNowUnix int64,
	timeNowHuman string,
) {
	count := len(snapshots)
	maxCount := policy.count
	var timeLast int64
	if count > 0 {
		timeLast = snapshots[count-1].creation
	}

	if maxCount == 0 {
		pending.Destroys = append(
			pending.Destroys,
			snapshotsToNames(snapshots)...)
		return
	}

	if timeLast+policy.interval < timeNowUnix+60 {
		pending.Snapshots = append(
			pending.Snapshots,
			fmt.Sprintf("%s@autosnap_%s_%s", zfsName, timeNowHuman, snapshotType),
		)
		count++
	}
	if count > policy.count {
		pending.Destroys = append(
			pending.Destroys,
			snapshotsToNames(snapshots[:count-maxCount])...)
	}
}

func main() {
	err := checkCallLuaCode(os.Args)
	checkErr(err)

	env, err := getEnvironment(os.Args)
	checkErr(err)

	if isTerminal() {
		fmt.Println("Running in terminal mode")
		updateCron()
	}

	executor := OSExec{}

	poolList, err := ZpoolList(executor)
	checkErr(err)

	vms, err := GetVMs(executor, env.hostname)
	checkErr(err)

	allVMIDs := GetAllVMIDs(vms)
	runningVMIDs := GetRunningVMIDs(vms)

	for _, pool := range poolList {
		pending := Pending{Pool: pool, Hosname: env.hostname}

		allZFS, err := ZFSlist(executor, pool)
		checkErr(err)

		// All datasets related to VMs
		allZFS = filterZfsInVms(allZFS, allVMIDs)

		// Datasets related to running VMs
		runningZFS := filterZfsInVms(allZFS, runningVMIDs)

		pendingStopZFS := getPendingStopZFS(allZFS, runningZFS, env.hostname)
		pendingStartZFS := getPendingStartZFS(allZFS, runningZFS, env.hostname)

		processPendingsZFS(&pending, pendingStopZFS, pendingStartZFS, env)

		// Filter nosnap datasets
		runningZFS = filterNoSnap(runningZFS)

		for _, zfs := range runningZFS {
			snapshots, err := ZfsListSnapshots(executor, zfs.name)
			checkErr(err)

			// Snapshots grouped by types and filtered by pattern
			groupedSnapshots := splitSnapshots(snapshots)

			processSnapshots(&pending, groupedSnapshots[yearly], zfs.name, yearly, env.policy[yearly], env.time.unix, env.time.human)
			processSnapshots(&pending, groupedSnapshots[monthly], zfs.name, monthly, env.policy[monthly], env.time.unix, env.time.human)
			processSnapshots(&pending, groupedSnapshots[daily], zfs.name, daily, env.policy[daily], env.time.unix, env.time.human)
			processSnapshots(&pending, groupedSnapshots[hourly], zfs.name, hourly, env.policy[hourly], env.time.unix, env.time.human)
			processSnapshots(&pending, groupedSnapshots[frequently], zfs.name, frequently, env.policy[frequently], env.time.unix, env.time.human)
			processSnapshots(&pending, groupedSnapshots[stopped], zfs.name, stopped, env.policy[stopped], env.time.unix, env.time.human)
		}
		err = pending.Run()
		checkErr(err)
	}
}
