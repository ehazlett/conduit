CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
COMMIT=`git rev-parse --short HEAD`
APP=conduit
REPO?=ehazlett/$(APP)
TAG?=latest

all: build

build:
	@cd cmd/$(APP) && go build -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT)" .

build-static:
	@cd cmd/$(APP) && go build -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/$(REPO)/version.GitCommit=$(COMMIT)" .

release: image
	@docker push $(REPO):$(TAG)

test: build
	@go test -v ./...

clean:
	@rm -rf cmd/$(APP)/$(APP)

.PHONY: build build-static release test clean
