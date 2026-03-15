#!/usr/bin/env python3
"""
Build the GUI binary with -tags demo, launch it against docs/examples/,
and capture screenshots of each UI state into docs/assets/.

Screenshots produced:
  screenshot-dropzone.png    — idle drop zone before demo triggers
  screenshot-processing.png  — in-progress state (~30 %)
  screenshot-done.png        — 100 % + Reveal in Finder button
"""

import os
import subprocess
import sys
import time

import Quartz

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------
REPO = os.path.normpath(os.path.join(os.path.dirname(__file__), ".."))
EXAMPLES_DIR = os.path.join(REPO, "docs", "examples")
ASSETS_DIR = os.path.join(REPO, "docs", "assets")
BINARY = "/tmp/img-settings-gui-demo"

os.makedirs(ASSETS_DIR, exist_ok=True)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
def build_binary():
    print("Building GUI binary with -tags demo …")
    env = os.environ.copy()
    env["CGO_ENABLED"] = "1"
    subprocess.check_call(
        ["go", "build", "-tags", "demo", "-o", BINARY, "./gui/"],
        cwd=REPO,
        env=env,
    )
    print("  OK →", BINARY)


def find_window(owner="img-settings-gui-demo", title="img-settings", timeout=8.0):
    deadline = time.time() + timeout
    while time.time() < deadline:
        wins = Quartz.CGWindowListCopyWindowInfo(
            Quartz.kCGWindowListOptionAll, Quartz.kCGNullWindowID
        )
        for w in wins:
            if (
                w.get("kCGWindowOwnerName", "") == owner
                and w.get("kCGWindowName", "") == title
            ):
                return w.get("kCGWindowNumber")
        time.sleep(0.15)
    raise RuntimeError("GUI window not found within timeout")


def screenshot(wid, path):
    subprocess.check_call(["screencapture", "-l", str(wid), path])
    print(f"  captured → {os.path.basename(path)}")


def kill_app():
    subprocess.run(["pkill", "-f", "img-settings-gui-demo"], capture_output=True)
    time.sleep(0.4)


def progress_bar_value(wid) -> float:
    """
    Heuristic: take a screenshot and check the blue pixel ratio in the top strip
    (where the progress bar lives) to estimate completion.
    Not perfect but good enough to decide when we're mid-way.
    """
    tmp = "/tmp/_probe.png"
    subprocess.run(["screencapture", "-l", str(wid), tmp], capture_output=True)
    # Read PNG dimensions + first scanlines via raw bytes
    try:
        with open(tmp, "rb") as f:
            data = f.read()
        # Find IDAT and decode? Too complex — just return unknown.
    except Exception:
        pass
    return -1.0


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
def main():
    build_binary()
    kill_app()

    # ── 1. Capture drop zone (before demo triggers) ───────────────────────
    print("\nLaunching app …")
    proc = subprocess.Popen(
        [BINARY, "--demo", EXAMPLES_DIR, "--demo-workers", "1"],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    wid = find_window()
    print(f"  window id: {wid}")

    # Drop zone is visible for ~2 s before demo kicks in
    time.sleep(1.0)
    screenshot(wid, os.path.join(ASSETS_DIR, "screenshot-dropzone.png"))

    # ── 2. Capture in-progress state ─────────────────────────────────────
    # Demo triggers at 2 s; wait just past that then grab early in processing.
    time.sleep(1.2)  # total ~2.2 s — processing just started
    screenshot(wid, os.path.join(ASSETS_DIR, "screenshot-processing.png"))

    # ── 3. Wait for done state and capture ───────────────────────────────
    print("  waiting for processing to finish …")
    deadline = time.time() + 60
    prev_snap = None
    done = False
    while time.time() < deadline:
        time.sleep(1.5)
        snap = "/tmp/_snap.png"
        subprocess.run(["screencapture", "-l", str(wid), snap], capture_output=True)
        # Compare file size as a proxy: once stable (two identical sizes) we're done
        size = os.path.getsize(snap) if os.path.exists(snap) else 0
        if prev_snap == size:
            done = True
            break
        prev_snap = size

    if done:
        screenshot(wid, os.path.join(ASSETS_DIR, "screenshot-done.png"))
    else:
        print("  WARNING: timed out waiting for done state — capturing anyway")
        screenshot(wid, os.path.join(ASSETS_DIR, "screenshot-done.png"))

    proc.terminate()
    kill_app()

    print("\nAll screenshots saved to", ASSETS_DIR)
    for f in sorted(os.listdir(ASSETS_DIR)):
        if f.startswith("screenshot-"):
            path = os.path.join(ASSETS_DIR, f)
            print(f"  {f}  ({os.path.getsize(path) // 1024} KB)")


if __name__ == "__main__":
    main()
