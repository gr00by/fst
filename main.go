package main

import (
	"github.com/gr00by87/fst/cmd"
)

// Version stores the application version.
var Version string

func main() {
	cmd.Execute(Version)
}
