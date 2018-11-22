package main

import (
	"testing"
)

func TestLoadHistoryInvalidPath(t *testing.T) {
	History = loadHistory("./badpath")
	if len(History) == 0 {
		t.Error("Invalid Path should return a blank History")
	}
}
func TestSaveAndLoadHistory(t *testing.T) {
	History = []string{
		"ls -al",
		"echo dog",
		"cat",
	}
	saveHistory(".", History, 1000)
	History = loadHistory(".")
	if len(History) != 3 {
		t.Errorf("Load of saved history failed. Got %d items expected 3", len(History))
		t.Errorf("%+v", History)
	}
}

func TestSaveHistoryLimit(t *testing.T) {
	History = []string{
		"ls -al",
		"echo dog",
		"cat",
	}
	saveHistory(".", History, 2)
	History = loadHistory(".")
	if len(History) != 2 {
		t.Errorf("Load of saved history failed. Got %d items expected 3", len(History))
		t.Errorf("%+v", History)
	}
	if History[0] != "echo dog" {
		t.Errorf("Expected first item to be 'echo dog' got %s", History[0])
	}
	if History[1] != "cat" {
		t.Errorf("Expected second item to be 'cat' got %s", History[1])
	}
}
