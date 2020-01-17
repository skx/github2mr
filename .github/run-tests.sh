#!/bin/bash

# Install tools to test our code-quality.
go get -u golang.org/x/lint/golint

# Failures cause aborts
set -e

# Run the linter
echo "Launching linter .."
golint -set_exit_status ./...
echo "Completed linter .."

# Run the vet-checker.
echo "Launching go vet check .."
go vet ./...
echo "Completed go vet check .."

