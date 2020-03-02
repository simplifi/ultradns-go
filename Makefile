TAG?=""

# Run all tests
.PHONY: test
test: fmt lint vet test-unit go-mod-tidy

# Run unit tests
.PHONY: test-unit
test-unit:
	go test -v -race ./...

# Clean go.mod
.PHONY: go-mod-tidy
go-mod-tidy:
	go mod tidy
	git diff --exit-code go.sum

# Check formatting
.PHONY: fmt
fmt:
	test -z "$(shell gofmt -l .)"

# Run linter
.PHONY: lint
lint:
	golint -set_exit_status ./...

# Run vet
.PHONY: vet
vet:
	go vet ./...

# Clean the examples
.PHONY: clean-examples
clean-examples:
	rm status

# Build the examples
.PHONY: examples
examples: clean-examples
	go build -o status examples/status/main.go
