package main

import (
	"net/http"
	"fmt"
	"time"
	"log"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
	"github.com/go-amz/amz/s3"
	"gopkg.in/amz.v1/aws"
)

type Minion struct {
	cancel 		chan bool
	current		*Job
	api 		string
	hostapi 	string
	token		string
	s3 		*s3.Bucket
}

func (server *Minion) handleCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	server.cancel <- true
}

func (server *Minion) viewJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(server.current.JobId))
}

func (server *Minion) viewCurrentState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(server.current.Serialize()))
}

func (server *Minion) serve()  {
	http.HandleFunc("/cancel", server.handleCancel)
	http.HandleFunc("/current", server.viewJob)
	http.HandleFunc("/current-state", server.viewCurrentState)
	http.ListenAndServe(server.hostapi, nil)
}

func (server *Minion) next() (*Job, bool) {
	conf := &struct {
		Job 	BuildConfig 	`json:"job"`
	}{BuildConfig{}}

	req := gorequest.New().
		Post(server.api + "/minions/jobs").
		Param("worker", server.hostapi).
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
	path := fmt.Sprintf("builds/%s/%s", server.current.JobFamily, server.current.JobId)
	err := server.s3.Put(path, out, "application/json", s3.ACL("public-read"))
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
		Param("stored_output_url", fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", server.s3.Region.Name, Config.S3Bucket, path)).
		End()

	if len(errs) > 0 {
		log.Printf("Could not patch updates %s %v", server.api, errs[0])
		//panic(errs[0])
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
	go server.serve()
	server.run()
}

func NewMinion() *Minion {
	reg, ok := aws.Regions[Config.S3Region]
	if !ok {
		panic(Config.S3Region + " is not a region")
	}

	auth := aws.Auth{
		AccessKey: Config.AwsAccessKeyID,
		SecretKey: Config.AwsSecretKeyID,
	}

	conn := s3.New(auth, reg)
	bucket := conn.Bucket(Config.S3Bucket)

	return &Minion{
		hostapi: Config.MinionApi,
		api: Config.SimpleCiApi,
		token: Config.MinionToken,
		s3: bucket,
		cancel: make(chan bool),
	}
}
