package storage

import (
	logging "github.com/ilius/go-logging"
)

var log = logging.GetLogger("storage").AddHandler(logging.NewStdoutHandler())
