package main

import "testing"

func TestSampleJob(t *testing.T) {
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

	job := NewJob("test", repo, build)
	job.Run()
}
