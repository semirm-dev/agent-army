"""Parse and write YAML frontmatter in markdown files.

Handles the simple YAML subset used by agent-army: scalar values and
inline ``[a, b]`` or block ``- item`` lists.  No full YAML parser is
needed -- this avoids a PyYAML dependency.
"""

from __future__ import annotations

import os
import re
from pathlib import Path


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def parse_frontmatter(content: str) -> dict[str, str | list[str]]:
    """Parse YAML frontmatter between ``---`` markers.

    Returns a dict where scalar values are ``str`` and list values
    (inline ``[a, b]`` or block ``- item``) are ``list[str]``.
    Missing keys are absent from the dict.

    Handles:
    - Quoted descriptions containing colons.
    - Inline empty lists ``[]``.
    - Block lists with ``- item`` entries.
    - Surrounding single/double quotes on scalar values.
    """
    fm_lines = _extract_frontmatter_lines(content)
    if not fm_lines:
        return {}
    return _parse_fm_lines(fm_lines)


def extract_h1(content: str) -> str:
    """Return the first ``# Heading`` line after frontmatter.

    If the file has no frontmatter the entire content is searched.
    Returns an empty string when no H1 is found.
    """
    past_frontmatter = False
    fm_count = 0
    for line in content.splitlines():
        if line.rstrip() == "---":
            fm_count += 1
            if fm_count >= 2:
                past_frontmatter = True
            continue
        if not past_frontmatter and fm_count == 0:
            # No frontmatter at all -- search from the start.
            past_frontmatter = True
        if past_frontmatter and line.startswith("# "):
            return line[2:].strip()
    return ""


def write_field(file_path: Path, field: str, values: list[str]) -> None:
    """Replace or insert a frontmatter field with *values*.

    If the field already exists in frontmatter it is replaced in-place.
    Otherwise it is inserted just before the closing ``---`` marker.

    Format is always inline: ``field: [val1, val2]`` (or ``field: []``
    for an empty list).  Writes atomically via a ``.tmp`` file and
    :func:`os.replace`.
    """
    content = file_path.read_text(encoding="utf-8")
    new_line = _format_field_line(field, values)
    updated = _replace_or_insert_field(content, field, new_line)

    tmp_path = file_path.with_suffix(file_path.suffix + ".tmp")
    tmp_path.write_text(updated, encoding="utf-8")
    os.replace(tmp_path, file_path)


# ---------------------------------------------------------------------------
# Private helpers
# ---------------------------------------------------------------------------


def _extract_frontmatter_lines(content: str) -> list[str]:
    """Return the lines between the first two ``---`` markers (exclusive)."""
    lines = content.splitlines()
    start: int | None = None
    for idx, line in enumerate(lines):
        if line.rstrip() == "---":
            if start is None:
                start = idx + 1
            else:
                return lines[start:idx]
    return []


def _parse_fm_lines(lines: list[str]) -> dict[str, str | list[str]]:
    """Parse a list of frontmatter lines into a key-value dict."""
    result: dict[str, str | list[str]] = {}
    i = 0
    while i < len(lines):
        line = lines[i]
        match = re.match(r"^([A-Za-z_][A-Za-z0-9_]*):\s*(.*)", line)
        if not match:
            i += 1
            continue

        key = match.group(1)
        raw_value = match.group(2).strip()

        # Inline list: [a, b, c] or []
        if raw_value.startswith("["):
            result[key] = _parse_inline_list(raw_value)
            i += 1
            continue

        # Check for block list on subsequent lines
        if raw_value == "":
            block_items, consumed = _parse_block_list(lines, i + 1)
            if block_items is not None:
                result[key] = block_items
                i += 1 + consumed
                continue
            # Empty scalar
            result[key] = ""
            i += 1
            continue

        # Scalar value -- strip surrounding quotes
        result[key] = _strip_quotes(raw_value)
        i += 1

    return result


def _parse_inline_list(raw: str) -> list[str]:
    """Parse an inline YAML list like ``[a, b, c]`` or ``[]``."""
    inner = raw.strip("[]").strip()
    if not inner:
        return []
    return [_strip_quotes(item.strip()) for item in inner.split(",") if item.strip()]


def _parse_block_list(lines: list[str], start: int) -> tuple[list[str] | None, int]:
    """Try to parse block list items starting at *start*.

    Returns ``(items, lines_consumed)`` on success, or ``(None, 0)``
    if the next line is not a block list item.
    """
    items: list[str] = []
    consumed = 0
    for i in range(start, len(lines)):
        match = re.match(r"^\s+-\s+(.*)", lines[i])
        if match:
            items.append(_strip_quotes(match.group(1).strip()))
            consumed += 1
        else:
            break

    if consumed == 0:
        return None, 0
    return items, consumed


def _strip_quotes(val: str) -> str:
    """Strip surrounding single or double quotes from a value."""
    if len(val) >= 2:
        if (val[0] == '"' and val[-1] == '"') or (val[0] == "'" and val[-1] == "'"):
            return val[1:-1]
    return val


def _format_field_line(field: str, values: list[str]) -> str:
    """Build a frontmatter line like ``field: [val1, val2]``."""
    if not values:
        return f"{field}: []"
    joined = ", ".join(values)
    return f"{field}: [{joined}]"


def _replace_or_insert_field(content: str, field: str, new_line: str) -> str:
    """Replace an existing field or insert before the closing ``---``."""
    lines = content.splitlines(keepends=True)
    fm_dash_count = 0
    field_replaced = False
    result_lines: list[str] = []

    for line in lines:
        stripped = line.rstrip("\n\r")

        if stripped == "---":
            fm_dash_count += 1
            if fm_dash_count == 2 and not field_replaced:
                # Insert before closing ---
                result_lines.append(new_line + "\n")
                field_replaced = True
            result_lines.append(line)
            continue

        if fm_dash_count == 1 and not field_replaced:
            # Inside frontmatter -- check for existing field
            if re.match(rf"^{re.escape(field)}:", stripped):
                result_lines.append(new_line + "\n")
                field_replaced = True
                continue

        result_lines.append(line)

    return "".join(result_lines)
