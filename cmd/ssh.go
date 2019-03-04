package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	// sshCmd represents the config command.
	sshCmd = &cobra.Command{
		Use:   "ssh",
		Short: "Connnect via ssh to an instance",
		Long:  "Not implemented yet",
		Run:   runSSH,
	}
)

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(sshCmd)
}

// runConfig executes the config command.
func runSSH(_ *cobra.Command, _ []string) {
	exitWithError(errors.New("ssh not implemented yet"))
}
