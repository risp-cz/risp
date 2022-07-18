package client

import (
	_context "context"
	"fmt"
	"io/fs"
	_runtime "runtime"

	"risp/config"
	"risp/protocol"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	*API
	context _context.Context
	Assets  fs.FS
}

func NewApp(config *config.Config, client protocol.RispClient) (app *App) {
	app = &App{}
	app.API = NewAPI(app, config, client)

	return
}

func (app *App) Context() _context.Context {
	return app.context
}

func (app *App) Run() error {
	return wails.Run(&options.App{
		Title:            "Risp",
		Width:            1024,
		Height:           768,
		Frameless:        false,
		AlwaysOnTop:      false,
		Assets:           app.Assets,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 0},
		Windows: &windows.Options{
			WindowIsTranslucent:  false,
			WebviewIsTransparent: false,
			Theme:                windows.SystemDefault,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:  windows.RGB(20, 20, 20),
				DarkModeTitleText: windows.RGB(200, 200, 200),
				DarkModeBorder:    windows.RGB(20, 0, 20),

				LightModeTitleBar:  windows.RGB(241, 239, 227),
				LightModeTitleText: windows.RGB(20, 20, 20),
				LightModeBorder:    windows.RGB(241, 239, 227),
			},
		},
		Mac: &mac.Options{
			WindowIsTranslucent:  false,
			WebviewIsTransparent: false,
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false, // true,
				HideTitle:                  false, // true,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       false,
			},
			About: &mac.AboutInfo{
				Title:   "Risp",
				Message: "© 2022 Patrik Šimunič",
			},
		},
		OnStartup:     app.OnStartup,
		OnShutdown:    app.OnShutdown,
		OnDomReady:    app.OnDOMReady,
		OnBeforeClose: app.OnBeforeClose,
		Menu:          app.Menu(),
		Bind: []interface{}{
			app,
		},
	})
}

func (app *App) OnStartup(context _context.Context) {
	app.context = context
}

func (app *App) OnShutdown(context _context.Context) {
}

func (app *App) OnDOMReady(context _context.Context) {
}

func (app *App) OnBeforeClose(context _context.Context) (preventClose bool) {
	return
}

func (app *App) Menu() (appMenu *menu.Menu) {
	appMenu = menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open", keys.CmdOrCtrl("o"), app.handleOpenFile)
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), app.handleQuit)

	if _runtime.GOOS == "darwin" {
		appMenu.Append(menu.EditMenu())
	}

	_ = appMenu.AddSubmenu("Resources")
	_ = appMenu.AddSubmenu("Sources")

	contextsMenu := appMenu.AddSubmenu("Contexts")
	contextsMenu.AddText("Import context(s)", nil, app.handleImportContexts)

	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Online documentation", nil, func(_ *menu.CallbackData) {})

	return
}

func (app *App) handleQuit(callbackData *menu.CallbackData) {
	runtime.Quit(app.context)
}

func (app *App) handleImportContexts(callbackData *menu.CallbackData) {
	filePath, err := runtime.OpenFileDialog(app.context, runtime.OpenDialogOptions{})

	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	fmt.Printf("selected path: %+v\n", filePath)
}

func (app *App) handleOpenFile(callbackData *menu.CallbackData) {
	var (
		err     error
		message string
	)

	if message, err = runtime.MessageDialog(app.context, runtime.MessageDialogOptions{
		Type:    runtime.WarningDialog,
		Title:   "Not implemented",
		Message: "...just yet...",
	}); err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	fmt.Printf("message: %+v\n", message)
}
