package code_errors

import (
	"fmt"

	"github.com/VanyaKrotov/xray_cshare/transfer"
)

type CodeError struct {
	Code    int
	Message string
}

func New(message string, codes ...int) *CodeError {
	code := 1
	if len(codes) > 0 {
		code = codes[0]
	}

	return &CodeError{Message: message, Code: code}
}

func (e *CodeError) ToString() string {
	return fmt.Sprintf("%d|%s", e.Code, e.Message)
}

func (e *CodeError) ToTransfer() *transfer.Response {
	return transfer.Failure(e.Message, e.Code)
}
