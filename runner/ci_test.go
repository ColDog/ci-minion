package runner

import "testing"

func testJob() *CiJob {
	conf := JobConfig{
		JobId: "test_1",
		JobFamily: "test",
		Repo: Repo{
			Provider: "github",
			Branch: "master",
			Organization: "coldog",
			Project: "ci-sample",
		},

		Build: Build{
			Env: []string{"TEST=true"},
			BaseImage: "ubuntu",
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
		},
	}

	return NewJob(conf)
}

func TestJob(t *testing.T) {
	ci := testJob()

	ci.Run()
}
