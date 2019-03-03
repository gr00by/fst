package config

import (
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/gr00by87/fst/config"
	"github.com/jessevdk/go-flags"
	"github.com/manifoldco/promptui"
)

var (
	opts   options
	parser = flags.NewParser(nil, flags.HelpFlag)
)

// options stores the go-flags parser flags.
type options struct {
	IsConfigured bool `short:"i" long:"is-configured" description:"Checks if application is configured"`
}

// init updates the go-flags parser options.
func init() {
	parser.Usage = "config [options]"
	parser.AddGroup("Options", "Options", &opts)
}

// Run runs the config command.
func Run(w *tabwriter.Writer) error {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintf(w, flagsErr.Message)
			return nil
		} else {
			return err
		}
	}

	if opts.IsConfigured {
		if _, err := config.LoadFromFile(); err != nil {
			return errors.New("application is not configured")
		}
		fmt.Fprintln(w, "application is configured")
		return nil
	}

	var (
		cfg = &config.Config{}
		err error
	)

	cfg.AWSCredentials.ID, err = runPrompt("AWS ID", 20)
	if err != nil {
		return err
	}

	cfg.AWSCredentials.Secret, err = runPrompt("AWS Secret", 40)
	if err != nil {
		return err
	}

	if err = config.SaveToFile(cfg); err != nil {
		return err
	}

	fmt.Fprintln(w, "configuration successful!")
	return nil
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
