package s5

import (
	"os"

	"github.com/mvisonneau/s5/cipher"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func (c *Client) getCipherEngineVault(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineVault != nil && v.CipherEngineVault.TransitKey != nil {
		return cipher.NewVaultClient(*v.CipherEngineVault.TransitKey)
	}

	if c.CipherEngineVault != nil && c.CipherEngineVault.TransitKey != nil {
		return cipher.NewVaultClient(*c.CipherEngineVault.TransitKey)
	}
	return cipher.NewVaultClient(os.Getenv("S5_VAULT_TRANSIT_KEY"))
}
