package main

import (
	"flag"
	"os"
)

func config() MinionConfig {
	config := MinionConfig{}

	config.AwsAccessKey = *flag.String("aws-access-key", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key")
	config.AwsSecretKey = *flag.String("aws-secret-key", os.Getenv("AWS_SECRET_KEY_ID"), "AWS secret key")
	config.SimpleCiApi = *flag.String("simplci-api", os.Getenv("SIMLECI_API"), "server to connect to")
	config.MinionApi = *flag.String("minion-api", os.Getenv("MINION_API"), "this minions endpoint")
	config.MinionToken = *flag.String("token", os.Getenv("MINION_SECRET"), "token to use in requests")
	config.S3Bucket = *flag.String("s3-bucket", os.Getenv("S3_BUCKET"), "s3 bucket")
	config.S3Region = *flag.String("s3-region", os.Getenv("S3_REGION"), "s3 region")

	if config.S3Bucket == "" {
		config.S3Bucket = "simplecistoreage"
	}

	if config.S3Region == "" {
		config.S3Region = "us-west-2"
	}

	flag.Parse()

	return config
}

func main() {
	minion := NewMinion(config())
	minion.Start()
}
