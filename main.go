package main

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/VanyaKrotov/xray_cshare/crypto_helpers"
	"github.com/VanyaKrotov/xray_cshare/testing"
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
func Start(cUuid *C.char, cJson *C.char) unsafe.Pointer {
	mu.Lock()
	defer mu.Unlock()

	uuid := C.GoString(cUuid)
	if instance, ok := instances[uuid]; ok && instance.IsRunning() {
		return transfer.Error("Xray server already started", xray.XrayAlreadyStarted).Pack()
	}

	json := C.GoString(cJson)
	inst, err := xray.Start(json)
	if err == nil {
		instances[uuid] = inst

		return transfer.Ok("Server started").Pack()
	}

	return err.ToTransfer().Pack()
}

//export Stop
func Stop(cUuid *C.char) {
	mu.Lock()
	defer mu.Unlock()

	uuid := C.GoString(cUuid)
	if instance, ok := instances[uuid]; ok {
		xray.Stop(instance)
		delete(instances, uuid)
	}
}

//export IsStarted
func IsStarted(cUuid *C.char) C.int {
	mu.Lock()
	defer mu.Unlock()

	uuid := C.GoString(cUuid)
	if instance, ok := instances[uuid]; !ok || !instance.IsRunning() {
		return 0
	}

	return 1
}

//export PingConfig
func PingConfig(jsonConfig *C.char, portsPtr *C.int, portsLen C.int, testingURL *C.char) unsafe.Pointer {
	goJSON := C.GoString(jsonConfig)
	url := C.GoString(testingURL)

	cPorts := unsafe.Slice(portsPtr, portsLen)
	ports := make([]int, int(portsLen))
	for i, v := range cPorts {
		ports[i] = int(v)
	}

	results, err := testing.PingConfig(goJSON, ports, url)
	if err != nil {
		return transfer.Error(err.Error()).Pack()
	}

	return transfer.Payload(results).Pack()
}

//export Ping
func Ping(port C.int, testingURL *C.char) unsafe.Pointer {
	url := C.GoString(testingURL)
	ping, err := testing.Ping(int(port), url)

	result := testing.PingResult{
		Port:    int(port),
		Timeout: ping,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return transfer.Payload(result).Pack()
}

//export GetXrayCoreVersion
func GetXrayCoreVersion() unsafe.Pointer {
	return transfer.Ok(core.Version()).Pack()
}

//export Curve25519Genkey
func Curve25519Genkey(cKey *C.char) unsafe.Pointer {
	key := C.GoString(cKey)
	result, err := crypto_helpers.Curve25519Genkey(key, base64.RawURLEncoding)
	if err != nil {
		return transfer.Error(err.Error()).Pack()
	}

	return transfer.Payload(result).Pack()
}

//export Curve25519GenkeyWG
func Curve25519GenkeyWG(cKey *C.char) unsafe.Pointer {
	key := C.GoString(cKey)
	result, err := crypto_helpers.Curve25519Genkey(key, base64.StdEncoding)
	if err != nil {
		return transfer.Error(err.Error()).Pack()
	}

	return transfer.Payload(result).Pack()
}

//export ExecuteUUID
func ExecuteUUID(cInput *C.char) unsafe.Pointer {
	input := C.GoString(cInput)
	uuid, err := crypto_helpers.ExecuteUUID(input)
	if err != nil {
		return transfer.Error(err.Error()).Pack()
	}

	return transfer.Ok(uuid).Pack()
}

//export ExecuteMLDSA65
func ExecuteMLDSA65(cInput *C.char) unsafe.Pointer {
	result, err := crypto_helpers.ExecuteMLDSA65(C.GoString(cInput))

	var response *transfer.Response
	if err == nil {
		response = transfer.Payload(result)
	} else {
		response = transfer.Error(err.Error())
	}

	return response.Pack()
}

//export ExecuteMLKEM768
func ExecuteMLKEM768(cInput *C.char) unsafe.Pointer {
	input := C.GoString(cInput)
	result, err := crypto_helpers.ExecuteMLKEM768(input)

	var response *transfer.Response
	if err == nil {
		response = transfer.Payload(result)
	} else {
		response = transfer.Error(err.Error())
	}

	return response.Pack()
}

//export ExecuteVLESSEnc
func ExecuteVLESSEnc() unsafe.Pointer {
	result := crypto_helpers.ExecuteVLESSEnc()

	return transfer.Payload(result).Pack()
}

//export GenerateCert
func GenerateCert(cDomains *C.char, cCommonName *C.char, cOrg *C.char, cIsCA C.int, cExpire *C.char) unsafe.Pointer {
	expireString := C.GoString(cExpire)

	var expire *time.Duration
	if len(expireString) > 0 {
		exp, err := time.ParseDuration(expireString)
		if err != nil {
			return transfer.Error(err.Error()).Pack()
		}

		expire = &exp
	}

	options := crypto_helpers.GenCertOptions{
		DomainNames:  strings.Split(C.GoString(cDomains), ","),
		IsCA:         cIsCA == 1,
		CommonName:   C.GoString(cCommonName),
		Organization: C.GoString(cOrg),
		Expire:       expire,
	}
	result, err := crypto_helpers.GenerateCert(options)

	var response *transfer.Response
	if err == nil {
		response = transfer.Payload(result)
	} else {
		response = transfer.Error(err.Error())
	}

	return response.Pack()
}

//export ExecuteCertChainHash
func ExecuteCertChainHash(cCert *C.char) unsafe.Pointer {
	result, err := crypto_helpers.ExecuteCertChainHash(C.GoString(cCert))

	var response *transfer.Response
	if err == nil {
		response = transfer.Ok(result)
	} else {
		response = transfer.Error(err.Error())
	}

	return response.Pack()
}

//export FreePointer
func FreePointer(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}

//export SetEnv
func SetEnv(cKey, cValue *C.char) unsafe.Pointer {
	err := os.Setenv(C.GoString(cKey), C.GoString(cValue))
	if err != nil {
		return transfer.Error(err.Error()).Pack()
	}

	return transfer.Ok("done").Pack()
}

func main() {
	print(fmt.Sprintf("Xray core lib. Core version - %s", core.Version()))
}
