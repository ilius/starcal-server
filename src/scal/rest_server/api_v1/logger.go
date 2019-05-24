package api_v1

import (
	logging "github.com/ilius/go-logging"
)

var log = logging.GetLogger("api_v1").AddHandler(logging.NewStdoutHandler())
