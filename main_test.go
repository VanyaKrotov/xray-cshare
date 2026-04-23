package main

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"unsafe"
)

const (
	testContentError   uint16 = 1
	testContentPayload uint16 = 2
	testContentMessage uint16 = 3
)

type packedResponse struct {
	Code        uint32
	ContentType uint16
	Body        string
}

func unpackResponse(t *testing.T, ptr unsafe.Pointer) packedResponse {
	t.Helper()

	if ptr == nil {
		t.Fatal("got nil response pointer")
	}

	header := unsafe.Slice((*byte)(ptr), 8)
	resp := packedResponse{
		Code:        binary.LittleEndian.Uint32(header[0:4]),
		ContentType: binary.LittleEndian.Uint16(header[4:6]),
	}
	bodyPtr := (*byte)(unsafe.Add(ptr, 8))
	bodyBytes := make([]byte, 0, 64)
	for i := 0; ; i++ {
		b := *(*byte)(unsafe.Add(unsafe.Pointer(bodyPtr), uintptr(i)))
		if b == 0 {
			break
		}
		bodyBytes = append(bodyBytes, b)
	}
	resp.Body = string(bodyBytes)

	FreePointer(ptr)

	return resp
}

func decodeJSONBody[T any](t *testing.T, body string) T {
	t.Helper()

	var value T
	if err := json.Unmarshal([]byte(body), &value); err != nil {
		t.Fatalf("failed to decode JSON body %q: %v", body, err)
	}

	return value
}

func reserveTCPPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve TCP port: %v", err)
	}
	defer ln.Close()

	return ln.Addr().(*net.TCPAddr).Port
}

func resetState(t *testing.T) {
	t.Helper()

	mu.Lock()
	ids := make([]string, 0, len(instances))
	for id := range instances {
		ids = append(ids, id)
	}
	starting = make(map[string]struct{})
	mu.Unlock()

	for _, id := range ids {
		stopString(id)
	}
}

func installStateCleanup(t *testing.T) {
	t.Helper()

	resetState(t)
	t.Cleanup(func() {
		resetState(t)
	})
}

func makeHTTPProxyConfig(port int) string {
	return `{
  "log": { "loglevel": "warning" },
  "inbounds": [
    {
      "listen": "127.0.0.1",
      "port": ` + strconv.Itoa(port) + `,
      "protocol": "http",
      "settings": {}
    }
  ],
  "outbounds": [
    {
      "protocol": "freedom",
      "settings": {}
    }
  ]
}`
}

func newHeadServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
}
