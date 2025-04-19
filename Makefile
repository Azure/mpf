#     MIT License
# 
#     Copyright (c) Microsoft Corporation.
# 
#     Permission is hereby granted, free of charge, to any person obtaining a copy
#     of this software and associated documentation files (the "Software"), to deal
#     in the Software without restriction, including without limitation the rights
#     to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#     copies of the Software, and to permit persons to whom the Software is
#     furnished to do so, subject to the following conditions:
# 
#     The above copyright notice and this permission notice shall be included in all
#     copies or substantial portions of the Software.
# 
#     THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#     IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#     FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#     AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#     LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#     OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
#     SOFTWARE

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Name of the binary output
BINARY_NAME = azmpf

# Main source file
# MAIN_FILE = main.go

# Output directory for the binary
OUTPUT_DIR = .

all: clean build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	# $(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)  ./cmd
	$(GOBUILD) -ldflags "-X 'main.version=$(shell git describe --tags --always --dirty)' -X 'main.commit=$(shell git rev-parse --short HEAD)' -X 'main.date=$(shell date -u '+%Y-%m-%d %H:%M:%S')'" -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd



test:
	@echo "Running tests..."
	$(GOTEST) -count=1 -v ./pkg/domain ./pkg/infrastructure/ARMTemplateShared ./pkg/infrastructure/mpfSharedUtils ./pkg/infrastructure/authorizationCheckers/terraform

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(OUTPUT_DIR)

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(OUTPUT_DIR)/$(BINARY_NAME)

deps:
	@echo "Fetching dependencies..."
	$(GOGET) ./...

.PHONY: all build test clean run deps

# build for darwin arm64, darwin amd64, linux amd64, and windows amd64
build-all:
	@echo "Building $(BINARY_NAME) for all target platforms..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd

test-e2e-arm: # arm and bicep tests
	@echo "Running end-to-end tests..."
	$(GOTEST) ./e2eTests -v -run TestARM

test-e2e-bicep: # bicep tests
	@echo "Running end-to-end tests..."
	$(GOTEST) ./e2eTests -v -run TestBicep

test-e2e-terraform: # terraform tests
	@echo "Running end-to-end tests..."
	$(GOTEST) ./e2eTests -v -timeout 20m -run TestTerraform

