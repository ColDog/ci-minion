package runner

import (
	"log"
	"strings"
	"time"
	"encoding/json"
)

type Runner struct {
	finished 	chan bool
	quit 		chan bool
	topic 		string
	startTime 	int64
	Status 		RunStatus
}
type Stage func() bool

func (runner *Runner) run(stages []Stage) bool {
	for i, stage := range stages {
		if runner.Status.Cancelled {
			return false
		}

		runner.topic = funcName(stage)
		log.Printf("\033[0;31mstep: %v %s\033[0m", i, runner.topic)
		ok := stage()
		if !ok {
			runner.Status.Failed = true
			last := runner.Status.Output[len(runner.Status.Output) - 1]
			runner.Status.Failure = strings.Join(last.Output, "\n")
			return false
		}
	}

	return true
}

func (runner *Runner) after(stages []Stage) bool {
	ok := true
	for i, stage := range stages {
		if runner.Status.Cancelled {
			return false
		}

		runner.topic = funcName(stage)
		log.Printf("\033[0;31mstep: %v %s\033[0m", i, runner.topic)
		stageOk := stage()
		if !stageOk {
			ok = false
		}
	}

	return ok
}

func (runner *Runner) ensure(stages []Stage) bool {
	for i, stage := range stages {
		runner.topic = funcName(stage)
		log.Printf("\033[0;31mstep: %v %s\033[0m", i, runner.topic)
		stage()
	}

	return true
}

func (runner *Runner) execute(main string, args ...string) bool {
	output := make(chan string, 10)

	res := &CommandResult{
		Topic: runner.topic,
		Args: main + " " + strings.Join(args, " "),
		Error: nil,
		Output: make([]string, 0),
	}

	runner.Status.Output = append(runner.Status.Output, res)
	go func() {
		for out := range output {
			res.Output = append(res.Output, out)
		}
	}()

	err, t := execute(runner.quit, output, main, args...)
	res.Error = err
	res.Time = t
	return res.Error == nil
}

func (runner *Runner) executeCmd(cmd string) bool {
	return runner.execute("/bin/sh", "-c", cmd)
}

func (runner *Runner) wait()  {
	<- runner.finished
}

func (runner *Runner) start() {
	runner.startTime = time.Now().Unix()
}

func (runner *Runner) finish() {
	runner.Status.TotalTime = time.Now().Unix() - runner.startTime
	runner.finished <- true
}

func (runner *Runner) cancel() {
	runner.quit <- true
	runner.Status.Cancelled = true
	runner.finish()
}

func (runner *Runner) serialize() []byte {
	res, err := json.MarshalIndent(runner.Status, " ", "  ")
	if err != nil {
		panic(err)
	}
	return res
}
