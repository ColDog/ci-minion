package main

import (
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := &App{
		cli: cli.NewApp(),
		AuthHeader: user + ":" + pass,
		SimpleCiApi: simapi,
		MinionApi: minapi,
	}

	app.cli.Run(os.Args)

}
