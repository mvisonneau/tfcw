package main

import (
	"os"

	"github.com/mvisonneau/tfcw/internal/cli"
)

var version = ""

func main() {
	cli.Run(version, os.Args)
}
