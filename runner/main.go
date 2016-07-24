package main

import (
	"flag"
	"os"
)

func config() MinionConfig {
	config := MinionConfig{}

	config.AwsAccessKey = *flag.String("aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key")
	config.AwsSecretKey = *flag.String("aws-secret-key", os.Getenv("AWS_SECRET_KEY_ID"), "AWS secret key")
	config.SimpleCiApi = *flag.String("simplci-api", "http://localhost:3000", "server to connect to")
	config.MinionApi = *flag.String("minion-api", "http://localhost:8000", "this minions endpoint")
	config.MinionToken = *flag.String("token", "secret", "token to use in requests")
	config.S3Bucket = *flag.String("s3-bucket", "simplecistorage", "s3 bucket")
	config.S3Region = *flag.String("s3-region", "us-west-2", "s3 region")

	flag.Parse()

	return config
}

func main() {
	minion := NewMinion(config())
	minion.Start()
}
