package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mvisonneau/go-helpers/logger"
	cli "github.com/urfave/cli/v2"
	"github.com/zclconf/go-cty/cty/function"

	tfcw "github.com/mvisonneau/tfcw/lib/client"
	"github.com/mvisonneau/tfcw/lib/functions"
	"github.com/mvisonneau/tfcw/lib/schemas"
	"github.com/mvisonneau/tfcw/lib/terraform"

	log "github.com/sirupsen/logrus"
)

var start time.Time

func configure(ctx *cli.Context) (c *tfcw.Client, cfg *schemas.Config, err error) {
	start = ctx.App.Metadata["startTime"].(time.Time)

	lc := &logger.Config{
		Level:  ctx.String("log-level"),
		Format: ctx.String("log-format"),
	}

	if err = lc.Configure(); err != nil {
		return
	}

	cfg = &schemas.Config{
		Runtime: schemas.Runtime{
			WorkingDir: ctx.String("working-dir"),
		},
	}

	tfcwConfigFile := computeConfigFilePath(cfg.Runtime.WorkingDir, ctx.String("config-file"))
	log.Debugf("Using config file at %s", tfcwConfigFile)

	// Create and EvalContext to define functions that we can use within the HCL for interpolation
	evalCtx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"env": functions.EnvFunction,
		},
	}

	err = hclsimple.DecodeFile(tfcwConfigFile, evalCtx, cfg)
	if err != nil {
		return c, cfg, fmt.Errorf("tfcw config/hcl: %s", err.Error())
	}

	if err = computeRuntimeConfigurationForTFC(cfg, ctx); err != nil {
		return
	}

	c, err = tfcw.NewClient(cfg)
	return
}

func exit(exitCode int, err error) cli.ExitCoder {
	defer log.WithFields(
		log.Fields{
			"execution-time": time.Since(start),
		},
	).Debug("exited..")

	if err != nil {
		log.Error(err.Error())
	}

	return cli.NewExitError("", exitCode)
}

// ExecWrapper gracefully logs and exits our `run` functions
func ExecWrapper(f func(ctx *cli.Context) (int, error)) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		return exit(f(ctx))
	}
}

func computeConfigFilePath(workingDir, configFile string) string {
	return strings.NewReplacer("<working-dir>", workingDir).Replace(configFile)
}

func computeRuntimeConfigurationForTFC(cfg *schemas.Config, ctx *cli.Context) (err error) {
	if cfg.TFC == nil {
		cfg.TFC = &schemas.TFC{}
	}

	if cfg.TFC.Workspace == nil {
		cfg.TFC.Workspace = &schemas.Workspace{}
	}

	// Address
	cfg.Runtime.TFC.Address, err = computeRuntimeTFCAddress(cfg.Runtime.WorkingDir, ctx.String("address"), cfg.TFC.Address)
	if err != nil {
		return
	}

	// Token
	cfg.Runtime.TFC.Token, err = computeRuntimeTFCToken(cfg.Runtime.WorkingDir, ctx.String("token"), cfg.TFC.Token)
	if err != nil {
		return
	}

	// Organization
	cfg.Runtime.TFC.Organization, err = computeRuntimeTFCOrganization(cfg.Runtime.WorkingDir, ctx.String("organization"), cfg.TFC.Organization)
	if err != nil {
		return
	}

	// Workspace
	cfg.Runtime.TFC.Workspace, err = computeRuntimeTFCWorkspace(cfg.Runtime.WorkingDir, ctx.String("workspace"), cfg.TFC.Workspace.Name)
	return
}

func computeRuntimeTFCAddress(workingDir, flagValue string, tfcwValue *string) (string, error) {
	if flagValue != "" {
		log.Debugf("Using TFC address '%s' from CLI flag (or env variable)", returnHTTPSPrefixedURL(flagValue))
		return returnHTTPSPrefixedURL(flagValue), nil
	}

	if tfcwValue != nil {
		log.Debugf("Using TFC address '%s' from TFCW config", returnHTTPSPrefixedURL(*tfcwValue))
		return returnHTTPSPrefixedURL(*tfcwValue), nil
	}

	rbc, err := terraform.GetRemoteBackendConfig(workingDir)
	if err != nil {
		return "", err
	}

	if rbc != nil && rbc.Hostname != "" {
		log.Debugf("Using TFC address '%s' from Terraform remote backend configuration", returnHTTPSPrefixedURL(rbc.Hostname))
		return returnHTTPSPrefixedURL(rbc.Hostname), nil
	}

	log.Debug("Using default TFC address 'https://app.terraform.io'")
	return "https://app.terraform.io", nil
}

func computeRuntimeTFCToken(workingDir, flagValue string, tfcwValue *string) (string, error) {
	if flagValue != "" {
		log.Debug("Using TFC token '***' from CLI flag (or env variable)")
		return flagValue, nil
	}

	if tfcwValue != nil {
		log.Debug("Using TFC token '***' from TFCW config")
		return *tfcwValue, nil
	}

	rbc, err := terraform.GetRemoteBackendConfig(workingDir)
	if err != nil {
		return "", err
	}

	if rbc != nil && rbc.Token != "" {
		log.Debug("Using TFC token '***' from Terraform remote backend configuration")
		return rbc.Token, nil
	}

	// By not returning an empty string, we allow the TFCW to be initiated, yay!
	return "_", nil
}

func computeRuntimeTFCOrganization(workingDir, flagValue string, tfcwValue *string) (string, error) {
	if flagValue != "" {
		log.Debugf("Using TFC organization '%s' from CLI flag (or env variable)", flagValue)
		return flagValue, nil
	}

	if tfcwValue != nil {
		log.Debugf("Using TFC organization '%s' from TFCW config", *tfcwValue)
		return *tfcwValue, nil
	}

	rbc, err := terraform.GetRemoteBackendConfig(workingDir)
	if err != nil {
		return "", err
	}

	if rbc != nil && rbc.Organization != "" {
		log.Debugf("Using TFC organization '%s' from Terraform remote backend configuration", rbc.Organization)
		return rbc.Organization, nil
	}

	return "", nil
}

func computeRuntimeTFCWorkspace(workingDir, flagValue string, tfcwValue *string) (string, error) {
	if flagValue != "" {
		log.Debugf("Using TFC workspace '%s' from CLI flag (or env variable)", flagValue)
		return flagValue, nil
	}

	if tfcwValue != nil {
		log.Debugf("Using TFC workspace '%s' from TFCW config", *tfcwValue)
		return *tfcwValue, nil
	}

	rbc, err := terraform.GetRemoteBackendConfig(workingDir)
	if err != nil {
		return "", err
	}

	if rbc != nil && rbc.Workspace != "" {
		log.Debugf("Using TFC workspace '%s' from Terraform remote backend configuration", rbc.Workspace)
		return rbc.Workspace, nil
	}

	return "", nil
}

func returnHTTPSPrefixedURL(url string) string {
	re := regexp.MustCompile(`^(http|https)://`)
	if re.MatchString(url) {
		return url
	}
	return fmt.Sprintf("https://%s", url)
}
