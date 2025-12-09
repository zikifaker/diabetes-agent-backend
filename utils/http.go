package utils

import (
	"net/http"
	"time"
)

type Option func(*http.Client)

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

func NewHTTPClient(opts ...Option) *http.Client {
	client := DefaultHTTPClient()
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}
