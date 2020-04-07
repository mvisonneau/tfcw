package client

import (
	"fmt"

	tfc "github.com/hashicorp/go-tfe"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRunWorkspaceOperationsValue(t *testing.T) {
	cfg := getTestConfig()
	c, err := NewClient(cfg)
	assert.NoError(t, err)
	w := &tfc.Workspace{}
	assert.Equal(t, fmt.Errorf("remote operations must be enabled on the workspace"), c.CreateRun(cfg, w, &TFCCreateRunOptions{}))

	w.Operations = true
	assert.NotEqual(t, fmt.Errorf("remote operations must be enabled on the workspace"), c.CreateRun(cfg, w, &TFCCreateRunOptions{}))
}
