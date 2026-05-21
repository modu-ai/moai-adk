#!/usr/bin/env bash
# migrate-frontmatter.sh — Hextra → hugo-geekdoc frontmatter 마이그레이션
#
# 제거 대상 (Hextra 전용 키):
#   - type: docs
#   - sidebar: 블록 (open: true 포함)
#   - toc: true
#
# 보존: title, description, weight, draft, tags, date 등 표준 Hugo 키
#
# Usage:
#   ./scripts/migrate-frontmatter.sh --dry-run   # 변경 사항 미리보기
#   ./scripts/migrate-frontmatter.sh             # 실제 적용
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCS_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONTENT_DIR="$DOCS_ROOT/content"

DRY_RUN=0
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=1
  echo "[DRY-RUN] No files will be modified."
fi

# 1) Hextra 키 (type/sidebar/toc) 가진 파일 찾기
FILES=$(grep -rl --include='*.md' -E '^(type:|sidebar:|toc:)' "$CONTENT_DIR" 2>/dev/null || true)

if [[ -z "$FILES" ]]; then
  echo "No files with Hextra-specific frontmatter found. Nothing to migrate."
  exit 0
fi

echo "Found $(echo "$FILES" | wc -l | tr -d ' ') file(s) to migrate:"
echo "$FILES" | sed 's|^|  - |'
echo ""

migrate_file() {
  local file="$1"
  local tmp
  tmp=$(mktemp)

  # frontmatter 블록 (--- 사이) 안에서 type/sidebar/toc 제거
  # awk로 frontmatter 영역 추적 + 라인 필터링
  awk '
    BEGIN { in_fm = 0; fm_count = 0; skip_block = 0 }
    /^---[[:space:]]*$/ {
      fm_count++
      if (fm_count == 1) { in_fm = 1 }
      else if (fm_count == 2) { in_fm = 0 }
      print; next
    }
    {
      if (in_fm) {
        # skip_block 모드: 들여쓰기 라인은 모두 skip, 들여쓰기 끝나면 모드 해제
        if (skip_block) {
          if ($0 ~ /^[[:space:]]+/) { next }
          else { skip_block = 0 }
        }
        # type: docs 한 줄 제거
        if ($0 ~ /^type:[[:space:]]/) { next }
        # toc: ... 한 줄 제거
        if ($0 ~ /^toc:[[:space:]]/) { next }
        # sidebar: 블록 시작 — 다음 들여쓰기 라인까지 skip
        if ($0 ~ /^sidebar:[[:space:]]*$/) { skip_block = 1; next }
        # sidebar: {} 인라인 — 한 줄만 제거
        if ($0 ~ /^sidebar:[[:space:]]/) { next }
      }
      print
    }
  ' "$file" > "$tmp"

  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "--- DRY-RUN diff for $file ---"
    diff "$file" "$tmp" || true
  else
    if ! cmp -s "$file" "$tmp"; then
      cp "$tmp" "$file"
      echo "  ✓ migrated: $file"
    else
      echo "  - skipped (no change): $file"
    fi
  fi
  rm -f "$tmp"
}

while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  migrate_file "$f"
done <<< "$FILES"

echo ""
echo "Migration complete."
