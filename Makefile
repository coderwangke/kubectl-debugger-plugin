BIN_NAME := kubectl-debugger
CUR := $(shell pwd)
BUILD_DIR := $(CUR)/_output

.PHONY: all clean darwin-amd64 darwin-arm64 linux-amd64 linux-arm64

all: clean darwin-amd64 darwin-arm64 linux-amd64 linux-arm64

# 在 $(1) 操作系统和 $(2) 架构下编译二进制文件
define build
	@echo "Building $(1)/$(2)..."
	@mkdir -p $(BUILD_DIR)/$(1)-$(2)
	@CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -o $(BUILD_DIR)/$(1)-$(2)/$(BIN_NAME) main.go
endef

define package
	@echo "Packaging $(1)/$(2)..."
	@tar -C $(BUILD_DIR) -czf $(BUILD_DIR)/$(BIN_NAME)-$1-$2.tar.gz $1-$2/$(BIN_NAME)
endef

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

darwin-amd64:
	$(call build,darwin,amd64)
	$(call package,darwin,amd64)

darwin-arm64:
	$(call build,darwin,arm64)
	$(call package,darwin,arm64)

linux-amd64:
	$(call build,linux,amd64)
	$(call package,linux,amd64)

linux-arm64:
	$(call build,linux,arm64)
	$(call package,linux,arm64)
