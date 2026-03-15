#!/usr/bin/env bash
# gen-docs.sh — regenerate documentation screenshots
#
# Usage:
#   ./scripts/gen-docs.sh          # full run: generate samples + capture
#   ./scripts/gen-docs.sh --shots  # skip sample generation, re-capture only
#
# Requirements (auto-installed if missing):
#   pip3 install Pillow piexif
#
set -euo pipefail
cd "$(dirname "$0")/.."

SKIP_SAMPLES=0
if [[ "${1:-}" == "--shots" ]]; then
  SKIP_SAMPLES=1
fi

# ── 1. Generate sample photos ─────────────────────────────────────────────
if [[ $SKIP_SAMPLES -eq 0 ]]; then
  echo "▶ Generating sample photos with EXIF …"
  python3 scripts/gen_samples.py
  echo
fi

# ── 2. Capture screenshots ────────────────────────────────────────────────
echo "▶ Capturing screenshots …"
python3 scripts/capture_screenshots.py
echo

echo "✓ Done. Docs assets updated in docs/assets/"
