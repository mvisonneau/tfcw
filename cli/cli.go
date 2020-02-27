package cli

import (
	"os"
	"time"

	"github.com/mvisonneau/tfcw/cmd"
	"github.com/urfave/cli"
)

// Run handles the instanciation of the CLI application
func Run(version string) {
	NewApp(version, time.Now()).Run(os.Args)
}

// NewApp configures the CLI application
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
			Name:   "tfc-address",
			EnvVar: "TFCW_TFC_ADDRESS",
			Usage:  "`address` to access Terraform Cloud API",
			Value:  "https://app.terraform.io",
		},
		cli.StringFlag{
			Name:   "tfc-token,t",
			EnvVar: "TFCW_TFC_TOKEN",
			Usage:  "`token` to access Terraform Cloud API",
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
					Flags:  []cli.Flag{dryRun},
				},
				{
					Name:   "local",
					Usage:  "render the variables locally, on disk",
					Action: cmd.ExecWrapper(cmd.Render),
				},
			},
		},
		{
			Name:  "run",
			Usage: "manipulate runs",
			Subcommands: []cli.Command{
				{
					Name:   "create",
					Usage:  "create a run on TFC",
					Action: cmd.ExecWrapper(cmd.RunCreate),
					Flags:  runCreate,
				},
				{
					Name:   "approve",
					Usage:  "approve a run given its 'ID'",
					Action: cmd.ExecWrapper(cmd.RunApprove),
					Flags:  []cli.Flag{currentRun},
				},
				{
					Name:   "discard",
					Usage:  "discard a run given its 'ID'",
					Action: cmd.ExecWrapper(cmd.RunDiscard),
					Flags:  []cli.Flag{currentRun},
				},
			},
		},
		{
			Name:  "workspace",
			Usage: "manipulate the workspace",
			Subcommands: []cli.Command{
				{
					Name:   "status",
					Usage:  "return the status of the workspace",
					Action: cmd.ExecWrapper(cmd.WorkspaceStatus),
				},
				{
					Name:   "current-run-id",
					Usage:  "return the id of the current run",
					Action: cmd.ExecWrapper(cmd.WorkspaceCurrentRunID),
				},
			},
		},
	}

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
