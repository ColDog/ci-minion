package main

import (
	"net/http"
	"fmt"
	"time"
	"log"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
	"os"
	"github.com/go-amz/amz/s3"
	"gopkg.in/amz.v1/aws"
)

type Minion struct {
	cancel 		chan bool
	current		*Job
	api 		string
	host 		string
	token		string
	s3 		*s3.Bucket
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
	err := server.s3.Put(server.current.JobId, out, "application/json", s3.ACL("public-read"))
	if err != nil {
		log.Printf("Could not upload file %v", err)
		panic(err)
	} else {
		log.Printf("uploaded file to %s", server.current.JobId)
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
		log.Printf("Could not patch updates %s %v", server.api, errs[0])
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

func (server *Minion) Start() {
	go server.serve(os.Getenv("MINION_PORT"))
	server.run()
}

func NewMinion() *Minion {
	s3region := os.Getenv("MINION_S3_REGION")
	reg, ok := aws.Regions[s3region]
	if !ok {
		panic(s3region + " is not a region")
	}

	auth := aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_KEY_ID"),
	}

	conn := s3.New(auth, reg)
	bucket := conn.Bucket(os.Getenv("MINION_S3_BUCKET"))

	return &Minion{
		host: os.Getenv("MINION_HOST"),
		api: os.Getenv("MINION_API"),
		token: os.Getenv("MINION_TOKEN"),
		s3: bucket,
		cancel: make(chan bool),
	}
}
