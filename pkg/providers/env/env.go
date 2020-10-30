package env

import (
	"os"

	"github.com/mvisonneau/tfcw/pkg/schemas"
)

// Client is a basic struct in order to support provider
// related functions
type Client struct{}

// GetValue returns a value from an environment variable
func (c *Client) GetValue(e *schemas.Env) string {
	return os.Getenv(e.Variable)
}
