package cli

import (
	"log"
	"os"
	"time"

	"github.com/mvisonneau/tfcw/cmd"
	"github.com/urfave/cli"
)

func Run(version string) {
	if err := NewApp(version, time.Now()).Run(os.Args); err != nil {
		log.Fatal("foo")
	}
}

func NewApp(version string, start time.Time) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "tfcw"
	app.Version = version
	app.Usage = "Terraform Cloud wrapper which can be used to manage variables dynamically"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config-path,c",
			EnvVar: "TFCW_CONFIG_PATH",
			Usage:  "`path` to a readable configuration file (.hcl or .json)",
			Value:  "./tfcw.hcl",
		},
		cli.StringFlag{
			Name:   "log-level",
			EnvVar: "TFCW_LOG_LEVEL",
			Usage:  "log `level` (debug,info,warn,fatal,panic)",
			Value:  "info",
		},
		cli.StringFlag{
			Name:   "log-format",
			EnvVar: "TFCW_LOG_FORMAT",
			Usage:  "log `format` (json,text)",
			Value:  "text",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "render",
			Usage: "render the variables",
			Subcommands: []cli.Command{
				{
					Name:   "tfc",
					Usage:  "update the variables on TFC directly",
					Action: cmd.ExecWrapper(cmd.Render),
					Flags:  append(tfc, dryRun),
				},
				{
					Name:   "local",
					Usage:  "render the variables locally, on disk",
					Action: cmd.ExecWrapper(cmd.Render),
				},
			},
		},
		{
			Name:   "plan",
			Usage:  "plans the config",
			Action: cmd.ExecWrapper(cmd.TFERun),
			Flags:  append(tfc, tf...),
		},
		{
			Name:   "apply",
			Usage:  "plans and applies the config",
			Action: cmd.ExecWrapper(cmd.TFERun),
			Flags:  append(tfc, tf...),
		},
	}

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
