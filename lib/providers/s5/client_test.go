package s5

import (
	"fmt"
	"testing"

	"github.com/mvisonneau/go-helpers/test"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func TestGetValueValid(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypeAES
	cipheredValue := "{{s5:NmRhN2I1YTFhNGE4ZjUzNzI5ZTNlMjk4YzQ3NWQzMDRiYmRkYjA6OTAzN2E3OGQ0YTFmY2U0ZDRmZmExYmU2}}"
	key := testAESKey

	c := &Client{}
	v := &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineAES: &schemas.S5CipherEngineAES{
			Key: &key,
		},
		Value: &cipheredValue,
	}

	value, err := c.GetValue(v)
	test.Expect(t, err, nil)
	test.Expect(t, value, "foo")
}

func TestGetValueInvalidCipherEngine(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineType("foo")
	c := &Client{}
	v := &schemas.S5{
		CipherEngineType: &cipherEngineType,
	}

	value, err := c.GetValue(v)
	test.Expect(t, err, fmt.Errorf("s5 error whilst getting cipher engine: engine 'foo' is not implemented yet"))
	test.Expect(t, value, "")
}

func TestGetValueInvalidInput(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypeAES
	invalidCipheredValue := "{{foo}}"
	key := testAESKey

	c := &Client{}
	v := &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineAES: &schemas.S5CipherEngineAES{
			Key: &key,
		},
		Value: &invalidCipheredValue,
	}

	value, err := c.GetValue(v)
	test.Expect(t, err, fmt.Errorf("s5 error whilst parsing input: Invalid string format, should be '{{s5:*}}'"))
	test.Expect(t, value, "")
}

func TestGetValueInvalidDecipher(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypeAES
	invalidCipheredValue := "{{s5:foo}}"
	key := testAESKey

	c := &Client{}
	v := &schemas.S5{
		CipherEngineType: &cipherEngineType,
		CipherEngineAES: &schemas.S5CipherEngineAES{
			Key: &key,
		},
		Value: &invalidCipheredValue,
	}

	value, err := c.GetValue(v)
	test.Expect(t, err, fmt.Errorf("s5 error whilst deciphering: base64decode error : illegal base64 data at input byte 0 - value : foo"))
	test.Expect(t, value, "")
}

func TestGetCipherEngineUndefined(t *testing.T) {
	c := &Client{}
	v := &schemas.S5{}

	cipherEngine, err := c.getCipherEngine(v)
	test.Expect(t, err, fmt.Errorf("you need to specify a S5 cipher engine"))
	test.Expect(t, cipherEngine, nil)
}

func TestGetCipherEngineInvalid(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineType("foo")
	c := &Client{}
	v := &schemas.S5{
		CipherEngineType: &cipherEngineType,
	}

	cipherEngine, err := c.getCipherEngine(v)
	test.Expect(t, err, fmt.Errorf("engine 'foo' is not implemented yet"))
	test.Expect(t, cipherEngine, nil)
}
