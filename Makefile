
# CC=go build

# BIN=bin
BUILD=build

.PHONY: clean clean-protocol protocol build run-service run-repl run-gui watch-gui

default: build

clean:
# rm -Rf $(BIN) || true
	rm -Rf $(BUILD) || true

clean-protocol:
	rm protocol/*.pb.go || true

protocol: clean-protocol
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		protocol/*.proto

build: clean protocol
# mkdir -p $(BIN)
# $(CC) -o $(BIN)/risp main.go

	wails build

ENV_PATRIK= \
	PATH_PID_FILE=/Users/patrik/projects/doceo/risp/tmp/risp.pid \
	PATH_LOG_FILE=/Users/patrik/projects/doceo/risp/tmp/risp.log \
	PATH_DATA=/Users/patrik/projects/doceo/risp/tmp \
	GRPC_PORT=9999

run-service: build
# clear && $(ENV_PATRIK) $(BIN)/risp service --no-exit
	clear && $(ENV_PATRIK) $(BUILD)/bin/risp.app/Contents/MacOS/risp service --no-exit

run-repl: # build
# clear && $(ENV_PATRIK) $(BIN)/risp
	clear && $(ENV_PATRIK) $(BUILD)/bin/risp.app/Contents/MacOS/risp

run-gui: build
# clear && $(ENV_PATRIK) $(BIN)/risp --gui
	clear && $(ENV_PATRIK) $(BUILD)/bin/risp.app/Contents/MacOS/risp --gui

watch-gui: clean
	DEFAULT_UI_MODE=gui $(ENV_PATRIK) wails dev
