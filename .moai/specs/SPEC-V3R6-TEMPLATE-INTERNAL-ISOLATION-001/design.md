---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Design — Template Internal-Content Isolation"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template, isolation, design, allowlist, substitution"
tier: M
---

# Design — Template Internal-Content Isolation

## §A. 설계 개요

본 design 문서는 spec.md의 13 REQ와 plan.md의 6 milestone을 실현하기 위한 3가지 핵심 아키텍처 결정을 다룬다:

- §B Substitution Dictionary: cleanup 시 일관된 변환 vocabulary
- §C Allowlist Mechanism: lint test에서 legitimate exception 처리
- §D CI Hook Placement: gating 위치 결정 (Actions vs make test vs pre-commit)

추가로 §E Decision Log에서 5개 주요 결정의 rationale를 기록한다.

## §B. Substitution Dictionary

35 leak files를 cleanup할 때 일관된 변환을 적용하기 위해 다음 6개 substitution rule을 정의한다. Predecessor partial cleanup pass 1 (`20a66df85`) + pass 2 (`40dc43f5b`)의 패턴을 학습하여 확장.

### B.1 변환 규칙 (substitution rules)

| Rule | Leak class | Pattern (regex) | Substitution strategy | Predecessor 사용 사례 |
|------|-----------|-----------------|----------------------|---------------------|
| S1 | C1 SPEC ID literal | `SPEC-V3R6-[A-Z0-9-]+-\d{3}` | "선행 SPEC" / "predecessor SPEC" / 일반 prose ("일부 cleanup SPEC"). 1차로 generic prose 대체 시도, 학습 가치 손실 시 design.md §B.4 외부화 옵션 적용 | pass 1 NOTICE.md 섹션 제거, pass 1 spec-frontmatter-schema 4-spot |
| S2 | C2 REQ token | `REQ-[A-Z]+-\d{3}` | 본문이 REQ token의 semantic을 명시할 경우 token 자체를 prose 설명으로 풀어쓰기; 단순 인용일 경우 삭제 | pass 1 agent-common-protocol REQ-ATR-009/014 제거, pass 2 spec-frontmatter-schema REQ-SPC-003-006 제거 |
| S3 | C3 Audit citation | `Audit \d+ Findings?\s*A?\d+(-A?\d+)?` or `Finding A[1-6]` | 전체 인용 단락 삭제 (메인테이너 audit narrative) — 학습 가치는 generic Anthropic best practice prose로 별도 표현 가능 | pass 1 NOTICE.md "Anthropic 2026 Alignment" 섹션 45-line 통째 삭제 |
| S4 | C4 Archive date | `archive-2026-05-25` or `\.moai/backups/agent-archive-2026-05-25/` | 경로 또는 날짜 인용을 generic placeholder ("archive directory" / "the migration backup location")로 변경 | pass 1 archived-agent-rejection 날짜 references 제거 |
| S5 | C5 Commit sha | `\b[0-9a-f]{7,40}\b` 7~40-hex (단 git log 본문 외) | 단순 인용 (`see commit b957a4d04`) 시 prose ("plan-phase 직후 commit")로 풀어쓰기 또는 삭제 | (predecessor 미사용, 본 SPEC 신규 패턴) |
| S6 | C6 Compound | 위 5 class 2개 이상 동시 발생 | 우선순위 S1 > S2 > S3 > S4 > S5 순서로 적용; 단락 전체 재작성이 효율적이면 단락 prose 재구성 | pass 1 NOTICE.md (S1+S3+S4 compound → 섹션 통째 삭제) |

### B.2 변환 priority (정보 보존 vs 격리)

각 leak token에 대해 다음 우선순위로 strategy 선택:

1. **재명명 (rename)**: SPEC ID → "선행 SPEC" 등 generic 대체로 1차 시도
2. **재구성 (rephrase)**: token 인용이 문법적으로 부자연스러우면 단락 전체 재작성
3. **삭제 (delete)**: 인용이 단순 인용 (학습 가치 0)일 때
4. **외부화 (externalize)**: 학습 가치 높은 historical narrative (예: archived agent migration table)는 generic prose로 변환 불가 → 별도 user-facing reference (`.moai/docs/` 영구 문서)로 이동하고 template에서는 단순 cross-reference만 유지

### B.3 변환 vocabulary (Korean / English mix per `documentation: ko`)

| Generic substitute | 사용 맥락 |
|--------------------|---------|
| "선행 SPEC" / "predecessor SPEC" | SPEC ID 인용 대체 (Korean 본문) |
| "선행 audit / earlier audit" | Audit 인용 대체 |
| "archive directory" / "archived agent location" | archive 날짜 인용 대체 |
| "초기 release commit" / "early release commit" | commit sha 인용 대체 |
| "rebuild / restructure" | 특정 SPEC 작업 narrative |

### B.4 외부화 옵션 (정보 보존 critical 경우)

`.claude/skills/moai-foundation-core/SKILL.md` line 248의 12 archived agent enumeration처럼 사용자 학습 가치 높은 historical narrative는 다음 중 선택:

- **Option A (선호)**: generic하게 "Some agents were retired during a refactor; see archived rejection guide for details" 식의 prose + cross-reference만 유지
- **Option B (대안)**: `.moai/docs/agent-migration-history.md` (template-local, 사용자 deploy 대상) 별도 작성하여 12 archived agent 정보 보존; template/SKILL.md에서는 단순 link만 유지

W1 M1 classification 단계에서 결정. 본 design.md는 default Option A 추천.

## §C. Allowlist Mechanism (lint test)

### C.1 설계 결정: in-test Go literal slice

3가지 후보 평가:

| 후보 | 장점 | 단점 | 결정 |
|------|-----|------|------|
| (A) In-test Go literal slice | 명시적 documentation, code review에서 visible, Go-idiomatic | grep allowlist보다 유연성 떨어짐 | **선택** |
| (B) YAML allowlist file (`internal/template/leak_allowlist.yaml`) | 외부 편집 용이, 비개발자 review 가능 | 추가 파일 + YAML parser 도입 | 미선택 |
| (C) Comment marker pattern (`<!-- isolation-allowlist: rationale -->`) | 파일 자체에 inline | 파편화, 누락 위험 | 미선택 |

**선택 이유**: 본 SPEC scope에서 allowlist 항목 수 0~5개 예상 (REQ-TII-007의 "minimal allowlist" SHOULD constraint). Go literal slice 5개 entry 정도면 가독성 + maintainability 충분.

### C.2 Allowlist 구조

```go
// in internal/template/internal_content_leak_test.go (proposed)
var allowlist = []allowlistEntry{
    // 현재 candidate 0개. 발견 시 W1 M1 classification 단계에서 추가.
    // 예시 entry (실제 구현 시 적합한 경우만):
    // {
    //     path: "internal/template/templates/.claude/rules/moai/NOTICE.md",
    //     pattern: "Apache 2.0 attribution",
    //     rationale: "Third-party license attribution dates must be preserved verbatim per Apache 2.0 §4(c)",
    // },
}

type allowlistEntry struct {
    path      string  // exact file path (relative to repo root)
    pattern   string  // human-readable description of the allowed token
    rationale string  // why this token is exempt (legal, attribution, etc.)
}
```

### C.3 Allowlist 정당화 룰 (REQ-TII-007 honor)

새 allowlist entry 추가 시 다음 3가지 중 1개 이상 만족해야 함:

1. **Legal/license requirement**: 제3자 라이선스 (Apache 2.0, MIT 등)에서 특정 attribution 인용을 verbatim 요구
2. **Domain-specific necessity**: pedagogical 가치가 generic 변환으로 손실되어 사용자 학습이 불가능
3. **Mirror source requirement**: source-of-truth 파일이 본 SPEC scope 외부 (e.g., `.claude/rules/moai/NOTICE.md` 자체)이고, mirror가 source와 동일해야 하는 contract 존재

각 entry는 `rationale` 필드에 위 3개 중 어느 것을 만족하는지 명시.

## §D. CI Hook Placement Decision

### D.1 후보 평가

| 위치 | 장점 | 단점 | 결정 |
|------|------|------|------|
| (a) `.github/workflows/test.yml`의 기존 `go test ./...` step (REQ-TII-009의 default) | 기존 CI 인프라 재사용, 추가 step 없음 | 다른 `go test` failure에 묻힐 가능성 | **선택 (primary)** |
| (b) 신규 `.github/workflows/template-isolation.yml` 독립 workflow | 명시적 step 이름, 빨리 식별 | workflow 파편화 | 미선택 (불필요한 복잡도) |
| (c) `.git/hooks/pre-commit` shell hook | 즉각 피드백 (local) | git hook은 template 동기화 안 됨 (§23.1) — 모든 기여자에게 강제 불가 | 보조 (선택적, W6 audit pass에서 추가 검토) |
| (d) `make test` Makefile target | 개발자 routine에 통합 | CI 보장이 없음 (개발자가 make test 안 돌리면 통과) | 미선택 (CI 게이팅 보장 부족) |

**1차 선택**: (a) — `go test ./internal/template/...` 가 기존 CI test job에 이미 포함될 가능성 높으므로, M5 W5 단계에서 inspection 후 docstring cross-reference만 추가하거나 누락 시 step 추가.

**보조 선택**: 사용자가 local에서 pre-commit hook 사용 시 `make test` 또는 `go test ./internal/template/...` 권장 사용을 §25 self-check checklist에 명시.

### D.2 Docstring cross-reference 형식

해당 workflow YAML 상단에 다음 형식 docstring 추가:

```yaml
# Template Internal-Content Isolation Policy:
# This test job enforces CLAUDE.local.md §25 (Template Internal-Content Isolation).
# Policy rationale: see local memory feedback file
# (/Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_template_internal_content_isolation.md)
# Failure mode: TestTemplateNoInternalContentLeak detects leak tokens in internal/template/templates/.
```

위 docstring은 본 SPEC ID 인용 회피 (REQ-TII-013) — generic policy 인용만 유지.

## §E. Decision Log

| ID | Decision | Rationale | Date |
|----|---------|----------|------|
| D1 | Substitution dictionary 6-rule 도입 | Predecessor 2-pass cleanup의 ad-hoc 변환이 일관성 부족 → systematic dictionary로 격상; manager-develop가 W2 sub-batch별로 일관 적용 가능 | 2026-05-25 |
| D2 | Allowlist를 in-test Go literal로 결정 | 외부 YAML 파일 도입은 본 SPEC scope (35 files, 예상 allowlist ≤5개)에 비해 over-engineering; Go literal slice가 explicit + code review-friendly | 2026-05-25 |
| D3 | CI gating을 기존 `go test ./...` workflow에 통합 (독립 workflow 신설하지 않음) | 별도 workflow는 fragmentation; 기존 test job에 templates 디렉토리는 이미 포함될 가능성 → M5 W5에서 확인 후 docstring 추가만 | 2026-05-25 |
| D4 | §25 작성 시 본 SPEC ID 외 다른 SPEC ID 인용 금지 (AC-TII-011) | Doctrine 문서는 generic policy 명시 — 특정 SPEC 인용 시 §25 자체가 새 internal-content leak 도입; cross-reference는 §15/§21/§24 generic section 인용만 허용 | 2026-05-25 |
| D5 | 외부화 옵션 (B.4 Option B) 1차 미적용, default Option A 채택 | 35 files 중 외부화 필요 후보 식별이 W1 M1 단계에서 이루어짐; 사전 추측보다 evidence-based 결정 — M1 classification 결과에서 외부화 필요 file 발견 시 design.md amend | 2026-05-25 |

## §F. Cross-references

- spec.md §B (REQ-TII-001~013) — 본 design이 honor하는 요구사항
- plan.md §F (M1~M6) — 본 design을 적용할 milestone
- acceptance.md §B (AC-TII-001~012) — 본 design 결정이 검증되는 commands
- research.md §B (Predecessor cleanup pattern 분석) — substitution dictionary 도출 근거
