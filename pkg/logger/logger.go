package logger

import (
	"log"
	"os"
)

func init() {
	logfile, err := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err == nil {
		log.SetOutput(logfile)
	}
}
