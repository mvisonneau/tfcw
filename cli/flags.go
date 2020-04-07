package cli

import (
	"github.com/urfave/cli"
)

var runCreate = []cli.Flag{
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
	cli.DurationFlag{
		Name:  "start-timeout,t",
		Usage: "time to wait for the plan to start (set to 0 to disable, it is the default)",
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

var ignoreTTLs = cli.BoolFlag{
	Name:  "ignore-ttls",
	Usage: "render all variables, unconditionnaly of their current expirations or configured TTLs",
}

var renderType = cli.StringFlag{
	Name:  "render-type,r",
	Usage: "where to render to values - options are : tfc, local or disabled",
	Value: "tfc",
}

var discardAllPendingRuns = cli.BoolFlag{
	Name:  "all,a",
	Usage: "discard all pending runs",
}
