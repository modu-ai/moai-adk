---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Plan — Template Internal-Content Isolation"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template, isolation, internal-content, plan"
tier: M
---

# Plan — Template Internal-Content Isolation

## §A. 맥락 (Context)

본 SPEC은 선행 Tier L SPEC (이하 "선행 SPEC") 진행 중 도입된 약 44 leak files 중, 2회의 partial cleanup 후 잔존하는 35 files를 영구 정리하고, 재발 방지 메커니즘 4종을 동시 도입한다. 1-pass plan-run-sync cycle을 목표로 하나, scope의 광범위성(35 files × 5 leak class)으로 인해 W2 phase는 sub-batch 분할.

선행 cleanup 패턴 학습:
- pass 1 (9 files, -80 lines): NOTICE.md 전체 section 제거 + SPEC ID prose 일반화 + REQ token 제거 + 4-source generalization (frontmatter schema, common protocol, archived-agent-rejection, orchestration-mode-selection) + 3-hook header cleanup
- pass 2 (2 files): agent-authoring.md 4-spot + spec-frontmatter-schema.md 2-spot SPEC ID/REQ 일반화

## §B. 알려진 이슈 (Known Issues)

- **K1**: `.moai/docs/generic-patterns-guide.md` (template 내)는 LNCO-001 산출물로 신규 작성됨. 의도가 "사용자 학습용 generic pattern reference"인지 "메인테이너 내부 history"인지 W6 review 시 분류 필요
- **K2**: `.gitignore` (template 내)에 `SPEC-V3R6-UPDATE-NOISE-001` 주석 인용 발견 — 단일 주석 라인. 변환 전략: 주석을 generic prose ("선행 cleanup SPEC")로 변경 또는 단순 삭제
- **K3**: `internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md` (line 248)이 archived agent 12종을 enum 형태로 노출 — 학습 가치 있으나 dev-internal narrative. 변환 전략 W3 design.md §B에서 결정

## §C. Pre-flight (재검증)

M1 시작 시 다음 4 명령을 parallel batch로 실행하여 ground truth 재확인:

```bash
# 1. 잔여 leak file count 재검증
grep -rln 'SPEC-V3R6-\|REQ-ATR-\|Audit 3\|Finding A[1-6]\|archive-2026-05-25' internal/template/templates/ | wc -l
# expected: 35 (if drift, surface to plan as blocker)

# 2. origin/main race check
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# expected: 0 0

# 3. embedded.go 현재 hash baseline
shasum internal/template/embedded.go

# 4. CLAUDE.local.md §25 부재 재확인
grep -c 'Template Internal-Content Isolation' CLAUDE.local.md
# expected: 0
```

## §D. 제약 (Constraints, derived from spec.md §C)

본 plan은 spec.md §C의 5 HARD + 3 SHOULD 제약을 무조건 honor.

추가 plan-level 제약:
- **[PLAN-1]** Sub-batch atomic boundary: 각 W2 sub-batch는 단일 git commit으로 land, sub-batch 내부에서 partial 적용 금지
- **[PLAN-2]** Allowlist 도입 시 design.md §C에 명시적 justification

## §E. 자가 검증 (Self-Verification)

각 sub-batch 완료 시 자가 검증 5-cmd batch:

```bash
# 1. Leak count 감소 추세 검증
grep -rln 'SPEC-V3R6-\|REQ-ATR-\|Audit 3\|Finding A[1-6]\|archive-2026-05-25' internal/template/templates/ | wc -l

# 2. make build 무결성
make build && shasum internal/template/embedded.go

# 3. Go test 무결성 (기존 테스트 회귀 없음)
go test ./internal/template/...

# 4. Lint test 자체 동작 (W4 이후)
go test -run TestTemplateNoInternalContentLeak ./internal/template/...

# 5. CLAUDE.local.md §25 존재 (W3 이후)
grep -c '## 25. Template Internal-Content Isolation' CLAUDE.local.md
```

## §F. Milestones

본 plan은 6-milestone (M1~M6) 구조를 가진다. 우선순위는 Priority High → Medium 순서.

### M1 — W1 Audit + Classification (Priority High)

**입력**: 35 leak files (spec.md §A.4 ground-truth 표)
**작업**:
- 35 files 각각에 대해 leak class C1~C6 분류 (Classification Table 작성)
- 각 file의 leak 위치(line number)와 context 추출
- Predecessor cleanup pattern과의 호환성 확인
- 각 file별 변환 strategy (단순 prose 일반화 / mirror re-generation / 본문 재구성) 결정
**산출물**: `plan.md §F.M1` 에 32+ rows classification table 첨부 (M1 commit 시 plan.md amend)
**검증**: 35 files 모두 분류됐고, 각 file당 최소 1개의 변환 strategy 명시
**Owner**: manager-develop (1차 직접 실행 권장)

### M2 — W3 CLAUDE.local.md §25 NEW (Priority High)

**입력**: spec.md §B.2 (REQ-TII-004, REQ-TII-005) + predecessor partial cleanup history
**작업**:
- CLAUDE.local.md 끝(§24 다음)에 `## 25. Template Internal-Content Isolation` 신규 작성
- 4 subsection: §25.1 정의 + §25.2 forbidden/allowed content classes + §25.3 5-item pre-commit self-check + §25.4 anti-pattern catalogue (≥3 worked examples)
- §15, §21, §24와의 cross-reference 명시
**산출물**: CLAUDE.local.md §25 (대략 80-120 lines)
**검증**: `grep -c '## 25. Template Internal-Content Isolation' CLAUDE.local.md` → 1
**Owner**: orchestrator-direct (CLAUDE.local.md은 maintainer-only file, manager-develop 위임 가능하나 직접 편집 효율적)

### M3 — W4 Go lint test 작성 + red-green proof (Priority High)

**입력**: spec.md §B.3 (REQ-TII-006~008) + design.md §C (allowlist 설계)
**작업**:
- `internal/template/internal_content_leak_test.go` 신규 작성
- `filepath.Walk` 기반 templates 디렉토리 순회 + 5-pattern regex 매칭
- Allowlist Go 슬라이스 (design.md §C 명시)
- Red proof: cleaned HEAD에서 PASS 확인
- Green-to-red proof: synthetic leak 1개 임시 주입 → FAIL → 복원 → PASS (CI에서는 자동 검증 불가, plan-phase 산출물로 demo 결과 acceptance.md AC-TII-007 commands에 포함)
**산출물**: `internal/template/internal_content_leak_test.go` (대략 100-150 LOC) + design.md §C allowlist
**검증**: `go test -run TestTemplateNoInternalContentLeak ./internal/template/...` PASS
**Owner**: manager-develop

### M4 — W2 35 files cleanup (sub-batched, Priority High)

**입력**: M1 classification table + M3 lint test (PASS 보장 안내)
**작업**: 35 files를 4개 sub-batch로 분할 처리
- **Sub-batch 4.1**: `.claude/agents/` (4+1 = 5 files) — manager-* core/ + meta/plan-auditor.md
- **Sub-batch 4.2**: `.claude/rules/moai/` (3+2+5 = 10 files) — core/development/workflow/ rules
- **Sub-batch 4.3**: `.claude/skills/` (2+7+6 = 15 files) — moai-foundation-core/ + moai/workflows/ + moai-workflow-*/
- **Sub-batch 4.4**: 단발 파일 5 files — output-styles, hooks, system.yaml.tmpl, generic-patterns-guide.md, .gitignore

각 sub-batch 완료 시 §E 5-cmd self-verification + atomic commit.

**산출물**: 4 atomic commits × 5/10/15/5 files = 35 cleaned files
**검증**: 각 sub-batch 후 leak count 35 → 30 → 20 → 5 → 0 단계적 감소
**Owner**: manager-develop (orchestrator는 sub-batch boundary verification)

### M5 — W5 CI workflow integration (Priority Medium)

**입력**: spec.md §B.4 (REQ-TII-009, REQ-TII-010) + 기존 `.github/workflows/test.yml` 또는 등가물
**작업**:
- `.github/workflows/` 디렉토리 inspection으로 기존 test job 위치 식별
- 기존 job에 `go test ./internal/template/...` 가 이미 포함되어 있으면 cross-reference docstring 추가만; 아니면 step 추가
- workflow file head에 policy rationale docstring 추가 (memory feedback file path 인용)
**산출물**: `.github/workflows/test.yml` (또는 등가) modified
**검증**: workflow YAML lint + GitHub Actions trigger simulation (직접 PR 시 검증)
**Owner**: manager-develop

### M6 — W6 Maintainer-only file audit (Priority Medium)

**입력**: spec.md §B.5 (REQ-TII-011, REQ-TII-012) + CLAUDE.local.md §21 dev-only file enumeration
**작업**:
- `internal/template/templates/` 전체 디렉토리에서 §21 forbidden file class 검색:
  - `97-*.md`, `98-*.md`, `99-*.md` 슬래시 커맨드
  - `settings.local.json` (template에 절대 금지)
  - `.moai/state/last-cc-version.json` (dev-only tracking)
  - `.moai/research/cc-update-*.md` (dev-only report)
- 발견 시 `git rm` + `.gitignore` 가드 추가
- 발견되지 않을 시 PASS 보고 (acceptance.md AC-TII-011 cmd에서 자동 검증)
**산출물**: `git rm` commits (조건부, 0~N files) + `.gitignore` 갱신 (조건부)
**검증**: `find internal/template/templates -name '9[789]-*.md' -o -name 'settings.local.json' -o -name 'last-cc-version.json'` → empty
**Owner**: manager-develop (audit 후 결과를 orchestrator에게 보고)

## §G. Anti-Patterns (M1~M6 진행 시 회피)

- **AP-1**: M4 sub-batch boundary 무시하고 한 commit으로 35 files 처리 → reviewability 손실
- **AP-2**: §25 작성 시 spec.md/plan.md 본문을 그대로 복사 → CLAUDE.local.md는 maintainer doctrine 위치, SPEC 본문과 의도 다름
- **AP-3**: Lint test에 allowlist 너무 광범위 도입 → 본 SPEC의 isolation 의도 무력화
- **AP-4**: Synthetic leak proof을 commit해서 main에 남김 → CI 회귀 / proof는 임시여야 함
- **AP-5**: M6 audit pass에서 §21 enumeration 외 임의 파일 제거 → scope creep

## §H. 위험 (Risks)

| ID | Risk | Mitigation |
|----|------|-----------|
| R1 | Mirror parity 실패 (make build 후 embedded.go drift) | M1 직후 baseline shasum, 각 sub-batch 후 재검증 + 회귀 시 즉시 revert |
| R2 | Allowlist over-permissive | design.md §C에 항목별 justification 강제; manager-spec 자가 review |
| R3 | 35 files 중 일부가 사용자 학습 가치 높은 historical narrative (e.g., archived agent migration table) → 일반화 시 정보 손실 | W1 M1 classification에서 "정보 보존 priority" column 추가; 정보 손실 가능성 시 design.md §B에 대안 (e.g., `.moai/docs/` 영구 reference로 외부화) |
| R4 | CI workflow file이 monorepo에서 multiple로 분산 → M5에서 일관성 보장 어려움 | M5 시작 시 `find .github/workflows/ -name '*.yml'` 전수 inspection 후 영향 받는 workflow file 전체 enumeration |
| R5 | .moai/docs/generic-patterns-guide.md (LNCO-001 산출물)이 사용자 deploy 의도였는지 maintainer-internal 의도였는지 ambiguous | M6 W6 audit pass에서 명시 결정; ambiguous 시 user-facing 가정 + leak token 제거로 처리 |

## §I. Cross-References

- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/spec.md` — REQ-TII 13개 정의
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/acceptance.md` — AC-TII 12개 verifiable commands
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/design.md` — Allowlist 설계 + substitution dictionary + decision log
- `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/research.md` — predecessor cleanup pattern 분석 + 기존 template test 조사
- CLAUDE.local.md §15 (Language Neutrality), §21 (Dev-Only Commands), §24 (Harness Namespace), §25 (NEW, 본 SPEC 산출물)
- 메모리 feedback 파일: `/Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_template_internal_content_isolation.md` (8996 bytes)
- Predecessor partial cleanup commits: `20a66df85` (pass 1, 9 files) + `40dc43f5b` (pass 2, 2 files)
