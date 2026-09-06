---
id: SPEC-LEAD-DEPUTY-001
title: "plan — 리드 세션 직렬 병목 해소 (상주 deputy 채택)"
version: "0.1.0"
created: 2026-09-03
updated: 2026-09-06
author: manager-spec (card t471)
tier: M
---

# PLAN: SPEC-LEAD-DEPUTY-001

## §A Context

- **Card**: t471 (리드 발행 2026-09-03, 운영자 지시) — Class C
- **Worktree**: `.claude/worktrees/t471` · branch `WT-lead-bottleneck` · HEAD `7835148d3e02e6b244371ba3d6c0c4b5b20886db` (= origin/develop, plan-phase 시점 실측)
- **Tier**: M (확정 — §F 마일스톤 판정 근거). Artifact set: spec.md + plan.md + acceptance.md (+ progress.md, 미집계)
- **SPEC artifacts**: `.moai/specs/SPEC-LEAD-DEPUTY-001/{spec,plan,acceptance,progress}.md`
- **plan-auditor**: 미실행 — Tier M PASS threshold 0.80. verdict 파일 예정 위치: `.moai/reports/t471/verdict.md`
- **depends_on**: `SPEC-LEAD-DEBOTTLENECK-001` — status: completed → fulfilled (run-gate pre-flight PASS 예상)
- **변경 대상 표면** (Template-First — 편집 원본은 전부 `internal/template/templates/` 쪽):
  - `internal/template/templates/.claude/agents/moai/manager-lead.md` → 로컬 미러 `.claude/agents/moai/manager-lead.md`
  - `internal/template/templates/.codex/agents/moai/manager-lead.toml` → `make agents-emit` 기계 방출 (수편집 금지)
  - `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` (+ `-detail.md` 필요 시) → 로컬 미러 동일 경로
- **기존 기구 (EXTEND, 재설계 금지)**: manager-lead § Deputy dispatch surface, `SendMessage`/`ListAgents` tools, delivery-shape 검증 문단, `cross-session-messaging.md` idle-통지 경계 절

## §B Known Issues (본 SPEC 도메인 관련만)

- **B4 Frontmatter**: canonical 12필드, `created:`/`updated:` (snake_case 금지) — 준수 완료
- **B6 spec-lint heading**: `### Out of Scope — <topic>` H3 + `-` 불릿 — spec.md §6 준수
- **B2 Cross-SPEC policy**: `SPEC-LEAD-DEBOTTLENECK-001` (completed)과 관계는 **채택(adopt)** — supersede 아님. `related_specs`로 상호 참조. t269 `SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001`의 부활 금지 교리를 deputy가 상속 (t283 상속 경로 유지)
- **Template neutrality** (`internal/template/templates/**`): `[리드 자체 계수]` 수치, 사고 일자, 카드 id(t471 등)는 템플릿에 금지 — SPEC 산출물(로컬 전용)에만 존재. 교리 문구는 "the lead"/"the deputy" 일반 표현 (REQ-LDP-010, AC-LDP-007)
- **`make agents-emit` (C2→C3)**: `.claude/agents/moai/manager-lead.md` 템플릿 사본(C2) 편집 후 반드시 재생성. `.codex/agents/moai/*.toml`(C3) 수편집 금지. 세 지점 검사: `make build`(선행 agents-emit-check, 읽기전용) / `go test ./internal/template/agentemit/...` / `make embed-check`
- **B8/B10 Tree hygiene**: `git add` 명시 pathspec만. 타 SPEC 디렉터리·`.moai/state/`·`.moai/harness/` 비접촉

## §C Pre-flight

```bash
git branch --show-current && git rev-parse HEAD   # WT-lead-bottleneck 기대
diff -q .claude/agents/moai/manager-lead.md internal/template/templates/.claude/agents/moai/manager-lead.md  # 편집 전 쌍 동기 확인 (의도된 분기 있을 수 있음 — diff 내용을 읽고 판정)
go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent' -count=1  # depth-seal baseline (RED-now 아님 — PRESERVE 기준선)
go test ./internal/template/agentemit/... -count=1
grep -rn "t471\|리드 자체 계수" internal/template/templates/ | wc -l   # neutrality 사전 기준선: 0 기대
```

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE**: §5 제약 목록 전부 — 특히 manager-lead 위임 5종/보유 6종 표·delivery-shape 문단 원문 보존, `DEPUTY-RETAINED-BY-LEAD` 6항목 변경 금지
- `internal/`·`pkg/`·`cmd/` 전체 비접촉 (REQ-LDP-011)
- `--no-verify`, `--amend`, force-push 금지; Conventional Commits + `🗿 MoAI` 트레일러
- 로컬 `.claude/rules/moai/**` 직접 편집 후 템플릿 미러 누락 금지 — Template-First (`moai update`가 로컬 편집 소멸)
- 템플릿에 내부 상태(SPEC-ID·자체 계수·날짜·카드 id) 반입 금지
- 동시 write-capable 에이전트 금지 — 본 SPEC은 단일 manager-develop 직렬 실행

## §E Self-Verification (manager-develop 반환 의무)

- E1: AC-LDP-001~010 PASS/FAIL 매트릭스 (명령+원문 출력+트리 SHA 귀속, VCI §3 5-섹션 형식)
- E2: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (마크다운 전용이지만 빌드 회귀 감시)
- E3: 해당 없음 (Go 소스 변경 0) — "변경 없음"을 `git diff --stat internal/ pkg/ cmd/` 빈 출력으로 입증
- E4: neutrality grep — `grep -rn "t471\|리드 자체 계수\|LEAD-DEPUTY" internal/template/templates/ | wc -l` → 0
- E5: `make agents-emit` 후 `git status --short internal/template/templates/.codex/` — toml 재생성 확인; `go test ./internal/template/agentemit/...`
- E6: depth-seal 테스트 PASS 출력 원문; 커밋 SHA 목록
- E7: blocker 발생 시 구조화 보고 (AskUserQuestion 금지)

## §F Milestones (결정-가역성 순 — 바뀔 가능성 큰 결정이 앞에)

### M1 — 축 A: 상주 deputy 채택 의무 (독립 가치, 단독 출하 가능)
- `kanban-dispatch.md` § Deputy dispatch surface에 **상주 모드** 절 추가: 배치 시작 시 deputy 1개 UNNAMED spawn 의무(REQ-LDP-001), 완료 보고 → deputy raw 판독 → `RECOMMEND:` 요약만 리드 턴으로(REQ-LDP-002), idle 통지 요청 방식(REQ-LDP-007, M3 선반영 가능). 기존 [HARD] 절 원문 보존의 확장
- `manager-lead.md` § Deputy dispatch surface에 상주 모드 + 완료-보고 경로 + delivery-shape 의무 상속 명시. 위임/보유 매트릭스는 변경 없음
- 판정 소재·운영자 게이트 절은 만지지 않는다 (§1.5 못박기 — 병목의 원인이자 보호 가치임을 명시하는 한 문단 추가 가능)

### M2 — 축 B: 회차 보고 위임 + 파일 분할
- 회차 보고 규율: 측정 배치·표 초안은 deputy 소관, 리드는 직접 단언 수치만 재저작, deputy 측정치는 `deputy 측정 (경로)` 귀속 표기 (REQ-LDP-005) — `kanban-dispatch.md` § Report milestones ↔ queue cards 인접부 또는 detail companion에 규율 절 추가
- 회차 보고 파일 구조: 회차별 파일 + 인덱스. 단일 대형 파일 전체 재작성 금지 (REQ-LDP-006). 이 규율은 교리(who writes what) 차원 — 실제 보고 파일은 리드 세션 산출물이므로 SPEC이 규율만 정의
- `[NEEDS CLARIFICATION 없음]` — 분할 단위(회차당 1파일)와 인덱스 형식은 본 계획이 확정

### M3 — 축 C: idle 통지 전환
- 폴링 → `notify_when_idle` 1회 요청 대체 규율 (REQ-LDP-007) — M1의 상주 절에 포함 가능하나 리뷰 단위로 분리
- idle-통지 경계: `cross-session-messaging.md` § An idle notice is a scheduling hint **인용 상속** (REQ-LDP-008) — 경계 절을 새로 만들지 않고 상호참조로 연결. 통지만으로 카드 전진 금지 명시 (VCI §1.1 surface 1 연결)

### M4 — 기계적 마무리 (마지막 — 결정 소요 없음)
- `make agents-emit` → `make build` → depth-seal 테스트 → neutrality grep(0) → 로컬↔템플릿 쌍 동기 확인 → spec-lint (`moai spec lint` 대상 SPEC 자가 점검)
- M1~M3 각 편집마다 Template-First가 적용되므로 M4는 통합 검증만 수행

**순서 근거**: M1은 단독 가치·단독 출하 가능(가장 바뀔 일 없는 핵심 결정) — M2/M3가 지체돼도 채택 자체는 착지. M2는 M1의 deputy가 상주해야 성립하는 종속 결정. M3는 M1이 정의한 감시 임무의 방식 변경. M4는 순수 기계적 검증.

## §G Anti-Patterns

- **deputy에게 판정 위임**: `RECOMMEND:`는 권고다. `FINAL VERDICT:` 토큰이 deputy 출력에 나타나면 구조 위반
- **named spawn**: deputy를 이름 붙여 spawn → in-process teammate 전환 → 결과 반환 두절 (GLM 위험)
- **통지만으로 카드 전진**: idle 통지는 "언제 읽을지"만 알려준다 — 증거를 읽기 전 전진은 미관측 완료 주장
- **턴 수를 지표로 보고**: 운영자 지시 빈도 교락 — 툴 배치 수만이 지표 (REQ-LDP-009)
- **로컬 먼저 편집**: `.claude/rules/moai/**` 직접 편집은 다음 `moai update`에 소멸 — 반드시 템플릿 원본 먼저
- **C3 수편집**: `.codex/agents/moai/*.toml`은 방출물 — C2 고치고 `make agents-emit`

## §H Cross-References

- spec.md §7 교차참조 목록 전체
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier M 5-섹션 위임 템플릿 (Section B는 본 §B 사용)
- acceptance.md §D.0 — RED-now 기준값과 `[리드 자체 계수]` 귀속
- `.moai/reports/t471/` — 카드 증거 경로 (plan-evidence.md, plan-auditor verdict.md 예정)
