package testing

import "C"
import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/VanyaKrotov/xray_cshare/xray"
)

const (
	PingTimeoutError int = 6
	PingError        int = 7

	_localhost string = "http://127.0.0.1:"
)

func PingConfig(jsonConfig string, port int, testingURL string) (int, error) {
	instance, err := xray.Start(jsonConfig)
	if err != nil {
		return -1, errors.New(err.Message)
	}

	proxyUrl, err1 := url.Parse(_localhost + strconv.Itoa(port))
	if err1 != nil {
		instance.Close()

		return -1, err1
	}

	timeout, err2 := PingProxy(testingURL, proxyUrl)

	instance.Close()

	return timeout, err2
}

func PingProxy(testUrl string, proxyUrl *url.URL) (int, error) {
	defaultTransport := &http.Transport{
		Proxy:               http.ProxyURL(proxyUrl),
		TLSHandshakeTimeout: time.Second * 5,
		DisableKeepAlives:   true,
	}

	return pingWithTransport(testUrl, defaultTransport)
}

func pingWithTransport(testUrl string, transport *http.Transport) (int, error) {
	start := time.Now()
	http.DefaultTransport = transport
	response, err := http.Head(testUrl)
	if err != nil {
		return -1, err
	}

	if response.StatusCode == 204 || response.StatusCode == 200 {
		return int(time.Since(start).Milliseconds()), nil
	}

	return -1, errors.New("Invalid ")
}

func Ping(port int, testUrl string) (int, error) {
	if port == 0 {
		defaultTransport := &http.Transport{
			TLSHandshakeTimeout: time.Second * 5,
			DisableKeepAlives:   true,
		}

		return pingWithTransport(testUrl, defaultTransport)
	}

	proxyUrl, err := url.Parse(_localhost + strconv.Itoa(port))
	if err != nil {
		return -1, err
	}

	return PingProxy(testUrl, proxyUrl)
}
