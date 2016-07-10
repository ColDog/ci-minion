package main

import (
	"flag"
)

const LOG_OUTPUT bool = false


func main() {
	action := flag.String("action", "server", "choose what to start")
	port := flag.String("port", "8000", "port to start the server on")

	branch := flag.String("branch", "master", "branch for the sandbox environment")
	provider := flag.String("provider", "github", "provider (bitbucket or github) for the sandbox environment")
	org := flag.String("org", "", "base url for the sandbox")
	project := flag.String("project", "", "project for the sandbox")


	if *action == "sandbox" {
		job := NewJob(*project, Repo{
			Branch: *branch,
			Provider: *provider,
			Organization: *org,
			Project: *project,
		})
		job.Sandbox()
	} else {
		minion := NewMinion()
		minion.Start(*port)
	}
}
