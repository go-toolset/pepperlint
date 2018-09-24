PEPPERLINT_PKGS=$(shell go list ./...)
PEPPERLINT_BASE_FOLDERS=$(shell ls -d */ | grep -v "testdata")

all: unit

unit: build verify
	@echo "Testing all pepperlint packages"
	@go test -cover -v ./...

unit-race: build verify
	@echo "Testing all pepperlint packages with race enabled"
	@go test -cover -v -race -cpu=1,2,4 ./...

build: get-deps
	@echo "Building pepperlint packages"
	@go build ${PEPPERLINT_PKGS}

verify: lint vet

get-deps:
	@echo "Getting dependencies"
	@go get github.com/golang/lint/golint
	@go get github.com/go-yaml/yaml

lint:
	@golint ./...

vet:
	@go tool vet --all -shadow ${PEPPERLINT_BASE_FOLDERS}

ci-test: unit-race
