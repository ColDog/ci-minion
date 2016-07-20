package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type BuildConfig struct {
	Key 		string 		`json:"key"`
	Build 		Build		`json:"build"`
	Repo 		Repo		`json:"repo"`
}

type Stage func() bool

type Repo struct {
	AuthUser 	string		`json:"auth_user"`
	AuthPass	string		`json:"auth_pass"`
	Provider 	string		`json:"provider"`
	Branch 		string		`json:"branch"`
	Organization	string		`json:"org"`
	Project 	string		`json:"project"`
}

func (repo Repo) Url() string {
	auth := ""
	if repo.AuthUser != "" {
		auth += repo.AuthUser + ":" + repo.AuthPass + "@"
	}

	return fmt.Sprintf("https://%s%s.com/%s/%s.git", auth, repo.Provider, repo.Organization, repo.Project)
}

type Service struct {
	Name 		string 		`json:"name"`
	Image 		string		`json:"image"`
	Env 		[]string	`json:"env"`
	OnStartup 	[]string	`json:"on_startup"`
}

type Build struct  {
	Env 		[]string	`json:"env"`
	BaseImage	string		`json:"base_image"`
	Services 	[]Service       `json:"services"`
	Before		[]string	`json:"before"`
	Main 		[]string	`json:"main"`
	After 		[]string	`json:"after"`
	OnSuccess	[]string	`json:"on_success"`
	OnFailure 	[]string	`json:"on_failure"`
}

type Job struct {
	JobId		string
	JobFamily 	string
	Repo 		Repo
	BuildFolder 	string
	Build		Build
	Failed 		bool
	Cancelled 	bool
	FailureOutput 	string
	Commands 	[]*CommandResult
	quit 		chan bool
	finished 	chan bool
}

func NewJob(id string, repo Repo, build Build) *Job {
	return &Job{
		JobId: id,
		Repo: repo,
		Build: build,
		Commands: make([]*CommandResult, 0),
		quit: make(chan bool, 1),
		finished: make(chan bool),
	}
}

func (job *Job) add(topic string, res *CommandResult) {
	job.Commands = append(job.Commands, res)
}

func (job *Job) execute(topic string, main string, args ...string) bool {
	output := make(chan string, 10)

	res := &CommandResult{
		Topic: topic,
		Args: strings.Join(args, " "),
		Error: nil,
		Output: make([]string, 0),
	}
	job.add(topic, res)

	go func() {
		for out := range output {
			res.Output = append(res.Output, out)
		}
	}()

	res.Error = execute(job.quit, output, main, args...)
	return res.Error == nil
}

func (job *Job) execInside(topic string, main string, args ...string) bool {
	cmds := []string{"exec", job.JobId, main}
	cmds = append(cmds, args...)
	return job.execute(topic, "docker", cmds...)
}

func (job *Job) execInsideSh(topic string, sh string) bool {
	return job.execInside(topic, "/bin/sh", "-c", sh)
}

func (job *Job) run(stages []Stage) bool {
	for i, stage := range stages {
		if job.Cancelled {
			return false
		}

		log.Printf("step: %v %s", i, FuncName(stage))
		ok := stage()
		if !ok {
			job.Failed = true
			return false
		}
	}

	return true
}

func (job *Job) StartServices() bool {
	job.execute("services", "docker", "network", "create", job.JobId + "-network")

	for _, service := range job.Build.Services {
		job.execute("services", "docker", "stop", service.Name)
		job.execute("services", "docker", "rm", service.Name)

		cmds := []string{"run", "-d", "--net=" + job.JobId + "-network", "--name=" + service.Name, "--net-alias=" + service.Name}
		for _, env := range service.Env {
			cmds = append(cmds, "-e", env)
		}

		cmds = append(cmds, service.Image)
		ok := job.execute("services", "docker", cmds...)
		if !ok {
			return false
		}

		for _, cmd := range service.OnStartup {
			ok := job.execute("services", "docker", "exec", service.Name, "/bin/sh", "-c", cmd)
			if !ok {
				return false
			}
		}
	}

	return true
}

func (job *Job) KillServices() bool {
	for _, service := range job.Build.Services {
		job.execute("services", "docker", "stop", service.Name)
		job.execute("services", "docker", "rm", service.Name)
	}

	return true
}

func (job *Job) Provision() bool {
	job.execute("provision", "docker", "stop", job.JobId)
	job.execute("provision", "docker", "rm", job.JobId)
	job.execute("provision", "docker", "pull", job.Build.BaseImage)

	run := []string{"run", "-it", "-d", "-v", job.BuildFolder + ":/opt/ci", "--name=" + job.JobId, "--net=" + job.JobId + "-network", "--net-alias=main"}
	for _, e := range job.Build.Env {
		run = append(run, "-e", e)
	}
	run = append(run, job.Build.BaseImage)

	return job.execute("provision", "docker", run...)
}

func (job *Job) Clone() bool {
	cwd, _ := os.Getwd()
	job.BuildFolder = cwd + "/builds/" + job.Repo.Project
	job.execute("clone", "rm", "-rf", job.BuildFolder)
	if job.Repo.Project != "" {
		return job.execute("clone", "git", "clone", "-b", job.Repo.Branch, job.Repo.Url(), job.BuildFolder)
	} else {
		return true
	}
}

func (job *Job) RunPre() bool {
	for _, cmd := range job.Build.Before {
		job.execInsideSh("before", cmd)
	}
	return true
}

func (job *Job) RunMain() bool {
	for _, cmd := range job.Build.Main {
		ok := job.execInsideSh("main", cmd)
		if !ok {
			return false
		}
	}
	return true
}

func (job *Job) RunPost() bool {
	for _, cmd := range job.Build.After {
		job.execInsideSh("after", cmd)
	}
	return true
}

func (job *Job) RunHooks() bool {
	if !job.Failed {
		for _, cmd := range job.Build.OnSuccess {
			ok := job.execInsideSh("on_success", cmd)
			if !ok {
				return false
			}
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
		job.Provision,
		job.RunPre,
		job.RunMain,
	})
}

func (job *Job) Run() {
	ok := job.run([]Stage{
		job.Clone,
		job.StartServices,
		job.Provision,
		job.RunPre,
		job.RunMain,
		job.RunPost,
		job.RunHooks,
		job.Cleanup,
		job.KillServices,
	})

	if !job.Failed && !ok {
		job.Failed = true
	}

	job.finished <- true
}

func (job *Job) Wait() {
	<- job.finished
}

func (job *Job) Quit() {
	job.Cancelled = true
	job.quit <- true
	job.Cleanup()
}

func (job *Job) Serialize() []byte {
	res, err := json.MarshalIndent(job, " ", "  ")
	if err != nil {
		panic(err)
	}
	return res
}
