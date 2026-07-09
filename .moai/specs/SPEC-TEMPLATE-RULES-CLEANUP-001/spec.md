---
id: SPEC-TEMPLATE-RULES-CLEANUP-001
title: "Template-distributed rules cleanup: broken refs, neutrality, backports, retired vocab, design drift + CI guard expansion"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/template/templates/.claude/rules"
lifecycle: spec-anchored
tier: L
tags: "template, rules, neutrality, ci-guard, cleanup, mirror-parity, backport"
---

# SPEC-TEMPLATE-RULES-CLEANUP-001 — 템플릿 배포 rules 정리 + CI 가드 확장

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-09 | 0.1.0 | 최초 작성 — 3-렌즈 감사(A~E finding) 반영, 사용자 결정 4건 고정 요구사항으로 인코딩 | GOOS행님 (manager-spec) |

## §A 개요 및 배경

`internal/template/templates/.claude/rules/` 하위 62개 템플릿 배포 rule 파일에 대한 3-렌즈 감사(2026-07-09, session 98a5197a; 본 세션에서 현재 트리 대비 재검증 완료)가 5개 finding 그룹을 확인했다. 전량이 현행 CI 가드(`go test ./internal/template/ -run 'Mirror|Leak|Neutral|Language'` — 전부 PASS)의 사각지대다:

- **Finding A** — 사용자 프로젝트에 배포되지 않는 문서를 인용하는 깨진 cross-reference 5건 (HIGH: 배포 후 404)
- **Finding B** — 중립성 위반 4클래스 (REQ/AC 토큰 7건, provenance 프로즈, lessons #N/W# 11건, 내부 작업 날짜 9건, CONST-V3R* 토큰, zone-registry.md 전체)
- **Finding C** — local → template 미백포트 2건 (선별적 — §B Group C 참조)
- **Finding D** — retired 어휘 soft 위반 2건 (Sprint Contract, cohort)
- **Finding E** — design 역드리프트 (`bda9a3f33`이 template만 갱신, local 미완료)

사용자 결정 4건(고정, 재질문 불가): (1) Finding E는 local에서 제거 완결, (2) zone-registry.md는 템플릿 배포에서 제거(local dev 사본 유지) + 참조 정리, (3) CI 가드 4종 전량 확장, (4) 한국어 콘텐츠 클래스는 후속 SPEC으로 명시 이연.

상세 증거·재검증 기록은 `research.md`, 가드 설계는 `design.md` 참조.

## §B 요구사항 (GEARS)

### Group P — 공통 프로세스 요구사항

- **REQ-TRC-001** (Ubiquitous): The rewritten template rule files shall preserve their normative semantics — a rewrite drops only unshipped path citations and internal provenance, never normative content. The per-file `[HARD]` marker count shall remain unchanged for every edited template rule file, with two documented exceptions: (a) `design/constitution.md` retired-annotation of the Sprint Contract block (REQ-TRC-040), (b) removal of `zone-registry.md` itself (REQ-TRC-024).
- **REQ-TRC-002** (While): **While** run-phase editing is in progress, every edit shall follow the per-file tree-classification treatment matrix (research.md §R7): (i) BYTE-PARITY ENROLLED files (`spec-workflow.md`, `session-handoff.md` — `rule_template_mirror_test.go` allowlist) receive identical dual-tree edits; (ii) SANITIZED PAIR files (`manager-develop-prompt-template.md`, `runtime-recovery-doctrine.md` — `sanitized_pair_parity_test.go` registry) receive template-side-only edits; (iii) UNENROLLED pairs receive identical dual-tree edits by default, with per-file exceptions where the local copy legitimately cites dev-only artifacts (decision recorded in progress.md). Template-First 순서(template 편집 → `make build` → local 반영)를 준수한다.
- **REQ-TRC-003** (Ubiquitous): The pre-existing guard suite (`go test ./internal/template/ -count=1`) shall remain green at every push boundary — including `TestLeakClassNoDateShaInDefaultTier`, `TestSanitizedPairParity`, and `TestRuleTemplateMirrorDrift`.

### Group A — 깨진 cross-reference (5건)

- **REQ-TRC-010** (Ubiquitous): The template rule tree shall contain zero path citations to the four unshipped documents: `.moai/docs/dev-only-commands-isolation.md`, `.moai/docs/git-local-workflow-doctrine.md`, `.moai/docs/git-workflow-doctrine.md`, `.moai/design/v3-redesign/synthesis/pattern-library.md`.
- **REQ-TRC-011** (When): **When** a rewrite removes an unshipped path citation, the rewrite shall inline the essential normative content as prose OR repoint to a shipped equivalent document — never delete the surrounding rule clause.
- **REQ-TRC-012** (Ubiquitous): The template `development/agent-authoring.md` shall not reference the retired `97-*`/`98-*` command-wrapper naming (superseded by the split-harness structure); the `.claude/agents/local/` protection contract (`moai update`가 절대 삭제/수정하지 않음)은 rewrite 후에도 보존되어야 한다.

### Group B — 중립성 위반 정리

- **REQ-TRC-020** (Ubiquitous): The template rule tree shall contain zero non-allowlisted tokens matching `\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b`. 확인된 7건 — `askuser-protocol.md:177` (AC-ADM-005..017), `sprint-round-naming.md:23,94` (AC-LR-009), `settings-management.md:181,344,355,364` (REQ-MIG003-006, REQ-WF006-006/015/011) — 을 중립 프로즈로 치환한다. Pedagogical placeholder(`REQ-XXX-*`/`AC-XXX-*`, 예: `manager-develop-prompt-template.md:175`)는 allowlist 대상이다.
- **REQ-TRC-021** (Ubiquitous): The template `core/askuser-protocol.md` Worked Example shall carry no internal provenance phrases (`Epic 7`, `TMC-001`, `L51`, `§24 namespace align`) — the pedagogical structure (preview key-set comparison example) shall be preserved with generic rationale text.
- **REQ-TRC-022** (Ubiquitous): The template rule tree shall contain zero `lessons #N` references and zero incident-provenance `W#` tokens (`W3 meta-analysis`, `lessons #21 W0`, `W1/W2`, `W3 케이스` 등 — 감사 확인 11건: `manager-develop-prompt-template.md:71,91,132,221,222`, `agent-common-protocol.md:274,409-413`, `session-handoff.md:291,332`, `spec-workflow.md:51`). 치환은 교훈의 내용을 일반화 프로즈로 유지한다.
- **REQ-TRC-023** (Ubiquitous): The template rule tree shall contain zero internal work-date provenance strings. 확인 9건: `design/constitution.md:12-15,423-424` (HISTORY/footer 날짜), `worktree-integration.md:44,46` + `spec-workflow.md:175,269,416` + `session-handoff.md:330` (반복 "2026-05-17 policy"), `zone-registry.md:629-630` (제거로 소멸). 날짜가 실린 정책 문장은 날짜만 제거하고 정책 내용은 유지한다.
- **REQ-TRC-024** (When): **When** `moai init` / `moai update` deploys the template tree, the deployed rule set shall not include `core/zone-registry.md` (내부 거버넌스 레지스트리 — 1,019줄 중 CONST-V3R* 121건). The local dev copy (`.claude/rules/moai/core/zone-registry.md`) shall be retained untouched.
- **REQ-TRC-025** (Ubiquitous): The template tree (`internal/template/templates/` 전체) shall contain zero references to `zone-registry` after removal — 확인된 참조 7줄 / 2파일 (`design/constitution.md` 5곳, `workflow/runtime-recovery-doctrine.md` 2곳)을 파일명 비인용 프로즈로 재서술한다.
- **REQ-TRC-026** (Ubiquitous): The template rule tree shall contain zero `CONST-V3R*` tokens (zone-registry 제거 후 잔여: `manager-develop-prompt-template.md` 4줄, `worktree-integration.md` 2곳) and zero standalone internal migration-ID headings (`settings-management.md:174` `MIG-003`).
- **REQ-TRC-027** (When): **When** a deployed user project lacks `zone-registry.md`, the `moai constitution list`, `moai doctor`, and `moai spec lint` commands shall degrade gracefully (non-crash, informative message). 런타임 소비자 3곳 확인: `internal/cli/constitution.go:24`, `internal/cli/doctor.go:580` (이미 not-found 메시지 처리), `internal/cli/spec_lint.go:159` (`detectRegistryPath`가 빈 문자열 반환 — graceful). Run-phase에서 scratch 배포 프로젝트로 실측 검증하고, 비-graceful 동작이 발견되면 최소 Go 수정으로 graceful 처리를 추가한다.

### Group C — 백포트 (local → template, 선별)

- **REQ-TRC-030** (Ubiquitous): The template `development/spec-frontmatter-schema.md` shall include the `§I Token Accounting` progress.md section-map row present in the local version (내용 중립 — SPEC ID 없음).
- **REQ-TRC-031** (Ubiquitous): The template `workflow/runtime-recovery-doctrine.md` shall carry the local tree's product-name terminology corrections (`MoAI` → `moai-adk`, 확인 ~8곳) and the local tree's more-neutral phrasings where they exist (예: "The sibling interrupt-ledger SPEC owns" → "The orchestrator-interrupt-ledger contract owns").
- **REQ-TRC-032** (Ubiquitous, unwanted): The backport shall NOT copy sanitization-stripped elements into the template mirror: the `.moai/research/dive-into-claude-code-archive.md` cross-reference, the `Version:`/`Origin:` footer lines, and the `CONST-V3R6-001` token — 이들의 template 부재는 sanitized-pair 계약(`sanitized_pair_parity_test.go` + leak test)에 의한 의도적 결과다 (research.md §R2).

### Group D — retired 어휘

- **REQ-TRC-040** (Ubiquitous): The template `design/constitution.md` Sprint Contract block (lines 323-350) shall carry a retired-historical annotation consistent with the file's existing RETIRED banner, and shall contain no un-annotated active-voice `[HARD]` clause citing the `.moai/sprints/` path. The literal `.moai/sprints` path string shall not remain in the template rules tree (config 키 `design.yaml gan_loop.sprint_contract.artifact_dir`는 라이브 Go 코드(`loader_design.go`)가 소비하므로 rename 대상이 아니며 본 SPEC 범위 밖이다 — rules 프로즈만 처리).
- **REQ-TRC-041** (Ubiquitous): The template `workflow/orchestration-mode-selection.md` shall not use the standalone retired term `cohort` (line 190 "SPEC cohorts" → 정식 분류 어휘로 치환, 예: "across SPECs" / "SPEC 집합"). SSOT: `.claude/rules/moai/development/sprint-round-naming.md` (AP-SRN-005).

### Group E — design 역드리프트 local 완결 (사용자 결정 1 고정)

- **REQ-TRC-050** (Ubiquitous): The local `.claude/rules/moai/design/constitution.md` shall be byte-identical to the post-cleanup template version (M5의 template 측 D 편집 이후 동기화 — 21,766B stale 사본을 갱신본으로 교체).
- **REQ-TRC-051** (Ubiquitous): The local `.moai/config/sections/design.yaml` shall contain no `design_docs:` block (`bda9a3f33`이 template/Go 양측에서 제거한 stale remnant), and its top-level key set under `design:` shall match the template `design.yaml` key set. 파일 삭제가 아니다 — `LoadDesignConfig`(`internal/config/loader_design.go`)가 라이브 소비하고 template이 여전히 배포하는 활성 config다 (research.md §R6).

### Group F — CI 가드 확장 (사용자 결정 3: FULL scope 4종)

- **REQ-TRC-060** (Where): **Where** the generalized REQ/AC-token guard is active, the guard shall detect any `\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b` token in template `.claude/rules/` files not covered by a pedagogical allowlist entry. 구현은 기존 `S3-req-ac-token-any-prefix` leak class(현재 `skillBodyScoped: true`, "agents/rules are owned elsewhere")의 rules 스코프 확장 또는 rules-스코프 sibling class 추가로 한다 (design.md §2).
- **REQ-TRC-061** (Where): **Where** the provenance guard is active, the guard shall detect `lessons #N` and incident-`W#` provenance patterns in template `.claude/rules/` files (RE2 호환 패턴 — lookahead 불가; false-positive 완화는 design.md §3).
- **REQ-TRC-062** (Where): **Where** the internal-date-provenance guard is active, the guard shall detect ISO work-date strings in template `.claude/rules/` files via a SEPARATE test function — NOT added to the default-tier `leakClasses` slice, preserving the tier-ownership contract enforced by `TestLeakClassNoDateShaInDefaultTier` (design.md §4).
- **REQ-TRC-063** (Where): **Where** the internal-governance-token guard is active, the `CONST-V3R*` / `SPEC-V3R*` detection scope shall extend from skill-body to template `.claude/rules/` files (기존 pattern line ~199 재사용, 스코프만 확장).
- **REQ-TRC-064** (When): **When** any new guard detects a violation, the test shall fail with an actionable finding naming the file path, line number, and matched token, plus a greppable sentinel string per guard class.
- **REQ-TRC-065** (While): **While** the content-cleanup milestones (M2-M6) are incomplete, the four new guards shall FAIL against the template tree (RED evidence — 각 가드 클래스가 기지 위반 ≥1건을 실명 검출한 출력이 progress.md §E.2에 기록됨); after cleanup completes, the guards shall PASS (GREEN). Push는 GREEN 이후에만 수행한다 (trunk CI 보호).
- **REQ-TRC-066** (When): **When** a future edit reintroduces a cleaned leak shape into a template rule file, the corresponding guard shall flag it — each new guard shall include a recurrence backstop self-test (synthetic re-leak probe fires, clean replacement passes; `TestSkillBodyLeakClassRecurrenceBackstop` 패턴 준용).

## §C 제외 범위

다음 항목은 본 SPEC의 out of scope다. 각 항목은 의도적 이연/거부이며, 여기 없는 암묵 확장(스킬 트리 재정비, config 키 rename 등)도 모두 범위 밖이다.

### Out of Scope — 한국어 콘텐츠 정규화 (사용자 결정 4: 후속 SPEC 이연)

- 템플릿 rules 15/62 파일의 한국어 콘텐츠(`manager-develop-prompt-template.md` 84줄, `askuser-protocol.md` 50줄, `session-handoff.md` 44줄 등)는 본 SPEC에서 다루지 않는다.
- 일부는 의도적 localization(cut-line 마커, ko worked example)이므로 파일별 판단이 필요하며 blind substitution은 금지 — 볼륨과 판단 비용상 후속 SPEC 소관.
- 본 SPEC의 M2-M6 편집이 한국어 문장을 만나면 provenance/토큰만 제거하고 언어는 바꾸지 않는다.

### Out of Scope — NOTICE.md import 날짜

- NOTICE.md의 서드파티 import 날짜류는 감사에서 판단 유보된 항목으로, 본 SPEC의 date-provenance 정리(REQ-TRC-023) 및 가드(REQ-TRC-062) 스코프에서 제외한다 (가드 스코프는 `.claude/rules/`로 한정).

### Out of Scope — design-system 제거 재검토 (거부)

- `bda9a3f33`의 design 시스템 제거 자체를 되돌리거나 재설계하는 것은 거부된 방향이다. 본 SPEC은 제거를 local에서 완결(Finding E)할 뿐이다.
- `design.yaml`의 `gan_loop.sprint_contract.*` 등 라이브 config 키 rename / `LoadDesignConfig` 제거도 범위 밖이다.

### Out of Scope — zone-registry 사용자-대면 대체 메커니즘

- 템플릿 제거 후 사용자 프로젝트를 위한 constitution registry 대체 설계(`moai constitution list`의 사용자-프로젝트 스토리 재설계)는 범위 밖이다. 본 SPEC은 graceful degradation 검증(REQ-TRC-027)까지만 보장한다.

### Out of Scope — skills 트리 중립성

- `.claude/skills/` 트리는 기존 SBN 가드 체계가 소유하며, 본 SPEC의 가드 확장 스코프는 `.claude/rules/`로 한정한다. skills 측 잔여 위반은 별도 소관.

## §D 성공 기준 요약

1. 배포 rule 트리에서 4개 unshipped 경로 인용 0건, 비-allowlist REQ/AC 토큰 0건, lessons/W# 0건, 내부 날짜 0건, CONST-V3R*/MIG-003 0건, zone-registry 파일·참조 0건.
2. 백포트 2건 반영 + sanitized-pair 계약 비침해.
3. retired 어휘 2건 해소.
4. local design 역드리프트 완결 (constitution.md byte-parity, design.yaml key-set 정렬).
5. 신규 가드 4종: pre-cleanup RED 증거 → post-cleanup GREEN, 기존 가드 스위트 무회귀, recurrence backstop 내장.
6. `moai spec lint` clean, `make build` 성공, 전체 `go test ./internal/template/` green 후에만 push.

기계 검증 명령 전체는 `acceptance.md` §D 참조.

## §E Cross-References

- `research.md` — 감사 증거 + 본 세션 재검증 기록 + 스코프 정제 R1-R9
- `design.md` — 가드 4종 상세 설계 (패턴/스코프/allowlist/오탐 완화/테스트 배치)
- `plan.md` — M1-M7 마일스톤, 트리 분류 매트릭스, 리스크
- `acceptance.md` — AC 매트릭스 (기계 검증 명령)
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — C1-C8 콘텐츠 클래스 SSOT
- `.claude/rules/moai/development/sprint-round-naming.md` — retired 어휘 SSOT
- `internal/template/rule_template_mirror_test.go` / `sanitized_pair_parity_test.go` / `internal_content_leak_test.go` — 트리 분류·가드 기존 계약
