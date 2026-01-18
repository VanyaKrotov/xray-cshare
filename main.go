package main

import "C"
import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/VanyaKrotov/xray_cshare/crypto_helpers"
	"github.com/VanyaKrotov/xray_cshare/transfer"
	"github.com/VanyaKrotov/xray_cshare/xray"

	"github.com/xtls/xray-core/core"
)

var (
	mu        sync.Mutex
	instances map[string]*core.Instance
)

func init() {
	instances = make(map[string]*core.Instance)
}

//export Start
func Start(cUuid *C.char, cJson *C.char) *C.char {
	mu.Lock()
	defer mu.Unlock()

	uuid := C.GoString(cUuid)
	if _, ok := instances[uuid]; ok {
		return C.CString(transfer.New(xray.XrayAlreadyStarted, "Xray server already started").ToString())
	}

	json := C.GoString(cJson)
	inst, resp := xray.Start(json)
	if resp.IsSuccess() {
		instances[uuid] = inst
	}

	return C.CString(resp.ToString())
}

//export Stop
func Stop(cUuid *C.char) *C.char {
	mu.Lock()
	defer mu.Unlock()

	uuid := C.GoString(cUuid)
	if instance, ok := instances[uuid]; ok {
		xray.Stop(instance)
		delete(instances, uuid)
	}

	return C.CString(transfer.Success("Server stopped").ToString())
}

//export IsStarted
func IsStarted(cUuid *C.char) C.int {
	mu.Lock()
	defer mu.Unlock()

	uuid := C.GoString(cUuid)
	if instance, ok := instances[uuid]; ok || !instance.IsRunning() {
		return 0
	}

	return 1
}

//export GetXrayCoreVersion
func GetXrayCoreVersion() *C.char {
	return C.CString(core.Version())
}

//export Curve25519Genkey
func Curve25519Genkey(cKey *C.char) *C.char {
	key := C.GoString(cKey)
	response := crypto_helpers.Curve25519Genkey(key, base64.RawURLEncoding)

	return C.CString(response.ToString())
}

//export Curve25519GenkeyWG
func Curve25519GenkeyWG(cKey *C.char) *C.char {
	key := C.GoString(cKey)
	response := crypto_helpers.Curve25519Genkey(key, base64.StdEncoding)

	return C.CString(response.ToString())
}

//export ExecuteUUID
func ExecuteUUID(cInput *C.char) *C.char {
	input := C.GoString(cInput)
	response := crypto_helpers.ExecuteUUID(input)

	return C.CString(response.ToString())
}

//export ExecuteMLDSA65
func ExecuteMLDSA65(cInput *C.char) *C.char {
	input := C.GoString(cInput)
	response := crypto_helpers.ExecuteMLDSA65(input)

	return C.CString(response.ToString())
}

//export ExecuteMLKEM768
func ExecuteMLKEM768(cInput *C.char) *C.char {
	input := C.GoString(cInput)
	response := crypto_helpers.ExecuteMLKEM768(input)

	return C.CString(response.ToString())
}

//export ExecuteVLESSEnc
func ExecuteVLESSEnc() *C.char {
	response := crypto_helpers.ExecuteVLESSEnc()

	return C.CString(response.ToString())
}

func main() {
	print(fmt.Sprintf("Xray core lib. Core version - %s", core.Version()))
}
