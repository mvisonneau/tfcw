package env

import (
	"os"
	"testing"

	"github.com/mvisonneau/go-helpers/assert"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func TestGetValue(t *testing.T) {
	os.Setenv("TEST_ENV", "foo")

	c := &Client{}
	e := &schemas.Env{
		Variable: "TEST_ENV",
	}

	assert.Equal(t, c.GetValue(e), "foo")
}
