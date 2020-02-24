TAG?=""

# Run all tests
.PHONY: test
test: fmt lint vet test-unit go-mod-tidy

# Clean up any cruft left over from old builds
.PHONY: clean
clean:
	rm -f ultradns

# Build a beta version of ultradns
.PHONY: build
build: clean
	CGO_ENABLED=0 go build -o ultradns

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
