package s5

import (
	"os"

	"github.com/mvisonneau/s5/pkg/cipher"
	"github.com/mvisonneau/tfcw/pkg/schemas"
)

func (c *Client) getCipherEngineGCP(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineGCP != nil && v.CipherEngineGCP.KmsKeyName != nil {
		return cipher.NewGCPClient(*v.CipherEngineGCP.KmsKeyName)
	}

	if c.CipherEngineGCP != nil && c.CipherEngineGCP.KmsKeyName != nil {
		return cipher.NewGCPClient(*c.CipherEngineGCP.KmsKeyName)
	}

	return cipher.NewGCPClient(os.Getenv("S5_GCP_KMS_KEY_NAME"))
}
