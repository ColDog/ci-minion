package main

import (
	"os/exec"
	"log"
	"strings"
	"runtime"
	"reflect"
	"github.com/parnurzeal/gorequest"
	"encoding/json"
)

type CommandResult struct {
	Args 		[]string
	Output		string
	Error 		error
}

func execute(quit chan bool, main string, args ...string) CommandResult {
	done := make(chan bool, 1)
	cmd := exec.Command(main, args...)

	go func() {
		select {
		case <- quit:
			cmd.Process.Kill()
			return
		case <- done:
			return
		}
	}()

	output, err := cmd.CombinedOutput()
	done <- true

	log.Printf("executing: %s %s err: %v", main, strings.Join(args, " "), err)

	if LOG_OUTPUT {
		log.Printf("output: %s", output)
	}

	return CommandResult{
		Args: args,
		Error: err,
		Output: string(output),
	}
}

func FuncName(i interface{}) string {
	s := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), ".")
	sa := strings.ToLower(s[len(s) - 1])
	return strings.Split(sa, "-")[0]
}

func patch(url string, res interface{}, params map[string] interface{}) error {
	_, body, errs := gorequest.New().Patch(url).Send(params).End()
	if len(errs) > 0 {
		return errs[0]
	}

	if res == nil {
		return nil
	}
	err := json.Unmarshal([]byte(body), res)
	return err
}
