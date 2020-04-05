package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExit(t *testing.T) {
	err := exit(20, fmt.Errorf("test"))
	assert.Equal(t, err.Error(), "")
	assert.Equal(t, err.ExitCode(), 20)
}

func TestComputeConfigFilePath(t *testing.T) {
	assert.Equal(t, "./tfcw.hcl", computeConfigFilePath(".", "<working-dir>/tfcw.hcl"))
	assert.Equal(t, "/foo/bar/tfcw.hcl", computeConfigFilePath(".", "/foo/bar/tfcw.hcl"))
}

func TestHTTPSPrefixedURL(t *testing.T) {
	assert.Equal(t, "https://app.terraform.io", returnHTTPSPrefixedURL("app.terraform.io"))
	assert.Equal(t, "http://app.terraform.io", returnHTTPSPrefixedURL("http://app.terraform.io"))
}
