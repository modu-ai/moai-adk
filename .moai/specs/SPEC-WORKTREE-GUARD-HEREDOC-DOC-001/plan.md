---
id: SPEC-WORKTREE-GUARD-HEREDOC-DOC-001
title: "Worktree-guard heredoc brace asymmetry — ownership pivot, doctrine extension, reporter reply, upstream draft"
version: "0.1.0"
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.2.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: M
issue_number: 1659
tags: "worktree, guard, heredoc, delimiter, documentation, claude-code-binary, upstream-draft"
---

# Plan — SPEC-WORKTREE-GUARD-HEREDOC-DOC-001

## 1. Plan Summary

### 1.1 Strategy

문서 전용 단일 웨이브. 3개 마일스톤이 순서 의존 없이 독립적이므로(run 세션의
판단에 따라 직렬 집행이 기본) 한 번의 run 페이즈로 마친다. 유일한 트리 변경은
M1(독트린 + 미러, 같은 커밋)이고 M2/M3는 카드 증거 경로의 초안 파일이다. 코드,
테스트, 빌드 산출물 없음 — `make build` 불필요(마크다운만 변경).

### 1.2 Base branch / 통합 경로

- 카드 워크트리: `.claude/worktrees/t512`, 브랜치 `WT-guard-heredoc`
  (develop에서 분기, plan 작성 시점 HEAD `6a46c0edb` = 로컬 develop tip).
- 통합: `git-flow` 카드 워크플로 (`.claude/rules/local/gitflow-lane-protocol.md`).
  레인은 push 하지 않는다 — run/sync 마친 뒤 리드의 develop 병합 창 경유로
  `--no-ff` 병합되고, `origin/develop` CI가 판정면이다.
- 모든 커밋 메시지에 카드 id `t512`를 남긴다(추적성 운반체). Conventional
  Commits + `🗿 MoAI` 트레일러.

### 1.3 Working location

- 이 워크트리 안에서 전 페이즈 집행. primary 체크아웃으로 나가지 않는다.
- 편집 대상은 절대경로 없이 워크트리 루트 기준 상대경로로 기술한다.

## 2. Wave Decomposition

단일 웨이브. 근거: 4개 편집이 모두 마크다운이고, M1의 두 파일은 같은 커밋으로
묶이는 결합 편집이며, M2/M3는 독립 신규 파일. 순증 LOC < 100.

## 3. Tasks

### M1 — 독트린 보강 + 미러 (T1, T2)

#### T1 — `worktree-integration.md` 소절 추가

**Deliverable**: `.claude/rules/moai/workflow/worktree-integration.md` 의
"### What has been observed to trip the worktree guard" 절의 **Gap 문단 뒤**,
"### Workarounds" 절 **앞**에 아래 소절을 삽입한다. 영어, 중립 문체(이슈 번호·
SPEC ID·날짜·리포 경로 금지 — REQ-WGHD-005). 기존 표·Gap 문단·Workarounds는
한 글자도 바꾸지 않는다.

삽입 초안 (run phase에서 이 문안을 기준으로 집행 — 문장 단위 재배열은 허용,
두 구분자 의미와 loosening-danger 경계는 불변):

```markdown
### Why the two heredoc delimiter forms differ

The distinction the observation above turns on is the delimiter quoting, and it
is bash semantics, not guard behavior:

- **A quoted delimiter (`<<'EOF'`) makes the body inert text.** Bash performs no
  expansion of any kind inside such a body — no parameter expansion, no command
  substitution, no arithmetic expansion, and no brace expansion. A brace there
  is a literal character and cannot be brace expansion. This is the same fact a
  guard relies on when it folds a command substitution appearing in such a
  body, as observed above.
- **An unquoted delimiter (`<<EOF`) makes the body live text.** Parameter
  expansion, command substitution, arithmetic expansion, and brace expansion
  all apply. No guard — and no reader of this file — may treat an
  unquoted-delimiter body as inert, and a refusal that errs on the side of
  caution there is correct behavior, not a defect.

The observed asymmetry is therefore narrow: in a position provably free of
expansion (the quoted-delimiter body), braces alone are treated as live, while
command substitutions in the same position are already folded.

The refusal half of this asymmetry was re-measured by paired probes in a
worktree-isolated session on this repository: the brace form was refused with
the `too complex to verify` sentence above before the command executed (the
target file was confirmed absent afterward), and the paired probe — the same
shape with a command substitution in the body — executed and passed with the
body preserved as a literal. This remains a record of observations, not a
specification; the boundary of the guard's analyzer is still unmeasured.
```

**Verification** (워크트리 루트에서):

```
grep -c "brace expansion" .claude/rules/moai/workflow/worktree-integration.md   # >= 2 (quoted 문장 2회 + unquoted 문장 1회 = 3 기준; 1이면 quoted 탈락 뮤턴트 통과)
grep -c "unquoted" .claude/rules/moai/workflow/worktree-integration.md          # >= 1
grep -c "re-measured" .claude/rules/moai/workflow/worktree-integration.md       # >= 1
grep -cE "1659|SPEC-WORKTREE-GUARD-HEREDOC" .claude/rules/moai/workflow/worktree-integration.md  # 0
```

**Implements**: REQ-WGHD-001, REQ-WGHD-002, REQ-WGHD-003, REQ-WGHD-005.

#### T2 — 미러 동기화 + 같은 커밋 패리티

**Deliverable**: T1과 **같은 커밋**에서
`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`
에 동일 내용을 적용한다. 커밋 전 `cmp` 로 바이트 동일을 확인하고, 뮤턴트 관측을
수행한다(아래 검증 절차).

**Verification**:

```
cmp .claude/rules/moai/workflow/worktree-integration.md \
    internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md   # exit 0
go test ./internal/template/ -run TestRuleTemplateMirrorDrift                          # PASS
```

뮤턴트 관측(RED 시연 — AC-WGHD-001의 알려진 실패 입력): 미러만 한 줄 다르게
건드린 뒤 `cmp` 가 실패하는 것을 관측하고, `git restore` 없이 **편집을 되돌리는
재편집**으로 복구한 뒤 `cmp` 재측정으로 exit 0을 다시 관측한다. (공유 트리
미커밋 작업 보호 규율상 `git restore -- <path>` 는 쓰지 않고, 편집-역편질다.
뮤턴트 상태에서 다른 커밋 작업을 하지 않는다.)

**Implements**: REQ-WGHD-004.

#### M1 커밋 규약

- 한 커밋: 독트린 + 미러 + (이 SPEC의 plan artifacts가 아직 미추적이면 함께)
  재현 원장 `.moai/reports/t512/`.
- 메시지 예: `docs(t512): worktree-guard heredoc two-delimiter distinction —
  doctrine + mirror (SPEC-WORKTREE-GUARD-HEREDOC-DOC-001)`
- 스테이징은 명시 경로목록으로(status 재판독과 같은 호출에서).

### M2 — 제보자 회신 초안 (T3)

**Deliverable**: `.moai/reports/t512/reporter-reply-1659.md`, 한국어, 초안
전용(게시 금지). 네 요소를 모두 담는다 (REQ-WGHD-006):

1. 분석이 옳았고 생으로 재현됐다는 확인 + 비대칭 매트릭스 인용(원장의 프로브
   1/2 — 거절 verbatim + 대조군 통과, `6a46c0edb` 병기).
2. 가드 소속 = Claude Code 바이너리 — 근거로 **이 리포 자체의 독트린**
   (`worktree-integration.md:481` — "No — there is no source here that
   implements or configures it")을 인용해 "리포 내 수정 불가"를 설명.
   `branch_guard.go` 가 아니라는 점(`:485`)도 한 줄로.
3. 우회 안내: 파일 내용은 Write 도구(가드가 본문을 파싱하지 않음), 명령은
   단순한 별개 명령으로 분할(거절 메시지 자체의 안내). 가드 약화 조언 없음.
4. 업스트림 제보가 올바른 채널이며 초안이 준비 중/완료됐음을 안내. 게시는
   리드가 카드 원격 착지 후 — 초안임을 명시.

**어조 제약 (리드 추가분 [HARD])**: 회신은 blame-shifting으로 읽혀서 안 된다 —
제보자 분석의 **검증**으로 시작하고, 우회는 **도움**으로 제시하며, 업스트림
회신(전달)은 소유권 거부가 아니라 **제보자 발견의 대변**으로 프레임한다.

**Verification**: `test -f` 존재 + 4요소 앵커 리뷰 + 어조 리뷰
(acceptance.md AC-WGHD-006).

**Implements**: REQ-WGHD-006.

### M3 — 업스트림 제보 초안 (T4)

**Deliverable**: `.moai/reports/t512/upstream-draft-claude-code.md`, 영어, 초안
전용(게시 금지). 다섯 요소를 모두 담는다 (REQ-WGHD-007):

1. 비대칭 매트릭스 — 두 프로브 verbatim (거절 문장 + 통과한 대조군, 원장 인용).
2. quoted-delimiter 본문이 증명 가능하게 불확성이라는 논증(bash 확장 규격 —
   제보자 원 논거를 credit 하며 재사용).
3. 수정 방향 1 — quoted-delimiter 본문의 중괄호를 접는다(명령 치환에 이미
   적용한 것과 같은 접기).
4. 수정 방향 2 — 거절 메시지가 문제 구문을 이름대게 해 사용자가 재구성하게
   한다.
5. 제보자(jjjh7401)의 원 관측 credit (#1659).

범위 제한: 스크립트 파일 우회 관측은 ask에 넣지 않는다(제보자가 별개 처리 요청;
Non-Goals). unquoted 본문 동작 변경 요구도 넣지 않는다(합당한 설계).

**Verification**: `test -f` 존재 + 5요소 앵커 리뷰(acceptance.md AC-WGHD-007).

**Implements**: REQ-WGHD-007.

## 4. Verification Plan (run-phase §E 매핑)

| 항목 | 명령 | 합격 판정 | §E |
|------|------|----------|-----|
| 미러 패리티 | `cmp` (T2) | exit 0 | E1 |
| 미러 drift 가드 | `go test ./internal/template/ -run TestRuleTemplateMirrorDrift` | PASS | E1 |
| 두 구분자 앵커 | T1의 grep 4종 | 판정값 준수 | E1 |
| 뮤턴트 관측 | T2 뮤턴트 절차 | cmp 실패→복구→exit 0 관측 기록 | E8 성격(RED 증거) |
| 초안 2종 | `test -f` + 앵커 리뷰 | 존재 + 요소 충족 | E1 |
| 범위 위반 스캔 | `git status --short` (편집 파일 목록이 5§ 대상만) | 이탈 0 | E5 대용 |

- `go build`/`go vet`/커버리지: 마크다운 전용 변경이라 대상 없음 — E2/E3는
  "대상 없음(문서 전용)"으로 보고하는 것이 정직한 표기다.
- 전체 스위트 `go test ./...` 를 로컬에서 돌리지 않는다(레인 규율). 영향 범위는
  `internal/template` 단일 패키지의 미러 테스트뿐이다.

## 5. Constraints (DO NOT VIOLATE)

- **금지**: `internal/hook/**`, `branch_guard.go`, 훅 래퍼 `.sh`/`.sh.tmpl`,
  t511 SPEC 디렉터리, `.moai/reports/t512/repro-heredoc-brace.md` 의 편집.
- **금지**: 회신·업스트림 초안의 게시(원격 행위). 초안 파일 작성까지만.
- **금지**: 독트린 텍스트에 이슈 번호·SPEC ID·내부 날짜·리포 경로(REQ-WGHD-005).
- **금지**: 어떤 가드의 약화·비활성·재구성 안내(REQ-WGHD-003).
- **금지**: `git add -A`/`git add .` — 명시 경로목록 스테이징 + 직전 status
  재판독.
- **금지**: 레인의 push (gitflow 레인 프로토콜 — 리드 일괄).
- 이 워크트리는 카드의 유일본이다. run 중 워크트리를 폐기하지 않는다.
