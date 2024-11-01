# Makefile
.PHONY: test test-verbose test-cover test-report

# Run all tests
test:
	go test ./...

# Run all tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage statistics
test-cover:
	go test -cover ./...

# Generate and open coverage report
test-report:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Run tests with race detection
test-race:
	go test -race ./...

# Run specific test pattern
test-pattern:
	@read -p "Enter test pattern: " pattern; \
	go test -v -run $$pattern ./...

# Clean test cache
test-clean:
	go clean -testcache