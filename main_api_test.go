package main

import (
	"strings"
	"testing"
)

func TestAPIStartStopLifecycle(t *testing.T) {
	installStateCleanup(t)

	port := reserveTCPPort(t)
	uuid := "api-lifecycle"
	config := makeHTTPProxyConfig(port)

	startResp := unpackResponse(t, startString(uuid, config))
	if startResp.Code != 0 || startResp.ContentType != testContentMessage || startResp.Body != "Server started" {
		t.Fatalf("expected successful start response, got %+v", startResp)
	}
	if isStartedString(uuid) != 1 {
		t.Fatal("expected instance to be marked as started")
	}

	stopString(uuid)

	if isStartedString(uuid) != 0 {
		t.Fatal("expected instance to be stopped")
	}
}

func TestAPIStartDuplicateUUID(t *testing.T) {
	installStateCleanup(t)

	port := reserveTCPPort(t)
	uuid := "api-duplicate"
	config := makeHTTPProxyConfig(port)

	first := unpackResponse(t, startString(uuid, config))
	if first.Code != 0 {
		t.Fatalf("expected first start to succeed, got %+v", first)
	}

	second := unpackResponse(t, startString(uuid, config))
	if second.Code != 5 || second.ContentType != testContentError {
		t.Fatalf("expected duplicate start error, got %+v", second)
	}
}

func TestAPIStartInvalidJSON(t *testing.T) {
	installStateCleanup(t)

	resp := unpackResponse(t, startString("api-invalid-json", "{invalid json"))
	if resp.Code != 1 || resp.ContentType != testContentError {
		t.Fatalf("expected JSON parse error, got %+v", resp)
	}
}

func TestAPIPingConfig(t *testing.T) {
	server := newHeadServer()
	defer server.Close()

	port := reserveTCPPort(t)
	config := makeHTTPProxyConfig(port)
	resp := unpackResponse(t, pingConfigInts(config, []int{port}, server.URL))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful ping config response, got %+v", resp)
	}

	results := decodeJSONBody[[]struct {
		Port    int    `json:"port"`
		Timeout int    `json:"timeout"`
		Error   string `json:"error"`
	}](t, resp.Body)
	if len(results) != 1 {
		t.Fatalf("expected one ping result, got %+v", results)
	}
	if results[0].Port != port || results[0].Timeout < 0 || results[0].Error != "" {
		t.Fatalf("unexpected ping config result: %+v", results[0])
	}
}

func TestAPIPingErrorEmbeddedInPayload(t *testing.T) {
	resp := unpackResponse(t, pingPort(0, "http://127.0.0.1:1"))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected payload response even on ping failure, got %+v", resp)
	}

	result := decodeJSONBody[struct {
		Port    int    `json:"port"`
		Timeout int    `json:"timeout"`
		Error   string `json:"error"`
	}](t, resp.Body)
	if result.Error == "" {
		t.Fatalf("expected embedded ping error, got %+v", result)
	}
}

func TestAPIResponseBufferLayout(t *testing.T) {
	resp := unpackResponse(t, GetXrayCoreVersion())
	if resp.Code != 0 {
		t.Fatalf("expected success code, got %+v", resp)
	}
	if resp.ContentType != testContentMessage {
		t.Fatalf("expected message content type, got %+v", resp)
	}
	if strings.TrimSpace(resp.Body) == "" {
		t.Fatalf("expected non-empty response body, got %+v", resp)
	}
}
