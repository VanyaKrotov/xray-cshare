package transfer

import "C"
import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"unsafe"
)

const (
	_offset int = 8

	_Error   uint16 = 1
	_Payload uint16 = 2
	_Message uint16 = 3
)

type Response struct {
	Code        uint
	Message     string
	contentType uint16
}

func Error(message string, codes ...uint) *Response {
	var code uint = 1
	if len(codes) > 0 {
		code = codes[0]
	}

	return &Response{Code: code, Message: message, contentType: _Error}
}

func Ok(message string) *Response {
	return &Response{Code: 0, Message: message, contentType: _Message}
}

func okContent(message string, contentType uint16) *Response {
	return &Response{Code: 0, Message: message, contentType: contentType}
}

func Payload(content interface{}) *Response {
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		return Error(fmt.Sprintf("failed to marshal payload: %v", err), 1)
	}

	return okContent(string(jsonBytes), _Payload)
}

func (r Response) Pack() unsafe.Pointer {
	bodyLen := len(r.Message)
	totalSize := _offset + bodyLen + 1

	ptr := C.malloc(C.size_t(totalSize))
	if ptr == nil {
		return nil
	}

	buffer := (*[1 << 6]byte)(ptr)[:totalSize:totalSize]

	// 1. Write code (4 byte)
	binary.LittleEndian.PutUint32(buffer[0:4], uint32(r.Code))

	// 2. Write content type (8 byte)
	binary.LittleEndian.PutUint16(buffer[4:6], uint16(r.contentType))

	// 2. Write body
	copy(buffer[_offset:], []byte(r.Message))

	buffer[totalSize-1] = 0

	return ptr
}

func (r Response) IsSuccess() bool {
	return r.Code == 0
}
