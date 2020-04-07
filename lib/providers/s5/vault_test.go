package s5

import (
	"os"
	"testing"

	"github.com/mvisonneau/s5/cipher"
	cipherVault "github.com/mvisonneau/s5/cipher/vault"
	"github.com/mvisonneau/tfcw/lib/schemas"
	"github.com/stretchr/testify/assert"
)

const (
	testVaultTransitKey string = "foo"
)

func TestGetCipherEngineVault(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypeVault
	key := testVaultTransitKey

	os.Setenv("VAULT_ADDR", "http://foo")
	os.Setenv("VAULT_TOKEN", "bar")

	// expected engine
	expectedEngine, err := cipher.NewVaultClient(key)
	assert.Nil(t, err)

	// all defined in client, empty variable config (default settings)
	v := &schemas.S5{}
	c := &Client{
		CipherEngineType: &cipherEngineType,
		CipherEngineVault: &schemas.S5CipherEngineVault{
			TransitKey: &key,
		},
	}

	cipherEngine, err := c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherVault.Client).Config)

	// all defined in variable, empty client config
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineVault: &schemas.S5CipherEngineVault{
			TransitKey: &key,
		},
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherVault.Client).Config)

	// key defined in environment variable
	os.Setenv("S5_VAULT_TRANSIT_KEY", testVaultTransitKey)
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherVault.Client).Config)

	// other engine & key defined in client, overridden in variable
	otherCipherEngineType := schemas.S5CipherEngineTypeAES
	otherTransitKey := "bar"

	c = &Client{
		CipherEngineType: &otherCipherEngineType,
		CipherEngineVault: &schemas.S5CipherEngineVault{
			TransitKey: &otherTransitKey,
		},
	}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineVault: &schemas.S5CipherEngineVault{
			TransitKey: &key,
		},
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherVault.Client).Config)
}
