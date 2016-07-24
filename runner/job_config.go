package runner

import "fmt"

var Config struct{
	MinionToken 	string
	CiNet 		string
}

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
}

func (job JobConfig) env() []string {
	return []string{
		"-e", "CI_BUILD_ID=" + job.JobId,
		"-e", "CI_MAIN_CONTAINER=" + job.JobId,
		"-e", "CI_BUILD_FAMILY=" + job.JobFamily,
		"-e", "CI_GIT_REPO=" + job.Repo.Project,
		"-e", "CI_GIT_OWNER=" + job.Repo.Organization,
		"-e", "CI_GIT_PROVIDER=" + job.Repo.Provider,
		"-e", "CI_GIT_BRANCH=" + job.Repo.Branch,
		"-e", "SIMPLECI_KEY=minion",
		"-e", "SIMPLECI_SECRET=" + fmt.Sprintf("%s.%v", Config.MinionToken, job.UserId),
	}
}
