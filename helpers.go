package main

import (
	"os/exec"
	"log"
	"strings"
	"fmt"
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
		fmt.Printf("output: %s", output)
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

func post(url string, res interface{}, params map[string] interface{}) error {
	request := gorequest.New()
	post := request.Post(url)

	for key, val := range params {
		post.Param(key, fmt.Sprintf("%v", val))
	}
	_, body, errs := post.End()
	if len(errs) > 0 {
		return errs[0]
	}

	err := json.Unmarshal([]byte(body), &res)
	return err
}

func patch(url string, params map[string] interface{}, res interface{}) error {
	request := gorequest.New()
	post := request.Patch(url)

	for key, val := range params {
		post.Param(key, fmt.Sprintf("%v", val))
	}
	_, body, errs := post.End()
	if len(errs) > 0 {
		return errs[0]
	}

	err := json.Unmarshal([]byte(body), &res)
	return err
}
