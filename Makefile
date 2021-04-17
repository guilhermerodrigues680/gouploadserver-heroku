MODULE=gouploadserver
VERSION=0.0.1-alpha.0
BUILDTIME=$(shell date +"%Y-%m-%dT%T%z")
# LDFLAGS= -ldflags '-X ...version=$(VERSION) -X ....buildTime=$(BUILDTIME)'

.PHONY: default
default: build

.PHONY: clearbin
clearbin:
	rm -rf ./bin

build: cmd/gouploadserver/main.go clearbin
	GOOS=linux GOARCH=amd64 go build -v -o ./bin/$(MODULE)-linux-amd64 ./cmd/gouploadserver/main.go
	GOOS=windows GOARCH=amd64 go build -v -o ./bin/$(MODULE)-windows-amd64.exe ./cmd/gouploadserver/main.go
	GOOS=darwin GOARCH=amd64 go build -v -o ./bin/$(MODULE)-darwin-amd64 ./cmd/gouploadserver/main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -o ./bin/$(MODULE)-alpine-linux-amd64 ./cmd/gouploadserver/main.go

.PHONY: cross
cross: cmd/gouploadserver/main.go clearbin
	go build -v -o ./bin/$(MODULE) ./cmd/gouploadserver/main.go

.PHONY: run
run: build
	./bin/$(MODULE)

install: cmd/gouploadserver/main.go
	go list -f '{{.Target}}'
	go install
