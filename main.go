package main

import (
	"os"
	"io/ioutil"
	"fmt"

	"github.com/urfave/cli"
	"github.com/ghodss/yaml"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
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

	//app.addCmd(cli.Command{
	//	Name: "sandbox",
	//	Usage: "build up your test docker image to play around with",
	//	Flags: []cli.Flag{
	//		cli.StringFlag{Name: "build-file, b", Value: "builds.yml", Usage: "specify build file"},
	//		cli.StringFlag{Name: "job, j", Usage: "specify job to run, defaults to first"},
	//	},
	//	Action: func(c *cli.Context) {
	//		val := map[string] struct{
	//			Env 		map[string] string 	`json:"env"`
	//			Services 	[]Service 		`json:"services"`
	//			Before 		[]string		`json:"before"`
	//			After 		[]string		`json:"after"`
	//			Main 		[]string		`json:"main"`
	//			OnSuccess 	[]string		`json:"on_success"`
	//			OnFailure 	[]string		`json:"on_failure"`
	//		}{}
	//
	//
	//		for k, v := range val[]
	//
	//
	//		Build{
	//			Env:
	//		}
	//
	//	},
	//})

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
		Name:    "emit",
		Usage:   "emit a custom event",
		ArgsUsage: "event [ payload ]",
		Action:  func(c *cli.Context) error {
			if c.Args().First() == "" {
				app.handleErr(errors.New("requires event name"))
			}

			payload := make(map[string] interface{})
			if c.Args().Get(1) != "" {
				err := json.Unmarshal([]byte(c.Args().Get(1)), &payload)
				app.handleErr(err)
			}

			err := app.post("/users/" + app.User + "/events", map[string] map[string] interface{} {
				"event": map[string] interface{} {
					"name": c.Args().First(),
					"payload": payload,
				},
			}, nil)
			app.handleErr(err)

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
