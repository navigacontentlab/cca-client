.PHONY: test
test:
	go test -short -v ./...
	golangci-lint run ./...
