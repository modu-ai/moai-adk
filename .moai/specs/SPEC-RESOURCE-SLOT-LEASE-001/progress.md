# SPEC-RESOURCE-SLOT-LEASE-001 — progress (card t607)

plan 단계 산출물을 트리 `c4ce42eca` @ `WT-heavy-test-slot`(worktree t607)에서 작성했다. Tier M. Status: draft.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- 산출물(Tier M): spec.md, plan.md, acceptance.md, progress.md(이 파일).
- SPEC ID 정규식 검사 — 실행한 명령과 출력:
  `ID="SPEC-RESOURCE-SLOT-LEASE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`
- ID 고유성: 이 워크트리 `.moai/specs/`에서 `RESOURCE`·`SLOT`·`LEASE` 이름을 가진 디렉터리는 `SPEC-SYNC-SHA-SLOT-FORMAT-001` 하나뿐이며 ID가 다르다.
- 프론트매터: 12개 정식 필드 + `tier: M`. `status: draft`. `phase`는 릴리스 대상(`"v3.2.0 target"`)이다.
- 요구사항: REQ-RSL-001..016(GEARS, IF/THEN 없음, Tier M 상한 16 이내).
- 인수 기준: AC-RSL-001..016(Tier M 상한 16 이내). 릴리스 차단 기준의 RED-now 셀은 acceptance.md 증거 원장(0.2.0 기준 EL-1..EL-8)을 인용한다.
- Out of Scope: `### Out of Scope — <주제>` H3 7개, 각각 `-` 항목 보유.
- spec lint — 이 트리(`c4ce42eca`)에서 빌드한 바이너리를 경로로 호출했다(설치본 `ed71054d3`은 HEAD의 조상이라 뒤처진 빌드다: `git merge-base --is-ancestor ed71054d3 HEAD` → 종료 코드 0):
  - `go build -o <scratch>/moai ./cmd/moai` → 성공
  - `<scratch>/moai spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001` → `✓ No findings — all SPEC documents are valid`
  - 음성 대조군: 같은 산출물의 사본에 `phase: plan`을 심고 같은 바이너리로 lint → `ERROR FrontmatterPhaseInvalid ... 1 error(s)`. lint가 이 디렉터리 모양을 실제로 읽는다는 증거다.
  - 참고: 요구 문장을 한국어로만 썼던 첫 초안은 설치본 lint에서 `ModalityUnjudged` 경고 16건을 받았다. 정본 요구 문장을 GEARS 영어로 바꾸고 한국어 설명을 하위 항목으로 옮긴 뒤 위 결과가 나왔다.
- M6(레인 문서 반영) 게이트 상태: `git merge-base --is-ancestor WT-acquire-branch-record develop` → 종료 코드 1(`WT-acquire-branch-record` = `f680dab46`, `develop` = `eb50af5a8`). 게이트 닫힘.

### plan-audit 1회차 수리 (v0.2.0)

plan-auditor 1회차는 FAIL 0.78이었다(Tier M 기준 0.80, 보고서 `.moai/reports/t607/plan-audit-iter1.md`, 커밋 `e50cfea93`). blocking D1-D6과 optional D7-D12를 모두 반영했다. 아래 측정은 이번 실행에서 트리 `e50cfea93`에 대해 했다.

- 코드 트리 불변 확인: `git diff --stat c4ce42eca HEAD -- internal cmd pkg` → 출력 없음. 문서 핀 `c4ce42eca`의 코드 측정이 그대로 유효하다.
- D1: 훅 루트 해석은 `internal/hook/path_resolve.go`:78-94에서 `CLAUDE_PROJECT_DIR` → `os.Getwd()`만 쓴다(직접 읽음). §B.3 정정, REQ-RSL-008(가드 루트 정규화), AC-RSL-016 추가.
- D2: 출처 `internal/kanban/integration_lock_cross_test.go`:46-63(500ms 풀림 타임아웃)과 :180-183(부모 pid 고정)을 직접 읽고 AC-RSL-001a를 다시 썼다.
- D4: lane 규칙 `.claude/rules/local/gitflow-lane-protocol.md` §8을 읽고 그 형태(`CARD_BASE=$(git merge-base develop HEAD)` + 대조군 + 병합 전 전용)로 바꿨다.
- D6: 언어 도구 토큰 목록의 기준선과 대조군을 쟀다 — `/usr/bin/grep -nwiE '<TOOL_TOKENS>' internal/template/templates/.moai/config/sections/workflow.yaml` → 출력 없음, 종료 코드 1(EL-7). 같은 목록으로 `/usr/bin/grep -cwiE '<TOOL_TOKENS>' internal/template/templates/.claude/rules/moai/languages/python.md` → `7`, 종료 코드 0(EL-8).
- D7: OQ-1은 감사 보고서 Evidence E3의 측정으로 닫았다. 런타임 버전은 E3에 없어 이번 실행의 `claude --version` → `2.1.268 (Claude Code)`을 참고로만 적었다.
- D9: M6 게이트 재측정 — `git merge-base --is-ancestor WT-acquire-branch-record develop` → 종료 코드 1. 이 시점 `git rev-parse --short develop` → `ac6c42c2d`(움직이는 ref의 측정 시점 값).
- spec lint(0.2.0): 트리 HEAD `e50cfea93`에서 `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=e50cfea93" -o <scratch>/moai-e50 ./cmd/moai`로 빌드하고 경로로 호출했다. 바이너리의 `version` 출력에 커밋 `e50cfea93`가 찍힌다. `<scratch>/moai-e50 spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001` → `✓ No findings — all SPEC documents are valid`. 음성 대조군: 0.2.0 산출물 사본에 `phase: plan`을 심고 같은 바이너리로 lint → `1 error(s), 0 warning(s)`.
- 예산: REQ 16 / AC 16 유지. 옛 REQ-RSL-008(pid 0)은 REQ-RSL-005로, 옛 AC-RSL-002(뮤턴트 관측)는 AC-RSL-001b로, 옛 AC-RSL-016(CLI·교차 플랫폼)은 AC-RSL-003c로 흡수했다.

## §E.2 Run-phase Evidence

_<run 단계 대기>_

## §E.3 Run-phase Audit-Ready Signal

_<run 단계 대기>_

## §E.4 Sync-phase Audit-Ready Signal

_<sync 단계 대기>_
