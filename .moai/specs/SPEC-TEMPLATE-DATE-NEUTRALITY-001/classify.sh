#!/usr/bin/env bash
# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — committed triage classifier (REQ-TDN-006).
#
# Emits one TSV row per OCCURRENCE-CLASS: (file, date-literal, line-shape, line-no, category).
# The guard's FINDING unit is (file, date-literal); an occurrence-class row is finer, because a
# single finding may appear under two conflicting line shapes in the same file (REQ-TDN-003b).
#
# Run from internal/template/ :   bash <path-to>/classify.sh
# Row count is NOT expected to equal the guard's finding count; see the reconcile block at the end.

set -uo pipefail
ROOT="${1:-templates}"

DATE_RE='202[6-9]-[0-1][0-9]-[0-3][0-9]'

# Emit: file<TAB>date<TAB>lineno<TAB>rawline  for every line carrying a date literal.
emit_lines() {
  find "$ROOT" -type f \( -name '*.md' -o -name '*.tmpl' -o -name '*.yaml' -o -name '*.yml' \
    -o -name '*.sh' -o -name '*.json' -o -name '*.js' -o -name '.gitignore' -o -name '.gitattributes' \) -print0 \
  | while IFS= read -r -d '' f; do
      awk -v F="${f#"$ROOT"/}" -v DRE="$DATE_RE" '
        # fence tracking: a line starting with ``` toggles fenced state
        /^[[:space:]]*```/ { fenced = !fenced }
        {
          line = $0
          tmp  = line
          # extract every distinct date literal on this line
          delete seen
          off = 0
          while (match(tmp, DRE)) {
            d     = substr(tmp, RSTART, RLENGTH)
            abs_s = off + RSTART              # 1-based index of match start in the full line
            abs_e = abs_s + RLENGTH - 1
            # Emulate the guard regex word boundaries (\b...\b). awk ERE has no \b,
            # so check the neighbouring characters are non-word. Without this,
            # "2026-01-06T10:00:00Z" and similar embedded literals produce findings
            # the Go guard does NOT report.
            before = (abs_s > 1)            ? substr(line, abs_s - 1, 1) : ""
            after  = (abs_e < length(line)) ? substr(line, abs_e + 1, 1) : ""
            ok = 1
            if (before ~ /[0-9A-Za-z_]/) ok = 0
            if (after  ~ /[0-9A-Za-z_]/) ok = 0
            if (ok && !(d in seen)) {
              seen[d] = 1
              printf "%s\t%s\t%d\t%d\t%s\n", F, d, NR, fenced, line
            }
            off = abs_e
            tmp = substr(tmp, RSTART + RLENGTH)
          }
        }' "$f"
    done
}

# Classify each emitted line into a line-shape and a category.
emit_lines | awk -F'\t' '
{
  file = $1; date = $2; lineno = $3; fenced = $4; line = $5

  # ---- line shape -------------------------------------------------------
  if (line ~ /^[[:space:]]+updated:[[:space:]]*"?20/)                     shape = (fenced ? "LS-FM-FENCED" : "LS-FM")
  else if (line ~ /^[[:space:]]*(#[[:space:]]*)?(\*\*)?(Last )?Updated(\*\*)?:[[:space:]]*"?20/) shape = "LS-PROSE-STAMP"
  else                                                                     shape = "LS-OTHER"

  # ---- category (evaluated in REQ-TDN-001 order) ------------------------
  # DC-3 first: a deadline literal outranks its carrier shape.
  if (date == "2026-11-22")                       cat = "DC-3"
  else if (file ~ /rules\/moai\/NOTICE\.md$/)     cat = "DC-4"
  else if (shape == "LS-FM")                      cat = "DC-1"
  else if (shape == "LS-FM-FENCED")               cat = "DC-5"
  else if (shape == "LS-PROSE-STAMP") {
      # DC-2b = mirror-capture stamp on a third-party documentation mirror (PRESERVE)
      if (file ~ /skills\/moai-foundation-cc\/reference\//) cat = "DC-2b"
      else                                                  cat = "DC-2a"
  }
  else                                            cat = "DC-5"

  printf "%s\t%s\t%s\t%s\t%s\n", file, date, shape, lineno, cat
}' | sort -u
