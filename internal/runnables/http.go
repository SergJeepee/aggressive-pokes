package runnables

import (
	"aggressive-pokes/internal/ltlogger"
	"aggressive-pokes/internal/stats"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func HttpRunnableWithSupplier(supplier *HttpRequestSupplier) func(reporter stats.Reporter) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxConnsPerHost:     1000,
			MaxIdleConnsPerHost: 1000,
		},
		Timeout: 30 * time.Second,
	}

	supplier.logger.Info("Initialized http client", "url", supplier.url, "timeout", httpClient.Timeout)

	return func(reporter stats.Reporter) {
		start := time.Now()
		req, err := supplier.request()
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
			reporter.Report(strconv.Itoa(res.StatusCode), time.Since(start))
		}
	}
}

func HttpRunnable(logger ltlogger.Logger, url string, body []byte, headers map[string]string) func(reporter stats.Reporter) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxConnsPerHost:     1000,
			MaxIdleConnsPerHost: 1000,
		},
		Timeout: 5 * time.Second,
	}

	logger.Info("Initialized http client", "url", url, "timeout", httpClient.Timeout)

	return func(reporter stats.Reporter) {
		start := time.Now()
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if err != nil {
			fmt.Printf("%s", err.Error())
			reporter.ReportFailure("request_setup_error", err.Error(), time.Since(start))
			return
		}

		res, err := httpClient.Do(req)
		if err != nil {
			fmt.Printf("%s", err.Error())
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
			reporter.Report(strconv.Itoa(res.StatusCode), time.Since(start))
		}
	}
}
