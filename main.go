package main

import (
	"log"
	"os"

	"github.com/doncicuto/openuem-console/internal/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "openuem-console",
		Commands:  getCommands(),
		Usage:     "The OpenUEM console allows and organization to manage its endpoints from a Web User Interface",
		Authors:   []*cli.Author{{Name: "Miguel Angel Alvarez Cabrerizo", Email: "mcabrerizo@sologitops.com"}},
		Copyright: "2024 - Miguel Angel Alvarez Cabrerizo <https://github.com/doncicuto>",
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
