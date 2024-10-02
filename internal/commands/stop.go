package commands

import "github.com/urfave/cli/v2"

func StopConsole() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop the OpenUEM console",
	}
}
