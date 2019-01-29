package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"

	"github.com/gkkkb/piston/log"
	"github.com/gkkkb/piston/metric"
)

// MonitorHTTP is a middleware used to monitor http function
func MonitorHTTP(action string, fn func(http.ResponseWriter, *http.Request, httprouter.Params) error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		startTime := time.Now()

		// get request ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			temp, err := uuid.NewRandom()
			if err != nil {
				mapLog(reqID, []string{action, "requestUUID"}, startTime, err)
				return
			}
			reqID = temp.String()
		}
		// get actor
		actor := r.Header.Get("Authorization")

		timeout, err := strconv.Atoi(r.Header.Get("X-Deadline"))
		if err != nil {
			timeout = 3000
		} else if timeout <= 0 {
			elapsedTime := time.Since(startTime).Seconds()
			metric.TraceRequestTime(r.Method, action, "fail", elapsedTime)
			mapLog(reqID, []string{action, "getDeadline"}, startTime, fmt.Errorf("Context Timeout: %v", err))
			w.WriteHeader(504)
			return
		}

		ctx := r.Context()
		// create request ID
		ctx = context.WithValue(ctx, "X-Request-ID", reqID)
		// create retry count
		ctx = context.WithValue(ctx, "Retry", 0)
		// create actor
		ctx = context.WithValue(ctx, "Authorization", actor)
		// create deadline
		ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
		defer cancel()

		err = fn(w, r.WithContext(ctx), params)
		elapsedTime := time.Since(startTime).Seconds()
		mapLog(reqID, []string{action}, startTime, err)
		if err == nil {
			metric.TraceRequestTime(r.Method, action, "ok", elapsedTime)
		} else {
			metric.TraceRequestTime(r.Method, action, "fail", elapsedTime)
		}
	}
}

func mapLog(reqID string, tags []string, startTime time.Time, err error) {
	elaps := time.Since(startTime).Seconds()
	elapsStr := strconv.FormatFloat(elaps, 'f', -1, 64)

	m := make(map[string]interface{})
	m["request_id"] = reqID
	m["tags"] = tags
	m["duration"] = elapsStr
	if err != nil {
		m["message"] = "fail"
		log.RequestError(err.Error(), m)
		return
	}
	m["message"] = "ok"
	log.RequestInfo("", m)
}
