#!/usr/bin/env bash
# docs-i18n-check.sh — 4-locale (ko/en/ja/zh) documentation parity validator.
#
# Source: SPEC-DOCS-SITE-001 AC-G3-03, CLAUDE.local.md §17.3 (Phase 5 deliverable).
#
# Validates:
#   1. File count/path parity across 4 locales
#   2. YAML frontmatter `title` presence (non-empty)
#   3. H1 heading existence in each .md file
#   4. MoAI glossary terms preserved verbatim across all locales
#
# Exits non-zero on any violation so CI can block merges.
#
# Usage:
#   scripts/docs-i18n-check.sh              # default strict mode
#   DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh   # warn-only mode (for bootstrap)

set -euo pipefail

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTENT_ROOT="${REPO_ROOT}/docs-site/content"

# Canonical source is ko. Other locales must mirror ko.
CANONICAL_LOCALE="ko"
TARGET_LOCALES=("en" "ja" "zh")
ALL_LOCALES=("${CANONICAL_LOCALE}" "${TARGET_LOCALES[@]}")

# MoAI canonical glossary (terms must appear verbatim in all locales; case-sensitive).
# Extend this list whenever a brand or product term enters the docs.
GLOSSARY_TERMS=(
  "MoAI-ADK"
  "SPEC-First"
  "EARS"
  "TRUST 5"
  "Claude Code"
  "Anthropic"
  "moai-adk"
)

# Strict mode: exit 1 on any error. Set DOCS_I18N_STRICT=0 to downgrade to warnings.
STRICT="${DOCS_I18N_STRICT:-1}"

# -----------------------------------------------------------------------------
# Output helpers
# -----------------------------------------------------------------------------
ERRORS=0
WARNINGS=0

err() {
  printf '❌ ERROR   %s\n' "$*" >&2
  ERRORS=$((ERRORS + 1))
}

warn() {
  printf '⚠️  WARN   %s\n' "$*" >&2
  WARNINGS=$((WARNINGS + 1))
}

info() {
  printf 'ℹ️  INFO   %s\n' "$*"
}

section() {
  printf '\n=== %s ===\n' "$*"
}

# -----------------------------------------------------------------------------
# Preflight: locale directories exist
# -----------------------------------------------------------------------------
section "Preflight: locale directory presence"

if [[ ! -d "${CONTENT_ROOT}" ]]; then
  err "content root not found: ${CONTENT_ROOT}"
  exit 1
fi

for loc in "${ALL_LOCALES[@]}"; do
  if [[ ! -d "${CONTENT_ROOT}/${loc}" ]]; then
    err "locale directory missing: content/${loc}"
  fi
done

if (( ERRORS > 0 )); then
  exit 1
fi

# -----------------------------------------------------------------------------
# Check 1: File path parity
# -----------------------------------------------------------------------------
section "Check 1: File path parity (${CANONICAL_LOCALE} ↔ target locales)"

TMPDIR="$(mktemp -d -t moai-docs-i18n.XXXXXX)"
trap 'rm -rf "${TMPDIR}"' EXIT

# Build relative-path file lists per locale.
for loc in "${ALL_LOCALES[@]}"; do
  ( cd "${CONTENT_ROOT}/${loc}" && find . -type f -name '*.md' | sed 's|^\./||' | sort ) > "${TMPDIR}/${loc}.txt"
done

CANONICAL_COUNT=$(wc -l < "${TMPDIR}/${CANONICAL_LOCALE}.txt" | tr -d '[:space:]')
info "${CANONICAL_LOCALE}: ${CANONICAL_COUNT} .md files (canonical)"

for loc in "${TARGET_LOCALES[@]}"; do
  loc_count=$(wc -l < "${TMPDIR}/${loc}.txt" | tr -d '[:space:]')
  info "${loc}: ${loc_count} .md files"

  # Files present in ko but missing in this locale
  missing="$(comm -23 "${TMPDIR}/${CANONICAL_LOCALE}.txt" "${TMPDIR}/${loc}.txt" || true)"
  if [[ -n "${missing}" ]]; then
    while IFS= read -r path; do
      [[ -z "${path}" ]] && continue
      err "missing translation: ${loc}/${path} (exists in ${CANONICAL_LOCALE})"
    done <<<"${missing}"
  fi

  # Files present in this locale but not in ko (orphan translations)
  orphan="$(comm -13 "${TMPDIR}/${CANONICAL_LOCALE}.txt" "${TMPDIR}/${loc}.txt" || true)"
  if [[ -n "${orphan}" ]]; then
    while IFS= read -r path; do
      [[ -z "${path}" ]] && continue
      warn "orphan translation: ${loc}/${path} has no ${CANONICAL_LOCALE} counterpart"
    done <<<"${orphan}"
  fi
done

# -----------------------------------------------------------------------------
# Check 2: YAML frontmatter `title` presence
# -----------------------------------------------------------------------------
section "Check 2: Frontmatter title presence"

check_frontmatter_title() {
  local file="$1"
  # Grab content between first two --- delimiters.
  local frontmatter
  frontmatter=$(awk '
    /^---[[:space:]]*$/ {
      fm_count++
      if (fm_count == 1) { in_fm = 1; next }
      if (fm_count == 2) { in_fm = 0; exit }
    }
    in_fm { print }
  ' "${file}")

  if [[ -z "${frontmatter}" ]]; then
    err "no frontmatter block in ${file#${REPO_ROOT}/}"
    return
  fi

  # Extract title value (supports "title: X" and `title: "X"`)
  local title
  title=$(printf '%s\n' "${frontmatter}" | sed -n 's/^title:[[:space:]]*//p' | head -n1)
  title="${title%\"}"
  title="${title#\"}"
  title="${title%\'}"
  title="${title#\'}"
  title="${title## }"
  title="${title%% }"

  if [[ -z "${title}" ]]; then
    err "empty or missing 'title:' in ${file#${REPO_ROOT}/}"
  fi
}

for loc in "${ALL_LOCALES[@]}"; do
  while IFS= read -r rel; do
    [[ -z "${rel}" ]] && continue
    check_frontmatter_title "${CONTENT_ROOT}/${loc}/${rel}"
  done < "${TMPDIR}/${loc}.txt"
done

# -----------------------------------------------------------------------------
# Check 3: H1 heading existence
# -----------------------------------------------------------------------------
section "Check 3: H1 heading existence"

check_h1() {
  local file="$1"
  # Skip frontmatter, then look for first non-empty H1.
  if ! awk '
    BEGIN { in_fm = 0; fm_done = 0 }
    /^---[[:space:]]*$/ {
      if (!fm_done) {
        in_fm = !in_fm
        if (!in_fm) fm_done = 1
        next
      }
    }
    in_fm { next }
    /^#[[:space:]]+.+/ { found = 1; exit }
    END { exit (found ? 0 : 1) }
  ' "${file}"; then
    err "no H1 heading in ${file#${REPO_ROOT}/}"
  fi
}

for loc in "${ALL_LOCALES[@]}"; do
  while IFS= read -r rel; do
    [[ -z "${rel}" ]] && continue
    # _index.md files can legitimately rely on title-only rendering; skip H1 check.
    [[ "$(basename "${rel}")" == "_index.md" ]] && continue
    check_h1 "${CONTENT_ROOT}/${loc}/${rel}"
  done < "${TMPDIR}/${loc}.txt"
done

# -----------------------------------------------------------------------------
# Check 4: Glossary term preservation
# -----------------------------------------------------------------------------
section "Check 4: Glossary term preservation (canonical: ${CANONICAL_LOCALE})"

# For each file in ko that contains a glossary term, every translated counterpart
# must also contain that term verbatim (case-sensitive).
for term in "${GLOSSARY_TERMS[@]}"; do
  # List ko files that contain the term.
  while IFS= read -r rel; do
    [[ -z "${rel}" ]] && continue
    ko_file="${CONTENT_ROOT}/${CANONICAL_LOCALE}/${rel}"
    if grep -Fq -- "${term}" "${ko_file}"; then
      for loc in "${TARGET_LOCALES[@]}"; do
        tr_file="${CONTENT_ROOT}/${loc}/${rel}"
        [[ ! -f "${tr_file}" ]] && continue  # missing file already reported in Check 1
        if ! grep -Fq -- "${term}" "${tr_file}"; then
          err "glossary term '${term}' missing in ${loc}/${rel} (present in ${CANONICAL_LOCALE})"
        fi
      done
    fi
  done < "${TMPDIR}/${CANONICAL_LOCALE}.txt"
done

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
section "Summary"
printf 'Errors:   %d\n' "${ERRORS}"
printf 'Warnings: %d\n' "${WARNINGS}"

if (( ERRORS > 0 )); then
  if [[ "${STRICT}" == "1" ]]; then
    printf '\nFAIL: %d error(s) detected. Fix them or re-run with DOCS_I18N_STRICT=0 for warn-only mode.\n' "${ERRORS}" >&2
    exit 1
  else
    printf '\nWARN: %d error(s) detected, but DOCS_I18N_STRICT=0 — exiting 0.\n' "${ERRORS}" >&2
    exit 0
  fi
fi

printf 'OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.\n'
exit 0
