#!/bin/bash
# t382 AC-EH3-006 — gate-restoration proof by defect injection.
#
# Copies one reclassified SPEC into a scratch tree, removes its "Out of Scope"
# H3 sub-headings so `MissingExclusions` fires, then runs `moai spec lint --strict
# --json` with the pre-fix and post-fix binaries and reports each rc + finding shape.
#
# The injection happens ONLY in the scratch copy; .moai/specs/ is never touched.
set -u

SPEC_ID="${1:-SPEC-KANBAN-WORKTREE-001}"
SCRATCH="$(mktemp -d)"
trap 'rm -rf "$SCRATCH"' EXIT

mkdir -p "$SCRATCH/.moai/specs/$SPEC_ID"
cp .moai/specs/"$SPEC_ID"/*.md "$SCRATCH/.moai/specs/$SPEC_ID/"

echo "scratch: $SCRATCH"
echo "before injection, Out of Scope headings:"
grep -c '^### Out of Scope' "$SCRATCH/.moai/specs/$SPEC_ID/spec.md"

# Isolate the rc signal: the scratch tree holds ONE SPEC, so this SPEC's
# `dependencies:` siblings are absent and DependencyExistsRule emits non-advisory
# MissingDependency ERRORs that would pin --strict rc to 1 on BOTH binaries,
# making the rc half of AC-EH3-006 undecidable. Drop the field in the SCRATCH copy
# only, so MissingExclusions is the sole thing rc can key on.
sed -i '' '/^dependencies:/d' "$SCRATCH/.moai/specs/$SPEC_ID/spec.md"

# Inject the defect: demote every "### Out of Scope —" H3 to plain text.
sed -i '' 's|^### Out of Scope|Out of Scope|' "$SCRATCH/.moai/specs/$SPEC_ID/spec.md"
echo "after injection, Out of Scope headings:"
grep -c '^### Out of Scope' "$SCRATCH/.moai/specs/$SPEC_ID/spec.md"

BIN_PRE="$PWD/bin/moai-pre-t382"
BIN_POST="$PWD/bin/moai"

for label in pre post; do
  if [ "$label" = pre ]; then BIN="$BIN_PRE"; else BIN="$BIN_POST"; fi
  echo
  echo "===== $label-fix binary: $BIN ====="
  ( cd "$SCRATCH" && "$BIN" spec lint --strict --json > lint.json 2>lint.err; echo "rc=$?" )
  echo "--- MissingExclusions findings + severity histogram ---"
  python3 - "$SCRATCH/lint.json" <<'PY'
import collections, json, sys
try:
    d = json.load(open(sys.argv[1], encoding="utf-8"))
except Exception as e:
    print("unparseable:", e); raise SystemExit
def walk(o):
    if isinstance(o, dict):
        if "code" in o and "severity" in o:
            yield o
        for v in o.values():
            yield from walk(v)
    elif isinstance(o, list):
        for v in o:
            yield from walk(v)
findings = list(walk(d))
for f in findings:
    if f.get("code") == "MissingExclusions":
        print({k: f.get(k) for k in ("code", "severity", "advisory", "message")})
hist = collections.Counter((f.get("severity"), bool(f.get("advisory"))) for f in findings)
print("severity histogram (severity, advisory):", dict(hist))
nonadv_err = [f.get("code") for f in findings
              if f.get("severity") == "error" and not f.get("advisory")]
print("non-advisory ERROR codes (these decide --strict rc):",
      dict(collections.Counter(nonadv_err)))
PY
done
