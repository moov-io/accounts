PLATFORM=$(shell uname -s | tr '[:upper:]' '[:lower:]')
VERSION := $(shell grep -Eo '(v[0-9]+[\.][0-9]+[\.][0-9]+([-a-zA-Z0-9]*)?)' version.go)

.PHONY: build generate

build:
	go fmt ./...
	@mkdir -p ./bin/
	go build github.com/moov-io/accounts
	CGO_ENABLED=1 go build -o ./bin/server github.com/moov-io/accounts/cmd/server

docker:
	docker build --pull -t moov/accounts:$(VERSION) -f Dockerfile .
	docker tag moov/accounts:$(VERSION) moov/accounts:latest

.PHONY: client
client:
# Versions from https://github.com/OpenAPITools/openapi-generator/releases
	@chmod +x ./openapi-generator
	@rm -rf ./client
	OPENAPI_GENERATOR_VERSION=4.0.1 ./openapi-generator generate -i openapi.yaml -g go -o ./client
	rm -f client/go.mod client/go.sum
	go fmt ./...
	go build github.com/moov-io/accounts/client
	go test ./client

clean:
	@rm -rf client/
	@rm -rf tmp/
	@rm -f openapi-generator-cli-*.jar

dist: clean client build
ifeq ($(OS),Windows_NT)
	CGO_ENABLED=1 GOOS=windows go build -o bin/accounts-windows-amd64.exe github.com/moov-io/accounts/cmd/server
else
	CGO_ENABLED=1 GOOS=$(PLATFORM) go build -o bin/accounts-$(PLATFORM)-amd64 github.com/moov-io/accounts/cmd/server
endif

release: docker AUTHORS
	go vet ./...
	go test -coverprofile=cover-$(VERSION).out ./...
	git tag -f $(VERSION)

release-push:
	docker push moov/accounts:$(VERSION)
	docker push moov/accounts:latest

# From https://github.com/genuinetools/img
.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@
