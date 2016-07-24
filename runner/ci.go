package runner

import "os"

type CiJob struct {
	Runner
	Job 		JobConfig
	folder 		string
}

func NewJob(job JobConfig)  {
	return &CiJob{Job: job}
}

func (ci *CiJob) GitSetup() bool {
	dir, _ := os.Getwd()
	ci.folder = dir

	if ci.Job.Repo.Project != "" {
		ok := ci.execute("git", "clone", "-b", ci.Job.Repo.Branch, ci.Job.Repo.url(), dir)
		if ok {
			// git setup
			ci.execute("git", "-C", dir, "log", "-n", "1")
		}
		return ok
	} else {
		return true
	}
}

func (ci *CiJob) SetupBuildImage() bool {
	image := ""
	if ci.Job.Build.BaseBuild {
		ok := ci.execute("docker", "build", "-t", ci.Job.JobId, ci.Job.Build.BaseBuild)
		if !ok {
			return false
		}
		image = ci.Job.JobId
	} else if ci.Job.Build.BaseImage {
		ok := ci.execute("docekr", "pull", ci.Job.Build.BaseImage)
		if !ok {
			return false
		}
		image = ci.Job.Build.BaseImage
	}

	if image != "" {
		c := Docker{
			Image: image,
			Env: ci.Job.env(),
			WorkDir: "/opt/ci",
			Net: Config.CiNet,
			NetAlias: "main",
			Name: "main",
		}

		ok := ci.execute("docker", c.start()...)
		if !ok {
			return false
		}
		ci.execute("docker", "cp", "main", ci.folder, "/opt/ci")
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
			WorkDir: "/opt/ci",
			Net: Config.CiNet,
			NetAlias: name,
			Name: name,
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
