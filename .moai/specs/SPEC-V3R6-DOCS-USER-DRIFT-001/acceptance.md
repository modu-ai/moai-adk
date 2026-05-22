# Acceptance Criteria — SPEC-V3R6-DOCS-USER-DRIFT-001

## 1. Acceptance Strategy

본 SPEC 은 docs-site 4-locale (ko/en/ja/zh) `workflow-commands/moai-sync.md` 페이지에 새 H2 섹션 1개를 삽입한다. 본 acceptance.md 는 **모든 AC 가 binary PASS/FAIL** 이며 각 AC 는 단일 shell 명령으로 검증 가능하다.

## 2. AC Matrix (Given-When-Then)

### AC-DUD-001: 새 H2 섹션이 4-locale 모두에 정확히 1개 존재

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료 직후 working tree.
**When**: 4-locale sync 페이지에서 새 H2 heading (locale-translated) 정확 매칭 카운트.
**Then**: 각 locale 파일에 정확히 1개 매칭. 합계 4건.

```bash
# Verification command (Given-When-Then atomic)
ko_count=$(grep -c "^## PR 머지 후 CI 모니터링" docs-site/content/ko/workflow-commands/moai-sync.md)
en_count=$(grep -c "^## CI monitoring after PR creation" docs-site/content/en/workflow-commands/moai-sync.md)
ja_count=$(grep -c "^## PR 作成後の CI モニタリング" docs-site/content/ja/workflow-commands/moai-sync.md)
zh_count=$(grep -c "^## PR 创建后的 CI 监控" docs-site/content/zh/workflow-commands/moai-sync.md)

# Expected: 1 1 1 1 (no more, no less)
[ "$ko_count" = "1" ] && [ "$en_count" = "1" ] && [ "$ja_count" = "1" ] && [ "$zh_count" = "1" ] || exit 1
```

**Maps to**: REQ-DUD-001, REQ-DUD-004.

### AC-DUD-002: PR #1045 파일 영역 무수정 (회귀 차단)

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료 직후 working tree.
**When**: `gh pr diff 1045 --name-only` 의 파일 list 와 본 SPEC 수정 파일 set 교집합 계산.
**Then**: 교집합 0건.

```bash
# Verification command
pr_files=$(gh pr diff 1045 --name-only | sort -u)
my_files=$(git diff origin/main..HEAD --name-only | sort -u)
overlap=$(comm -12 <(echo "$pr_files") <(echo "$my_files") | wc -l | tr -d ' ')

# Expected: 0
[ "$overlap" = "0" ] || exit 1
```

**Maps to**: REQ-DUD-006.

### AC-DUD-003: 새 섹션 본문이 doctrine 3 pillar 모두 포함

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료 후 ko locale 새 섹션 본문 추출.
**When**: 새 섹션 (다음 `## ` heading 직전까지 범위) 본문에서 doctrine 키워드 검색.
**Then**: 3 pillar 키워드 모두 매칭 (Wave 1 폴링 / 30초 / 30분, Wave 2 / 3 iter / 자동 fix, escalation / 4 iter / AskUserQuestion).

```bash
# Extract new section body from ko locale (between target H2 and next ##)
section=$(awk '/^## PR 머지 후 CI 모니터링/{flag=1; next} /^## /{flag=0} flag' docs-site/content/ko/workflow-commands/moai-sync.md)

# Pillar 1: Wave 1 polling (30s + 30min)
echo "$section" | grep -qE "30[ ]*(초|s|sec)" || exit 1
echo "$section" | grep -qE "30[ ]*분|30[ ]*min|30 minutes" || exit 1

# Pillar 2: Wave 2 auto-fix loop max 3 iterations
echo "$section" | grep -qE "3[ ]*(회|iter|iteration)" || exit 1
echo "$section" | grep -qE "auto-?fix|자동[ ]*fix|자동[ ]*수정" || exit 1

# Pillar 3: Escalation at iter >= 4 via AskUserQuestion blocking
echo "$section" | grep -qE "4(회|iter)|iteration.*4|escalation|에스컬레이션" || exit 1
echo "$section" | grep -q "AskUserQuestion" || exit 1
```

**Maps to**: REQ-DUD-002.

### AC-DUD-004: protected files subsection 또는 callout 존재

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료 후 ko locale 새 섹션 본문 추출.
**When**: 본문에서 보호 파일 5종 키워드 검색.
**Then**: `.env`, credentials, `scripts/ci-watch/run.sh`, `.github/required-checks.yml` 4개 키워드 모두 매칭 (4-locale 모두).

```bash
# Verification command (ko locale; en/ja/zh equivalent assumed via REQ-DUD-004 parity)
section=$(awk '/^## PR 머지 후 CI 모니터링/{flag=1; next} /^## /{flag=0} flag' docs-site/content/ko/workflow-commands/moai-sync.md)

echo "$section" | grep -q '\.env' || exit 1
echo "$section" | grep -qE "credential|크리덴셜|认证|認証情報" || exit 1
echo "$section" | grep -q "scripts/ci-watch/run.sh" || exit 1
echo "$section" | grep -q ".github/required-checks.yml" || exit 1
```

**Maps to**: REQ-DUD-003.

### AC-DUD-005: 4-locale 본문 구조 parity (h3 + 표 + callout 동일)

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료.
**When**: 4-locale 각 sync 페이지에서 새 H2 섹션 범위 내 (a) h3 subsection 개수 + (b) markdown 표 행 개수 + (c) shortcode callout 개수 추출.
**Then**: 4-locale 세 값 모두 동일.

```bash
# Verification command
extract_section() {
  local locale=$1
  local heading=$2
  awk -v h="^## ${heading}" '$0 ~ h {flag=1; next} /^## /{flag=0} flag' \
    docs-site/content/${locale}/workflow-commands/moai-sync.md
}

for l_h in "ko:PR 머지 후 CI 모니터링" "en:CI monitoring after PR creation" "ja:PR 作成後の CI モニタリング" "zh:PR 创建后的 CI 监控"; do
  locale="${l_h%%:*}"
  heading="${l_h#*:}"
  section=$(extract_section "$locale" "$heading")
  h3_count=$(echo "$section" | grep -c '^### ')
  table_rows=$(echo "$section" | grep -c '^|')
  callout_count=$(echo "$section" | grep -c '{{< callout')
  echo "$locale: h3=$h3_count table=$table_rows callout=$callout_count"
done | awk '{
  gsub(/[a-z]+:/, "", $1); gsub(/[a-z0-9]+=/, "", $2); gsub(/[a-z0-9]+=/, "", $3); gsub(/[a-z0-9]+=/, "", $4);
  key = $2 "_" $3 "_" $4;
  seen[key]++
} END {
  if (length(seen) != 1) { print "PARITY MISMATCH:"; for (k in seen) print "  " k " x " seen[k]; exit 1 }
}'
```

**Maps to**: REQ-DUD-004.

### AC-DUD-006: 금지 콘텐츠 0건 (forbidden URL, Mermaid LR, emoji, dev-only refs)

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료.
**When**: 4-locale 새 섹션 본문에서 금지 패턴 검색.
**Then**: 모든 금지 패턴 0건 매칭.

```bash
# Verification command — applied to all 4 locales
forbidden_failed=0
for l_h in "ko:PR 머지 후 CI 모니터링" "en:CI monitoring after PR creation" "ja:PR 作成後の CI モニタリング" "zh:PR 创建后的 CI 监控"; do
  locale="${l_h%%:*}"
  heading="${l_h#*:}"
  section=$(awk -v h="^## ${heading}" '$0 ~ h {flag=1; next} /^## /{flag=0} flag' \
    docs-site/content/${locale}/workflow-commands/moai-sync.md)

  # (a) Forbidden URLs (only canonical adk.mo.ai.kr + github.com/modu-ai allowed)
  if echo "$section" | grep -qE 'docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr'; then
    echo "FAIL $locale: forbidden URL"; forbidden_failed=1
  fi

  # (b) Mermaid LR / BT / RL (only TD/TB allowed)
  if echo "$section" | grep -qE 'flowchart[ ]+(LR|BT|RL)|graph[ ]+(LR|BT|RL)'; then
    echo "FAIL $locale: Mermaid non-TD direction"; forbidden_failed=1
  fi

  # (c) Emojis (Unicode block U+1F300..U+1FAFF + heart/etc.)
  if echo "$section" | LC_ALL=C grep -qP '[\xF0\x9F]'; then
    echo "FAIL $locale: emoji detected"; forbidden_failed=1
  fi

  # (d) Dev-only command refs (97-*, 98-github, 99-*, /98-github)
  if echo "$section" | grep -qE '/97-|/98-github|/99-|97-release-update|release-update\.md'; then
    echo "FAIL $locale: dev-only reference"; forbidden_failed=1
  fi
done

[ "$forbidden_failed" = "0" ] || exit 1
```

**Maps to**: REQ-DUD-005.

### AC-DUD-007: Hugo 빌드 exit 0 + 4-locale public/ 페이지 모두 새 섹션 포함

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료 직전, working tree clean (untracked 무관).
**When**: `cd docs-site && hugo --gc --minify` 실행 + `public/{ko,en,ja,zh}/workflow-commands/moai-sync/index.html` 4 파일 새 H2 heading 텍스트 매칭.
**Then**: hugo exit code 0, 4 페이지 모두 새 heading 텍스트 매칭 1건 이상.

```bash
# Verification command
cd docs-site && hugo --gc --minify
hugo_exit=$?
[ "$hugo_exit" = "0" ] || exit 1

# Check 4-locale public pages contain new section heading
ko_ok=$(grep -l "PR 머지 후 CI 모니터링" public/ko/workflow-commands/moai-sync/index.html 2>/dev/null | wc -l | tr -d ' ')
en_ok=$(grep -l "CI monitoring after PR creation" public/en/workflow-commands/moai-sync/index.html 2>/dev/null | wc -l | tr -d ' ')
ja_ok=$(grep -l "PR 作成後の CI モニタリング" public/ja/workflow-commands/moai-sync/index.html 2>/dev/null | wc -l | tr -d ' ')
zh_ok=$(grep -l "PR 创建后的 CI 监控" public/zh/workflow-commands/moai-sync/index.html 2>/dev/null | wc -l | tr -d ' ')

[ "$ko_ok" = "1" ] && [ "$en_ok" = "1" ] && [ "$ja_ok" = "1" ] && [ "$zh_ok" = "1" ] || exit 1
```

**Maps to**: REQ-DUD-007.

### AC-DUD-008: 4-locale 본문 라인 수 편차 ±20% 이하 (asymmetry guard)

**Given**: SPEC-V3R6-DOCS-USER-DRIFT-001 run-phase 완료.
**When**: 새 섹션 라인 수 4-locale 추출.
**Then**: max(line_count) / min(line_count) ≤ 1.2.

```bash
# Verification command
counts=()
for l_h in "ko:PR 머지 후 CI 모니터링" "en:CI monitoring after PR creation" "ja:PR 作成後の CI モニタリング" "zh:PR 创建后的 CI 监控"; do
  locale="${l_h%%:*}"
  heading="${l_h#*:}"
  c=$(awk -v h="^## ${heading}" '$0 ~ h {flag=1; next} /^## /{flag=0} flag' \
    docs-site/content/${locale}/workflow-commands/moai-sync.md | wc -l | tr -d ' ')
  counts+=("$c")
done

max=$(printf '%s\n' "${counts[@]}" | sort -n | tail -1)
min=$(printf '%s\n' "${counts[@]}" | sort -n | head -1)
ratio=$(awk -v a="$max" -v b="$min" 'BEGIN{printf "%.3f", a/b}')

# Expected: ratio <= 1.20
awk -v r="$ratio" 'BEGIN{exit (r > 1.20) ? 1 : 0}'
```

**Maps to**: REQ-DUD-004 (parity edge case EC-DUD-003).

## 3. Traceability Matrix

| REQ | AC | Verification |
|-----|----|--------------|
| REQ-DUD-001 (new H2 inserted in 4-locale) | AC-DUD-001 | grep heading count = 1 per locale |
| REQ-DUD-002 (doctrine 3 pillars) | AC-DUD-003 | grep pillar keywords (Wave1/Wave2/escalation) |
| REQ-DUD-003 (protected files subsection) | AC-DUD-004 | grep .env / credentials / scripts/ci-watch/run.sh / required-checks.yml |
| REQ-DUD-004 (4-locale parity) | AC-DUD-001 + AC-DUD-005 + AC-DUD-008 | heading count parity + h3/table/callout parity + line count ±20% |
| REQ-DUD-005 (no forbidden content) | AC-DUD-006 | grep forbidden URL/Mermaid LR/emoji/dev-only refs (all 0) |
| REQ-DUD-006 (no PR #1045 overlap) | AC-DUD-002 | `gh pr diff 1045 --name-only` intersection = 0 |
| REQ-DUD-007 (Hugo build clean) | AC-DUD-007 | hugo exit 0 + 4 public/ pages contain heading |

100% REQ↔AC traceability verified — 7 REQs all mapped to at least 1 AC; 8 ACs all mapped to at least 1 REQ.

## 4. Edge Case Test Cases

| Edge Case | Test scenario | Expected | Maps to AC |
|-----------|---------------|----------|------------|
| EC-DUD-001 | PR #1045 머지 후 본 SPEC run-phase 진입 | 본 SPEC 의 변경 파일 set 은 PR #1045 의 변경 파일 set 과 disjoint | AC-DUD-002 |
| EC-DUD-002 | PR #1045 가 본 SPEC 보다 늦게 머지 | merge conflict 0건 (file overlap 사전 검증) | AC-DUD-002 |
| EC-DUD-003 | locale별 본문 라인 수 비대칭 (예: ja > ko 약 30%) | max/min ≤ 1.20 허용 | AC-DUD-008 |
| EC-DUD-004 | en/ja/zh locale 에 `## 품질 게이트` heading 부재 | run-phase pre-flight 의무 — fallback heading 4-locale 동일 사용 | (run-phase 별도 pre-flight 보고서) |
| EC-DUD-005 | Hugo 빌드 weight 충돌 fail | PR #1045 책임 — 본 SPEC blocker 보고 | AC-DUD-007 (만약 fail 시 AC FAIL) |

## 5. Definition of Done (DoD)

- [ ] AC-DUD-001 PASS — 4-locale 새 H2 heading 정확 1건씩
- [ ] AC-DUD-002 PASS — PR #1045 overlap 0건
- [ ] AC-DUD-003 PASS — doctrine 3 pillars 본문 포함
- [ ] AC-DUD-004 PASS — protected files 5종 키워드 매칭
- [ ] AC-DUD-005 PASS — 4-locale 구조 parity (h3+table+callout)
- [ ] AC-DUD-006 PASS — 금지 콘텐츠 0건
- [ ] AC-DUD-007 PASS — Hugo 빌드 exit 0 + public/ 4 페이지 새 heading 포함
- [ ] AC-DUD-008 PASS — 4-locale 본문 라인 수 ±20% 이내
- [ ] Working tree PRESERVE list (spec.md §6.2) 무결성 확인 — dirty/untracked PRESERVE 항목 수정 0건
- [ ] git diff 가 `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` 4 파일만 + SPEC 산출물 (`spec.md`/`acceptance.md`/`plan.md`/`progress.md`) 만 표시
- [ ] SPEC frontmatter status `draft → implemented`, version `0.1.0 → 0.2.0`, HISTORY 항목 추가

## 6. Test Execution Strategy

run-phase 완료 직전 단일 검증 스크립트 실행 — 모든 AC 를 sequential 로 실행 후 결과 집계:

```bash
# acceptance-verify.sh (run-phase 끝에서 실행)
set -e
echo "[AC-DUD-001] new H2 in 4-locale ..." && (검증 명령 §2 참조)
echo "[AC-DUD-002] no PR #1045 overlap ..." && (검증 명령 §2 참조)
echo "[AC-DUD-003] doctrine 3 pillars ..." && (검증 명령 §2 참조)
echo "[AC-DUD-004] protected files ..." && (검증 명령 §2 참조)
echo "[AC-DUD-005] 4-locale parity ..." && (검증 명령 §2 참조)
echo "[AC-DUD-006] no forbidden content ..." && (검증 명령 §2 참조)
echo "[AC-DUD-007] hugo build clean ..." && (검증 명령 §2 참조)
echo "[AC-DUD-008] line-count ±20% ..." && (검증 명령 §2 참조)
echo "ALL 8 ACs PASS"
```

각 AC 가 binary PASS/FAIL — 1개 fail = SPEC run-phase FAIL.
