default: fmt lint test build

PROVIDER_NAME    := nullcloud
NAMESPACE        := we-work-in-the-cloud
REGISTRY         := registry.terraform.io
VERSION          := 0.5.0
BINARY_VERSIONED := terraform-provider-$(PROVIDER_NAME)_v$(VERSION)

OS_NAME  := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH_RAW := $(shell uname -m)
LOCAL_OS   := $(OS_NAME)
LOCAL_ARCH := $(if $(filter arm64 aarch64,$(ARCH_RAW)),arm64,amd64)
INSTALL_DIR := \
	$(HOME)/.terraform.d/plugins/$(REGISTRY)/$(NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(LOCAL_OS)_$(LOCAL_ARCH)

.PHONY: default fmt lint test testacc build install dist docs backend-setup

fmt:
	gofmt -s -w -e .

lint:
	golangci-lint run

test: backend-setup
	@echo "Running tests..."; \
	BACKEND_BINARY=$(abspath $(BACKEND_DIR)/nullcloud-backend) TF_ACC=1 go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout=120m ./...

docs:
	cd tools && go generate ./...

build:
	CGO_ENABLED=0 go build -trimpath -o terraform-provider-$(PROVIDER_NAME) .

dist:
	goreleaser release --snapshot --clean

install:
	@mkdir -p "$(INSTALL_DIR)"
	GOOS=$(LOCAL_OS) GOARCH=$(LOCAL_ARCH) CGO_ENABLED=0 \
		go build -trimpath -o "$(INSTALL_DIR)/$(BINARY_VERSIONED)" .
	@echo "Installed: $(INSTALL_DIR)/$(BINARY_VERSIONED)"

# System Tests with Live Backend
# Each test manages its own backend lifecycle using StartBackend() from system_test_helper.go
# See SYSTEM_TESTS.md for detailed documentation
#
# Available targets:
#   make test      - Run unit tests and acceptance tests with live backend
#   make backend-setup - Clone and build the backend binary in .backend-test/
BACKEND_DIR := .backend-test

backend-setup:
	@if [ ! -d "$(BACKEND_DIR)" ]; then \
		echo "Cloning backend repository..."; \
		git clone https://github.com/we-work-in-the-cloud/backend-nullcloud $(BACKEND_DIR); \
	fi
	@echo "Building backend..."; \
	cd $(BACKEND_DIR) && make build
	@echo "Backend built successfully at $(BACKEND_DIR)/nullcloud-backend"
