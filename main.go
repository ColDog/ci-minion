package main

import (
	"os"
)

var (
	LOG_OUTPUT bool
)

func main() {
	LOG_OUTPUT = os.Getenv("LOG_OUTPUT") == "true"

	validateEnvVars([]string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_KEY_ID",
		"MINION_HOST",
		"MINION_PORT",
		"MINION_API",
		"MINION_TOKEN",
		"MINION_S3_REGION",
		"MINION_S3_BUCKET",
	})

	minion := NewMinion()
	minion.Start()
}
