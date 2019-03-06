package cmd

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gr00by87/fst/config"
	"github.com/gr00by87/fst/core"
	"github.com/spf13/cobra"
)

var (
	identityFile *string
	loginName    *string

	// sshCmd represents the config command.
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
	identityFile = sshCmd.Flags().StringP("identity-file", "i", "", "identity file location")
	loginName = sshCmd.Flags().StringP("login-name", "l", "", "login user name")
}

// runSSH executes the ssh command.
func runSSH(_ *cobra.Command, args []string) {
	cfg, err := config.LoadFromFile()
	if err != nil {
		exitWithError(err)
	}

	if len(cfg.BastionHosts) == 0 {
		exitWithError(errors.New("bastion hosts not configured, use `fst config` to configure"))
	}

	server, err := core.GetSingleServer(cfg.AWSCredentials, core.NewServerID(args[0]))
	if err != nil {
		exitWithError(err)
	}

	bastionHost, ok := cfg.BastionHosts[server.Region]
	if !ok {
		exitWithError(fmt.Errorf("bastion host not found for region: %s", server.Region))
	}

	cmd := []string{"ssh", "-J", randomHost(bastionHost)}
	if *identityFile != "" {
		cmd = append(cmd, "-i", *identityFile)
	}
	if *loginName != "" {
		cmd = append(cmd, "-l", *loginName)
	}
	cmd = append(cmd, server.PrivateIP)

	fmt.Println(strings.Join(cmd, " "))
	os.Exit(3)
}

// randomHost selects a random host from hosts slice.
func randomHost(hosts []string) string {
	rand.Seed(time.Now().Unix())
	return hosts[rand.Intn(len(hosts))]
}
