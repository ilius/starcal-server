package api_v1

import (
	"fmt"
	"log"
	"scal/storage"

	"github.com/ilius/restpc"
)

type ErrorModel struct {
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
	return "errors_" + m.Code
}

func SetMongoErrorDispatcher() {
	db, err := storage.GetDB()
	if err != nil {
		panic(err)
	}
	restpc.SetErrorDispatcher(func(request restpc.Request, rpcErr restpc.RPCError) {
		errorModel := &ErrorModel{
			URL:       request.URL().String(),
			Code:      rpcErr.Code().String(),
			Message:   rpcErr.Message(),
			Details:   rpcErr.Details(),
			Request:   request.FullMap(),
			Traceback: rpcErr.Traceback().MapRecords(),
		}
		privateErr := rpcErr.Private()
		if privateErr != nil {
			errorModel.PrivateMessage = privateErr.Error()
			errorModel.PrivateType = fmt.Sprintf("%T", privateErr)
		}
		err := db.Insert(errorModel)
		if err != nil {
			log.Println(err)
		}
		log.Println(rpcErr.Code(), rpcErr.Error(), rpcErr.Details())
	})
}
