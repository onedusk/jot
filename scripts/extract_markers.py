#!/usr/bin/env python3
"""Extract key action lines from the session transcript for downstream analysis."""
from __future__ import annotations

from pathlib import Path

MARKERS = {"⏺", "⎿", ">"}
MAX_LEADING_SPACES = 4
SOURCE_FILE = Path(__file__).resolve().parent.parent / "2025-10-03-this-session-is-being-continued-from-a-previous-co.txt"
OUTPUT_FILE = Path(__file__).resolve().parent.parent / "2025-10-03-extracted.txt"


def is_target_line(line: str) -> bool:
    """Return True if the line should be captured based on marker and indentation rules."""
    if not line.strip():
        return False

    stripped = line.lstrip(" ")
    leading_spaces = len(line) - len(stripped)
    if leading_spaces > MAX_LEADING_SPACES:
        return False

    stripped = stripped.lstrip("\u00a0")
    if not stripped:
        return False

    if stripped[0].isdigit():
        return False

    return stripped[0] in MARKERS


def extract_lines(path: Path) -> list[str]:
    """Collect lines that conform to the extraction rules while preserving format."""
    lines: list[str] = []
    with path.open(encoding="utf-8") as handle:
        for line in handle:
            if is_target_line(line):
                lines.append(line)
    return lines


def write_output(lines: list[str], path: Path) -> None:
    """Write extracted lines to the destination file, preserving original newlines."""
    path.write_text("".join(lines), encoding="utf-8")


def main() -> None:
    if not SOURCE_FILE.exists():
        raise FileNotFoundError(f"Source transcript not found: {SOURCE_FILE}")

    lines = extract_lines(SOURCE_FILE)
    write_output(lines, OUTPUT_FILE)
    print(f"Wrote {len(lines)} lines to {OUTPUT_FILE}")


if __name__ == "__main__":
    main()
