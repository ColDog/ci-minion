package minion

import (
	"net/http"
	"fmt"
)

type Minion struct {
	cancel 		chan bool
}

func (server *Minion) handle(w http.ResponseWriter, r *http.Request) {
	server.cancel <- true
}

func (server *Minion) serve()  {
	http.HandleFunc("/cancel", server.handle)
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

func (server *Minion) run() {
	for {
		job := server.GetNextJob()
		go job.Run()

		select {
		case job.finished:
			out := job.Serialize()
			fmt.Printf("\n%s\n", out)
			continue

		case server.cancel:
			job.Quit()
		}
	}
}

func (server *Minion) Start() {
	go server.serve()
	server.run()
}

func NewMinion() *Minion {
	return Minion{
		cancel: make(chan bool),
	}
}
