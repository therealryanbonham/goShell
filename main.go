package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/cosiner/argv"
	"github.com/therealryanbonham/go-prompt" //Loading Custom version of go-prompt to allow custom ansi colors in prompt prefix

	"github.com/go-playground/ansi"
	log "github.com/sirupsen/logrus"
)

// Result Result of the command run
type Result struct {
	Message string
	Error   error
}

// History current command line history
var History []string

// HomeDirs contains current HomeDir and previous HomeDir if it has changed
var HomeDirs struct {
	current string
	last    string
}

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
			userHomeDir()
			//reader := bufio.NewReader(os.Stdin)
			input := prompt.Input(buildPrompt(), completer,
				prompt.OptionHistory(History),
				prompt.OptionTitle(buildTitleBar()))

			// fmt.Print(buildPrompt())
			// input, _ := reader.ReadString('\n')
			cmdreturn := parseCmdString(input)
			if cmdreturn.Error != nil {
				fmt.Println(cmdreturn.Error.Error())
			} else {
				fmt.Println(cmdreturn.Message)
				// Do Not Save history if it is a repeat of last item.
				lastItem := ""
				if len(History) > 0 {
					lastItem = History[len(History)-1]
				}
				if lastItem != input {
					History = append(History, input)
					saveHistory(HomeDirs.current, History, 1000)
				}
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
	var promptPath string
	if len == 1 {
		promptPath = ansi.Cyan + "/" + ansi.White
	} else {
		promptPath = ansi.Cyan + string(strings.Split(pwd, "/")[len]) + ansi.White
	}
	suffix := addPromptSuffix()
	if suffix != "" {
		promptPath = promptPath + " " + suffix
	}
	return promptPath + " >"

}
func addPromptSuffix() string {
	var suffix string
	//See if this is a git repo.
	b, err := ioutil.ReadFile("./.git/HEAD") // just pass the file name
	if err == nil {
		head := strings.Split(strings.Trim(string(b), "\n"), "/")
		branch := head[len(head)-1]
		suffix = ansi.Blue + "git:(" + ansi.Red + branch + ansi.Blue + ")" + ansi.White
	}
	return suffix
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
		subs := re.FindAllString(s, -1)
		for _, sub := range subs {
			log.WithFields(log.Fields{"s": s, "sub": sub}).Debug("Found Sub Cmd")
			// cmdResult := re.ReplaceAllStringFunc(s, parseCmdSubString)
			cmdResult = parseCmdSubString(sub)
			if cmdResult.Error == nil {
				subReplace := cmdResult.Message
				log.WithFields(log.Fields{"s": s, "sub": sub, "subReplace": subReplace}).Debug("Sub Cmd Result")
				s = strings.Replace(s, sub, subReplace, 1)
			} else {
				return cmdResult
			}
		}
	}
	log.WithFields(log.Fields{"s": s}).Debug("Final Cmd")
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
		if len(command) == 1 || command[1] == "" {
			cmdResult.Error = errors.New("You must specify a path")
			o <- cmdResult
			return
		}
		if command[1] == "~" {
			command[1] = HomeDirs.current
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
	if command[0] == "sudo" && command[1] == "su" {
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
	} else {
		stdin := strings.NewReader(in)
		cmd.Stdin = stdin
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

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
