package main

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	f()
	return buf.String()
}

func TestRun(t *testing.T) {
	output := captureOutput(run)
	if !strings.Contains(output, "Welcome to Helix!") {
		t.Errorf("expected output to contain 'Welcome to Helix!', got: %s", output)
	}
}

func TestMainFunc(t *testing.T) {
	output := captureOutput(main)
	if !strings.Contains(output, "Welcome to Helix!") {
		t.Errorf("expected output to contain 'Welcome to Helix!', got: %s", output)
	}
}
