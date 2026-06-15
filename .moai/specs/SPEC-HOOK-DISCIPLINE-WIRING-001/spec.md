---
id: SPEC-HOOK-DISCIPLINE-WIRING-001
title: "Discipline Hook Wiring — Phase-2 Realization (status-transition + sync-gate, warn-first)"
version: "0.1.0"
status: implemented
created: 2026-06-11
updated: 2026-06-15
author: manager-spec
priority: P2
phase: "v0.2.0 target"
module: "internal/template/templates/.claude/settings.json.tmpl, .claude/hooks/moai/sync-phase-quality-gate.sh"
lifecycle: spec-anchored
tags: "hooks, settings, wiring, template-neutrality, language-detection, advisory"
tier: M
---

# SPEC-HOOK-DISCIPLINE-WIRING-001 — Discipline Hook Wiring (Phase-2 Realization)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-11 | manager-spec | Initial draft. Plan-phase artifacts authored (spec/plan/acceptance/design/research). Tier M. Realizes the pre-announced Phase-2 wiring deferred at the agent-team-rebuild milestone. |

---

## A. Background — Deferred-Wiring Realization (NOT a bug fix)

세 개의 "discipline hook" 스크립트는 이전 에이전트 팀 재구축 마일스톤에서 작성되었습니다. 이 훅들은 지금은 폐기된(archived) phantom 에이전트(`manager-quality` / `expert-security`)에게 위임되던 오케스트레이터 규율 의무를 기계적으로(mechanically) 강제하기 위한 것입니다:

- `.claude/hooks/moai/status-transition-ownership.sh` — PostToolUse, advisory (exit 0 only)
- `.claude/hooks/moai/sync-phase-quality-gate.sh` — Stop, blocking-capable (현재는 advisory로 한정)
- `.claude/hooks/moai/team-ac-verify.sh` — TaskCompleted, dormant/advisory

[HARD] 이 SPEC은 결함 수정(bug fix)이 **아닙니다**. 이것은 **사전에 예고된(pre-announced) Phase-2 wiring의 실현(realization)**입니다. 근거(verified by investigation):

1. 선행 SPEC(에이전트 팀 재구축)의 acceptance criteria는 훅 **파일 존재(file existence) + 내부 정확성(syntax)**에만 결합되었고, settings.json 등록(registration)에는 결합되지 않았습니다. (선행 acceptance.md의 sync-gate AC pass criterion은 "syntax 통과 AND 스크립트 본문이 키워드를 포함" — 등록 여부 미검증.)
2. 선행 design 문서는 wiring을 명시적으로 "orchestrator responsibility / future SPEC or manual configuration"으로 기술했습니다.
3. 대조군 SPEC(DB-sync 계열)은 "SHALL register a PostToolUse matcher in settings.json.tmpl"을 **명시적으로** 요구했습니다. 본 영역에서 그 요구의 부재는 누락이 의도적이었음을 증명합니다.
4. `--skip-hook` 플래그는 훅이 wiring된 이후를 위한 런타임 opt-out으로 이미 설계되어 있습니다.

따라서 본 SPEC은 그 예고된 Phase-2 wiring을 실현합니다. 결함이 아니라 의도된 단계적 진행입니다.

### A.1 Confirmed ground-truth (재발견하지 말 것)

- 세 훅 파일은 local `.claude/hooks/moai/` 과 template mirror `internal/template/templates/.claude/hooks/moai/` 양쪽에 정적 `.sh`(템플릿 변수 없음)로 존재합니다 — byte-identical (3625 / 4379 / 3066 bytes).
- `settings.json.tmpl`은 이미 PostToolUse(`handle-post-tool.sh`, matcher `Write|Edit`, async), Stop(`handle-stop.sh` + 조건부 `handle-harness-observe-stop.sh`), TaskCompleted(`handle-task-completed.sh`) 이벤트 블록을 보유합니다. 신규 훅은 이 배열에 **항목을 추가(ADD)**하며, advisory exit-0 훅은 기존 핸들러와 공존합니다(conflict 없음).
- 세 discipline 스크립트는 self-contained shell(jq/grep/git/언어 도구)이며 `moai hook <event>`를 호출하지 **않습니다** — 위임 wrapper가 아니라 최종 핸들러입니다. 따라서 wiring = `.sh`를 직접 가리키는 command 엔트리 추가이며, 새로운 `handle-*.sh` wrapper 생성이 아닙니다.
- `sync-phase-quality-gate.sh`는 `go vet` / `go test` / `golangci-lint`(Go 전용)를 하드코딩합니다. 16개 언어 템플릿에 현재 상태로 배포하면 중립성(neutrality) 위반입니다 (Python/Rust/JS 프로젝트에서 no-op이거나 깨짐).

---

## B. Problem Statement

세 discipline 훅은 정확하게 작성되어 있으나 `settings.json.tmpl`에 등록되지 않아 **휴면(dormant)** 상태입니다 — 어떤 이벤트에서도 실행되지 않습니다. 결과적으로, archived phantom 에이전트가 수행하던 오케스트레이터 규율(status-transition ownership 감사, sync-phase 품질 게이트)이 기계적으로 강제되지 않습니다.

추가로, `sync-phase-quality-gate.sh`는 Go 전용 도구를 하드코딩하므로, 그대로 등록·배포하면 비-Go 프로젝트(16개 언어 배포 대상)에서 훅이 깨지거나 무의미해집니다 — 템플릿 언어 중립성 원칙 위반.

---

## C. Goal

진보적(progressive), warn-first wiring을 실현합니다:

1. `status-transition-ownership.sh`를 settings.json.tmpl PostToolUse에 등록(advisory, exit 0 보존 — 감사 로그만, blocking 없음). 이 훅은 SPEC frontmatter status를 검사하므로 이미 언어 중립적입니다.
2. `sync-phase-quality-gate.sh`를 (a) 먼저 하드코딩된 Go 도구를 CLAUDE.md §7 언어 자동 감지 매트릭스로 일반화한 뒤, (b) settings.json.tmpl Stop에 등록. 본 SPEC에서는 **advisory/warn-first**로 한정 — blocking(exit 2) 경로는 후속 SPEC으로 명시적으로 연기됩니다.
3. `team-ac-verify.sh`는 wiring에서 **제외**됩니다 (Go `task_completed.go` 핸들러가 이미 더 강력한 SPEC-AC 검증을 exit-2로 수행하므로 advisory+dormant+기능적 중복). 파일은 삭제하지 않고 등록만 하지 않습니다.

---

## D. Requirements (GEARS)

### REQ-HDW-001 — status-transition hook registration (Ubiquitous)

The settings.json.tmpl **shall** register `.claude/hooks/moai/status-transition-ownership.sh` as a PostToolUse hook entry that coexists with the existing `handle-post-tool.sh` entry in the same PostToolUse hooks array.

- 근거(motivation): 등록되지 않은 훅은 휴면 상태이므로 status-transition ownership 감사가 실행되지 않는다. 기존 핸들러와의 공존이 보장되어야 회귀가 없다.

### REQ-HDW-002 — status-transition advisory preservation (Ubiquitous)

The `status-transition-ownership.sh` hook **shall** remain advisory (exit 0 on every path that is not `--skip-hook`); the registration **shall not** introduce any blocking (exit 2) behavior.

- 근거: PostToolUse는 blocking 불가 이벤트이며, 이 훅은 감사 로그(audit-log) 전용으로 설계되었다. blocking 도입은 본 SPEC scope 밖이다.

### REQ-HDW-003 — sync-gate language generalization (Ubiquitous)

The `sync-phase-quality-gate.sh` hook **shall** detect the project language from canonical project markers and run the matching toolchain per the CLAUDE.md §7 language-auto-detect matrix (Go / Node.js / Python / Rust), replacing the currently hardcoded Go-only `go vet` / `go test` / `golangci-lint` invocations.

- 근거: 16개 언어 템플릿 배포 대상에서 Go 전용 하드코딩은 중립성 위반이다. 언어 자동 감지로 일반화해야 모든 사용자 프로젝트에서 의미 있게 동작한다.

### REQ-HDW-004 — graceful skip on absent tooling (Event-driven)

**When** the detected language's toolchain is absent (the required CLI is not on `PATH`), the `sync-phase-quality-gate.sh` hook **shall** skip that toolchain step gracefully (exit 0, no error) rather than failing the Stop event.

- 근거: CLAUDE.md §7 — "tools that are not installed are skipped gracefully." 도구 부재가 사용자 세션을 중단시켜서는 안 된다.

### REQ-HDW-005 — silent pass on no language marker (Event-driven)

**When** no recognized language marker is present in the project root (no `go.mod` / `package.json` / `pyproject.toml` / `requirements.txt` / `Cargo.toml`), the `sync-phase-quality-gate.sh` hook **shall** pass silently (exit 0) without running any toolchain.

- 근거: CLAUDE.md §7 — "projects with no recognized language marker pass the gate silently." 언어 마커가 없는 프로젝트(문서 전용 등)는 게이트를 무음 통과해야 한다.

### REQ-HDW-006 — sync-gate registration warn-first (Ubiquitous)

The settings.json.tmpl **shall** register the (language-generalized) `sync-phase-quality-gate.sh` as a Stop hook entry that coexists with the existing `handle-stop.sh` and conditional `handle-harness-observe-stop.sh` entries, and the registered configuration **shall** operate advisory/warn-first within this SPEC (the exit-2 blocking path is deferred — see §Out-of-Scope).

- 근거: warn-first는 점진적 도입(progressive wiring) 원칙이다. blocking을 즉시 켜면 sync 작업이 예기치 않게 차단될 위험이 있으므로 후속 SPEC으로 연기한다.

### REQ-HDW-007 — team-ac-verify exclusion (Ubiquitous, unwanted-behavior)

The settings.json.tmpl **shall not** register `team-ac-verify.sh` in any event block; the hook file **shall** remain present on disk (not deleted).

- 근거: Go `task_completed.go` 핸들러가 이미 더 강력한 SPEC-AC 검증을 exit-2로 수행하므로 team-ac-verify.sh는 기능적으로 중복이다. 파일 삭제는 향후 재활성화 여지를 없애므로 보존한다.

### REQ-HDW-008 — platform-conditional quoting (Ubiquitous)

The new settings.json.tmpl hook entries **shall** follow the existing platform-conditional quoting pattern (`{{- if eq .Platform "windows"}} bash "$CLAUDE_PROJECT_DIR/..." {{- else}} "$CLAUDE_PROJECT_DIR/..." {{- end}}`) and use the `$CLAUDE_PROJECT_DIR` path variable, matching every existing wrapper entry.

- 근거: 기존 패턴과의 정합은 cross-platform(Windows bash vs Unix direct) 정확성과 렌더링 일관성을 보장한다.

### REQ-HDW-009 — template neutrality (Ubiquitous)

The edited settings.json.tmpl and the edited `sync-phase-quality-gate.sh` **shall** contain zero internal SPEC IDs, internal REQ/AC tokens, internal paths, or audit citations such that the template-neutrality CI guard (`internal/template/internal_content_leak_test.go` + `template_neutrality_audit_test.go`) passes.

- 근거: `internal/template/templates/**` 산출물은 16개 언어 사용자에게 배포되는 범용 자산이다. 내부 흔적 누출은 CI guard로 차단된다 (CLAUDE.local.md §15/§25).
- **중요 (D3 — CI guard scope 정정)**: 이 neutrality CI guard는 **내부 콘텐츠 누출(internal-content leak)만** 탐지한다. 그 탐지 클래스는 C1 macOS-bias-path / C2 V3R-sigil / C4 feedback-memory-ref / C5 CLAUDE.local-ref / C6 PR#-ref / C7 SHA / C8 GOOS-preserve / spec-id-date 이며, **언어 편향(Go-bias / language-bias) 탐지 클래스는 존재하지 않는다**. 증거: 현재(일반화 전) `sync-phase-quality-gate.sh`가 `go vet`를 4회 하드코딩하고 있음에도 neutrality guard는 오늘 GREEN으로 통과한다. 따라서 이 CI guard 통과는 16개 언어 중립성을 **증명하지 못한다**. 언어 중립성(Go-bias 부재)은 **별도로** AC-HDW-002(static marker grep + 실 git-repo non-Go fixture 런타임 검증)와 AC-HDW-009(Go-tool 토큰이 Go case branch 내부에만 존재하는지 case-block boundedness 검증)로 입증한다 — 이 두 AC가 central neutrality risk에 대한 실제 자동 guard다.

### REQ-HDW-010 — local↔template parity (Ubiquitous)

**Where** `make build` regenerates embedded assets, the local `.claude/settings.json.tmpl`-rendered output and the template source **shall** remain in parity (the wiring addition applied to both the template source and the local git-tracked `.claude/settings.json`), and the wiring addition **shall not** disturb the dev-intent keys (`defaultMode`, `env.PATH`, `teammateMode`) documented in CLAUDE.local.md §22.

- 근거: 로컬 `.claude/settings.json`은 git-tracked이며 의도적 dev-intent 키를 보유한다. wiring은 hook 배열 엔트리만 ADD해야 하며 dev-intent 키를 건드리면 안 된다.

---

## Exclusions (What NOT to Build)

[HARD] 본 SPEC은 아래를 **포함하지 않습니다**:

1. **Blocking / exit-2 경로 (DEFERRED)** — `sync-phase-quality-gate.sh`의 blocking(exit 2) 강제 경로는 본 SPEC에서 활성화하지 않습니다. 본 SPEC은 advisory/warn-first만 배송합니다. exit-2 게이트(품질 실패 시 Stop 차단)는 명시적으로 **후속 SPEC**으로 연기됩니다. 선행 SPEC의 aspirational "exits with code 2" 기준은 본 SPEC에서 강제되지 않습니다.

2. **`team-ac-verify.sh` 등록 (EXCLUDED)** — TaskCompleted용 `team-ac-verify.sh`는 wiring 대상에서 제외됩니다. 근거: Go `task_completed.go` 핸들러가 이미 SPEC-AC 검증을 exit-2로 수행하여 기능적으로 더 강력하고, team-ac-verify.sh는 advisory+dormant 중복입니다. **파일은 삭제하지 않고 보존**하되 어떤 이벤트 블록에도 등록하지 않습니다.

3. **비-Go 언어 도구 설치 (OUT OF SCOPE)** — `sync-phase-quality-gate.sh`의 언어 일반화는 도구 **감지(detect)와 우아한 skip**만 구현합니다. eslint / ruff / cargo / npm 등 비-Go 도구의 **설치(installation)**는 본 SPEC scope 밖이며, 도구 부재 시 graceful skip(REQ-HDW-004)으로 처리됩니다.

4. **신규 `handle-*.sh` wrapper 생성** — discipline 스크립트는 self-contained 최종 핸들러이므로 위임 wrapper를 새로 만들지 않습니다. wiring은 `.sh`를 직접 가리키는 command 엔트리 추가입니다.

5. **Go `post_tool.go` / `stop.go` / `task_completed.go` 핸들러 수정** — discipline 훅의 로직은 Go 핸들러와 별개입니다. 기존 Go 핸들러는 수정하지 않습니다.

6. **Coverage regression 차단 / coverage 델타 baseline 파일** — `sync-phase-quality-gate.sh`의 coverage 측정은 advisory(현행 그대로)로 유지하며, baseline 비교 기반 regression 차단은 구현하지 않습니다.

---

## E. Cross-References

- CLAUDE.md §7 "Language-Specific Guidelines" — 언어 자동 감지 매트릭스 (canonical)
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface — 3개 discipline 훅의 trigger/owning-REQ/exit-code 표
- `.claude/rules/moai/core/hooks-system.md` — Hook 이벤트 / settings.json 구성 패턴
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — status-transition-ownership.sh가 참조하는 canonical 매트릭스
- CLAUDE.local.md §15 (언어 중립성), §22 (dev settings intent), §25 (Template Internal-Content Isolation)
- `internal/template/internal_content_leak_test.go`, `template_neutrality_audit_test.go` — neutrality CI guard
