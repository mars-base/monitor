# 定义编译输出目录
BUILD_DIR := build

# 定义目标文件名称
APP_NAME := monitor

# 定义目标平台
PLATFORMS := windows linux darwin-amd64 darwin-arm64

# 定义每个平台的输出文件名称
WINDOWS_OUTPUT := $(BUILD_DIR)/$(APP_NAME).exe
LINUX_OUTPUT := $(BUILD_DIR)/$(APP_NAME)-linux
MACOS_AMD64_OUTPUT := $(BUILD_DIR)/$(APP_NAME)-darwin-amd64
MACOS_ARM64_OUTPUT := $(BUILD_DIR)/$(APP_NAME)-darwin-arm64

# 默认目标
.PHONY: all
all: $(PLATFORMS)

# 创建build目录
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# 开发模式
.PHONY: dev
dev:
	CompileDaemon \
	-graceful-kill=true \
	-exclude-dir=".git,build" \
	-pattern=".*\.go" \
	-color=true \
	-build="make linux" -command="./$(LINUX_OUTPUT) run -c -m -n -i 2"

windows: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o $(WINDOWS_OUTPUT)

linux: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o $(LINUX_OUTPUT)
	@which upx >/dev/null 2>&1 && upx $(LINUX_OUTPUT) || echo "upx not found, skipping compression"

# macOS目标 (amd64)
darwin-amd64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o $(MACOS_AMD64_OUTPUT)

# macOS目标 (arm64)
darwin-arm64: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -ldflags '-extldflags "-static" -s -w' -o $(MACOS_ARM64_OUTPUT)

# 清理构建文件
clean:
	rm -rf $(BUILD_DIR)

install:
	mkdir -p ~/bucket/tools/monitor
	cp $(BUILD_DIR)/* ~/bucket/tools/monitor
	sudo cp $(LINUX_OUTPUT) /usr/local/bin/monitor
