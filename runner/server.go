package runner

import (
	"net/http"
	"fmt"
	"time"
	"log"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
	"github.com/go-amz/amz/s3"
	"gopkg.in/amz.v1/aws"
	"strings"
)

type MinionConfig struct {
	MinionApi 	string 		`json:"minion_api"`
	SimpleCiApi 	string 		`json:"simpleci_api"`
	MinionToken 	string 		`json:"minion_token"`
	AwsAccessKey 	string 		`json:"aws_access_key"`
	AwsSecretKey 	string 		`json:"aws_secret_key"`
	S3Region 	string 		`json:"s3_region"`
	S3Bucket 	string 		`json:"s3_bucket"`
}

type Minion struct {
	cancel 		chan bool
	current		*CiJob
	api 		string
	hostapi 	string
	token		string
	s3bucket 	string
	exitAfterBuild 	bool
	s3 		*s3.Bucket
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
	spl := strings.Split(minion.hostapi, ":")
	port := spl[len(spl) - 1]

	log.Printf("serving on %s", port)

	http.HandleFunc("/cancel", minion.handleCancel)
	http.HandleFunc("/current-state", minion.viewCurrentState)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

func (minion *Minion) next() (JobConfig, bool) {
	conf := JobConfig{}
	conf.token = minion.token

	req := gorequest.New().
		Post(minion.api + "/minions/jobs").
		Param("worker", minion.hostapi).
		Set("Authorization", fmt.Sprintf("minion:%s", minion.token))

	r, body, errs := req.End()

	if errs != nil && len(errs) > 0 {
		panic(errs[0])
	}

	if r.StatusCode != 200 {
		return conf, false
	}

	err := json.Unmarshal([]byte(body), conf)
	if err != nil {
		panic(err)
	}

	return conf, true
}

func (minion *Minion) save() {
	out := minion.current.serialize()
	path := fmt.Sprintf("builds/%s/%s", minion.current.Job.JobFamily, minion.current.Job.JobId)

	err := minion.s3.Put(path, out, "application/json", s3.ACL("public-read"))
	if err != nil {
		log.Printf("Could not upload file %v", err)
		// panic(err)
	} else {
		log.Printf("uploaded file to %s", minion.current.Job.JobId)
	}

	_, _, errs := gorequest.New().
		Patch(minion.api + "/minions/jobs/" + minion.current.Job.JobId).
		Param("complete", fmt.Sprintf("%v", true)).
		Param("cancelled", fmt.Sprintf("%v", minion.current.Status.Cancelled)).
		Param("failed", fmt.Sprintf("%v", minion.current.Status.Failed)).
		Param("failure", minion.current.Status.Failure).
		Param("total_time", fmt.Sprintf("%v", minion.current.Status.TotalTime)).
		Param("stored_output_url", fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", minion.s3.Region.Name, minion.s3bucket, path)).
		Set("Authorization", fmt.Sprintf("minion:%s.%v", minion.token, minion.current.Job.UserId)).
		End()

	if len(errs) > 0 {
		log.Printf("Could not patch updates %s %v", minion.api, errs[0])
	} else {
		log.Println("saved!")
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

		if minion.exitAfterBuild {
			return
		}
	}
}

func (server *Minion) Start() {
	go server.serve()
	server.run()
}

func NewMinion(conf MinionConfig) *Minion {
	reg, ok := aws.Regions[conf.S3Region]
	if !ok {
		panic(conf.S3Region + " is not a region")
	}

	auth := aws.Auth{
		AccessKey: conf.AwsAccessKey,
		SecretKey: conf.AwsSecretKey,
	}

	conn := s3.New(auth, reg)
	bucket := conn.Bucket(conf.S3Bucket)

	return &Minion{
		hostapi: conf.MinionApi,
		api: conf.SimpleCiApi,
		token: conf.MinionToken,
		s3: bucket,
		s3bucket: conf.S3Bucket,
		exitAfterBuild: false,
		cancel: make(chan bool),
	}
}
