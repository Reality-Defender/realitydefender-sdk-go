# Reality Defender Go SDK Justfile
# Install Just: https://github.com/casey/just

# List available recipes
default:
    @just --list

# Install development dependencies
install-deps:
    go mod download
    go install golang.org/x/lint/golint@latest
    go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Run linting with golint
lint:
    golint -set_exit_status ./realitydefender/...

# Check code formatting with gofmtju
fmt-check:
    gofmt -l -d $(find ./realitydefender -type f -name '*.go')
    test -z "$(gofmt -l $(find ./realitydefender -type f -name '*.go'))"

# Format code with gofmt
fmt:
    gofmt -w $(find ./realitydefender -type f -name '*.go')

# Run tests with go test
test:
    go test -v -timeout=30s ./realitydefender/...

# Run tests with ginkgo
test-ginkgo:
    ginkgo -r --randomize-all --randomize-suites --fail-on-pending --trace --race --show-node-events --timeout=60s realitydefender/

# Run tests with ginkgo and coverage
test-coverage:
    ginkgo -r --randomize-all --randomize-suites --fail-on-pending --cover --trace --race --show-node-events --timeout=60s --output-dir=coverage --coverprofile=coverprofile.out realitydefender/
    go tool cover -html=coverage/coverprofile.out -o coverage/coverage.html

# Run tests with ginkgo and coverage (CI-friendly version)
test-coverage-ci:
    mkdir -p coverage
    ginkgo -r --randomize-all --randomize-suites --fail-on-pending --cover --trace --race --timeout=60s --output-dir=coverage --no-color --keep-going --keep-separate-coverprofiles realitydefender/

# Run all quality checks
check: lint fmt-check test-ginkgo

# Clean generated files
clean:
    rm -f coverprofile.out coverage.html

# Run examples
run-basic:
    cd examples/basic && go run main.go

run-events:
    cd examples/events && go run main.go

run-channels:
    cd examples/channels && go run main.go

run-social url:
    cd examples/social && go run main.go {{ url }}

run-results:
    cd examples/results && go run main.go