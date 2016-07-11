package main

import (
	"flag"
	"os"
)

var (
	LOG_OUTPUT bool
	SECRET string = os.Getenv("SECRET_KEY_BASE")
)

func main() {
	port 		:= flag.String("port", "8000", "port to start the server on")
	api 		:= flag.String("api", "", "url for the main api")
	host 		:= flag.String("host", "", "my host to broadcast")
	logOutput 	:= flag.Bool("log-out", true, "should the stdout from the commands be included in the logs")

	LOG_OUTPUT = *logOutput
	flag.Parse()

	minion := NewMinion(*api, *host)
	minion.Start(*port)
}
