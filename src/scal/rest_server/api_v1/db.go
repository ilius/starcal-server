package api_v1

import (
	"scal/storage"
)

var globalDb, globalDbErr = storage.GetDB()

func init() {
	if globalDbErr != nil {
		panic(globalDbErr)
	}
}
