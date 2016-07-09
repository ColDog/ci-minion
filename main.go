package main

import "time"

const LOG_OUTPUT bool = false


func main()  {
	//minion := NewMinion()
	//minion.Start()

	repo := Repo{
		Branch: "master",
		Provider: "github",
		Organization: "coldog",
		Project: "ci-sample",
	}

	job := NewJob("test1", repo)

	go func() {
		time.Sleep(10 * time.Second)
		job.Quit()
	}()

	job.Run()
}
