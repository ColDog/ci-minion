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
		Before: []string{"echo 'pre test'"},
		Main: []string{"echo 'test'"},
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
}
