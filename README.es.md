# img-settings

> Agregá los datos de tu cámara como marca de agua en tus fotos — listas para compartir por WhatsApp.

**[🇬🇧 Read in English](README.md)**

---

img-settings lee los datos EXIF de tus fotos (apertura, velocidad de obturación, ISO, distancia focal, modelo de cámara) y los estampa como una marca de agua discreta en la parte inferior central de la imagen. El resultado siempre es un JPG optimizado para WhatsApp HD (máx. 2560 px).

---

## Descarga

Entrá a la página de [Releases](https://github.com/esteban-plaza/img-settings/releases/latest) y descargá el archivo para tu plataforma:

| Plataforma | Archivo |
|---|---|
| macOS — Apple Silicon (M1/M2/M3/M4) | `img-settings-darwin-arm64` |
| macOS — Intel | `img-settings-darwin-amd64` |
| Windows | `img-settings-windows-amd64.exe` |

**En macOS**, ejecutá esto una sola vez después de descargar para poder abrir la app:
```bash
xattr -d com.apple.quarantine img-settings-darwin-arm64
chmod +x img-settings-darwin-arm64
```

---

## Cómo usarlo

### Interfaz gráfica

1. Abrí la app
2. **Arrastrá y soltá** tus fotos o una carpeta en la ventana — o hacé clic para buscar
3. Ajustá las opciones en la barra inferior:
   - **Subfolders** — procesar también las fotos dentro de subcarpetas (desactivado por defecto)
   - **Opacity** — qué tan visible es la marca de agua (82% por defecto)
   - **Output** — dónde guardar los resultados (por defecto: una carpeta `out/` junto a tus fotos)
4. Listo — hacé clic en **Reveal in Finder** (macOS) o **Open in Explorer** (Windows) para ver tus fotos

### Línea de comandos

```bash
# Procesar todas las fotos de una carpeta
img-settings-cli-macos /ruta/a/fotos/

# Un archivo individual
img-settings-cli-macos foto.jpg

# Carpeta de salida y opacidad personalizadas
img-settings-cli-macos -out /ruta/salida -opacity 0.65 /ruta/a/fotos/
```

---

## Formatos soportados

| Formato | Soporte |
|---|---|
| JPG / JPEG | ✓ |
| PNG | ✓ |
| ARW (Sony RAW) | ✓ — usa el preview embebido; usa `dcraw` o ImageMagick como respaldo |

Para archivos ARW puede que necesites instalar alguna de estas herramientas:
```bash
brew install dcraw          # recomendado
brew install imagemagick    # alternativa
```

---

## Compilar desde el código fuente

Requiere Go 1.21+ con CGO habilitado.

```bash
git clone https://github.com/esteban-plaza/img-settings.git
cd img-settings

make dev      # build rápido para la máquina actual
make macos    # macOS arm64 + amd64
make windows  # Windows — requiere: brew install mingw-w64
make all      # compilar todo
```

---

## Licencia

[MIT](LICENSE)
