package main

import (
	"os/exec"
	"strings"
	"runtime"
	"reflect"
	"os"
	"log"
	"bufio"
	"fmt"
	"time"
)

type CommandResult struct {
	Topic 		string
	Args 		string
	Time 		int64
	Output		[]string
	Error 		error
}

func execute(quit chan bool, output chan string, main string, args ...string) (error, int64) {
	t1 := time.Now().Unix()

	done := make(chan bool, 1)
	cmd := exec.Command(main, args...)

	log.Printf("exec: %s %s", main, strings.Join(args, " "))

	go func() {
		select {
		case <- quit:
			cmd.Process.Kill()
			return
		case <- done:
			return
		}
	}()

	// capture the output and error pipes
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		t2 := time.Now().Unix()
		return err, t2 - t1
	}

	go func() {
		buff := bufio.NewScanner(stderr)

		for buff.Scan() {
			fmt.Printf("	> %s\n", buff.Text())
			output <- buff.Text()
		}
	}()

	go func() {
		buff := bufio.NewScanner(stdout)

		for buff.Scan() {
			fmt.Printf("	> %s\n", buff.Text())
			output <- buff.Text()
		}
	}()

	err = cmd.Wait()
	done <- true
	t2 := time.Now().Unix()
	return err, t2 - t1
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
