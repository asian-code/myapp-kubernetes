package logger

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

func Init(serviceName string) *log.Entry {
	l := log.New()

	// JSON format for structured logging
	l.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	l.SetOutput(os.Stdout)
	l.SetLevel(log.InfoLevel)

	if os.Getenv("LOG_LEVEL") == "debug" {
		l.SetLevel(log.DebugLevel)
	}

	return l.WithField("service", serviceName)
}

func WithContext(ctx context.Context, logger *log.Entry) *log.Entry {
	if requestID := ctx.Value("request-id"); requestID != nil {
		logger = logger.WithField("request_id", requestID)
	}
	return logger
}
