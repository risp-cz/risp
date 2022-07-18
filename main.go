package main

import (
	"embed"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"risp/client"
	Config "risp/config"
	"risp/service"
)

//go:embed frontend/dist
var assets embed.FS

var ClientFlags []cli.Flag = []cli.Flag{
	&cli.BoolFlag{
		Name:  "gui",
		Usage: "run client in GUI mode",
	},
	&cli.StringFlag{
		Name:  "c",
		Usage: "risp config file; usage: -c <path to *.yaml>",
	},
}

var ServiceFlags []cli.Flag = []cli.Flag{
	&cli.StringFlag{
		Name:  "s",
		Usage: "service signal; usage: -s [start|stop]",
	},
	&cli.BoolFlag{
		Name:  "no-exit",
		Usage: "run program in the current process instead of detaching as a service",
	},
	&cli.StringFlag{
		Name:  "c",
		Usage: "risp config file; usage: -c <path to *.yaml>",
	},
}

func detectLocalDaemon() (port int64, err error) {
	return
}

func runClient(context *cli.Context) (err error) {
	var (
		config     *Config.Config
		configPath = context.String("c")
	)

	if config, err = Config.NewConfig(&Config.Options{}); err != nil {
		return
	}

	if config.GRPCPort == 0 {
		if config.GRPCPort, err = detectLocalDaemon(); err != nil {
			return
		}
	}

	if len(configPath) > 0 {
		config.ParseConfigFile(configPath, nil)
	}

	if context.Bool("gui") {
		config.UIMode = Config.GUI
	}

	err = client.RunClient(config, assets)
	return
}

func runService(context *cli.Context) (err error) {
	var (
		config     *Config.Config
		configPath = context.String("c")
	)

	if config, err = Config.NewConfig(&Config.Options{}); err != nil {
		return
	}

	if len(configPath) > 0 {
		config.ParseConfigFile(configPath, nil)
	}

	service := service.NewService(config)
	service.NoExit = context.Bool("no-exit")
	service.Signal = context.String("s")

	err = service.Run()
	return
}

func main() {
	cliApp := &cli.App{
		Flags:  ClientFlags,
		Action: runClient,
		Commands: []*cli.Command{{
			Name:   "client",
			Flags:  ClientFlags,
			Action: runClient,
		}, {
			Name:   "service",
			Flags:  ServiceFlags,
			Action: runService,
		}},
	}

	if err := cliApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
