package runnables

import (
	"aggressive-pokes/internal/ltlogger"
	"bytes"
	"net/http"
)

type HttpRequestSupplier struct {
	logger         ltlogger.Logger
	method         string
	url            string
	body           []byte
	headers        map[string]string
	payloadMutator func([]byte) []byte
}

func NewHttpRequestSupplier(logger ltlogger.Logger, method, url string, body []byte, headers map[string]string, payloadMutator func([]byte) []byte) *HttpRequestSupplier {
	if payloadMutator == nil {
		payloadMutator = func(b []byte) []byte {
			return b
		}
	}

	return &HttpRequestSupplier{
		logger:         logger,
		method:         method,
		url:            url,
		body:           body,
		headers:        headers,
		payloadMutator: payloadMutator,
	}
}

func (s *HttpRequestSupplier) request() (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(s.payloadMutator(s.body)))
	if s.headers != nil {
		for k, v := range s.headers {
			req.Header.Set(k, v)
		}
	}
	if err != nil {
		return nil, err
	}
	return req, nil
}
