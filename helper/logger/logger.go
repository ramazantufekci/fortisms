package logger
import (
	"log"
	"os"
)

var Log *log.Logger

func Init(logPath string) error {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	Log = log.New(logFile,"",log.LstdFlags)
	return nil
}
