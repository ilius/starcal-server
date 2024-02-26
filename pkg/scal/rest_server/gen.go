package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/ilius/starcal-server/pkg/scal/event_lib"

	"github.com/ilius/starcal-server/pkg/scal/genlib"

	"github.com/kardianos/osext"
)

// disable formatting file (with go fmt or goreturns)
var disableFormatGoFile = true

// no flag for it for now
var useGoreturns = true

var (
	myDir        string
	apiDir       string
	myParentDir  string
	templatesDir string
)

var activeEventModels = []any{
	event_lib.AllDayTaskEventModel{},
	event_lib.CustomEventModel{},
	event_lib.DailyNoteEventModel{},
	event_lib.LargeScaleEventModel{},
	event_lib.LifetimeEventModel{},
	event_lib.MonthlyEventModel{},
	event_lib.TaskEventModel{},
	event_lib.UniversityClassEventModel{},
	event_lib.UniversityExamEventModel{},
	event_lib.WeeklyEventModel{},
	event_lib.YearlyEventModel{},
}

var formatWaitGroup sync.WaitGroup

var eventModelByEventType = map[string]any{}

func init() {
	var err error
	myDir, err = osext.ExecutableFolder()
	if err != nil {
		panic(err)
	}
	apiDir = filepath.Join(myDir, "api_v1")
	myParentDir = filepath.Dir(myDir)
	templatesDir = filepath.Join(myParentDir, "templates")
	for _, eventModel := range activeEventModels {
		eventType := eventModel.(WithType).Type()
		eventModelByEventType[eventType] = eventModel
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

func extractModelParams(model any) []ParamRow {
	modelType := reflect.TypeOf(model)
	params := make([]ParamRow, modelType.NumField())
	for i := range modelType.NumField() {
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

func extractModelPatchParams(model any) []ParamRow {
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

func genEventTypeHandlers(eventType string) {
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

	eventModels := activeEventModels
	if eventType != "" {
		eventModels = []any{
			eventModelByEventType[eventType],
		}
	}

	basePatchParams := extractModelPatchParams(event_lib.BaseEventModel{})
	for _, eventModel := range eventModels {
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
		err = ioutil.WriteFile(goPath, []byte(goText), 0o644)
		if err != nil {
			panic(err)
		}
		if !disableFormatGoFile {
			formatWaitGroup.Add(1)
			go genlib.FormatGoFile(goPath, &formatWaitGroup, useGoreturns)
		}
	}
}

func main() {
	eventType := ""
	flag.StringVar(&eventType, "event-type", "", "Event Type")

	flag.BoolVar(&disableFormatGoFile, "no-fmt", false, "Do not run format Go files")

	flag.Parse()

	genEventTypeHandlers(eventType)
	formatWaitGroup.Wait()
}
