build:
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.rice/rice
	rice embed-go
	go build .
