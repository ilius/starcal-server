package genlib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

func RunCommand3(name string, args ...string) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			exitCode = 1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	stdout = strings.TrimSpace(stdout)
	stderr = strings.TrimSpace(stderr)
	return
}

func FormatGoFile(fpath string, waitGroup *sync.WaitGroup, useGoreturns bool) {
	if waitGroup != nil {
		defer waitGroup.Done()
	}
	fmt.Println("formatting", fpath)
	var cmdParts []string
	if useGoreturns {
		cmdParts = []string{"goreturns", "-w", fpath}
	} else {
		cmdParts = []string{"go", "fmt", fpath}
	}
	stdout, stderr, exitCode := RunCommand3(cmdParts[0], cmdParts[1:len(cmdParts)]...)
	stdout = strings.TrimSpace(stdout)
	if stdout != "" {
		fmt.Println(stdout)
	}
	if exitCode != 0 {
		if stderr != "" {
			panic(stderr)
		}
		panic(fmt.Sprintf("goreturns exited with status %v", exitCode))
	}
	fmt.Println("formatted", fpath)
}
