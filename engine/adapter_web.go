package engine

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/necessitates/clover"

	"risp/protocol"
)

type AdapterDataWeb struct {
	Scheme string
	Host   string
	User   string
}

func (adapterDataWeb *AdapterDataWeb) MarshalMap() map[string]interface{} {
	return map[string]interface{}{
		"scheme": adapterDataWeb.Scheme,
		"host":   adapterDataWeb.Host,
		"user":   adapterDataWeb.User,
	}
}

func (adapterDataWeb *AdapterDataWeb) MarshalProtocol(source *protocol.Source) {
	if source == nil {
		return
	}

	source.AdapterData = &protocol.Source_Web{
		Web: &protocol.AdapterDataWeb{
			Scheme: adapterDataWeb.Scheme,
			Host:   adapterDataWeb.Host,
			User:   adapterDataWeb.User,
		},
	}
}

func (adapterDataWeb *AdapterDataWeb) UnmarshalMap(value map[string]interface{}) (err error) {
	if value["adapterData"] == nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if value["adapterData"].(map[string]interface{})[key] != nil {
			*field = value["adapterData"].(map[string]interface{})[key].(string)
		}
	}

	unmarshalString(&adapterDataWeb.Scheme, "scheme")
	unmarshalString(&adapterDataWeb.Host, "host")
	unmarshalString(&adapterDataWeb.User, "user")
	return
}

func (adapterDataWeb *AdapterDataWeb) UnmarshalDBDocument(document *clover.Document) (err error) {
	if document == nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if document.Get(fmt.Sprintf("adapterData.%s", key)) != nil {
			*field = document.Get(fmt.Sprintf("adapterData.%s", key)).(string)
		}
	}

	unmarshalString(&adapterDataWeb.Scheme, "scheme")
	unmarshalString(&adapterDataWeb.Host, "host")
	unmarshalString(&adapterDataWeb.User, "user")
	return
}

type AdapterWeb struct {
	Adapter
	source   *Source
	database *clover.DB
	index    bleve.Index
}

func NewAdapterWeb(source *Source, database *clover.DB, index bleve.Index) *AdapterWeb {
	return &AdapterWeb{
		source:   source,
		database: database,
		index:    index,
	}
}

func (adapterWeb *AdapterWeb) Type() AdapterType {
	return AdapterTypeWeb
}

func (adapterWeb *AdapterWeb) UnmarshalMap(value map[string]interface{}) error {
	if adapterWeb.source.AdapterData == nil {
		adapterWeb.source.AdapterData = &AdapterDataWeb{}
	}

	return adapterWeb.source.AdapterData.UnmarshalMap(value)
}

func (adapterWeb *AdapterWeb) UnmarshalDBDocument(document *clover.Document) error {
	if adapterWeb.source.AdapterData == nil {
		adapterWeb.source.AdapterData = &AdapterDataWeb{}
	}

	return adapterWeb.source.AdapterData.UnmarshalDBDocument(document)
}

func (adapterWeb *AdapterWeb) Index() (err error) {
	var (
		parsedURI *url.URL
		documents []*clover.Document
	)

	if parsedURI, err = url.Parse(adapterWeb.source.CanonicalURI); err != nil {
		return fmt.Errorf("invalid URI '%s': %+v", adapterWeb.source.CanonicalURI, err)
	}

	if parsedURI.Scheme != "http" && parsedURI.Scheme != "https" {
		return fmt.Errorf("invalid URI scheme '%s', expected 'http(s)'", parsedURI.Scheme)
	}

	canonicalURI := &url.URL{
		Scheme: parsedURI.Scheme,
		Host:   parsedURI.Host,
		User:   parsedURI.User,
	}

	adapterWeb.source.CanonicalURI = canonicalURI.String()

	if documents, err = adapterWeb.database.Query(ColSources).Where(
		clover.Field("urn").Eq(adapterWeb.source.MarshalURN()),
	).FindAll(); err != nil {
		return
	}

	if len(documents) > 0 {
		adapterWeb.source.ID = documents[0].ObjectId()

		if err = adapterWeb.source.UnmarshalDBDocument(documents[0]); err != nil {
			return
		}

		err = adapterWeb.crawlURI(parsedURI)
		return
	}

	adapterWeb.source.AdapterData = &AdapterDataWeb{
		Scheme: canonicalURI.Scheme,
		Host:   canonicalURI.Host,
		User:   canonicalURI.User.String(),
	}

	document := clover.NewDocument()
	document.SetAll(adapterWeb.source.MarshalMap())

	if adapterWeb.source.ID, err = adapterWeb.database.InsertOne(ColSources, document); err != nil {
		return
	}

	record := make(Record).
		SetType(RecordSource).
		SetAll(adapterWeb.source.MarshalMap())

	if err = adapterWeb.index.Index(adapterWeb.source.ID, record); err != nil {
		return
	}

	err = adapterWeb.crawlURI(parsedURI)
	return
}

func (adapterWeb *AdapterWeb) crawlURI(resourceURI *url.URL) (err error) {
	if resourceURI, err = adapterWeb.prependSourceURI(resourceURI); err != nil {
		return
	}

	var (
		response    *http.Response
		contentType string
	)

	if response, err = http.Get(resourceURI.String()); err != nil {
		return
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("failed with code %d; GET %s", response.StatusCode, resourceURI.String())
	}

	if contentType = response.Header.Get("Content-Type"); contentType != "" {
		var mediatype string
		// params    map[string]string

		if mediatype, _, err = mime.ParseMediaType(contentType); err != nil {
			return
		}

		switch mediatype {
		case "text/html", "html":
			err = adapterWeb.processResponseHTML(resourceURI, response)
		default:
			fmt.Printf("Unknown response Content-Type '%s', skipping...\n", contentType)
		}
	}

	return
}

func (adapterWeb *AdapterWeb) processResponseHTML(resourceURI *url.URL, response *http.Response) (err error) {
	var documents []*clover.Document

	resourceWebPage := NewResourceWebPage(adapterWeb.source, resourceURI)

	if documents, err = adapterWeb.database.Query(ColResources).Where(
		clover.Field("urn").Eq(resourceWebPage.MarshalURN()),
	).FindAll(); err != nil {
		return
	}

	if len(documents) > 0 {
		resourceWebPage.SetID(documents[0].ObjectId())

		if err = resourceWebPage.UnmarshalDBDocument(documents[0]); err != nil {
			return
		}
	} else {
		var resourceWebPageID string

		document := clover.NewDocument()
		document.SetAll(resourceWebPage.MarshalMap())

		if resourceWebPageID, err = adapterWeb.database.InsertOne(ColResources, document); err != nil {
			return
		}

		resourceWebPage.SetID(resourceWebPageID)
	}

	if err = resourceWebPage.parseHTML(response.Body); err != nil {
		return
	}

	err = resourceWebPage.Index(adapterWeb)
	return
}

func (adapterWeb *AdapterWeb) prependSourceURI(uri *url.URL) (resourceURI *url.URL, err error) {
	resourceURI = &url.URL{}
	if uri != nil {
		*resourceURI = *uri
	}

	adapterDataWeb, isAdapterDataWeb := adapterWeb.source.AdapterData.(*AdapterDataWeb)
	if !isAdapterDataWeb {
		return resourceURI, fmt.Errorf("invalid adapter data: expected type '*AdapterDataWeb', got type '%T'", adapterWeb.source.AdapterData)
	}

	resourceURI.Scheme = adapterDataWeb.Scheme
	resourceURI.Host = adapterDataWeb.Host
	resourceURI.User = nil

	if adapterDataWeb.User != "" {
		credentials := strings.Split(adapterDataWeb.User, ":")

		if len(credentials) > 1 {
			resourceURI.User = url.UserPassword(credentials[0], credentials[1])
		} else {
			resourceURI.User = url.User(credentials[0])
		}
	}

	// region Imaginary code here
	//
	// Imagine a very complex, very interesting, very.. sophisticated! algorithm here that
	// normalizes a URI query to prevent duplication erros due to commutative nature of URI queries
	// ...
	// ...
	// ...
	// Now, stop that! no need, since the issue has been, partially, accidentally! solved under the hood,
	// since *URI.Query is by default sorted alphabetically, to mask for the implementation deficiency of
	// loss of queries' order since the use of map[string]string
	//
	// end region

	return resourceURI, nil
}
