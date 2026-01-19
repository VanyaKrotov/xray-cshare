package transfer

import "fmt"

type Response struct {
	Code    int
	Message string
}

func Failure(message string, codes ...int) *Response {
	var code int = 1
	if len(codes) > 0 {
		code = codes[0]
	}

	return &Response{Code: code, Message: message}
}

func FailureString(message string, codes ...int) string {

	return Failure(message, codes...).ToString()
}

func Success(message string) *Response {
	return &Response{Code: 0, Message: message}
}

func SuccessString(message string) string {
	return Success(message).ToString()
}

func (r Response) ToString() string {
	return fmt.Sprintf("%d|%s", r.Code, r.Message)
}

func (r Response) IsSuccess() bool {
	return r.Code == 0
}
