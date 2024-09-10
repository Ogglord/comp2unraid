#!/bin/sh
env GOOS=darwin GOARCH=amd64 go build -v -o bin/$(basename $(pwd)) && go test -v && go vet