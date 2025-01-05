package main

import (
	"log"
	"os"

	"github.com/open-uem/openuem-console/internal/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "openuem-console",
		Commands:  getCommands(),
		Usage:     "The OpenUEM console allows and organization to manage its endpoints from a Web User Interface",
		Authors:   []*cli.Author{{Name: "Miguel Angel Alvarez Cabrerizo", Email: "mcabrerizo@openuem.eu"}},
		Copyright: "2024 - Miguel Angel Alvarez Cabrerizo <https://github.com/open-uem>",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getCommands() []*cli.Command {
	return []*cli.Command{
		commands.StartConsole(),
		commands.StopConsole(),
	}
}
