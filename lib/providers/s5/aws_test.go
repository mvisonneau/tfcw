package s5

import (
	"os"
	"testing"

	"github.com/mvisonneau/go-helpers/test"
	"github.com/mvisonneau/s5/cipher"
	cipherAWS "github.com/mvisonneau/s5/cipher/aws"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

const (
	testAWSKMSKeyArn string = "arn:aws:kms:*:111111111111:key/mykey"
)

func TestGetCipherEngineAWS(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypeAWS
	kmsKeyArn := testAWSKMSKeyArn

	// expected engine
	expectedEngine, err := cipher.NewAWSClient(kmsKeyArn)
	test.Expect(t, err, nil)

	// all defined in client, empty variable config (default settings)
	v := &schemas.S5{}
	c := &Client{
		CipherEngineType: &cipherEngineType,
		CipherEngineAWS: &schemas.S5CipherEngineAWS{
			KmsKeyArn: &kmsKeyArn,
		},
	}

	cipherEngine, err := c.getCipherEngine(v)
	test.Expect(t, err, nil)
	test.Expect(t, cipherEngine.(*cipherAWS.Client).Config, expectedEngine.Config)

	// all defined in variable, empty client config
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineAWS: &schemas.S5CipherEngineAWS{
			KmsKeyArn: &kmsKeyArn,
		},
	}

	cipherEngine, err = c.getCipherEngine(v)
	test.Expect(t, err, nil)
	test.Expect(t, cipherEngine.(*cipherAWS.Client).Config, expectedEngine.Config)

	// key defined in environment variable
	os.Setenv("S5_AWS_KMS_KEY_ARN", testAWSKMSKeyArn)
	c = &Client{}
	v = &schemas.S5{
		CipherEngineType: &cipherEngineType,
	}

	cipherEngine, err = c.getCipherEngine(v)
	test.Expect(t, err, nil)
	test.Expect(t, cipherEngine.(*cipherAWS.Client).Config, expectedEngine.Config)

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
	test.Expect(t, err, nil)
	test.Expect(t, cipherEngine.(*cipherAWS.Client).Config, expectedEngine.Config)
}
