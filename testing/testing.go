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
	PingTimeoutError uint = 6
	PingError        uint = 7

	_localhost string = "http://127.0.0.1:"
)

type PingResult struct {
	Port    int    `json:"port"`
	Timeout int    `json:"timeout"`
	Error   string `json:"error"`
}

func PingConfig(jsonConfig string, ports []int, testingURL string) ([]PingResult, error) {
	instance, err := xray.Start(jsonConfig)
	if err != nil {
		return []PingResult{}, errors.New(err.Message)
	}

	results := make([]PingResult, 0, len(ports))
	for _, port := range ports {
		result := PingResult{Port: port}
		proxyUrl, err1 := url.Parse(_localhost + strconv.Itoa(port))
		if err1 != nil {
			result.Error = err1.Error()
		} else {
			timeout, err2 := PingProxy(testingURL, proxyUrl)

			if err2 != nil {
				result.Error = err2.Error()
			}

			result.Timeout = timeout
		}

		results = append(results, result)
	}

	instance.Close()

	return results, nil
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
