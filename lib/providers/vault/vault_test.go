package vault

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/mvisonneau/go-helpers/test"
	"github.com/mvisonneau/tfcw/lib/schemas"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
)

var (
	wd, _ = os.Getwd()
)

func TestGetClient(t *testing.T) {
	os.Setenv("HOME", wd)
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("VAULT_TOKEN")

	_, err := GetClient("", "")
	test.Expect(t, err, fmt.Errorf("Vault address is not defined"))

	_, err = GetClient("foo", "")
	test.Expect(t, err, fmt.Errorf("Vault token is not defined (VAULT_TOKEN or ~/.vault-token)"))

	_, err = GetClient("foo", "bar")
	test.Expect(t, err, nil)

	os.Setenv("VAULT_ADDR", "foo")
	os.Setenv("VAULT_TOKEN", "bar")
	_, err = GetClient("", "")
	test.Expect(t, err, nil)
}

func TestGetValues(t *testing.T) {
	ln, client := createTestVault(t)
	defer ln.Close()
	c := Client{client}

	// Undefined path
	v := &schemas.Vault{}
	r, err := c.GetValues(v)
	test.Expect(t, err, fmt.Errorf("no path defined for retrieving vault secret"))
	test.Expect(t, r, map[string]string{})

	// Valid secret
	validPath := "secret/foo"
	v.Path = &validPath

	r, err = c.GetValues(v)
	test.Expect(t, err, nil)
	test.Expect(t, r, map[string]string{"secret": "bar"})

	// Unexistent secret
	invalidPath := "secret/baz"
	v.Path = &invalidPath

	r, err = c.GetValues(v)
	test.Expect(t, err, fmt.Errorf("no results/keys returned for secret : secret/baz"))
	test.Expect(t, r, map[string]string{})

	// Invalid method
	invalidMethod := "foo"
	v.Method = &invalidMethod
	r, err = c.GetValues(v)
	test.Expect(t, err, fmt.Errorf("unsupported method 'foo'"))
	test.Expect(t, r, map[string]string{})

	// Write method
	writeMethod := "write"
	params := map[string]string{"foo": "bar"}
	v.Method = &writeMethod
	v.Path = &validPath
	v.Params = &params

	r, err = c.GetValues(v)
	test.Expect(t, err, fmt.Errorf("no results/keys returned for secret : secret/foo"))
	test.Expect(t, r, map[string]string{})
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
		"secret": "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	return ln, client
}
