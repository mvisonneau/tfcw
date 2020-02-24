package client

import (
	"bytes"
	"net"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	providerVault "github.com/mvisonneau/tfcw/lib/providers/vault"
	"github.com/mvisonneau/tfcw/lib/schemas"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestGetAndProcessVaultValues(t *testing.T) {
	ln, client := createTestVault(t)
	defer ln.Close()
	c := Client{
		Vault: &providerVault.Client{
			Client: client,
		},
	}

	path := "secret/foo"
	key := "foo"
	v := &schemas.Variable{
		Vault: &schemas.Vault{
			Path: &path,
			Key:  &key,
		},
	}

	value, err := c.getAndProcessVaultValues(v)
	assert.Equal(t, err, nil)
	assert.Equal(t, value, "bar")
}

func TestIsVariableAlreadyProcessed(t *testing.T) {
	c := &Client{
		ProcessedVariables: map[string]schemas.VariableKind{},
	}

	v1 := "foo"
	assert.Equal(t, c.isVariableAlreadyProcessed(v1, schemas.VariableKindEnvironment), false)
	assert.Equal(t, c.isVariableAlreadyProcessed(v1, schemas.VariableKindEnvironment), true)
	assert.Equal(t, c.isVariableAlreadyProcessed(v1, schemas.VariableKindTerraform), false)
	assert.Equal(t, c.isVariableAlreadyProcessed(v1, schemas.VariableKindTerraform), true)
}

func TestLogVariable(t *testing.T) {
	// redirect logs to str variable
	var str bytes.Buffer
	log.SetOutput(&str)

	// dry-run mode with no value
	v := &schemas.Variable{
		Name: "foo",
		Kind: schemas.VariableKindEnvironment,
	}

	logVariable(v, true)
	assert.Equal(t, str.String()[28:], "level=info msg=\"[DRY-RUN] Set variable 'foo' (environment) : **********\"\n")

	// no dry-mode with value set
	v.Value = "love"
	str.Reset()
	logVariable(v, false)
	assert.Equal(t, str.String()[28:], "level=info msg=\"Set variable 'foo' (environment)\"\n")
}

func TestSecureSensitiveString(t *testing.T) {
	assert.Equal(t, secureSensitiveString("foo"), "**********")
	assert.Equal(t, secureSensitiveString("love"), "l********e")
}

func createTestVault(t *testing.T) (net.Listener, *api.Client) {
	t.Helper()

	// Create an in-memory, unsealed core (the "backend", if you will).
	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
	_ = keyShares

	// Start an HTTP server for the core.
	ln, addr := http.TestServer(t, core)

	// Create a client that talks to the server, initially authenticating with
	// the root token.
	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	client.SetToken(rootToken)

	// Setup required secrets, policies, etc.
	_, err = client.Logical().Write("secret/foo", map[string]interface{}{
		"foo": "bar",
		"baz": "baz",
	})
	if err != nil {
		t.Fatal(err)
	}

	return ln, client
}
