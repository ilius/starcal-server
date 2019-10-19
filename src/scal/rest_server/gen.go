package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"scal/event_lib"
	"strings"
	"sync"
	"text/template"

	"github.com/kardianos/osext"
)

// format file (with go fmt or goreturns)
var enableFormatFile = true

// no flag for it for now
var useGoreturns = true

var myDir string
var apiDir string
var myParentDir string
var templatesDir string

var activeEventModels = []interface{}{
	event_lib.AllDayTaskEventModel{},
	event_lib.DailyNoteEventModel{},
	event_lib.LargeScaleEventModel{},
	event_lib.LifeTimeEventModel{},
	event_lib.MonthlyEventModel{},
	event_lib.TaskEventModel{},
	event_lib.UniversityClassEventModel{},
	event_lib.UniversityExamEventModel{},
	event_lib.WeeklyEventModel{},
	event_lib.YearlyEventModel{},
	event_lib.CustomEventModel{},
}

var fmtWG sync.WaitGroup

func formatFile(fpath string) {
	var cmd *exec.Cmd
	if useGoreturns {
		cmd = exec.Command("goreturns", "-w", fpath)
	} else {
		cmd = exec.Command("go", "fmt", fpath)
	}
	err := cmd.Run()
	fmtWG.Done()
	if err != nil {
		panic(err)
	}
}

func init() {
	var err error
	myDir, err = osext.ExecutableFolder()
	if err != nil {
		panic(err)
	}
	apiDir = filepath.Join(myDir, "api_v1")
	myParentDir = filepath.Dir(myDir)
	templatesDir = filepath.Join(myParentDir, "templates")
	if len(os.Args) > 1 && os.Args[1] == "--no-fmt" {
		enableFormatFile = false
	}
}

func lowerFirstLetter(paramCap string) string {
	return strings.ToLower(string(paramCap[0])) + paramCap[1:]
}

func upperFirstLetter(param string) string {
	return strings.ToUpper(string(param[0])) + param[1:]
}

type ParamRow struct {
	PARAM      string
	CAP_PARAM  string
	PARAM_TYPE string
	PARAM_KIND string
	PARAM_INT  bool
}

var intKinds = map[reflect.Kind]bool{
	reflect.Int:    true,
	reflect.Int8:   true,
	reflect.Int16:  true,
	reflect.Int32:  true,
	reflect.Int64:  true,
	reflect.Uint:   true,
	reflect.Uint8:  true,
	reflect.Uint16: true,
	reflect.Uint32: true,
	reflect.Uint64: true,
}

func extractModelParams(model interface{}) []ParamRow {
	modelType := reflect.TypeOf(model)
	params := make([]ParamRow, modelType.NumField())
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		paramType := field.Type.String()
		paramKind := field.Type.Kind()
		paramInt := intKinds[paramKind]

		paramCap := field.Name

		if paramType == "event_lib.BaseEventModel" {
		}

		param := lowerFirstLetter(paramCap)
		params[i] = ParamRow{
			PARAM:      param,
			CAP_PARAM:  upperFirstLetter(param),
			PARAM_TYPE: paramType,
			PARAM_KIND: paramKind.String(),
			PARAM_INT:  paramInt,
		}
	}
	return params
}

func extractModelPatchParams(model interface{}) []ParamRow {
	params := extractModelParams(model)
	patchParams := make([]ParamRow, 0, len(params))
	for _, row := range params {
		param := row.PARAM
		switch param {
		case "id", "sha1", "dummyType":
			continue
		case "groupId", "meta":
			continue
		case "baseEventModel":
			continue
		}
		patchParams = append(patchParams, row)
	}
	return patchParams
}

type WithType interface {
	Type() string
}

type EventHandlersTemplateParams struct {
	EVENT_TYPE         string
	CAP_EVENT_TYPE     string
	EVENT_PATCH_PARAMS []ParamRow
}

var emptyLineRE = regexp.MustCompile(`(?m)^\s+\n`)

func genEventTypeHandlers() {
	tpl_path := path.Join(templatesDir, "event_handlers.go.tpl")
	tpl_bytes, err := ioutil.ReadFile(tpl_path)
	if err != nil {
		panic(err)
	}
	tpl_str := string(tpl_bytes)
	tpl, err := template.New("event_handlers").Parse(tpl_str)
	if err != nil {
		panic(err)
	}

	basePatchParams := extractModelPatchParams(event_lib.BaseEventModel{})
	for _, eventModel := range activeEventModels {
		eventType := eventModel.(WithType).Type()
		eventTypeCap := upperFirstLetter(eventType)
		typePatchParams := append(basePatchParams, extractModelPatchParams(eventModel)...)

		var outBuff bytes.Buffer
		err := tpl.Execute(&outBuff, &EventHandlersTemplateParams{
			EVENT_TYPE:         eventType,
			CAP_EVENT_TYPE:     eventTypeCap,
			EVENT_PATCH_PARAMS: typePatchParams,
		})
		if err != nil {
			panic(err)
		}
		goBytes := outBuff.Bytes()
		goPath := path.Join(apiDir, fmt.Sprintf("event_handlers_%v.go", eventType))
		goText := string(goBytes)
		goText = emptyLineRE.ReplaceAllString(goText, "")
		err = ioutil.WriteFile(goPath, []byte(goText), 0644)
		if err != nil {
			panic(err)
		}
		if enableFormatFile {
			fmtWG.Add(1)
			go formatFile(goPath)
		}
	}
}

func main() {
	genEventTypeHandlers()
	fmtWG.Wait()
}
