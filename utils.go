package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"
)

func loadHistory(path string) []string {
	history, err := readLines(path + "/.goShellHistory")
	if err != nil {
		return []string{""}
	}
	return history
}
func saveHistory(path string, h []string, i int) {
	if i < 0 {
		log.Error("History Limit can not be negative.")
	}
	if len(h) > i {
		start := len(h) - i
		h = h[start:]
	}
	err := writeLines(h, path+"/.goShellHistory")
	if err != nil {
		log.Error(err.Error())
	}

}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}
