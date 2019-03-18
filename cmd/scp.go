package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gr00by87/fst/core"
	"github.com/spf13/cobra"
)

var (
	scpConfigFile   *string
	scpIdentityFile *string

	// instanceRe is used to extract the instance identifier from command args.
	instanceRe = regexp.MustCompile(`^(?:.*@|)(.*)\:.*`)

	// scpCmd represents the scp command.
	scpCmd = &cobra.Command{
		Use:   "scp",
		Args:  cobra.MinimumNArgs(2),
		Short: "Copy file to, from, or between instances",
		Long:  "This subcommand allows files to be copied to, from, or between instances. It accepts either server's public ip address, private ip address or it's name as instance identifier.",
		Run:   runSCP,
	}
)

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(scpCmd)
	scpConfigFile = scpCmd.Flags().StringP("config-file", "F", "", "configuration file location")
	scpIdentityFile = scpCmd.Flags().StringP("identity-file", "i", "", "identity file location")
}

// runSCP executes the scp command.
func runSCP(_ *cobra.Command, args []string) {
	cfg := checkBastionHosts()

	region := ""
	for i, arg := range args {
		if matches := instanceRe.FindStringSubmatch(arg); len(matches) == 2 {
			server, err := core.GetSingleServer(cfg.AWSCredentials, core.NewServerID(matches[1]))
			if err != nil {
				exitWithError(err)
			}

			if region != "" && server.Region != region {
				exitWithError(errors.New("servers are not within the same aws region"))
			}

			region = server.Region
			args[i] = strings.Replace(arg, matches[1], server.PrivateIP, 1)
		}
	}

	bastionHost, ok := cfg.BastionHosts[region]
	if !ok {
		exitWithError(fmt.Errorf("bastion host not found for region: %s", region))
	}

	cmd := []string{"scp", fmt.Sprintf("-o 'ProxyJump %s'", randomHost(bastionHost))}
	if *scpConfigFile != "" {
		cmd = append(cmd, "-F", *scpConfigFile)
	}
	if *scpIdentityFile != "" {
		cmd = append(cmd, "-i", *scpIdentityFile)
	}
	cmd = append(cmd, args...)

	fmt.Println(strings.Join(cmd, " "))
	os.Exit(3)
}
