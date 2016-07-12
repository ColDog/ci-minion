package main

import (
	"net/http"
	"fmt"
	"time"
	"log"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
	"os"
)

type Minion struct {
	cancel 		chan bool
	current		*Job
	api 		string
	host 		string
	token		string
	s3 		*S3Client
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

func (server *Minion) next() (*Job, bool) {
	conf := &struct {
		Job 	BuildConfig 	`json:"job"`
	}{BuildConfig{}}

	req := gorequest.New().
		Post(server.api + "/minions/jobs").
		Param("worker", server.host).
		Param("token", server.token)

	_, body, errs := req.End()
	if len(errs) > 0 {
		return &Job{}, false
	}

	err := json.Unmarshal([]byte(body), conf)
	if err != nil {
		return &Job{}, false
	}

	if conf.Job.Key != "" {
		job := NewJob(conf.Job.Key, conf.Job.Repo, conf.Job.Build)
		return job, true
	} else {
		return &Job{}, false
	}
}

func (server *Minion) save() {
	out := server.current.Serialize()
	err := server.s3.Upload(server.current.JobId, out)
	if err != nil {
		panic(err)
	}

	_, _, errs := gorequest.New().
		Patch(server.api + "/minions/jobs/" + server.current.JobId).
		Param("complete", fmt.Sprintf("%v", true)).
		Param("cancelled", fmt.Sprintf("%v", server.current.Cancelled)).
		Param("failed", fmt.Sprintf("%v", server.current.Failed)).
		Param("failure", server.current.FailureOutput).
		Param("token", server.token).
		End()

	if len(errs) > 0 {
		panic(errs[0])
	}
}

func (server *Minion) run() {
	for {
		cur, ok := server.next()
		if !ok {
			log.Printf("could not find any new jobs")
			// sleep before starting up again
			time.Sleep(5 * time.Second)
			continue
		}

		server.current = cur
		go server.current.Run()

		select {
		case <- server.current.finished:
			log.Printf("job finished! %s", server.current.JobId)

		case <- server.cancel:
			server.current.Quit()
		}

		server.save()

		// sleep before starting up again
		time.Sleep(5 * time.Second)
	}
}

func (server *Minion) Start(port string) {
	go server.serve(port)
	server.run()
}

func NewMinion(api, host, token, s3bucket string) *Minion {
	s3key := os.Getenv("AWS_ACCESS_KEY_ID")
	s3secret := os.Getenv("AWS_SECRET_KEY_ID")

	return &Minion{
		host: host,
		api: api,
		token: token,
		s3: NewS3Client(s3key, s3secret, s3bucket),
		cancel: make(chan bool),
	}
}
