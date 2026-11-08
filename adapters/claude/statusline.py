#!/usr/bin/env python3
"""Render only known Claude stdin fields; no files, subprocesses or secrets."""
import json, math, sys
try:
    obj = json.load(sys.stdin)
    if not isinstance(obj, dict):
        obj = {}
except (ValueError, OSError):
    obj = {}
model = obj.get("model")
model = model.get("display_name") if isinstance(model, dict) else None
if not isinstance(model, str):
    model = None
if model:
    model = "".join(c for c in model if c.isprintable() and c not in "\x1b\r\n")[:80].strip()
context = obj.get("context_window")
percent = context.get("used_percentage") if isinstance(context, dict) else None
if isinstance(percent, bool) or not isinstance(percent, (int, float)) or not 0 <= percent <= 100 or not math.isfinite(percent):
    percent = None
parts = ["🤖 MoAI", model or "model: --", f"context: {percent:g}%" if percent is not None else "context: --"]
print(" │ ".join(parts))
