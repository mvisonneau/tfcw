package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/pkg/configs"
	log "github.com/sirupsen/logrus"
)

// RemoteBackendConfig represents partial set of values
// for the remote backend configuration we can set in
// a Terraform file.
type RemoteBackendConfig struct {
	Hostname     string
	Organization string
	Token        string
	Workspace    string
}

// GetRemoteBackendConfig attempts to return the remote backend configuration
// from a Terraform file.
func GetRemoteBackendConfig(workingDir string) (*RemoteBackendConfig, error) {
	p := configs.NewParser(nil)

	c, err := p.LoadConfigDir(workingDir)
	if err.HasErrors() {
		return nil, err
	}

	if c.Backend == nil {
		log.Debug("terraform remote backend not configured, skipping this evaluation..")
		return nil, nil
	}

	if c.Backend.Type != "remote" {
		return nil, fmt.Errorf("Terraform state backend has not been configured as 'remote'")
	}

	rbc := &RemoteBackendConfig{}

	remote, _, err := c.Backend.Config.PartialContent(remoteBackendSchema)
	if err != nil {
		return nil, err
	}

	if _, ok := remote.Attributes["hostname"]; ok {
		hostnameVar, err := remote.Attributes["hostname"].Expr.Value(nil)
		if err != nil {
			return nil, err
		}
		rbc.Hostname = hostnameVar.AsString()
	}

	if _, ok := remote.Attributes["organization"]; ok {
		organizationVar, err := remote.Attributes["organization"].Expr.Value(nil)
		if err != nil {
			return nil, err
		}
		rbc.Organization = organizationVar.AsString()
	}

	if _, ok := remote.Attributes["token"]; ok {
		tokenVar, err := remote.Attributes["token"].Expr.Value(nil)
		if err != nil {
			return nil, err
		}
		rbc.Token = tokenVar.AsString()
	}

	for _, block := range remote.Blocks {
		w, _, err := block.Body.PartialContent(remoteBackendWorkspaceSchema)
		if err != nil {
			return nil, err
		}

		if _, ok := w.Attributes["name"]; ok {
			workspaceVar, err := w.Attributes["name"].Expr.Value(nil)
			if err != nil {
				return nil, err
			}
			rbc.Workspace = workspaceVar.AsString()
		}
	}

	return rbc, nil
}

var remoteBackendSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "organization",
		},
		{
			Name: "hostname",
		},
		{
			Name: "token",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "workspaces",
		},
	},
}

var remoteBackendWorkspaceSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "name",
		},
	},
}
