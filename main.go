package main

import (
	"os"
	"flag"
)

var Config struct {
	LogOutput bool
	AwsAccessKeyID string
	AwsSecretKeyID string
	MinionApi string
	SimpleCiApi string
	MinionToken string
	S3Bucket string
	S3Region string
}

func Configure()  {
	Config.LogOutput = *flag.Bool("log", true, "should log output")
	Config.AwsAccessKeyID = *flag.String("aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key")
	Config.AwsSecretKeyID = *flag.String("aws-secret-key", os.Getenv("AWS_SECRET_KEY_ID"), "AWS secret key")
	Config.SimpleCiApi = *flag.String("simplci-api", "http://localhost:3000", "server to connect to")
	Config.MinionApi = *flag.String("minion-api", "http://localhost:8000", "this minions endpoint")
	Config.MinionToken = *flag.String("token", "secret", "token to use in requests")
	Config.S3Bucket = *flag.String("s3-bucket", "simplecistorage", "s3 bucket")
	Config.S3Region = *flag.String("s3-region", "us-west-2", "s3 region")

	flag.Parse()
}

func main() {
	Configure()

	minion := NewMinion()
	minion.Start()
}
