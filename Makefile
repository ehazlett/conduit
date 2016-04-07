CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${TAG:-latest}
NAME=conduit
REPO=ehazlett/$(NAME)

all: deps build

clean:
	@rm -rf Godeps/_workspace $(NAME)

build:
	@godep go build -a -tags 'netgo' -ldflags '-w -linkmode external -extldflags -static' .

build-container:
	@docker build -t $(NAME)-build -f Dockerfile.build .
	@docker run -it -e BUILD -e TAG --name $(NAME)-build -ti $(NAME)-build make build
	@docker cp $(NAME)-build:/go/src/github.com/$(REPO)/$(NAME) ./
	@docker rm -fv $(NAME)-build

image: build
	@echo Building $(NAME) image $(TAG)
	@docker build -t $(REPO):$(TAG) .

release: deps build image
	@docker push $(REPO):$(TAG)

test: clean 
	@godep go test -v ./...

.PHONY: all deps build clean image test release
