package env

import (
	"os"
	"testing"

	"github.com/mvisonneau/go-helpers/test"
	"github.com/mvisonneau/tfcw/lib/schemas"
)

func TestGetValue(t *testing.T) {
	os.Setenv("TEST_ENV", "foo")

	c := &Client{}
	e := &schemas.Env{
		Variable: "TEST_ENV",
	}

	test.Expect(t, c.GetValue(e), "foo")
}
