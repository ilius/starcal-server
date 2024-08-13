package genlib

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
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

func formatGoFileBuiltin(fpath string) {
	src, err := os.ReadFile(fpath)
	if err != nil {
		panic(err)
	}
	src2, err2 := format.Source(src)
	if err2 != nil {
		panic(err2)
	}
	err3 := os.WriteFile(fpath, src2, 0o644)
	if err3 != nil {
		panic(err3)
	}
}

func FormatGoFile(fpath string, waitGroup *sync.WaitGroup, useGoreturns bool) {
	if waitGroup != nil {
		defer waitGroup.Done()
	}
	fmt.Println("formatting", fpath)
	defer fmt.Println("formatted", fpath)
	if !useGoreturns {
		formatGoFileBuiltin(fpath)
		return
	}
	cmdParts := []string{"goreturns", "-w", fpath}
	stdout, stderr, exitCode := RunCommand3(cmdParts[0], cmdParts[1:]...)
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
}
