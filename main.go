package main

import "C"
import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

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
		return C.CString(transfer.FailureString("Xray server already started", xray.XrayAlreadyStarted))
	}

	json := C.GoString(cJson)
	inst, err := xray.Start(json)
	if err == nil {
		instances[uuid] = inst

		return C.CString(transfer.SuccessString("Server started"))
	}

	return C.CString(err.ToTransfer().ToString())
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

	return C.CString(transfer.SuccessString("Server stopped"))
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
	result, err := crypto_helpers.Curve25519Genkey(key, base64.RawURLEncoding)
	if err != nil {
		return C.CString(transfer.FailureString(err.Error()))
	}

	return C.CString(transfer.SuccessString(fmt.Sprintf("%s|%s|%s", result.PrivateKey, result.Password, result.Hash32)))
}

//export Curve25519GenkeyWG
func Curve25519GenkeyWG(cKey *C.char) *C.char {
	key := C.GoString(cKey)
	result, err := crypto_helpers.Curve25519Genkey(key, base64.StdEncoding)
	if err != nil {
		return C.CString(transfer.FailureString(err.Error()))
	}

	return C.CString(transfer.SuccessString(fmt.Sprintf("%s|%s|%s", result.PrivateKey, result.Password, result.Hash32)))
}

//export ExecuteUUID
func ExecuteUUID(cInput *C.char) *C.char {
	input := C.GoString(cInput)
	uuid, err := crypto_helpers.ExecuteUUID(input)
	if err != nil {
		return C.CString(transfer.FailureString(err.Error()))
	}

	return C.CString(transfer.Success(uuid).ToString())
}

//export ExecuteMLDSA65
func ExecuteMLDSA65(cInput *C.char) *C.char {
	result, err := crypto_helpers.ExecuteMLDSA65(C.GoString(cInput))

	var response *transfer.Response
	if err == nil {
		response = transfer.Success(fmt.Sprintf("%s|%s", result.Seed, result.Verify))
	} else {
		response = transfer.Failure(err.Error())
	}

	return C.CString(response.ToString())
}

//export ExecuteMLKEM768
func ExecuteMLKEM768(cInput *C.char) *C.char {
	input := C.GoString(cInput)
	result, err := crypto_helpers.ExecuteMLKEM768(input)

	var response *transfer.Response
	if err == nil {
		response = transfer.Success(fmt.Sprintf("%s|%s|%s", result.Seed, result.Client, result.Hash32))
	} else {
		response = transfer.Failure(err.Error())
	}

	return C.CString(response.ToString())
}

//export ExecuteVLESSEnc
func ExecuteVLESSEnc() *C.char {
	result := crypto_helpers.ExecuteVLESSEnc()
	response := transfer.Success(fmt.Sprintf("%s|%s|%s|%s", result.Decryption, result.Encryption, result.DecryptionPQ, result.EncryptionPQ))

	return C.CString(response.ToString())
}

//export GenerateCert
func GenerateCert(cDomains *C.char, cCommonName *C.char, cOrg *C.char, cIsCA C.int, cExpire *C.char) *C.char {
	expireString := C.GoString(cExpire)

	var expire *time.Duration
	if len(expireString) > 0 {
		exp, err := time.ParseDuration(expireString)
		if err != nil {
			return C.CString(transfer.Failure(err.Error()).ToString())
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
		response = transfer.Success(fmt.Sprintf("%s|%s", strings.Join(result.Certificate, ","), strings.Join(result.Key, ",")))
	} else {
		response = transfer.Failure(err.Error())
	}

	return C.CString(response.ToString())
}

//export ExecuteCertChainHash
func ExecuteCertChainHash(cCert *C.char) *C.char {
	result, err := crypto_helpers.ExecuteCertChainHash(C.GoString(cCert))

	var response *transfer.Response
	if err == nil {
		response = transfer.Success(result)
	} else {
		response = transfer.Failure(err.Error())
	}

	return C.CString(response.ToString())
}

func main() {
	print(fmt.Sprintf("Xray core lib. Core version - %s", core.Version()))
}
