.PHONY: all build tag push

DOCKER_LOCAL_NAME               = barito-router
DOCKER_REMOTE_NAME              = barito-router
DOCKER_REMOTE_REPOSITORY    	= barito
VERSION                         = git-$(shell git rev-parse --short HEAD)

DOCKER_LOCAL_TARGET				= $(DOCKER_LOCAL_NAME):$(VERSION)
DOCKER_REMOTE_TARGET			= $(DOCKER_REMOTE_REPOSITORY)/$(DOCKER_REMOTE_NAME):$(VERSION)

all: build tag push

build:
	docker build . -t $(DOCKER_LOCAL_TARGET)

tag:
	docker tag $(DOCKER_LOCAL_TARGET) $(DOCKER_REMOTE_TARGET)

push:
	docker push $(DOCKER_REMOTE_TARGET)

test:
	go test -v ./router

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest .

deadcode:
	go run golang.org/x/tools/cmd/deadcode@latest .
