package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"
)

func loadHistory(path string) []string {
	historyPath := path + "/.goShellHistory"
	log.WithFields(log.Fields{"historyPath": historyPath}).Debug("Loading History")
	history, err := readLines(historyPath)
	if err != nil {
		log.WithFields(log.Fields{"historyPath": historyPath}).Debug("Failed to Load History")
		return []string{""}
	}

	return history
}
func saveHistory(path string, h []string, i int) {
	historyPath := path + "/.goShellHistory"
	if i < 0 {
		log.Error("History Limit can not be negative.")
	}
	if len(h) > 0 {
		if len(h) > i {
			start := len(h) - i
			h = h[start:]
		}
		err := writeLines(h, historyPath)
		if err != nil {
			log.Error(err.Error())
		}
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

func userHomeDir() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	HomeDirs.current = currentUser.HomeDir
	if HomeDirs.current != HomeDirs.last {
		HomeDirs.last = HomeDirs.current
		History = loadHistory(HomeDirs.current)
	}
}

func buildTitleBar() string {

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	userName := currentUser.Username
	homeDir := currentUser.HomeDir
	displayPath := strings.Replace(pwd, homeDir, "~", 1)
	return userName + "@" + hostname + ": " + displayPath
}
