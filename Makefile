CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${TAG:-latest}
NAME=conduit
REPO=ehazlett/$(NAME)

all: deps build

deps:
	@godep restore

clean:
	@rm -rf Godeps/_workspace $(NAME)

build: deps
	@godep go build -a -tags 'netgo' -ldflags '-w -linkmode external -extldflags -static' .

image: build
	@echo Building $(NAME) image $(TAG)
	@docker build -t $(REPO):$(TAG) .

release: deps build image
	@docker push $(REPO):$(TAG)

test: clean 
	@godep go test -v ./...

.PHONY: all deps build clean image test release
