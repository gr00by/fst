package cmd

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gr00by87/fst/config"
	"github.com/logrusorgru/aurora"
	surveyCore "gopkg.in/AlecAivazis/survey.v1/core"
)

var (
	// status symbols.
	success = aurora.Green("✓")
	failure = aurora.Red("✗")
)

// init updates the survey error template.
func init() {
	surveyCore.ErrorTemplate = fmt.Sprintf("%s %s\n", failure.String(), aurora.Red("{{.Error}}").String())
}

// checkBastionHosts loads the config file and checks if bastion hosts are
// configured. Exits with error otherwise.
func checkBastionHosts() *config.Config {
	cfg, err := config.LoadFromFile()
	if err != nil {
		exitWithError(err)
	}

	if len(cfg.BastionHosts) == 0 {
		exitWithError(errors.New("bastion hosts not configured, use `fst config` to configure"))
	}

	return cfg
}

// randomHost selects a random host from hosts slice.
func randomHost(hosts []string) string {
	rand.Seed(time.Now().Unix())
	return hosts[rand.Intn(len(hosts))]
}
