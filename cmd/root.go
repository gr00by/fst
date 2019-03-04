package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use: "fst",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(version string) {
	rootCmd.Long = fmt.Sprintf("fst version %s", version)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// exitWithError prints the error and exits with code 1.
func exitWithError(err error) {
	fmt.Println(err)
	os.Exit(1)
}
