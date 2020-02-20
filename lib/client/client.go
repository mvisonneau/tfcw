package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	providerEnv "github.com/mvisonneau/tfcw/lib/providers/env"
	providerS5 "github.com/mvisonneau/tfcw/lib/providers/s5"
	providerVault "github.com/mvisonneau/tfcw/lib/providers/vault"
	"github.com/mvisonneau/tfcw/lib/schemas"

	tfe "github.com/hashicorp/go-tfe"
)

type Client struct {
	Vault                   *providerVault.Client
	S5                      *providerS5.Client
	Env                     *providerEnv.Client
	TFE                     *tfe.Client
	Context                 context.Context
	ProcessedVariablesMutex sync.Mutex
	ProcessedVariables      map[string]schemas.VariableKind
	Backoff                 *backoff.Backoff
}

type Config struct {
	*schemas.Config
	Runtime struct {
		TFE struct {
			Disabled bool
			Address  string
			Token    string
		}
	}
}

func NewClient(cfg *Config) (c *Client, err error) {
	vaultClient := &providerVault.Client{}
	if cfg.isVaultClientRequired() {
		// Initializing Vault client with default values
		var vaultAddress, vaultToken string
		if cfg.Defaults != nil {
			if cfg.Defaults.Vault != nil {
				if cfg.Defaults.Vault.Address != nil {
					vaultAddress = *cfg.Defaults.Vault.Address
				}

				if cfg.Defaults.Vault.Address != nil {
					vaultToken = *cfg.Defaults.Vault.Token
				}
			}
		}

		vaultClient, err = providerVault.GetClient(vaultAddress, vaultToken)
		if err != nil {
			return nil, fmt.Errorf("vault: %s", err)
		}
	}

	// Initializing S5 client with default values
	s5Client := &providerS5.Client{}
	if cfg.Defaults != nil {
		if cfg.Defaults.S5 != nil {
			if cfg.Defaults.S5.CipherEngineType != nil {
				s5Client.CipherEngineType = cfg.Defaults.S5.CipherEngineType
			}
			if cfg.Defaults.S5.CipherEngineAES != nil {
				s5Client.CipherEngineAES = cfg.Defaults.S5.CipherEngineAES
			}
			if cfg.Defaults.S5.CipherEngineAWS != nil {
				s5Client.CipherEngineAWS = cfg.Defaults.S5.CipherEngineAWS
			}
			if cfg.Defaults.S5.CipherEngineGCP != nil {
				s5Client.CipherEngineGCP = cfg.Defaults.S5.CipherEngineGCP
			}
			if cfg.Defaults.S5.CipherEnginePGP != nil {
				s5Client.CipherEnginePGP = cfg.Defaults.S5.CipherEnginePGP
			}
			if cfg.Defaults.S5.CipherEngineVault != nil {
				s5Client.CipherEngineVault = cfg.Defaults.S5.CipherEngineVault
			}
		}
	}

	// Configure TFE client
	tfeClient := &tfe.Client{}
	if !cfg.Runtime.TFE.Disabled {
		tfeClient, err = tfe.NewClient(&tfe.Config{
			Address: cfg.Runtime.TFE.Address,
			Token:   cfg.Runtime.TFE.Token,
		})

		if err != nil {
			return nil, fmt.Errorf("terraform cloud: %s", err)
		}
	}

	// Create a context
	ctx := context.Background()

	c = &Client{
		Vault:              vaultClient,
		S5:                 s5Client,
		Env:                &providerEnv.Client{},
		TFE:                tfeClient,
		Context:            ctx,
		ProcessedVariables: map[string]schemas.VariableKind{},
		Backoff: &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    15 * time.Second,
			Factor: 2,
			Jitter: false,
		},
	}

	return
}

func (cfg *Config) isVaultClientRequired() bool {
	for _, v := range append(cfg.TerraformVariables, cfg.EnvironmentVariables...) {
		if v.Vault != nil {
			return true
		}
	}
	return false
}
