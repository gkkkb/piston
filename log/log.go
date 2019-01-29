// Package log is used to write log to stdout
package log

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger          *zap.Logger
	mandatoryFields []string
	errNotValid     error
)

func init() {
	logger, _ = zap.NewProduction()
	mandatoryFields = []string{"request_id", "tags", "message", "duration", "severity"}
	errNotValid = errors.New("Missing required field(s)")
}

// JobLog is function that returns error
type JobLog func() error

// RequestInfo writes log with severity = info.
// It is intended to be used to log request.
func RequestInfo(message string, args map[string]interface{}) error {
	args["severity"] = "INFO"
	if fields, valid := extractArgs(args); valid {
		logger.Info(message, fields...)
		return nil
	}
	return errNotValid
}

// RequestError writes log with severity = error.
// It is intended to be used to log request.
func RequestError(message string, args map[string]interface{}) error {
	args["severity"] = "ERROR"
	if fields, valid := extractArgs(args); valid {
		logger.Error(message, fields...)
		return nil
	}
	return errNotValid
}

// Job is a function that wraps a JobLog and write its error
func Job(ctx context.Context, name string, job JobLog) error {
	reqID, _ := ctx.Value("X-Request-ID").(string)
	tags := []string{name}

	args := make(map[string]interface{})
	args["request_id"] = reqID
	args["tags"] = tags

	start := time.Now()
	err := job()
	elapsed := time.Since(start).Seconds()

	elapStr := strconv.FormatFloat(elapsed, 'f', -1, 64)
	args["duration"] = elapStr

	if err == nil {
		args["severity"] = "INFO"
		args["message"] = "job was run successfully!"
		return RequestInfo(name, args)
	}

	args["severity"] = "ERROR"
	args["message"] = err.Error()
	return RequestError(name, args)
}

func extractArgs(args map[string]interface{}) ([]zapcore.Field, bool) {
	var fields []zapcore.Field
	keys := make(map[string]bool)

	for k, v := range args {
		field := zap.Any(k, v)
		fields = append(fields, field)

		keys[k] = true
	}

	for _, field := range mandatoryFields {
		if _, found := keys[field]; !found {
			return fields, false
		}
	}

	return fields, true
}
