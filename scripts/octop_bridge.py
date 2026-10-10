#!/usr/bin/env python3
"""Launcher script for Octop Bridge (Go executable)."""

import os
import shutil
import subprocess
import sys
from pathlib import Path

def main():
    binary = shutil.which("octop-bridge")
    if not binary:
        # Check ~/.local/bin or local project bin
        candidates = [
            Path.home() / ".local" / "bin" / "octop-bridge",
            Path(__file__).resolve().parent.parent / "bin" / "octop-bridge",
        ]
        for c in candidates:
            if c.is_file() and os.access(c, os.X_OK):
                binary = str(c)
                break

    if binary:
        sys.exit(subprocess.call([binary] + sys.argv[1:]))

    # Fallback to 'go run'
    project_root = Path(__file__).resolve().parent.parent
    cmd = ["go", "run", str(project_root / "main.go")] + sys.argv[1:]
    sys.exit(subprocess.call(cmd))

if __name__ == "__main__":
    main()
