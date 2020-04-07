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

	assert.Equal(t, false, isVaultClientRequired(cfg))

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
	assert.Equal(t, false, isVaultClientRequired(cfg))

	path := "foo"
	cfg.EnvironmentVariables = schemas.Variables{
		&schemas.Variable{
			Vault: &schemas.Vault{
				Path: &path,
			},
		},
	}
	assert.Equal(t, true, isVaultClientRequired(cfg))
}

func TestNewClient(t *testing.T) {
	cfg := &schemas.Config{
		Runtime: schemas.Runtime{},
	}

	// We need to set the TFC token otherwise the client won't initiate correctly
	cfg.Runtime.TFC.Token = "_"

	c, err := NewClient(cfg)
	assert.Equal(t, nil, err)
	assert.IsType(t, &providerVault.Client{}, c.Vault)
	assert.IsType(t, &providerS5.Client{}, c.S5)
	assert.IsType(t, &providerEnv.Client{}, c.Env)
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
	assert.Equal(t, nil, err)
	assert.Equal(t, fooString, c.Address())
	assert.Equal(t, fooString, c.Token())
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
	assert.Equal(t, cipherEngineType, *c.CipherEngineType)
	assert.Equal(t, cipherEngineAES, *c.CipherEngineAES)
	assert.Equal(t, cipherEngineAWS, *c.CipherEngineAWS)
	assert.Equal(t, cipherEngineGCP, *c.CipherEngineGCP)
	assert.Equal(t, cipherEnginePGP, *c.CipherEnginePGP)
	assert.Equal(t, cipherEngineVault, *c.CipherEngineVault)
}
