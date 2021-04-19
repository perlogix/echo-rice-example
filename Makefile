MAIN_PACKAGE := echo-rice-example
PACKAGES:=$(shell go list ./... | grep -v /vendor/)
GO_LINUX := GOOS=linux GOARCH=amd64
GO_OSX := GOOS=darwin GOARCH=amd64
GO_WIN := GOOS=darwin GOARCH=amd64
LDFLAGS := '-s -w'

default: build

build:
	rice embed-go
	GOOS=$(GOOS) CGO_ENABLED=0 go build -a -installsuffix cgo -o $(MAIN_PACKAGE) -ldflags $(LDFLAGS) .

osx:
	rice embed-go
	CGO_ENABLED=0 $(GO_OSX) go build -a -installsuffix cgo -o $(MAIN_PACKAGE) -ldflags $(LDFLAGS) .

linux:
	rice embed-go
	CGO_ENABLED=0 $(GO_LINUX) go build -a -installsuffix cgo -o $(MAIN_PACKAGE) -ldflags $(LDFLAGS) .

windows:
	rice embed-go
	CGO_ENABLED=0 $(GO_WIN) go build -a -installsuffix cgo -o $(MAIN_PACKAGE).exe -ldflags $(LDFLAGS) .

clean:
	find . -name *_gen.go -type f -delete
	rm -f ./$(MAIN_PACKAGE)
	rm -f ./rice-box.go

gofmt:
	go fmt ./...

lint: gofmt
	$(GOPATH)/bin/golint $(PACKAGES)
	$(GOPATH)/bin/golangci-lint run
	$(GOPATH)/bin/gosec -quiet -no-fail ./...

run:
	go run main.go

update-deps:
	go get -u ./...
	go mod tidy

install:
	go get -u github.com/GeertJohan/go.rice/rice
