package response

import (
	"errors"
	"fmt"
)

func NewServiceError(code int32, msg string) func() *Error {
	return func() *Error {
		return &Error{
			Code: code,
			Msg:  msg,
		}
	}
}

type ServiceError interface {
	ErrCode() int32
	ErrMsg() string
	ErrReason() string
}

type Error struct {
	Code   int32
	Msg    string
	Reason string
}

func (e *Error) Into(code int32, message string) {
	e.Code, e.Msg = code, message
}

func (e *Error) Error() string {
	return fmt.Sprintf("code=(%v),msg=(%s),reason=(%s)", e.Code, e.Msg, e.Reason)
}

func (e *Error) ErrCode() int32 {
	return e.Code
}

func (e *Error) ErrMsg() string {
	return e.Msg
}

func (e *Error) ErrReason() string {
	return e.Reason
}

func (e *Error) SetMsg(msg string) *Error {
	e.Msg = msg
	return e
}

func (e *Error) SetErrReason(reason string) *Error {
	e.Reason = reason
	return e
}

func (e *Error) Is(err error) bool {
	if err == nil {
		return false
	}

	ne := new(Error)
	if errors.As(err, &ne) {
		return e.Code == ne.Code && e.Msg == ne.Msg
	}

	return false
}
