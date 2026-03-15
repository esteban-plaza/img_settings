# img-settings — Manual de usuario

**[🇬🇧 Read in English](USER_GUIDE.md)**

---

## Índice

1. [¿Qué es img-settings?](#1-qué-es-img-settings)
2. [Instalación](#2-instalación)
3. [Usar la aplicación (interfaz gráfica)](#3-usar-la-aplicación-interfaz-gráfica)
4. [Usar la línea de comandos (CLI)](#4-usar-la-línea-de-comandos-cli)
5. [Entendiendo la marca de agua](#5-entendiendo-la-marca-de-agua)
6. [Trabajar con archivos RAW de Sony (ARW)](#6-trabajar-con-archivos-raw-de-sony-arw)
7. [Solución de problemas](#7-solución-de-problemas)

---

## 1. ¿Qué es img-settings?

img-settings es una herramienta que toma tus fotos y les agrega una marca de agua discreta con los datos de la cámara con la que fueron tomadas — apertura, velocidad de obturación, ISO, distancia focal y modelo de cámara. El resultado se exporta como JPG optimizado para WhatsApp HD (máx. 2560 px en el lado más largo, calidad 92).

Funciona con archivos JPG, PNG y ARW (RAW de Sony).

---

## 2. Instalación

### Descarga

Entrá a la [página de Releases](https://github.com/esteban-plaza/img-settings/releases/latest) y descargá el archivo para tu sistema:

| Sistema | Archivo a descargar |
|---|---|
| Mac con Apple Silicon (M1/M2/M3/M4) | `img-settings-darwin-arm64` |
| Mac con procesador Intel | `img-settings-darwin-amd64` |
| Windows | `img-settings-windows-amd64.exe` |

¿No sabés qué Mac tenés? Hacé clic en el menú  → **Acerca de esta Mac**. Si dice "Apple M..." tenés Apple Silicon. Si dice "Intel" tenés un Mac Intel.

### Primera apertura en macOS

macOS bloquea la app la primera vez porque fue descargada de internet y no está firmada con un certificado de desarrollador Apple. Para permitirla:

**Opción A — Terminal (más rápido):**
```bash
xattr -d com.apple.quarantine ~/Downloads/img-settings-darwin-arm64
chmod +x ~/Downloads/img-settings-darwin-arm64
```

Luego hacé doble clic en el archivo para abrirlo.

**Opción B — Configuración del sistema:**
1. Intentá abrir la app — macOS la bloqueará y mostrará una alerta
2. Abrí **Configuración del Sistema → Privacidad y seguridad**
3. Bajá hasta la sección de seguridad y hacé clic en **Abrir de todos modos**

### Windows

Hacé doble clic en `img-settings-windows-amd64.exe`. Si Windows Defender SmartScreen muestra una advertencia, hacé clic en **Más información → Ejecutar de todos modos**.

---

## 3. Usar la aplicación (interfaz gráfica)

### Agregar tus fotos

![Zona de arrastre — lista para recibir fotos](assets/screenshot-dropzone.png)

Al abrir la app verás una zona de arrastre grande en el centro de la ventana. Podés:

- **Arrastrar y soltar** una o más fotos directamente en la ventana
- **Arrastrar y soltar una carpeta** para procesar todas las fotos que contiene
- **Hacer clic en la zona** para abrir el explorador de carpetas

La app empieza a procesar inmediatamente después de que soltás los archivos.

### Barra de configuración

En la parte inferior de la ventana hay una barra con tres opciones:

#### Subfolders (Subcarpetas)
Cuando arrastrás una carpeta, esta opción controla si la app también busca fotos dentro de las subcarpetas.

- **Desactivado (por defecto):** solo procesa las fotos que están directamente dentro de la carpeta que soltaste
- **Activado:** encuentra todas las fotos en cada subcarpeta de forma recursiva

#### Opacity (Opacidad)
Controla qué tan visible es la marca de agua, de 0% (invisible) a 100% (completamente opaca). El valor por defecto es **82%**, que da un aspecto limpio sin ser demasiado intrusivo.

Mové el slider hacia la izquierda para hacer la marca de agua más transparente, o hacia la derecha para hacerla más visible.

#### Output (Carpeta de salida)
La carpeta donde se guardarán las fotos procesadas.

- **Por defecto:** se crea automáticamente una carpeta `out/` dentro de la carpeta que soltaste
- **Personalizado:** hacé clic en el ícono de carpeta para elegir un destino diferente
- **Resetear:** hacé clic en el botón × para volver al valor automático

> Las fotos originales nunca se modifican. img-settings siempre escribe en la carpeta de salida.

### Progreso y resultados

![Procesando archivos](assets/screenshot-processing.png)

Una vez que empieza el procesamiento, la zona de arrastre se reemplaza por una barra de progreso y una lista de archivos. Cada archivo muestra:

- ✓ y el tamaño del archivo cuando termina correctamente
- ✗ y un mensaje de error si algo salió mal

![Listo — todos los archivos procesados](assets/screenshot-done.png)

Cuando terminan todos los archivos, aparecen dos botones:

- **Reveal in Finder / Open in Explorer** — abre la carpeta de salida para que veas tus fotos
- **Process more** — vuelve a la zona de arrastre para procesar otro lote

---

## 4. Usar la línea de comandos (CLI)

La versión CLI es para usuarios que prefieren la terminal o quieren automatizar el procesamiento por lotes.

### Uso básico

```bash
# Procesar todas las fotos de una carpeta
img-settings-cli-macos /ruta/a/fotos/

# Procesar un archivo individual
img-settings-cli-macos foto.jpg

# Procesar varias carpetas a la vez
img-settings-cli-macos ~/Desktop/sesion1/ ~/Desktop/sesion2/
```

### Opciones

| Opción | Por defecto | Descripción |
|---|---|---|
| `-out <carpeta>` | `out/` | Dónde guardar las fotos procesadas |
| `-opacity <0.0–1.0>` | `0.82` | Opacidad de la marca de agua (0 = invisible, 1 = completamente opaca) |

### Ejemplos

```bash
# Guardar en una carpeta específica
img-settings-cli-macos -out ~/Desktop/con-marca/ ~/Fotos/sesion/

# Marca de agua más transparente
img-settings-cli-macos -opacity 0.5 foto.jpg

# Combinar opciones
img-settings-cli-macos -out ~/Desktop/out -opacity 0.9 ~/Fotos/
```

### Salida en consola

La CLI imprime una línea por cada archivo a medida que termina:

```
processing 12 file(s) → out/  [8 workers]

  DSC00123.ARW                              OK  (2.1 MB)
  DSC00124.ARW                              OK  (1.9 MB)
  DSC00125.jpg                              OK  (0.8 MB)
  ...

done: 12 ok, 0 failed
```

---

## 5. Entendiendo la marca de agua

La marca de agua es una etiqueta oscura en forma de píldora ubicada en la **parte inferior central** de la imagen. Muestra la siguiente información cuando está disponible en los datos EXIF de la foto:

| Ícono | Campo | Ejemplo |
|---|---|---|
| Cámara | Modelo de cámara | ILCE-7RM5 |
| Diafragma | Apertura | f/2.8 |
| Reloj | Velocidad de obturación | 1/250 |
| Chip | ISO | ISO 400 |
| Círculos de lente | Distancia focal | 85mm |

**Los campos sin datos simplemente no se muestran.** Si una foto no tiene datos EXIF (por ejemplo, una captura de pantalla), se exporta tal cual sin ninguna marca de agua.

El tamaño de la marca de agua se adapta automáticamente a la imagen — escala con la altura de la foto y reduce la tipografía si es necesario para que todo entre dentro del ancho de la imagen.

---

## 6. Trabajar con archivos RAW de Sony (ARW)

img-settings soporta archivos ARW de Sony. Prueba tres métodos en orden:

1. **Preview JPEG embebido** — la mayoría de las cámaras Sony incluyen un JPEG de resolución completa dentro del archivo ARW. img-settings lo extrae directamente, lo cual es rápido y no requiere herramientas adicionales.

2. **dcraw** — si no se encuentra un preview utilizable, img-settings intenta usar dcraw para decodificar los datos RAW.

3. **ImageMagick** — si dcraw no está instalado, recurre a ImageMagick.

Si ninguno funciona, el archivo se omite con un mensaje de error. Para instalar las herramientas de respaldo en macOS:

```bash
brew install dcraw
# o
brew install imagemagick
```

---

## 7. Solución de problemas

**La app no abre en macOS**
Seguí los pasos en [Primera apertura en macOS](#primera-apertura-en-macos). macOS bloquea las apps que no están firmadas con un certificado de desarrollador Apple.

**La marca de agua no aparece**
La foto probablemente no tiene datos EXIF. Esto puede ocurrir con capturas de pantalla, fotos exportadas desde algunas aplicaciones, o imágenes que tuvieron sus metadatos eliminados. Probá con una foto tomada directamente con una cámara o teléfono.

**Los archivos ARW se omiten con un error**
Instalá dcraw o ImageMagick (ver [Trabajar con archivos RAW de Sony](#6-trabajar-con-archivos-raw-de-sony-arw)).

**La foto de salida tiene el mismo tamaño que el original**
Si tu foto ya es de 2560 px o menos en su lado más largo, img-settings no la redimensiona. El redimensionado solo se aplica a imágenes más grandes.

**Solté una carpeta pero algunas fotos no se procesaron**
Verificá que las fotos estén en un formato soportado (JPG, PNG, ARW). Si están en subcarpetas, asegurate de tener el toggle **Subfolders** activado.
