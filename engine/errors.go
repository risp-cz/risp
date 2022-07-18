package engine

import (
	"risp/protocol"
)

const (
	ErrAllGood int64 = iota
	ErrUnknown
	ErrInvalidCommand
	ErrInvalidQuery
	ErrInvalidSourceURI
	ErrInvalidContext
	ErrInvalidSource
	ErrInvalidResource
)

func NewProtocolError(opts ...interface{}) *protocol.Error {
	protocolError := &protocol.Error{
		Code: ErrAllGood,
	}

	for _, opt := range opts {
		switch _opt := opt.(type) {
		case int64:
			protocolError.Code = _opt
		case string:
			protocolError.Message = &_opt
		case error:
			message := _opt.Error()
			protocolError.Message = &message
		}
	}

	return protocolError
}
