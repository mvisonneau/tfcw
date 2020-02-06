package s5

import (
	"fmt"
	"os"

	"github.com/mvisonneau/s5/cipher"
	"github.com/mvisonneau/tfcs/lib/schemas"
)

type Client struct {
	CipherEngineType  *schemas.S5CipherEngineType
	CipherEngineAES   *schemas.S5CipherEngineAES
	CipherEngineAWS   *schemas.S5CipherEngineAWS
	CipherEngineGCP   *schemas.S5CipherEngineGCP
	CipherEnginePGP   *schemas.S5CipherEnginePGP
	CipherEngineVault *schemas.S5CipherEngineVault
}

func (c *Client) GetValue(v *schemas.S5) (string, error) {
	variableCipher, err := c.getCipherEngine(v)
	if err != nil {
		return "", err
	}

	parsedValue, err := cipher.ParseInput(*v.Value)
	if err != nil {
		return "", err
	}

	return variableCipher.Decipher(parsedValue)
}

func (c *Client) getCipherEngine(v *schemas.S5) (cipher.Engine, error) {
	var cipherEngineType *schemas.S5CipherEngineType
	if v.CipherEngineType != nil {
		cipherEngineType = v.CipherEngineType
	} else if c.CipherEngineType != nil {
		cipherEngineType = c.CipherEngineType
	} else {
		return nil, fmt.Errorf("You need to specify a cipher engine")
	}

	switch *cipherEngineType {
	case schemas.S5CipherEngineTypeAES:
		return c.getCipherEngineAES(v)
	case schemas.S5CipherEngineTypeAWS:
		return c.getCipherEngineAWS(v)
	case schemas.S5CipherEngineTypeGCP:
		return c.getCipherEngineGCP(v)
	case schemas.S5CipherEngineTypePGP:
		return c.getCipherEnginePGP(v)
	case schemas.S5CipherEngineTypeVault:
		return c.getCipherEngineVault(v)
	default:
		return nil, fmt.Errorf("Engine %s is not implemented yet", *cipherEngineType)
	}
}

func (c *Client) getCipherEngineAES(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineAES != nil && v.CipherEngineAES.Key != nil {
		return cipher.NewAES(*v.CipherEngineAES.Key)
	}

	if c.CipherEngineAES != nil && c.CipherEngineAES.Key != nil {
		return cipher.NewAES(*c.CipherEngineAES.Key)
	}

	return cipher.NewAES(os.Getenv("S5_AES_KEY"))
}

func (c *Client) getCipherEngineAWS(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineAWS != nil && v.CipherEngineAWS.KmsKeyArn != nil {
		return cipher.NewAWS(*v.CipherEngineAWS.KmsKeyArn)
	}

	if c.CipherEngineAWS != nil && c.CipherEngineAWS.KmsKeyArn != nil {
		return cipher.NewAWS(*c.CipherEngineAWS.KmsKeyArn)
	}

	return cipher.NewAWS(os.Getenv("S5_AWS_KMS_KEY_ARN"))
}

func (c *Client) getCipherEngineGCP(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineGCP != nil && v.CipherEngineGCP.KmsKeyName != nil {
		return cipher.NewAWS(*v.CipherEngineGCP.KmsKeyName)
	}

	if c.CipherEngineGCP != nil && c.CipherEngineGCP.KmsKeyName != nil {
		return cipher.NewAWS(*c.CipherEngineGCP.KmsKeyName)
	}

	return cipher.NewGCP(os.Getenv("S5_GCP_KMS_KEY_NAME"))
}

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

	return cipher.NewPGP(publicKeyPath, privateKeyPath)
}

func (c *Client) getCipherEngineVault(v *schemas.S5) (cipher.Engine, error) {
	if v.CipherEngineVault != nil && v.CipherEngineVault.TransitKey != nil {
		return cipher.NewAWS(*v.CipherEngineVault.TransitKey)
	}

	if c.CipherEngineVault != nil && c.CipherEngineVault.TransitKey != nil {
		return cipher.NewAWS(*c.CipherEngineVault.TransitKey)
	}

	return cipher.NewVault(os.Getenv("S5_VAULT_TRANSIT_KEY"))
}
