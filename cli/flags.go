package cli

import (
	"github.com/urfave/cli"
)

var runCreate = []cli.Flag{
	cli.StringFlag{
		Name:   "tf-config-folder,f",
		EnvVar: "TFCW_TF_CONFIG_FOLDER",
		Usage:  "`path` to the Terraform configuration folder",
		Value:  ".",
	},
	cli.BoolFlag{
		Name:  "auto-discard",
		Usage: "will automatically discard the run once planned",
	},
	cli.BoolFlag{
		Name:  "auto-approve",
		Usage: "automatically approve the run once planned",
	},
	cli.BoolFlag{
		Name:  "no-prompt",
		Usage: "will not prompt for approval once planned",
	},
	cli.BoolFlag{
		Name:  "no-render",
		Usage: "do not attempt to render variables before applying",
	},
	cli.BoolFlag{
		Name:  "render-local",
		Usage: "render files locally instead of updating their values in TFC",
	},
	cli.StringFlag{
		Name:  "output,o",
		Usage: "file on which to write the run ID",
	},
}

var dryRun = cli.BoolFlag{
	Name:  "dry-run",
	Usage: "simulate what TFCW would do onto the TFC API",
}

var currentRun = cli.BoolFlag{
	Name:  "current",
	Usage: "perform the action against the current run",
}

var message = cli.StringFlag{
	Name:  "message,m",
	Usage: "custom message for the action",
	Value: "from TFCW",
}
