#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
HOOK="$ROOT/.claude/hooks/moai/sync-phase-quality-gate.sh"
RULE="$ROOT/.claude/rules/moai/workflow/language-routing-contract.md"
DOC="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-quality.md"

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

set -- ""
source "$HOOK"

mkdir -p "$TMP_ROOT/kotlin"
printf 'plugins { kotlin("jvm") version "2.0.0" }\n' > "$TMP_ROOT/kotlin/build.gradle.kts"
test "$(detect_language "$TMP_ROOT/kotlin")" = kotlin

mkdir -p "$TMP_ROOT/mono"
: > "$TMP_ROOT/mono/go.mod"
: > "$TMP_ROOT/mono/package.json"
: > "$TMP_ROOT/mono/service.go"
: > "$TMP_ROOT/mono/app.ts"
candidates="$(detect_languages "$TMP_ROOT/mono")"
grep -Fxq go <<<"$candidates"
grep -Fxq node <<<"$candidates"

grep -Fq 'all candidates' "$RULE"
grep -Fq 'multiple candidates' "$RULE"
grep -Fq 'Do not use first-match wins' "$DOC"

printf 'PASS: language routing distinguishes Kotlin and preserves monorepo candidates\n'
