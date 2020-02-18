package cli

import (
	"github.com/urfave/cli"
)

var tfc = []cli.Flag{
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
}

var tf = []cli.Flag{
	cli.StringFlag{
		Name:   "tf-config-folder,f",
		EnvVar: "TFCW_TF_CONFIG_FOLDER",
		Usage:  "`path` to the Terraform configuration folder",
		Value:  ".",
	},
	cli.BoolFlag{
		Name:   "no-render",
		EnvVar: "TFCW_NO_RENDER",
		Usage:  "do not attempt to render variables before applying",
	},
	cli.BoolFlag{
		Name:   "render-local",
		EnvVar: "TFCW_RENDER_LOCAL",
		Usage:  "render files locally instead of updating their values in TFC",
	},
}

var dryRun = cli.BoolFlag{
	Name:   "dry-run",
	EnvVar: "TFCW_DRY_RUN",
	Usage:  "simulate what TFCW would do onto the TFC API",
}
