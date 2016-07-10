package main

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"os"
)

type Stage func() bool

type Repo struct {
	AuthUser 	string
	AuthPass	string
	Provider 	string
	Branch 		string
	Organization	string
	Project 	string
}

func (repo Repo) Url() string {
	auth := ""
	if repo.AuthUser != "" {
		auth += repo.AuthUser + ":" + repo.AuthPass + "@"
	}

	return fmt.Sprintf("https://%s%s.com/%s/%s.git", auth, repo.Provider, repo.Organization, repo.Project)
}

type Build struct  {
	BaseImage	string		`yaml:"base_image"`
	PreTest		[]string	`yaml:"pre_test"`
	Test 		[]string	`yaml:"test"`
	PostTest 	[]string	`yaml:"post_test"`
	OnSuccess	[]string	`yaml:"on_success"`
	OnFailure 	[]string	`yaml:"on_failure"`
}

type Job struct {
	JobId		string
	Repo 		Repo
	BuildFolder 	string
	Build		Build
	Failed 		bool
	Cancelled 	bool
	FailureOutput 	string
	Commands 	map[string] []CommandResult
	quit 		chan bool
	finished 	chan bool
}

func NewJob(id string, repo Repo) *Job {
	return &Job{
		JobId: id,
		Repo: repo,
		Commands: make(map[string] []CommandResult),
		quit: make(chan bool, 1),
		finished: make(chan bool),
	}
}

func (job *Job) add(topic string, res CommandResult) {
	if _, ok := job.Commands[topic]; !ok {
		job.Commands[topic] = make([]CommandResult, 1)
	}
	job.Commands[topic] = append(job.Commands[topic], res)
}

func (job *Job) execute(topic string, main string, args ...string) bool {
	res := execute(job.quit, main, args...)
	job.add(topic, res)
	return res.Error == nil
}

func (job *Job) execInside(topic string, main string, args ...string) bool {
	cmds := []string{"exec", job.JobId, main}
	cmds = append(cmds, args...)
	res := execute(job.quit, "docker", cmds...)
	job.add(topic, res)
	return res.Error == nil
}

func (job *Job) execInsideSh(topic string, sh string) bool {
	res := execute(job.quit, "docker", "exec", job.JobId, "/bin/sh", "-c", sh)
	job.add(topic, res)
	return res.Error == nil
}

func (job *Job) execInsideShOut(topic string, sh string) (string, bool) {
	res := execute(job.quit, "docker", "exec", job.JobId, "/bin/sh", "-c", sh)
	job.add(topic, res)
	return res.Output, res.Error == nil
}

func (job *Job) run(stages []Stage) bool {
	for i, stage := range stages {
		if job.Cancelled {
			return false
		}

		log.Printf("step: %v %s", i, FuncName(stage))
		ok := stage()
		if !ok {
			return false
		}
	}

	return true
}

func (job *Job) GetBuild() bool {
	data, err := ioutil.ReadFile(job.BuildFolder + "/ci.yml")
	if err == nil {
		b := Build{}
		yaml.Unmarshal(data, &b)
		job.Build = b
		log.Printf("build: %+v", b)
		return true
	}

	return true
}

func (job *Job) Provision() bool {
	job.execute("provision", "docker", "stop", job.JobId)
	job.execute("provision", "docker", "rm", job.JobId)
	job.execute("provision", "docker", "pull", job.Build.BaseImage)
	return job.execute("provision", "docker", "run", "-it", "-d", "-v", job.BuildFolder + ":/opt/ci", "--name=" + job.JobId, job.Build.BaseImage)
}

func (job *Job) Clone() bool {
	cwd, _ := os.Getwd()
	job.BuildFolder = cwd + "/builds/" + job.Repo.Project
	job.execute("clone", "rm", "-rf", job.BuildFolder)
	return job.execute("clone", "git", "clone", "-b", job.Repo.Branch, job.Repo.Url(), job.BuildFolder)
}

func (job *Job) RunPre() bool {
	for _, cmd := range job.Build.PreTest {
		job.execInsideSh("pre_test", cmd)
	}
	return true
}

func (job *Job) RunTests() bool {
	for _, cmd := range job.Build.Test {
		out, ok := job.execInsideShOut("test", cmd)
		if !ok {
			job.Failed = true
			job.FailureOutput = out
			return true
		}
	}
	return true
}

func (job *Job) RunPost() bool {
	for _, cmd := range job.Build.PostTest {
		job.execInsideSh("post_test", cmd)
	}
	return true
}

func (job *Job) RunHooks() bool {
	if !job.Failed {
		for _, cmd := range job.Build.OnSuccess {
			job.execInsideSh("on_success", cmd)
		}
	} else {
		for _, cmd := range job.Build.OnFailure {
			job.execInsideSh("on_failure", cmd)
		}
	}

	return true
}

func (job *Job) Cleanup() bool {
	job.execute("cleanup", "docker", "stop", job.JobId)
	job.execute("cleanup", "docker", "rm", job.JobId)
	job.execute("cleanup", "rm", "-rf", job.BuildFolder)
	return true
}

func (job *Job) Sandbox() {
	job.run([]Stage{
		job.Clone,
		job.GetBuild,
		job.Provision,
		job.RunPre,
		job.RunTests,
	})
}

func (job *Job) Run() {
	ok := job.run([]Stage{
		job.Clone,
		job.GetBuild,
		job.Provision,
		job.RunPre,
		job.RunTests,
		job.RunPost,
		job.RunHooks,
		job.Cleanup,
	})

	if !job.Failed && !ok {
		job.Failed = true
	}
}

func (job *Job) Quit() {
	job.Cancelled = true
	job.quit <- true
	job.Cleanup()
}

func (job *Job) Serialize() []byte {
	res, err := json.Marshal(job)
	if err != nil {
		panic(err)
	}
	return res
}
