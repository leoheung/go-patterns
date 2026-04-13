package clients

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

var (
	sharedHTTPClient *http.Client
	once             sync.Once
)

// HTTPClientConfig 定义共享 HTTP Client 的配置
// 注意：不包含 Timeout，因为应该在 request level 通过 context 控制
type HTTPClientConfig struct {
	// 连接池配置
	MaxIdleConns        int           // 最大空闲连接数
	MaxIdleConnsPerHost int           // 每个主机的最大空闲连接数
	MaxConnsPerHost     int           // 每个主机的最大连接数
	IdleConnTimeout     time.Duration // 空闲连接超时时间

	// TLS 配置
	TLSConfig *tls.Config

	// 重定向策略
	CheckRedirect func(req *http.Request, via []*http.Request) error

	// 代理配置
	Proxy func(*http.Request) (*url.URL, error)

	// 其他传输层配置
	DisableKeepAlives  bool
	DisableCompression bool
	ForceAttemptHTTP2  bool
}

// defaultHTTPClientConfig 返回默认的 HTTP Client 配置
var defaultHTTPClientConfig = &HTTPClientConfig{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	MaxConnsPerHost:     100,
	IdleConnTimeout:     90 * time.Second,
	ForceAttemptHTTP2:   true,
}

// InitSharedHTTPClient 使用默认配置初始化共享 HTTP Client
func InitDefaultSharedHTTPClient() {
	InitSharedHTTPClientWithConfig(defaultHTTPClientConfig)
}

// InitSharedHTTPClientWithConfig 使用自定义配置初始化共享 HTTP Client
func InitSharedHTTPClientWithConfig(config *HTTPClientConfig) {
	once.Do(func() {
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

		sharedHTTPClient = &http.Client{
			Transport:     transport,
			CheckRedirect: config.CheckRedirect,
			// 注意：不设置 Timeout，由 request level 的 context 控制
		}
		utils.DevLogInfo(utils.PrettyObjStr(config))
	})
}

func Request(req *http.Request) (data []byte, httpCode int, err error) {
	defer func() {
		if r := recover(); r != nil {
			data = nil
			httpCode = http.StatusInternalServerError
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if sharedHTTPClient == nil {
		data = nil
		httpCode = http.StatusInternalServerError
		err = fmt.Errorf("shared http client is not initialized yet")
		return
	}

	if req == nil {
		data = nil
		httpCode = http.StatusInternalServerError
		err = fmt.Errorf("req is nil")
		return
	}

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		data = nil
		httpCode = 0
		err = fmt.Errorf("failed to send request: %w", err)
		return
	}

	if resp == nil {
		data = nil
		httpCode = 0
		err = fmt.Errorf("response is nil")
		return
	}

	defer resp.Body.Close()

	httpCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		data = nil
		err = fmt.Errorf("failed to read response body: %w", err)
		return
	}

	data = body
	err = nil
	return
}

func ParseResponse[T any](data []byte, dest *T) error {
	return json.Unmarshal(data, dest)
}
