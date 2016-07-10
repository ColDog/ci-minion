package main

import (
	"net/http"
	"fmt"
	"time"
)

type Minion struct {
	cancel 		chan bool
	current		*Job
	api 		string
	host 		string
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

func (server *Minion) serve(port string)  {
	http.HandleFunc("/cancel", server.handleCancel)
	http.HandleFunc("/current", server.viewJob)
	http.HandleFunc("/current-state", server.viewCurrentState)
	http.ListenAndServe("0.0.0.0:" + port, nil)
}

func (server *Minion) next() *Job {
	data, err := post(server.api + "/jobs", map[string] interface{} {
		"worker": server.host,
	})
	if err != nil {
		panic(err)
	}

	repo := Repo{
		Branch: data["branch"].(string),
		Provider: data["provider"].(string),
		Organization: data["org"].(string),
		Project: data["project"].(string),
	}

	job := NewJob(data["key"].(string), repo)
	return job
}

func (server *Minion) run() {
	for {
		server.current = server.next()
		go server.current.Run()

		select {
		case <- server.current.finished:
			continue

		case <- server.cancel:
			server.current.Quit()
		}

		// todo: save the output to permanent storage
		out := server.current.Serialize()
		fmt.Printf("\n%s\n", out)

		// update the app
		patch(server.api + "/jobs/" + server.current.JobId, map[string] interface{} {
			"completed": true,
			"cancelled": server.current.Cancelled,
			"failed": server.current.Failed,
			"failure": server.current.FailureOutput,
		})

		// sleep before starting up again
		time.Sleep(5 * time.Second)
	}
}

func (server *Minion) Start(port string) {
	go server.serve(port)
	server.run()
}

func NewMinion() *Minion {
	return &Minion{
		cancel: make(chan bool),
	}
}
