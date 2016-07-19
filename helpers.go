package main

import (
	"os/exec"
	"log"
	"strings"
	"runtime"
	"reflect"
	"os"
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

	log.Printf("executing: `%s %s` err: %v", main, strings.Join(args, " "), err)

	//if Config.LogOutput {
		log.Printf("output: %s", output)
	//}

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

func validateEnvVars(vars []string) {
	for _, v := range vars {
		if os.Getenv(v) == "" {
			panic("Could not find environment variable " + v)
		}
	}
}
