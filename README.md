# xray-cshare (v25.10.15)

**Description**: A small wrapper library for `xray-core` with cryptographic helper utilities. It provides a C-compatible API to manage Xray instances and several utilities for key generation and encoding.

**Files**:
- [main.go](main.go): C API exports for external applications.
- [xray/server.go](xray/server.go): Start/stop logic for `xray-core`.
- [transfer/response.go](transfer/response.go): `Response` structure for unified replies.
- [crypto_helpers/*](crypto_helpers): Cryptographic utilities (Curve25519, ML-DSA-65, ML-KEM-768, UUID, etc.).

**Exported C API (in `main.go`)**
- `Start(cUuid *C.char, cJson *C.char) *C.char`: Starts an Xray instance using a JSON config. Returns a response string in the format `"<code>|<message>"`. On success, an instance is created and stored by UUID.
- `Stop(cUuid *C.char) *C.char`: Stops the server associated with the given UUID and frees resources. Returns a response string.
- `IsStarted(cUuid *C.char) C.int`: Checks whether an instance is started for the given UUID. Returns `0` or `1` (C.int).
- `GetXrayCoreVersion() *C.char`: Returns the embedded `xray-core` version string.
- `Curve25519Genkey(cKey *C.char) *C.char`: Generates or accepts an existing Curve25519 private key in base64 (RawURLEncoding). Returns private key, public key and a 32-byte hash encoded in the same base64 encoding.
- `Curve25519GenkeyWG(cKey *C.char) *C.char`: Same as `Curve25519Genkey` but uses `base64.StdEncoding` (compatibility for WireGuard).
- `ExecuteUUID(cInput *C.char) *C.char`: Generates or normalizes a UUID (input length must be ≤ 30 bytes when provided).
- `ExecuteMLDSA65(cInput *C.char) *C.char`: Generates an ML-DSA-65 keypair. Accepts an optional seed in base64.
- `ExecuteMLKEM768(cInput *C.char) *C.char`: Generates ML-KEM-768 key material. Accepts an optional seed in base64.
- `ExecuteVLESSEnc() *C.char`: Returns a set of VLESS configuration strings (X25519 and ML-KEM-768 variants).

**Package `xray` (file: [xray/server.go](xray/server.go))**
- `Start(jsonConfig string) (*core.Instance, *transfer.Response)`: Parses JSON configuration, builds `xray-core` configuration, creates and starts an instance. Returns `nil` and a `*transfer.Response` with an error code on failure.
- `Stop(instance *core.Instance)`: Gracefully stops the `xray-core` instance, closes resources and triggers OS memory cleanup.

**Package `transfer` (file: [transfer/response.go](transfer/response.go))**
- `type Response struct { Code int; Message string }`: Unified response structure.
- `New(code int, message string) *Response`: Constructor with custom code and message.
- `Success(message string) *Response`: Helper constructor with `Code == 0`.
- `(r Response) ToString() string`: Serializes response to `"<code>|<message>"`.
- `(r Response) IsSuccess() bool`: Returns `true` when `Code == 0`.

**Package `crypto_helpers` (files: [crypto_helpers](crypto_helpers))**
- `ExecuteUUID(input string) *transfer.Response` (uuid.go): Generates a new UUID when `input` is empty; parses and normalizes when input length ≤ 30; otherwise returns an error.
- `ExecuteMLDSA65(input string) *transfer.Response` (MLDSA65.go): Generates an ML-DSA-65 keypair from an optional seed (base64 Raw). Returns seed and public key in base64.
- `ExecuteMLKEM768(input_mlkem768 string) *transfer.Response` (MLKEM768.go): Generates ML-KEM-768 seed/client/hash; accepts an optional seed (base64 Raw). Returns seed, client and hash in base64.
- `genMLKEM768(inputSeed *[64]byte) (seed [64]byte, client []byte, hash32 [32]byte)`: Internal helper for ML-KEM-768 generation (not exported).
- `Curve25519Genkey(input_base64 string, enc *base64.Encoding) *transfer.Response` (x25519.go): Generates X25519 private key or accepts an existing private key in base64. Returns `privateKey|password|hash32` in the specified encoding.
- `genCurve25519(inputPrivateKey []byte) (privateKey []byte, password []byte, hash32 [32]byte, returnErr error)`: Internal helper for X25519 generation/normalization (not exported).
- `ExecuteVLESSEnc() *transfer.Response` (VLESSEnc.go): Assembles VLESS configuration pairs for X25519 and ML-KEM-768, returning four dot-joined strings separated by `|`.
- `generateDotConfig(fields ...string) string`: Small helper that joins fields with a dot (not exported).

**Error codes (some constants defined in code)**
- `JsonParseError` / `LoadConfigError` / `InitXrayError` / `StartXrayError` / `XrayAlreadyStarted` — errors related to `xray-core` startup (see [xray/server.go](xray/server.go)).
- `UuidInputOverflow` — input length error for UUID (see `crypto_helpers/uuid.go`).
- Other file-specific codes are defined in the respective files (`crypto_helpers/x25519.go`, `MLDSA65.go`, etc.).