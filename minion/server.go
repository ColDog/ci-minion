package main

import (
	"net/http"
)

type Minion struct {
	cancel 		chan bool
}

func (server *Minion) Handle(w http.ResponseWriter, r *http.Request) {
	server.cancel <- true
}

func (server *Minion) Serve()  {
	http.HandleFunc("/", server.Handle)
	http.ListenAndServe(":8000", nil)
}

func (server *Minion) GetNextJob() *Job {
	repo := Repo{
		Branch: "master",
		Provider: "github",
		Organization: "coldog",
		Project: "ci-sample",
	}

	job := NewJob("test1", repo)

	return job
}

func (server *Minion) Run() {
	for {
		job := server.GetNextJob()
		go job.Run()

		select {
		case job.finished:
			continue

		case server.cancel:
			job.Quit()
		}
	}
}

func main() {
}
