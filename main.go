///usr/bin/env go run "$0" "$@"; exit

package main

import (
	"fmt"
	"os"
	"regexp"
	"slices"
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
	stopped    = "stopped" // для остановленных VM
)

type policy struct {
	count    int
	interval int64
}

type environment struct {
	path string
	time struct {
		unix  int64
		human string
	}
	policy map[string]policy
}

func help() {
	fmt.Println("Все параметры параметры должены иметь формат '<one_letter><int>'")
	fmt.Println("  Пример использования: ./pve-zfs-snap f100000")
	fmt.Println("Возожные ключи и их описания:")
	fmt.Println("  f<int> - количество frequently снимков")
	fmt.Println("  h<int> - количество hourly снимков")
	fmt.Println("  d<int> - количество daily снимков")
	fmt.Println("  m<int> - количество monthly снимков")
	fmt.Println("  y<int> - количество yearly снимков")
}

// готовим регулярное выражение для поиска типа снимка
var snapshotTypeRE = regexp.MustCompile(`@autosnap_[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}:[0-9]{2}:[0-9]{2}_([a-z]+)`)

func init() {
	os.Setenv("PATH", os.Getenv("PATH")+":/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin")
}

func checkCallLuaCode(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("минимальное число параметров - 1")
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
	case "lua_unset_running":
		fmt.Println(lua_unset_running())
		os.Exit(0)
	}
	return nil
}

func getEnvironment(args []string) (environment, error) {
	if len(args) < 2 {
		return environment{}, fmt.Errorf("минимальное число параметров - 1")
	}

	var env = environment{
		policy: make(map[string]policy),
	}

	for _, arg := range args[1:] {
		i, err := strconv.Atoi(arg[1:])
		if err != nil {
			return environment{}, fmt.Errorf("параметр '%s' не является числом", arg)
		}
		switch arg[0] {
		case 'f':
			env.policy[frequently] = policy{count: i, interval: 0}
		case 'h':
			env.policy[hourly] = policy{count: i, interval: 3600}
		case 'd':
			env.policy[hourly] = policy{count: i, interval: 3600 * 24}
		case 'm':
			env.policy[hourly] = policy{count: i, interval: 3600 * 24 * 30}
		case 'y':
			env.policy[hourly] = policy{count: i, interval: 3600 * 24 * 365}
		default:
			return environment{}, fmt.Errorf("неизвестный параметр '%s'", arg)
		}
	}
	env.path = args[0]
	env.time.human = time.Now().Format("2006-01-02_15:04:05")
	env.time.unix = time.Now().Unix()
	return env, nil
}

// Получение Running VMIDs
func getRunningVMIDs(qmList []qm, pctList []pct) []string {
	var vmIDs []string
	for _, vm := range qmList {
		if vm.status == "running" {
			vmIDs = append(vmIDs, vm.vmid)
		}
	}
	for _, ct := range pctList {
		if ct.status == "running" {
			vmIDs = append(vmIDs, ct.vmid)
		}
	}
	return vmIDs
}

// Получение датасетов, связанных со списом VM
// Шаблон для поиска vm-<num>-disk-|subvol-<num>-disk-
func filterZfsInVms(zfsList []zfs, vmIDs []string) []zfs {
	var filteredZfs []zfs
	// Компилируем регулярное выражение
	// Шаблон для поиска vm-<num>-disk-|subvol-<num>-disk-
	vmidPatern := strings.Join(vmIDs, "|")
	// Используем vmidPatern в выражении
	re := regexp.MustCompile(fmt.Sprintf("vm-(%s)-disk-|subvol-(%s)-disk-", vmidPatern, vmidPatern))
	for _, zfs := range zfsList {
		if re.MatchString(zfs.name) {
			filteredZfs = append(filteredZfs, zfs)
		}
	}
	return filteredZfs
}

// Получение датасетов, после остановки VM
func getPendingStopZFS(allZFS []zfs, runningZFS []zfs) []zfs {
	var pendingStoppedZFS []zfs
	for _, zfs := range allZFS {
		if slices.Contains(runningZFS, zfs) {
			continue
		}
		if zfs.running {
			pendingStoppedZFS = append(pendingStoppedZFS, zfs)
		}
	}
	return pendingStoppedZFS
}

// Получение датасетов, после остановки VM
func getPendingStartZFS(allZFS []zfs, runningZFS []zfs) []zfs {
	var pendingStartZFS []zfs
	for _, zfs := range allZFS {
		if !slices.Contains(runningZFS, zfs) {
			continue
		}
		if !zfs.running {
			pendingStartZFS = append(pendingStartZFS, zfs)
		}
	}
	return pendingStartZFS
}

func processPendingsZFS(pending *Pending, pendingStopZFS []zfs, pendingStartZFS []zfs, env environment) {
	for _, zfs := range pendingStopZFS {
		pending.Snapshots = append(pending.Snapshots, fmt.Sprintf("%s@autosnap_%s_stopped", zfs.name, env.time.human))
		pending.UnsetRunning = append(pending.UnsetRunning, zfs.name)
	}
	for _, zfs := range pendingStartZFS {
		pending.SetRunning = append(pending.SetRunning, zfs.name)
	}
}

// функция разбивки снимков на группы по типу
// Пример: @autosnap_[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}_yearly
// Типы: frequently, hourly, daily, monthly, yearly
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

// Фильтрация nosnap zfs датасетов
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

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		help()
		os.Exit(1)
	}
}

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

	if timeLast+policy.interval < timeNowUnix {
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
	// fmt.Printf("%+v\n", env)

	if isTerminal() {
		fmt.Println("Запуск в режиме терминала")
		updateCron()
	}

	executor := OSExec{}

	poolList, err := ZpoolList(executor)
	checkErr(err)

	qmList, err := QmList(executor)
	checkErr(err)
	pctList, err := PctList(executor)
	checkErr(err)
	runningVMIDs := getRunningVMIDs(qmList, pctList)

	for _, pool := range poolList {
		pending := Pending{Pool: pool}

		allZFS, err := ZFSlist(executor, pool)
		checkErr(err)

		runningZFS := filterZfsInVms(allZFS, runningVMIDs)

		pendingStopZFS := getPendingStopZFS(allZFS, runningZFS)

		pendingStartZFS := getPendingStartZFS(allZFS, runningZFS)

		processPendingsZFS(&pending, pendingStopZFS, pendingStartZFS, env)

		// Фильтруем nosnap zfs датасеты
		runningZFS = filterNoSnap(runningZFS)

		for _, zfs := range runningZFS {
			snapshots, err := ZfsListSnapshots(executor, zfs.name)
			checkErr(err)

			// снапшоты сгруппированные по типам и отфильтрованы по шаблону
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
