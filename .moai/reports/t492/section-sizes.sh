#!/usr/bin/env bash
# t492 — byte size of each top-level (##) and second-level (###) section of a markdown file.
# Fenced code blocks are tracked so a '#' inside a fence is not mistaken for a heading.
set -u
awk '
  /^```/ { fence = !fence }
  {
    if (!fence && ($0 ~ /^## / || $0 ~ /^### /)) {
      if (hdr != "") printf "%8d  %s\n", bytes, hdr
      hdr = $0
      bytes = 0
    }
    bytes += length($0) + 1
  }
  END { if (hdr != "") printf "%8d  %s\n", bytes, hdr }
' "$1"
