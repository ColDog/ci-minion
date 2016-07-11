package main

import (
	"net/http"
	"fmt"
	"time"
	"log"
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
	conf := struct {
		Job 	BuildConfig 	`json:"job"`
	}{
		Job: 	BuildConfig{},
	}

	err := post(server.api + "/minions/jobs", conf, map[string] interface{} {
		"worker": server.host,
		"token": SECRET,
	})
	if err != nil {
		log.Printf("error: %v", err)
	}

	job := NewJob(conf.Job.Key, conf.Job.Repo, conf.Job.Build)
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

		out := server.current.Serialize()
		fmt.Printf("\n%s\n", out)

		// update the app
		res := make(map[string] interface{})
		patch(server.api + "/minions/jobs/" + server.current.JobId, res, map[string] interface{} {
			"completed": true,
			"cancelled": server.current.Cancelled,
			"failed": server.current.Failed,
			"failure": server.current.FailureOutput,
			"token": SECRET,
		})

		// sleep before starting up again
		time.Sleep(5 * time.Second)
	}
}

func (server *Minion) Start(port string) {
	go server.serve(port)
	server.run()
}

func NewMinion(api, host string) *Minion {
	return &Minion{
		host: host,
		api: api,
		cancel: make(chan bool),
	}
}
