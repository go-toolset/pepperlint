PEPPERLINT_PKGS=$(shell go list ./...)
PEPPERLINT_BASE_FOLDERS=$(shell ls -d */ | grep -v "testdata")

all: unit

unit: build verify
	@echo "Testing all pepperlint packages"
	@go test -v ./...

unit-race: build verify
	@echo "Testing all pepperlint packages with race enabled"
	@go test -v -race -cpu=1,2,4 ./...

build:
	@echo "Building pepperlint packages"
	@go build ${PEPPERLINT_PKGS}

verify: get-deps lint vet

get-deps:
	@go get github.com/golang/lint/golint

lint:
	@golint ./...

vet:
	@go tool vet --all -shadow ${PEPPERLINT_BASE_FOLDERS}

ci-test: unit-race
