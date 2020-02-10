package env

import (
	"os"

	"github.com/mvisonneau/tfcw/lib/schemas"
)

type Client struct{}

func (c *Client) GetValue(e *schemas.Env) string {
	return os.Getenv(e.Variable)
}
