.PHONY: lint
lint:
	golangci-lint config verify
	golangci-lint run ./...

.PHONY: fmt
fmt:
	go fmt ./...
