package cmd

import (
	"errors"
	"fmt"

	"github.com/gr00by87/fst/config"
	"github.com/gr00by87/fst/core"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"
	surveyCore "gopkg.in/AlecAivazis/survey.v1/core"
)

var (
	// flag variables.
	awsCredentials *bool
	bastionHosts   *bool
	checkStatus    *bool

	// status symbols.
	success = aurora.Green("✓")
	failure = aurora.Red("✗")

	// configCmd represents the config command.
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Setup configuration file",
		Long:  "This subcommand displays prompts to setup a configuration file. If no flag is passed, the whole configuration is run.",
		Run:   runConfig,
	}
)

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(configCmd)
	awsCredentials = configCmd.Flags().BoolP("aws-credentials", "a", false, "displays prompts to setup aws credentials")
	bastionHosts = configCmd.Flags().BoolP("bastion-hosts", "b", false, "updates bastion hosts list")
	checkStatus = configCmd.Flags().BoolP("check-status", "c", false, "checks configuration status")

	surveyCore.QuestionIcon = "🔒"
	surveyCore.ErrorTemplate = fmt.Sprintf("%s %s\n", failure.String(), aurora.Red("{{.Error}}").String())
}

// runConfig executes the config command.
func runConfig(_ *cobra.Command, _ []string) {
	if *checkStatus {
		checkConfigStatus()
		return
	}

	var (
		cfg *config.Config
		err error
	)

	// No switch passed, run the whole configuration.
	if !*awsCredentials && !*bastionHosts {
		*awsCredentials, *bastionHosts = true, true
	}

	if *awsCredentials {
		if err = saveConfig(cfg, getAWSCredentials); err != nil {
			exitWithError(err)
		}
		fmt.Println(success, "AWS credentials succesfully updated")
	}

	if *bastionHosts {
		if err = saveConfig(cfg, getBastionHosts); err != nil {
			exitWithError(err)
		}
		fmt.Println(success, "Bastion hosts list succesfully updated")
	}
}

// saveConfig is a wrapper around configuration functions to save the changes
// after each configuration step.
func saveConfig(cfg *config.Config, cfgFunc func(*config.Config) error) error {
	var err error

	if cfg == nil {
		cfg, err = config.LoadFromFile()
		if err != nil {
			cfg = &config.Config{}
		}
	}

	if err = cfgFunc(cfg); err != nil {
		return err
	}

	return config.SaveToFile(cfg)
}

// getAWSCredentials runs aws credentials configuration.
func getAWSCredentials(cfg *config.Config) error {
	prompts := []*survey.Question{
		{
			Name:     "id",
			Prompt:   &survey.Input{Message: "Enter AWS ID:"},
			Validate: validateLength("AWS Secret", 20),
		},
		{
			Name:     "secret",
			Prompt:   &survey.Input{Message: "Enter AWS Secret:"},
			Validate: validateLength("AWS Secret", 40),
		},
	}

	return survey.Ask(prompts, &cfg.AWSCredentials)
}

// getBastionHosts runs bastion hosts configuration.
func getBastionHosts(cfg *config.Config) error {
	fmt.Println(aurora.Cyan("ⓘ"), "Updating bastion hosts list...")

	typeFilter := core.NewFilter(core.TagType, []string{"bastion"}, core.Equals, false)
	servers, err := core.GetAllServers(cfg.AWSCredentials, core.AllowedRegions, typeFilter)
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		return errors.New("no servers found")
	}

	cfg.BastionHosts = make(map[string][]string)
	for _, server := range servers {
		cfg.BastionHosts[server.Region] = append(cfg.BastionHosts[server.Region], server.PublicIP)
	}

	return nil
}

// checkConfigStatus checks the current configuration status.
func checkConfigStatus() {
	awsCredentialsStatus, bastionHostsStatus := aurora.Red("✗"), aurora.Red("✗")

	if cfg, err := config.LoadFromFile(); err == nil {
		if cfg.AWSCredentials.ID != "" && cfg.AWSCredentials.Secret != "" {
			awsCredentialsStatus = aurora.Green("✓")
		}

		if len(cfg.BastionHosts) > 0 {
			bastionHostsStatus = aurora.Green("✓")
		}
	}

	fmt.Println(awsCredentialsStatus, "AWS credentials configuration")
	fmt.Println(bastionHostsStatus, "Bastion hosts configuration")
}

// validateLength validates the prompts input value length.
func validateLength(label string, length int) func(interface{}) error {
	return func(val interface{}) error {
		if str := val.(string); len(str) != length {
			return fmt.Errorf("%s must have exactly %d characters", label, length)
		}
		return nil
	}
}
