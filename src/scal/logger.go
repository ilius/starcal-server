package scal

import (
	logging "github.com/ilius/go-logging"
)

var log = logging.GetLogger("").AddHandler(logging.NewStdoutHandler())

var Log = log
