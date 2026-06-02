package clients

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

type SharedHTTPClient struct {
	client *http.Client
}

func NewDefaultSharedHTTPClient() *SharedHTTPClient {
	return NewSharedHTTPClientWithConfig(defaultHTTPClientConfig)
}

func NewSharedHTTPClientWithConfig(config *HTTPClientConfig) *SharedHTTPClient {
	if config == nil {
		config = defaultHTTPClientConfig
	}

	transport := &http.Transport{
		Proxy:                 config.Proxy,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSClientConfig:       config.TLSConfig,
		DisableKeepAlives:     config.DisableKeepAlives,
		DisableCompression:    config.DisableCompression,
		ForceAttemptHTTP2:     config.ForceAttemptHTTP2,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport:     transport,
		CheckRedirect: config.CheckRedirect,
	}

	utils.DevLogInfo(utils.PrettyObjStr(config))

	return &SharedHTTPClient{
		client: client,
	}
}

func (c *SharedHTTPClient) Request(req *http.Request) (data []byte, headers http.Header, httpCode int, err error) {
	defer func() {
		if r := recover(); r != nil {
			headers = nil
			data = nil
			httpCode = http.StatusInternalServerError
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if c == nil {
		headers = nil
		data = nil
		httpCode = http.StatusInternalServerError
		err = fmt.Errorf("shared http client instance is nil")
		return
	}

	if c.client == nil {
		headers = nil
		data = nil
		httpCode = http.StatusInternalServerError
		err = fmt.Errorf("shared http client is not initialized yet")
		return
	}

	if req == nil {
		headers = nil
		data = nil
		httpCode = http.StatusInternalServerError
		err = fmt.Errorf("req is nil")
		return
	}

	resp, err := c.client.Do(req)
	if err != nil {
		headers = nil
		data = nil
		httpCode = 0
		err = fmt.Errorf("failed to send request: %w", err)
		return
	}

	if resp == nil {
		headers = nil
		data = nil
		httpCode = 0
		err = fmt.Errorf("response is nil")
		return
	}

	defer resp.Body.Close()

	httpCode = resp.StatusCode
	headers = resp.Header

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		headers = nil
		data = nil
		err = fmt.Errorf("failed to read response body: %w", err)
		return
	}

	data = body
	err = nil
	return
}
