package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// Overall job flow
// 1. clone
//	- clone's the git repo locally, catches any bugs early
// 2. provision build container
// 	- starts up the build container, this is destroyed after every build so secret keys can be copied with confidence
//	- runs setup on the build container
// 3. services
//	- starts up services inside the docker network
//	- runs any setup on those services
// 4. provision
// 	- provisions the build container by building / pulling the image
// 	- copies over files
// 	- starts the build image
// 5. testing
//	- runs before -> main -> after -> (on success | on failure) inside the main image
// 6. post
//	- runs the post -> after -> (post success | post failure) inside our build image
// 7. cleanup
// 	- removes the network, services, build and test image
//

var BuildImg string = "coldog/simpleci-base"
var BuildName string = "simpleci-base"

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
	Organization	string		`json:"owner"`
	Project 	string		`json:"project"`
}

func (repo Repo) Url() string {
	auth := ""
	if repo.AuthUser != "" && repo.AuthPass != "" {
		auth += repo.AuthUser + ":" + repo.AuthPass + "@"
	}

	return fmt.Sprintf("https://%s%s.com/%s/%s.git", auth, repo.Provider, repo.Organization, repo.Project)
}

type Service struct {
	Image 		string		`json:"image"`
	Env 		[]string	`json:"env"`
	OnStartup 	[]string	`json:"on_startup"`
}

type Build struct  {
	Env 		[]string		`json:"env"`
	BaseImage	string			`json:"base_image"`
	BaseBuild	string			`json:"base_build"`
	Services 	map[string] Service     `json:"services"`
	Setup 		[]string		`json:"setup"`
	Before		[]string		`json:"before"`
	Main 		[]string		`json:"main"`
	After 		[]string		`json:"after"`
	OnSuccess	[]string		`json:"on_success"`
	OnFailure 	[]string		`json:"on_failure"`
	Post 		[]string		`json:"post"`
	PostSuccess	[]string		`json:"post_success"`
	PostFailure	[]string		`json:"post_failure"`
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

func (job *Job) execIn(topic string, image string, sh string) bool {
	return job.execute(topic, "docker", "exec", "-i", image,  "/bin/sh", "-c", sh)
}

func (job *Job) run(stages []Stage) bool {
	for i, stage := range stages {
		if job.Cancelled {
			return false
		}

		log.Printf("\033[0;31mstep: %v %s\033[0m", i, FuncName(stage))
		ok := stage()
		if !ok {
			job.Failed = true
			last := job.Commands[len(job.Commands) - 1]
			job.FailureOutput = strings.Join(last.Output, "\n")
			return false
		}
	}

	return true
}

func (job *Job) ensure(stages []Stage) bool {
	ok := true
	for i, stage := range stages {

		log.Printf("\033[0;31mstep: %v %s\033[0m", i, FuncName(stage))
		stageOk := stage()
		if !stageOk {
			ok = false
		}
	}

	return ok
}

func (job *Job) addEnvVars(cmds []string) []string {
	add := []string{
		"-e", "CI_BUILD_ID=" + job.JobId,
		"-e", "CI_MAIN_CONTAINER=" + job.JobId,
		"-e", "CI_BUILD_FAMILY=" + job.JobFamily,
		"-e", "CI_FAILED=" + fmt.Sprintf("%v", job.Failed),
		"-e", "CI_GIT_REPO=" + job.Repo.Project,
		"-e", "CI_GIT_OWNER=" + job.Repo.Organization,
		"-e", "CI_GIT_PROVIDER=" + job.Repo.Provider,
		"-e", "CI_GIT_BRANCH=" + job.Repo.Branch,
	}

	return append(cmds, add...)
}

func (job *Job) StartBuildContainer() bool {
	job.execute("provision_build_container", "docker", "rm", "-f", BuildName)

	cmds := []string{"run", "-i", "-d", "--name=" + BuildName, "-v", "/var/run/docker.sock:/var/run/docker.sock"}
	for _, env := range job.Build.Env {
		cmds = append(cmds, "-e", env)
	}
	cmds = job.addEnvVars(cmds)
	cmds = append(cmds, BuildImg, "bash")

	return job.execute("provision_build_container", "docker", cmds...)
}

func (job *Job) StartServices() bool {
	job.execute("services", "docker", "network", "create", job.JobId + "-network")

	for name, service := range job.Build.Services {
		job.execute("services", "docker", "rm", "-f", name)

		cmds := []string{"run", "-d", "--net=" + job.JobId + "-network", "--name=" + name, "--net-alias=" + name}
		for _, env := range service.Env {
			cmds = append(cmds, "-e", env)
		}
		cmds = job.addEnvVars(cmds)

		cmds = append(cmds, service.Image)
		ok := job.execute("services", "docker", cmds...)
		if !ok {
			return false
		}

		for _, cmd := range service.OnStartup {
			ok := job.execute("services", "docker", "exec", name, "/bin/sh", "-c", cmd)
			if !ok {
				return false
			}
		}
	}

	return true
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

func (job *Job) Provision() bool {
	job.execute("provision", "docker", "rm", "-f", job.JobId)

	isImage := true
	if job.Build.BaseImage != "" {
		job.execute("provision", "docker", "pull", job.Build.BaseImage)
	} else if job.Build.BaseBuild != "" {
		isImage = false

		b := job.Build.BaseBuild
		if job.Build.BaseBuild == "." {
			b = job.BuildFolder
		}

		job.execute("provision", "docker", "build", "-t", job.JobId, b)
	} else {
		log.Println("could not build or pull an image")
		return false
	}

	run := []string{"run", "-it", "-d", "--name=" + job.JobId, "--net=" + job.JobId + "-network", "--net-alias=main"}

	if isImage {
		run = append(run, "-w", "/opt/ci/" + job.Repo.Project)
	}

	for _, e := range job.Build.Env {
		run = append(run, "-e", e)
	}

	run = job.addEnvVars(run)

	if isImage {
		run = append(run, job.Build.BaseImage)
	} else {
		run = append(run, job.JobId)
	}

	ok := job.execute("provision", "docker", run...)
	if !ok {
		return false
	}

	if isImage {
		job.execute("provision", "docker", "cp", job.BuildFolder, job.JobId + ":/opt/ci/")
	}

	return true
}

func (job *Job) RunSetup() bool {
	for _, cmd := range job.Build.Setup {
		ok := job.execIn("setup", BuildName, cmd)
		if !ok {
			return false
		}
	}

	return true
}

func (job *Job) RunPre() bool {
	for _, cmd := range job.Build.Before {
		job.execIn("before", job.JobId, cmd)
	}
	return true
}

func (job *Job) RunMain() bool {
	for _, cmd := range job.Build.Main {
		ok := job.execIn("main", job.JobId, cmd)
		if !ok {
			return false
		}
	}
	return true
}

func (job *Job) RunAfter() bool {
	for _, cmd := range job.Build.After {
		job.execIn("after", job.JobId, cmd)
	}
	return true
}

func (job *Job) RunHooks() bool {
	if !job.Failed {
		for _, cmd := range job.Build.OnSuccess {
			job.execIn("on_success", job.JobId, cmd)
		}
	} else {
		for _, cmd := range job.Build.OnFailure {
			job.execIn("on_failure", job.JobId, cmd)
		}
	}

	return true
}

func (job *Job) RunPost() bool {
	for _, cmd := range job.Build.Post {
		job.execIn("post", BuildName, cmd)
	}

	return true
}

func (job *Job) RunPostHooks() bool {
	if !job.Failed {
		for _, cmd := range job.Build.OnSuccess {
			job.execIn("post_success", BuildName, cmd)
		}
	} else {
		for _, cmd := range job.Build.OnFailure {
			job.execIn("post_failure", BuildName, cmd)
		}
	}

	return true
}

func (job *Job) KillServices() bool {
	for name, _ := range job.Build.Services {
		job.execute("services", "docker", "stop", name)
		job.execute("services", "docker", "rm", name)
	}

	return true
}

func (job *Job) Cleanup() bool {
	job.execute("cleanup", "docker", "stop", job.JobId)
	job.execute("cleanup", "docker", "rm", job.JobId)
	job.execute("cleanup", "rm", "-rf", job.BuildFolder)
	job.execute("cleanup", "docker", "network", "rm", job.JobId + "-network")
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
		job.StartBuildContainer,
		job.RunSetup,
		job.StartServices,
		job.Provision,
		job.RunPre,
		job.RunMain,
	})

	job.ensure([]Stage{
		job.RunAfter,
		job.RunHooks,
		job.RunPostHooks,
		job.RunPost,
		job.Cleanup,
		job.KillServices,
	})

	if !job.Failed && !ok {
		job.Failed = true
	}

	log.Printf("job finished with failure status: %v %s", job.Failed, job.FailureOutput)
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
