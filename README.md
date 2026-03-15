# img-settings

> Add your camera settings as a watermark to your photos — ready to share on WhatsApp.

**[🇪🇸 Leer en español](README.es.md)**

---

img-settings reads the EXIF data embedded in your photo (aperture, shutter speed, ISO, focal length, camera model) and stamps it as a clean pill-shaped watermark at the bottom centre of the image. Output is always JPG at WhatsApp HD quality (max 2560 px).

---

## Download

Go to the [Releases](https://github.com/esteban-plaza/img-settings/releases/latest) page and download the file for your platform:

| Platform | File |
|---|---|
| macOS — Apple Silicon (M1/M2/M3/M4) | `img-settings-darwin-arm64` |
| macOS — Intel | `img-settings-darwin-amd64` |
| Windows | `img-settings-windows-amd64.exe` |

**On macOS**, run this once after downloading to allow the app to open:
```bash
xattr -d com.apple.quarantine img-settings-darwin-arm64
chmod +x img-settings-darwin-arm64
```

---

## How to use

### GUI

1. Open the app
2. **Drag and drop** your photos or a folder onto the window — or click to browse
3. Tweak the settings in the bottom bar:
   - **Subfolders** — also process photos inside subfolders (off by default)
   - **Opacity** — how visible the watermark is (default 82%)
   - **Output** — where to save the results (default: an `out/` folder next to your photos)
4. Done — click **Reveal in Finder** (macOS) or **Open in Explorer** (Windows) to see your photos

### Command line

```bash
# Watermark all photos in a folder
img-settings-cli-macos /path/to/photos/

# Single file
img-settings-cli-macos photo.jpg

# Custom output folder and opacity
img-settings-cli-macos -out /path/to/output -opacity 0.65 /path/to/photos/
```

---

## Supported formats

| Format | Support |
|---|---|
| JPG / JPEG | ✓ |
| PNG | ✓ |
| ARW (Sony RAW) | ✓ — uses embedded preview; falls back to `dcraw` or ImageMagick |

For ARW files, you may need to install one of these tools:
```bash
brew install dcraw          # recommended
brew install imagemagick    # alternative
```

---

## Build from source

Requires Go 1.21+ with CGO enabled.

```bash
git clone https://github.com/esteban-plaza/img-settings.git
cd img-settings

make dev      # quick build for the current machine
make macos    # macOS arm64 + amd64
make windows  # Windows — requires: brew install mingw-w64
make all      # build everything
```

---

## License

[MIT](LICENSE)
