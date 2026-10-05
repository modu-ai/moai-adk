---
id: SPEC-WORKTREE-GUARD-HEREDOC-DOC-001
title: "Research — worktree-guard heredoc brace asymmetry"
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

# Research — SPEC-WORKTREE-GUARD-HEREDOC-DOC-001

> 측정 트리: `.claude/worktrees/t512` · HEAD `6a46c0edb` · 브랜치
> `WT-guard-heredoc` · 카드 커밋 0개 (plan 작성 시점). 카드 t512 = GH #1659.
> 디스패치가 확립 사실로 내려준 것은 재연구하지 않고 인용하되, 본 세션에서
> 직접 재측정한 항목은 그 사실을 별도 표기한다.

## 1. 생재현 — 비대칭의 양방향 측정 (커밋 예정 증거)

원장: `.moai/reports/t512/repro-heredoc-brace.md` (2964바이트, 워크트리에
존재 — **단, 미추적 상태**. §6 관측 차이 1 참조).

- **프로브 1** (제보 형태: `<<'T512EOF'` 본문에 `{"guard": ...}` JSON 라인):
  실행되지 않고 거절. 거절 문면 verbatim — "This session is isolated in the
  worktree …, but this command is too complex to verify that it stays inside
  the worktree. Refusing to run it — … Split it into plain, separate commands
  and run them from …". 미실행 증명: 대상 파일 부재 (`ls` → rc=1).
- **프로브 2** (대조군: 같은 형태, 본문에 `$(echo …)` 명령 치환): 실행됨,
  rc=0, 55바이트 파일 생성, 본문 리터럴 보존(치환 미발생 — quoted 규격대로).
  즉 가드는 명령 치환을 접어 통과시켰다.
- **판정**: 같은 heredoc 구조에서 본문 명령 치환은 통과, 본문 중괄호는 거절 —
  접기 비대칭이 생으로 재현됐다(제보자 주장과 일치).
- **스크래치 정리 확인 (본 세션 재측정)**: `t512-repro-subst.txt`,
  `repro-subst.txt` 모두 부재 (`ls` → rc=1) — 원장의 정리 계획이 실행된 상태.
  남은 스크래치 없음.

## 2. 소속 판별 — 결함의 몸은 바이너리에 있다

본 세션 재측정 (HEAD `6a46c0edb`):

- `grep -rn 'too complex to verify' internal/ .claude/hooks/` → 소스 0건.
  적중 3건은 모두 문서 인용: `.claude/rules/moai/workflow/worktree-integration.md`
  의 `:481`, `:485`, `:491` (라인별 verbatim 확인).
- `:481` — 소속 표: `too complex to verify` 거절의 소속 = **The Claude Code
  binary**, "No — there is no source here that implements or configures it".
- `:485` — `internal/hook/branch_guard.go` 는 1차 체크아웃의 브랜치 상태
  방어 가드로, 이 워크트리 가드를 구현·설정하지 않으며, 편집해도 이 거절에
  영향 없음.
- `:491` — 거절 전문 인용 블록.
- 이는 #1658 스레드에서 유지자가 제보자에게 말한 소속 설명과 일치한다(디스패치
  확립 사실 — 재연구 없이 인용).

## 3. 기존 독트린 상태 — M1 델타의 정밀화

`worktree-integration.md` § "Refused Commands in a Worktree-Isolated Session"
(라인 475 전후부터 Workarounds 절까지)의 현재 상태:

| 이미 존재하는 것 | 내용 |
|---|---|
| 소속 판별 표 | `Dangerous command blocked:` → 리포 자체 훅(`internal/hook/pre_tool.go`); `too complex to verify` → 바이너리. "메시지가 유일한 판별자" |
| 관측 표 | 2행 — (1) quoted-delimiter heredoc 본문의 중괄호(quoted key/value 랩) 트리거, **First-hand**, 비고에 "크기·파이프·백틱·명령 치환은 트리거 아님 — 6.7KB 산문 통과, 10글자 JSON 라인 거절, 본문 명령 치환은 접혀 통과" (2) 복합 명령 번들 — second-hand |
| Gap 문단 | 수용/거절 경계가 양 행 모두 미측정임을 명시 |
| Workarounds 소절 | 두 상황 두 해법 — 파일 내용 → Write 도구(관례, 우선), 복합 명령 → 단순 별개 명령 분할 (`unset … &&` 쌍은 분할 금지 주의 포함) |

**따라서 M1의 실제 델타는 두 가지다** (디스패치 문구가 암시하는 "비대칭 추가 +
우회 추가"보다 좁다 — §6 관측 차이 2):

1. **두 구분자 구분** — quoted 본문이 불확성인 **이유**(bash 가 어떤 확장도
   하지 않음 → 중괄호는 brace expansion 불가)와 unquoted 본문이 확장한다는
   경고. 현재 부재 — RED 측정: `grep -c "brace expansion"` → 0/rc=1,
   `grep -c "unquoted"` → 0/rc=1.
2. **세션 내 양방향 재측정 provenance** — 기존 행은 (제보 조사 시점의) 페어
   프로브 관측을 실었지만, 이 세션의 생재현(거절 verbatim + 미실행 증명 +
   대조군 통과) 기록은 없다. RED 측정: `grep -c "re-measured"` → 0/rc=1.

Workarounds는 이미 정확히 실려 있어 재작성하지 않는다(scope discipline —
PRESERVE 목록 근거).

## 4. 미러 기계론

- 소유자: `internal/template/rule_template_mirror_test.go` —
  `TestRuleTemplateMirrorDrift` (:137), sentinel `RULE_TEMPLATE_MIRROR_DRIFT`.
  allowlist에 본 파일 명시 (:57, SPEC-WORKTREE-ENTRY-STRATEGY-001 배선).
- 현재 패리티: `cmp <live> <mirror>` → **BYTE-IDENTICAL** (본 세션 측정).
- 배포 경로: `go:embed all:templates` → `moai init`/`moai update` 가 미러 쪽을
  뿌린다. drift 시 사용자 프로젝트가 낡은 룰을 받는다.

## 5. 템플릿 중립성 제약 — 바이트 동일과의 결합

`CLAUDE.local.md` §2.1 + `.moai/docs/template-internal-isolation-doctrine.md`
§25.1: 템플릿 금지 클래스에 이슈 번호·SPEC ID·내부 날짜가 있다(CI 가드
`template-neutrality-check.yaml`). 바이트 동일 미러 요구와 결합하면:

- 로컬 독트린에만 `#1659`를 넣을 수 없다 — 미러와 갈라진다.
- 결론: **독트린 텍스트는 처음부터 중립으로 쓴다**(하나의 텍스트가 두 트리를
  섬긴다). 이슈 번호·원장 경로·날짜는 회신 초안·업스트림 초안·SPEC artifacts에만
  산다.
- 베이스라인 (본 세션 측정): 독트린 양쪽 모두 `grep -c "1659"` → 0/0,
  rc=1. `SPEC-WORKTREE-GUARD-HEREDOC` → 0, rc=1.

## 6. GH #1659 본문 분석 (제보자 원문 기반)

- 제보자: jjjh7401, 2026-08-26. 환경: moai-adk 3.1.2 / Darwin 25.5.0 /
  Claude Code 2.1.238.
- 제보자 매트릭스 (본문 격리 실측): 파이프 85+백틱 통과 / 한글 산문 6737바이트
  통과 / 중괄호 1개 통과 / 쉼표 2항 중괄호 통과 / **명령 치환 통과(결정적
  대조군)** / **quoted key/value 중괄호(JSON 한 줄) 거절**.
- 제보자의 스코핑 — 초안 작성 시 그대로 존중한다: 살아있는 명령 텍스트의
  중괄호 거절은 합당한 설계(루프 파라미터 확장, 조건부 명령 치환 형태는 정탐);
  결함은 **확장이 원천적으로 일어나지 않는 자리에서 중괄호만 살아있는 것으로
  취급되는 것**으로 한정.
- 실무 피해: 훅 판정 JSON을 보고서에 heredoc으로 기록 불가 → 산문으로 풀어야
  했음(기록 정확도 저하).
- 제보자 제안: 명령 치환에 이미 하는 접기를 중괄호 판정에도 적용.
- 부수 관측: 스크립트 파일 우회(가드가 파일 안을 읽지 않음) — 제보자가 별개
  처리를 요청. M3 초안의 ask에서 제외하기로 함(Non-Goals).
- 관련: #1658 (위험명령 가드 쪽 별개 결함).

## 7. RED-now 베이스라인 배치 (전부 `6a46c0edb` 핀, 본 세션 관측)

| 셀 | 명령 | stdout | rc |
|----|------|--------|-----|
| 두 구분자 앵커 A | `grep -c "brace expansion" .claude/rules/moai/workflow/worktree-integration.md` | `0` | 1 |
| 두 구분자 앵커 B | `grep -c "unquoted" .claude/rules/moai/workflow/worktree-integration.md` | `0` | 1 |
| 재측정 provenance | `grep -c "re-measured" .claude/rules/moai/workflow/worktree-integration.md` | `0` | 1 |
| 초안 2종 부재 | `ls .moai/reports/t512/reporter-reply-1659.md .moai/reports/t512/upstream-draft-claude-code.md` | `No such file or directory` ×2 | 1 |
| 중립성 `1659` | `grep -c "1659" <live> <mirror>` | `:0` ×2 | 1 |
| 중립성 SPEC-ID | `grep -c "SPEC-WORKTREE-GUARD-HEREDOC" <live>` | `0` | 1 |

## 8. 카드 제약 (run이 지킬 것)

- note2 [HARD] loosening-danger: 가드 약화 조언 금지 — 우회는 거절 메시지
  자체의 안내(단순 별개 명령) + Write 도구 경로뿐.
- 초안 전용: 회신·업스트림 모두 게시 금지. **업스트림 제출은 리드 소관**,
  회신 게시는 원격 착지 후 리드.
- 어조 [HARD] (리드 추가분): 회신은 blame-shifting으로 읽히지 않게 — 검증으로
  시작, 우회는 도움, 업스트림은 제보자 발견의 대변.
- 건드리지 않음: `internal/hook/**`, t511 SPEC 디렉터리, 원장 파일 자체.

## 9. 디스패치 대비 관측 차이

1. **원장 추적 상태**: 디스패치는 원장을 "committed evidence"로 지칭했으나
   실측 `git status --short` → `?? .moai/reports/t512/` — **미추적**이다.
   run 첫 커밋(M1 커밋)이 원장을 함께 실어야 "커밋 증거"가 성립한다. plan.md
   M1 커밋 규약에 반영됨.
2. **M1 범위**: 디스패치 문구는 "비대칭 재현 + 두 구분자 구분 + 우회를
   추가하라"로 읽히지만, 비대칭 관측(관측 표 행)과 우회(Workarounds 소절)는
   이미 독트린에 존재한다(§3). 실제 델타는 두 구분자 구분 + 재측정
   provenance다 — plan.md T1이 이에 맞춰 범위를 정밀화했다. 관측 차이이지
   모순은 아니다: 디스패치의 세 요소 중 두 가지가 이미 충족된 상태라는 뜻이다.
