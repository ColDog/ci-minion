package main

import (
	"fmt"
	"os"
	"encoding/json"
	"errors"

	"github.com/urfave/cli"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

type App struct {
	SimpleCiApi	string
	MinionApi	string
	AuthHeader 	string
	S3Region	string
	S3Bucket 	string
	AwsAccess 	string
	AwsSecret 	string
	User 		string

	cli 		*cli.App
	s3 		*s3.Bucket
}

func (app *App) configS3() {
	reg, ok := aws.Regions[app.S3Region]
	if !ok {
		panic(app.S3Region + " is not a region")
	}

	auth := aws.Auth{
		AccessKey: app.AwsAccess,
		SecretKey: app.AwsSecret,
	}

	conn := s3.New(auth, reg)
	bucket := conn.Bucket(app.S3Bucket)

	app.s3 = bucket
}

func (app *App) configure() {
	app.cli.Flags = []cli.Flag{
		cli.StringFlag{Name: "user, u", Value: "me", EnvVar: "MINION_USER"},
		cli.StringFlag{Name: "secret, s", EnvVar: "SIMPLECI_SECRET"},
		cli.StringFlag{Name: "key, k", EnvVar: "SIMPLECI_KEY"},
		cli.StringFlag{Name: "minion-api", EnvVar: "MINION_API"},
		cli.StringFlag{Name: "simpleci-api", EnvVar: "SIMPLECI_API"},
		cli.StringFlag{Name: "s3-bucket", EnvVar: "MINION_S3_BUCKET"},
		cli.StringFlag{Name: "s3-region", EnvVar: "MINION_S3_REGION"},
		cli.StringFlag{Name: "aws-access", EnvVar: "MINION_AWS_ACCESS_KEY"},
		cli.StringFlag{Name: "aws-secret", EnvVar: "MINION_AWS_SECRET_KEY"},
	}

	app.cli.Before = func(c *cli.Context) error {
		app.User = c.GlobalString("user")
		app.AuthHeader = c.GlobalString("key") + ":" + c.GlobalString("secret")
		app.MinionApi = c.GlobalString("minion-api")
		app.SimpleCiApi = c.GlobalString("simpleci-api")
		app.S3Region = c.GlobalString("s3-region")
		app.S3Bucket = c.GlobalString("s3-bucket")
		app.AwsAccess = c.GlobalString("aws-access")
		app.AwsSecret = c.GlobalString("aws-secret")

		app.configS3()

		return nil
	}


}

func (app *App) addCmd(cmd cli.Command)  {
	app.cli.Commands = append(app.cli.Commands, cmd)
}

func (app *App) post(path string, params interface{}, res interface{}) error {
	data := app.parseReq(params)

	resp, body, errs := gorequest.New().Post(app.SimpleCiApi + "/api" + path).
		Set("Accepts", "application/json").
		Set("Authorization", app.AuthHeader).
		Send(data).
		End()

	return app.handleHttp(resp, body, errs, res)
}

func (app *App) patch(path string, params interface{}, res interface{}) error {
	data := app.parseReq(params)

	resp, body, errs := gorequest.New().Patch(app.SimpleCiApi + "/api" + path).
		Set("Accepts", "application/json").
		Set("Authorization", app.AuthHeader).
		Send(data).
		End()

	return app.handleHttp(resp, body, errs, res)
}

func (app *App) get(path string, params map[string] interface{}, res interface{}) error {
	req := gorequest.New().Get(app.SimpleCiApi + "/api" + path).
		Set("Accepts", "application/json").
		Set("Authorization", app.AuthHeader)

	for key, val := range params {
		req.Param(key, fmt.Sprintf("%v", val))
	}

	resp, body, errs := req.End()
	return app.handleHttp(resp, body, errs, res)
}

func (app *App) parseReq(params map[string] interface{}) string {
	var data []byte
	var err error
	if params != nil {
		data, err = json.Marshal(params)
		if err != nil {
			panic(err)
		}
	}

	return string(data)
}

func (app *App) handleHttp(resp gorequest.Response, body string, errs []error, res interface{}) error {
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Status is %v: %s", resp.StatusCode, resp.Status))
	}

	if len(errs) > 0 {
		return errs[0]
	}

	if res == nil {
		js := make(map[string] interface{})
		json.Unmarshal([]byte(body), &js)
		d, _ := json.MarshalIndent(js, " ", "  ")
		fmt.Printf("%s\n", d)
		return nil
	} else {
		return json.Unmarshal([]byte(body), res)
	}
}

func (app *App) handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
