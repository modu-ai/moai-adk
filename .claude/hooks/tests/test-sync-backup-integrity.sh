#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SCRIPT="$ROOT/.claude/hooks/moai/verify-sync-backup.sh"
RULE="$ROOT/.claude/rules/moai/workflow/sync-backup-integrity.md"
DOC="$ROOT/.claude/skills/moai/workflows/sync/doc-execution.md"

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT
mkdir -p "$TMP_ROOT/project/docs" "$TMP_ROOT/backup"
printf 'read me\n' > "$TMP_ROOT/project/README.md"
printf 'doc\n' > "$TMP_ROOT/project/docs/index.md"
bash "$SCRIPT" create "$TMP_ROOT/backup" "$TMP_ROOT/project" README.md docs/index.md
bash "$SCRIPT" verify "$TMP_ROOT/backup"
printf 'tampered\n' >> "$TMP_ROOT/backup/README.md"
if bash "$SCRIPT" verify "$TMP_ROOT/backup" >/dev/null 2>&1; then
    printf 'FAIL: tampered backup was accepted\n' >&2
    exit 1
fi

grep -Fq 'manifest.tsv' "$RULE"
grep -Fq 'SHA-256' "$RULE"
grep -Fq 'non-empty directory' "$DOC"
grep -Fq 'verify-sync-backup.sh verify' "$DOC"

printf 'PASS: sync backup uses a hash manifest and rejects tampering\n'
