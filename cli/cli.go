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
	app.Usage = "Terraform Cloud Wrapper"
	app.EnableBashCompletion = true

	app.Flags = cli.FlagsByName{
		cli.StringFlag{
			Name:   "address,a",
			EnvVar: "TFCW_ADDRESS",
			Usage:  "`address` to access Terraform Cloud API",
		},
		cli.StringFlag{
			Name:   "config-file,c",
			EnvVar: "TFCW_CONFIG_FILE",
			Usage:  "`path` of a readable TFCW configuration file (.hcl or .json)",
			Value:  "<working-dir>/tfcw.hcl",
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
		cli.StringFlag{
			Name:   "organization,o",
			EnvVar: "TFCW_ORGANIZATION",
			Usage:  "`organization` to use on Terraform Cloud API",
		},
		cli.StringFlag{
			Name:   "token,t",
			EnvVar: "TFCW_TOKEN",
			Usage:  "`token` to access Terraform Cloud API",
		},
		cli.StringFlag{
			Name:   "working-dir,d",
			EnvVar: "TFCW_WORKING_DIR",
			Usage:  "`path` of the directory containing your Terraform files",
			Value:  ".",
		},
		cli.StringFlag{
			Name:   "workspace,w",
			EnvVar: "TFCW_WORKSPACE",
			Usage:  "`workspace` to use on Terraform Cloud API",
		},
	}

	app.Commands = cli.CommandsByName{
		{
			Name:   "render",
			Usage:  "render variables values",
			Action: cmd.ExecWrapper(cmd.Render),
			Flags:  cli.FlagsByName{renderType, ignoreTTLs, dryRun},
		},
		{
			Name:  "run",
			Usage: "manipulate runs",
			Subcommands: cli.Commands{
				{
					Name:   "approve",
					Usage:  "approve a run given its 'ID'",
					Action: cmd.ExecWrapper(cmd.RunApprove),
					Flags:  cli.FlagsByName{currentRun, message},
				},
				{
					Name:   "create",
					Usage:  "create a run on TFC",
					Action: cmd.ExecWrapper(cmd.RunCreate),
					Flags:  append(runCreate, message, renderType, ignoreTTLs),
				},
				{
					Name:   "discard",
					Usage:  "discard a run given its 'ID'",
					Action: cmd.ExecWrapper(cmd.RunDiscard),
					Flags:  cli.FlagsByName{currentRun, message},
				},
				{
					Name:   "current-id",
					Usage:  "return the id of the current run",
					Action: cmd.ExecWrapper(cmd.RunCurrentID),
				},
			},
		},
		{
			Name:      "workspace",
			ShortName: "ws",
			Usage:     "manipulate the workspace",
			Subcommands: cli.Commands{
				{
					Name:      "configure",
					ShortName: "cfg",
					Usage:     "apply defined configuration to the workspace",
					Action:    cmd.ExecWrapper(cmd.WorkspaceConfigure),
					Flags:     cli.FlagsByName{dryRun},
				},
				{
					Name:   "status",
					Usage:  "return the status of the workspace",
					Action: cmd.ExecWrapper(cmd.WorkspaceStatus),
				},
				{
					Name:  "operations",
					Usage: "manages the operations value of the workspace",
					Subcommands: cli.Commands{
						{
							Name:   "enable",
							Usage:  "enable remote operations on the workspace",
							Action: cmd.ExecWrapper(cmd.WorkspaceEnableOperations),
						},
						{
							Name:   "disable",
							Usage:  "disable remote operations on the workspace",
							Action: cmd.ExecWrapper(cmd.WorkspaceDisableOperations),
						},
					},
				},
			},
		},
	}

	app.Metadata = map[string]interface{}{
		"startTime": start,
	}

	return
}
