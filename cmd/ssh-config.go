package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/alecthomas/template"
	"github.com/gr00by87/fst/templates"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

var (
	proxyJumpRegion *string

	// templateName stores the template name.
	templateName = "ssh-config"

	// configTemplate stores the ssh config template.
	configTemplate = template.Must(template.New(templateName).Parse(templates.SSHConfig))

	// sshConfigCmd represents the ssh-config command.
	sshConfigCmd = &cobra.Command{
		Use:   "ssh-config",
		Short: "Create ssh config file",
		Long:  "This subcommand generates a ssh config file containing all the bastion hosts and ProxyJump configuration for selected region.",
		Run:   runSSHConfig,
	}
)

// templateData stores the ssh config template data.
type templateData struct {
	JumpHost     string
	BastionHosts []bastionHost
}

// bastionHost stores a single bastion host data.
type bastionHost struct {
	Region string
	IP     string
	ID     int
}

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(sshConfigCmd)
	proxyJumpRegion = sshConfigCmd.Flags().StringP("region", "r", "us-east-1", "region to use in ProxyJump configuration, one of: us-east-1,us-west-2,eu-west-1,ap-northeast-1,ap-southeast-2")
}

// runSSHConfig executes the ssh-config command.
func runSSHConfig(_ *cobra.Command, _ []string) {
	cfg := checkBastionHosts()

	if _, ok := cfg.BastionHosts[*proxyJumpRegion]; !ok {
		exitWithError(fmt.Errorf("invalid region: %s", *proxyJumpRegion))
	}

	if !proceed() {
		return
	}

	bastionHosts := []bastionHost{}
	for region, hosts := range cfg.BastionHosts {
		for id, ip := range hosts {
			bastionHosts = append(bastionHosts, bastionHost{
				Region: region,
				IP:     ip,
				ID:     id + 1,
			})
		}
	}

	sshConfigFile, err := openConfigFile()
	if err != nil {
		exitWithError(fmt.Errorf("error opening ssh config: %v", err))
	}
	defer sshConfigFile.Close()

	if err = configTemplate.ExecuteTemplate(sshConfigFile, templateName, templateData{
		JumpHost:     fmt.Sprintf("%s-01", *proxyJumpRegion),
		BastionHosts: bastionHosts,
	}); err != nil {
		exitWithError(fmt.Errorf("error saving ssh config: %v", err))
	}

	fmt.Println(success, "SSH config updated successfully")
}

// proceed displays a confirmation prompt.
func proceed() (proceed bool) {
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "This will overwrite your current ssh config file, do you want to proceed?",
		},
		&proceed,
		survey.Required,
	); err != nil {
		exitWithError(err)
	}
	return
}

// openConfigFile opens ssh config file for writing.
func openConfigFile() (*os.File, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(path.Join(usr.HomeDir, ".ssh/config"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}
