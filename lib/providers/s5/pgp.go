package s5

import (
	"os"

	"github.com/mvisonneau/s5/cipher"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func (c *Client) getCipherEnginePGP(v *schemas.S5) (cipher.Engine, error) {
	var publicKeyPath, privateKeyPath string

	if v.CipherEnginePGP != nil && v.CipherEnginePGP.PublicKeyPath != nil {
		publicKeyPath = *v.CipherEnginePGP.PublicKeyPath
	} else if c.CipherEnginePGP != nil && c.CipherEnginePGP.PublicKeyPath != nil {
		publicKeyPath = *c.CipherEnginePGP.PublicKeyPath
	} else {
		publicKeyPath = os.Getenv("S5_PGP_PUBLIC_KEY_PATH")
	}

	if v.CipherEnginePGP != nil && v.CipherEnginePGP.PrivateKeyPath != nil {
		privateKeyPath = *v.CipherEnginePGP.PrivateKeyPath
	} else if c.CipherEnginePGP != nil && c.CipherEnginePGP.PrivateKeyPath != nil {
		privateKeyPath = *c.CipherEnginePGP.PrivateKeyPath
	} else {
		privateKeyPath = os.Getenv("S5_PGP_PUBLIC_KEY_PATH")
	}

	return cipher.NewPGPClient(publicKeyPath, privateKeyPath)
}
