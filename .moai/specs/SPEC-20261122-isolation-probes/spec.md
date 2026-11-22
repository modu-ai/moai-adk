---
id: SPEC-20261122-isolation-probes
title: "four probes measure whether a lane worktree session and the macOS sandbox can run the landing command, add worktrees and write the moai home"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# four probes measure whether a lane worktree session and the macOS sandbox can run the landing command, add worktrees and write the moai home

## Intent

흐름 재설계의 전제 확인 조사(리더 전체문서 보고서 2026-11-22 6-4절과 16장 11번 항목, 사용자 선택 2026-11-22: 16장 모든 항목이 카드가 된다). 착지가 레인 작업트리 세션에서 되는지, macOS 샌드박스가 켜져도 되는지, 샌드박스가 `~/.moai` 원장 쓰기를 허용하는지, 레인 안에서 `git worktree add` 가 되는지 — 네 탐침으로 임시 저장소·임시 설정 폴더에서만 측정하고, 실제 저장소·실제 설정 폴더는 건드리지 않는다.

공식 문서 원문(통째로 읽음, 2026-11-22): worktrees.md 의 "How Claude Code enforces isolation" 절은 네 검사(주 체크아웃을 겨냥한 파일 편집, 작업 디렉터리, git 리다이렉트 `git -C`·`--git-dir`·`GIT_DIR`·`GIT_WORK_TREE`·`cd`, 검증할 수 없는 명령 모양)를 정의하고 "검사들은 경로·디렉터리·명령 텍스트를 읽을 뿐 어떤 파일을 썼는지 추적하지 않는다", 그리고 이 검사는 끌 수 없다고 한다. sandboxing.md 는 샌드박스가 기본 꺼짐이며 켜면 기본 쓰기 허용이 작업 디렉터리·사용자 임시 폴더·추가 디렉터리뿐이고, 연결 작업트리에서는 공용 `.git` 쓰기가 허용되나 그 안 `hooks/`·`config` 는 거부된다고 한다. 이 조각은 SPEC 먼저: 측정 계약을 정하고, 보고서가 네 답을 내면 카드는 보고서로 닫힌다.

## Non-goals

- 검사를 끄는 방법이나 우회: 문서가 "끌 수 없다"고 하므로 측정만 하고 우회는 설계하지 않는다.
- 실제 저장소·실제 `~/.moai`·실제 원격 대상 측정: 네 탐침 모두 임시 대상에서만 돈다(카드 본문의 규칙).
- 착지 명령 자체의 변경: 어디서 착지가 돌아가야 하는지에 대한 진술만 담는다(REQ-5).
- Linux·WSL2 샌드박스 변형: 이 개발기는 macOS 다(측정 환경 절에 기록).

## Requirements

- REQ-1: The report shall answer each of the four probes with pass or refuse, the exact message and the Claude Code version string measured with. · failure: none — a probe that could not run is answered not-measured with the reason, never guessed.
- REQ-2: The landing probe shall run the built moai binary's landing command against one temporary project and ledger from a real lane worktree session and record whether the isolation checks refuse the call and whether the landing itself completes. · failure: none — no real repository, real ledger or real remote is touched.
- REQ-3: The sandbox probes shall run the same landing call and a write into a moai home folder with the sandbox enabled, and record the exact refusal and the setting that would allow each. · failure: none — the moai home probe writes inside one temporary HOME only, and the real home is never written.
- REQ-4: The worktree probe shall measure a plain `git worktree add` and the `git -C` redirect form from a lane worktree session against one temporary repository, and record each exact message. · failure: none — the real repository's worktree list is unchanged.
- REQ-5: When a probe refuses the landing from a lane, the report shall state where landing runs instead · failure: none — the statement names the leader session in the primary checkout or the user's terminal, after the whole-docs report's finding.

## Acceptance

- AC-1 (REQ-1): the report holds four answers, each pass or refuse (or not-measured with the reason), each with the exact message and the version string.
  verify: the report file under the moai reports folder, read whole by the leader; the card closes on it (`card finish --evidence-report`)
- AC-2 (REQ-2, REQ-3): the two landing answers exist — one from the lane worktree session, one from inside the sandboxed run — and each names the exact message when refused.
  verify: the same report
- AC-3 (REQ-3): the moai-home answer names the exact refusal and the one setting (`sandbox.filesystem.allowWrite`) that would allow the write.
  verify: the same report
- AC-4 (REQ-4): the two worktree answers exist — the plain form and the `git -C` form — each with the exact message.
  verify: the same report
- AC-5 (REQ-5): the report ends with the one-sentence conclusion of where landing runs if a lane cannot land.
  verify: the same report

## Scope

- write: `.moai/specs/SPEC-20261122-isolation-probes/**`
- read-only: `.moai/specs/SPEC-20261121-flow-skeleton/**`, the whole-docs report `claude-code-native-redesign-2026-11-22.md` (6-4절·16장 11번), the official pages worktrees.md and sandboxing.md (원문, 통째로)
- outside the repository, not a scope line: the report goes to the moai reports folder of the user's HOME (경로는 Plan 3에 적는다) — 저장소 밖이라 예약 밖이며, 카드는 보고서로 닫힌다
- forbidden: any write into the real repository's refs or worktree list; any write into the real moai home folder or the Claude config folder; any change of this session's or the user's sandbox settings; printing a key or its file

## Plan

1. This SPEC, committed on the branch.
2. The probes, each recorded as it runs: the version string first; the landing probe from this lane worktree session against one temporary project; the worktree probes (plain and `git -C`) against one temporary repository; the sandbox probes inside one sandboxed run with a temporary HOME for the moai-home write.
3. The report under the moai reports folder: four answers, exact messages, the version string, the conclusion of REQ-5, and the list of what was measured live in this conversation before the probes ran (the two isolation refusals the lane session already met).
4. The card closes on the report (`card finish --evidence-report`), the SPEC's Result names it.

## Risks

- 중첩 claude 실행(샌드박스 변형 측정)이 시작조차 못 할 수 있다(임시 HOME 이 인증을 흩면): 그 탐침은 not-measured 로 답하고 이유를 적는다(REQ-1).
- 샌드박스의 "~" 해석이 HOME 환경변수를 따르는지 문서가 확언하지 않는다: 측정값을 있는 그대로 기록하고 문서 문장을 함께 적는다.
- 격리 검사의 거절문은 버전에 따라 다를 수 있다: 보고서가 버전 문자열을 담는 이유다(REQ-1).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 격리 검사 측정은 진짜 레인 작업트리 세션(이 세션)에서, 샌드박스 측정은 임시 설정으로 켠 한 번의 실행에서 한다. | worktrees.md: 검사는 세션의 작업트리에 묶인다 | 카드 열림 |
| D-2 | 버전 문자열은 측정 순간의 `claude --version` 출력 그대로다. | 카드 기준 1 | 카드 열림 |
| D-3 | moai 홈 쓰기 탐침은 임시 HOME 안의 폴더에만 쓴다; 실제 홈은 읽지도 않는다. | 카드 본문 "never the real ones"; sandboxing.md 기본 쓰기 정책은 경로와 무관 | 카드 열림 |
| D-4 | 이 대화에서 이미 만난 두 격리 거절(명령 모양·git -C 리다이렉트)은 실측으로 치고 원문 그대로 보고서에 싣는다. | 관측된 사실, 시각 기록 | 카드 열림 |

## Result

(status draft — written when the slice lands)
