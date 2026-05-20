# Plan — SPEC-V3R5-DOCS-SECURITY-001

## Tier 판단 + 근거

**Tier: M (Medium)**

- 영향 파일 수: 4 페이지 × 4 locale = **16 markdown files** (Tier S 임계 <5 files 초과)
- 추정 LOC (단어 환산): security-notes.md ~3K 단어/locale × 4 = ~12K 단어 + 3 페이지 부분 갱신 × 4 = ~4-5K 단어 → 약 16-17K 단어 ≈ 1-1.5K LOC (Tier M 범위 300-1500 LOC 적합)
- CHANGELOG.md 갱신 추가 → 총 17 files affected
- 본 SPEC 자체 산출물: 3-artifact set (spec.md + plan.md + acceptance.md)
- plan-auditor PASS threshold: **0.80** (Tier M)
- Section A-E 위임 prompt template **required** (manager-develop 위임 시)
- research doc §8.2 의 Tier S 권고는 per-locale 기준 추정으로 보이며, 4-locale parity 의무 (`docs-site-i18n-rules.md` §17.3) 를 반영하면 Tier M 이 정확. SPEC-V3R5-WORKFLOW-LEAN-001 spec-workflow.md § SPEC Complexity Tier 의 anti-pattern 회피 (Tier S 로 잘못 분류해 overhead 회피 시도 차단).

## Approach (High-level)

### Phase 1 — 골격 준비 (M1)

- ko canonical 의 `security-notes.md` 골격 작성 (6 섹션, ~3K 단어)
- ko canonical 의 `settings-json.md` 패치 (settings.local.json 0o600 섹션 추가)
- ko canonical 의 `update.md` 패치 (mandatory checksum 섹션 추가)
- ko canonical 의 `cg-mode.md` 패치 (tmux source-file injection 섹션 추가)
- CHANGELOG.md `[Unreleased]` SECURITY 섹션 추가

### Phase 2 — 4-locale 동기 (M2)

- ko → en 번역 (영문 보안 용어는 원문 유지)
- ko → ja 번역 (en bridge 활용)
- ko → zh 번역 (en bridge 활용)
- Heading slug 일관성 확보 (영문 explicit slug `{#cwe-732}` 등)

### Phase 3 — 정합성 검증 (M3)

- `scripts/docs-i18n-check.sh` 실행 → exit 0 확인
- 금지 URL grep (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`) → 0 matches
- Mermaid LR/BR grep → 0 matches
- 본문 이모지 grep → 0 matches
- `moai spec lint --strict` 본 SPEC 디렉토리 → 0 NEW 위반

### Phase 4 — manager-develop run-phase 위임 준비

본 SPEC 의 run-phase 는 별도 진입 (`/moai run SPEC-V3R5-DOCS-SECURITY-001`). 위임 prompt 는 manager-develop-prompt-template.md Section A-E 5-section 구조 사용.

위임 prompt 의 Section B (Known Issues) 필터링: 본 SPEC 은 documentation-only 이므로 B1 (cross-platform build tags) / B2 (cross-SPEC 충돌, retired SPEC) / B3 (C-HRA-008 subagent boundary) / B7 (observer.go path) 는 **N/A**. 적용 카테고리: **B4** (frontmatter canonical schema), **B5** (CI 3-tier: spec-lint + docs-i18n-check + lint baseline), **B6** (spec-lint heading 규약 — `### N.M Out of Scope` h3), **B8** (working tree hygiene — `.moai/state/` `internal/hook/.moai/` `.moai/harness/usage-log.jsonl` 변경 금지).

cycle_type 선택: **ddd** (documentation-only 의 경우 ANALYZE-PRESERVE-IMPROVE 적합 — ANALYZE: 기존 페이지 구조 분석, PRESERVE: hugo frontmatter + 4-locale parity 보존, IMPROVE: 신규 섹션 + 패치 적용).

## Milestones

### M1 — ko canonical 작성 (Priority: High)

- 산출물:
  - `docs-site/content/ko/advanced/security-notes.md` (신규, ~3K 단어, 6 섹션)
  - `docs-site/content/ko/advanced/settings-json.md` (패치, settings.local.json 0o600 섹션 추가)
  - `docs-site/content/ko/getting-started/update.md` (패치, mandatory checksum 섹션 추가)
  - `docs-site/content/ko/multi-llm/cg-mode.md` (패치, tmux source-file 섹션 추가)
- DoD: 4 페이지 모두 hugo render 통과 + heading slug 영문 explicit (`{#cwe-732}` 등)

### M2 — en/ja/zh 동시 작성 (Priority: High, M1 의존)

- 산출물 (per locale × 3 locales = 12 files):
  - `docs-site/content/{en,ja,zh}/advanced/security-notes.md`
  - `docs-site/content/{en,ja,zh}/advanced/settings-json.md`
  - `docs-site/content/{en,ja,zh}/getting-started/update.md`
  - `docs-site/content/{en,ja,zh}/multi-llm/cg-mode.md`
- DoD: 4-locale parity (`scripts/docs-i18n-check.sh` PASS), heading slug 동일

### M3 — CHANGELOG + 정합성 검증 (Priority: High, M2 의존)

- 산출물:
  - `CHANGELOG.md` `[Unreleased]` SECURITY 섹션 추가 (4 commit hashes + CWE 키워드)
  - 자체 검증 보고서 (commit body 에 binary AC matrix)
- DoD: AC-DSEC-001 ~ AC-DSEC-007 모두 binary PASS

## Technical Approach

### Heading Slug 전략

hugo 자동 slug 는 한글/일본어/중국어 제목에서 unstable 한 결과를 낼 수 있다. 4-locale anchor cross-reference 일관성 확보를 위해 **영문 explicit slug** 채택:

```markdown
### CWE-732 — settings.local.json permission hardening {#cwe-732}
### CWE-214 — tmux IPC token argv exposure {#cwe-214}
### CWE-345 — Update flow mandatory checksum {#cwe-345}
```

이 slug 는 4-locale 모두에 동일하게 적용되어 cross-reference (`[CWE-732 섹션](#cwe-732)`) 가 작동.

### hugo Frontmatter 규약

각 페이지의 frontmatter 는 다음 구조:

```yaml
---
title: "Security Notes"          # locale 별 번역
description: "..."                # locale 별 번역
weight: 25                        # 4-locale 동일
tags: ["security", "cwe", "audit"] # 4-locale 동일
---
```

`weight` 와 `tags` 는 4-locale 간 identical. `title`/`description` 만 locale 별 번역.

### CWE 표기

CWE 번호 + 영문 명칭은 4-locale 모두 원문 유지. 한국어/일본어/중국어 본문은 번역하되 `CWE-732 (Incorrect Permission Assignment for Critical Resource)` 같이 영문 병기.

### 점검 체크리스트 명령 예시

`advanced/security-notes.md` §5 (점검 체크리스트) 에 포함될 명령:

```bash
# CWE-732 점검 — settings.local.json permission
stat -c '%a' .claude/settings.local.json    # Linux: expect 600
stat -f '%A' .claude/settings.local.json    # macOS: expect 600

# CWE-214 점검 — tmux env 노출 확인 (cg-mode 실행 중)
tmux show-environment -g | grep -i ANTHROPIC
# argv 에는 token 없음 / source-file 경유 확인:
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'  # expect: 0 matches

# CWE-345 점검 — update flow checksum 동작 확인
moai update --check                          # release 정보 + checksums.txt 존재 확인
moai update                                  # 실패 시 ErrChecksumUnavailable 메시지 확인
```

### Cross-reference 패턴

각 페이지에 다음 cross-reference 추가:

- `advanced/settings-json.md` → `advanced/security-notes.md#cwe-732`
- `getting-started/update.md` → `advanced/security-notes.md#cwe-345`
- `multi-llm/cg-mode.md` → `advanced/security-notes.md#cwe-214`
- `advanced/security-notes.md` → `CHANGELOG.md` `[Unreleased]` 섹션

## Section A-E Delegation Strategy (run-phase 위임 시 적용)

본 SPEC 의 run-phase 위임 시 manager-develop 에게 전달할 prompt 골격 (manager-develop-prompt-template.md 5-section 구조):

### Section A — Context

- 작업 위치: `/Users/goos/MoAI/moai-adk-go` (main worktree, Late-Branch 정책)
- 현재 branch: (run-phase 진입 시점), HEAD (run-phase 진입 시점)
- SPEC 산출물: `.moai/specs/SPEC-V3R5-DOCS-SECURITY-001/{spec,plan,acceptance}.md`
- plan-auditor verdict: (plan-phase 완료 후 기록)
- PRESERVE: 기존 페이지의 hugo frontmatter, 다른 섹션 콘텐츠, 다른 페이지

### Section B — Known Issues (필터링 적용)

- B4 (frontmatter canonical 12-field)
- B5 (CI 3-tier: spec-lint + docs-i18n-check + markdown lint)
- B6 (`### N.M Out of Scope` h3 sub-section)
- B8 (`.moai/state/` `internal/hook/.moai/` `.moai/harness/usage-log.jsonl` PRESERVE)
- B1 / B2 / B3 / B7 N/A (documentation-only)

### Section C — Pre-flight Check

```bash
# 1. 현재 branch + baseline 확인
git branch --show-current
git rev-parse HEAD

# 2. 기존 페이지 부재 확인 (security-notes.md, 신규 대상)
ls docs-site/content/{ko,en,ja,zh}/advanced/security-notes.md 2>&1

# 3. 4-locale 기준 페이지 존재 확인
ls docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md
ls docs-site/content/{ko,en,ja,zh}/getting-started/update.md
ls docs-site/content/{ko,en,ja,zh}/multi-llm/cg-mode.md

# 4. i18n script 실행 가능성 확인
ls scripts/docs-i18n-check.sh

# 5. spec-lint baseline
go run ./cmd/moai spec lint --strict 2>&1 | tail -5
```

### Section D — Constraints

- PRESERVE: 기존 페이지 frontmatter (`title`/`description` 외 필드), 다른 섹션 콘텐츠, `.moai/harness/usage-log.jsonl`, `internal/hook/.moai/` (anomaly dir), `.moai/specs/SPEC-V3R5-ATOMIC-WRITE-001/` (in-flight)
- 무관 working tree modified files (banner.go / wizard.go / doctor.go 등 21건) commit 포함 금지
- 금지: `--no-verify`, `--amend`, force-push to main
- 사용 의무: Conventional Commits + `🗿 MoAI` trailer
- docs-i18n-rules.md §17.1 금지 URL / §17.2 Mermaid TD only / §17.3 4-locale 동시 준수
- Late-Branch 정책: main 직진 commit, PR 시점 branch 분리

### Section E — Self-Verification Deliverables

manager-develop 완료 보고 시 다음 자체 검증 결과 강제 포함:

| 항목 | 검증 명령 | Expected |
|------|-----------|----------|
| 4-locale parity | `find docs-site/content/{ko,en,ja,zh}/advanced/security-notes.md \| wc -l` | 4 |
| i18n script | `scripts/docs-i18n-check.sh` | exit 0 |
| 금지 URL grep | `grep -rE 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content/` | 0 matches |
| Mermaid TD only | `grep -rE 'flowchart LR\|graph LR\|flowchart BR\|graph BR' docs-site/content/{ko,en,ja,zh}/{advanced/security-notes.md,advanced/settings-json.md,getting-started/update.md,multi-llm/cg-mode.md}` | 0 matches |
| CHANGELOG SECURITY | `grep -E 'SECURITY\|### Security\|CWE-' CHANGELOG.md` | ≥1 match |
| 본문 이모지 | `grep -P '[\x{1F300}-\x{1F9FF}]' docs-site/content/{ko,en,ja,zh}/{advanced/security-notes.md,advanced/settings-json.md,getting-started/update.md,multi-llm/cg-mode.md}` | 0 matches |
| spec-lint clean | `go run ./cmd/moai spec lint --strict 2>&1 \| grep DOCS-SECURITY` | 0 NEW 위반 |

## Risks (Implementation Phase)

| ID | Risk | Mitigation |
|----|------|------------|
| **R-IMPL-001** | en/ja/zh 번역 시 보안 용어 의미 유실 | 영문 보안 용어 (CWE, source-file injection, mandatory checksum 등) 원문 병기 + en 페이지를 ja/zh bridge 로 사용 |
| **R-IMPL-002** | hugo build 실패 (frontmatter 오타, broken link) | 각 페이지 작성 후 `hugo --baseURL=...` 로 dry-run 시도 (manager-develop 자체 검증) |
| **R-IMPL-003** | `scripts/docs-i18n-check.sh` 가 신규 페이지의 parity 자동 검출 못 함 | M3 phase 에서 script 동작 사전 확인. 필요 시 script 보강을 별도 follow-up SPEC (`SPEC-V3R5-DOCS-I18N-CHECK-001` 가칭) 으로 분리. 본 SPEC scope 는 콘텐츠만. |
| **R-IMPL-004** | CHANGELOG.md `[Unreleased]` 섹션 부재 → 신규 섹션 생성 필요 | CHANGELOG.md 사전 확인 (Phase 1 첫 작업). 부재 시 `[Unreleased]` 헤더 + 4 카테고리 (Added/Changed/Fixed/Security) 골격 생성. |

## Open Questions (plan-auditor 가 검토할 후보)

1. `advanced/security-notes.md` 의 `weight` 값 — 기존 `advanced/_meta.yaml` 정렬 순서와 충돌 없는지? (현재 추정 25, 검증 필요)
2. `scripts/docs-i18n-check.sh` 의 동작 — 신규 파일 자동 발견 vs 명시적 등록 필요? (M3 단계에서 검증)
3. CHANGELOG.md `[Unreleased]` 섹션 현재 상태 — 다른 entry 들이 이미 있는지? (M3 단계 사전 확인)

이 3개 Open Question 은 run-phase Section C pre-flight 단계에서 manager-develop 이 즉시 해소 가능 (low-risk inquiry).
