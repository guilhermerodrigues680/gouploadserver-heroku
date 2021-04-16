MODULE=gouploadserver
VERSION=0.0.1-alpha.0
BUILDTIME=$(shell date +"%Y-%m-%dT%T%z")
# LDFLAGS= -ldflags '-X ...version=$(VERSION) -X ....buildTime=$(BUILDTIME)'

.PHONY: default
default: build

.PHONY: clearbin
clearbin:
	rm -rf ./bin

build: main.go clearbin
	go build -v -o ./bin/$(MODULE) ./main.go

.PHONY: run
run: build
	./bin/$(MODULE)

install: main.go
	go list -f '{{.Target}}'
	go install
