package config

import (
	"net"
	"os"
	"reflect"
	"strconv"

	"risp/dump"
)

type UIMode int

const (
	CLI UIMode = iota
	GUI
)

type Options struct {
	ValidateConfiguration bool
}

type Config struct {
	intBitSize    int
	UIMode        UIMode
	ReplPrompt    string
	ReplContextID string
	PIDFilePath   string
	LogFilePath   string
	PathData      string
	GRPCHost      string
	GRPCPort      int64
	GRPCListener  net.Listener
}

func NewConfig(options *Options) (config *Config, err error) {
	var intValue int = 0

	if options == nil {
		options = &Options{}
	}

	config = &Config{
		intBitSize:  reflect.TypeOf(intValue).Bits(),
		UIMode:      CLI,
		ReplPrompt:  "Risp > ",
		PIDFilePath: os.Getenv("PATH_PID_FILE"),
		LogFilePath: os.Getenv("PATH_LOG_FILE"),
		PathData:    os.Getenv("PATH_DATA"),
		GRPCHost:    os.Getenv("GRPC_HOST"),
		GRPCPort:    0,
	}

	switch os.Getenv("DEFAULT_UI_MODE") {
	case "gui":
		config.UIMode = GUI
	case "cli":
		fallthrough
	default:
		config.UIMode = CLI
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if len(grpcPort) > 0 {
		if config.GRPCPort, err = strconv.ParseInt(grpcPort, 10, config.intBitSize); err != nil {
			err = nil
			return
		}
	}

	if len(config.PIDFilePath) < 1 {
		config.PIDFilePath = "/var/run/risp.pid"
	}

	if len(config.LogFilePath) < 1 {
		config.LogFilePath = "/var/log/risp.log"
	}

	if options.ValidateConfiguration {
		if err = config.validate(); err != nil {
			return
		}
	}

	return
}

func (config *Config) ParseConfigFile(path string, options *Options) (err error) {
	var (
		data       []byte
		configYAML *dump.ConfigYAML
	)

	if options == nil {
		options = &Options{}
	}

	if data, err = os.ReadFile(path); err != nil {
		return
	}

	if configYAML, err = dump.DecodeConfigYAML(data); err != nil {
		return
	}

	if len(configYAML.PathPidFile) > 0 {
		config.PIDFilePath = configYAML.PathPidFile
	}

	if len(configYAML.PathLogFile) > 0 {
		config.LogFilePath = configYAML.PathLogFile
	}

	if len(configYAML.PathData) > 0 {
		config.PathData = configYAML.PathData
	}

	if len(configYAML.GRPC.Host) > 0 {
		config.GRPCHost = configYAML.GRPC.Host
	}

	if configYAML.GRPC.Port > 0 {
		config.GRPCPort = int64(configYAML.GRPC.Port)
	}

	if len(configYAML.Repl.Prompt) > 0 {
		config.ReplPrompt = configYAML.Repl.Prompt
	}

	if options.ValidateConfiguration {
		if err = config.validate(); err != nil {
			return
		}
	}

	return
}

func (config *Config) validate() (err error) {
	// stat, err := os.Stat(config.IndexPath)
	// if err != nil {
	// 	return err
	// }

	return
}
