.PHONY: all clean

TARGETS := \
    linux-amd64 \
    darwin-arm64

BIN_NAME := kubectl-debugger
CUR := $(shell pwd)
BUILD_DIR := $(CUR)/_output

ifeq ($(OS),Windows_NT)
    BIN_NAME := $(BIN_NAME).exe
endif

# 在 $(1) 操作系统和 $(2) 架构下编译二进制文件
define build
	@echo "Building $(1)/$(2)..."
	@mkdir -p $(BUILD_DIR)/$(1)-$(2)
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -o $(BUILD_DIR)/$(1)-$(2)/$(BIN_NAME) main.go
endef

package-all: $(foreach t,$(TARGETS),$(BUILD_DIR)/$(t)/$(BIN_NAME))
	@echo "Packaging all targets..."
	@cd $(BUILD_DIR) \
		&& for t in $(TARGETS); do \
			echo "Creating tar.gz archive for $$t..."; \
			tar -czf $(BIN_NAME)-$$t.tar.gz $$t; \
		done

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

darwin-amd64:
	$(call build,darwin,amd64)

darwin-arm64:
	$(call build,darwin,arm64)

linux-amd64:
	$(call build,linux,amd64)

windows-amd64:
	$(call build,windows,amd64)

linux-arm64:
	$(call build,linux,arm64)