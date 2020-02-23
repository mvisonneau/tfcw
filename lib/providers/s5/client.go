package s5

import (
	"fmt"

	"github.com/mvisonneau/s5/cipher"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

// Client is here to support provider related functions
type Client struct {
	CipherEngineType  *schemas.S5CipherEngineType
	CipherEngineAES   *schemas.S5CipherEngineAES
	CipherEngineAWS   *schemas.S5CipherEngineAWS
	CipherEngineGCP   *schemas.S5CipherEngineGCP
	CipherEnginePGP   *schemas.S5CipherEnginePGP
	CipherEngineVault *schemas.S5CipherEngineVault
}

// GetValue returns a deciphered value from S5
func (c *Client) GetValue(v *schemas.S5) (string, error) {
	variableCipher, err := c.getCipherEngine(v)
	if err != nil {
		return "", fmt.Errorf("s5 error whilst getting cipher engine: %s", err.Error())
	}

	parsedValue, err := cipher.ParseInput(*v.Value)
	if err != nil {
		return "", fmt.Errorf("s5 error whilst parsing input: %s", err.Error())
	}

	value, err := variableCipher.Decipher(parsedValue)
	if err != nil {
		return "", fmt.Errorf("s5 error whilst deciphering: %s", err.Error())
	}

	return value, nil
}

func (c *Client) getCipherEngine(v *schemas.S5) (cipher.Engine, error) {
	var cipherEngineType *schemas.S5CipherEngineType
	if v.CipherEngineType != nil {
		cipherEngineType = v.CipherEngineType
	} else if c.CipherEngineType != nil {
		cipherEngineType = c.CipherEngineType
	} else {
		return nil, fmt.Errorf("you need to specify a S5 cipher engine")
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
		return nil, fmt.Errorf("engine %s is not implemented yet", *cipherEngineType)
	}
}
