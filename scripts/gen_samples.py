#!/usr/bin/env python3
"""
Generate sample JPEG photos with embedded EXIF data for documentation screenshots.
Produces large (4032x3024) images so the GUI has enough work to show a progress bar.

Requirements: Pillow, piexif
  pip3 install Pillow piexif
"""

import math
import os
import struct
import sys

# ---------------------------------------------------------------------------
# Dependency check
# ---------------------------------------------------------------------------
try:
    from PIL import Image, ImageDraw
    import piexif
except ImportError:
    print("Installing Pillow and piexif...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "--break-system-packages", "Pillow", "piexif"])
    from PIL import Image, ImageDraw
    import piexif

# ---------------------------------------------------------------------------
# Sample EXIF profiles
# ---------------------------------------------------------------------------
SAMPLES = [
    dict(model="ILCE-7RM5", fnumber=(28, 10), exposure=(1, 250), iso=400,  focal=(85, 1),  hue=0),
    dict(model="ILCE-7RM5", fnumber=(18, 10), exposure=(1, 500), iso=200,  focal=(50, 1),  hue=30),
    dict(model="ILCE-7RM5", fnumber=(56, 10), exposure=(1, 60),  iso=800,  focal=(135, 1), hue=60),
    dict(model="ILCE-7RM5", fnumber=(40, 10), exposure=(1, 125), iso=1600, focal=(24, 1),  hue=120),
    dict(model="ILCE-7RM5", fnumber=(14, 10), exposure=(1, 30),  iso=3200, focal=(16, 1),  hue=180),
    dict(model="ILCE-7RM5", fnumber=(22, 10), exposure=(1, 1000),iso=100,  focal=(200, 1), hue=210),
    dict(model="ILCE-7RM5", fnumber=(50, 10), exposure=(1, 200), iso=640,  focal=(70, 1),  hue=240),
    dict(model="ILCE-7RM5", fnumber=(28, 10), exposure=(1, 320), iso=320,  focal=(35, 1),  hue=300),
]

W, H = 4032, 3024   # large enough to produce visible resize work


def make_gradient(width: int, height: int, hue_degrees: float) -> Image.Image:
    """Create a smooth gradient image in a given hue."""
    img = Image.new("RGB", (width, height))
    draw = ImageDraw.Draw(img)
    h = hue_degrees / 360.0
    for x in range(width):
        t = x / width
        # HSV -> RGB: vary lightness across the image
        r, g, b = hsv_to_rgb(h, 0.6 + 0.3 * math.sin(t * math.pi), 0.4 + 0.5 * t)
        draw.line([(x, 0), (x, height)], fill=(int(r * 255), int(g * 255), int(b * 255)))
    return img


def hsv_to_rgb(h, s, v):
    if s == 0:
        return v, v, v
    i = int(h * 6)
    f = h * 6 - i
    p, q, t = v * (1 - s), v * (1 - s * f), v * (1 - s * (1 - f))
    return [(v, t, p), (p, v, t), (t, p, v), (t, p, v), (q, p, v), (v, p, q)][i % 6][::-1]  # noqa


def r(numerator: int, denominator: int):
    """piexif rational tuple."""
    return (numerator, denominator)


def build_exif(profile: dict) -> bytes:
    exif_dict = {
        "0th": {
            piexif.ImageIFD.Make: b"SONY",
            piexif.ImageIFD.Model: profile["model"].encode(),
        },
        "Exif": {
            piexif.ExifIFD.FNumber:          r(*profile["fnumber"]),
            piexif.ExifIFD.ExposureTime:     r(*profile["exposure"]),
            piexif.ExifIFD.ISOSpeedRatings:  profile["iso"],
            piexif.ExifIFD.FocalLength:      r(*profile["focal"]),
        },
        "GPS": {},
        "1st": {},
    }
    return piexif.dump(exif_dict)


def main():
    out_dir = os.path.join(os.path.dirname(__file__), "..", "docs", "examples")
    out_dir = os.path.normpath(out_dir)
    os.makedirs(out_dir, exist_ok=True)

    # Remove old samples
    for f in os.listdir(out_dir):
        if f.lower().endswith((".jpg", ".jpeg")):
            os.remove(os.path.join(out_dir, f))

    print(f"Generating {len(SAMPLES)} sample images → {out_dir}/")
    for i, profile in enumerate(SAMPLES, 1):
        img = make_gradient(W, H, profile["hue"])
        exif_bytes = build_exif(profile)
        path = os.path.join(out_dir, f"sample_{i:02d}.jpg")
        img.save(path, "JPEG", quality=85, exif=exif_bytes)
        size_mb = os.path.getsize(path) / 1_048_576
        print(f"  sample_{i:02d}.jpg  ({size_mb:.1f} MB)  "
              f"f/{profile['fnumber'][0]/profile['fnumber'][1]:.1f}  "
              f"1/{profile['exposure'][1]}s  ISO {profile['iso']}  "
              f"{profile['focal'][0]}mm")

    print("Done.")


if __name__ == "__main__":
    main()
