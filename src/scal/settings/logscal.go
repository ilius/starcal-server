package settings

import (
	"fmt"

	logging "github.com/ilius/go-logging"
)

func NewLogger(name string) logging.Logger {
	logger := logging.GetLogger(name)
	handler := logging.NewStdoutHandler()

	format := "%(asctime)s %(levelname)s: %(name)s: %(filename)s:%(lineno)d: %(message)s"
	dateFormat := "%Y-%m-%d %H:%M:%S.%3n"
	formatter := logging.NewStandardFormatter(format, dateFormat)
	handler.SetFormatter(formatter)

	logger.AddHandler(handler)

	levelName := LOG_LEVEL
	level, ok := logging.GetLevelByName(levelName)
	if !ok {
		panic(fmt.Errorf("invalid settings.LOG_LEVEL=%#v", levelName))
	}
	logger.SetLevel(level)

	return logger
}
