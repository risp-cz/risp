package engine

import (
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/necessitates/clover"

	"risp/protocol"
)

type SearchResultHit struct {
	Score      float64
	Resource   Resource
	Highlights map[string][]string
}

type SearchResult struct {
	MaxScore float64
	Hits     []*SearchResultHit
}

func (searchResult *SearchResult) MarshalProtocol() (response *protocol.QueryResponse) {
	response = &protocol.QueryResponse{
		Error:      &protocol.Error{},
		Edges:      make([]*protocol.QueryHit, 0),
		EdgesTotal: 0,
	}

	response.MaxScore = float32(searchResult.MaxScore)
	response.EdgesTotal = int64(len(searchResult.Hits))

	for _, hit := range searchResult.Hits {
		highlights := make([]*protocol.QueryHighlight, 0)
		for key, values := range hit.Highlights {
			highlights = append(highlights, &protocol.QueryHighlight{
				Key:    key,
				Values: values,
			})
		}

		response.Edges = append(response.Edges, &protocol.QueryHit{
			Score:      float32(hit.Score),
			Resource:   hit.Resource.MarshalProtocol(),
			Highlights: highlights,
		})
	}

	return
}

type Context struct {
	ID        string
	Name      string
	IsDefault bool
	engine    *Engine
	index     bleve.Index
}

func (context *Context) MarshalMap() (value map[string]interface{}) {
	return map[string]interface{}{
		"name":      context.Name,
		"isDefault": context.IsDefault,
	}
}

func (context *Context) UnmarshalMap(value map[string]interface{}) error {
	if value == nil {
		return fmt.Errorf("cannot unmarshal nil map to context")
	}

	context.Name = value["name"].(string)
	context.IsDefault = value["isDefault"].(bool)

	return nil
}

func (context *Context) MarshalProtocol() *protocol.Context {
	return &protocol.Context{
		Id:        context.ID,
		Name:      context.Name,
		IsDefault: context.IsDefault,
	}
}

func (context *Context) UnmarshalDBDocument(document *clover.Document) error {
	if document == nil {
		return fmt.Errorf("cannot unmarshal nil document to context")
	}

	context.Name = document.Get("name").(string)
	context.IsDefault = document.Get("isDefault").(bool)

	return nil
}

func (context *Context) GetIndexPath() string {
	return fmt.Sprintf("%s/%s", context.engine.config.PathData, context.ID)
}

func (context *Context) SourceURI(uri string) (source *Source, err error) {
	fmt.Printf("Source URI '%s'\n", uri)

	source = &Source{
		ContextID:    context.ID,
		CanonicalURI: uri,
	}

	switch true {
	case strings.HasPrefix(uri, "file:"):
		source.AdapterType = AdapterTypeFS
	case strings.HasPrefix(uri, "http:") || strings.HasPrefix(uri, "https:"):
		source.AdapterType = AdapterTypeWeb
	}

	err = source.Adapter(context.engine.database, context.index).Index()
	return
}

func (context *Context) GetSource(sourceID string) (source *Source, err error) {
	var (
		document         *clover.Document
		sourcesColletion = context.engine.database.Query(ColSources)
	)

	if document, err = sourcesColletion.FindById(sourceID); err != nil {
		return
	}

	source = &Source{
		ContextID: context.ID,
		ID:        sourceID,
	}

	err = source.UnmarshalDBDocument(document)
	return
}

func (context *Context) GetSourcesByCriteria(criteria *clover.Criteria, limit, offset int) (sources []*Source, total int, err error) {
	var (
		documents  []*clover.Document
		whereQuery = context.engine.database.Query(ColSources)
	)

	_criteria := clover.Field("contextId").Eq(context.ID)

	if criteria != nil {
		_criteria = _criteria.And(criteria)
	}

	whereQuery = whereQuery.Where(_criteria)

	if total, err = whereQuery.Count(); err != nil {
		return
	}

	query := whereQuery.Limit(limit).Skip(offset)

	if documents, err = query.FindAll(); err != nil {
		return
	}

	sources = make([]*Source, 0)

	for _, document := range documents {
		source := &Source{
			ContextID: context.ID,
			ID:        document.ObjectId(),
		}

		if err = source.UnmarshalDBDocument(document); err != nil {
			return nil, 0, err
		}

		sources = append(sources, source)
	}

	return
}

func (context *Context) GetSources(limit, offset int) (sources []*Source, total int, err error) {
	return context.GetSourcesByCriteria(nil, limit, offset)
}

func (context *Context) GetResource(resourceID string) (resource Resource, err error) {
	var (
		document           *clover.Document
		resourcesColletion = context.engine.database.Query(ColResources)
	)

	if document, err = resourcesColletion.FindById(resourceID); err != nil {
		return
	}

	if resource, err = UnmarshalResource(document); err != nil {
		return
	}

	resource.SetID(document.ObjectId())

	return
}

func (context *Context) GetResourcesByCriteria(criteria *clover.Criteria, limit, offset int) (resources []Resource, total int, err error) {
	var (
		documents  []*clover.Document
		whereQuery = context.engine.database.Query(ColResources)
	)

	_criteria := clover.Field("contextId").Eq(context.ID)

	if criteria != nil {
		_criteria = _criteria.And(criteria)
	}

	whereQuery = whereQuery.Where(_criteria)

	if total, err = whereQuery.Count(); err != nil {
		return
	}

	query := whereQuery.Limit(limit).Skip(offset)

	if documents, err = query.FindAll(); err != nil {
		return
	}

	resources = make([]Resource, 0)

	for _, document := range documents {
		var resource Resource

		if resource, err = UnmarshalResource(document); err != nil {
			return nil, 0, err
		}

		resources = append(resources, resource)
	}

	return
}

func (context *Context) GetResources(limit, offset int) (resources []Resource, total int, err error) {
	return context.GetResourcesByCriteria(nil, limit, offset)
}

func (context *Context) Search(queryString string, highlightStyle string) (result *SearchResult, err error) {
	var (
		searchRequest *bleve.SearchRequest
		searchResult  *bleve.SearchResult
	)

	queryString = fmt.Sprintf("%s +%s:%s", queryString, RecordTypeField, RecordResource)

	query := bleve.NewQueryStringQuery(queryString)

	searchRequest = bleve.NewSearchRequest(query)

	if highlightStyle != "" {
		searchRequest.Highlight = bleve.NewHighlightWithStyle(highlightStyle)

		searchRequest.Highlight.AddField(fmt.Sprintf("%s.contents_text", ResFSFile))
		searchRequest.Highlight.AddField(fmt.Sprintf("%s.contents_html", ResFSFile))

		searchRequest.Highlight.AddField(fmt.Sprintf("%s.title", ResWebPage))
		searchRequest.Highlight.AddField(fmt.Sprintf("%s.body", ResWebPage))
	}

	if searchResult, err = context.index.Search(searchRequest); err != nil {
		return
	}

	result = &SearchResult{
		MaxScore: searchResult.MaxScore,
		Hits:     make([]*SearchResultHit, 0),
	}

	for _, hit := range searchResult.Hits {
		searchResultHit := &SearchResultHit{
			Score:      hit.Score,
			Highlights: hit.Fragments,
		}

		if searchResultHit.Resource, err = context.GetResource(hit.ID); err != nil {
			return
		}

		result.Hits = append(result.Hits, searchResultHit)
	}

	return
}

func (context *Context) initializeIndex() (err error) {
	context.index, err = bleve.Open(context.GetIndexPath())

	if err == bleve.ErrorIndexPathDoesNotExist {
		err = nil

		indexMapping, err := BuildIndexMapping()
		if err != nil {
			return err
		}

		if context.index, err = bleve.New(context.GetIndexPath(), indexMapping); err != nil {
			return err
		}
	}

	return
}
