package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"risp/config"
	"risp/protocol"
)

type CLI struct {
	*API
}

func NewCLI(config *config.Config, client protocol.RispClient) (cli *CLI) {
	cli = &CLI{}
	cli.API = NewAPI(cli, config, client)

	return
}

func (cli *CLI) Context() context.Context {
	return context.TODO()
}

func (cli *CLI) Run() (err error) {
	input := bufio.NewReader(os.Stdin)
	buffer := []byte{}

	fmt.Printf(cli.Config.ReplPrompt)
	for {
		var char byte
		if char, err = input.ReadByte(); err != nil {
			err = nil
			continue
		}

		switch char {
		// case 0x0c:
		// 	fmt.Printf("\033[H\033[2J%s%s", config.ReplPrompt, buffer)
		case '\n':
			line := string(buffer)

			if strings.HasPrefix(line, "query ") {
				response := cli.Query(line[6:])

				fmt.Printf("response: %+v\n", response)
			} else {
				response := cli.Execute(line)

				fmt.Printf("response: %+v\n", response)
			}

			buffer = make([]byte, 0)
			fmt.Printf("%s", cli.Config.ReplPrompt)
		default:
			buffer = append(buffer, char)
		}
	}
}
