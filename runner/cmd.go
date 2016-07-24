package runner

import (
	"time"
	"bufio"
	"fmt"
	"strings"
	"runtime"
	"reflect"
	"os/exec"
	"log"
)

type CommandResult struct {
	Topic 		string		`json:"topic"`
	Args 		string		`json:"args"`
	Time 		int64		`json:"time"`
	Output		[]string	`json:"output"`
	Error 		error		`json:"error"`
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

func funcName(i interface{}) string {
	s := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	spl := strings.Split(s, ".")
	s = strings.Split(spl[len(spl) - 1], "-")[0]
	s = strings.Replace(s, ")", "", -1)
	s = strings.Replace(s, "(", "", -1)
	return s
}