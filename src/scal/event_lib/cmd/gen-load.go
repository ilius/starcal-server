package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
)

var useGoreturns = true

var capEventTypes = []string{
	"AllDayTask",
	"Custom",
	"DailyNote",
	"LargeScale",
	"LifeTime",
	"Monthly",
	"Task",
	"UniversityClass",
	"UniversityExam",
	"Weekly",
	"Yearly",
}

var tpl = template.Must(template.New("load").Parse(`
func Load{{.CAP_EVENT_TYPE}}EventModel(db storage.Database, sha1 string) (
	*{{.CAP_EVENT_TYPE}}EventModel,
	error,
) {
	model := {{.CAP_EVENT_TYPE}}EventModel{}
	model.Sha1 = sha1
	err := db.Get(&model)
	return &model, err
}
`))

type TemplateParams struct {
	CAP_EVENT_TYPE string
}

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

func formatFile(fpath string) {
	var cmdParts []string
	if useGoreturns {
		cmdParts = []string{"goreturns", "-w", fpath}
	} else {
		cmdParts = []string{"go", "fmt", fpath}
	}
	stdout, stderr, exitCode := RunCommand3(cmdParts[0], cmdParts[1:len(cmdParts)]...)
	fmt.Println(stdout)
	if exitCode != 0 {
		if stderr != "" {
			panic(stderr)
		}
		panic(fmt.Sprintf("goreturns exited with status %v", exitCode))
	}
}

func main() {
	goPath := os.Getenv("GOPATH")
	rootDir := goPath
	if !strings.HasSuffix(rootDir, "starcal-server") {
		rootDir = filepath.Join(rootDir, "src/github.com/ilius/starcal-server")
	}
	libDir := filepath.Join(rootDir, "src/scal/event_lib")
	// myDir, err := osext.ExecutableFolder()
	// if err != nil {
	// 	panic(err)
	// }

	src := `package event_lib
import "scal/storage"
	`
	for _, capEventType := range capEventTypes {
		var outBuff bytes.Buffer
		err := tpl.Execute(&outBuff, &TemplateParams{
			CAP_EVENT_TYPE: capEventType,
		})
		if err != nil {
			panic(err)
		}
		goBytes := outBuff.Bytes()
		src += string(goBytes)
	}
	loadGoPath := filepath.Join(libDir, "load.go")
	fmt.Println(loadGoPath)
	file, err := os.Create(loadGoPath)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(src)
	if err != nil {
		panic(err)
	}
	formatFile(loadGoPath)
}
