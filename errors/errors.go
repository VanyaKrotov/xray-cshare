package code_errors

import (
	"github.com/VanyaKrotov/xray_cshare/transfer"
)

type CodeError struct {
	Code    uint
	Message string
}

func New(message string, codes ...uint) *CodeError {
	var code uint = 1
	if len(codes) > 0 {
		code = codes[0]
	}

	return &CodeError{Message: message, Code: code}
}

func (e *CodeError) ToTransfer() *transfer.Response {
	return transfer.Error(e.Message, e.Code)
}
