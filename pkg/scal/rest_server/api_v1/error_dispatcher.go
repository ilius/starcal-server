package api_v1

import (
	"fmt"
	"github.com/ilius/starcal-server/pkg/scal/settings"
	"github.com/ilius/starcal-server/pkg/scal/storage"
	"time"

	"github.com/ilius/ripo"
)

var errorChan = make(chan *errorChanItem, settings.ERRORS_CHANNEL_SIZE)

type errorChanItem struct {
	Time    time.Time
	Request ripo.ExtendedRequest
	Error   ripo.RPCError
}

func errorDispatcher(request ripo.ExtendedRequest, rpcErr ripo.RPCError) {
	errorChan <- &errorChanItem{
		Time:    time.Now().UTC(),
		Request: request,
		Error:   rpcErr,
	}
	log.Error(fmt.Sprintf(
		"%v %v %v %v",
		rpcErr.Code(),
		rpcErr.Error(),
		rpcErr.Details(),
		rpcErr.Cause(),
	))
}

type ErrorModel struct {
	Time         time.Time                `bson:"time"`
	HandlerName  string                   `bson:"handlerName"`
	URL          string                   `bson:"url"`
	Code         string                   `bson:"code"`
	Message      string                   `bson:"message"`
	CauseMessage string                   `bson:"causeMessage"`
	CauseType    string                   `bson:"causeType"`
	Details      map[string]interface{}   `bson:"details"`
	Request      map[string]interface{}   `bson:"request"`
	Traceback    []map[string]interface{} `bson:"traceback"`
}

func errorCollection(code ripo.Code) string {
	return storage.C_errorsPrefix + code.String()
}

func DispatchError(request ripo.Request, rpcErr ripo.RPCError) {
	requestExt, ok := request.(ripo.ExtendedRequest)
	if !ok {
		log.Error(rpcErr)
		log.Error("CRITICAL: DispatchError: request is not ExtendedRequest")
		return
	}
	errorDispatcher(requestExt, rpcErr)
}

func SetMongoErrorDispatcher() {
	ripo.SetErrorDispatcher(errorDispatcher)
}

func saveErrors(byCode map[ripo.Code][]*errorChanItem) {
	defer func() {
		r := recover()
		if r != nil {
			log.Error("panic in saveErrors: ", r)
		}
	}()
	db, err := storage.GetDB()
	if err != nil {
		log.Error("error in saveErrors: storage.GetDB: ", err)
		return
	}
	for code, items := range byCode {
		if len(items) == 0 {
			continue
		}
		errorModels := make([]interface{}, len(items))
		for index, item := range items {
			rpcErr := item.Error
			request := item.Request
			handlerName := request.HandlerName()
			traceback := rpcErr.Traceback(handlerName)
			errorModel := &ErrorModel{
				Time:        item.Time,
				HandlerName: handlerName,
				URL:         request.URL().String(),
				Code:        rpcErr.Code().String(),
				Message:     rpcErr.Message(),
				Details:     rpcErr.Details(),
				Request:     request.FullMap(),
				Traceback:   traceback.MapRecords(),
			}
			causeErr := rpcErr.Cause()
			if causeErr != nil {
				errorModel.CauseMessage = causeErr.Error()
				errorModel.CauseType = fmt.Sprintf("%T", causeErr)
			}
			errorModels[index] = errorModel
		}
		err := db.InsertMany(errorCollection(code), errorModels)
		if err != nil {
			log.Error("error in saveErrors: db.InsertMany: ", err)
			continue
		}
	}
}

func ErrorSaverLoop() {
	byCode := map[ripo.Code][]*errorChanItem{}
	ticker := time.NewTicker(settings.ERRORS_LOOP_SLEEP_DURATION_SECONDS * time.Second)
	for {
		select {
		case item := <-errorChan:
			code := item.Error.Code()
			byCode[code] = append(byCode[code], item)
			// log.Debug("-- added to map: Error=", item.Error)
		case <-ticker.C:
			if len(byCode) > 0 {
				// log.Debug("---- saveErrors starting, len(byCode) = ", len(byCode))
				saveErrors(byCode)
				byCode = map[ripo.Code][]*errorChanItem{}
				// log.Debug("---- saveErrors finished")
			}
		}
	}
}
