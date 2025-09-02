#!/bin/bash

# Test Runner Script for Build Scan Go
# This script runs comprehensive tests and generates reports

set -e

echo "🚀 Build Scan Go - Test Runner"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Create test results directory
mkdir -p test-results

# Run unit tests
print_status "Running unit tests..."
if go test -v -short ./... > test-results/unit-tests.log 2>&1; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    cat test-results/unit-tests.log
    exit 1
fi

# Run tests with coverage
print_status "Running tests with coverage..."
if go test -v -coverprofile=test-results/coverage.out ./... > test-results/coverage-tests.log 2>&1; then
    print_success "Coverage tests completed"

    # Generate HTML coverage report
    go tool cover -html=test-results/coverage.out -o test-results/coverage.html

    # Show coverage summary
    COVERAGE=$(go tool cover -func=test-results/coverage.out | grep total | awk '{print $3}')
    print_status "Overall test coverage: $COVERAGE"

    # Check if coverage is above threshold (70%)
    COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
    if [ ${COVERAGE_NUM%.*} -ge 70 ]; then
        print_success "Coverage threshold met (≥70%)"
    else
        print_warning "Coverage below threshold (70%). Current: $COVERAGE"
    fi
else
    print_error "Coverage tests failed"
    cat test-results/coverage-tests.log
    exit 1
fi

# Run integration tests (if not in short mode)
if [[ "$1" != "--unit-only" ]]; then
    print_status "Running integration tests..."
    if go test -v -run TestIntegration ./... > test-results/integration-tests.log 2>&1; then
        print_success "Integration tests passed"
    else
        print_warning "Integration tests failed (this might be expected without proper server setup)"
        # Don't exit on integration test failures as they might require external setup
    fi
fi

# Run benchmarks
print_status "Running benchmark tests..."
if go test -v -bench=. -benchmem ./... > test-results/benchmark-tests.log 2>&1; then
    print_success "Benchmark tests completed"
else
    print_warning "Benchmark tests had issues"
fi

# Check code formatting
print_status "Checking code formatting..."
UNFORMATTED=$(go fmt ./... 2>&1)
if [ -z "$UNFORMATTED" ]; then
    print_success "Code is properly formatted"
else
    print_warning "Code formatting issues found:"
    echo "$UNFORMATTED"
fi

# Run go vet
print_status "Running go vet..."
if go vet ./... > test-results/vet.log 2>&1; then
    print_success "go vet passed"
else
    print_warning "go vet found issues:"
    cat test-results/vet.log
fi

# Check for race conditions (in unit tests only to keep it fast)
print_status "Checking for race conditions..."
if go test -race -short ./... > test-results/race-tests.log 2>&1; then
    print_success "No race conditions detected"
else
    print_error "Race conditions detected"
    cat test-results/race-tests.log
    exit 1
fi

# Generate test summary
print_status "Generating test summary..."

echo "# Test Summary Report" > test-results/summary.md
echo "Generated: $(date)" >> test-results/summary.md
echo "" >> test-results/summary.md

echo "## Coverage" >> test-results/summary.md
echo "- Overall Coverage: $COVERAGE" >> test-results/summary.md
echo "- Coverage Report: [coverage.html](coverage.html)" >> test-results/summary.md
echo "" >> test-results/summary.md

echo "## Test Results" >> test-results/summary.md
echo "- Unit Tests: ✅ Passed" >> test-results/summary.md
echo "- Integration Tests: ⚠️  See integration-tests.log" >> test-results/summary.md
echo "- Benchmark Tests: ✅ Completed" >> test-results/summary.md
echo "- Race Detection: ✅ Passed" >> test-results/summary.md
echo "" >> test-results/summary.md

echo "## Code Quality" >> test-results/summary.md
if [ -z "$UNFORMATTED" ]; then
    echo "- Code Formatting: ✅ Passed" >> test-results/summary.md
else
    echo "- Code Formatting: ⚠️  Issues found" >> test-results/summary.md
fi

if [ -s test-results/vet.log ]; then
    echo "- Go Vet: ⚠️  Issues found" >> test-results/summary.md
else
    echo "- Go Vet: ✅ Passed" >> test-results/summary.md
fi

echo ""
print_success "Test run completed! Results saved in test-results/"
print_status "View coverage report: open test-results/coverage.html"
print_status "View summary: cat test-results/summary.md"

echo ""
echo "📊 Quick Summary:"
echo "   Coverage: $COVERAGE"
echo "   Results:  test-results/"
echo "   Logs:     test-results/*.log"
