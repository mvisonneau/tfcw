package s5

import (
	"os"

	"github.com/mvisonneau/s5/cipher"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func (c *Client) getCipherEngineAES(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineAES != nil && v.CipherEngineAES.Key != nil {
		return cipher.NewAESClient(*v.CipherEngineAES.Key)
	}

	if c.CipherEngineAES != nil && c.CipherEngineAES.Key != nil {
		return cipher.NewAESClient(*c.CipherEngineAES.Key)
	}

	return cipher.NewAESClient(os.Getenv("S5_AES_KEY"))
}
