package api_v1

import (
	"fmt"
	"log"
	"scal/settings"
	"scal/storage"
	"time"

	"github.com/ilius/ripo"
)

var errorChan = make(chan *errorChanItem, settings.ERRORS_CHANNEL_SIZE)

var errorLoopSleepDuration time.Duration = 10 * time.Second

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
	log.Println(rpcErr.Code(), rpcErr.Error(), rpcErr.Details(), rpcErr.Cause())
}

type ErrorModel struct {
	Time           time.Time                `bson:"time"`
	HandlerName    string                   `bson:"handlerName"`
	URL            string                   `bson:"url"`
	Code           string                   `bson:"code"`
	Message        string                   `bson:"message"`
	PrivateMessage string                   `bson:"privateMessage"`
	PrivateType    string                   `bson:"privateType"`
	Details        map[string]interface{}   `bson:"details"`
	Request        map[string]interface{}   `bson:"request"`
	Traceback      []map[string]interface{} `bson:"traceback"`
}

func errorCollection(code ripo.Code) string {
	return storage.C_errorsPrefix + code.String()
}

func DispatchError(request ripo.Request, rpcErr ripo.RPCError) {
	requestExt, ok := request.(ripo.ExtendedRequest)
	if !ok {
		log.Println(rpcErr)
		log.Println("CRITICAL: DispatchError: request is not ExtendedRequest")
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
			log.Println("panic in saveErrors:", r)
		}
	}()
	db, err := storage.GetDB()
	if err != nil {
		log.Println("error in saveErrors: storage.GetDB:", err)
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
				errorModel.PrivateMessage = causeErr.Error()
				errorModel.PrivateType = fmt.Sprintf("%T", causeErr)
			}
			errorModels[index] = errorModel
		}
		err := db.InsertMany(errorCollection(code), errorModels)
		if err != nil {
			log.Println("error in saveErrors: db.InsertMany:", err)
			continue
		}
	}
}

func ErrorSaverLoop() {
	byCode := map[ripo.Code][]*errorChanItem{}
	ticker := time.NewTicker(errorLoopSleepDuration)
	for {
		select {
		case item := <-errorChan:
			code := item.Error.Code()
			byCode[code] = append(byCode[code], item)
			// log.Println("-- added to map:", item.Error)
		case <-ticker.C:
			if len(byCode) > 0 {
				// log.Println("---- saveErrors starting, len(byCode) = ", len(byCode))
				saveErrors(byCode)
				byCode = map[ripo.Code][]*errorChanItem{}
				// log.Println("---- saveErrors finished")
			}
		}
	}
}
