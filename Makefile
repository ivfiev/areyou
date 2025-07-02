
BIN_DIR := bin

all: server cli

server:
	go vet ./cmd/server
	GOAMD64=v4 go build -o $(BIN_DIR)/server ./cmd/server

cli:
	go vet ./cmd/cli
	GOAMD64=v4 go build -o $(BIN_DIR)/cli ./cmd/cli

clean:
	rm -rf $(BIN_DIR)/*

.PHONY: all server cli clean