package engine

import (
	_context "context"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/necessitates/clover"
	"google.golang.org/grpc"

	"risp/config"
	"risp/dump"
	"risp/protocol"
)

const (
	ColContexts  string = "Contexts"
	ColSources   string = "Sources"
	ColResources string = "Resources"
)

var collections = []string{ColContexts, ColSources, ColResources}

type Engine struct {
	protocol.UnimplementedRispServer
	config   *config.Config
	database *clover.DB
	contexts map[string]*Context
	// stopSignal chan bool
}

func NewEngine(config *config.Config) *Engine {
	return &Engine{
		config:   config,
		contexts: map[string]*Context{},
		// stopSignal: make(chan bool),
	}
}

func (engine *Engine) Start() (err error) {
	fmt.Printf("Starting Risp engine...\n")

	if err = engine.initializeDatabase(); err != nil {
		return
	}

	if err = engine.loadContexts(); err != nil {
		return
	}

	if err = engine.setupDefaultContext(); err != nil {
		return
	}

	if err = engine.listen(); err != nil {
		return
	}

	// <-engine.stopSignal
	return
}

func (engine *Engine) Stop() (err error) {
	fmt.Printf("Engine stopping\n")

	// engine.stopSignal <- true
	return
}

func (engine *Engine) Execute(context _context.Context, request *protocol.ExecuteRequest) (response *protocol.ExecuteResponse, err error) {
	response = &protocol.ExecuteResponse{
		Error: NewProtocolError(),
	}

	fmt.Printf("Execute called\n")

	commands := strings.Split(request.Command, " ")
	if len(commands) < 1 {
		response.Error = NewProtocolError(ErrInvalidCommand, "Empty Command")
		return
	}

	fmt.Printf("  request.Command: '%s'\n", request.Command)

	switch commands[0] {
	case "source":
		if len(commands) < 2 || len(commands[1]) < 1 {
			response.Error = NewProtocolError(ErrInvalidSourceURI, "Missing Source URI")
			return
		}

		fmt.Printf("  sourcing '%s'\n", commands[1])

	resolveContext:
		for _, context := range engine.contexts {
			if context.IsDefault {
				// var source Source
				if _, err = context.SourceURI(commands[1]); err != nil {
					return
				}

				break resolveContext
			}
		}

		return
	}

	response.Error = NewProtocolError(ErrInvalidCommand, fmt.Sprintf("Invalid Command: '%s'", commands[0]))
	return
}

func (engine *Engine) Query(context _context.Context, request *protocol.QueryRequest) (response *protocol.QueryResponse, err error) {
	response = &protocol.QueryResponse{
		Error: NewProtocolError(),
	}

	fmt.Printf("Query called\n")

	if len(request.Value) < 1 {
		response.Error = NewProtocolError(ErrInvalidQuery, "Empty Query")
		return
	}

	highlightStyle := "html"
	if request.HighlightStyle != nil {
		switch *request.HighlightStyle {
		case protocol.QueryHighlightStyle_HTML:
			highlightStyle = "html"
		case protocol.QueryHighlightStyle_ASCI:
			highlightStyle = "asci"
		}
	}

	fmt.Printf("  request.Context: %s\n", request.ContextId)
	fmt.Printf("  request.Value: '%s'\n", request.Value)
	fmt.Printf("  request.HighlightStyle: '%s' (actual used: '%s')\n", request.HighlightStyle, highlightStyle)

	for _, context := range engine.contexts {
		if context.IsDefault {
			var searchResult *SearchResult

			if searchResult, err = context.Search(request.Value, highlightStyle); err != nil {
				return
			}

			fmt.Printf("  searchResult: %+v\n", searchResult)

			response = searchResult.MarshalProtocol()
			break
		}
	}

	return
}

func (engine *Engine) IndexURI(context _context.Context, request *protocol.IndexURIRequest) (response *protocol.IndexURIResponse, err error) {
	response = &protocol.IndexURIResponse{
		Error: NewProtocolError(),
	}

	fmt.Printf("Index URI called\n")
	fmt.Printf("  request.Context: %s\n", request.ContextId)
	fmt.Printf("  request.Uri: '%s'\n", request.Uri)

	for _, context := range engine.contexts {
		// request.ContextId
		if context.IsDefault {
			var source *Source

			if source, err = context.SourceURI(request.Uri); err != nil {
				return
			}

			response.Source = source.MarshalProtocol()
		}
	}

	fmt.Printf("  response: '%+v'\n", response)

	return
}

func (engine *Engine) GetContext(context _context.Context, request *protocol.GetContextRequest) (response *protocol.GetContextResponse, err error) {
	response = &protocol.GetContextResponse{
		Error: NewProtocolError(),
	}

	for contextID, context := range engine.contexts {
		if (request.ContextId == nil || *request.ContextId == "") && context.IsDefault {
			response.Context = context.MarshalProtocol()
			break
		}

		if (request.ContextId != nil && *request.ContextId != "") && contextID == *request.ContextId {
			response.Context = context.MarshalProtocol()
			break
		}
	}

	if response.Context == nil {
		response.Error = NewProtocolError(ErrInvalidContext, "Context Not Found")
	}

	return
}

func (engine *Engine) GetContexts(context _context.Context, request *protocol.GetContextsRequest) (response *protocol.GetContextsResponse, err error) {
	response = &protocol.GetContextsResponse{
		Error: NewProtocolError(),
	}

	response.ContextsTotal = int64(len(engine.contexts))
	response.Contexts = make([]*protocol.Context, 0)

	for _, context := range engine.contexts {
		response.Contexts = append(response.Contexts, context.MarshalProtocol())
	}

	return
}

// func (engine *Engine) GetSource(context _context.Context, request *protocol.GetSourceRequest) (response *protocol.GetSourceResponse, err error) {
// 	response = &protocol.GetSourceResponse{
// 		Error: NewProtocolError(),
// 	}

// 	// request.Id

// 	return
// }

func (engine *Engine) GetSources(context _context.Context, request *protocol.GetSourcesRequest) (response *protocol.GetSourcesResponse, err error) {
	var (
		limit  = 100
		offset = 0
	)

	response = &protocol.GetSourcesResponse{
		Error: NewProtocolError(),
	}

	if request.Limit > 0 {
		limit = int(request.Limit)
	}

	if request.Offset >= 0 {
		offset = int(request.Offset)
	}

	response.Sources = make([]*protocol.Source, 0)

	for contextID, context := range engine.contexts {
		if context.IsDefault {
			var (
				documents      []*clover.Document
				documentsTotal int
			)

			query := context.engine.database.Query(ColSources).Where(clover.Field("contextId").Eq(contextID))

			if documentsTotal, err = query.Count(); err != nil {
				return
			}

			response.SourcesTotal = int64(documentsTotal)

			if documents, err = query.Limit(limit).Skip(offset).FindAll(); err != nil {
				return
			}

			for _, document := range documents {
				source := &Source{
					ContextID: contextID,
					ID:        document.ObjectId(),
				}

				if err = source.UnmarshalDBDocument(document); err != nil {
					return
				}

				response.Sources = append(response.Sources, source.MarshalProtocol())
			}
		}
	}

	return
}

func (engine *Engine) GetResources(context _context.Context, request *protocol.GetResourcesRequest) (response *protocol.GetResourcesResponse, err error) {
	var (
		limit  = 100
		offset = 0
	)

	response = &protocol.GetResourcesResponse{
		Error: NewProtocolError(),
	}

	if request.Limit > 0 {
		limit = int(request.Limit)
	}

	if request.Offset >= 0 {
		offset = int(request.Offset)
	}

	response.Resources = make([]*protocol.Resource, 0)

	for contextID, context := range engine.contexts {
		if context.IsDefault {
			var (
				documents      []*clover.Document
				documentsTotal int
			)

			query := context.engine.database.Query(ColResources).Where(clover.Field("contextId").Eq(contextID))

			if documentsTotal, err = query.Count(); err != nil {
				return
			}

			response.ResourcesTotal = int64(documentsTotal)

			if documents, err = query.Limit(limit).Skip(offset).FindAll(); err != nil {
				return
			}

			for _, document := range documents {
				var resource Resource

				if resource, err = UnmarshalResource(document); err != nil {
					return
				}

				resource.SetID(document.ObjectId())

				response.Resources = append(response.Resources, resource.MarshalProtocol())
			}
		}
	}

	return
}

func (engine *Engine) CreateContext(context _context.Context, request *protocol.CreateContextRequest) (response *protocol.CreateContextResponse, err error) {
	var createdContext *Context

	response = &protocol.CreateContextResponse{
		Error: NewProtocolError(),
	}

	if createdContext, err = engine.createContext(&Context{
		Name: request.Name,
	}); err != nil {
		return
	}

	response.Context = createdContext.MarshalProtocol()
	return
}

func (engine *Engine) ExportContexts(context _context.Context, request *protocol.ExportContextsRequest) (response *protocol.ExportContextsResponse, err error) {
	var contextsYAMLData []byte

	response = &protocol.ExportContextsResponse{
		Error: NewProtocolError(),
	}

	data := &dump.DataYAML{
		Contexts: make([]*dump.ContextYAML, 0),
	}

	fmt.Printf("Export contexts called\n")
	fmt.Printf("  request.OutputPath: %+v\n", request.OutputPath)
	fmt.Printf("  request.ContextIds: %+v\n", request.ContextIds)

	for _, contextID := range request.ContextIds {
		if engine.contexts[contextID] == nil {
			continue
		}

		contextYAML := &dump.ContextYAML{
			Name:      engine.contexts[contextID].Name,
			IsDefault: engine.contexts[contextID].IsDefault,
			Sources:   make([]*dump.SourceYAML, 0),
		}

		sourcesOffset := 0
		sourcesBatchSize := 100
		for {
			var sources = make([]*Source, 0)

			if sources, _, err = engine.contexts[contextID].GetSources(sourcesBatchSize, sourcesOffset); err != nil {
				return nil, err
			}

			if len(sources) < 1 {
				break
			}

			for _, source := range sources {
				sourceYAML := &dump.SourceYAML{
					URI:       source.CanonicalURI,
					Resources: make(dump.Resources, 0),
				}

				if true { // if source.AdapterType != AdapterTypeFS {
					var (
						resourcesOffset    = 0
						resourcesBatchSize = 100
					)

				loopResources:
					for {
						var resources = make([]Resource, 0)

						if resources, _, err = engine.contexts[contextID].GetResourcesByCriteria(
							clover.Field("sourceId").Eq(source.ID),
							resourcesBatchSize,
							resourcesOffset,
						); err != nil {
							return nil, err
						}

						if len(resources) < 1 {
							break loopResources
						}

						for _, resource := range resources {
							sourceYAML.Resources = append(sourceYAML.Resources, resource.CanonicalURI())
						}

						resourcesOffset += resourcesBatchSize
					}
				}

				sort.Sort(sourceYAML.Resources)

				contextYAML.Sources = append(contextYAML.Sources, sourceYAML)
			}

			sourcesOffset += sourcesBatchSize
		}

		data.Contexts = append(data.Contexts, contextYAML)
	}

	fmt.Printf("  converting to YAML\n")

	if contextsYAMLData, err = dump.EncodeDataYAML(data); err != nil {
		return nil, err
	}

	if err = os.WriteFile(request.OutputPath, contextsYAMLData, 0666); err != nil {
		return
	}

	fmt.Printf("  done writing to '%s'\n", request.OutputPath)

	return
}

func (engine *Engine) initializeDatabase() (err error) {
	fmt.Printf("Initializing master database\n")
	engine.database, err = clover.Open(fmt.Sprintf("%s/__master", engine.config.PathData))

	for _, name := range collections {
		var hasCollection bool

		if hasCollection, err = engine.database.HasCollection(name); err != nil {
			return
		}

		if !hasCollection {
			if engine.database.CreateCollection(name); err != nil {
				return
			}
		}
	}

	return
}

func (engine *Engine) loadContexts() (err error) {
	var (
		documents []*clover.Document
		query     = engine.database.Query(ColContexts)
	)

	fmt.Printf("Loading existing contexts\n")

	if documents, err = query.FindAll(); err != nil {
		return
	}

	for _, document := range documents {
		contextID := document.ObjectId()

		context := &Context{
			ID:     contextID,
			engine: engine,
		}

		if err = context.UnmarshalDBDocument(document); err != nil {
			return
		}

		if err = context.initializeIndex(); err != nil {
			return
		}

		engine.contexts[contextID] = context
	}

	return
}

func (engine *Engine) setupDefaultContext() (err error) {
	fmt.Printf("Setting up default context\n")

	hasDefaultContext := false
	for _, context := range engine.contexts {
		fmt.Printf("context: %+v\n", context)
		if context.IsDefault {
			hasDefaultContext = true
			break
		}
	}

	if !hasDefaultContext {
		defaultContext, err := engine.createContext(&Context{
			Name:      "_default",
			IsDefault: true,
		})
		if err != nil {
			return err
		}

		if _, err = defaultContext.SourceURI("https://en.wikipedia.org/wiki/Diopeithes"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("file:///Users/patrik/projects/doceo/risp/tmp/hello"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("https://en.wikipedia.org/wiki/Richard_Whatmore"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("https://en.wikipedia.org/wiki/John_D._Wickhem"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("https://en.wikipedia.org/wiki/Asana_(software)"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("https://en.wikipedia.org/wiki/Backpacking_with_animals"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("https://www.urbandictionary.com/define.php?term=foo"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("https://www.techtarget.com/searchapparchitecture/definition/foo-in-software-programming"); err != nil {
			return err
		}
		if _, err = defaultContext.SourceURI("https://en.wikipedia.org/wiki/Ball_transfer_unit"); err != nil {
			return err
		}
	}

	return
}

func (engine *Engine) listen() (err error) {
	server := grpc.NewServer()

	if engine.config.GRPCListener, err = net.Listen(
		"tcp",
		fmt.Sprintf(":%d", engine.config.GRPCPort),
	); err != nil {
		return
	}

	protocol.RegisterRispServer(server, engine)

	fmt.Printf("GRPC server listening on %s\n", engine.config.GRPCListener.Addr().String())
	return server.Serve(engine.config.GRPCListener)
}

func (engine *Engine) createContext(context *Context) (*Context, error) {
	var err error

	if context.engine == nil {
		context.engine = engine
	}

	document := clover.NewDocument()
	document.SetAll(context.MarshalMap())

	if context.ID, err = engine.database.InsertOne(ColContexts, document); err != nil {
		return context, err
	}

	if err = context.initializeIndex(); err != nil {
		return context, err
	}

	engine.contexts[context.ID] = context
	return context, nil
}
