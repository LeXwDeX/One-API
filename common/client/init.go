package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/LeXwDeX/one-api/common/config"
	"github.com/LeXwDeX/one-api/common/logger"
)

var HTTPClient *http.Client
var ImpatientHTTPClient *http.Client
var UserContentRequestHTTPClient *http.Client

const defaultUserAgent = "Mozilla/5.0 (compatible; One-API/1.0; +https://github.com/LeXwDeX/one-api)"

type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", defaultUserAgent)
	}
	return t.base.RoundTrip(req)
}

func ensureTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &userAgentTransport{base: base}
}

func Init() {
	var userContentTransport http.RoundTripper
	if config.UserContentRequestProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as proxy to fetch user content", config.UserContentRequestProxy))
		proxyURL, err := url.Parse(config.UserContentRequestProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		userContentTransport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	userContentTransport = ensureTransport(userContentTransport)
	UserContentRequestHTTPClient = &http.Client{
		Transport: userContentTransport,
	}
	if config.UserContentRequestTimeout > 0 {
		UserContentRequestHTTPClient.Timeout = time.Second * time.Duration(config.UserContentRequestTimeout)
	}

	var relayTransport http.RoundTripper
	if config.RelayProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as api relay proxy", config.RelayProxy))
		proxyURL, err := url.Parse(config.RelayProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		relayTransport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	relayTransport = ensureTransport(relayTransport)

	if config.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: relayTransport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(config.RelayTimeout) * time.Second,
			Transport: relayTransport,
		}
	}

	ImpatientHTTPClient = &http.Client{
		Timeout:   5 * time.Second,
		Transport: relayTransport,
	}
}
