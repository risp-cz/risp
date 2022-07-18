package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/necessitates/clover"
	"golang.org/x/net/html"

	"risp/protocol"
)

const ResWebPage ResourceType = "web-page"

type ResourceWebPage struct {
	*ResourceBase
	Path             string
	Query            string
	Title            string
	body             string
	skipFetchOnIndex bool
}

func NewResourceWebPage(source *Source, resourceURI *url.URL) *ResourceWebPage {
	canonicalURI := &url.URL{}
	if resourceURI != nil {
		*canonicalURI = *resourceURI
	}

	canonicalURI.Scheme = ""
	canonicalURI.Host = ""
	canonicalURI.User = nil

	if !strings.HasPrefix(canonicalURI.Path, "/") {
		canonicalURI.Path = fmt.Sprintf("/%s", canonicalURI.Path)
	}

	return &ResourceWebPage{
		Path:  canonicalURI.Path,
		Query: canonicalURI.RawQuery,
		ResourceBase: &ResourceBase{
			resourceType: ResWebPage,
			source:       source,
			sourceID:     source.ID,
			contextID:    source.ContextID,
			canonicalURI: canonicalURI.String(),
		},
	}
}

func (resourceWebPage *ResourceWebPage) MarshalMap() (value map[string]interface{}) {
	value = resourceWebPage.ResourceBase.MarshalMap()

	value[ResWebPage.String()] = map[string]interface{}{
		"path":  resourceWebPage.Path,
		"query": resourceWebPage.Query,
		"title": resourceWebPage.Title,
	}

	return
}

func (resourceWebPage *ResourceWebPage) MarshalRecord(record Record) {
	resourceWebPage.ResourceBase.MarshalRecord(record)

	record[ResWebPage.String()] = map[string]interface{}{
		"path":  resourceWebPage.Path,
		"query": resourceWebPage.Query,
		"title": resourceWebPage.Title,
		"body":  resourceWebPage.body,
	}
}

func (resourceWebPage *ResourceWebPage) MarshalProtocol() *protocol.Resource {
	resource := resourceWebPage.ResourceBase.MarshalProtocol()

	data, _ := json.Marshal(resourceWebPage.MarshalMap()[ResWebPage.String()])

	resource.DataJson = string(data)

	return resource
}

func (resourceWebPage *ResourceWebPage) UnmarshalMap(value map[string]interface{}) (err error) {
	if err = resourceWebPage.ResourceBase.UnmarshalMap(value); err != nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if value[ResWebPage.String()].(map[string]interface{})[key] != nil {
			*field = value[ResWebPage.String()].(map[string]interface{})[key].(string)
		}
	}

	if value[ResWebPage.String()] != nil {
		unmarshalString(&resourceWebPage.Path, "path")
		unmarshalString(&resourceWebPage.Query, "query")
		unmarshalString(&resourceWebPage.Title, "title")
	}

	return nil
}

func (resourceWebPage *ResourceWebPage) UnmarshalDBDocument(document *clover.Document) (err error) {
	if err = resourceWebPage.ResourceBase.UnmarshalDBDocument(document); err != nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if document.Get(fmt.Sprintf("%s.%s", ResWebPage, key)) != nil {
			*field = document.Get(fmt.Sprintf("%s.%s", ResWebPage, key)).(string)
		}
	}

	unmarshalString(&resourceWebPage.Path, "path")
	unmarshalString(&resourceWebPage.Query, "query")
	unmarshalString(&resourceWebPage.Title, "title")

	return nil
}

func (resourceWebPage *ResourceWebPage) Index(adapter Adapter) (err error) {
	if adapter.Type() != AdapterTypeWeb {
		return fmt.Errorf("invalid adapter '%s': ResourceWebPage expects adapter type '%s'", adapter.Type(), AdapterTypeWeb)
	}

	if resourceWebPage.ID() == nil || *resourceWebPage.ID() == "" {
		return fmt.Errorf("cannot index ResourceWebPage without ID")
	}

	if !resourceWebPage.skipFetchOnIndex {
		var response *http.Response

		if response, err = resourceWebPage.httpGET(adapter); err != nil {
			return
		}

		if err = resourceWebPage.parseHTML(response.Body); err != nil {
			return
		}
	}

	record := make(Record).SetType(RecordResource)

	resourceWebPage.MarshalRecord(record)

	if err = adapter.(*AdapterWeb).index.Index(*resourceWebPage.ID(), record); err != nil {
		return
	}

	return
}

func (resourceWebPage *ResourceWebPage) parseHTML(reader io.Reader) (err error) {
	var (
		webpageNode *html.Node
		buffer      bytes.Buffer
	)

	if webpageNode, err = html.Parse(reader); err != nil {
		return
	}

	if titleNode, hasTitleNode := findHTMLNode(webpageNode, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "title"
	}); hasTitleNode {
		resourceWebPage.Title = titleNode.FirstChild.Data
	}

	if buffer, err = sanitizeHTMLDocument(webpageNode); err != nil {
		return
	}

	resourceWebPage.body = buffer.String()
	resourceWebPage.skipFetchOnIndex = true
	return
}

func (resourceWebPage *ResourceWebPage) httpGET(adapter Adapter) (response *http.Response, err error) {
	var (
		contentType  string
		canonicalURI *url.URL
		resourceURI  *url.URL
	)

	if canonicalURI, err = url.Parse(resourceWebPage.CanonicalURI()); err != nil {
		return
	}

	if resourceURI, err = adapter.(*AdapterWeb).prependSourceURI(canonicalURI); err != nil {
		return
	}

	if response, err = http.Get(resourceURI.String()); err != nil {
		return
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response, fmt.Errorf("failed with code %d; GET %s", response.StatusCode, resourceURI.String())
	}

	if contentType = response.Header.Get("Content-Type"); contentType != "" {
		var mediatype string
		// params    map[string]string

		if mediatype, _, err = mime.ParseMediaType(contentType); err != nil {
			return
		}

		switch mediatype {
		case "text/html", "html":
			return
		default:
			return response, fmt.Errorf("invalid response: unknown Content-Type '%s'", contentType)
		}
	}

	return response, fmt.Errorf("invalid response: missing Content-Type")
}
