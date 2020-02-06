package cli

import (
	"time"

	"github.com/mvisonneau/tfcs/cmd"
	"github.com/urfave/cli"
)

// Init : Generates CLI configuration for the application
func Init(version *string, start time.Time) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "tfcs"
	app.Version = *version
	app.Usage = "Terraform Cloud wrapper which can be used to manage variables dynamically"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config-path,c",
			EnvVar: "TFCS_CONFIG_PATH",
			Usage:  "`path` to a readable configuration file (.hcl or .json)",
			Value:  "./tfcs.hcl",
		},
		cli.StringFlag{
			Name:   "log-level",
			EnvVar: "TFCS_LOG_LEVEL",
			Usage:  "log `level` (debug,info,warn,fatal,panic)",
			Value:  "info",
		},
		cli.StringFlag{
			Name:   "log-format",
			EnvVar: "TFCS_LOG_FORMAT",
			Usage:  "log `format` (json,text)",
			Value:  "text",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "validate",
			Usage:  "validate the config and access to all the providers included",
			Action: cmd.Validate,
		},
	}

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
