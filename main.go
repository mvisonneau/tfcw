package main

import (
	"os"
	"time"

	"github.com/mvisonneau/tfcw/cli"
	log "github.com/sirupsen/logrus"
)

var version = ""

func main() {
	if err := cli.Init(&version, time.Now()).Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
