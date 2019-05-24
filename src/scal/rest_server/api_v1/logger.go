package api_v1

import (
	logging "github.com/hhkbp2/go-logging"
)

var log logging.Logger

func init() {
	if log == nil {
		log = logging.GetLogger("api_v1")
		log.AddHandler(logging.NewStdoutHandler())
	}
}
