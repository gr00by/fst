package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gr00by87/fst/commands/config"
	"github.com/gr00by87/fst/commands/listservers"
)

var Version string

func main() {
	var (
		w      = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		module string
		err    error
	)

	if len(os.Args) > 1 {
		module = os.Args[1]
	}

	switch module {
	case "config":
		err = config.Run(w)
	case "list-servers":
		err = listservers.Run(w)
	case "ssh":
		err = errors.New("ssh not implemented yet")
	case "scp":
		err = errors.New("scp not implemented yet")
	case "-v", "--version":
		fmt.Fprintf(w, "fst version %s\n", Version)
	case "-h", "--help", "":
		printUsage(w)
	default:
		err = fmt.Errorf("command not found: %s", module)
	}
	if err != nil {
		fmt.Fprintln(w, err.Error())
		os.Exit(1)
	}

	w.Flush()
}

// printUsage prints the application usage data.
func printUsage(w *tabwriter.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "\tfst [help options]")
	fmt.Fprintln(w, "\tfst [command] [command options]\n")

	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "\tconfig\tSetup configuration file")
	fmt.Fprintln(w, "\tlist-servers\tList available servers")
	fmt.Fprintln(w, "\tssh\tConnnect via ssh to an instance")
	fmt.Fprintln(w, "\tscp\tCopy file to, from, or between instances\n")

	fmt.Fprintln(w, "Help Options:")
	fmt.Fprintln(w, "\t-h, --help\tShow this help message")
	fmt.Fprintln(w, "\t-v, --version\tDisplay version information")
}
