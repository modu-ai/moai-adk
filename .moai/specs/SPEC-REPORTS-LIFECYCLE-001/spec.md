---
id: SPEC-REPORTS-LIFECYCLE-001
title: "Reports directory lifecycle: root reports/ relocation into .moai/reports/historical/, html-report skill default path fix, worktree evidence hoist, and .moai/reports archive policy"
version: "0.1.0"
status: in-progress
created: 2026-09-29
updated: 2026-09-29
author: manager-spec (card t1320)
priority: P2
phase: "v3.3.0 target"
module: "internal/cli/worktree, internal/cli, .claude/skills/moai-domain-html-report, internal/template/templates/.claude/skills/moai-domain-html-report, .gitignore, internal/template/templates/.gitignore"
lifecycle: spec-anchored
tags: "reports, evidence-lifecycle, archive-policy, worktree-hoist, gitignore, html-report, template-mirror"
tier: M
---

# SPEC: Reports directory lifecycle (root relocation + evidence hoist + archive policy)

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-29 | Initial draft — plan-phase artifact set (card t1320, Tier M) |

---

## §A Context

### A.1 측정된 기선 (basis: develop 6879cfa5e, 이 워크트리)

카드 본문의 수치와 이 트리의 실측이 몇 군데 다르다. 아래 표가 본 SPEC 의 기선이며, 카드 본문 수치와 다른 항목은 A.3 에 사유와 함께 적는다.

| 대상 | 카드 본문 | 본 트리 실측 (2026-09-29) | 판정 |
|---|---|---|---|
| 루트 `reports/` | 34 항목 / 7.3MB / 08-26~09-25 산출물 | **27 최상위 항목 / 9.2MB / 20260911~20260923** / **tracked 511 파일** (`git ls-files reports/`) | 실측 채택. 카드가 놓친 핵심: **루트 reports/ 는 git 추적 중이다** — 이동은 로컬 디렉터리 이동이 아니라 tracked 재배치(`git mv`)다 |
| `.moai/reports/` | 2,565 항목 / 564MB | 본 트리 421 항목 / 12MB | 카드 수치는 **primary 체크아웃의 머신 로컬 누적량**이다. 저장소 상태가 아니므로 아카이브 정책의 기준값으로 쓰지 않는다 |
| 워크트리 증거 오염 | 5 트리 | (위임자 재측정) **27 트리 × 8.9~12MB** | 위임자 측정 채택 — 본 워크트리 안에서는 재측정 불가(대상이 primary 의 `.moai/worktrees/`) |
| `.gitignore` 커버리지 (item 5) | "하위 디렉터리 미커버" 가정 | **이미 충족**: `.gitignore:235` `.moai/reports/*` (하위 포함) · `:306` `.moai/worktrees/` · `:399` `.moai/reports/*.md` / 템플릿 미러 `:267`·`:234` | 카드 관측은 main 브랜치의 primary 사본 기준 오독. REQ 를 회귀 가드로 재형상화 (§B REQ-RLC-008) |
| html-report skill 기본 경로 | SKILL.md:65·:74 `<cwd>/reports/` | **확인**: 로컬·템플릿 미러 양쪽 65행(`output_path` 기본값)·74행(Output 절)에 `<cwd>/reports/` 존재 | 카드 그대로 |
| 워크트리 done 의 hoist | "반영 필요" | **부재 + L1 거부 확인**: `done.go` 에 hoist 0건(grep, exit 1)이고, done 은 **L1 트리(`.moai/worktrees/` 포함)를 두 경로(auto `:80`, interactive `:284`)에서 모두 거부**(SPEC-WORKTREE-DONE-TIER-001, `--force` 로도 불가). `worktree-integration.md` 에 인출 의무 조항 0건 | RED 성립 + **기계 구조 정정**: 카드 트리는 전부 L1 이므로 done 에 hoist 를 심는 것만으로는 실행 경로가 없다 — REQ-RLC-005 를 독립 동사 경유로 재설계(plan §F.0 D2) |

### A.2 `.gitignore:399` 의 용도 — 조사 결과, 제거 대상 아님

`.moai/reports/*.md` (399행)는 235행 `.moai/reports/*` 뒤에 있어 겉보기엔 중복이다. 그러나 gitignore 는 **마지막 매칭 규칙이 이긴다**: 235행 뒤의 디렉터리 단위 negation 블록(`plan-audit`, `t338`, `t528`, `t530`, `t229`)이 미래에 어떤 `.md` 재포함 규칙을 얻더라도 399행이 그것을 **무조건 재배제**한다. 즉 399행은 "card t1059 로 철회된 verdict 재포함 예외가 패턴 차원에서 재삽입되는 것"을 막는 방어심층 가드다. 본 SPEC 은 이 행을 **보존**하고, 규칙 집합 전체를 회귀 가드의 대상으로 삼는다(REQ-RLC-008).

### A.3 카드 본문 정정 목록

1. **루트 reports/ 는 tracked** — 카드 본문의 "move" 를 무작정 `mv` 로 수행하면 511 파일이 remote 에서 사라진다(무시 규칙은 추적 중인 파일을 untrack 하지 않지만, untracked 이동은 애초에 remote 에 없던 상태가 된다). 정확한 동사는 `git mv` 이며 이력과 원격 존재가 보존된다. `.moai/reports/*` 무시 규칙과의 상호작용은 REQ-RLC-002 가 규정한다.
2. **item 5 는 이미 충족** — 신규 규칙 추가가 아니라 `git check-ignore` 행동 매트릭스를 고정하는 회귀 가드가 산출물이다.
3. **564MB / 2,565 항목은 머신 로컬량** — 아카이브 정책의 임계값은 이 수치가 아니라 연령 기반으로 설계한다(머신마다 다른 값에 정책을 붙이면 재현이 안 된다).

---

## §B Requirements (GEARS)

### 이동 (card item 1)

- **REQ-RLC-001** — **When** 이동 마일스톤이 착지하면, the repository shall contain no tracked entries under root `reports/` and shall contain the same set of tracked files under `.moai/reports/historical/` with relative paths preserved (측정 기준: 511 files @ 6879cfa5e), relocated via `git mv` so blob identity, file history, and remote presence are retained.
- **REQ-RLC-002** — **While** `.moai/reports/*` remains the governing ignore rule, the ruleset shall NOT re-include `.moai/reports/historical/` for future files: preservation rests on tracked-file continuity of the relocated set, and newly created files under `historical/` shall stay local-only (ignored).

### skill 기본 경로 (card item 2)

- **REQ-RLC-003** — The html-report skill shall declare `.moai/reports/<slug>-<YYYYMMDD>.html` as the `output_path` default and `.moai/reports/<slug>-<YYYYMMDD>.md` for the sidecar, in **both** the local skill (`.claude/skills/moai-domain-html-report/SKILL.md`) and the template mirror (`internal/template/templates/.claude/skills/moai-domain-html-report/SKILL.md`), committed in the same commit and followed by `make build` (catalog.yaml hash regeneration).
- **REQ-RLC-004** — **When** the resolved output directory does not exist, the skill's report-writing flow shall create it before writing either file.

### 워크트리 증거 인출 (card item 3)

- **REQ-RLC-005** — The CLI shall provide `moai worktree hoist <tree-path>` as an independent verb that copies the tree's `.moai/reports/` content to the project root's `.moai/reports/worktrees/<tree-name>/` preserving relative paths, prints the hoisted entry count and byte total, refuses paths outside the project root, and completes before any disposal of that tree. **When** `moai worktree done` removes an L2 tree whose `.moai/reports/` contains files, the CLI shall invoke the same hoist routine before removal. The disposal-flow documentation (`.claude/rules/moai/workflow/worktree-integration.md` 와 그 템플릿 미러) shall make hoist-before-dispose an explicit obligation of the session-end L1 disposal path — card trees under `.moai/worktrees/` 는 L1 이라 done 이 절대 제거하지 않으므로(SPEC-WORKTREE-DONE-TIER-001), 그 폐기 경로의 유일한 실행 메커니즘은 이 동사다.
- **REQ-RLC-006** — **When** the hoist destination already holds an entry at the same relative path with different content, the CLI shall never silently overwrite it: it shall keep the existing file and report every unresolved path (skip-and-report) in the command output.

### 아카이브 정책 (card item 4)

- **REQ-RLC-007** — **Where** `moai clean` is invoked with the reports-archive action, the CLI shall **move** (never delete) archive candidates into `.moai/reports/archive/<YYYY-MM>/`, print the moved count and byte total, and apply a **default-deny predicate**: a candidate is a top-level entry under `.moai/reports/` whose name matches the evidence-dir naming (`^t[0-9]+$` 또는 `^SPEC-[A-Z0-9-]+[0-9]{3}$`) **and** whose mtime is older than the retention window (default 90 days) **and** that contains no git-tracked files (`git ls-files` 비어 있음). Everything else is out of scope by construction. Explicitly protected — predicate 적중 여부와 무관하게 절대 이동 대상이 아니다: `.moai/reports/historical/`, `.moai/reports/plan-audit/`, `.moai/reports/worktrees/`(hoist 산출물), `.moai/reports/archive/`(자기 자신), 그리고 tracked 파일을 포함하는 모든 항목(`t338`/`t528`/`t530`/`t229` fixture 가 이 규칙으로 자동 보호된다).

### 가드와 배포 (card items 5-6)

- **REQ-RLC-008** — A guard test shall assert the reports/worktrees ignore matrix — local `.gitignore` 의 `.moai/reports/*`(:235)·`.moai/worktrees/`(:306)·`.moai/reports/*.md` 방어심층 행(:399)과 템플릿 미러의 대응 규칙 — 의 존재와 `git check-ignore` 행동을 검증하고, 규칙 삭제 또는 순서 변경으로 negation 이 방어심층을 이기게 되면 실패한다.
- **REQ-RLC-009** — The template mirror shall carry the same conventions (skill default path, ignore matrix, hoist/archive behavior shipped via the binary) with template content free of card provenance (no card ids, no internal dates, no SPEC ids, no commit SHAs), and `make build` shall pass its emit-drift checks after every template-mirror edit.

---

## §C Acceptance Criteria

Tier M — AC 본문은 `acceptance.md` §D 를 본다. 요약 매트릭스:

| AC | REQ | 판정 요지 |
|---|---|---|
| AC-RLC-001 | REQ-RLC-001 | `git ls-files reports/` → 0, `git ls-files .moai/reports/historical/` → 이동 전 511 과 동일 집합 |
| AC-RLC-002 | REQ-RLC-001 | blob SHA 총합 불변 + `git log --follow` 로 이동 전 이력 도달 |
| AC-RLC-003 | REQ-RLC-002 | check-ignore 매트릭스: 신규 파일은 무시, tracked 이행 파일은 추적 유지, `git status` 에 D 없음 |
| AC-RLC-004 | REQ-RLC-008 | 가드 테스트 통과 + 변이 주입(규칙 삭제) 시 실패 관측 |
| AC-RLC-005 | REQ-RLC-003, 009 | 양 미러에서 `<cwd>/reports/` 0건·`.moai/reports/` 존재 + `make build` 드리프트 클린 |
| AC-RLC-006 | REQ-RLC-003, 004 | output_path 미지정 렌더가 `.moai/reports/` 에 html+md 생성, 디렉터리 자동 생성 |
| AC-RLC-007 | REQ-RLC-005 | `moai worktree hoist` 가 증거 있는 트리의 내용을 루트 `.moai/reports/worktrees/<tree>/` 로 인출 + 건수/바이트 출력; L2 제거 경로(done)는 같은 루틴을 제거 전 호출 |
| AC-RLC-008 | REQ-RLC-006 | 동일 경로 상이 내용 → 미덮어쓰기 + 미해결 경로 보고 |
| AC-RLC-009 | REQ-RLC-007 | 술어 적합 후보(이름 패턴 + 창 초과 + tracked 0)가 move 로 아카이브 이동 + 건수/바이트 출력, 비적합 항목 무영향 |
| AC-RLC-010 | REQ-RLC-007 | historical/·plan-audit·worktrees/·archive/·tracked 포함 항목은 술어 적중 여부와 무관하게 무영향 |
| AC-RLC-011 | REQ-RLC-009 | 템플릿 미러 변경분에 카드 유래물 0건 (t1320 / SPEC id / 본 카드 날짜 / SHA grep) |
| AC-RLC-012 | REQ-RLC-005 | 폐기 플로우 문서(양쪽 사본)에 hoist-before-dispose 의무 조항 존재 |

---

## §D Constraints

1. **Template neutrality (C1-C8)** — `internal/template/templates/**` 변경분은 카드 유래물(card id, 내부 날짜, SPEC id, commit SHA)을 포함하지 않는다. CI 가드 `template-neutrality-check.yaml` 이 안전망이며 본 SPEC 의 AC-RLC-011 이 1차 검증이다.
2. **Template-First 순서** — skill 기본 경로 수정은 로컬·템플릿 미러를 **같은 커밋**에서 고치고 `make build` 로 임베드를 재생성한다(§2 규율). C3(`.codex` 방출)는 손으로 대상이 아니다 — 필요하면 `make agents-emit` 이 유일한 재생성 동사다.
3. **moai update 비대상 확인** — `CleanMoaiManagedPaths` 의 관리 대상 뿌리(`.moai/config` 등)에 `.moai/reports` 는 없다. 이동·아카이브된 증거가 update 에 지워지지 않는다는 것은 측정된 사실이며, plan 의 리스크 절이 이 근거를 인용한다.
4. **증거 디렉터리** — 본 카드의 판정 증거는 `.moai/reports/t1320/` (gitignored, 커밋 금지)에 둔다. plan-audit 보고서가 이 디렉터리에 기록된다면 첫 줄에 감사 모델을 명시한다.
5. **t1240 독립성** — 본 SPEC 의 어떤 마일스톤도 t1240(factory F2)의 착지에 의존하지 않는다. hoist 구현은 `internal/cli/worktree` 안에서 자기완결적이다.
6. **로컬 전체 스위트 금지** — 검증은 변경 패키지 한정(`go test -timeout 30m ./internal/<pkg>/...`), 전 패키지 판정은 CI 몫이다.
7. **시간 예측 금지** — 마일스톤은 우선순위 라벨과 순서로만 표기한다.

---

## §E Exclusions

### Out of Scope — 기존 증거의 소급 아카이브

- 이미 90일을 넘은 기존 `.moai/reports/` 디렉터리들의 일괄 소급 이동은 run phase 의 산출물이 아니다. 정책(명령)만 착지시키고, 실제 소급 실행은 운영자가 명령을 돌려 판단하는 별도 행위다.

### Out of Scope — 머신 로컬량 기준의 하드 카텍

- 564MB 같은 머신별 누적량에 하드 상한을 두고 자동 삭제하는 동작은 만들지 않는다. 아카이브는 move 전용이며, 삭제 경로는 본 SPEC 이 열지 않는다.

### Out of Scope — 리포트 내용·포맷 변경

- html-report skill 의 렌더링 동작(모드, tier, 다이어그램 정책)은 전부 현행 유지다. 바뀌는 것은 기본 출력 디렉터리 하나와 디렉터리 자동 생성뿐이다.

### Out of Scope — primary 체크아웃·타 워크트리의 즉시 수거

- 이미 존재하는 27개 카드 워크트리 안의 증거를 본 SPEC 이 일괄 인출하지 않는다. hoist 는 미래의 `moai worktree done` 호출에 적용되는 절차다. 기존 트리의 처리는 각 카드의 통합 창에서 자연 발생하거나 운영자 판단 사항이다.

---

## §F 결정 기록 (iter2 확정)

계획 단계에서 제안이던 4개 결정 포인트는 plan-audit iter1 (score 0.88, MP-7 미해결 마커 2건으로 FAIL) 이후 **레인 채택으로 확정됐다**(카드가 운영자 게이트를 명시하지 않으므로 kickoff-autonomy 정책상 레인 결정; 감사 측정 근거는 `.moai/reports/t1320/plan-audit-iter1.md` 및 본 SPEC §A·plan §F.0).

1. **D1 — tracked 연속성 확정**: `git mv` 이동, `historical/` 재포함 negation 없음 (REQ-RLC-002 현행 유지).
2. **D2 — hoist = 독립 동사 확정**: `moai worktree hoist` + done(L2) 내부 호출 + L1 폐기 플로우 문서 의무 (REQ-RLC-005 재설계). 선정 사유와 기각 기록은 plan §F.0 D2.
3. **D3 — skill 기본 경로 무조건 `.moai/reports/` 확정** (REQ-RLC-003 현행 유지).
4. **D4 — 아카이브 파라미터 확정**: 연령 기반 90일(mtime), `<YYYY-MM>` 분할, 플래그 + `defaults.go`(신설 설정 파일 없음). 세부 3값(창 길이, 충돌 정책 skip-and-report, 1GB 경고 임계)은 열린 마커가 아니라 **보수적으로 채택된 문서화 기본값**이다 — 재측정 근거가 마련되면 후속 카드에서 조정한다.

미해결 요구 확인 마커: 0건 — 네 결정 모두 위 기록으로 확정됐다.
