#!/bin/bash
# t357: D1 대상(비-spec.md 산출물의 frontmatter status: 라인) 모집단별 계수
# closed-only vs 전체 696 — 두 모집단의 차이를 드러낸다
cd "$1" || exit 1
all=0; closed=0
fm_of() { awk 'NR==1&&/^---/{f=1;next} f&&/^---/{exit} f' "$1"; }
for d in .moai/specs/SPEC-*/; do
  spec="${d}spec.md"
  st=""
  [ -f "$spec" ] && st=$(fm_of "$spec" | sed -n 's/^status:[[:space:]]*//p' | head -1 | tr -d '"' | tr -d "'" | tr -d '\r')
  for a in plan acceptance design research; do
    f="${d}${a}.md"
    [ -f "$f" ] || continue
    if fm_of "$f" | grep -qE '^status:[[:space:]]'; then
      all=$((all+1))
      case "$st" in completed|implemented) closed=$((closed+1)) ;; esac
    fi
  done
done
echo "HEAD=$(git rev-parse --short HEAD)"
echo "D1 전체 696 모집단 = $all"
echo "D1 종결(633) 모집단 = $closed"
