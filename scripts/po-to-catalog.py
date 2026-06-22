#!/usr/bin/env python3
"""Convert SearXNG .po translation files to SearGo JSON catalogs.

Usage:
  python3 scripts/po-to-catalog.py \
    --po translations/zh_Hans_CN/LC_MESSAGES/messages.po \
    --locale zh-CN \
    --out-go internal/i18n/catalogs/zh-CN.json \
    --out-web web/public/locales/zh-CN.json

If --locale is omitted, derives it from the po file path (e.g. zh_Hans_CN -> zh-CN).
"""

import argparse
import json
import os
import re
import sys
from pathlib import Path


def parse_po(path: str) -> dict[str, str]:
    """Parse a .po file and return a msgid -> msgstr dict for non-plural entries."""
    messages: dict[str, str] = {}
    msgid_lines: list[str] = []
    msgstr_lines: list[str] = []
    in_msgid = False
    in_msgstr = False
    in_msgid_plural = False

    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.rstrip("\n")

            if line.startswith("msgid_plural "):
                in_msgid_plural = True
                continue
            if line.startswith("msgstr[") and in_msgid_plural:
                continue

            if line.startswith("msgid "):
                in_msgid = True
                in_msgstr = False
                in_msgid_plural = False
                msgid_lines = [_parse_quoted(line[6:])]
                msgstr_lines = []
            elif line.startswith("msgstr "):
                in_msgid = False
                in_msgstr = True
                in_msgid_plural = False
                msgstr_lines = [_parse_quoted(line[7:])]
            elif line.startswith('"') and in_msgid:
                msgid_lines.append(_parse_quoted(line))
            elif line.startswith('"') and in_msgstr:
                msgstr_lines.append(_parse_quoted(line))
            elif line.strip() == "":
                if msgid_lines and msgstr_lines and not in_msgid_plural:
                    msgid = "".join(msgid_lines)
                    msgstr = "".join(msgstr_lines)
                    if msgid and msgstr:
                        messages[msgid] = msgstr
                msgid_lines = []
                msgstr_lines = []
                in_msgid = False
                in_msgstr = False
                in_msgid_plural = False

    # Don't miss the last entry
    if msgid_lines and msgstr_lines and not in_msgid_plural:
        msgid = "".join(msgid_lines)
        msgstr = "".join(msgstr_lines)
        if msgid and msgstr:
            messages[msgid] = msgstr

    return messages


def _parse_quoted(s: str) -> str:
    """Extract content from a quoted PO string, handling escapes."""
    # Remove surrounding quotes
    s = s.strip()
    if s.startswith('"') and s.endswith('"'):
        s = s[1:-1]
    # Handle common PO escapes
    s = s.replace('\\"', '"')
    s = s.replace("\\n", "\n")
    s = s.replace("\\t", "\t")
    return s


def convert_python_format(s: str) -> str:
    """Convert Python %-format placeholders to ICU/Go/i18next format.
    %(name)s -> {name}
    %s -> {0} (sequential)
    """
    # Named placeholders: %(name)s -> {name}
    s = re.sub(r"%\((\w+)\)s", r"{\1}", s)
    # Unnamed placeholders: %s -> {n}
    count = 0
    def replace_unnamed(m):
        nonlocal count
        result = "{" + str(count) + "}"
        count += 1
        return result
    s = re.sub(r"%s", replace_unnamed, s)
    return s


def main():
    parser = argparse.ArgumentParser(description="Convert SearXNG .po to SearGo JSON catalogs")
    parser.add_argument("--po", required=True, help="Path to the .po file")
    parser.add_argument("--locale", help="Locale tag (e.g. zh-CN). Derived from path if omitted.")
    parser.add_argument("--out-go", help="Output path for Go catalog JSON")
    parser.add_argument("--out-web", help="Output path for frontend locale JSON")
    args = parser.parse_args()

    po_path = Path(args.po)
    if not po_path.exists():
        print(f"Error: {po_path} not found", file=sys.stderr)
        sys.exit(1)

    locale = args.locale
    if not locale:
        # Derive from path: translations/zh_Hans_CN/... -> zh-Hans-CN
        parts = po_path.parts
        for p in parts:
            if p not in ("translations", "LC_MESSAGES"):
                # zh_Hans_CN -> zh-Hans-CN
                locale = p.replace("_", "-")
                break
        if not locale:
            print("Error: could not derive locale from path, use --locale", file=sys.stderr)
            sys.exit(1)

    print(f"Parsing {po_path} for locale {locale}")

    messages = parse_po(str(po_path))
    print(f"  Found {len(messages)} translated messages")

    converted: dict[str, str] = {}
    for msgid, msgstr in messages.items():
        converted_msgid = convert_python_format(msgid)
        converted_msgstr = convert_python_format(msgstr)
        converted[converted_msgid] = converted_msgstr

    output = json.dumps(converted, ensure_ascii=False, indent=2, sort_keys=True)

    if args.out_go:
        out_path = Path(args.out_go)
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(output, encoding="utf-8")
        print(f"  Wrote Go catalog: {out_path}")

    if args.out_web:
        out_path = Path(args.out_web)
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(output, encoding="utf-8")
        print(f"  Wrote web locale: {out_path}")

    if not args.out_go and not args.out_web:
        print("  No output paths specified (--out-go or --out-web)")

    print("Done.")


if __name__ == "__main__":
    main()
