//go:build !darwin
// +build !darwin

package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	assert.NotPanics(t, func() { Run("0.0.0", []string{"vac", "--version"}) })
}

func TestNewApp(t *testing.T) {
	app := NewApp("0.0.0", time.Now())
	assert.Equal(t, "tfcw", app.Name)
	assert.Equal(t, "0.0.0", app.Version)
}
