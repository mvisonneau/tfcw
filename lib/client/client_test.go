package client

import (
	"testing"

	"github.com/mvisonneau/tfcw/lib/schemas"
)

func TestIsVaultClientRequired(t *testing.T) {
	// Validate Vault client is not required if config is empty
	cfg := &Config{
		Config: &schemas.Config{},
	}

	if cfg.isVaultClientRequired() {
		t.Fatalf("Vault client should not be required")
	}

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
	if cfg.isVaultClientRequired() {
		t.Fatalf("Vault client should not be required")
	}

	path := "foo"
	cfg.EnvironmentVariables = schemas.Variables{
		&schemas.Variable{
			Vault: &schemas.Vault{
				Path: &path,
			},
		},
	}
	if !cfg.isVaultClientRequired() {
		t.Fatalf("Vault client should be required")
	}
}
