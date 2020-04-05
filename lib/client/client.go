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

	tfc "github.com/hashicorp/go-tfe"
)

// Client aggregates provider clients
type Client struct {
	Vault                   *providerVault.Client
	S5                      *providerS5.Client
	Env                     *providerEnv.Client
	TFC                     *tfc.Client
	Context                 context.Context
	ProcessedVariablesMutex sync.Mutex
	ProcessedVariables      map[string]schemas.VariableKind
	Backoff                 *backoff.Backoff
}

// NewClient instantiate a Client from a provider Config
func NewClient(cfg *schemas.Config) (c *Client, err error) {
	vaultClient, err := getVaultClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting vault client: %s", err)
	}

	tfcClient, err := getTFCClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting terraform cloud client: %s", err)
	}

	c = &Client{
		Vault:              vaultClient,
		S5:                 getS5Client(cfg),
		Env:                &providerEnv.Client{},
		TFC:                tfcClient,
		Context:            context.Background(),
		ProcessedVariables: map[string]schemas.VariableKind{},
		Backoff: &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    20 * time.Second,
			Factor: 1.5,
			Jitter: false,
		},
	}

	return
}

func getVaultClient(cfg *schemas.Config) (c *providerVault.Client, err error) {
	if isVaultClientRequired(cfg) {
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

		c, err = providerVault.GetClient(vaultAddress, vaultToken)
	}
	return
}

func getS5Client(cfg *schemas.Config) (c *providerS5.Client) {
	c = &providerS5.Client{}
	if cfg.Defaults != nil && cfg.Defaults.S5 != nil {
		if cfg.Defaults.S5.CipherEngineType != nil {
			c.CipherEngineType = cfg.Defaults.S5.CipherEngineType
		}
		if cfg.Defaults.S5.CipherEngineAES != nil {
			c.CipherEngineAES = cfg.Defaults.S5.CipherEngineAES
		}
		if cfg.Defaults.S5.CipherEngineAWS != nil {
			c.CipherEngineAWS = cfg.Defaults.S5.CipherEngineAWS
		}
		if cfg.Defaults.S5.CipherEngineGCP != nil {
			c.CipherEngineGCP = cfg.Defaults.S5.CipherEngineGCP
		}
		if cfg.Defaults.S5.CipherEnginePGP != nil {
			c.CipherEnginePGP = cfg.Defaults.S5.CipherEnginePGP
		}
		if cfg.Defaults.S5.CipherEngineVault != nil {
			c.CipherEngineVault = cfg.Defaults.S5.CipherEngineVault
		}
	}
	return
}

func getTFCClient(cfg *schemas.Config) (c *tfc.Client, err error) {
	if !cfg.Runtime.TFC.Disabled {
		c, err = tfc.NewClient(&tfc.Config{
			Address: cfg.Runtime.TFC.Address,
			Token:   cfg.Runtime.TFC.Token,
		})
	}
	return
}

func isVaultClientRequired(cfg *schemas.Config) bool {
	for _, v := range append(cfg.TerraformVariables, cfg.EnvironmentVariables...) {
		if v.Vault != nil {
			return true
		}
	}
	return false
}
