---
id: SPEC-WORKTREE-GUARD-HEREDOC-DOC-001
title: "Worktree-guard heredoc brace asymmetry — ownership pivot, doctrine extension, reporter reply, upstream draft"
version: "0.1.0"
status: draft
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

# SPEC-WORKTREE-GUARD-HEREDOC-DOC-001 — 워크트리 가드 heredoc 중괄호 접기 비대칭: 소속 전환 + 독트린 보강 + 회신·업스트림 초안

## HISTORY

| Version | Date       | Author       | Change |
|---------|------------|--------------|--------|
| 0.1.0   | 2026-09-07 | manager-spec | Initial draft. 카드 t512 (GH #1659). 생재현이 강제한 소속 전환(pivot)을 본문으로 채택. |

## 1. Overview

### 1.1 Goal

워크트리 격리 세션의 Bash 거절 가드가 따옴표 붙은 구분자(quoted-delimiter) heredoc
본문 안의 중괄호에만 접기(folding)를 적용하지 않는 비대칭(같은 자리의 명령 치환은
정상 접힘)에 대해, 이 저장소가 실제로 낼 수 있는 산출물을 낸다:

1. **독트린 보강** — `worktree-integration.md` 가드 소속 섹션에 두 구분자 구분
   (quoted = 불확성 텍스트 / unquoted = 확장 일어남)과 생재현 기록을 추가하고,
   템플릿 미러에 같은 커밋으로 바이트 동일 반영한다.
2. **제보자 회신 초안** (한국어) — 분석이 옳았고 생으로 재현됐음을 알리고,
   가드의 소속(바이너리)과 우회로를 안내한다.
3. **업스트림 제보 초안** (영어) — 비대칭 매트릭스와 두 가지 수용 가능한 수정
   방향을 Claude Code 저장소에 보낼 형태로 정리한다.

### 1.2 Problem — the pivot

카드 t512는 "리포 안의 가드를 수리한다"는 전제로 발행됐다. **생재현이 그 전제를
뒤집었다. 결함의 몸은 이 저장소 밖 — Claude Code 바이너리 — 에 있다.**

세 개의 독립 측정이 이 소속 판정을 뒷받침한다 (모두 HEAD `6a46c0edb` 기준, 재현
원장 `.moai/reports/t512/repro-heredoc-brace.md`):

1. **거절 문면이 리포 소스에 0건** — `grep -rn 'too complex to verify' internal/
   .claude/hooks/` 의 적중은 문서 인용 3건뿐이다: `worktree-integration.md:481,
   :485, :491`. 그 가드를 구현하거나 설정하는 소스는 여기에 없다.
2. **리포 자체 독트린이 소속을 명시한다** (`worktree-integration.md:481`) —
   `too complex to verify` 거절의 소속은 **The Claude Code binary**이며, 여기에
   구현도 설정도 없다("No — there is no source here that implements or
   configures it"). 이는 #1658 스레드에서 유지자가 제보자에게 말한 것과 일치한다.
3. **`branch_guard.go` 는 이 가드가 아니다** (`worktree-integration.md:485`) —
   그 파일은 1차 체크아웃의 브랜치 상태 변경을 방어하는 별개 가드이며, 편집해도
   이 거절에 영향이 없다.

동시에 **비대칭 자체는 생으로 재현됐다** (같은 세션, 양방향 2-프로브):

- 프로브 1 (제보 형태 — 따옴표 붙은 구분자 본문에 중괄호/JSON 라인): 가드가
  명령을 **실행하지 않고** 거절. 거절 문면 verbatim 확보, 미실행은 대상 파일
  부재로 증명.
- 프로브 2 (대조군 — 같은 형태, 본문에 명령 치환): 실행됨, rc=0, 본문은 리터럴
  그대로(치환 미발생 — quoted 구분자 규격대로). 즉 가드의 분석은 명령 치환을
  접어 통과시켰다.

따라서 이 리포가 낼 수 있는 것은 코드 수정이 아니라: (a) 결함의 **소속 확립**을
독트린에 새기고, (b) 기술적 핵심인 **두 구분자 구분**을 문서가 정확히 운반하게
하고, (c) 제보자에게 **올바른 채널**(업스트림)로 연결되는 초안을 준비하는 것이다.

### 1.3 Why now

- 제보자(jjjh7401)가 #1659 (2026-08-26)에서 실무 피해를 구체적으로 기술했다 —
  훅 판정 JSON을 보고서에 heredoc으로 기록하려다 거절됐고, JSON을 산문으로
  풀어야 했다. 회신과 업스트림 전달이 늦어질수록 제보자의 대기가 길어진다.
- 기존 독트린이 비대칭 **관측**은 이미 흡수하고 있지만(관측 표의 provenance 행),
  **왜** quoted 본문이 불확성인지(두 구분자 구분)는 실려 있지 않다 — 다음
  독자가 같은 오판을 같은 자리에서 다시 할 수 있다.
- 생재현 원장이 이 세션에서 이미 확보됐다. 원장을 인용하는 문서·회신·초안을
  같은 카드 안에서 마감하는 것이 귀속이 가장 깨끗하다.

### 1.4 Impact

- **수정 파일**: 2 (독트린 본체 + 템플릿 미러, 같은 커밋) + 신규 2 (회신 초안,
  업스트림 초안). 문서 전용, 예상 순증 ~50-70 LOC.
- **사용자 영향**: 배포 사용자는 `moai update` 경유로 보강된 독트린을 받는다.
  가드 동작 자체는 바이너리 소관이라 변화 없다 — 이 SPEC은 그 사실을 문서가
  정직하게 말하게 만들 뿐이다.
- **Tier 판정(정직)**: LOC·파일 수는 Tier S 범위(<300 LOC, <5파일)다. 그러나
  산출물 세트는 디스패치 지시로 Tier M 구성(spec + plan + acceptance + 연구
  문서)을 채택했고, 수용 기준 면(미러 바이트 패리티 기계 검증 + 두 구분자
  의미의 뮤턴트 저항 검토)이 독립 acceptance.md 를 요구한다. 프론트매터
  `tier: M`, 근거는 이 절에 기록.

## 2. Goals

1. 독트린에 두 구분자 구분을 새긴다: quoted 구분자 본문은 bash 가 **어떤 확장도
   하지 않는** 불확성 텍스트(중괄호는 brace expansion 이 될 수 없음), unquoted
   구분자 본문은 **확장이 일어나는** 살아있는 텍스트(어떤 가드도, 어떤 안내도
   불확성으로 다뤄선 안 됨).
2. 독트린에 양방향 생재현 기록을 중립적 문체로 추가한다 (관측 표의 기존
   provenance 행을 확장, 중복 아님).
3. loosening-danger 경계를 유지한다: 보강 내용에 어떤 가드의 약화·비활성·재구성
   안내도 넣지 않고, 소속 판정(바이너리 소유)과 "문서 수정이 거절을 고치는
   것 아님"을 명시적으로 유지한다.
4. 템플릿 미러를 같은 커밋에 바이트 동일로 반영한다 (`TestRuleTemplateMirrorDrift`
   소관, sentinel `RULE_TEMPLATE_MIRROR_DRIFT`).
5. 제보자 회신 초안(한국어)과 업스트림 제보 초안(영어)을 카드 증거 경로에
   작성한다. 초안 전용 — 게시는 이 카드의 소관이 아니다.

## 3. Non-Goals

### 3.1 Out of Scope

아래는 의도적으로 범위 밖이다. 각 항목에 이유를 남긴다:

- **모든 Go 변경** — `internal/hook/branch_guard.go` 포함. 독트린 `:485` 가
  명시하듯 그 파일을 편집해도 이 거절에 영향이 없다(소속이 다른 가드). 결함의
  몸이 바이너리에 있는 이상 이 리포의 Go 코드에는 고칠 대상이 없다.
- **바이너리 가드의 억제·재구성 시도** — loosening-danger [HARD] (카드 note2).
  문서도 어떤 가드도 약화하는 조언을 넣지 않는다. 허용되는 우회는 거절 메시지
  자체의 안내(단순한 별개 명령으로 분할)와 파일 내용의 Write 도구 경로뿐이다.
- **회신·업스트림 이슈의 게시** — 이 카드는 초안까지만 만든다. 게시는 원격
  착지 후 리드의 별도 행위다(M2/M3 초안 전용 조항 참조).
- **스크립트 파일 우회(제보자의 부수 관측)의 처리** — 제보자가 별개 처리를
  요청했고(#1659 본문), 이 초안의 ask 목록에 넣지 않는다. 본 SPEC의 M3 초안도
  그 관측을 ask 로 포함하지 않는다.
- **unquoted-delimiter 본문에 대한 가드 동작 변경 요구** — 살아있는 명령
  텍스트의 중괄호 거절은 합당한 설계다(제보자 본인이 명시). 초안은 이 동작을
  건드리는 어떤 방향도 제안하지 않는다.
- **t511 SPEC 디렉터리와 훅 소스 일체의 변경** — 카드 디스패치 금지 사항.

## 4. GEARS Requirements

> 표기: GEARS 5상 (Ubiquitous / Event-driven / State-driven / Where /
> Event-detected). 이 SPEC의 요구는 모두 항시(Ubiquitous) 성격이다 — 문서 내용
> 규정과 산출물 존재 규정이므로 트리거 이벤트가 없다.

### REQ-WGHD-001 — 두 구분자 구분의 독트린 반영 (Ubiquitous)

The doctrine shall state the two-delimiter distinction: a quoted-delimiter
(`<<'EOF'`) heredoc body is inert text on which bash performs no expansion of any
kind — parameter, command substitution, arithmetic, or brace — so a brace there
cannot be brace expansion; and an unquoted-delimiter (`<<EOF`) heredoc body DOES
expand (all four apply), so neither a guard nor any guidance in the doctrine may
treat it as inert.

근거: 이 구분이 비대칭의 기술적 핵심이다. 가드가 명령 치환을 이미 접는다는
관측(기존 provenance 행)만으로는 "왜"가 없어 다음 독자가 같은 오판을 재생산한다.
위치: `worktree-integration.md` 의 "What has been observed to trip the worktree
guard" 절 뒤, Workarounds 절 앞(§ plan.md T1 초안 참조).

### REQ-WGHD-002 — 생재현 기록의 독트린 반영 (Ubiquitous)

The doctrine shall record the live bidirectional re-measurement of the fold
asymmetry — the brace-form probe refused with the verbatim `too complex to
verify` sentence before execution (non-execution proven by the target file's
absence), the paired command-substitution probe in the same position executed and
passed with the literal preserved — in template-neutral prose, extending the
existing observation row's provenance without duplicating its measured cells.

근거: 재현 원장(리포 내부 경로)은 템플릿에 실릴 수 없으므로(REQ-WGHD-005),
독트린은 관측의 사실과 1차성(first-hand, 세션 내 측정)만 중립 문체로 운반한다.
원장 경로 인용은 SPEC artifacts(research/acceptance)의 소관이다.

### REQ-WGHD-003 — loosening-danger 경계 유지 (Ubiquitous)

The doctrine extension shall contain no guidance to disable, weaken, suppress,
or reconfigure any guard; shall preserve the existing ownership finding that the
`too complex to verify` refusal is owned by the Claude Code binary; and shall
not present the documentation extension as a fix of the refusal.

근거: 카드 note2 [HARD]. 유일한 코드-인접 변경이 문서라는 것이 이 SPEC의 정직성
조건이다 — 문서가 "고쳤다"고 읽히면 결함 소속 판정 자체가 오염된다.

### REQ-WGHD-004 — 템플릿 미러 바이트 동일 (Ubiquitous)

The template mirror
(`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`)
shall be updated byte-identically to the live rule, in the SAME commit, parity
owned by `internal/template/rule_template_mirror_test.go`
(`TestRuleTemplateMirrorDrift`, sentinel `RULE_TEMPLATE_MIRROR_DRIFT`; the file
is on the test's explicit allowlist).

근거: `go:embed` 로 배포되는 것이 미러 쪽이다. 측정 기준(2026-09-07, HEAD
`6a46c0edb`) 현재 두 파일은 `cmp` 바이트 동일이다 — 이 불변을 M1 편집이 깨지
않게 하는 것이 기계 검증 가능한 수용면이다.

### REQ-WGHD-005 — 독트린 텍스트의 템플릿 중립성 (Ubiquitous)

The mirrored doctrine text shall remain template-neutral — no issue numbers, no
SPEC IDs, no internal dates, no internal report paths — so that one text serves
both trees (live rule and template mirror) byte-identically.

근거: §2.1 템플릿 중립성(CI 가드 `template-neutrality-check.yaml`)과 바이트
동일 미러는 결합 제약이다: 로컬판에만 `#1659`를 넣으면 미러와 갈라진다. 따라서
이슈 번호·SPEC ID·날짜·리포 경로는 회신·업스트림 초안과 SPEC artifacts에만
살고, 독트린 텍스트는 처음부터 중립으로 쓴다.

### REQ-WGHD-006 — 제보자 회신 초안 (Ubiquitous)

The card evidence set shall carry a reporter-reply DRAFT in Korean at
`.moai/reports/t512/reporter-reply-1659.md` carrying four elements: (a)
acknowledgment that the analysis was correct and was reproduced live (the
asymmetry matrix cited), (b) the guard's ownership — the Claude Code binary —
citing this repository's own doctrine as the source of that finding, so no
in-repo fix exists, (c) the two workarounds (the `Write` tool for file content;
plain, separate commands), (d) the upstream-report note (a draft exists; posting
follows after the card lands). Tone constraint (lead addendum, [HARD]): the
draft MUST NOT read as blame-shifting — it opens by validating the reporter's
analysis, offers the workarounds as help, and frames the upstream handoff as
advocating the reporter's finding, not as declining ownership of the report.
Draft only — the lead posts it after remote landing.

### REQ-WGHD-007 — 업스트림 제보 초안 (Ubiquitous)

The card evidence set shall carry an upstream-issue DRAFT in English at
`.moai/reports/t512/upstream-draft-claude-code.md` carrying five elements: (a)
the asymmetry matrix with both probes verbatim (refusal sentence and the passing
control), (b) the argument that a quoted-delimiter body is provably inert, (c)
fix direction 1 — fold braces in quoted-delimiter heredoc bodies, the same fold
already applied to command substitutions, (d) fix direction 2 — make the
refusal message name the offending construct so users can restructure, (e)
credit to the reporter's original observation. Draft only — submission to the
upstream tracker stays with the lead; this card does not post it.

## 5. Affected Files

### MODIFY (2)

| File | Est. delta | Change |
|------|-----------|--------|
| `.claude/rules/moai/workflow/worktree-integration.md` | ~20 LOC added | 관측 절 뒤 두 구분자 구분 소절 + 생재현 provenance 문장 (REQ-WGHD-001/002/003). 영어, 중립 문체. |
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | byte-identical | 같은 커밋, `cmp` 동일 유지 (REQ-WGHD-004/005). |

### NEW (2)

| File | Est. delta | Change |
|------|-----------|--------|
| `.moai/reports/t512/reporter-reply-1659.md` | ~40 LOC | 제보자 회신 초안, 한국어 (REQ-WGHD-006). |
| `.moai/reports/t512/upstream-draft-claude-code.md` | ~70 LOC | 업스트림 제보 초안, 영어 (REQ-WGHD-007). |

### PRESERVE (건드리지 않음)

- `internal/hook/**` 전체 (branch_guard.go 포함) — 소속 다른 가드, 편집 무효.
- 기존 "Workarounds" 소절 — 이미 두 우회를 정확히 실리고 있어 재작성 금지
  (scope discipline; REQ-WGHD-001 위치 선정 근거이기도 하다).
- 관측 표의 기존 행(제보 측 매트릭스) — M1 은 확장이지 덮어쓰기가 아니다.
- `.moai/reports/t512/repro-heredoc-brace.md` — 커밋 예정 증거, 편집 금지.
- t511 SPEC 디렉터리.

## 6. References

- 재현 원장: `.moai/reports/t512/repro-heredoc-brace.md` (커밋 예정 증거 —
  이 SPEC 작성 시점에 워크트리에 존재, 미추적 상태. run 첫 커밋이 함께 실는다)
- 독트린 대상 섹션: `.claude/rules/moai/workflow/worktree-integration.md`
  § "Refused Commands in a Worktree-Isolated Session" (`:475` 전후 — 소속 표
  `:481`, branch_guard 구분 `:485`, 거절 전문 `:491`)
- 미러 패리티 소유자: `internal/template/rule_template_mirror_test.go`
  (`TestRuleTemplateMirrorDrift`; allowlist `:57` 에 본 파일 명시)
- 원 제보: GH issue #1659 (jjjh7401, 2026-08-26; moai-adk 3.1.2 / Darwin 25.5.0
  / Claude Code 2.1.238) — 관련 스레드 #1658 (위험명령 가드 쪽 별개 결함)
- 템플릿 중립성: `CLAUDE.local.md` §2.1 + `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (C1-C8)
- 프론트매터 스키마: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- 수용 기준 규율: `.claude/rules/moai/development/verification-completeness.md`
  (two-cell, §1.2 3-part check, §2.1 RED 4요소)
