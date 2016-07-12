package main

import "testing"

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
		PreTest: []string{"echo 'pre test'"},
		Test: []string{"echo 'test'"},
		OnSuccess: []string{"echo 'success!'"},
		OnFailure: []string{"echo 'failure :('"},
	}

	return NewJob("test", repo, build)
}

func TestSampleJob(t *testing.T) {
	job := testJob()
	go job.Run()
	job.Wait()
}
