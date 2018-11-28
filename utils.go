package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"
	prompt "github.com/therealryanbonham/go-prompt"
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
func getExecutablesinCurrentDir() []prompt.Suggest {
	paths := []string{"."}
	return getExecutables(paths)
}
func getExecutablesInPath() []prompt.Suggest {
	result := parseCmdString("echo $PATH | tr ':' '\n'")
	if result.Error != nil {
		log.Error("Failed to load paths")
	}
	paths := strings.Split(result.Message, "\n")
	return getExecutables(paths)
}
func getExecutables(paths []string) []prompt.Suggest {
	var execs []prompt.Suggest

	out := make(chan Result)
	for _, path := range paths {
		go runCmd(out, `find "`+path+`" -type f -maxdepth 1 -print0 -exec test -x {} ;`, "")
	}
	for range paths {
		findResults := <-out
		log.WithFields(log.Fields{"findResults": findResults}).Debug("Exec Find Results")
		foundExecs := strings.Split(findResults.Message, "\x00")
		for _, f := range foundExecs {
			if f != "" {
				fs := strings.Split(f, "/")
				//log.WithFields(log.Fields{"f": f, "fs": fs, "execs": execs}).Debug("Adding Exec")
				cmd := fs[len(fs)-1]
				if cmd != "cd" {
					var exec prompt.Suggest
					exec.Text = strings.Replace(cmd, " ", "\\ ", -1)
					exec.Description = f
					execs = append(execs, exec)
				}
			}
		}
	}
	return execs
}

func getDirectories(paths []string) []prompt.Suggest {
	var execs []prompt.Suggest

	out := make(chan Result)
	for _, path := range paths {
		go runCmd(out, `find "`+path+`" -type d -maxdepth 1 -print0`, "")
	}
	for range paths {
		findResults := <-out
		if findResults.Error == nil {
			log.WithFields(log.Fields{"findResults": findResults}).Debug("Dir Find Results")
			foundExecs := strings.Split(findResults.Message, "\x00")
			for _, f := range foundExecs {
				if f != "" {
					//fs := strings.Split(f, "/")
					//log.WithFields(log.Fields{"f": f, "fs": fs, "execs": execs}).Debug("Adding Exec")
					var exec prompt.Suggest
					exec.Text = strings.Replace(f, " ", "\\ ", -1)
					exec.Description = f
					execs = append(execs, exec)
				}
			}
		}

	}
	return execs
}

func suggestCDirectory(i string, withExecs bool) []prompt.Suggest {

	lastSlash := strings.LastIndex(i, "/")

	searchPath := i
	log.WithFields(log.Fields{"i": i, "lastSlash": lastSlash}).Debug("suggestChangeDirectory")
	if lastSlash == -1 {
		searchPath = "."
	} else if lastSlash != len(i)-1 {
		split := strings.Split(i, "/")
		searchPath = strings.Join(split[:len(split)-1], "/")
		log.WithFields(log.Fields{"split": split, "searchPath": searchPath}).Debug("Search Path")
		if searchPath == "" {
			searchPath = "/"
		}
	}
	paths := []string{searchPath}
	items := getDirectories(paths)
	if withExecs {
		execs := getExecutables(paths)
		items = append(items, execs...)

	}
	log.WithFields(log.Fields{"i": i, "searchPath": searchPath, "items": items}).Debug("Found Path")
	return items

}
