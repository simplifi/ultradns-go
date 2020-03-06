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

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Clean the examples
.PHONY: clean-examples
clean-examples:
	rm status

# Build the examples
.PHONY: examples
examples: clean-examples
	make status

# Build any example with make <name>
% :: examples/%/main.go
	go build -o $* examples/$*/main.go
