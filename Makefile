.PHONY: test
test:
	go test -v
	golangci-lint run ./...
