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

# Acceptance Criteria — SPEC-WORKTREE-GUARD-HEREDOC-DOC-001

모든 기준은 binary PASS/FAIL. 규율: `.claude/rules/moai/development/verification-completeness.md`
(two-cell 채택 §2, RED 4요소 §2.1, 결여형 기준의 정직한 분류 §2 undecidable
disposition). **문서 전용 SPEC이다 — 실행 가능한 RED가 있는 셀만 release-blocking으로
채택하고, 없는 셀은 regression-guard로 정직하게 분류한다.** 기준별 분류를 각 AC에
명시한다.

문서 핀 규칙: 이 SPEC의 문서 수준 핀은 **`6a46c0edb`** (plan 작성 시점 HEAD, 카드
커밋 0개 상태)이며, 자체 핀 없는 모든 AC에 적용된다. 개별 AC의 핀이 있으면 그것이
우선한다.

## AC-WGHD-001 — 미러 바이트 패리티 불변

**분류: release-blocking (기계적 불변). RED의 성격 — 사전 RED가 원리적으로
불가능한 불변형이므로, RED는 뮤턴트 관측으로 시연한다 (아래 절차).**

베이스 트리에서 두 파일은 이미 바이트 동일이다(측정: `cmp` exit 0, `6a46c0edb`).
이 AC의 힘은 "변화 검출"이 아니라 "M1 편집 과정이 불변을 깨는 것의 검출"에 있다.
RED 4요소(사전 관측) 형태를 정직하게 채울 수 없는 이유가 이것이고, 대신 알려진
실패 입력(뮤턴트)에서의 실패 관측을 채택 요구로 삼는다 (verification-completeness
§1.2(b) — 입력이 검사를 붉게 만들고, 그 실패가 관측될 것).

**GREEN 경로**: M1(T2)이 같은 커밋에서 양쪽을 편집한 뒤 —

```
cmp .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
```

exit 0 (무출력). 그리고:

```
go test ./internal/template/ -run TestRuleTemplateMirrorDrift
```

`PASS` (sentinel `RULE_TEMPLATE_MIRROR_DRIFT` 미발생).

**뮤턴트 관측 의무 (run-phase, §E에 기록)**: 미러 한쪽만 임의 1바이트 다르게
편집 → `cmp` 가 다름을 보고하는 것을 관측 → 역편집으로 복구(공유 트리 규율상
`git restore` 금지, 편집-역편집) → `cmp` exit 0 재관측. 실패 관측 없이 이 AC를
PASS로 채택할 수 없다.

**Implements**: REQ-WGHD-004.

## AC-WGHD-002 — 두 구분자 구분의 양쪽 셀 존재

**분류: release-blocking. RED-now: 측정됨 (아래, `6a46c0edb` 핀).**

독트린이 quoted(불확성)와 unquoted(확장) **양쪽**을 이름대고 올바른 의미를 담는
것이 요구다. 한쪽만 실은 뮤턴트가 이 AC를 이기도록 하는 것이 설계 의도다 —
quoted 문장을 떼면 `brace expansion` 카운트가 문턱 미달로 떨어지고, unquoted
문장을 떼면 `unquoted` 앵커가 0이 되어, 어느 방향의 뮤턴트든 기계 검증이
실패한다. 앵커는 필요조건이고, 의미의 올바름은 리뷰 셀(AC-WGHD-004의 리뷰와
별개로, 본 AC의 리뷰 문단)이 충분조건을 담당한다.

**RED-now 셀** (측정: 2026-09-07, 트리 `.claude/worktrees/t512`, HEAD
`6a46c0edb`, 카드 커밋 0개):

```
$ grep -c "brace expansion" .claude/rules/moai/workflow/worktree-integration.md
0
# exit code: 1
```

```
$ grep -c "unquoted" .claude/rules/moai/workflow/worktree-integration.md
0
# exit code: 1
```

RED인 이유: 두 앵커 모두 독트린에 존재하지 않는다 — 즉 "두 구분자 구분이 실려
있다"는 요구가 현재 트리에서 충족되지 않으며, M1의 문서 삽입이 뒤집을 수 있는
자리다(잘못된-이유 RED 아님: 앵커 부재는 M1이 건드리는 바로 그 파일 안에서,
M1이 넣을 바로 그 문장의 키워드로 측정됐다).

**GREEN 경로**: M1(T1) 후 `grep -c "brace expansion" …` ≥ **2** 이고
`grep -c "unquoted" …` ≥ 1이다. **뮤턴트 근거 — 왜 2인가**: plan.md T1 초안에서
`brace expansion`은 quoted 불확성 문장에 2회, unquoted 확장 문장에 1회 등장한다
(합계 3). 문턱이 1이면 quoted 문장을 떼어낸 뮤턴트가 `brace expansion`=1·
`unquoted`=1로 **기계 검증을 통과**한다 — ≥2는 그 방향을 실패시키고(1<2),
반대 방향(unquoted 문장 탈락)은 `unquoted` 앵커 0으로 실패한다. 두 앵커의
조합만이 양방향 단일-문장-탈락 뮤턴트를 모두 잡는다.
리뷰 문단: run/sync 리뷰가 삽입 소절을 읽고 (a) quoted 본문 = 어떤 확장도
없음, (b) unquoted 본문 = 확장 일어남 + 불확성 취급 금지, (c) 살아있는 텍스트의
거절은 합당한 설계라는 경계 — 세 문장이 plan.md T1 초안의 의미와 일치하는지
확인한다.

**Implements**: REQ-WGHD-001.

## AC-WGHD-003 — 생재현 provenance의 독트린 기록

**분류: release-blocking. RED-now: 측정됨 (아래, `6a46c0edb` 핀).**

**RED-now 셀** (동일 트리/HEAD 핀):

```
$ grep -c "re-measured" .claude/rules/moai/workflow/worktree-integration.md
0
# exit code: 1
```

RED인 이유: 재측정 기록 문장이 독트린에 없다. 기존 provenance 행(제보 측
매트릭스)은 존재하지만 이 SPEC이 추가하는 세션 내 양방향 재측정 기록은 없다 —
M1이 정확히 뒤집는 자리다.

**GREEN 경로**: M1 후 `grep -c "re-measured" ...` ≥ 1이고, 리뷰가 그 문장이
(a) 거절 프로브(실행 전 거절 + 대상 파일 부재 확인)와 (b) 통과한 대조군
프로브(리터럴 보존)를 모두 운반하는지 확인한다. 재현 원장 경로 인용은
독트린이 아니라 SPEC artifacts(research.md, acceptance.md, 커밋 메시지)의
소관이다(REQ-WGHD-005 — 원장 경로는 중립 텍스트가 아니다).

**Implements**: REQ-WGHD-002.

## AC-WGHD-004 — loosening-danger 경계 (약화 금지 + 소속 유지 + fixed 주장 금지)

**분류: regression-guard. release-blocking 아님 — 결여형(absence) 기준은
기능 부재 시 저절로 만족되는 공허한 초록의 위험이 있고, §2의 채택 규율상
결여형만으로 release-blocking 채택이 불가하다. 무게중심은 리뷰 셀이 진다.**

**리뷰 셀 (1차)**: run/sync 리뷰가 삽입 소절 전문을 plan.md T1 초안과 대조해
(a) 가드 약화·비활성·억제·재구성 안내가 없는지, (b) 소속 판정(바이너리 소유)이
유지되는지, (c) 문서 수정이 거절의 수정으로 읽히지 않는지 — 를 판정한다.

**보조 셀 (기계, 참고용)**: 삽입된 소절 범위에서
`disable|weaken|suppress|reconfigure` 스캔 — 기대 0. 단, 이 스캔의 0은
"안내 부재"의 간접 신호일 뿐이며 단독으로는 아무것도 증명하지 않는다(부재
가드의 한계를 명시적으로 기록한다).

**Implements**: REQ-WGHD-003.

## AC-WGHD-005 — 독트린 텍스트의 템플릿 중립성

**분류: release-blocking (결여형 + 대조 쌍).**

결여형의 공허한 초록 위험(문자열이 아직 없어서 지금 초록)을 대조 쌍으로
깬다: **회신 초안은 `1659`를 반드시 실는다**(AC-WGHD-006) — 즉 "이슈 번호가
리포에는 있고 독트린에는 없다"는 대비가 성립할 때 이 AC는 실제 판별력을 가진다.
뮤턴트: 이슈 번호를 인용하는 독트린 소절 초안은 이 AC를 반드시 실패시켜야 한다.

**RED-now/베이스라인 셀** (측정: `6a46c0edb` 핀 — 인자 순서대로 live 먼저,
mirror 나중; 원장의 프로브 순서와 같은 방향):

```
$ grep -c "1659" .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
.claude/rules/moai/workflow/worktree-integration.md:0
internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md:0
# exit code: 1
```

```
$ grep -c "SPEC-WORKTREE-GUARD-HEREDOC" .claude/rules/moai/workflow/worktree-integration.md
0
# exit code: 1
```

**GREEN 경로**: M1 후에도 위 스캔이 동일하게 0을 유지한다(편집이 중립성을 깨지
않음). 대조 셀: `.moai/reports/t512/reporter-reply-1659.md` 에서
`grep -c "1659"` ≥ 1 (AC-WGHD-006과 짝).

**Implements**: REQ-WGHD-005.

## AC-WGHD-006 — 제보자 회신 초안 (한국어)

**분류: release-blocking (카드 범위). RED-now: 측정됨.**

**RED-now 셀** (`6a46c0edb` 핀):

```
$ ls .moai/reports/t512/reporter-reply-1659.md .moai/reports/t512/upstream-draft-claude-code.md
# stdout: (empty)
# stderr:
# ls: .moai/reports/t512/reporter-reply-1659.md: No such file or directory
# ls: .moai/reports/t512/upstream-draft-claude-code.md: No such file or directory
# exit code: 1
```

스트림 귀속은 별도 관측으로 확인했다(같은 명령에 `2>/dev/null` → stdout 완전
공백 + rc=1). 공백 stdout + 비영exit은 doctrine §2.1 이 완전한 관측으로
인정하는 형태다.

**GREEN 경로**: M2(T3)가 파일을 만들고, 리뷰가 네 요소 앵커를 확인한다:
(a) 재현 확인 + 비대칭 매트릭스 인용(프로브 1 거절 verbatim + 프로브 2 통과,
`6a46c0edb` 병기), (b) 소속 = Claude Code 바이너리 + 리포 자체 독트린 인용
(`worktree-integration.md:481`) + `branch_guard.go` 아님(`:485`), (c) 우회 2종
(Write 도구 / 단순 별개 명령 분할), (d) 업스트림 채널 안내 + 초안 전용 명시,
(e) 어조 — blame-shifting으로 읽히지 않는다(리드 추가분 [HARD]): 검증으로
시작, 우회는 도움으로 제시, 업스트림 전달은 소유권 거부가 아닌 제보자 발견의
대변으로 프레임. 대조 셀: 파일 안 `grep -c "1659"` ≥ 1.

**Implements**: REQ-WGHD-006.

## AC-WGHD-007 — 업스트림 제보 초안 (영어)

**분류: release-blocking (카드 범위). RED-now: AC-WGHD-006의 ls 셀과 동일
관측(두 파일 모두 부재, `6a46c0edb` 핀).**

**GREEN 경로**: M3(T4)가 파일을 만들고, 리뷰가 다섯 요소 앵커를 확인한다:
(a) 비대칭 매트릭스 — 두 프로브 verbatim, (b) quoted 본문 불확성 논증,
(c) 수정 방향 1 (quoted 본문 중괄호 접기 — 명령 치환과 같은 접기), (d) 수정
방향 2 (거절 메시지가 문제 구문을 이름대기), (e) 제보자(jjjh7401, #1659) credit.
범위 제한 준수: 스크립트 파일 우회 관측과 unquoted 동작 변경이 ask로 들어가지
않았는지 리뷰한다.

**Implements**: REQ-WGHD-007.

## 수용 매트릭스 요약

| AC | 분류 | RED-now | GREEN 소관 | 검증 수단 |
|----|------|---------|-----------|----------|
| AC-WGHD-001 | release-blocking (불변) | 뮤턴트 관측으로 시연 | M1/T2 | `cmp` + mirror test |
| AC-WGHD-002 | release-blocking | 측정됨 (0/rc=1 ×2, `6a46c0edb`) | M1/T1 + 리뷰 | grep 앵커 ×2 + 리뷰 |
| AC-WGHD-003 | release-blocking | 측정됨 (0/rc=1, `6a46c0edb`) | M1/T1 | grep 앵커 + 리뷰 |
| AC-WGHD-004 | regression-guard | 결여형 — 채택 불가, 리뷰 1차 | M1 리뷰 | 리뷰 + 보조 스캔 |
| AC-WGHD-005 | release-blocking (대조 쌍) | 측정됨 (0/0/rc=1, `6a46c0edb`) | M1 유지 + M2 대조 | grep + 대조 셀 |
| AC-WGHD-006 | release-blocking (카드) | 측정됨 (ls 부재, `6a46c0edb`) | M2/T3 | `test -f` + 4요소 리뷰 |
| AC-WGHD-007 | release-blocking (카드) | 측정됨 (동일 ls 셀) | M3/T4 | `test -f` + 5요소 리뷰 |
