APP_NAME ?= intelligent-agent-system
VERSION ?= 1.0.0
ENVIRONMENT ?= staging
VARIANT ?= ias-general

BUILD_DIR ?= bin
PACKAGE_NAME ?= $(APP_NAME)

GITLAB_URL ?= https://ev-gitlab.mataelang.net
PROJECT_ID ?= 54

GOOS := linux
GOARCH ?= amd64
CGO_ENABLED ?= 1

TARGET_ARCHES ?= amd64 arm64

VARIANT_DIRS := $(wildcard cmd/*)
ALL_VARIANTS := $(patsubst cmd/%,%,$(VARIANT_DIRS))
VARIANTS ?= $(ALL_VARIANTS)

BINARY_NAME ?= $(APP_NAME)-$(VARIANT)
ENTRYPOINT := ./cmd/$(VARIANT)/main.go

TARGET_DIR := $(BUILD_DIR)/$(VARIANT)/$(GOOS)-$(GOARCH)
BINARY_PATH := $(TARGET_DIR)/$(BINARY_NAME)

ARCHIVE_FILE := $(APP_NAME)-$(VARIANT)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz
ARCHIVE := $(BUILD_DIR)/$(ARCHIVE_FILE)

# Optional cross C compilers for CGO builds.
# These are only used when cross-compiling from another architecture.
CC_AMD64 ?= x86_64-linux-gnu-gcc
CC_ARM64 ?= aarch64-linux-gnu-gcc
SYSROOT_ARM64 ?= /usr/aarch64-redhat-linux/sys-root/fc42

CGO_CFLAGS ?= -O2 -D_FORTIFY_SOURCE=3 -fstack-protector-strong -fPIE
CGO_CPPFLAGS ?= -O2 -D_FORTIFY_SOURCE=3 -fstack-protector-strong -fPIE
CGO_CXXFLAGS ?= -O2 -D_FORTIFY_SOURCE=3 -fstack-protector-strong -fPIE
CGO_LDFLAGS ?= -Wl,-z,relro,-z,now -Wl,-z,noexecstack -Wl,-z,defs -pie

GOFLAGS_RELEASE := -trimpath -buildvcs=true -buildmode=pie
LDFLAGS_RELEASE := -s -w \
	-linkmode=external \
	-extldflags '$(CGO_LDFLAGS)' \
	-X 'main.Version=$(VERSION)' \
	-X 'main.Environment=$(ENVIRONMENT)' \
	-X 'main.Variant=$(VARIANT)'

.DEFAULT_GOAL := build

.PHONY: \
	all \
	build \
	build-variants \
	release \
	release-build \
	release-arches \
	release-variants \
	release-all \
	package \
	verify-hardening \
	checksec \
	list-variants \
	check-env \
	publish \
	publish-arches \
	publish-variants \
	publish-all \
	upload \
	validate-linux \
	validate-variant \
	clean

all: build

build: validate-variant
	@echo ">> Building local variant: $(VARIANT)"
	@mkdir -p "$(BUILD_DIR)"
	@go build -o "$(BUILD_DIR)/$(BINARY_NAME)" "$(ENTRYPOINT)"

build-variants:
	@set -e; \
	for variant in $(VARIANTS); do \
		echo ""; \
		echo "==> Building local variant $$variant"; \
		$(MAKE) build VARIANT="$$variant"; \
	done

release: release-build verify-hardening package

release-build: validate-linux validate-variant
	@echo ">> Building hardened $(GOOS)/$(GOARCH) binary for variant: $(VARIANT)"
	@mkdir -p "$(TARGET_DIR)"
	@set -e; \
		host_target="$$(go env GOHOSTOS)/$$(go env GOHOSTARCH)"; \
		target="$(GOOS)/$(GOARCH)"; \
		target_cc=""; \
		target_sysroot=""; \
		case "$(GOARCH)" in \
			amd64) \
				target_cc="$(CC_AMD64)" ;; \
			arm64) \
				target_cc="$(CC_ARM64)"; \
				target_sysroot="$(SYSROOT_ARM64)" ;; \
			*) \
				echo "ERROR: unsupported GOARCH=$(GOARCH)"; \
				echo "Supported architectures: $(TARGET_ARCHES)"; \
				exit 1; \
				;; \
		esac; \
		build_cgo_cflags="$(CGO_CFLAGS)"; \
		build_cgo_cppflags="$(CGO_CPPFLAGS)"; \
		build_cgo_cxxflags="$(CGO_CXXFLAGS)"; \
		build_cgo_ldflags="$(CGO_LDFLAGS)"; \
		if [ "$(CGO_ENABLED)" = "1" ] && [ "$$target" != "$$host_target" ]; then \
			if [ -z "$$target_cc" ]; then \
				echo "ERROR: cross-compiling CGO target $$target requires a target C compiler."; \
				exit 1; \
			fi; \
			if ! command -v "$$target_cc" >/dev/null 2>&1; then \
				echo "ERROR: target C compiler not found: $$target_cc"; \
				exit 1; \
			fi; \
			if [ -n "$$target_sysroot" ]; then \
				if [ ! -f "$$target_sysroot/usr/include/stdlib.h" ]; then \
					echo "ERROR: ARM64 sysroot not found or incomplete: $$target_sysroot"; \
					echo "Install it with: sudo dnf install sysroot-aarch64-fc42-glibc"; \
					exit 1; \
				fi; \
				build_cgo_cflags="--sysroot=$$target_sysroot $$build_cgo_cflags"; \
				build_cgo_cppflags="--sysroot=$$target_sysroot $$build_cgo_cppflags"; \
				build_cgo_cxxflags="--sysroot=$$target_sysroot $$build_cgo_cxxflags"; \
				build_cgo_ldflags="--sysroot=$$target_sysroot $$build_cgo_ldflags"; \
				echo ">> Using target sysroot: $$target_sysroot"; \
			fi; \
			echo ">> Cross-compiling from $$host_target to $$target"; \
			echo ">> Using target C compiler: $$target_cc"; \
			export CC="$$target_cc"; \
		fi; \
		CGO_ENABLED="$(CGO_ENABLED)" \
		GOOS="$(GOOS)" \
		GOARCH="$(GOARCH)" \
		CGO_CFLAGS="$$build_cgo_cflags" \
		CGO_CPPFLAGS="$$build_cgo_cppflags" \
		CGO_CXXFLAGS="$$build_cgo_cxxflags" \
		CGO_LDFLAGS="$$build_cgo_ldflags" \
		go build $(GOFLAGS_RELEASE) \
			-ldflags="$(LDFLAGS_RELEASE)" \
			-o "$(BINARY_PATH)" \
			"$(ENTRYPOINT)"

release-arches:
	@set -e; \
	for arch in $(TARGET_ARCHES); do \
		echo ""; \
		echo "==> Releasing variant $(VARIANT) for $(GOOS)/$$arch"; \
		$(MAKE) release VARIANT="$(VARIANT)" GOARCH="$$arch"; \
	done

release-variants:
	@set -e; \
	for variant in $(VARIANTS); do \
		echo ""; \
		echo "==> Releasing variant $$variant for $(GOOS)/$(GOARCH)"; \
		$(MAKE) release VARIANT="$$variant" GOARCH="$(GOARCH)"; \
	done

release-all:
	@set -e; \
	for variant in $(VARIANTS); do \
		for arch in $(TARGET_ARCHES); do \
			echo ""; \
			echo "==> Releasing variant $$variant for $(GOOS)/$$arch"; \
			$(MAKE) release VARIANT="$$variant" GOARCH="$$arch"; \
		done; \
	done

package:
	@echo ">> Packaging artifact..."
	@tar -czf "$(ARCHIVE)" -C "$(TARGET_DIR)" "$(BINARY_NAME)"
	@echo ">> Artifact ready: $(ARCHIVE)"

verify-hardening:
	@echo ">> Verifying binary hardening for $(VARIANT) $(GOOS)/$(GOARCH)..."
	@if [ ! -f "$(BINARY_PATH)" ]; then \
		echo "ERROR: missing binary: $(BINARY_PATH)"; \
		echo "Run: make release-build VARIANT=$(VARIANT) GOARCH=$(GOARCH)"; \
		exit 1; \
	fi
	@if ! command -v readelf >/dev/null 2>&1; then \
		echo "ERROR: readelf is required to verify Linux hardening but was not found."; \
		echo "Install binutils, then rerun this target."; \
		exit 1; \
	fi
	@if ! readelf -h "$(BINARY_PATH)" >/dev/null 2>&1; then \
		echo "ERROR: $(BINARY_PATH) is not a readable ELF binary."; \
		exit 1; \
	fi
	@if readelf -h "$(BINARY_PATH)" | grep -q 'Type:[[:space:]]*DYN'; then \
		echo "   OK: PIE ELF type detected"; \
	else \
		echo "ERROR: binary is not PIE; ELF type is not DYN."; \
		readelf -h "$(BINARY_PATH)" | grep 'Type:'; \
		exit 1; \
	fi
	@if ! readelf -W -l "$(BINARY_PATH)" | grep -q 'GNU_STACK'; then \
		echo "ERROR: GNU_STACK program header was not found."; \
		exit 1; \
	fi
	@if readelf -W -l "$(BINARY_PATH)" | grep 'GNU_STACK' | grep -q 'E'; then \
		echo "ERROR: GNU_STACK is executable."; \
		readelf -W -l "$(BINARY_PATH)" | grep 'GNU_STACK'; \
		exit 1; \
	else \
		echo "   OK: GNU_STACK is not executable"; \
	fi
	@if readelf -W -l "$(BINARY_PATH)" | grep -q 'GNU_RELRO'; then \
		echo "   OK: GNU_RELRO segment detected"; \
	else \
		echo "ERROR: GNU_RELRO segment was not found."; \
		exit 1; \
	fi
	@if readelf -W -d "$(BINARY_PATH)" 2>/dev/null | grep -E 'BIND_NOW|FLAGS.*NOW' >/dev/null; then \
		echo "   OK: immediate binding / BIND_NOW detected"; \
	else \
		echo "ERROR: BIND_NOW / immediate binding was not detected."; \
		exit 1; \
	fi

checksec: verify-hardening

list-variants:
	@printf '%s\n' $(ALL_VARIANTS)

validate-linux:
	@if [ "$(GOOS)" != "linux" ]; then \
		echo "ERROR: this Makefile currently supports Linux release targets only."; \
		echo "Received GOOS=$(GOOS)"; \
		exit 1; \
	fi

validate-variant:
	@if [ ! -d "cmd/$(VARIANT)" ]; then \
		echo "ERROR: unknown variant: $(VARIANT)"; \
		echo "Available variants:"; \
		printf '  %s\n' $(ALL_VARIANTS); \
		exit 1; \
	fi

check-env:
	@if [ -z "$(GITLAB_TOKEN)" ]; then \
		echo "ERROR: GITLAB_TOKEN is not set!"; \
		echo "Run: export GITLAB_TOKEN=glpat-xxxx"; \
		exit 1; \
	fi

publish: check-env release upload

publish-arches: check-env
	@set -e; \
	for arch in $(TARGET_ARCHES); do \
		echo ""; \
		echo "==> Publishing variant $(VARIANT) for $(GOOS)/$$arch"; \
		$(MAKE) publish VARIANT="$(VARIANT)" GOARCH="$$arch"; \
	done

publish-variants: check-env
	@set -e; \
	for variant in $(VARIANTS); do \
		echo ""; \
		echo "==> Publishing variant $$variant for $(GOOS)/$(GOARCH)"; \
		$(MAKE) publish VARIANT="$$variant" GOARCH="$(GOARCH)"; \
	done

publish-all: check-env
	@set -e; \
	for variant in $(VARIANTS); do \
		for arch in $(TARGET_ARCHES); do \
			echo ""; \
			echo "==> Publishing variant $$variant for $(GOOS)/$$arch"; \
			$(MAKE) publish VARIANT="$$variant" GOARCH="$$arch"; \
		done; \
	done

upload:
	@echo ">> Uploading $(ARCHIVE_FILE) to GitLab Package Registry..."
	@curl --fail --silent --show-error \
		--header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" \
		--upload-file "$(ARCHIVE)" \
		"$(GITLAB_URL)/api/v4/projects/$(PROJECT_ID)/packages/generic/$(PACKAGE_NAME)/$(VERSION)/$(ARCHIVE_FILE)"
	@echo ">> Upload complete."

clean:
	@echo "Cleaning artifact directory..."
	rm -rf "$(BUILD_DIR)"
	@echo "Done."
