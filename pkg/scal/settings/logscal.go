package settings

import (
	"fmt"

	logging "github.com/hhkbp2/go-logging"
)

func NewLogger(name string) logging.Logger {
	logger := logging.GetLogger(name)
	handler := logging.NewStdoutHandler()

	formatter := logging.NewStandardFormatter(LOG_FORMAT, LOG_DATE_FORMAT)
	handler.SetFormatter(formatter)

	logger.AddHandler(handler)

	levelName := LOG_LEVEL
	level, ok := logging.GetNameLevel(levelName)
	if !ok {
		panic(fmt.Errorf("invalid settings.LOG_LEVEL=%#v", levelName))
	}
	err := logger.SetLevel(level)
	if err != nil {
		panic(err)
	}

	return logger
}
