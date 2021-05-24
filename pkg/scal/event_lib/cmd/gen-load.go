package main

import (
	"bytes"
	"fmt"
	"github.com/ilius/starcal-server/pkg/scal/genlib"
	"os"
	"path/filepath"
	"text/template"
)

var useGoreturns = true

var capEventTypes = []string{
	"AllDayTask",
	"Custom",
	"DailyNote",
	"LargeScale",
	"Lifetime",
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

func main() {
	goPath := os.Getenv("GOPATH")
	rootDir := filepath.Join(goPath, "src/github.com/ilius/starcal-server")
	libDir := filepath.Join(rootDir, "pkg/scal/event_lib")
	// myDir, err := osext.ExecutableFolder()
	// if err != nil {
	// 	panic(err)
	// }

	src := `// Do not modify this file, it's auto-generated
package event_lib
import "github.com/ilius/starcal-server/pkg/scal/storage"
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
	genlib.FormatGoFile(loadGoPath, nil, useGoreturns)
}
