package main

import (
	"reflect"
	"testing"
)

func TestGetSnapshotNames(t *testing.T) {
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
	expected := []string{
		"vm-100-disk-1@autosnap_2020-10-21_04:00:02_yearly",
		"vm-100-disk-1@autosnap_2021-10-21_04:00:02_yearly",
		"vm-100-disk-1@autosnap_2022-10-21_04:00:02_yearly",
		"vm-100-disk-1@autosnap_2022-10-21_04:00:02_monthly",
		"vm-100-disk-1@autosnap_2022-11-21_04:00:02_monthly",
		"vm-100-disk-1@autosnap_2022-12-21_04:00:02_monthly",
		"vm-100-disk-1@autosnap_2023-01-21_04:00:02_monthly",
		"vm-100-disk-1@autosnap_2023-01-22_04:00:02_daily",
		"vm-100-disk-1@autosnap_2023-01-23_04:00:02_daily",
		"vm-100-disk-1@autosnap_2023-01-24_04:00:02_daily",
		"vm-100-disk-1@autosnap_2023-01-24_05:00:02_hourly",
		"vm-100-disk-1@autosnap_2023-01-24_06:00:02_hourly",
		"vm-100-disk-1@autosnap_2023-01-24_07:00:02_hourly",
		"vm-100-disk-1@autosnap_2023-01-24_07:15:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_07:30:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_07:45:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_08:00:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_08:15:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_08:30:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_08:45:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_09:00:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_09:15:02_frequently",
		"vm-100-disk-1@autosnap_2023-01-24_09:30:02_frequently",
	}
	got := snapshotsToNames(snapshots)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("getSnapshotNames() = %v, want %v", got, expected)
	}
}
