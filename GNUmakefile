# GNUmakefile — terraform-provider-nullcloud
# Mirrors hashicorp/terraform-provider-scaffolding-framework, extended with
# cross-platform build targets.

default: fmt lint test build

PROVIDER_NAME    := nullcloud
NAMESPACE        := we-work-in-the-cloud
REGISTRY         := registry.terraform.io
VERSION          := 0.2.0
BINARY_VERSIONED := terraform-provider-$(PROVIDER_NAME)_v$(VERSION)

# ── Platforms ──────────────────────────────────────────────────────────────────
PLATFORMS := \
	darwin_amd64 \
	darwin_arm64 \
	linux_amd64 \
	linux_arm64 \
	windows_amd64

# ── Local OS/arch for install target ──────────────────────────────────────────
OS_NAME  := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH_RAW := $(shell uname -m)
LOCAL_OS   := $(OS_NAME)
LOCAL_ARCH := $(if $(filter arm64 aarch64,$(ARCH_RAW)),arm64,amd64)
INSTALL_DIR := \
	$(HOME)/.terraform.d/plugins/$(REGISTRY)/$(NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(LOCAL_OS)_$(LOCAL_ARCH)

.PHONY: default fmt lint test testacc build install clean FORCE

fmt:
	gofmt -s -w -e .

lint:
	golangci-lint run

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout=120m ./...

## build: Cross-compile for all platforms into dist/
build: $(addprefix build-,$(PLATFORMS))

## build-<os>_<arch>: Build for one platform, e.g. make build-darwin_arm64
# FORCE ensures the recipe always runs (works on GNU Make 3.81 and newer).
build-%: FORCE
	$(eval GOOS   := $(word 1,$(subst _, ,$*)))
	$(eval GOARCH := $(word 2,$(subst _, ,$*)))
	$(eval EXT    := $(if $(filter windows,$(GOOS)),.exe,))
	@mkdir -p dist
	@echo "→ $(GOOS)/$(GOARCH)"
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 \
		go build -trimpath -o dist/$(BINARY_VERSIONED)_$(GOOS)_$(GOARCH)$(EXT) .

## install: Build for local OS/arch and install into ~/.terraform.d/plugins
install: FORCE
	@mkdir -p "$(INSTALL_DIR)"
	GOOS=$(LOCAL_OS) GOARCH=$(LOCAL_ARCH) CGO_ENABLED=0 \
		go build -trimpath -o "$(INSTALL_DIR)/$(BINARY_VERSIONED)" .
	@echo "Installed: $(INSTALL_DIR)/$(BINARY_VERSIONED)"

clean:
	rm -rf dist

# FORCE is a phony target with no recipe — any target that depends on it
# will always run, regardless of whether its output file exists.
FORCE:
