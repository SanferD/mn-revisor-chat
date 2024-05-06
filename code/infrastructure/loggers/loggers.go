package loggers

import (
	"fmt"
	"log"
	"os"
)

type MultiLogger struct {
	loggers []*log.Logger
}

func InitializeMultiLogger(logToStdout bool) (*MultiLogger, error) {
	loggers := make([]*log.Logger, 0)
	if logToStdout {
		stdoutLogger := log.New(os.Stdout, "", log.LstdFlags)
		loggers = append(loggers, stdoutLogger)
	}
	return &MultiLogger{loggers: loggers}, nil
}

func (multiLogger *MultiLogger) Info(msg string, args ...any) {
	multiLogger.doLog("info", msg, args...)
}

func (multiLogger *MultiLogger) Warn(msg string, args ...any) {
	multiLogger.doLog("warn", msg, args...)
}

func (multiLogger *MultiLogger) Debug(msg string, args ...any) {
	multiLogger.doLog("debug", msg, args...)
}

func (multiLogger *MultiLogger) Error(msg string, args ...any) {
	multiLogger.doLog("error", msg, args...)
}

func (multiLogger *MultiLogger) Fatal(msg string, args ...any) {
	multiLogger.doLog("fatal", msg, args...)
}

func (multiLogger *MultiLogger) doLog(prefix, msg string, args ...any) {
	for _, logger := range multiLogger.loggers {
		logger.SetPrefix(prefix + ":")
		result := fmt.Sprintf(msg, args...)
		logger.Println(result)
	}
	if prefix == "fatal" {
		os.Exit(1)
	}
}
