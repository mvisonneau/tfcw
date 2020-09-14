package main

import (
	"os"

	"github.com/mvisonneau/tfcw/cli"
)

var version = ""

func main() {
	cli.Run(version, os.Args)
}
