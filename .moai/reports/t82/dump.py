#!/usr/bin/env python3
"""Dump clause blocks with an index id for classification."""
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
b = json.load(open(os.path.join(HERE, "blocks.json"), encoding="utf-8"))
for i, x in enumerate(b):
    print("=== B%03d | %s:%d-%d | %dB | %s" % (
        i, x["file"], x["line"], x["endline"], x["bytes"],
        "clause" if x["clause_initial"] else "PROSE-MENTION"))
    print(x["text"].rstrip())
    print()
