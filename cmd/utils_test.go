package cmd

import (
	"fmt"
	"testing"

	"github.com/mvisonneau/go-helpers/test"
)

func TestExit(t *testing.T) {
	err := exit(fmt.Errorf("test"), 20)
	test.Expect(t, err.Error(), "")
	test.Expect(t, err.ExitCode(), 20)
}
