package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gr00by87/fst/core"
	"github.com/spf13/cobra"
)

var (
	forwardPort                  *string
	loginName                    *string
	sshConfigFile                *string
	sshIdentityFile              *string
	sshDoNotExecuteRemoteCommand *bool

	// sshCmd represents the ssh command.
	sshCmd = &cobra.Command{
		Use:   "ssh",
		Args:  cobra.ExactArgs(1),
		Short: "Connect via ssh to an instance",
		Long:  "This subcommand connects via ssh to an instance. It accepts exactly 1 argument - an instance identifier, which can be either server's public ip address, private ip address or it's name",
		Run:   runSSH,
	}
)

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(sshCmd)
	forwardPort = sshCmd.Flags().StringP("forward-port", "L", "", "forward local port")
	loginName = sshCmd.Flags().StringP("login-name", "l", "", "login user name")
	sshConfigFile = sshCmd.Flags().StringP("config-file", "F", "", "configuration file location")
	sshIdentityFile = sshCmd.Flags().StringP("identity-file", "i", "", "identity file location")
	sshDoNotExecuteRemoteCommand = sshCmd.Flags().BoolP("do-not-execute", "N", false, "do not execute a remote command (this is useful for just forwarding ports)")
}

// runSSH executes the ssh command.
func runSSH(_ *cobra.Command, args []string) {
	cfg := checkBastionHosts()

	server, err := core.GetSingleServer(cfg.AWSCredentials, core.NewServerID(args[0]))
	if err != nil {
		exitWithError(err)
	}

	bastionHosts := cfg.BastionHosts[server.Region]
	if len(bastionHosts) == 0 && !disableBastionHostCheck[server.Region] {
		exitWithError(fmt.Errorf("bastion host not found for region: %s", server.Region))
	}

	cmd := []string{"ssh"}
	if len(bastionHosts) > 0 {
		cmd = append(cmd, "-J", randomHost(bastionHosts))
	}
	if *forwardPort != "" {
		cmd = append(cmd, "-L", *forwardPort)
	}
	if *loginName != "" {
		cmd = append(cmd, "-l", *loginName)
	}
	if *sshConfigFile != "" {
		cmd = append(cmd, "-F", *sshConfigFile)
	}
	if *sshIdentityFile != "" {
		cmd = append(cmd, "-i", *sshIdentityFile)
	}
	cmd = append(cmd, server.PrivateIP)
	if *sshDoNotExecuteRemoteCommand {
		cmd = append(cmd, "-N")
	}

	fmt.Println(strings.Join(cmd, " "))
	os.Exit(3)
}
