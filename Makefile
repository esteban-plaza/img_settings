##
## img_settings — build targets
## Build from macOS for macOS (arm64/amd64 universal) and Windows (amd64).
##

GUI_PKG    := ./gui/
CLI_PKG    := ./cli/
DIST       := dist/
ICON       := gui/assets/AppIcon.png
CC_WIN     := x86_64-w64-mingw32-gcc
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS    := -s -w -X main.version=$(VERSION)

.PHONY: all macos windows cli bundle-macos clean fyne-setup check-mingw

## Build everything
all: macos windows cli

## ── macOS ────────────────────────────────────────────────────────────────────

## Native macOS universal binary (arm64 + amd64 joined with lipo)
macos: dist-dir
	@echo "→ Building macOS arm64…"
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
	  go build -ldflags="$(LDFLAGS)" \
	  -o $(DIST)img_settings-darwin-arm64 $(GUI_PKG)

	@echo "→ Building macOS amd64…"
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
	  go build -ldflags="$(LDFLAGS)" \
	  -o $(DIST)img_settings-darwin-amd64 $(GUI_PKG)

	@echo "✓ $(DIST)img_settings-darwin-arm64 (Apple Silicon)"
	@echo "✓ $(DIST)img_settings-darwin-amd64 (Intel)"

## macOS .app bundle (requires fyne CLI: make fyne-setup)
bundle-macos: dist-dir
	@echo "→ Packaging macOS .app bundle…"
	cd gui && ~/go/bin/fyne package \
	  -os darwin \
	  -icon ../$(ICON) \
	  --name img_settings \
	  --appID io.img_settings.app
	mv gui/img_settings.app $(DIST)
	@echo "✓ $(DIST)img_settings.app"

## ── Windows ──────────────────────────────────────────────────────────────────

## Windows amd64 cross-compile (requires mingw-w64: brew install mingw-w64)
windows: dist-dir check-mingw
	@echo "→ Building Windows amd64…"
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=$(CC_WIN) \
	  go build \
	  -ldflags="-H windowsgui $(LDFLAGS)" \
	  -o $(DIST)img_settings-windows-amd64.exe $(GUI_PKG)
	@echo "✓ $(DIST)img_settings-windows-amd64.exe"

## ── CLI ──────────────────────────────────────────────────────────────────────

## CLI binary (no CGO needed)
cli: dist-dir
	@echo "→ Building CLI…"
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" \
	  -o $(DIST)img_settings-cli $(CLI_PKG)
	@echo "✓ $(DIST)img_settings-cli"

## ── Helpers ──────────────────────────────────────────────────────────────────

dist-dir:
	@mkdir -p $(DIST)

check-mingw:
	@which $(CC_WIN) > /dev/null 2>&1 || \
	  (echo "✗ $(CC_WIN) not found. Install with: brew install mingw-w64" && exit 1)

## Install fyne v2 CLI tool
fyne-setup:
	go install fyne.io/tools/cmd/fyne@latest

clean:
	rm -rf $(DIST)
	rm -f gui/*.syso

## Quick dev build for the current platform only
dev: dist-dir
	@echo "→ Dev build (native)…"
	CGO_ENABLED=1 go build -o $(DIST)img_settings-gui-dev $(GUI_PKG)
	@echo "✓ $(DIST)img_settings-gui-dev"
