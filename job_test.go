package main

import (
	"testing"
	"fmt"
)

func testJob() *Job {
	repo := Repo{
		Provider: "github",
		Branch: "master",
		Organization: "coldog",
		Project: "ci-sample",
	}

	build := Build{
		Env: []string{"TEST=true"},
		BaseImage: "ubuntu",
		Services: []Service{
			Service{
				Name: "mysql",
				Image: "mysql:5.7",
				Env: []string{"MYSQL_ROOT_PASSWORD=pass"},
				OnStartup: []string{"echo 'hello from mysql'"},
			},
		},
		Before: []string{"echo 'pre test'", "apt-get update && apt-get install -y curl && apt-get clean"},
		Main: []string{"echo 'test'", "sleep 60"},
		After: []string{"echo 'after'"},
		OnSuccess: []string{"echo 'success!'"},
		OnFailure: []string{"echo 'failure :('"},
	}

	return NewJob("test", repo, build)
}

func TestSampleJob(t *testing.T) {
	job := testJob()
	go job.Run()
	job.Wait()

	fmt.Printf("\n\n%s\n", job.Serialize())
}
