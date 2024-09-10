#!/bin/sh
env GOOS=darwin GOARCH=amd64 go build -v -o bin/$(basename $(pwd))_darwin_amd64 && go test -v && go vet
env GOOS=darwin GOARCH=arm64 go build -v -o bin/$(basename $(pwd))_darwin_arm64 && go test -v && go vet
env GOOS=linux GOARCH=amd64 go build -v -o bin/$(basename $(pwd))_linux_amd64 && go test -v && go vet
