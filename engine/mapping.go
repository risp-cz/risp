package engine

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/char/html"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/web"
	"github.com/blevesearch/bleve/v2/mapping"
)

func excludeFieldMapping() *mapping.FieldMapping {
	excludeFieldMapping := bleve.NewKeywordFieldMapping()
	excludeFieldMapping.Store = true
	excludeFieldMapping.Index = false
	excludeFieldMapping.IncludeInAll = false
	excludeFieldMapping.IncludeTermVectors = false
	return excludeFieldMapping
}

func booleanFieldMapping() *mapping.FieldMapping {
	booleanFieldMapping := bleve.NewBooleanFieldMapping()
	return booleanFieldMapping
}

func keywordFieldMapping() *mapping.FieldMapping {
	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	return keywordFieldMapping
}

func textFieldMapping() *mapping.FieldMapping {
	textFieldMapping := bleve.NewTextFieldMapping()
	return textFieldMapping
}

func htmlFieldMapping() *mapping.FieldMapping {
	htmlFieldMapping := bleve.NewTextFieldMapping()
	htmlFieldMapping.Analyzer = "risp-html"
	// htmlFieldMapping.Store = true
	// htmlFieldMapping.Index = true
	// htmlFieldMapping.IncludeTermVectors = true

	return htmlFieldMapping
}

func BuildIndexMapping() (indexMapping *mapping.IndexMappingImpl, err error) {
	indexMapping = bleve.NewIndexMapping()
	indexMapping.TypeField = RecordTypeField

	if err = indexMapping.AddCustomAnalyzer("risp-html", map[string]interface{}{
		"type": custom.Name,
		// "tokenizer": unicode.Name,
		"tokenizer": web.Name,
		"char_filters": []string{
			html.Name,
		},
		// "token_filters": []string{},
	}); err != nil {
		return
	}

	excludeFieldMapping := excludeFieldMapping()
	booleanFieldMapping := booleanFieldMapping()
	keywordFieldMapping := keywordFieldMapping()
	textFieldMapping := textFieldMapping()
	htmlFieldMapping := htmlFieldMapping()

	// Source
	sourceMapping := bleve.NewDocumentMapping()

	sourceMapping.AddFieldMappingsAt(RecordTypeField, keywordFieldMapping)
	sourceMapping.AddFieldMappingsAt("contextId", keywordFieldMapping)
	sourceMapping.AddFieldMappingsAt("adapterType", keywordFieldMapping)
	sourceMapping.AddFieldMappingsAt("canonicalUri", keywordFieldMapping)
	sourceMapping.AddFieldMappingsAt("urn", excludeFieldMapping)

	adapterDataMapping := bleve.NewDocumentMapping()
	// Source [FS]
	adapterDataMapping.AddFieldMappingsAt("path", keywordFieldMapping)
	adapterDataMapping.AddFieldMappingsAt("isDir", booleanFieldMapping)
	adapterDataMapping.AddFieldMappingsAt("isDot", booleanFieldMapping)
	// Source [Web]
	adapterDataMapping.AddFieldMappingsAt("scheme", keywordFieldMapping)
	adapterDataMapping.AddFieldMappingsAt("host", keywordFieldMapping)
	adapterDataMapping.AddFieldMappingsAt("user", excludeFieldMapping)

	sourceMapping.AddSubDocumentMapping("adapterData", adapterDataMapping)

	// Resource
	resourceMapping := bleve.NewDocumentMapping()

	resourceMapping.AddFieldMappingsAt(RecordTypeField, keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt("contextId", keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt("sourceId", keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt("type", keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt("canonicalUri", keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt("urn", excludeFieldMapping)

	// Resource [FSFile]
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.path", ResFSFile), keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.filename", ResFSFile), keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.filetype", ResFSFile), keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.isDot", ResFSFile), booleanFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.contents_keywords", ResFSFile), keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.contents_text", ResFSFile), textFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.contents_html", ResFSFile), htmlFieldMapping)

	// Resource [WebPage]
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.path", ResWebPage), keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.query", ResWebPage), keywordFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.title", ResWebPage), textFieldMapping)
	resourceMapping.AddFieldMappingsAt(fmt.Sprintf("%s.body", ResWebPage), htmlFieldMapping)

	indexMapping.AddDocumentMapping(string(RecordSource), sourceMapping)
	indexMapping.AddDocumentMapping(string(RecordResource), resourceMapping)

	return
}
