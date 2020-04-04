package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/gr00by87/fst/config"
	"github.com/gr00by87/fst/core"
	"github.com/gr00by87/fst/vpn"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"
	surveyCore "gopkg.in/AlecAivazis/survey.v1/core"
)

var (
	awsCredentials *bool
	bastionHosts   *bool
	vpnConfig      *bool
	checkStatus    *bool

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
	vpnConfig = configCmd.Flags().BoolP("vpn-config", "v", false, "displays prompts to setup vpn (optional)")
	checkStatus = configCmd.Flags().BoolP("check-status", "c", false, "checks configuration status")
}

// runConfig executes the config command.
func runConfig(_ *cobra.Command, _ []string) {
	if *checkStatus {
		checkConfigStatus()
		return
	}

	var (
		cfg    *config.Config
		err    error
		runAll bool
	)

	// No switch passed, run the whole configuration.
	if !*awsCredentials && !*vpnConfig && !*bastionHosts {
		runAll = true
	}

	if *awsCredentials || runAll {
		if err = saveConfig(cfg, getAWSCredentials); err != nil {
			exitWithError(err)
		}
		fmt.Println(success, "AWS credentials updated successfully")
	}

	if *bastionHosts || runAll {
		if err = saveConfig(cfg, getBastionHosts); err != nil {
			exitWithError(err)
		}
		fmt.Println(success, "Bastion hosts list updated successfully")
	}

	switch {
	case runAll:
		if !proceed("Do you want to run VPN configuration?") {
			break
		}
		fallthrough
	case *vpnConfig:
		if err = saveConfig(cfg, getVPNConfig); err != nil {
			exitWithError(err)
		}
		fmt.Println(success, "VPN config updated successfully")
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
	surveyCore.QuestionIcon = "🔒"
	prompts := []*survey.Question{
		{
			Name:     "ID",
			Prompt:   &survey.Input{Message: "Enter AWS ID:"},
			Validate: validateLength("AWS Secret", 20),
		},
		{
			Name:     "Secret",
			Prompt:   &survey.Input{Message: "Enter AWS Secret:"},
			Validate: validateLength("AWS Secret", 40),
		},
	}

	return survey.Ask(prompts, &cfg.AWSCredentials)
}

// getVPNConfig runs vpn configuration.
func getVPNConfig(cfg *config.Config) error {
	pritunl, err := vpn.NewPritunl()
	if err != nil {
		return err
	}

	profiles, err := pritunl.ListProfiles()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tPROFILE NAME\tPROFILE ID")

	for i, profile := range profiles {
		fmt.Fprintf(w, "%d\t%s\t%s\n", i+1, profile.Name, profile.ID)
	}
	w.Flush()

	surveyCore.QuestionIcon = "🔒"
	prompts := []*survey.Question{
		{
			Name:   "ProfileID",
			Prompt: &survey.Input{Message: "Select Profile:"},
			Validate: func(val interface{}) error {
				id, err := strconv.Atoi(val.(string))
				if err != nil {
					return errors.New("ID must be a number")
				}
				if id == 0 || id > len(profiles) {
					return errors.New("Invalid ID")
				}
				return nil
			},
			Transform: func(val interface{}) interface{} {
				id, _ := strconv.Atoi(val.(string))
				return profiles[id-1].ID
			},
		},
		{
			Name:   "OTPSecret",
			Prompt: &survey.Input{Message: "Enter OTP Secret (optional):"},
			Validate: func(val interface{}) error {
				if len(val.(string)) > 0 {
					return validateLength("OTP Secret", 16)(val)
				}
				return nil
			},
		},
	}

	return survey.Ask(prompts, &cfg.VPNConfig)
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
	awsCredentialsStatus, bastionHostsStatus, vpnConfigStatus := failure, failure, failure

	if cfg, err := config.LoadFromFile(); err == nil {
		if cfg.AWSCredentials.ID != "" && cfg.AWSCredentials.Secret != "" {
			awsCredentialsStatus = success
		}
		if len(cfg.BastionHosts) > 0 {
			bastionHostsStatus = success
		}
		if cfg.VPNConfig.ProfileID != "" {
			vpnConfigStatus = success
		}
	}

	fmt.Println(awsCredentialsStatus, "AWS credentials configuration")
	fmt.Println(bastionHostsStatus, "Bastion hosts configuration")
	fmt.Println(vpnConfigStatus, "VPN configuration")
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

// proceed displays a confirmation prompt.
func proceed(msg string) (proceed bool) {
	surveyCore.QuestionIcon = "?"
	if err := survey.AskOne(
		&survey.Confirm{
			Message: msg,
		},
		&proceed,
		survey.Required,
	); err != nil {
		exitWithError(err)
	}
	return
}
