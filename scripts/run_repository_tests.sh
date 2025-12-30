#!/bin/bash

# Script to run repository tests

set -e

echo "====================================="
echo "Running Repository Tests"
echo "====================================="

cd "$(dirname "$0")/.."

# Check if required dependencies are installed
echo "Checking test dependencies..."
go mod download

# Run tests with coverage
echo ""
echo "Running tests..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./internal/repository/... 2>&1

# Display coverage summary
echo ""
echo "====================================="
echo "Coverage Summary"
echo "====================================="
go tool cover -func=coverage.out | grep total

# Optionally generate HTML coverage report
if [ "$1" == "html" ]; then
    echo ""
    echo "Generating HTML coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report generated: coverage.html"
fi

echo ""
echo "====================================="
echo "Tests Complete!"
echo "====================================="
