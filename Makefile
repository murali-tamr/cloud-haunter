BINARY=cr
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")

deps: deps-errcheck
	go get -u github.com/golang/dep/cmd/dep
	# go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/kisielk/errcheck

deps-errcheck:
	go get -u github.com/kisielk/errcheck

format:
	@gofmt -w ${GOFILES_NOVENDOR}

vet:
	go vet ./...

test:
	go test -timeout 30s -race ./...

errcheck:
	errcheck -ignoretests ./...

build: errcheck format vet test build-darwin build-linux

build-docker:
	@#USER_NS='-u $(shell id -u $(whoami)):$(shell id -g $(whoami))'
	docker run --rm ${USER_NS} -v "${PWD}":/go/src/github.com/hortonworks/cloud-cost-reducer -w /go/src/github.com/hortonworks/cloud-cost-reducer golang:1.9 make deps-errcheck build

build-darwin:
	GOOS=darwin CGO_ENABLED=0 go build -o build/Darwin/${BINARY} main.go

build-linux:
	GOOS=linux CGO_ENABLED=0 go build -o build/Linux/${BINARY} main.go