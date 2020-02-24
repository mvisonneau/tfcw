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

// Client aggregates provider clients
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

// Config is a subset of schemas.Config with a few more runtime
// related values
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

// NewClient instantiate a Client from a provider Config
func NewClient(cfg *Config) (c *Client, err error) {
	vaultClient, err := getVaultClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting vault client: %s", err)
	}

	tfeClient, err := getTFEClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting terraform cloud client: %s", err)
	}

	c = &Client{
		Vault:              vaultClient,
		S5:                 getS5Client(cfg),
		Env:                &providerEnv.Client{},
		TFE:                tfeClient,
		Context:            context.Background(),
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

func getVaultClient(cfg *Config) (c *providerVault.Client, err error) {
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

		c, err = providerVault.GetClient(vaultAddress, vaultToken)
	}
	return
}

func getS5Client(cfg *Config) (c *providerS5.Client) {
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

func getTFEClient(cfg *Config) (c *tfe.Client, err error) {
	if !cfg.Runtime.TFE.Disabled {
		c, err = tfe.NewClient(&tfe.Config{
			Address: cfg.Runtime.TFE.Address,
			Token:   cfg.Runtime.TFE.Token,
		})
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
