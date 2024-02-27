package runnables

import (
	"aggressive-pokes/internal/stats"
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

func HttpRunnable(url string, body []byte) func(reporter stats.Reporter) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxConnsPerHost:     1000,
			MaxIdleConnsPerHost: 1000,
		},
		Timeout: 5 * time.Second,
	}

	return func(reporter stats.Reporter) {
		start := time.Now()
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			reporter.ReportFailure("request_setup_error", err.Error(), time.Since(start))
			return
		}

		res, err := httpClient.Do(req)
		if err != nil {
			if res == nil {
				if errors.As(err, &http.ErrHandlerTimeout) {
					reporter.ReportFailure("http_timeout", err.Error(), time.Since(start))
				} else {
					reporter.ReportFailure("http_error", err.Error(), time.Since(start))
				}
			}
		} else {
			if res.Body != nil {
				_, _ = io.Copy(io.Discard, res.Body)
				defer res.Body.Close()
			}
			reporter.ReportSuccess(strconv.Itoa(res.StatusCode), time.Since(start))
		}

	}
}
