// +build !darwin
// There seems to be a bug in a lib importer by hashicorp/vault/api that prevents the test from running
// correctly on darwin..

package vault

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/mvisonneau/tfcw/pkg/schemas"
	"github.com/stretchr/testify/assert"
)

var wd, _ = os.Getwd()

func TestGetClient(t *testing.T) {
	os.Setenv("HOME", wd)
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("VAULT_TOKEN")

	_, err := GetClient("", "")
	assert.Equal(t, fmt.Errorf("Vault address is not defined"), err)

	_, err = GetClient("foo", "")
	assert.Equal(t, fmt.Errorf("Vault token is not defined (VAULT_TOKEN or ~/.vault-token)"), err)

	_, err = GetClient("foo", "bar")
	assert.Nil(t, err)

	os.Setenv("VAULT_ADDR", "foo")
	os.Setenv("VAULT_TOKEN", "bar")
	_, err = GetClient("", "")
	assert.Nil(t, err)
}

// func TestGetValues(t *testing.T) {
// 	ln, client := createTestVault(t)
// 	defer ln.Close()
// 	c := Client{client}

// 	// Undefined path
// 	v := &schemas.Vault{}
// 	r, err := c.GetValues(v)
// 	assert.Equal(t, fmt.Errorf("no path defined for retrieving vault secret"), err)
// 	assert.Equal(t, map[string]string{}, r)

// 	// Valid secret
// 	validPath := "secret/foo"
// 	v.Path = &validPath

// 	r, err = c.GetValues(v)
// 	assert.Nil(t, err)
// 	assert.Equal(t, map[string]string{"secret": "bar"}, r)

// 	// Unexistent secret
// 	invalidPath := "secret/baz"
// 	v.Path = &invalidPath

// 	r, err = c.GetValues(v)
// 	assert.Equal(t, fmt.Errorf("no results/keys returned for secret : secret/baz"), err)
// 	assert.Equal(t, map[string]string{}, r)

// 	// Invalid method
// 	invalidMethod := "foo"
// 	v.Method = &invalidMethod
// 	r, err = c.GetValues(v)
// 	assert.Equal(t, fmt.Errorf("unsupported method 'foo'"), err)
// 	assert.Equal(t, map[string]string{}, r)

// 	// Write method
// 	writeMethod := "write"
// 	params := map[string]string{"foo": "bar"}
// 	v.Method = &writeMethod
// 	v.Path = &validPath
// 	v.Params = &params

// 	r, err = c.GetValues(v)
// 	assert.Equal(t, fmt.Errorf("no results/keys returned for secret : secret/foo"), err)
// 	assert.Equal(t, map[string]string{}, r)
// }

// func createTestVault(t *testing.T) (net.Listener, *api.Client) {
// 	t.Helper()

// 	// Create an in-memory, unsealed core (the "backend", if you will).
// 	core, keyShares, rootToken := vault.TestCoreUnsealed(t)
// 	_ = keyShares

// 	// Start an HTTP server for the core.
// 	ln, addr := http.TestServer(t, core)

// 	// Create a client that talks to the server, initially authenticating with
// 	// the root token.
// 	conf := api.DefaultConfig()
// 	conf.Address = addr

// 	client, err := api.NewClient(conf)
// 	assert.Nil(t, err)
// 	client.SetToken(rootToken)

// 	// Setup required secrets, policies, etc.
// 	_, err = client.Logical().Write("secret/foo", map[string]interface{}{
// 		"secret": "bar",
// 	})
// 	assert.Nil(t, err)

// 	return ln, client
// }
