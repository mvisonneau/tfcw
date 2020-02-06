package cli

import (
	"testing"
	"time"
)

func TestRunCli(t *testing.T) {
	version := "0.0.0"
	app := Init(&version, time.Now())
	if app.Name != "tfcs" {
		t.Fatalf("Expected app.Name to be tfcs, got '%s'", app.Name)
	}

	if app.Version != "0.0.0" {
		t.Fatalf("Expected app.Version to be 0.0.0, got '%s'", app.Version)
	}
}
