package main

import (
	"os/exec"
	"log"
	"strings"
	"runtime"
	"reflect"
	"net/http"

	"gopkg.in/amz.v1/s3"
	"gopkg.in/amz.v1/aws"
)

type CommandResult struct {
	Args 		[]string
	Output		string
	Error 		error
}

func execute(quit chan bool, main string, args ...string) CommandResult {
	done := make(chan bool, 1)
	cmd := exec.Command(main, args...)

	go func() {
		select {
		case <- quit:
			cmd.Process.Kill()
			return
		case <- done:
			return
		}
	}()

	output, err := cmd.CombinedOutput()
	done <- true

	log.Printf("executing: %s %s err: %v", main, strings.Join(args, " "), err)

	if LOG_OUTPUT {
		log.Printf("output: %s", output)
	}

	return CommandResult{
		Args: args,
		Error: err,
		Output: string(output),
	}
}

func FuncName(i interface{}) string {
	s := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), ".")
	sa := strings.ToLower(s[len(s) - 1])
	return strings.Split(sa, "-")[0]
}


func NewS3Client(key, secret, region string)  {
	reg, ok := aws.Regions[region]
	if !ok {
		panic(region + " is not a region")
	}

	auth := aws.Auth{
		AccessKey: key, // change this to yours
		SecretKey: secret,
	}

	return &S3Client{
		Client: s3.New(auth, reg),
	}
}

type S3Client struct {
	BucketName 	string
	Client 		*s3.S3
}

func (client *S3Client) Upload(path string, bytes []byte) error {
	filetype := http.DetectContentType(bytes)

	bucket := client.Client.Bucket(client.BucketName)
	return bucket.Put(path, bytes, filetype, s3.ACL("public-read"))
}
