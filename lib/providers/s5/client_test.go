package s5

import (
	"fmt"
	"testing"

	"github.com/mvisonneau/go-helpers/test"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func TestGetCipherEngineUndefined(t *testing.T) {
	c := &Client{}
	v := &schemas.S5{}

	// Empty cipher engine
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

	// Empty cipher engine
	cipherEngine, err := c.getCipherEngine(v)
	test.Expect(t, err, fmt.Errorf("engine 'foo' is not implemented yet"))
	test.Expect(t, cipherEngine, nil)
}
