package vault

import (
	"fmt"
	"io/ioutil"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"

	"github.com/mvisonneau/tfcs/lib/schemas"
)

type Client struct {
	*vault.Client
}

// GetClient : Get a Vault client using Vault official params
func GetClient(address, token string) (*Client, error) {
	c, err := vault.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating Vault client: %s", err.Error())
	}

	if len(address) > 0 {
		c.SetAddress(address)
	} else if len(os.Getenv("VAULT_ADDR")) > 0 {
		c.SetAddress(os.Getenv("VAULT_ADDR"))
	} else {
		return nil, fmt.Errorf("Vault address is not defined")
	}

	if len(token) > 0 {
		c.SetToken(token)
	} else {
		token := os.Getenv("VAULT_TOKEN")
		if len(token) == 0 {
			home, _ := homedir.Dir()
			f, err := ioutil.ReadFile(home + "/.vault-token")
			if err != nil {
				return nil, fmt.Errorf("Vault token is not defined (VAULT_TOKEN or ~/.vault-token)")
			}

			token = string(f)
		}

		c.SetToken(token)
	}

	return &Client{c}, nil
}

func (c *Client) GetValues(v *schemas.Vault) (results map[string]string, err error) {
	results = make(map[string]string)
	if v != nil && v.Path != nil {
		//log.Infof("Using Vault for variable '%s'", v.Name)
		var secret *vault.Secret

		if v.Method != nil {
			switch *v.Method {
			case "read":
				secret, err = c.Logical().Read(*v.Path)
			case "write":
				params := map[string]interface{}{}
				if v.Params != nil {
					for k, v := range *v.Params {
						params[k] = v
					}
				}
				secret, err = c.Logical().Write(*v.Path, params)
			default:
				return results, fmt.Errorf("unsupported method '%s'", *v.Method)
			}
		} else {
			secret, err = c.Logical().Read(*v.Path)
		}

		if err != nil {
			return results, fmt.Errorf("vault error : %s", err)
		}

		if len(secret.Data) == 0 {
			return results, fmt.Errorf("no results/keys returned for secret : %s", *v.Path)
		}

		for k, v := range secret.Data {
			results[k] = v.(string)
		}

		return
	}

	// if v.S5 != nil && v.S5.Value != nil {
	// 	log.Infof("Using S5 for variable '%s'", v.Name)

	// 	c, s5Err := getS5CipherEngine(v.S5, c.Defaults.S5)
	// 	if s5Err != nil {
	// 		return results, s5Err
	// 	}

	// 	input, s5Err := cipher.ParseInput(*v.S5.Value)
	// 	if s5Err != nil {
	// 		return results, s5Err
	// 	}

	// 	o, s5Err := c.Decipher(input)
	// 	if s5Err != nil {
	// 		return results, s5Err
	// 	}

	// 	results["s5"] = o
	// 	return
	// }

	return results, fmt.Errorf("No provider defined for variable '%v'", v)
}
