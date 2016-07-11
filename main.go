package main

import (
	"flag"
)

var (
	LOG_OUTPUT bool
)

func main() {
	port 		:= flag.String("port", "8000", "port to start the server on")
	api 		:= flag.String("api", "http://localhost:3000", "url for the main api")
	host 		:= flag.String("host", "http://localhost:8000", "my host to broadcast")
	secret 		:= flag.String("secret", "secret", "secret key")
	logOutput 	:= flag.Bool("log-out", true, "should the stdout from the commands be included in the logs")

	LOG_OUTPUT = *logOutput
	flag.Parse()

	minion := NewMinion(*api, *host, *secret)
	minion.Start(*port)
}
