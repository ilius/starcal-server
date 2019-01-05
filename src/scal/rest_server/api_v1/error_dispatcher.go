package api_v1

import (
	"fmt"
	"log"
	"scal/storage"
	"time"

	"github.com/ilius/ripo"
)

var errorDispatcher func(request ripo.ExtendedRequest, rpcErr ripo.RPCError)

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

func (m *ErrorModel) Collection() string {
	return storage.C_errorsPrefix + m.Code
}

func newErrorDispatcher(db storage.Database) func(request ripo.ExtendedRequest, rpcErr ripo.RPCError) {
	return func(request ripo.ExtendedRequest, rpcErr ripo.RPCError) {
		handlerName := request.HandlerName()
		traceback := rpcErr.Traceback(handlerName)
		errorModel := &ErrorModel{
			Time:        time.Now().UTC(),
			HandlerName: handlerName,
			URL:         request.URL().String(),
			Code:        rpcErr.Code().String(),
			Message:     rpcErr.Message(),
			Details:     rpcErr.Details(),
			Request:     request.FullMap(),
			Traceback:   traceback.MapRecords(),
		}
		privateErr := rpcErr.Cause()
		if privateErr != nil {
			errorModel.PrivateMessage = privateErr.Error()
			errorModel.PrivateType = fmt.Sprintf("%T", privateErr)
		}
		err := db.Insert(errorModel)
		if err != nil {
			log.Println(err)
		}
		log.Println(rpcErr.Code(), rpcErr.Error(), rpcErr.Details(), privateErr)
	}
}

func DispatchError(request ripo.Request, rpcErr ripo.RPCError) {
	requestExt, ok := request.(ripo.ExtendedRequest)
	if !ok {
		log.Println(rpcErr)
		log.Println("CRTITICAL: DispatchError: request is not ExtendedRequest")
		return
	}
	errorDispatcher(requestExt, rpcErr)
}

func SetMongoErrorDispatcher() {
	if errorDispatcher == nil {
		db, err := storage.GetDB()
		if err != nil {
			panic(err)
		}
		errorDispatcher = newErrorDispatcher(db)
	}
	ripo.SetErrorDispatcher(errorDispatcher)
}
