package client

import (
	"testing"

	providerEnv "github.com/mvisonneau/tfcw/lib/providers/env"
	providerS5 "github.com/mvisonneau/tfcw/lib/providers/s5"
	providerVault "github.com/mvisonneau/tfcw/lib/providers/vault"
	"github.com/mvisonneau/tfcw/lib/schemas"
	"github.com/stretchr/testify/assert"
)

func TestIsVaultClientRequired(t *testing.T) {
	// Validate Vault client is not required if config is empty
	cfg := &schemas.Config{}

	assert.Equal(t, isVaultClientRequired(cfg), false)

	// Validate Vault client is not required if config contains other variables with
	// different providers is empty
	s5CipherEngineType := schemas.S5CipherEngineTypeAES
	cfg.EnvironmentVariables = schemas.Variables{
		&schemas.Variable{
			S5: &schemas.S5{
				CipherEngineType: &s5CipherEngineType,
			},
		},
	}
	assert.Equal(t, isVaultClientRequired(cfg), false)

	path := "foo"
	cfg.EnvironmentVariables = schemas.Variables{
		&schemas.Variable{
			Vault: &schemas.Vault{
				Path: &path,
			},
		},
	}
	assert.Equal(t, isVaultClientRequired(cfg), true)
}

func TestNewClient(t *testing.T) {
	cfg := &schemas.Config{}

	// We will have to figure out how to test TFE init but for now lets disable it
	cfg.Runtime.TFC.Disabled = true

	c, err := NewClient(cfg)
	assert.Equal(t, err, nil)
	assert.IsType(t, c.Vault, &providerVault.Client{})
	assert.IsType(t, c.S5, &providerS5.Client{})
	assert.IsType(t, c.Env, &providerEnv.Client{})
}

func TestGetVaultClient(t *testing.T) {
	fooString := "foo"
	cfg := &schemas.Config{
		Defaults: &schemas.Defaults{
			Vault: &schemas.Vault{
				Address: &fooString,
				Token:   &fooString,
			},
		},
		EnvironmentVariables: schemas.Variables{
			&schemas.Variable{
				Vault: &schemas.Vault{
					Path: &fooString,
				},
			},
		},
	}

	c, err := getVaultClient(cfg)
	assert.Equal(t, err, nil)
	assert.Equal(t, c.Address(), fooString)
	assert.Equal(t, c.Token(), fooString)
}

func TestGetS5Client(t *testing.T) {
	cipherEngineType := schemas.S5CipherEngineTypeAES
	cipherEngineAES := schemas.S5CipherEngineAES{}
	cipherEngineAWS := schemas.S5CipherEngineAWS{}
	cipherEngineGCP := schemas.S5CipherEngineGCP{}
	cipherEnginePGP := schemas.S5CipherEnginePGP{}
	cipherEngineVault := schemas.S5CipherEngineVault{}

	cfg := &schemas.Config{
		Defaults: &schemas.Defaults{
			S5: &schemas.S5{
				CipherEngineType:  &cipherEngineType,
				CipherEngineAES:   &cipherEngineAES,
				CipherEngineAWS:   &cipherEngineAWS,
				CipherEngineGCP:   &cipherEngineGCP,
				CipherEnginePGP:   &cipherEnginePGP,
				CipherEngineVault: &cipherEngineVault,
			},
		},
	}

	c := getS5Client(cfg)
	assert.Equal(t, *c.CipherEngineType, cipherEngineType)
	assert.Equal(t, *c.CipherEngineAES, cipherEngineAES)
	assert.Equal(t, *c.CipherEngineAWS, cipherEngineAWS)
	assert.Equal(t, *c.CipherEngineGCP, cipherEngineGCP)
	assert.Equal(t, *c.CipherEnginePGP, cipherEnginePGP)
	assert.Equal(t, *c.CipherEngineVault, cipherEngineVault)
}
