GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT = $(shell git rev-parse HEAD)
IMAGE_NAME = local/comp2unraid

.PHONY: make docker

make:
ifeq (, $(shell which go))
	@echo "ERROR: go binary is not in PATH"
	exit 1
endif
	@echo "Compiling binaries..."
	env GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Commit=${GIT_COMMIT} -X main.Branch=${GIT_BRANCH}" -o bin/$(basename $(pwd))_linux_arm64
	env GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Commit=${GIT_COMMIT} -X main.Branch=${GIT_BRANCH}" -o bin/$(basename $(pwd))_darwin_arm64
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Commit=${GIT_COMMIT} -X main.Branch=${GIT_BRANCH}" -o bin/$(basename $(pwd))_linux_amd64
	go test
	go vet

docker:
ifeq (, $(shell which docker))
	@echo "ERROR: docker binary is not in PATH"
	exit 1
endif
	@echo "Building docker images..."
	ARGS="--build-arg GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD) --build-arg GIT_COMMIT=$(git rev-parse HEAD) --build-arg IMAGE_NAME=local/comp2unraid"
	-docker build --platform="darwin/arm64" ${ARGS} -t local/comp2unraid .
	docker build --platform="linux/amd64"  ${ARGS} -t local/comp2unraid .
	-docker build --platform="linux/arm64"  ${ARGS} -t local/comp2unraid .