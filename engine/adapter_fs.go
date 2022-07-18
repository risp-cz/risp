package engine

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	Path "path"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/necessitates/clover"

	"risp/protocol"
)

type AdapterDataFS struct {
	Path  string
	IsDir bool
	IsDot bool
}

func (adapterDataFS *AdapterDataFS) MarshalMap() map[string]interface{} {
	return map[string]interface{}{
		"path":  adapterDataFS.Path,
		"isDir": adapterDataFS.IsDir,
		"isDot": adapterDataFS.IsDot,
	}
}

func (adapterDataFS *AdapterDataFS) MarshalProtocol(source *protocol.Source) {
	if source == nil {
		return
	}

	source.AdapterData = &protocol.Source_Fs{
		Fs: &protocol.AdapterDataFS{
			Path:  adapterDataFS.Path,
			IsDir: adapterDataFS.IsDir,
			IsDot: adapterDataFS.IsDot,
		},
	}
}

func (adapterDataFS *AdapterDataFS) UnmarshalMap(value map[string]interface{}) (err error) {
	if value["adapterData"] == nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if value["adapterData"].(map[string]interface{})[key] != nil {
			*field = value["adapterData"].(map[string]interface{})[key].(string)
		}
	}

	unmarshalBool := func(field *bool, key string) {
		if value["adapterData"].(map[string]interface{})[key] != nil {
			*field = value["adapterData"].(map[string]interface{})[key].(bool)
		}
	}

	unmarshalString(&adapterDataFS.Path, "path")
	unmarshalBool(&adapterDataFS.IsDir, "isDir")
	unmarshalBool(&adapterDataFS.IsDot, "isDot")
	return
}

func (adapterDataFS *AdapterDataFS) UnmarshalDBDocument(document *clover.Document) (err error) {
	if document == nil {
		return
	}

	unmarshalString := func(field *string, key string) {
		if document.Get(fmt.Sprintf("adapterData.%s", key)) != nil {
			*field = document.Get(fmt.Sprintf("adapterData.%s", key)).(string)
		}
	}

	unmarshalBool := func(field *bool, key string) {
		if document.Get(fmt.Sprintf("adapterData.%s", key)) != nil {
			*field = document.Get(fmt.Sprintf("adapterData.%s", key)).(bool)
		}
	}

	unmarshalString(&adapterDataFS.Path, "path")
	unmarshalBool(&adapterDataFS.IsDir, "isDir")
	unmarshalBool(&adapterDataFS.IsDot, "isDot")
	return
}

type AdapterFS struct {
	Adapter
	source   *Source
	database *clover.DB
	index    bleve.Index
}

func NewAdapterFS(source *Source, database *clover.DB, index bleve.Index) *AdapterFS {
	return &AdapterFS{
		source:   source,
		database: database,
		index:    index,
	}
}

func (adapterFS *AdapterFS) Type() AdapterType {
	return AdapterTypeFS
}

func (adapterFS *AdapterFS) UnmarshalMap(value map[string]interface{}) error {
	if adapterFS.source.AdapterData == nil {
		adapterFS.source.AdapterData = &AdapterDataFS{}
	}

	return adapterFS.source.AdapterData.UnmarshalMap(value)
}

func (adapterFS *AdapterFS) UnmarshalDBDocument(document *clover.Document) error {
	if adapterFS.source.AdapterData == nil {
		adapterFS.source.AdapterData = &AdapterDataFS{}
	}

	return adapterFS.source.AdapterData.UnmarshalDBDocument(document)
}

func (adapterFS *AdapterFS) Index() (err error) {
	var (
		parsedURI *url.URL
		documents []*clover.Document
		pathStat  fs.FileInfo
	)

	if parsedURI, err = url.Parse(adapterFS.source.CanonicalURI); err != nil {
		return fmt.Errorf("invalid URI '%s': %+v", adapterFS.source.CanonicalURI, err)
	}

	if parsedURI.Scheme != "file" {
		return fmt.Errorf("invalid URI scheme '%s', expected 'file'", parsedURI.Scheme)
	}

	canonicalURI := &url.URL{
		Scheme: parsedURI.Scheme,
		Host:   parsedURI.Host,
		User:   parsedURI.User,
		Path:   parsedURI.Path,
	}

	adapterFS.source.CanonicalURI = canonicalURI.String()

	if documents, err = adapterFS.database.Query(ColSources).Where(
		clover.Field("urn").Eq(adapterFS.source.MarshalURN()),
	).FindAll(); err != nil {
		return
	}

	if len(documents) > 0 {
		adapterFS.source.ID = documents[0].ObjectId()

		if err = adapterFS.source.UnmarshalDBDocument(documents[0]); err != nil {
			return
		}

		err = adapterFS.crawlPath(".")
		return
	}

	if pathStat, err = os.Stat(canonicalURI.Path); err != nil {
		if os.IsNotExist(err) {
			return
		}

		return
	}

	adapterFS.source.AdapterData = &AdapterDataFS{
		Path:  canonicalURI.Path,
		IsDir: pathStat.IsDir(),
		IsDot: strings.HasPrefix(Path.Base(canonicalURI.Path), "."),
	}

	document := clover.NewDocument()
	document.SetAll(adapterFS.source.MarshalMap())

	if adapterFS.source.ID, err = adapterFS.database.InsertOne(ColSources, document); err != nil {
		return
	}

	record := make(Record).
		SetType(RecordSource).
		SetAll(adapterFS.source.MarshalMap())

	if err = adapterFS.index.Index(adapterFS.source.ID, record); err != nil {
		return
	}

	err = adapterFS.crawlPath(".")
	return
}

func (adapterFS *AdapterFS) crawlPath(path string) (err error) {
	var (
		adapterDataFS = adapterFS.source.AdapterData.(*AdapterDataFS)
		resourceURI   *url.URL
		resourceStat  os.FileInfo
	)

	if path == "." || path == "/" || path == "" {
		path = "."
		resourceURI, err = url.Parse(adapterFS.source.CanonicalURI)
	} else if adapterDataFS.IsDir {
		resourceURI, err = adapterFS.prependBasePath(path)
	} else {
		err = fmt.Errorf("cannot crawl subPath '%s' of file source '%s'", path, adapterFS.source.CanonicalURI)
	}

	if err != nil {
		return
	}

	if resourceStat, err = os.Stat(resourceURI.Path); err != nil {
		return
	}

	if resourceStat.IsDir() {
		var entries []os.DirEntry

		if entries, err = os.ReadDir(resourceURI.Path); err != nil {
			return
		}

		for _, entry := range entries {
			if err = adapterFS.crawlPath(
				Path.Join(path, entry.Name()),
			); err != nil {
				return
			}
		}

		return
	}

	var (
		documents []*clover.Document
		data      []byte
	)

	resourceFSFile := NewResourceFSFile(adapterFS.source, path)

	if documents, err = adapterFS.database.Query(ColResources).Where(
		clover.Field("urn").Eq(resourceFSFile.MarshalURN()),
	).FindAll(); err != nil {
		return
	}

	if len(documents) > 0 {
		resourceFSFile.SetID(documents[0].ObjectId())

		if err = resourceFSFile.UnmarshalDBDocument(documents[0]); err != nil {
			return
		}
	} else {
		var resourceFSFileID string

		document := clover.NewDocument()
		document.SetAll(resourceFSFile.MarshalMap())

		if resourceFSFileID, err = adapterFS.database.InsertOne(ColResources, document); err != nil {
			return
		}

		resourceFSFile.SetID(resourceFSFileID)
	}

	if data, err = resourceFSFile.readFile(adapterFS); err != nil {
		return
	}

	if err = resourceFSFile.parseFile(adapterFS, data); err != nil {
		return
	}

	err = resourceFSFile.Index(adapterFS)
	return
}

func (adapterFS *AdapterFS) prependBasePath(resourcePath string) (resourceURI *url.URL, err error) {
	if resourceURI, err = url.Parse(adapterFS.source.CanonicalURI); err != nil {
		return
	}

	resourceURI.Path = Path.Join(resourceURI.Path, resourcePath)
	return
}
