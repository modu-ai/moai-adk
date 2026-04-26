---
id: SPEC-V3R3-PATTERNS-001
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: High
labels: [patterns, harness, plan, v3r3]
issue_number: null
---

# SPEC-V3R3-PATTERNS-001 — Implementation Plan

## 1. Strategy Overview

3-Wave 작업으로 분해. Wave 1은 prior art 분석 (read-only), Wave 2는 6 rule 파일 작성 (write), Wave 3은 Template-First 동기화 + 회귀 테스트.

| Wave | Goal | Agent | Mode | Isolation |
|------|------|-------|------|-----------|
| Wave 1 | harness clone + 6 reference 분석 + import 매핑 | Explore (general-purpose, mode: plan) | Read-only | None |
| Wave 2 | 6 rule 파일 + NOTICE 작성 (markdown artifact) | manager-tdd (single sub-agent) | acceptEdits | None |
| Wave 3 | Template mirror 동기화 + `make build` + Go test 회귀 | manager-git (sub-agent) | acceptEdits | None |

**Why no team mode?** 산출물이 markdown rule 파일 6개로 file count < 10, domain count = 1 (rules). 팀 모드 임계치 (`domains>=3, files>=10`) 미달. Sub-agent 시퀀셜이 token 효율적.

---

## 2. Milestones (Priority-based, no time estimates)

### M1 (Priority High) — Prior Art Analysis

- harness 저장소 clone (HTTPS, depth=1) `/tmp/harness-analysis/harness/`
- 6 reference docs 위치 확인: `skills/harness/references/{agent-patterns,boundary-verification,skill-ab-testing,team-pattern-cookbook,orchestrator-templates,skill-writing-craft}.md`
- 각 reference 핵심 섹션 추출 (Explore agent가 read-only 분석)
- 라이선스 파일 (LICENSE) 확인 → Apache 2.0 검증
- import 매핑 표 작성 (source section → target section)

**Exit criteria**: 6 reference 모두 읽기 완료, Apache 2.0 LICENSE 사본 확보, import 매핑 표 progress.md에 기록.

### M2 (Priority High) — 6 rule 파일 + NOTICE 작성

- `.claude/rules/moai/quality/` 디렉토리 생성
- 6 파일 작성 (manager-tdd가 markdown artifact 작성):
  1. `development/agent-patterns.md`
  2. `quality/boundary-verification.md`
  3. `development/skill-ab-testing.md`
  4. `workflow/team-pattern-cookbook.md`
  5. `development/orchestrator-templates.md`
  6. `development/skill-writing-craft.md`
- 각 파일 frontmatter (paths, description) + 상단 attribution 주석 추가
- `NOTICE.md` 작성 (6 source 일람 + Apache 2.0 텍스트 사본)

**Exit criteria**: 6 파일 + NOTICE 모두 작성, frontmatter `paths` 글로브 검증, attribution 주석 grep 확인.

### M3 (Priority Medium) — Template-First 동기화

- 7 파일 모두 `internal/template/templates/.claude/rules/moai/`로 미러
- `make build` 실행 → `internal/template/embedded.go` 재생성
- `go test ./internal/template/...` 회귀 확인
- `internal/template/commands_audit_test.go` 통과 확인

**Exit criteria**: byte-identical mirror, build 성공, all template tests green.

### M4 (Priority Medium) — Validation

- frontmatter `paths` 글로브 패턴 실제 매칭 검증 (테스트 파일 수정 시 boundary-verification.md auto-load 확인)
- 16-language neutral 검증: 6 파일 grep으로 특정 언어 hard-coding 확인 (예: `python`/`go`/`javascript` 편향 검사)
- AC-001 ~ AC-005 (acceptance.md) 모두 PASS

**Exit criteria**: 모든 AC PASS, plan-auditor (있다면) PASS verdict.

---

## 3. Technical Approach

### 3.1 Cookbook Structure

각 rule 파일은 다음 구조:

```markdown
<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->
---
description: <one-line — when this cookbook applies>
paths: "<glob>"
---

# <Title>

> Imported from revfactory/harness Apache 2.0 (date: 2026-04-26).
> See `.claude/rules/moai/NOTICE.md` for license terms.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0.0 | 2026-04-26 | manager-spec | Initial import from harness `skills/harness/references/<file>.md` |

## <Section 1>
...
```

### 3.2 Apache 2.0 → MIT 호환성

MoAI-ADK는 MIT 라이선스. Apache 2.0 코드를 MIT 프로젝트에 흡수하는 것은 허용되며 다음 조건 충족:

- Apache 2.0 LICENSE 텍스트 사본 보존 → `NOTICE.md` 또는 `LICENSE-APACHE.txt`
- 원본 attribution 명시 → 각 파일 상단 주석 + NOTICE.md
- 변경사항 표기 → "Imported from harness on 2026-04-26, no modifications" 또는 변경 시 변경 내역 기록
- NOTICE 파일 보존 의무 (harness가 NOTICE 파일을 가지고 있다면 사본 포함)

### 3.3 Frontmatter `paths` 활용 패턴

Claude Code rules 시스템은 `paths` 글로브에 파일 매칭 시 해당 rule을 conditional load. 이를 활용해:

- `agent-patterns.md`는 agent 작성 중에만 로드 (`paths: ".claude/agents/**"`)
- `skill-ab-testing.md`는 skill 작성 중에만 로드 (`paths: ".claude/skills/**"`)

→ 토큰 효율 + 적절한 컨텍스트.

### 3.4 16-Language Neutrality 보장

원본 harness 문서는 일부 코드 예시를 포함할 수 있음. 흡수 시:

- 코드 예시가 특정 언어인 경우 → 의사코드(pseudocode)로 변환 또는 multi-language 변형 추가
- 도구명/패키지명만 언급 (예: `pytest` 단독 → "test framework (e.g. pytest, jest, go test)")
- file extension 가정 금지 (예: `.py` → `<source file>`)

---

## 4. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| harness LICENSE가 실제로 Apache 2.0이 아닌 경우 | Low | Critical | M1에서 LICENSE 직접 확인 → 다른 라이선스면 SPEC abort, 사용자 협의 |
| harness reference docs 일부가 부재 | Medium | High | M1에서 6 파일 모두 존재 확인 → 누락 시 abort 또는 사용자 협의로 대체 |
| 흡수 후 16-language neutrality 위반 | Medium | Medium | M2에서 작성 시 grep 자가검증 + M4에서 validation grep |
| Template mirror 누락 | Low | High | M3 이후 `find` 비교 + `diff` 확인 |
| frontmatter `paths` 글로브 매칭 실패 | Low | Medium | M4에서 실제 파일 수정 후 rule 로딩 확인 |
| Apache 2.0 NOTICE 누락 | Low | Critical | M2에서 NOTICE.md 별도 작성 + grep 검증 |

---

## 5. Dependencies

- [BLOCKING] git CLI: harness 저장소 clone에 필요 (이미 환경에 존재 확인)
- [BLOCKING] make: `make build` 실행 (이미 Makefile 존재 확인)
- [BLOCKING] go: `go test ./internal/template/...` 실행
- [SOFT] SPEC-V3R3-DEF-007 (Convention Sweep) — frontmatter 베이스라인 정의된 상태에서 시작 가능 (이미 implemented)
- [SOFT] SPEC-V3R3-ARCH-003 (Expert tool uplift) — agent body 일관성 baseline (이미 완료)

---

## 6. Open Questions

| ID | Question | Resolution Path |
|----|----------|-----------------|
| OQ-001 | NOTICE 파일 위치: `.claude/rules/moai/NOTICE.md` vs 프로젝트 root `NOTICE` vs `LICENSE-third-party.md`? | M1에서 결정. 권장: `.claude/rules/moai/NOTICE.md` (cookbook 인접) + `LICENSE-third-party.md` (root 표준) 둘 다. |
| OQ-002 | harness reference 파일이 실제 6개 모두 존재하는가? | M1 clone 후 `ls skills/harness/references/` 확인. |
| OQ-003 | 16-language neutrality 위반 시 의사코드 변환 vs multi-variant? | M2 작성 시 case-by-case. 짧은 예시는 의사코드, 긴 예시는 multi-variant. |
| OQ-004 | frontmatter `paths` 글로브가 실제 자동 로딩되는지 검증 메커니즘은? | M4에서 실제 파일 touch 후 rule 로딩 로그 확인. Claude Code log/debug mode 활용. |

---

## 7. Out of Scope (Reaffirmation)

- harness skill / agent 본체 흡수
- 한국어 번역
- 기존 rule 파일 수정 (frozen content)
- harness 저장소 upstream sync 자동화
