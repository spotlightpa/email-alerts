package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/email-alerts/pkg/emailalerts"
)

func main() {
	exitcode.Exit(emailalerts.CLI(os.Args[1:]))
}
