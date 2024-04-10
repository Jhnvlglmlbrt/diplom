package logger

import (
	"io"
	"log"
	"os"

	kitlog "github.com/go-kit/log"
)

var logger kitlog.Logger

func Init() {
	var logout io.Writer
	logpath := os.Getenv("SENDIT_LOG")
	if logpath == "" {
		log.Fatal("SENDIT_LOG environment variable is not set")
	}
	if _, err := os.Stat(logpath); err != nil {
		_, err := os.Create(logpath)
		if err != nil {
			log.Fatal(err)
		}
	}
	logfile, err := os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logout = io.MultiWriter(logfile, os.Stderr)

	logger = kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(logout))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.Caller(4))
}

func Log(args ...any) error {
	return logger.Log(args...)
}
