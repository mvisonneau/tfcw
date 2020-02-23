package s5

import (
	"os"

	"github.com/mvisonneau/s5/cipher"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func (c *Client) getCipherEngineAWS(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineAWS != nil && v.CipherEngineAWS.KmsKeyArn != nil {
		return cipher.NewAWSClient(*v.CipherEngineAWS.KmsKeyArn)
	}

	if c.CipherEngineAWS != nil && c.CipherEngineAWS.KmsKeyArn != nil {
		return cipher.NewAWSClient(*c.CipherEngineAWS.KmsKeyArn)
	}

	return cipher.NewAWSClient(os.Getenv("S5_AWS_KMS_KEY_ARN"))
}
