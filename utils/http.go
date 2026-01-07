package utils

import (
	"net/http"
	"time"
)

var GlobalHTTPClient = defaultHTTPClient()

type Option func(*http.Client)

func NewHTTPClient(opts ...Option) *http.Client {
	client := defaultHTTPClient()
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *http.Client) {
		c.Timeout = timeout
	}
}

func WithMaxIdleConns(max int) Option {
	return func(c *http.Client) {
		if t, ok := c.Transport.(*http.Transport); ok && t != nil {
			t.MaxIdleConns = max
		}
	}
}

func WithMaxIdleConnsPerHost(max int) Option {
	return func(c *http.Client) {
		if t, ok := c.Transport.(*http.Transport); ok && t != nil {
			t.MaxIdleConnsPerHost = max
		}
	}
}

func WithIdleConnTimeout(timeout time.Duration) Option {
	return func(c *http.Client) {
		if t, ok := c.Transport.(*http.Transport); ok && t != nil {
			t.IdleConnTimeout = timeout
		}
	}
}

func WithTLSHandshakeTimeout(timeout time.Duration) Option {
	return func(c *http.Client) {
		if t, ok := c.Transport.(*http.Transport); ok && t != nil {
			t.TLSHandshakeTimeout = timeout
		}
	}
}
