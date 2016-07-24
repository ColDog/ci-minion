package runner

import "fmt"

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
