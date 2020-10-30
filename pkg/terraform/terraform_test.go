package terraform

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var emptyTerraformConfig = `
terraform {
  backend "remote" {
		organization = "foo"
  }
}
`

var completeTerraformConfig = `
terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "foo"
		token        = "bar"
		
		workspaces {
      name = "baz"
    }
  }
}
`

func TestGetRemoteBackendConfig(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tfcw-test")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		config   string
		expected RemoteBackendConfig
	}{
		{
			emptyTerraformConfig,
			RemoteBackendConfig{
				Organization: "foo",
			},
		},
		{
			completeTerraformConfig,
			RemoteBackendConfig{
				Hostname:     "app.terraform.io",
				Organization: "foo",
				Token:        "bar",
				Workspace:    "baz",
			},
		},
	}

	for _, test := range tests {
		f, err := os.Create(tmpDir + "/terraform.tf")
		assert.Nil(t, err)
		f.WriteString(test.config)
		f.Close()

		rbc, err := GetRemoteBackendConfig(tmpDir)
		assert.Nil(t, err)
		assert.NotNil(t, rbc)
		assert.Equal(t, test.expected, *rbc)
	}
}
