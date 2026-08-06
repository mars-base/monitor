# Build output directory
BUILD_DIR := build

# Application name
APP_NAME := monitor

# Build metadata (injected via ldflags)
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS   := -s -w \
	-X 'main.AppVersion=$(VERSION)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.BuildTime=$(BUILD_TIME)'

# Target platforms
PLATFORMS := windows linux linux-arm64 darwin-amd64 darwin-arm64

# Output file names per platform
WINDOWS_OUTPUT     := $(BUILD_DIR)/$(APP_NAME).exe
LINUX_OUTPUT       := $(BUILD_DIR)/$(APP_NAME)-linux
LINUX_ARM64_OUTPUT := $(BUILD_DIR)/$(APP_NAME)-linux-arm64
MACOS_AMD64_OUTPUT := $(BUILD_DIR)/$(APP_NAME)-darwin-amd64
MACOS_ARM64_OUTPUT := $(BUILD_DIR)/$(APP_NAME)-darwin-arm64

# Installation directories
PREFIX  ?= /usr/local
BINDIR  ?= $(PREFIX)/bin
DESTDIR ?=

# Detect current OS and architecture via uname (works under sudo without go)
CURRENT_OS   := $(shell uname -s 2>/dev/null || echo Windows)
CURRENT_ARCH := $(shell uname -m 2>/dev/null || echo x86_64)

# Resolve the binary for the current platform
ifeq ($(CURRENT_OS),Windows_NT)
  CURRENT_OUTPUT := $(WINDOWS_OUTPUT)
  INSTALL_NAME   := $(APP_NAME).exe
else ifeq ($(CURRENT_OS)_$(CURRENT_ARCH),Linux_aarch64)
  CURRENT_OUTPUT := $(LINUX_ARM64_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
else ifeq ($(CURRENT_OS),Linux)
  CURRENT_OUTPUT := $(LINUX_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
else ifeq ($(CURRENT_OS)_$(CURRENT_ARCH),Darwin_arm64)
  CURRENT_OUTPUT := $(MACOS_ARM64_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
else ifeq ($(CURRENT_OS),Darwin)
  CURRENT_OUTPUT := $(MACOS_AMD64_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
endif

# ---- Build targets ----

.PHONY: all $(PLATFORMS) dev clean install uninstall

all: $(PLATFORMS)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

windows: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o $(WINDOWS_OUTPUT) .

linux: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o $(LINUX_OUTPUT) .

linux-arm64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -o $(LINUX_ARM64_OUTPUT) .

darwin-amd64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o $(MACOS_AMD64_OUTPUT) .

darwin-arm64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -o $(MACOS_ARM64_OUTPUT) .

clean:
	rm -rf $(BUILD_DIR)

# ---- Dev mode ----

dev:
	CompileDaemon \
	-graceful-kill=true \
	-exclude-dir=".git,build" \
	-pattern=".*\.go" \
	-color=true \
	-build="make linux" -command="./$(LINUX_OUTPUT) run -c -m -n -i 2"

# ---- Install / Uninstall ----

install:
	@if [ ! -f "$(CURRENT_OUTPUT)" ]; then echo "Error: $(CURRENT_OUTPUT) not found. Run 'make' first."; exit 1; fi
	@echo "Installing $(INSTALL_NAME) to $(DESTDIR)$(BINDIR)"
ifeq ($(CURRENT_OS),windows)
	@if not exist "$(DESTDIR)$(BINDIR)" mkdir "$(DESTDIR)$(BINDIR)"
	copy /Y "$(subst /,\,$(CURRENT_OUTPUT))" "$(subst /,\,$(DESTDIR)$(BINDIR)\$(INSTALL_NAME))"
else
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(CURRENT_OUTPUT) $(DESTDIR)$(BINDIR)/$(INSTALL_NAME)
endif

uninstall:
	@echo "Uninstalling from $(DESTDIR)$(BINDIR)"
ifeq ($(CURRENT_OS),windows)
	del /Q "$(subst /,\,$(DESTDIR)$(BINDIR)\$(INSTALL_NAME))" 2>nul
else
	rm -f $(DESTDIR)$(BINDIR)/$(INSTALL_NAME)
endif
