package runnables

import (
	"aggressive-pokes/internal/stats"
	"bytes"
	"net/http"
	"strconv"
	"time"
)

func HttpRunnable(url string, body []byte) func(reporter stats.Reporter) {
	return func(reporter stats.Reporter) {
		start := time.Now()
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			reporter.ReportFailure("request_setup_error", err.Error(), time.Since(start))
			return
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			if res == nil {
				reporter.ReportFailure("http_error", err.Error(), time.Since(start))
			} else {
				reporter.ReportFailure(strconv.Itoa(res.StatusCode), err.Error(), time.Since(start))
			}
			return
		}
		reporter.ReportSuccess(strconv.Itoa(res.StatusCode), time.Since(start))
	}
}
