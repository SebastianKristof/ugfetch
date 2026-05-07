package main

import (
	"os"

	"ugfetch/internal/app"
	"ugfetch/internal/cli"
)

func main() {
	opts, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fatal(err)
	}
	if err := app.Run(opts, app.DefaultDependencies()); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	if err == nil {
		return
	}
	if err.Error() == cli.UsageText {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
	os.Stderr.WriteString(err.Error() + "\n")
	os.Exit(1)
}
