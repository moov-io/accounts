PLATFORM=$(shell uname -s | tr '[:upper:]' '[:lower:]')
VERSION := $(shell grep -Eo '(v[0-9]+[\.][0-9]+[\.][0-9]+([-a-zA-Z0-9]*)?)' version.go)

.PHONY: build generate

build:
	go fmt ./...
	@mkdir -p ./bin/
	go build github.com/moov-io/gl
	CGO_ENABLED=1 go build -o ./bin/server github.com/moov-io/gl/cmd/server

docker:
	docker build --pull -t moov/gl:$(VERSION) -f Dockerfile .
	docker tag moov/gl:$(VERSION) moov/gl:latest

.PHONY: client
client:
# Versions from https://github.com/OpenAPITools/openapi-generator/releases
	@chmod +x ./openapi-generator
	@rm -rf ./client
	OPENAPI_GENERATOR_VERSION=4.0.0-beta2 ./openapi-generator generate -i openapi.yaml -g go -o ./client
	go fmt ./client
	go build github.com/moov-io/gl/client
	go test ./client

clean:
	@rm -rf client/
	@rm -rf tmp/
	@rm -f openapi-generator-cli-*.jar

dist: clean generate build
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=1 GOOS=windows go build -o bin/gl-windows-amd64.exe github.com/moov-io/gl/cmd/server
else
	CGO_ENABLED=1 GOOS=$(PLATFORM) go build -o bin/gl-$(PLATFORM)-amd64 github.com/moov-io/gl/cmd/server
endif

release: docker AUTHORS
	go vet ./...
	go test -coverprofile=cover-$(VERSION).out ./...
	git tag -f $(VERSION)

release-push:
	docker push moov/gl:$(VERSION)

# From https://github.com/genuinetools/img
.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@
