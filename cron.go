package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

// Провряем, что терминал подключен
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func updateCron() {
	// Ваша команда и аргументы
	executable := os.Args[0]
	args := os.Args[1:]

	// Сформируйте строку для добавления в cron
	command := fmt.Sprintf("*/15 * * * * %s %s", executable, strings.Join(args, " "))

	// Чтение текущих заданий cron
	out, err := exec.Command("crontab", "-l").Output()
	if err != nil {
		fmt.Println("Ошибка при получении текущих заданий cron:", err)
		os.Exit(1)
	}

	lines := strings.Split(string(out), "\n")
	found := false
	newCronLines := []string{}

	// Проверяем каждую строку на наличие нашей команды
	for _, line := range lines {
		if strings.Contains(line, executable) && !strings.HasPrefix(line, "#") {
			found = true
			newCronLines = append(newCronLines, command) // заменяем существующую строку
		} else if line != "" {
			newCronLines = append(newCronLines, line)
		}
	}

	// Если команда не найдена, добавляем ее
	if !found {
		newCronLines = append(newCronLines, command)
	}

	// Создаем новую задачу cron
	newCron := strings.Join(newCronLines, "\n") + "\n"
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCron)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		os.Exit(1)
	}

	fmt.Println("Задание cron обновлено.")
}
