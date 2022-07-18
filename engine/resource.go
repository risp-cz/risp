package engine

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/necessitates/clover"

	"risp/protocol"
)

type ResourceType string

func (resourceType ResourceType) String() string {
	return string(resourceType)
}

type Resource interface {
	ID() *string
	SetID(id string)

	Type() ResourceType
	SourceURN() string
	CanonicalURI() string

	MarshalURN() string
	MarshalMap() map[string]interface{}
	MarshalRecord(Record)
	MarshalProtocol() *protocol.Resource

	UnmarshalMap(map[string]interface{}) error
	UnmarshalDBDocument(*clover.Document) error

	Index(Adapter) error
}

func UnmarshalResource(resource interface{}) (Resource, error) {
	var err error

	resolveResourceType := func(resourceType string) (Resource, error) {
		resourceBase := &ResourceBase{
			resourceType: ResourceType(resourceType),
		}

		switch resourceBase.resourceType {
		case ResFSFile:
			return &ResourceFSFile{ResourceBase: resourceBase}, nil
		case ResWebPage:
			return &ResourceWebPage{ResourceBase: resourceBase}, nil
		}

		return nil, fmt.Errorf("invalid resource type '%s'", resourceType)
	}

	switch item := resource.(type) {
	case *clover.Document:
		var resource Resource

		if !item.Has("type") || item.Get("type") == "" {
			return resource, fmt.Errorf("missing resource type")
		}

		if resource, err = resolveResourceType(item.Get("type").(string)); err != nil {
			return resource, err
		}

		return resource, resource.UnmarshalDBDocument(item)
	case map[string]interface{}:
		var resource Resource

		if item["type"] == nil || item["type"] == "" {
			return resource, fmt.Errorf("missing resource type")
		}

		if resource, err = resolveResourceType(item["type"].(string)); err != nil {
			return resource, err
		}

		return resource, resource.UnmarshalMap(item)
	}

	err = fmt.Errorf("cannot unmarshal resource: invalid type '%T'", resource)
	return nil, err
}

type ResourceBase struct {
	id           *string
	canonicalURI string
	contextID    string
	sourceID     string
	source       *Source
	resourceType ResourceType
}

func (resourceBase *ResourceBase) ID() *string {
	return resourceBase.id
}

func (resourceBase *ResourceBase) SetID(id string) {
	resourceBase.id = &id
}

func (resourceBase *ResourceBase) Type() ResourceType {
	return resourceBase.resourceType
}

func (resourceBase *ResourceBase) SourceURN() string {
	return resourceBase.source.MarshalURN()
}

func (resourceBase *ResourceBase) CanonicalURI() string {
	return resourceBase.canonicalURI
}

func (resourceBase *ResourceBase) MarshalURN() string {
	return fmt.Sprintf(
		"%s/resources/%s/%s",
		resourceBase.source.MarshalURN(),
		resourceBase.Type(),
		url.PathEscape(resourceBase.canonicalURI),
	)
}

func (resourceBase *ResourceBase) unmarshalURN(urn string) (err error) {
	if resourceBase.source == nil {
		resourceBase.source = &Source{}
	}

	if err = resourceBase.source.unmarshalURN(urn); err != nil {
		return fmt.Errorf("invalid resource URN: %s", err)
	}

	resourceBase.contextID = resourceBase.source.ContextID

	pathParts := strings.Split(urn, "/")

	if len(pathParts) < 8 || pathParts[5] != "resources" {
		return fmt.Errorf("invalid resource URN scheme: '%s' (expected '<SOURCE_URN>/resources/<TYPE>/<path_escaped(URI)>')", urn)
	}

	resourceBase.resourceType = ResourceType(pathParts[6])

	if resourceBase.canonicalURI, err = url.PathUnescape(pathParts[7]); err != nil {
		return fmt.Errorf("malformatted resource URN component: expected path encoded source URI, got '%s'", pathParts[7])
	}

	return
}

func (resourceBase *ResourceBase) MarshalMap() (value map[string]interface{}) {
	value = map[string]interface{}{}

	value["contextId"] = resourceBase.contextID
	value["sourceId"] = resourceBase.sourceID
	value["type"] = resourceBase.resourceType
	value["canonicalUri"] = resourceBase.canonicalURI
	value["urn"] = resourceBase.MarshalURN()

	return
}

func (resourceBase *ResourceBase) MarshalRecord(record Record) {
	record.SetAll(resourceBase.MarshalMap())
}

func (resourceBase *ResourceBase) MarshalProtocol() (resourceProto *protocol.Resource) {
	resourceProto = &protocol.Resource{
		ContextId:    resourceBase.contextID,
		Id:           *resourceBase.id,
		Urn:          resourceBase.MarshalURN(),
		CanonicalUri: resourceBase.canonicalURI,
	}

	if resourceBase.source != nil {
		resourceProto.SourceUrn = resourceBase.source.MarshalURN()
		resourceProto.SourceCanonicalUri = resourceBase.source.CanonicalURI
	}

	switch resourceBase.Type() {
	case ResFSFile:
		resourceProto.Type = protocol.ResourceType_FS_FILE
	case ResWebPage:
		resourceProto.Type = protocol.ResourceType_WEB_PAGE
	default:
		// unknown resource type
	}

	return
}

func (resourceBase *ResourceBase) UnmarshalMap(value map[string]interface{}) error {
	if value == nil {
		return fmt.Errorf("cannot unmarshal nil (type map[string]interface{}) to resource")
	}

	if value["urn"] != nil && value["urn"] != "" {
		if err := resourceBase.unmarshalURN(value["urn"].(string)); err != nil {
			return err
		}
	} else {
		if value["contextId"] != nil && value["contextId"] != "" {
			resourceBase.contextID = value["contextId"].(string)
		} else {
			return fmt.Errorf("missing resource contextId")
		}

		if value["type"] != nil && value["type"] != "" {
			resourceBase.resourceType = ResourceType(value["type"].(string))
		} else {
			return fmt.Errorf("missing resource type")
		}

		if value["canonicalUri"] != nil && value["canonicalUri"] != "" {
			resourceBase.canonicalURI = value["canonicalUri"].(string)
		} else {
			return fmt.Errorf("missing resource canonicalUri")
		}
	}

	if value["sourceId"] != nil && value["sourceId"] != "" {
		resourceBase.sourceID = value["sourceId"].(string)
	}

	return nil
}

func (resourceBase *ResourceBase) UnmarshalDBDocument(document *clover.Document) error {
	if document == nil {
		return fmt.Errorf("cannot unmarshal nil (type *clover.Document) to resource")
	}

	if document.Get("urn") != nil && document.Get("urn") != "" {
		if err := resourceBase.unmarshalURN(document.Get("urn").(string)); err != nil {
			return err
		}
	} else {
		if document.Get("contextId") != nil && document.Get("contextId") != "" {
			resourceBase.contextID = document.Get("contextId").(string)
		} else {
			return fmt.Errorf("missing resource contextId")
		}

		if document.Get("type") != nil && document.Get("type") != "" {
			resourceBase.resourceType = ResourceType(document.Get("type").(string))
		} else {
			return fmt.Errorf("missing resource type")
		}

		if document.Get("canonicalUri") != nil && document.Get("canonicalUri") != "" {
			resourceBase.canonicalURI = document.Get("canonicalUri").(string)
		} else {
			return fmt.Errorf("missing resource canonicalUri")
		}
	}

	if document.Get("sourceId") != nil && document.Get("sourceId") != "" {
		resourceBase.sourceID = document.Get("sourceId").(string)
	}

	return nil
}

func (resourceBase *ResourceBase) Index(adapter Adapter) error {
	return fmt.Errorf("unimplemented")
}

// type ResourceWebPage struct{}
// type ResourceWebTable struct{}

// type ResourceFileImage struct{}
// type ResourceFilePDF struct{}
// type ResourceFileDocx struct{}
// type ResourceFileXlsx struct{}
