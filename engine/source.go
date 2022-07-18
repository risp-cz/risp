package engine

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/necessitates/clover"

	"risp/protocol"
)

type Source struct {
	ContextID    string
	ID           string
	CanonicalURI string
	AdapterType  AdapterType
	AdapterData  AdapterData
}

func (source *Source) Adapter(database *clover.DB, index bleve.Index) Adapter {
	switch source.AdapterType {
	case AdapterTypeFS:
		return NewAdapterFS(source, database, index)
	case AdapterTypeWeb:
		return NewAdapterWeb(source, database, index)
	}

	return nil
}

func (source *Source) MarshalURN() string {
	return fmt.Sprintf(
		"contexts/%s/sources/%s/%s",
		source.ContextID,
		source.AdapterType,
		url.PathEscape(source.CanonicalURI),
	)
}

func (source *Source) unmarshalURN(urn string) (err error) {
	pathParts := strings.Split(urn, "/")

	if len(pathParts) < 5 || pathParts[0] != "contexts" || pathParts[2] != "sources" {
		// return fmt.Errorf("invalid source URN scheme: '%s' (expected 'contexts/<CONTEXT_ID>/sources/<SOURCE_TYPE>/<CANONICAL_URI>')", urn)
		return fmt.Errorf("invalid source URN scheme: '%s' (expected 'contexts/<ID>/sources/<TYPE>/<path_escaped(URI)>')", urn)
	}

	source.ContextID = pathParts[1]
	source.AdapterType = AdapterType(pathParts[3])

	if source.CanonicalURI, err = url.PathUnescape(pathParts[4]); err != nil {
		return fmt.Errorf("malformatted source URN component: expected path encoded source URI, got '%s'", pathParts[4])
	}

	return
}

func (source *Source) MarshalMap() (value map[string]interface{}) {
	value = map[string]interface{}{
		"contextId":    source.ContextID,
		"adapterType":  source.AdapterType,
		"canonicalUri": source.CanonicalURI,
		"urn":          source.MarshalURN(),
	}

	if source.AdapterData != nil {
		value["adapterData"] = source.AdapterData.MarshalMap()
	}

	return
}

func (source *Source) MarshalProtocol() (sourceProto *protocol.Source) {
	sourceProto = &protocol.Source{
		ContextId:    source.ContextID,
		Id:           source.ID,
		CanonicalUri: source.CanonicalURI,
		Urn:          source.MarshalURN(),
	}

	switch source.AdapterType {
	case AdapterTypeFS:
		sourceProto.AdapterType = protocol.AdapterType_FS
	case AdapterTypeWeb:
		sourceProto.AdapterType = protocol.AdapterType_WEB
	}

	if source.AdapterData != nil {
		source.AdapterData.MarshalProtocol(sourceProto)
	}

	return
}

func (source *Source) UnmarshalMap(value map[string]interface{}) error {
	if value == nil {
		return fmt.Errorf("cannot unmarshal nil (type map[string]interface{}) to source")
	}

	if value["urn"] != nil && value["urn"] != "" {
		if err := source.unmarshalURN(value["urn"].(string)); err != nil {
			return err
		}
	} else {
		if value["adapterType"] != nil && value["adapterType"] != "" {
			source.AdapterType = AdapterType(value["adapterType"].(string))

			if source.Adapter(nil, nil) == nil {
				return fmt.Errorf("invalid adapter: '%s'", source.AdapterType)
			}
		} else {
			return fmt.Errorf("missing source adapter type")
		}

		if value["contextId"] != nil && value["contextId"] != "" {
			source.ContextID = value["contextId"].(string)
		} else {
			return fmt.Errorf("missing source contextId")
		}

		if value["canonicalUri"] != nil && value["canonicalUri"] != "" {
			source.CanonicalURI = value["canonicalUri"].(string)
		} else {
			return fmt.Errorf("missing source canonicalUri")
		}
	}

	return source.Adapter(nil, nil).UnmarshalMap(value)
}

func (source *Source) UnmarshalDBDocument(document *clover.Document) error {
	if document == nil {
		return fmt.Errorf("cannot unmarshal nil (type *clover.Document) to source")
	}

	if document.Get("urn") != nil && document.Get("urn") != "" {
		if err := source.unmarshalURN(document.Get("urn").(string)); err != nil {
			return err
		}
	} else {
		if document.Get("adapterType") != nil && document.Get("adapterType") != "" {
			source.AdapterType = AdapterType(document.Get("adapterType").(string))

			if source.Adapter(nil, nil) == nil {
				return fmt.Errorf("invalid adapter: '%s'", source.AdapterType)
			}
		} else {
			return fmt.Errorf("missing source adapter type")
		}

		if document.Get("contextId") != nil && document.Get("contextId") != "" {
			source.ContextID = document.Get("contextId").(string)
		} else {
			return fmt.Errorf("missing source contextId")
		}

		if document.Get("canonicalUri") != nil && document.Get("canonicalUri") != "" {
			source.CanonicalURI = document.Get("canonicalUri").(string)
		} else {
			return fmt.Errorf("missing source canonicalUri")
		}
	}

	return source.Adapter(nil, nil).UnmarshalDBDocument(document)
}
