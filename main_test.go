package main

import (
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	prompt := buildPrompt()
	if prompt != "goShell >" {
		t.Errorf("Expected prompt of 'goShell >' got %s", prompt)
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

func TestreturnCmd(t *testing.T) {
	result := returnCmd("echo cat", "")
	if result.Message != "cat" {
		t.Errorf("Expected Result 'cat' got %s", result.Message)
	}
}

func TeststdIn(t *testing.T) {
	result := returnCmd("cat", "dog")
	if result.Message != "dog" {
		t.Errorf("Expected Result 'dog' got %s", result.Message)
	}
}

func TestparseCmdSubString(t *testing.T) {
	result := parseCmdSubString("cat $(echo dog)")
	if result.Message != "dog" {
		t.Errorf("Expected Result 'dog' got %s", result.Message)
	}
}
func TestfindSubCmdStrings(t *testing.T) {
	result := findSubCmdStrings("cat $(echo dog)", "")
	if result.Message != "dog" {
		t.Errorf("Expected Result 'dog' got %s", result.Message)
	}
}

func TestparseCmdString(t *testing.T) {
	result := parseCmdString("echo $(echo dog) and cat| cat")
	if result.Message != "dog and cat" {
		t.Errorf("Expected Result 'dog and cat' got %s", result.Message)
	}
}

func TestparseCmdStringBadCommand(t *testing.T) {
	result := parseCmdString("kfasjfasfjaspojf4")
	if result.Error != nil {
		t.Errorf("Expected Error Bad Commandt %s", result.Error.Error())
	}
}
