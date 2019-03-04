package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	// scpCmd represents the config command.
	scpCmd = &cobra.Command{
		Use:   "scp",
		Short: "Copy file to, from, or between instances",
		Long:  "Not implemented yet",
		Run:   runSCP,
	}
)

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(scpCmd)
}

// runConfig executes the config command.
func runSCP(_ *cobra.Command, _ []string) {
	exitWithError(errors.New("scp not implemented yet"))
}
