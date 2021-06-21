all: deps build test
SRC_DIR := $(shell go env GOPATH)

deps:
	go get -d
	cd $(SRC_DIR)/pkg/mod/github.com/albanseurat/go-fdkaac@v1.0.3 && chmod u+w . && make && cd -

build:
	go build

test:
	go test

clean:
	go clean