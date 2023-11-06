package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Pending struct {
	Pool       string
	Hosname    string
	Snapshots  []string
	Destroys   []string
	SetRunning []string
	SetStopped []string
}

func (p *Pending) Run() error {
	if p.Pool == "" {
		return fmt.Errorf("pool is empty")
	}
	if len(p.Snapshots) > 0 {
		_, _ = program(p.Pool, "lua_snapshot", p.Snapshots)
		// fmt.Println("snapshots", p.Snapshots)
	}
	if len(p.Destroys) > 0 {
		_, _ = program(p.Pool, "lua_destroy", p.Destroys)
		// fmt.Println("destroys", p.Destroys)
	}
	if len(p.SetRunning) > 0 {
		args := append([]string{p.Hosname}, p.SetRunning...)
		_, _ = program(p.Pool, "lua_set_running", args)
		// fmt.Println("set_running", p.SetRunning)
	}
	if len(p.SetStopped) > 0 {
		args := append([]string{"stopped"}, p.SetStopped...)
		_, _ = program(p.Pool, "lua_set_running", args)
		// fmt.Println("unset_running", p.UnsetRunning)
	}
	return nil
}

func program(pool string, program string, args []string) (string, error) {
	output, err := exec.Command("bash", "-c",
		fmt.Sprintf("zfs program %s <(%s %s) %s",
			pool, os.Args[0], program, strings.Join(args, " ")),
	).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
