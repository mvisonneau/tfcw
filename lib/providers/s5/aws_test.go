package s5

import (
	"os"
	"testing"

	"github.com/mvisonneau/s5/cipher"
	cipherAWS "github.com/mvisonneau/s5/cipher/aws"
	"github.com/mvisonneau/tfcw/lib/schemas"
	"github.com/stretchr/testify/assert"
)

const (
	testAWSKMSKeyArn string = "arn:aws:kms:*:111111111111:key/mykey"
)

func TestGetCipherEngineAWS(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypeAWS
	kmsKeyArn := testAWSKMSKeyArn

	// expected engine
	expectedEngine, err := cipher.NewAWSClient(kmsKeyArn)
	assert.Nil(t, err)

	// all defined in client, empty variable config (default settings)
	v := &schemas.S5{}
	c := &Client{
		CipherEngineType: &cipherEngineType,
		CipherEngineAWS: &schemas.S5CipherEngineAWS{
			KmsKeyArn: &kmsKeyArn,
		},
	}

	cipherEngine, err := c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherAWS.Client).Config)

	// all defined in variable, empty client config
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineAWS: &schemas.S5CipherEngineAWS{
			KmsKeyArn: &kmsKeyArn,
		},
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherAWS.Client).Config)

	// key defined in environment variable
	os.Setenv("S5_AWS_KMS_KEY_ARN", testAWSKMSKeyArn)
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherAWS.Client).Config)

	// other engine & key defined in client, overridden in variable
	otherCipherEngineType := schemas.S5CipherEngineTypeVault
	otherKmsKeyArn := "arn:aws:kms:*:111111111111:key/myotherkey"

	c = &Client{
		CipherEngineType: &otherCipherEngineType,
		CipherEngineAWS: &schemas.S5CipherEngineAWS{
			KmsKeyArn: &otherKmsKeyArn,
		},
	}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineAWS: &schemas.S5CipherEngineAWS{
			KmsKeyArn: &kmsKeyArn,
		},
	}

	cipherEngine, err = c.getCipherEngine(v)
	assert.Nil(t, err)
	assert.Equal(t, expectedEngine.Config, cipherEngine.(*cipherAWS.Client).Config)
}
