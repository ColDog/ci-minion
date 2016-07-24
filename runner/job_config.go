package main

import "fmt"

type RunStatus struct {
	Failed 		bool			`json:"failed"`
	Failure 	string			`json:"failure"`
	Cancelled 	bool			`json:"cancelled"`
	TotalTime 	int64			`json:"total_time"`
	Output 		[]*CommandResult 	`json:"output"`
}


type Repo struct {
	AuthUser 	string			`json:"auth_user"`
	AuthPass	string			`json:"auth_pass"`
	Provider 	string			`json:"provider"`
	Branch 		string			`json:"branch"`
	Organization	string			`json:"owner"`
	Project 	string			`json:"project"`
}

func (repo Repo) url() string {
	auth := ""
	if repo.AuthUser != "" && repo.AuthPass != "" {
		auth += repo.AuthUser + ":" + repo.AuthPass + "@"
	}

	return fmt.Sprintf("https://%s%s.com/%s/%s.git", auth, repo.Provider, repo.Organization, repo.Project)
}

type Service struct {
	Image 		string			`json:"image"`
	Env 		[]string		`json:"env"`
	OnStartup 	[]string		`json:"on_startup"`
}

type Build struct  {
	Env 		[]string		`json:"env"`
	BaseImage	string			`json:"base_image"`
	BaseBuild	string			`json:"base_build"`
	Services 	map[string] Service     `json:"services"`
	Before		[]string		`json:"before"`
	Main 		[]string		`json:"main"`
	After 		[]string		`json:"after"`
	OnSuccess	[]string		`json:"on_success"`
	OnFailure 	[]string		`json:"on_failure"`
}

type JobConfig struct {
	JobId 		string			`json:"job_id"`
	JobFamily	string			`json:"job_family"`
	UserId		int			`json:"user_id"`
	Build 		Build			`json:"build"`
	Repo 		Repo			`json:"repo"`
	token 		string
}

func (job JobConfig) env() []string {
	e := []string{
		"CI_BUILD_ID=" + job.JobId,
		"CI_MAIN_CONTAINER=main",
		"CI_BUILD_FAMILY=" + job.JobFamily,
		"CI_GIT_REPO=" + job.Repo.Project,
		"CI_GIT_OWNER=" + job.Repo.Organization,
		"CI_GIT_PROVIDER=" + job.Repo.Provider,
		"CI_GIT_BRANCH=" + job.Repo.Branch,
		"SIMPLECI_KEY=minion",
		"SIMPLECI_SECRET=" + fmt.Sprintf("%s.%v", job.token, job.UserId),
	}

	e = append(e, job.Build.Env...)
	return e
}
