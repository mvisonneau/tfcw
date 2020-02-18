package cli

import (
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	app := newApp("0.0.0", time.Now())
	if app.Name != "tfcw" {
		t.Fatalf("Expected app.Name to be tfcw, got '%s'", app.Name)
	}

	if app.Version != "0.0.0" {
		t.Fatalf("Expected app.Version to be 0.0.0, got '%s'", app.Version)
	}
}
