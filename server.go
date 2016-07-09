package main

import (
	"net/http"
	"fmt"
)

type Minion struct {
	cancel 		chan bool
	current		*Job
}

func (server *Minion) handleCancel(w http.ResponseWriter, r *http.Request) {
	server.cancel <- true
}

func (server *Minion) viewJob(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(server.current.JobId))
}

func (server *Minion) viewCurrentState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(server.current.Serialize()))
}

func (server *Minion) serve()  {
	http.HandleFunc("/cancel", server.handleCancel)
	http.HandleFunc("/current", server.viewJob)
	http.HandleFunc("/current-state", server.viewCurrentState)
	http.ListenAndServe("0.0.0.0:8000", nil)
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
		server.current = server.GetNextJob()
		go server.current.Run()

		select {
		case <- server.current.finished:
			out := server.current.Serialize()
			fmt.Printf("\n%s\n", out)
			continue

		case <- server.cancel:
			server.current.Quit()
		}
	}
}

func (server *Minion) Start() {
	go server.serve()
	server.run()
}

func NewMinion() *Minion {
	return &Minion{
		cancel: make(chan bool),
	}
}
