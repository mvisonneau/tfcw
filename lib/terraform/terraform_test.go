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
	if err != nil {
		t.Fatal(err)
	}
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
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString(test.config)
		f.Close()

		rbc, err := GetRemoteBackendConfig(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, rbc)
		assert.Equal(t, test.expected, *rbc)
	}
}
