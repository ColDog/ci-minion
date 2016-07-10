package main

import (
	"flag"
)

var (
	LOG_OUTPUT bool
)

func main() {
	action 		:= flag.String("action", "server", "choose what to start")
	port 		:= flag.String("port", "8000", "port to start the server on")

	api 		:= flag.String("api", "", "url for the main api")
	host 		:= flag.String("host", "", "my host to broadcast")
	logOutput 	:= flag.Bool("log-out", false, "should the stdout from the commands be included in the logs")

	branch 		:= flag.String("branch", "master", "branch for the sandbox environment")
	provider 	:= flag.String("provider", "github", "provider (bitbucket or github) for the sandbox environment")
	org 		:= flag.String("org", "", "base url for the sandbox")
	project 	:= flag.String("project", "", "project for the sandbox")

	LOG_OUTPUT = *logOutput

	if *action == "sandbox" {
		job := NewJob(*project, Repo{
			Branch: *branch,
			Provider: *provider,
			Organization: *org,
			Project: *project,
		})
		job.Sandbox()
	} else {
		minion := NewMinion(*api, *host)
		minion.Start(*port)
	}
}
