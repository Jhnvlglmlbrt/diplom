package logger

import (
	"os"

	kitlog "github.com/go-kit/log"
)

var logger kitlog.Logger

func Init() {
	logout := os.Stderr
	if logpath := os.Getenv("SENDIT_LOG"); logpath != "" {
		if file, err := os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logout = file
		} else {
			panic(err)
		}
	}
	logger = kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(logout))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.Caller(4))
}

func Log(args ...interface{}) error {
	return logger.Log(args...)
}
