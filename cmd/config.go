package cmd

import (
	"errors"
	"fmt"

	"github.com/gr00by87/fst/config"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	isConfigured *bool

	// configCmd represents the config command.
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Setup configuration file",
		Long:  "This subcommand displays prompts to setup a configuration file",
		Run:   runConfig,
	}
)

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(configCmd)
	isConfigured = configCmd.Flags().BoolP("is-configured", "i", false, "checks if application is configured")
}

// runConfig executes the config command.
func runConfig(_ *cobra.Command, _ []string) {
	if *isConfigured {
		if _, err := config.LoadFromFile(); err != nil {
			exitWithError(errors.New("application is not configured"))
		}
		fmt.Println("application is configured")
		return
	}

	var (
		cfg = &config.Config{}
		err error
	)

	cfg.AWSCredentials.ID, err = runPrompt("AWS ID", 20)
	if err != nil {
		exitWithError(err)
	}

	cfg.AWSCredentials.Secret, err = runPrompt("AWS Secret", 40)
	if err != nil {
		exitWithError(err)
	}

	if err = config.SaveToFile(cfg); err != nil {
		exitWithError(err)
	}

	fmt.Println("configuration successful!")
}

// runPrompt executes the prompt.
func runPrompt(label string, length int) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validateLength(label, length),
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	} else if result == "" {
		return "", errors.New("configuration canceled, config file not saved")
	}
	return result, nil
}

// validateLength validates the prompts input value length.
func validateLength(label string, length int) func(string) error {
	return func(input string) error {
		if len(input) != length {
			return fmt.Errorf("%s must have exactly %d characters", label, length)
		}
		return nil
	}
}
