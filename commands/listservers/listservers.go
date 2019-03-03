package listservers

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/gr00by87/fst/config"
	"github.com/gr00by87/fst/core"
	"github.com/jessevdk/go-flags"
)

var (
	opts   options
	parser = flags.NewParser(nil, flags.HelpFlag)
)

// options stores the go-flags parser flags.
type options struct {
	Name       commaDelimited `short:"n" long:"name" description:"Filter servers by Name tag. Multiple comma delimited values are allowed"`
	Env        commaDelimited `short:"e" long:"env" description:"Filter servers by Env tag. Multiple comma delimited values are allowed"`
	Region     commaDelimited `short:"r" long:"region" description:"Look for servers in selected AWS region(s). Any of: us-east-1|us-west-2|eu-west-1|ap-northeast-1|ap-southeast-2|all" default:"us-east-1"`
	IgnoreCase bool           `short:"i" long:"ignore-case" description:"Ignore case in tag filters"`
}

// commaDelimited handles the comma delimited flag values.
type commaDelimited []string

// UnmarshalFlag satisfies the go-flags Unmarshaler interface.
func (cd *commaDelimited) UnmarshalFlag(value string) error {
	vals := strings.Split(value, ",")
	for _, val := range vals {
		*cd = commaDelimited(append([]string(*cd), val))
	}
	return nil
}

// val returns the []string value of commaDelimited.
func (cd *commaDelimited) val() []string {
	return []string(*cd)
}

// init updates the go-flags parser options.
func init() {
	parser.Usage = "list-servers [options]"
	parser.AddGroup("Options", "Options", &opts)
}

// Run runs the list-servers command.
func Run(w *tabwriter.Writer) error {
	cfg, err := config.LoadFromFile()
	if err != nil {
		return err
	}

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintf(w, flagsErr.Message)
			return nil
		} else {
			return err
		}
	}

	nameFilter := core.NewFilter(opts.Name.val(), core.Contains, opts.IgnoreCase)
	envFilter := core.NewFilter(opts.Env.val(), core.Equals, opts.IgnoreCase)
	servers, err := core.GetAllServers(cfg.AWSCredentials, nameFilter, envFilter, opts.Region.val())
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "NAME\tENVIRONMENT\tADDRESS")
	for _, server := range servers {
		fmt.Fprintf(w, "%s\t%s\t%s\n", server.Name, server.Env, server.Address)
	}

	return nil
}
