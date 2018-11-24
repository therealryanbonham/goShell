package main

import (
	"testing"

	"github.com/acarl005/stripansi"
)

func TestBuildPrompt(t *testing.T) {
	prompt := stripansi.Strip(buildPrompt())
	if prompt != "goShell git:(master) >" {
		t.Errorf("Expected prompt of 'goShell git:(master) >' got %s", prompt)
	}
}

func TestRunCmd(t *testing.T) {
	out := make(chan Result)
	go runCmd(out, "echo cat", "")
	result := <-out
	if result.Error != nil {
		t.Errorf("Unexpected Error %s", result.Error.Error())
	}
	if result.Message != "cat" {
		t.Errorf("Expected Result 'cat' got %s", result.Message)
	}
}
func TestRunChangeDir(t *testing.T) {
	out := make(chan Result)
	go runCmd(out, "cd ..", "")
	result := <-out
	if result.Error != nil {
		t.Errorf("Unexpected Error %s", result.Error.Error())
	}
}

func TestRunChangeBadDir(t *testing.T) {
	out := make(chan Result)
	go runCmd(out, "cd bAdD!r", "")
	result := <-out
	if result.Error == nil {
		t.Errorf("Expected an error as this is directory does not exist")
	}
}

func TestRunChangeDirToFile(t *testing.T) {
	out := make(chan Result)
	go runCmd(out, "cd main.go", "")
	result := <-out
	if result.Error == nil {
		t.Errorf("Expected an error as this is  a file not a directory")
	}
}

func TestRunChangeDirNoArg(t *testing.T) {
	out := make(chan Result)
	go runCmd(out, "cd", "")
	result := <-out
	if result.Error == nil {
		t.Errorf("Expected an error no dir was passed")
	}
}
func TestReturnCmd(t *testing.T) {
	result := returnCmd("echo cat", "")
	if result.Message != "cat" {
		t.Errorf("Expected Result 'cat' got %s", result.Message)
	}
}

func TestStdIn(t *testing.T) {
	result := returnCmd("cat", "dog")
	if result.Message != "dog" {
		t.Errorf("Expected Result 'dog' got %s", result.Message)
	}
}

func TestParseCmdSubString(t *testing.T) {
	result := parseCmdSubString("$(echo dog)")
	if result.Message != "dog" {
		t.Errorf("Expected Result 'dog' got %s", result.Message)
	}
}
func TestFindSubCmdStrings(t *testing.T) {
	result := findSubCmdStrings("echo $(echo dog)", "")
	if result.Message != "dog" {
		t.Errorf("Expected Result 'dog' got %s", result.Message)
	}
}

func TestParseCmdString(t *testing.T) {
	result := parseCmdString("echo $(echo dog) and cat| cat")
	if result.Message != "dog and cat" {
		t.Errorf("Expected Result 'dog and cat' got %s", result.Message)
	}
}

func TestParseCmdStringBadCommand(t *testing.T) {
	result := parseCmdString("kfasjfasfjaspojf4")
	if result.Error == nil {
		t.Errorf("Expected Error Bad Command %s", result.Error.Error())
	}
}
