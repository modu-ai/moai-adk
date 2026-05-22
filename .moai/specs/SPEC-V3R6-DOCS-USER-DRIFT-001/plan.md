# Plan — SPEC-V3R6-DOCS-USER-DRIFT-001

## 1. Scope Recap

본 SPEC 은 `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` 4 파일에 **새 H2 섹션 1개** (PR 머지 후 CI 모니터링 doctrine — Wave 1 ci-watch + Wave 2 ci-autofix) 를 삽입한다. 단일 concern, Tier S, manager-spec 자체 판단으로 plan-auditor 호출 optional (LEAN Tier S precedent: SPEC-V3R6-CATALOG-SSOT-001 / SPEC-V3R6-HARNESS-RENAME-001 1-pass).

baseline `.moai/research/moai-adk-current-state-2026-05-22.md` §8 F1-F8 중 F1/F2/F3/F4 는 선행 PR #1045 (`SPEC-V3R6-DOCS-CMD-CATALOG-001`, OPEN MERGEABLE) 가 완전 해소. F6 은 commit `a3239d3de` 정리 완료. F5/F7 은 별도 SPEC 후보. **본 SPEC = F8 only**.

## 2. Implementation Approach (Section A — Context + Approach)

### 2.1 Approach (high-level)

1. **Pre-flight check**: 4-locale `moai-sync.md` 의 H2 list 추출, `## 품질 게이트` (또는 locale equivalent) heading 존재 확인. PR #1045 머지 상태 확인 (`gh pr view 1045 --json state`).
2. **ko 원문 작성**: 새 섹션 본문 (3 pillars + protected files subsection + 자동/사람 결정 분류 표 + 2 cross-ref 링크) 을 한국어로 작성.
3. **en/ja/zh 번역**: ko 원문을 기준으로 3개 locale 동시 번역. h3 subsection 개수 + 표 행 수 + callout 수 정확히 동일 (REQ-DUD-004 parity).
4. **삽입**: 4 파일 Edit tool 로 `## 품질 게이트` (또는 locale equivalent) 직전에 새 H2 섹션 삽입. MultiEdit 사용 가능 (4 파일 simultaneous).
5. **Hugo 로컬 빌드 검증**: `cd docs-site && hugo --gc --minify` exit 0. public/{ko,en,ja,zh}/workflow-commands/moai-sync/index.html 4 페이지 새 heading 포함 확인.
6. **AC 검증 실행**: acceptance.md §2 AC-DUD-001..008 8건 sequential 실행. 1개 fail = SPEC run-phase FAIL.
7. **PRESERVE 검증**: §6.2 PRESERVE list 무결성 확인 — dirty/untracked 17개 항목 수정 0건. `git diff` 가 4 sync.md 파일 + SPEC 산출물만 표시.

### 2.2 새 섹션 ko 원문 outline (run-phase 작성 sketch — non-binding)

```markdown
## PR 머지 후 CI 모니터링

`/moai sync` 가 PR 을 생성한 직후, MoAI-ADK 는 두 단계의 자동 모니터링을
실행합니다.

### Wave 1 — CI 결과 폴링

- 30초 간격으로 `gh pr checks` 호출 (GitHub API rate limit 존중)
- 30분 hard timeout — 그 시간 안에 required check 가 완료되지 않으면
  watch loop 가 exit code 3 으로 종료
- required check 정의 SSoT: `.github/required-checks.yml`
- auxiliary check 는 fail 해도 merge blocker 가 아님 (warning 만)

### Wave 2 — 자동 fix 루프 (최대 3 회)

required check 가 fail 하면 MoAI-ADK 는 자동 fix 루프에 진입합니다.

- 매 iteration 마다 새 commit 으로 fix 적용 (force-push / amend 금지)
- 최대 3 iterations per PR push (per-session 이 아님)
- iteration ≥ 4 가 되면 사용자에게 blocking AskUserQuestion 으로 escalation

### 자동 처리 vs 사람 결정 필요

| 결함 유형 | 자동 처리? | 비고 |
|-----------|-----------|------|
| lint error | 자동 | `golangci-lint` autofix 가능한 항목 |
| format drift | 자동 | `gofmt` / `prettier` 등 |
| test syntax error | 자동 | import 누락 / 컴파일 에러 |
| **data race** | **사람 결정** | semantic failure — 의도된 동시성? |
| **deadlock** | **사람 결정** | semantic failure |
| **panic** | **사람 결정** | semantic failure |
| **test assertion failure** | **사람 결정** | spec 또는 코드 둘 중 무엇이 옳은지 사람 판단 |

### auto-fix 가 절대 건드리지 않는 파일

{{< callout type="warning" >}}
auto-fix 루프는 다음 파일을 **절대 수정하지 않습니다**:

- `.env`, `.env.*` (환경변수 / 비밀)
- credentials 파일
- `scripts/ci-watch/run.sh` (Wave 2 infrastructure)
- `.github/required-checks.yml` (Wave 1 SSoT)
{{< /callout >}}

### 관련 문서

- 폴링 doctrine SSoT: `.claude/rules/moai/workflow/ci-watch-protocol.md`
- auto-fix doctrine SSoT: `.claude/rules/moai/workflow/ci-autofix-protocol.md`
```

위 outline 은 run-phase 의 starting point — 실제 본문은 manager-develop 위임 시 locale-appropriate phrasing 으로 다듬는다.

### 2.3 4-locale 번역 매핑

| Element | ko | en | ja | zh |
|---------|----|----|----|----|
| H2 heading | `PR 머지 후 CI 모니터링` | `CI monitoring after PR creation` | `PR 作成後の CI モニタリング` | `PR 创建后的 CI 监控` |
| H3 #1 | `Wave 1 — CI 결과 폴링` | `Wave 1 — CI result polling` | `Wave 1 — CI 結果ポーリング` | `Wave 1 — CI 结果轮询` |
| H3 #2 | `Wave 2 — 자동 fix 루프 (최대 3 회)` | `Wave 2 — Auto-fix loop (max 3 iterations)` | `Wave 2 — 自動 fix ループ (最大 3 回)` | `Wave 2 — 自动修复循环 (最多 3 次)` |
| H3 #3 | `자동 처리 vs 사람 결정 필요` | `Auto-handled vs human decision required` | `自動処理 vs 人間判断必要` | `自动处理 vs 需人工决定` |
| H3 #4 | `auto-fix 가 절대 건드리지 않는 파일` | `Files auto-fix never modifies` | `auto-fix が絶対に変更しないファイル` | `auto-fix 绝不修改的文件` |
| H3 #5 | `관련 문서` | `Related documentation` | `関連ドキュメント` | `相关文档` |

5 h3 subsections × 4 locales = 20 h3 headings total. 1 markdown 표 (7 행 header + 데이터) × 4 locales. 1 warning callout × 4 locales. 라인 수 baseline (ko) 약 35-45 라인.

## 3. Section B — Known Issues (자동 주입)

본 SPEC 은 docs-site Hugo 콘텐츠만 다루므로 standard B1-B8 카테고리의 적용성을 다음과 같이 필터링:

| Cat | 적용 | 적용 시 처리 |
|-----|------|--------------|
| B1 Cross-platform build tags | 적용 안 됨 (Go 소스 변경 0) | n/a |
| B2 Cross-SPEC 정책 충돌 | **적용** | PR #1045 (SPEC-V3R6-DOCS-CMD-CATALOG-001) 의 4 파일과 file overlap 0건 사전 검증 (AC-DUD-002). 본 SPEC 의 1 파일 set 은 PR #1045 의 16 파일 set 과 disjoint. |
| B3 C-HRA-008 Subagent Boundary | 적용 안 됨 (`internal/harness/`, `internal/hook/` 무관) | n/a |
| B4 Frontmatter Canonical Schema | **적용** | SPEC 산출물 frontmatter 12-field SSOT (`.claude/rules/moai/development/spec-frontmatter-schema.md`) 준수 — `created:` (not `created_at`) / `updated:` (not `updated_at`) / `tags:` string (not `labels:` array) / 12 필드 모두 존재. spec.md 의 `tier: S` 명시. |
| B5 CI 3-tier 인지 | **적용** | spec-lint + golangci-lint + Test (per OS) 별도 fail 가능. 본 SPEC 은 Go 소스 변경 0 이므로 golangci-lint 영향 최소 (pre-existing baseline residual 만 가능). spec-lint 는 frontmatter SSOT 준수 + Out of Scope h3 subsection 패턴으로 PASS 예상. |
| B6 spec-lint Heading 규약 | **적용** | spec.md §3 `## Scope` (h2) 하위에 `### 3.1 In Scope` + `### 3.2.1 Out of Scope — F1-F4` 등 **h3 sub-section** 으로 Out of Scope 표기 (h2 `## Out of Scope` 단독 사용 시 `MissingExclusions` ERROR — W3 발견). 본 SPEC 의 `## 3. Scope` 하위에 `### 3.2 Out of Scope (별도 SPEC 후보)` 가 존재하므로 spec-lint PASS 예상. |
| B7 observer.go capture path | 적용 안 됨 (hook 작업 0) | n/a |
| B8 Working Tree Hygiene | **적용** | §6.2 PRESERVE list 명시 — `.moai/harness/usage-log.jsonl` / `docs-site/{hugo.toml, layouts/_default/baseof.html, layouts/partials/menu.html}` modified, `docs-site/content/*/book/`, `docs-site/scripts/gen_menu.py`, `docs-site/static/book/`, `docs-site/data/menu/extra.yaml`, `docs-site/layouts/_default/redirect.html`, `internal/hook/.moai/`, `.moai/research/*-2026-05-22.md` untracked 모두 PRESERVE. `git add` 는 specific path 만 (`git add .moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/ docs-site/content/ko/workflow-commands/moai-sync.md docs-site/content/en/workflow-commands/moai-sync.md docs-site/content/ja/workflow-commands/moai-sync.md docs-site/content/zh/workflow-commands/moai-sync.md`). |

## 4. Section C — Pre-flight Check List

run-phase 시작 시 실행 의무:

```bash
# 1. 현재 branch + baseline 확인
git branch --show-current                  # 예상: feat/SPEC-V3R6-DOCS-USER-DRIFT-001 또는 main
git rev-parse HEAD                         # baseline commit SHA 기록

# 2. PR #1045 머지 상태 확인 (EC-DUD-001/002 시나리오 결정)
gh pr view 1045 --json state,mergeable     # state: OPEN | MERGED

# 3. 4-locale sync.md 라인 수 + H2 구조 baseline
wc -l docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md
for l in ko en ja zh; do
  echo "--- $l H2 list ---"
  grep '^## ' docs-site/content/$l/workflow-commands/moai-sync.md
done

# 4. F8 reproducibility 재확인 (sync 페이지에 ci-watch/ci-autofix 키워드 0건)
grep -l 'ci-watch\|ci-autofix\|auto-fix\|CI 모니터' \
  docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md || echo "F8 reproducible (no matches)"

# 5. PRESERVE list 보호 baseline 측정
git status --porcelain | head -30          # PRESERVE 대상 파일 list 기록

# 6. PR #1045 file overlap pre-check
gh pr diff 1045 --name-only | grep moai-sync && echo "OVERLAP DETECTED — STOP" || echo "no overlap"

# 7. Hugo 로컬 빌드 가능성 사전 확인
cd docs-site && hugo --gc --minify --buildDrafts=false 2>&1 | tail -3
cd ..
```

위 7개 명령 모두 정상 출력 + PR #1045 overlap 0건 + F8 reproducibility 재확인 후 implementation 진입.

## 5. Section D — Constraints (DO NOT VIOLATE)

### 5.1 PRESERVE 대상 (절대 수정 금지) — Section A enumeration

spec.md §6.2 의 PRESERVE list 모두. 핵심:

- `.moai/harness/usage-log.jsonl`
- `docs-site/{hugo.toml, layouts/_default/baseof.html, layouts/partials/menu.html}` (modified)
- `docs-site/content/{ko,en,ja,zh}/book/` (untracked 별도 워크스트림)
- `docs-site/scripts/gen_menu.py`, `docs-site/static/book/`, `docs-site/data/menu/extra.yaml`, `docs-site/layouts/_default/redirect.html` (untracked)
- `internal/hook/.moai/` (CLAUDE.local.md §2 working tree hygiene)
- `.moai/research/{config-audit, lsp-yaml-v2-audit, moai-adk-current-state, v3.0-design}-*.md` (untracked, baseline 보고서)
- **PR #1045 영역** (REQ-DUD-006): `moai-db.md` / `moai-github.md` / `moai-harness.md` / `moai-gate.md` × 4-locale + `_meta.yaml` × 12 + `main.yaml`

### 5.2 금지 명령

- `--no-verify` / `--amend` / force-push to main (CLAUDE.local.md §23)
- `git reset --hard` (sandbox 차단; `--keep` 대체 §23.5)
- `git add -A` 또는 `git add .` (specific path 만 — B8 Working Tree Hygiene)

### 5.3 사용 의무 명령

- Conventional Commits 형식 (`docs(SPEC-V3R6-DOCS-USER-DRIFT-001): ...`)
- `🗿 MoAI <email@mo.ai.kr>` trailer
- 단일 commit 또는 logical-grouping 으로 commit (Tier S 1-2 commits 권장)

### 5.4 Binary Constraints

- AC-DUD-001 grep heading count = 1 per locale (4 locales)
- AC-DUD-002 PR #1045 file overlap = 0
- AC-DUD-006 forbidden URL / Mermaid LR / emoji / dev-only refs = 0 매칭
- AC-DUD-007 Hugo 빌드 exit code = 0

## 6. Section E — Self-Verification Deliverables (run-phase 완료 보고 시 의무)

manager-develop (또는 orchestrator-direct 실행 시 orchestrator) 가 보고할 항목:

### E1. AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-DUD-001 | (PASS/FAIL) | acceptance.md §2 grep heading | (실제 출력) |
| AC-DUD-002 | (PASS/FAIL) | acceptance.md §2 gh pr diff overlap | (실제 출력) |
| AC-DUD-003 | (PASS/FAIL) | acceptance.md §2 pillar keywords | (실제 출력) |
| AC-DUD-004 | (PASS/FAIL) | acceptance.md §2 protected files keywords | (실제 출력) |
| AC-DUD-005 | (PASS/FAIL) | acceptance.md §2 parity awk | (실제 출력) |
| AC-DUD-006 | (PASS/FAIL) | acceptance.md §2 forbidden 4 patterns | (실제 출력) |
| AC-DUD-007 | (PASS/FAIL) | acceptance.md §2 hugo + public/ check | (실제 출력) |
| AC-DUD-008 | (PASS/FAIL) | acceptance.md §2 line ratio ≤ 1.20 | (실제 출력) |

### E2. Cross-Platform 결과

본 SPEC 은 Go 소스 변경 0 → cross-platform build n/a.

### E3. Coverage 측정

본 SPEC 은 Go 코드 0 → coverage n/a. spec-lint 와 hugo 빌드 만이 quality gate.

### E4. Subagent Boundary Grep (C-HRA-008)

본 SPEC 영역 `docs-site/` 는 subagent boundary 와 무관 → n/a.

### E5. Lint Status (NEW vs baseline)

- `golangci-lint`: pre-existing baseline residual 만 (본 SPEC 변경 영역 무관)
- `spec-lint`: 본 SPEC 산출물 (`spec.md`/`acceptance.md`/`plan.md`) PASS — frontmatter 12-field SSOT + Out of Scope h3 sub-section + EARS REQ 패턴
- Hugo 빌드: AC-DUD-007 검증 결과

### E6. Branch HEAD + Push 상태

- 새 commits SHA 리스트
- `git push origin <branch>` 결과
- PR 생성 시 PR URL

### E7. Blocker Report (있을 시)

위임 prompt 에서 명시 안 된 사용자 결정 필요 항목 발견 시 structured 보고 (AskUserQuestion 절대 호출 금지).

## 7. Milestones (Priority-based, no time estimates)

| Milestone | Priority | Verification |
|-----------|----------|--------------|
| M1: Pre-flight check 완료 | High | §4 pre-flight 7 명령 모두 정상 + PR #1045 state 기록 |
| M2: 4-locale 새 섹션 작성 + 삽입 | High | git diff 가 4 moai-sync.md 만 수정 + 새 H2 1개씩 |
| M3: Hugo 로컬 빌드 통과 | High | AC-DUD-007 PASS |
| M4: 8 ACs 모두 PASS | High | acceptance.md §2 검증 명령 8건 sequential 실행 |
| M5: PRESERVE 무결성 + commit + push | Medium | git status PRESERVE 항목 무수정 + commit message Conventional + 🗿 trailer |

M1 → M2 → M3 → M4 → M5 sequential. M1 fail (예: PR #1045 file overlap 발견) 시 즉시 STOP + blocker 보고.

## 8. Risks (상세, spec.md §7 요약 보완)

### R-DUD-001 — 4-locale 동시 업데이트 중 일부 locale 누락

- Severity: **High** (HARD 위반 시 docs-site-i18n-rules §17.3)
- Likelihood: Low (MultiEdit 단일 작업으로 4 파일 동시 처리)
- Mitigation:
  - manager-develop 위임 prompt 에 "4-locale 모두 동일 PR" 명시
  - AC-DUD-001 (4-locale 정확 1건씩) + AC-DUD-005 (parity) + AC-DUD-008 (line ratio ≤ 1.20) 3중 검증
  - git diff name-only 가 4 파일 모두 표시 확인
- Contingency: 일부 locale 만 수정된 상태로 commit 시도 시 pre-commit hook 차단 (장기 — 현재는 manual check)

### R-DUD-002 — Hugo 빌드 fail

- Severity: **Medium**
- Likelihood: Low (본 SPEC 은 frontmatter 변경 없음 + 새 본문은 standard markdown + Hugo callout shortcode 만 사용)
- Mitigation:
  - run-phase pre-flight (§4 step 7) 으로 baseline hugo 빌드 정상 사전 확인
  - 새 섹션 본문에 unsupported shortcode 사용 금지 (`{{< callout >}}` 만 사용, 기존 sync.md 패턴 그대로)
  - AC-DUD-007 (hugo --gc --minify exit 0) 완료 직전 검증
- Contingency: hugo fail 시 즉시 stash + 본문 revise + 재시도 (최대 3회)

### R-DUD-003 — PR #1045 와 file 충돌

- Severity: **Low** (file overlap pre-flight 검증 = 0 confirmed)
- Likelihood: Very Low
- Mitigation:
  - pre-flight `gh pr diff 1045 --name-only | grep moai-sync` (empty 확인)
  - AC-DUD-002 (comm -12 overlap = 0) 검증
- Contingency: 만약 overlap 발견 시 즉시 STOP + 사용자 결정 대기

### R-DUD-004 — 새 섹션 본문이 source rules 와 drift

- Severity: **Medium** (장기 maintenance 부담)
- Likelihood: Medium (rules 본문은 evolvable 영역)
- Mitigation:
  - 새 섹션 끝 `### 관련 문서` 에 `.claude/rules/moai/workflow/ci-{watch,autofix}-protocol.md` 명시 cross-ref
  - 본문에는 spec 수치 (30초, 3 iter, 30분) 만 인용 — rules body 의 sub-detail (정확한 exit code, JSON envelope) 는 인용 안 함 → drift 표면 축소
- Contingency: 향후 ci-watch / ci-autofix doctrine 변경 시 동반 PR 로 sync.md 업데이트

### R-DUD-005 — en/ja/zh 번역 품질 부족

- Severity: **Low**
- Likelihood: Medium
- Mitigation:
  - 기존 sync.md 의 본문 스타일 (ko 기존 → en/ja/zh 기존 번역) 패턴 따라 동일 톤
  - locale별 native phrase 차이는 EC-DUD-003 (±20% 허용 범위) 내 자연 변동 허용
- Contingency: 머지 후 native speaker 검토에서 issue 발견 시 patch SPEC

## 9. Traceability — Files Affected vs SPEC artifacts

### 9.1 In-scope files (run-phase 수정 대상)

| File | Change |
|------|--------|
| `docs-site/content/ko/workflow-commands/moai-sync.md` | INSERT new H2 section before `## 품질 게이트` |
| `docs-site/content/en/workflow-commands/moai-sync.md` | INSERT new H2 section before `## Quality gates` (or fallback) |
| `docs-site/content/ja/workflow-commands/moai-sync.md` | INSERT new H2 section before `## 品質ゲート` (or fallback) |
| `docs-site/content/zh/workflow-commands/moai-sync.md` | INSERT new H2 section before `## 质量门` (or fallback) |

총 4 in-scope files. 라인 수 변화 baseline (ko 570 / en 511 / ja 511 / zh 512 = 총 2,104) → 예상 (+~150 lines, ko 35-45 lines × 4 locales).

### 9.2 SPEC artifacts (plan-phase 산출물, 본 디렉토리)

| File | Purpose |
|------|---------|
| `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/spec.md` | Requirements + EARS REQs + Scope + Edge cases + Risks |
| `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/acceptance.md` | 8 binary AC + Traceability + DoD + verification commands |
| `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/plan.md` | 본 문서 — Section A-E template (Tier S optional but applied for clarity) |

run-phase 추가 산출물: `progress.md` (M1-M5 진행 기록).

## 10. Cross-references

- baseline 보고서: `.moai/research/moai-adk-current-state-2026-05-22.md` §8 (F1-F8 table)
- v3.0 환골탈태 설계: `.moai/research/v3.0-design-2026-05-22.md` Wave 0 (라인 351-356)
- Rules SSoT: `.claude/rules/moai/workflow/ci-{watch,autofix}-protocol.md`
- i18n doctrine: `.moai/docs/docs-site-i18n-rules.md`
- Frontmatter SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- 선례 SPEC: SPEC-V3R6-DOCS-CMD-CATALOG-001 (PR #1045 OPEN MERGEABLE) — depends_on
- 선례 SPEC: SPEC-V3R5-DOCS-SECURITY-001 (commit `c94d8b203`) — 4-locale 신설 패턴 참조
- Section A-E delegation template: `.claude/rules/moai/development/manager-develop-prompt-template.md` (Tier S optional, 본 SPEC 은 적용)
