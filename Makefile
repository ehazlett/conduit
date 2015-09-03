CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${TAG:-latest}
NAME=conduit
REPO=ehazlett/$(NAME)
COMMIT=`git rev-parse --short HEAD`

all: build

clean:
	@rm -f $(NAME)

build:
	@godep go build -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/ehazlett/conduit/version.GitCommit=$(COMMIT)" .

image: build
	@echo Building $(NAME) image $(TAG)
	@docker build -t $(REPO):$(TAG) .

release: build image
	@docker push $(REPO):$(TAG)

test: clean 
	@godep go test -v ./...

.PHONY: all build clean image test release
