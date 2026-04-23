# xray-cshare

Languages: [English](README.md) | [Russian](RU_README.md)

`xray-cshare` is a C-compatible shared library wrapper around [`xray-core`](https://github.com/XTLS/Xray-core). It exposes a small ABI for starting and stopping Xray instances from JSON configs, checking connectivity through local proxy ports, querying the embedded `xray-core` version, and running helper routines for UUID, TLS, REALITY, VLESS encryption, and certificate-related data.

This repository is aimed at integrators that want to call Xray from another runtime through a generated shared library such as `.dll`, `.so`, or `.dylib`, rather than embedding Go packages directly.

## Features

- Start and stop `xray-core` instances from JSON configuration strings
- Check whether a named instance is currently running
- Ping a target URL directly or through local proxy ports
- Retrieve the embedded `xray-core` version
- Generate X25519, ML-DSA-65, ML-KEM-768, VLESS encryption helper values, UUIDs, and TLS certificate materials
- Calculate a SHA-256 certificate chain hash using Xray's TLS utilities
- Expose a compact binary response format suitable for FFI consumers

## Build

The project is built as a C-shared library.

Example local build:

```bash
go build -buildmode=c-shared -o build/xray_sdk.dll
```

The CI workflow currently builds:

- `windows_amd64.dll`
- `linux_amd64.so`
- `linux_arm64.so`
- `darwin_amd64.dylib`
- `darwin_arm64.dylib`
- `darwin_universal.dylib`

The generated header file contains the exported C ABI. A sample generated header is included at `build/xray_sdk.h`.

## Quick Overview

Typical usage flow:

1. Build the shared library and load it from your host application.
2. Call `Start(uuid, jsonConfig)` to create and run an Xray instance associated with your own UUID key.
3. Inspect the returned response buffer to determine success or failure.
4. Optionally call `IsStarted(uuid)`, `Ping(...)`, or `PingConfig(...)`.
5. Call `Stop(uuid)` when the instance is no longer needed.
6. Release any returned buffers with `FreePointer`.

## Response Buffer Format

Most exported functions return a raw pointer to a buffer allocated with `C.malloc`. The buffer format is defined in `transfer/response.go`.

- Bytes `0..3`: little-endian `uint32` status code
- Bytes `4..5`: little-endian `uint16` content type
- Bytes `6..7`: unused padding
- Bytes starting at offset `8`: null-terminated UTF-8 body string

Content type values:

- `1`: error string
- `2`: JSON payload string
- `3`: plain success message string

Status code behavior:

- `0`: success
- non-zero: error code

Memory management:

- Any pointer returned by an exported function that allocates a response must be released with `FreePointer`.
- `Stop` and `IsStarted` do not allocate response buffers.

## Error Codes

The wrapper defines the following explicit codes in the source:

| Code | Name | Meaning |
| --- | --- | --- |
| `1` | `JsonParseError` | JSON config parsing failed |
| `2` | `LoadConfigError` | Xray config build failed |
| `3` | `InitXrayError` | `core.New(...)` failed |
| `4` | `StartXrayError` | `instance.Start()` failed |
| `5` | `XrayAlreadyStarted` | The same UUID already maps to a running instance |
| `6` | `PingTimeoutError` | Declared in code but not currently returned by the public API |
| `7` | `PingError` | Declared in code but not currently returned by the public API |

For helper functions, general failures are typically returned as code `1` with an error string in the response body.

## Public API

The exported C signatures below match `build/xray_sdk.h`.

### Instance Lifecycle

#### `void* Start(char* cUuid, char* cJson);`

Starts an Xray instance from a JSON config and stores it under the caller-provided UUID key.

Parameters:

- `cUuid`: caller-defined identifier used as the instance map key
- `cJson`: full Xray JSON configuration

Success:

- Content type `3`
- Message body: `"Server started"`

Failure:

- Returns content type `1`
- Error codes may be `1`, `2`, `3`, `4`, or `5`

Underlying `xray-core` calls:

- `serial.DecodeJSONConfig(strings.NewReader(jsonConfig))`
- `config.Build()`
- `core.New(coreCfg)`
- `instance.Start()`
- `instance.Close()` on start failure cleanup

Notes:

- If the UUID already exists and the stored instance is still running, the call fails with code `5`.
- Successful instances are stored in an in-memory map protected by a mutex.

#### `void Stop(char* cUuid);`

Stops and removes the Xray instance stored for the given UUID.

Parameters:

- `cUuid`: instance key used in `Start`

Success:

- No return value

Underlying `xray-core` calls:

- `instance.Close()`

Notes:

- The function is a no-op if the UUID is not found.
- The Go wrapper also triggers `runtime.GC()` and `debug.FreeOSMemory()` after shutdown.

#### `int IsStarted(char* cUuid);`

Checks whether a stored instance exists and is running.

Parameters:

- `cUuid`: instance key used in `Start`

Success:

- Returns `1` if the instance exists and `instance.IsRunning()` is true
- Returns `0` otherwise

Underlying `xray-core` calls:

- `instance.IsRunning()`

### Connectivity Helpers

#### `void* PingConfig(char* jsonConfig, int* portsPtr, int portsLen, char* testingURL);`

Starts a temporary Xray instance from the provided JSON config, then performs HTTP `HEAD` requests through the given local ports.

Parameters:

- `jsonConfig`: Xray JSON config used for the temporary instance
- `portsPtr`: pointer to an array of local ports
- `portsLen`: number of ports in `portsPtr`
- `testingURL`: target URL for HTTP `HEAD`

Success:

- Content type `2`
- JSON array of:

```json
[
  {
    "port": 1080,
    "timeout": 123,
    "error": ""
  }
]
```

Failure:

- Content type `1`
- Error message string

Underlying `xray-core` calls:

- Indirectly uses the same startup flow as `Start` through `xray.Start(jsonConfig)`
- Calls `instance.Close()` after all probes finish

Notes:

- Each port is tested concurrently.
- A timeout is reported in milliseconds.
- The function does not store the instance in the global UUID map.

#### `void* Ping(int port, char* testingURL);`

Performs a direct HTTP `HEAD` request or a proxied one through `127.0.0.1:<port>`.

Parameters:

- `port`: proxy port; `0` means direct request without proxy
- `testingURL`: target URL for HTTP `HEAD`

Success:

- Content type `2`
- JSON object:

```json
{
  "port": 1080,
  "timeout": 123,
  "error": ""
}
```

Failure behavior:

- The function still returns a payload object
- On request failure, `timeout` is typically `-1` and `error` contains the runtime error string

Underlying `xray-core` calls:

- None directly

### Version

#### `void* GetXrayCoreVersion(void);`

Returns the embedded `xray-core` version string.

Success:

- Content type `3`
- Message body is the result of `core.Version()`

Underlying `xray-core` calls:

- `core.Version()`

### Cryptographic Helpers

#### `void* Curve25519Genkey(char* cKey);`

Generates or re-derives X25519 key material using URL-safe base64 encoding without padding.

Parameters:

- `cKey`: optional base64-encoded 32-byte private key in `base64.RawURLEncoding`

Success:

- Content type `2`
- JSON object:

```json
{
  "private_key": "...",
  "password": "...",
  "hash32": "..."
}
```

Failure:

- Content type `1`
- Error string such as invalid private key length

Underlying `xray-core` calls:

- None directly

#### `void* Curve25519GenkeyWG(char* cKey);`

Same behavior as `Curve25519Genkey`, but uses standard padded base64 encoding intended for WireGuard-style formatting.

Underlying `xray-core` calls:

- None directly

#### `void* ExecuteUUID(char* cInput);`

Generates or normalizes a UUID string using Xray's UUID package.

Parameters:

- `cInput`: empty string for a random UUID, or a short input that is passed to `uuid.ParseString`

Success:

- Content type `3`
- Message body contains the UUID string

Failure:

- Content type `1`
- Error string, for example when input length exceeds 30 bytes

Underlying `xray-core` calls:

- `uuid.New()`
- `uuid.ParseString(input)`

Notes:

- The source code comment mentions UUIDv5/VLESS wording, but the actual implementation only generates a random UUID for empty input or parses the supplied string.

#### `void* ExecuteMLDSA65(char* cInput);`

Generates ML-DSA-65 verification material from a supplied or random seed.

Parameters:

- `cInput`: optional 32-byte seed encoded with `base64.RawURLEncoding`

Success:

- Content type `2`
- JSON object:

```json
{
  "seed": "...",
  "verify": "..."
}
```

Underlying `xray-core` calls:

- None directly

#### `void* ExecuteMLKEM768(char* cInput);`

Generates ML-KEM-768 encapsulation material from a supplied or random seed.

Parameters:

- `cInput`: optional 64-byte seed encoded with `base64.RawURLEncoding`

Success:

- Content type `2`
- JSON object:

```json
{
  "seed": "...",
  "client": "...",
  "hash32": "..."
}
```

Underlying `xray-core` calls:

- None directly

#### `void* ExecuteVLESSEnc(void);`

Generates helper strings for VLESS encryption and post-quantum variants.

Success:

- Content type `2`
- JSON object:

```json
{
  "decryption": "...",
  "encryption": "...",
  "decryption_pq": "...",
  "encryption_pq": "..."
}
```

Underlying `xray-core` calls:

- None directly

### TLS and Certificate Helpers

#### `void* GenerateCert(char* cDomains, char* cCommonName, char* cOrg, int cIsCA, char* cExpire);`

Generates a PEM certificate and private key pair.

Parameters:

- `cDomains`: comma-separated DNS names
- `cCommonName`: certificate common name
- `cOrg`: organization name
- `cIsCA`: `1` to generate a CA-style certificate, otherwise `0`
- `cExpire`: optional Go duration string such as `2160h`

Success:

- Content type `2`
- JSON object:

```json
{
  "certificate": [
    "-----BEGIN CERTIFICATE-----",
    "..."
  ],
  "key": [
    "-----BEGIN PRIVATE KEY-----",
    "..."
  ]
}
```

Failure:

- Content type `1`
- Error string

Underlying `xray-core` calls:

- `cert.Generate(nil, opts...)`
- `cert.Authority(true)`
- `cert.KeyUsage(...)`
- `cert.NotAfter(...)`
- `cert.CommonName(...)`
- `cert.DNSNames(...)`
- `cert.Organization(...)`

Notes:

- The implementation uses Xray's TLS certificate generation package.
- The source intends to default the expiry to 90 days and names to `"Xray Inc"`, but the current implementation should be treated carefully because `initDefaults()` dereferences `Expire` when it is `nil`. This README documents the actual code path rather than assuming corrected behavior.

#### `void* ExecuteCertChainHash(char* cCert);`

Calculates the SHA-256 hash of a PEM certificate chain using Xray's TLS helper.

Parameters:

- `cCert`: either PEM text content or a filesystem path to a PEM file

Success:

- Content type `3`
- Message body contains the hash string

Failure:

- Content type `1`
- Error string if file reading fails

Underlying `xray-core` calls:

- `tls.CalculatePEMCertChainSHA256Hash(certContent)`

### Utility Functions

#### `void* SetEnv(char* cKey, char* cValue);`

Sets a process environment variable inside the host process.

Parameters:

- `cKey`: environment variable name
- `cValue`: environment variable value

Success:

- Content type `3`
- Message body: `"done"`

Failure:

- Content type `1`
- Error string from `os.Setenv`

Underlying `xray-core` calls:

- None

#### `void FreePointer(void* ptr);`

Releases a buffer previously returned by this library.

Parameters:

- `ptr`: pointer returned by a response-producing function

Success:

- No return value

Notes:

- Only free pointers allocated by this library.

## How This Wrapper Uses `xray-core`

### Direct runtime integration

The main runtime path in `xray/server.go` uses `xray-core` directly:

- `infra/conf/serial.DecodeJSONConfig` parses JSON input
- `config.Build()` converts the parsed config into a core config
- `core.New(...)` creates a new Xray instance
- `(*core.Instance).Start()` starts it
- `(*core.Instance).Close()` stops or cleans it up
- `core.Version()` exposes the embedded core version

### Registration through blank imports

`xray/imports.go` imports many `xray-core` packages for their `init()` side effects so that handlers, features, proxies, transports, and transport headers are registered before configs are built.

This includes:

- mandatory app features such as dispatcher and proxyman
- optional services such as commander, stats, routing, logging, DNS, metrics, and observatory
- inbound and outbound proxies such as VLESS, VMess, SOCKS, Trojan, Shadowsocks, Freedom, DNS, and WireGuard
- transports such as gRPC, TCP, KCP, TLS, WebSocket, SplitHTTP, HTTP upgrade, UDP, and REALITY
- transport headers such as HTTP, SRTP, UTP, WeChat, TLS, and WireGuard

Without these imports, config loading may fail because components referenced in JSON would not be registered.

### Helper usage through `xray-core` packages

Some helper functions are not part of the runtime lifecycle, but still depend on `xray-core` packages:

- `ExecuteUUID` uses `github.com/xtls/xray-core/common/uuid`
- `GenerateCert` uses `github.com/xtls/xray-core/common/protocol/tls/cert`
- `ExecuteCertChainHash` uses `github.com/xtls/xray-core/transport/internet/tls`
- `crypto_helpers/cert.go` also imports `github.com/xtls/xray-core/main/commands/base` for `stringList.Set`

## Example FFI Workflow

The exact host-side code depends on your language, but the expected lifecycle looks like this:

```c
void* resp = Start("instance-1", "{...json config...}");
/* read status code, content type, and body from the packed buffer */
FreePointer(resp);

if (IsStarted("instance-1")) {
    void* version = GetXrayCoreVersion();
    /* read packed response */
    FreePointer(version);
}

Stop("instance-1");
```

When decoding a returned pointer:

1. Read a 32-bit little-endian status code from offset `0`.
2. Read a 16-bit little-endian content type from offset `4`.
3. Read the null-terminated body string starting at offset `8`.
4. Call `FreePointer` once you are done with the buffer.

## Known Caveats

- Instance ownership is keyed entirely by the UUID string provided by the caller.
- The instance map exists only in process memory and is not persisted.
- `PingConfig` creates a temporary Xray instance from the supplied config, performs checks, and closes it afterward.
- `Ping` returns a payload object even when the request fails; in that case the error is embedded inside the JSON payload.
- `PingTimeoutError` and `PingError` constants exist in code but are not currently surfaced as dedicated response codes by the exported API.
- `GenerateCert` should be consumed with caution when `cExpire` is empty, because the current implementation path around default expiry initialization is unsafe as written.
- `ExecuteCertChainHash` treats its input as a file path if that path exists on disk; otherwise it treats the input as PEM content.

## Source of Truth

This README is based on the current implementation in:

- `main.go`
- `xray/server.go`
- `xray/imports.go`
- `crypto_helpers/*.go`
- `testing/testing.go`
- `transfer/response.go`
- `build/xray_sdk.h`

If implementation and documentation ever diverge, treat the source code and generated header as authoritative.
