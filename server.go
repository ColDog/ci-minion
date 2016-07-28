package main

import (
	"net/http"
	"fmt"
	"time"
	"log"
	"strings"

	"gopkg.in/amz.v1/s3"
)

type Minion struct {
	app		*App
	cancel 		chan bool
	current		*CiJob
	exitPostBuild 	bool
}

func (minion *Minion) handleCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	minion.cancel <- true
}

func (minion *Minion) viewCurrentState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(minion.current.serialize()))
}

func (minion *Minion) serve()  {
	spl := strings.Split(minion.app.MinionApi, ":")
	port := spl[len(spl) - 1]

	log.Printf("serving on %s", port)

	http.HandleFunc("/cancel", minion.handleCancel)
	http.HandleFunc("/current-state", minion.viewCurrentState)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

func (minion *Minion) next() (JobConfig, bool) {
	conf := struct {
		Job 	JobConfig 	`json:"job"`
	}{}

	minion.app.setAuth("minion:" + minion.app.SimpleCiSecret)
	conf.Job.token = minion.app.SimpleCiSecret

	err := minion.app.post("/minions/jobs", map[string] string{ "worker": minion.app.MinionApi}, &conf)
	if err != nil {
		return conf.Job, false
	} else {
		auth := fmt.Sprintf("minion:%s.%v", minion.app.SimpleCiSecret, conf.Job.UserId)
		minion.app.setAuth(auth)

		res := make(map[string] []struct{
			Key 	string 	`json:"key"`
			Value 	string 	`json:"value"`
		})

		err := minion.app.get("/api/users/me/secrets", nil, &res)
		if err != nil {
			panic(err)
		}

		for _, sec := range res["secrets"] {
			conf.Job.Build.Env = append(conf.Job.Build.Env, fmt.Sprintf("%s=%s", sec.Key, sec.Value))
		}

		return conf.Job, true
	}
}

func (minion *Minion) save() {
	out := minion.current.serialize()
	path := fmt.Sprintf("builds/%s/%s", minion.current.Job.JobFamily, minion.current.Job.JobId)

	err := minion.app.s3.Put(path, out, "application/json", s3.ACL("public-read"))
	if err != nil {
		log.Printf("Could not upload file %v", err)
		// panic(err)
	} else {
		log.Printf("uploaded file to %s", minion.current.Job.JobId)
	}

	minion.app.setAuth("minion:" + minion.app.SimpleCiSecret)
	err = minion.app.patch("/minions/jobs/" + minion.current.Job.JobId, map[string] interface{} {
		"complete": true,
		"cancelled": minion.current.Status.Cancelled,
		"failed": minion.current.Status.Failed,
		"failure": minion.current.Status.Failure,
		"total_time": minion.current.Status.TotalTime,
		"stored_output_url": fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", minion.app.S3Region, minion.app.S3Bucket, path),
	}, nil)

	if err != nil {
		panic(err)
	}
}

func (minion *Minion) run() {
	for {
		cur, ok := minion.next()
		if !ok {
			log.Printf("could not find any new jobs")
			// sleep before starting up again
			time.Sleep(5 * time.Second)
			continue
		}

		minion.current = NewJob(cur)
		go minion.current.Run()

		select {
		case <- minion.current.finished:
			log.Printf("job finished! %s", minion.current.Job.JobId)

		case <- minion.cancel:
			minion.current.cancel()
		}

		minion.save()

		// this allows the scheduler to build up a new fresh image upon completion
		if minion.exitPostBuild {
			return
		}
	}
}

func (server *Minion) Start() {
	go server.serve()
	server.run()
}

