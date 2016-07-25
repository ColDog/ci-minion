package main

import (
	"os"
	"strings"
	"encoding/json"
	"fmt"
)

type CiJob struct {
	Runner
	Job 		JobConfig
	folder 		string
}

func NewJob(job JobConfig) *CiJob {
	return &CiJob{
		Job: job,
		Runner: Runner{
			finished: make(chan bool, 1),
			quit: make(chan bool, 1),
			idx: 0,
			id: job.JobId,
			Status: RunStatus{},
		},
	}
}

func (ci *CiJob) Setup() bool {
	ci.execute("docker", "network", "create", "test-net")

	for _, env := range ci.Job.env() {
		k := strings.Split(env, "=")[0]
		v := strings.Replace(env, k + "=", "", -1)
		os.Setenv(k, v)
	}

	return true
}

func (ci *CiJob) GitSetup() bool {
	dir, _ := os.Getwd()
	ci.folder = dir + "/builds/" + ci.Job.Repo.Project

	ci.execute("rm", "-rf", ci.folder)
	if ci.Job.Repo.Project != "" {
		ok := ci.execute("git", "clone", "-b", ci.Job.Repo.Branch, ci.Job.Repo.url(), ci.folder)
		if ok {
			// git setup
			ci.execute("git", "-C", ci.folder, "log", "-n", "1")
		}
		return ok
	} else {
		return true
	}
}

func (ci *CiJob) SetupBuildImage() bool {
	image := ""
	if ci.Job.Build.BaseBuild != "" {
		ok := ci.execute("docker", "build", "-t", ci.Job.JobId, ci.Job.Build.BaseBuild)
		if !ok {
			return false
		}
		image = ci.Job.JobId
	} else if ci.Job.Build.BaseImage != "" {
		ok := ci.execute("docker", "pull", ci.Job.Build.BaseImage)
		if !ok {
			return false
		}
		image = ci.Job.Build.BaseImage
	}

	if image != "" {
		c := Docker{
			Image: image,
			Env: ci.Job.env(),
			WorkDir: "/opt/ci/" + ci.Job.Repo.Project,
			Net: "test-net",
			NetAlias: "main",
			Name: "main",
			FlagI: true,
			FlagD: true,
		}

		ci.execute("docker", "rm", "-f", c.Name)
		ok := ci.execute("docker", c.start()...)
		if !ok {
			return false
		}
		ci.execute("docker", "cp", ci.folder, "main:/opt/ci")
	}

	return true

}

func (ci *CiJob) SetupServices() bool {
	for name, service := range ci.Job.Build.Services {
		ok := ci.execute("docker", "pull", service.Image)
		if !ok {
			return false
		}

		c := Docker{
			Image: service.Image,
			Env: service.Env,
			Net: "test-net",
			NetAlias: name,
			Name: name,
			FlagI: true,
			FlagD: true,
		}

		ok = ci.execute("docker", c.start()...)
		if !ok {
			return false
		}

		for _, cmd := range service.OnStartup {
			ci.execute("docker", "exec", "-i", name, "/bin/sh", "-c", cmd)
		}
	}

	return true
}

func (ci *CiJob) Before() bool {
	for _, cmd := range ci.Job.Build.Before {
		ok := ci.executeCmd(cmd)
		if !ok {
			return false
		}
	}
	return true
}

func (ci *CiJob) Main() bool {
	for _, cmd := range ci.Job.Build.Main {
		ok := ci.executeCmd(cmd)
		if !ok {
			return false
		}
	}
	return true
}

func (ci *CiJob) After() bool {
	for _, cmd := range ci.Job.Build.After {
		ci.executeCmd(cmd)
	}
	return true
}

func (ci *CiJob) OnSuccess() bool {
	if !ci.Status.Failed {
		for _, cmd := range ci.Job.Build.OnSuccess {
			ok := ci.executeCmd(cmd)
			if !ok {
				return false
			}
		}
	}
	return true
}

func (ci *CiJob) OnFailure() bool {
	if ci.Status.Failed {
		for _, cmd := range ci.Job.Build.OnFailure {
			ci.executeCmd(cmd)
		}
	}
	return true
}

func (ci *CiJob) Cleanup() bool {
	ci.execute("rm", "-rf", ci.folder)
	ci.execute("docker", "stop", "main")
	ci.execute("docker", "rm", "-f", "main")
	for name, _ := range ci.Job.Build.Services {
		ci.execute("docker", "stop", name)
		ci.execute("docker", "rm", "-f", name)
	}
	ci.execute("docker", "network", "rm", "test-net")
	return true
}

func (ci *CiJob) Run() {
	ci.start()
	res, _ := json.MarshalIndent(ci.Job, " ", "  ")
	fmt.Printf("%s\n", res)

	ci.run([]Stage{
		ci.Setup,
		ci.GitSetup,
		ci.SetupBuildImage,
		ci.SetupServices,
		ci.Before,
		ci.Main,
	})

	ci.after([]Stage{
		ci.After,
		ci.OnSuccess,
		ci.OnFailure,
	})

	ci.ensure([]Stage{
		ci.Cleanup,
	})

	ci.finish()
}
