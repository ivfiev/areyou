
BIN_DIR := bin

all: server e2e

server:
	go vet ./cmd/server
	GOAMD64=v4 go build -o $(BIN_DIR)/areyou ./cmd/server

e2e: server
	go vet ./cmd/e2e
	GOAMD64=v4 go run ./cmd/e2e ./bin/areyou

clean:
	rm -rf $(BIN_DIR)/*

.PHONY: all server e2e clean
