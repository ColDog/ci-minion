package main

import (
	"testing"
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
		Setup: []string{"echo 'hello from setup'"},
		Services: map[string] Service {
			"mysql": Service{
					Image: "mysql:5.7",
					Env: []string{"MYSQL_ROOT_PASSWORD=pass"},
					OnStartup: []string{"echo 'hello from mysql'"},
				},
		},
		Before: []string{"echo 'pre test'"},
		Main: []string{"echo 'test'", "sleep 5"},
		After: []string{"echo 'after'"},
		OnSuccess: []string{"echo 'success!'"},
		OnFailure: []string{"echo 'failure :('"},
		Post: []string{"echo 'on build machine'"},
		PostSuccess: []string{"echo 'success!'"},
		PostFailure: []string{"echo 'failure :('"},
	}

	return NewJob("test", repo, build, 1)
}

func TestSampleJob(t *testing.T) {
	job := testJob()
	go job.Run()
	job.Wait()

	//fmt.Printf("\n\n%s\n", job.Serialize())
}
