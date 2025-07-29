all: test

vendor:
	go mod vendor

test:
	go test -coverprofile=coverage.out ./...

.PHONY: test
