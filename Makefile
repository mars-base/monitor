# Build output directory
BUILD_DIR := build

# Application name
APP_NAME := monitor

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

# Detect current OS and architecture
CURRENT_OS   := $(shell go env GOOS)
CURRENT_ARCH := $(shell go env GOARCH)

# Resolve the binary for the current platform
ifeq ($(CURRENT_OS),windows)
  CURRENT_OUTPUT := $(WINDOWS_OUTPUT)
  INSTALL_NAME   := $(APP_NAME).exe
else ifeq ($(CURRENT_OS)_$(CURRENT_ARCH),linux_arm64)
  CURRENT_OUTPUT := $(LINUX_ARM64_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
else ifeq ($(CURRENT_OS),linux)
  CURRENT_OUTPUT := $(LINUX_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
else ifeq ($(CURRENT_OS)_$(CURRENT_ARCH),darwin_arm64)
  CURRENT_OUTPUT := $(MACOS_ARM64_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
else ifeq ($(CURRENT_OS),darwin)
  CURRENT_OUTPUT := $(MACOS_AMD64_OUTPUT)
  INSTALL_NAME   := $(APP_NAME)
endif

# ---- Build targets ----

.PHONY: all $(PLATFORMS) dev clean install uninstall

all: $(PLATFORMS)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

windows: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o $(WINDOWS_OUTPUT) .

linux: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o $(LINUX_OUTPUT) .

linux-arm64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o $(LINUX_ARM64_OUTPUT) .

darwin-amd64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o $(MACOS_AMD64_OUTPUT) .

darwin-arm64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -o $(MACOS_ARM64_OUTPUT) .

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

# Build for current platform before install
install: $(CURRENT_OUTPUT)
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
