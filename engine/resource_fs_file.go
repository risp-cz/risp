package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/necessitates/clover"
	"golang.org/x/net/html"

	"risp/protocol"
)

const ResFSFile ResourceType = "fs-file"

type ResourceFSFile struct {
	*ResourceBase
	Path              string
	Filename          string
	Filetype          string
	IsDot             bool
	contents_keywords string
	contents_text     string
	contents_html     string
	skipReadOnIndex   bool
}

func NewResourceFSFile(source *Source, resourcePath string) *ResourceFSFile {
	filename := path.Base(resourcePath)
	filenameParts := strings.Split(filename, ".")
	filetype := ""
	isDot := false

	if strings.HasPrefix(filename, ".") {
		isDot = true
	} else if len(filenameParts) > 0 {
		filetype = filenameParts[len(filenameParts)-1]
	}

	return &ResourceFSFile{
		Path:     resourcePath,
		Filename: filename,
		Filetype: filetype,
		IsDot:    isDot,
		ResourceBase: &ResourceBase{
			resourceType: ResFSFile,
			source:       source,
			sourceID:     source.ID,
			contextID:    source.ContextID,
			canonicalURI: resourcePath,
		},
	}
}

func (resourceFSFile *ResourceFSFile) MarshalMap() (value map[string]interface{}) {
	value = resourceFSFile.ResourceBase.MarshalMap()

	value[ResFSFile.String()] = map[string]interface{}{
		"path":     resourceFSFile.Path,
		"filename": resourceFSFile.Filename,
		"filetype": resourceFSFile.Filetype,
		"isDot":    resourceFSFile.IsDot,
	}

	return
}

func (resourceFSFile *ResourceFSFile) MarshalRecord(record Record) {
	resourceFSFile.ResourceBase.MarshalRecord(record)

	record[ResFSFile.String()] = map[string]interface{}{
		"path":              resourceFSFile.Path,
		"filename":          resourceFSFile.Filename,
		"filetype":          resourceFSFile.Filetype,
		"isDot":             resourceFSFile.IsDot,
		"contents_keywords": resourceFSFile.contents_keywords,
		"contents_text":     resourceFSFile.contents_text,
		"contents_html":     resourceFSFile.contents_html,
	}
}

func (resourceFSFile *ResourceFSFile) MarshalProtocol() *protocol.Resource {
	resource := resourceFSFile.ResourceBase.MarshalProtocol()

	data, _ := json.Marshal(resourceFSFile.MarshalMap()[ResFSFile.String()])

	resource.DataJson = string(data)

	return resource
}

func (resourceFSFile *ResourceFSFile) UnmarshalMap(value map[string]interface{}) (err error) {
	if err = resourceFSFile.ResourceBase.UnmarshalMap(value); err != nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if value[ResFSFile.String()].(map[string]interface{})[key] != nil {
			*field = value[ResFSFile.String()].(map[string]interface{})[key].(string)
		}
	}

	unmarshalBool := func(field *bool, key string) {
		if value[ResFSFile.String()].(map[string]interface{})[key] != nil {
			*field = value[ResFSFile.String()].(map[string]interface{})[key].(bool)
		}
	}

	if value[ResFSFile.String()] != nil {
		unmarshalString(&resourceFSFile.Path, "path")
		unmarshalString(&resourceFSFile.Filename, "filename")
		unmarshalString(&resourceFSFile.Filetype, "filetype")
		unmarshalBool(&resourceFSFile.IsDot, "isDot")
	}

	return nil
}

func (resourceFSFile *ResourceFSFile) UnmarshalDBDocument(document *clover.Document) (err error) {
	if err = resourceFSFile.ResourceBase.UnmarshalDBDocument(document); err != nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if document.Get(fmt.Sprintf("%s.%s", ResFSFile, key)) != nil {
			*field = document.Get(fmt.Sprintf("%s.%s", ResFSFile, key)).(string)
		}
	}

	unmarshalBool := func(field *bool, key string) {
		if document.Get(fmt.Sprintf("%s.%s", ResFSFile, key)) != nil {
			*field = document.Get(fmt.Sprintf("%s.%s", ResFSFile, key)).(bool)
		}
	}

	unmarshalString(&resourceFSFile.Path, "path")
	unmarshalString(&resourceFSFile.Filename, "filename")
	unmarshalString(&resourceFSFile.Filetype, "filetype")
	unmarshalBool(&resourceFSFile.IsDot, "isDot")

	return nil
}

func (resourceFSFile *ResourceFSFile) Index(adapter Adapter) (err error) {
	if adapter.Type() != AdapterTypeFS {
		return fmt.Errorf("invalid adapter '%s': ResourceWebPage expects adapter type '%s'", adapter.Type(), AdapterTypeWeb)
	}

	if resourceFSFile.ID() == nil || *resourceFSFile.ID() == "" {
		return fmt.Errorf("cannot index ResourceFSFile without ID")
	}

	if !resourceFSFile.skipReadOnIndex {
		var data []byte

		if data, err = resourceFSFile.readFile(adapter); err != nil {
			return
		}

		if err = resourceFSFile.parseFile(adapter, data); err != nil {
			return
		}
	}

	record := make(Record).SetType(RecordResource)

	resourceFSFile.MarshalRecord(record)

	if err = adapter.(*AdapterFS).index.Index(*resourceFSFile.ID(), record); err != nil {
		return
	}

	return
}

func (resourceFSFile *ResourceFSFile) parseFile(adapter Adapter, data []byte) (err error) {
	switch resourceFSFile.Filetype {
	case "txt":
		resourceFSFile.contents_text = string(data)
	case "html":
		var (
			webpageNode *html.Node
			buffer      bytes.Buffer
		)

		if webpageNode, err = html.Parse(bytes.NewBuffer(data)); err != nil {
			return
		}

		if buffer, err = sanitizeHTMLDocument(webpageNode); err != nil {
			return
		}

		resourceFSFile.contents_html = buffer.String()
	}

	resourceFSFile.skipReadOnIndex = true
	return
}

func (resourceFSFile *ResourceFSFile) readFile(adapter Adapter) (data []byte, err error) {
	resourceURI, err := adapter.(*AdapterFS).prependBasePath(resourceFSFile.Path)
	if err != nil {
		return data, err
	}

	return os.ReadFile(resourceURI.Path)
}
