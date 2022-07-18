package engine

import (
	"github.com/necessitates/clover"

	"risp/protocol"
)

type AdapterType string

const (
	AdapterTypeFS  AdapterType = "fs"
	AdapterTypeWeb AdapterType = "web"
)

type Adapter interface {
	Type() AdapterType

	UnmarshalMap(value map[string]interface{}) error
	UnmarshalDBDocument(document *clover.Document) error

	Index() error
}

type AdapterData interface {
	MarshalMap() map[string]interface{}
	MarshalProtocol(*protocol.Source)

	UnmarshalMap(map[string]interface{}) error
	UnmarshalDBDocument(*clover.Document) error
}
