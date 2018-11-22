package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/cosiner/argv"
	log "github.com/sirupsen/logrus"
)

// Result Result of the command run
type Result struct {
	Message string
	Error   error
}

// History current command line history
var History []string

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		// {Text: "users", Description: "Store the username and age"},
		// {Text: "articles", Description: "Store the article text posted by user"},
		// {Text: "comments", Description: "Store the text commented to articles"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	log.SetOutput(os.Stdout)
	for _, v := range os.Args[1:] {
		switch v {
		case "--debug":
			log.SetLevel(log.DebugLevel)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	//Shell Loop
	func() {
		for {
			home := userHomeDir()
			//reader := bufio.NewReader(os.Stdin)
			input := prompt.Input(buildPrompt(), completer,
				prompt.OptionHistory(loadHistory(home)))
			// fmt.Print(buildPrompt())
			// input, _ := reader.ReadString('\n')
			cmdreturn := parseCmdString(input)
			if cmdreturn.Error != nil {
				fmt.Println(cmdreturn.Error.Error())
			} else {
				fmt.Println(cmdreturn.Message)
				History = append(History, input)
				saveHistory(home, History, 1000)
			}

		}
	}()
}

func printHelp() {
	fmt.Println(`Usage: 
		 [--help|-h] for help
		 [--debug] for debug logs`)
}
func buildPrompt() string {
	//Get Current path
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	len := len(strings.Split(pwd, "/")) - 1
	return string(strings.Split(pwd, "/")[len]) + " >"

}
func parseCmdString(s string) Result {
	var cmdResult Result
	args, err := argv.Argv([]rune(s), argv.ParseEnv(os.Environ()), argv.Run)
	if err != nil {
		cmdResult.Error = err
		return cmdResult
	}

	log.WithFields(log.Fields{"args": args}).Debug("parseCmdString Args")
	log.WithFields(log.Fields{"s": s}).Debug("Calling Find Sub Cmd")
	for _, cmd := range args {
		cmdResult = findSubCmdStrings(strings.Join(cmd, " "), cmdResult.Message)
		if cmdResult.Error != nil {
			return cmdResult
		}
	}
	return cmdResult
}
func findSubCmdStrings(s string, in string) Result {
	var cmdResult Result
	re := regexp.MustCompile("\\$\\([^)\\\\]*(?:\\\\.[^)\\\\]*)*\\)")
	hasSub := re.FindStringIndex(s)
	if hasSub != nil {
		log.WithFields(log.Fields{"s": s}).Debug("Found Sub Cmd")
		subs := re.FindAllString(s, -1)
		for _, sub := range subs {
			// cmdResult := re.ReplaceAllStringFunc(s, parseCmdSubString)
			if cmdResult.Error != nil {
				subReplace := cmdResult.Message
				s = strings.Replace(s, sub, subReplace, 1)
			} else {
				return cmdResult
			}
		}
	}
	log.WithFields(log.Fields{"s": s}).Debug("Final Sub Cmd")
	return returnCmd(s, in)

}
func parseCmdSubString(s string) Result {
	s = s[2:]
	s = s[:len(s)-1]
	return parseCmdString(s)
}
func returnCmd(c string, in string) Result {
	out := make(chan Result)
	go runCmd(out, c, in)
	cmdreturn := <-out
	return cmdreturn
}
func runCmd(o chan Result, c string, in string) {
	var stdout, stderr bytes.Buffer
	var cmdResult Result
	command := strings.Split(c, " ")
	switch command[0] {
	case "exit":
		os.Exit(0)
	case "cd":
		if command[1] == "" {
			cmdResult.Error = errors.New("You must specify a path")
			o <- cmdResult
			return
		}
		fi, err := os.Stat(command[1])
		if err != nil {
			cmdResult.Error = err
			o <- cmdResult
			return
		}
		if !fi.Mode().IsDir() {
			cmdResult.Error = fmt.Errorf("%s is not a directory", command[1])
			o <- cmdResult
			return
		}
		err = os.Chdir(command[1])
		if err != nil {
			cmdResult.Error = err
			o <- cmdResult
		} else {
			cmdResult.Error = nil
			cmdResult.Message = ""
			o <- cmdResult
		}
		return
	}
	cmd := exec.Command(command[0], command[1:]...)
	log.WithFields(log.Fields{"stdin": in}).Debug("Setting stdIn")
	stdin := strings.NewReader(in)
	cmd.Stdin = stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.WithFields(log.Fields{"cmd": cmd}).Debug("Cmd")
	err := cmd.Run()
	if err != nil {
		cmdResult.Error = err
		o <- cmdResult
	} else {
		cmdResult.Error = nil
		cmdResult.Message = strings.Trim(string(stdout.Bytes()), "\n")
		o <- cmdResult
	}
	return

}
