package client

import (
	"context"
	"fmt"
	"io/fs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	Config "risp/config"
	"risp/protocol"
)

type ClientRuntime interface {
	Context() context.Context
}

func RunClient(config *Config.Config, assets fs.FS) (err error) {
	var connection *grpc.ClientConn

	if connection, err = grpc.Dial(
		fmt.Sprintf("%s:%d", config.GRPCHost, config.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	); err != nil {
		return
	}

	defer connection.Close()

	client := protocol.NewRispClient(connection)

	switch config.UIMode {
	case Config.CLI:
		cli := NewCLI(config, client)

		err = cli.Run()
	case Config.GUI:
		app := NewApp(config, client)
		app.Assets = assets

		err = app.Run()
	}

	return
}
