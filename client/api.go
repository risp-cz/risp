package client

import (
	_context "context"
	"fmt"

	"github.com/skratchdot/open-golang/open"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"risp/config"
	"risp/engine"
	"risp/protocol"
)

type API struct {
	Runtime ClientRuntime
	Config  *config.Config
	Client  protocol.RispClient
}

func NewAPI(runtime ClientRuntime, config *config.Config, client protocol.RispClient) *API {
	return &API{
		Runtime: runtime,
		Config:  config,
		Client:  client,
	}
}

func (api *API) OpenURI(uri string) *error {
	if err := open.Run(uri); err != nil {
		return &err
	}

	return nil
}

func (api *API) Execute(command string) (response *protocol.ExecuteResponse) {
	var err error

	request := &protocol.ExecuteRequest{
		Command: command,
	}

	if response, err = api.Client.Execute(_context.TODO(), request); err != nil {
		response = &protocol.ExecuteResponse{
			Error: engine.NewProtocolError(engine.ErrUnknown, err),
		}
	}

	return
}

func (api *API) Query(value string) (response *protocol.QueryResponse) {
	var err error

	request := &protocol.QueryRequest{
		Value:     value,
		ContextId: api.Config.ReplContextID,
	}

	if response, err = api.Client.Query(_context.TODO(), request); err != nil {
		response = &protocol.QueryResponse{
			Error: engine.NewProtocolError(engine.ErrUnknown, err),
		}
	}

	return
}

func (api *API) IndexURI(uri string) (response *protocol.IndexURIResponse) {
	var err error

	request := &protocol.IndexURIRequest{
		ContextId: api.Config.ReplContextID,
		Uri:       uri,
	}

	if response, err = api.Client.IndexURI(_context.TODO(), request); err != nil {
		response = &protocol.IndexURIResponse{
			Error: engine.NewProtocolError(engine.ErrUnknown, err),
		}
	}

	return
}

func (api *API) GetContexts() (response *protocol.GetContextsResponse) {
	var err error

	request := &protocol.GetContextsRequest{}

	if response, err = api.Client.GetContexts(_context.TODO(), request); err != nil {
		response = &protocol.GetContextsResponse{
			Error: engine.NewProtocolError(engine.ErrUnknown, err),
		}
	}

	return
}

func (api *API) GetSources() (response *protocol.GetSourcesResponse) {
	var err error

	request := &protocol.GetSourcesRequest{
		ContextId: api.Config.ReplContextID,
	}

	if response, err = api.Client.GetSources(_context.TODO(), request); err != nil {
		response = &protocol.GetSourcesResponse{
			Error: engine.NewProtocolError(engine.ErrUnknown, err),
		}
	}

	return
}

func (api *API) GetResources() (response *protocol.GetResourcesResponse) {
	var err error

	request := &protocol.GetResourcesRequest{
		ContextId: api.Config.ReplContextID,
	}

	if response, err = api.Client.GetResources(_context.TODO(), request); err != nil {
		return
	}

	return
}

func (api *API) CreateContext(name string) (response *protocol.CreateContextResponse) {
	var err error

	request := &protocol.CreateContextRequest{
		Name: name,
	}

	if response, err = api.Client.CreateContext(_context.TODO(), request); err != nil {
		return
	}

	return
}

func (api *API) ExportContexts(contextIds []string) (response *protocol.ExportContextsResponse) {
	var err error

	request := &protocol.ExportContextsRequest{
		ContextIds: contextIds,
	}

	title := "Save"
	title = fmt.Sprintf("%s %d context", title, len(contextIds))
	if len(contextIds) > 1 {
		title = fmt.Sprintf("%ss", title)
	}
	title = fmt.Sprintf("%s as", title)

	if request.OutputPath, err = runtime.SaveFileDialog(api.Runtime.Context(), runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: "context(s).yaml",
		Filters: []runtime.FileFilter{{
			DisplayName: "YAML Files (*.yaml, *.yml)",
			Pattern:     "*.yaml;*.yml",
		}},
		ShowHiddenFiles:      false,
		CanCreateDirectories: true,
	}); err != nil {
		response = &protocol.ExportContextsResponse{
			Error: engine.NewProtocolError(engine.ErrUnknown, err),
		}
		return
	}

	fmt.Printf("Exporting contexts: %+v\n", request)

	if response, err = api.Client.ExportContexts(_context.TODO(), request); err != nil {
		response = &protocol.ExportContextsResponse{
			Error: engine.NewProtocolError(engine.ErrUnknown, err),
		}
	}

	fmt.Printf("Response: %+v\n", response)

	return
}
