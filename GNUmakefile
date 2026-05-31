default: fmt lint test build

PROVIDER_NAME    := nullcloud
NAMESPACE        := we-work-in-the-cloud
REGISTRY         := registry.terraform.io
VERSION          := 0.3.0
BINARY_VERSIONED := terraform-provider-$(PROVIDER_NAME)_v$(VERSION)

OS_NAME  := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH_RAW := $(shell uname -m)
LOCAL_OS   := $(OS_NAME)
LOCAL_ARCH := $(if $(filter arm64 aarch64,$(ARCH_RAW)),arm64,amd64)
INSTALL_DIR := \
	$(HOME)/.terraform.d/plugins/$(REGISTRY)/$(NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(LOCAL_OS)_$(LOCAL_ARCH)

.PHONY: default fmt lint test testacc build install docs

fmt:
	gofmt -s -w -e .

lint:
	golangci-lint run

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout=120m ./...

docs:
	cd tools && go generate ./...

build:
	CGO_ENABLED=0 go build -trimpath -o terraform-provider-$(PROVIDER_NAME) .

install:
	@mkdir -p "$(INSTALL_DIR)"
	GOOS=$(LOCAL_OS) GOARCH=$(LOCAL_ARCH) CGO_ENABLED=0 \
		go build -trimpath -o "$(INSTALL_DIR)/$(BINARY_VERSIONED)" .
	@echo "Installed: $(INSTALL_DIR)/$(BINARY_VERSIONED)"
