package main

/*
#include <stdlib.h>
*/
import "C"

import "unsafe"

func cStringPtr(value string) *C.char {
	return C.CString(value)
}

func freeCString(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

func startString(uuid string, json string) unsafe.Pointer {
	cUUID := cStringPtr(uuid)
	defer freeCString(cUUID)

	cJSON := cStringPtr(json)
	defer freeCString(cJSON)

	return Start(cUUID, cJSON)
}

func stopString(uuid string) {
	cUUID := cStringPtr(uuid)
	defer freeCString(cUUID)

	Stop(cUUID)
}

func isStartedString(uuid string) int {
	cUUID := cStringPtr(uuid)
	defer freeCString(cUUID)

	return int(IsStarted(cUUID))
}

func pingConfigInts(json string, ports []int, testingURL string) unsafe.Pointer {
	cJSON := cStringPtr(json)
	defer freeCString(cJSON)

	cURL := cStringPtr(testingURL)
	defer freeCString(cURL)

	cPorts := make([]C.int, len(ports))
	for i, port := range ports {
		cPorts[i] = C.int(port)
	}

	if len(cPorts) == 0 {
		return PingConfig(cJSON, nil, 0, cURL)
	}

	return PingConfig(cJSON, &cPorts[0], C.int(len(cPorts)), cURL)
}

func pingPort(port int, testingURL string) unsafe.Pointer {
	cURL := cStringPtr(testingURL)
	defer freeCString(cURL)

	return Ping(C.int(port), cURL)
}

func curve25519GenkeyString(key string) unsafe.Pointer {
	cKey := cStringPtr(key)
	defer freeCString(cKey)

	return Curve25519Genkey(cKey)
}

func curve25519GenkeyWGString(key string) unsafe.Pointer {
	cKey := cStringPtr(key)
	defer freeCString(cKey)

	return Curve25519GenkeyWG(cKey)
}

func executeUUIDString(input string) unsafe.Pointer {
	cInput := cStringPtr(input)
	defer freeCString(cInput)

	return ExecuteUUID(cInput)
}

func executeMLDSA65String(input string) unsafe.Pointer {
	cInput := cStringPtr(input)
	defer freeCString(cInput)

	return ExecuteMLDSA65(cInput)
}

func executeMLKEM768String(input string) unsafe.Pointer {
	cInput := cStringPtr(input)
	defer freeCString(cInput)

	return ExecuteMLKEM768(cInput)
}

func generateCertStrings(domains string, commonName string, org string, isCA int, expire string) unsafe.Pointer {
	cDomains := cStringPtr(domains)
	defer freeCString(cDomains)

	cCommonName := cStringPtr(commonName)
	defer freeCString(cCommonName)

	cOrg := cStringPtr(org)
	defer freeCString(cOrg)

	cExpire := cStringPtr(expire)
	defer freeCString(cExpire)

	return GenerateCert(cDomains, cCommonName, cOrg, C.int(isCA), cExpire)
}

func executeCertChainHashString(cert string) unsafe.Pointer {
	cCert := cStringPtr(cert)
	defer freeCString(cCert)

	return ExecuteCertChainHash(cCert)
}

func setEnvStrings(key string, value string) unsafe.Pointer {
	cKey := cStringPtr(key)
	defer freeCString(cKey)

	cValue := cStringPtr(value)
	defer freeCString(cValue)

	return SetEnv(cKey, cValue)
}
