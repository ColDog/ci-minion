package main

import (
	"os"
	"io/ioutil"
	"fmt"

	"github.com/urfave/cli"
	"github.com/ghodss/yaml"
)

func main() {
	app := &App{
		cli: cli.NewApp(),
	}

	app.configure()

	app.addCmd(cli.Command{
		Name:    "apply",
		Aliases: []string{"a"},
		Usage:   "register the requred builds",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "build-file, b",
				Value: "builds.yml",
				Usage: "specify build file to apply",
			},
		},
		Action:  func(c *cli.Context) error {
			b := c.String("build-file")
			res := make(map[string] map[string] interface{})

			data, err := ioutil.ReadFile(b)
			app.handleErr(err)

			err = yaml.Unmarshal([]byte(data), &res)
			app.handleErr(err)

			for key, config := range res {
				config["name"] = key
				fmt.Printf("creating job definition %s\n", key)
				err := app.post("/users/" + app.User + "/job_definitions", config, nil)
				app.handleErr(err)
			}

			return nil
		},
	})

	app.addCmd(cli.Command{
		Name:    "provision-secrets",
		Usage:   "load in secret environment variables",
		Action:  func(c *cli.Context) error {
			res := make(map[string] []struct{
				Key 	string 	`json:"key"`
				Value 	string 	`json:"value"`
			})

			err := app.get("/users/" + app.User + "/secrets", nil, &res)
			app.handleErr(err)

			for _, secret := range res["secrets"] {
				fmt.Printf("export %s=%s\n", secret.Key, secret.Value)
			}
			return nil
		},
	})

	app.addCmd(cli.Command{
		Name: "server",
		Usage: "start the minion server",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name: "exit-post-build, e",
				Usage: "should the server stop after a build",
			},
		},
		Action: func(c *cli.Context) {
			minion := &Minion{
				cancel: make(chan bool),
				app: app,
				exitPostBuild: c.Bool("exit-post-build"),
			}

			minion.Start()
		},
	})

	app.cli.Run(os.Args)

}
