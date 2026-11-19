---
id: SPEC-20261119-self-check-diff
title: "A pre-PR self-check list for implementers, with one command for the machine items"
status: draft
revision: 1
created: 2026-11-19
---

# A pre-PR self-check list for implementers, with one command for the machine items

## Intent

PR 역사 분석(752 PR·감사 FAIL 193건)에서 가장 큰 실패 클래스(경계 시험·변이 부재 63, CI 빨강 47, 심볼릭링크/FIFO 37, HOME 격리 24, 크기 한도 21, 도움말 규칙 10 등)를 implementer 가 PR 을 열기 전에 스스로 잡게 해요(사용자 선택 2026-11-17; 리더 ORIGIN note 2344 — Addy Osmani 검증 표). `skills/moai/references/internal/self-check.md` 가 검사마다 명령과 함께 한 줄씩 목록을 주고, `tools/self_check.sh` 가 기계로 확인 가능한 항목을 한 명령으로 돌려 한 줄씩 결과를 내어요. 이 SPEC 은 PR #3203·#3211 시대에 병합된 같은 일의 SDD 뒤붙이기예요(전환기 예외라 SPEC 이 없었음).

## Non-goals

- 감사 관문·CI 대체 — 이 목록은 스스로 돌리는 사전 점검이지 통과 증명이 아니에요.
- `mode-agent.md` 에 이름 올리기 — 파일이 65,509/65,536 바이트로 자리가 없어 리더 수정(2057)으로 `references/chains/index.md` 가 대신하고, mode-agent.md 는 자리가 나면 따라와요.
- 새 검사 항목 발명 — 인도된 19개 항목(시험 10·패키지·문서 4·PR 5)이 기준이에요.

## Requirements

- REQ-1: Where an implementer reads `references/internal/self-check.md`, the file shall list each pre-PR check with the command to run or the thing to look for, one line per check, grouped as Tests, Package and documents, and Pull request.
- REQ-2: When `tools/self_check.sh` runs in the moai-adk repository, it shall run the machine items in one go, keep going past a failure, and print one `self-check: <item> exit=<code>` line per item and one summary line.
- REQ-3 (REQ-2): IF a heavy item is refused by the machine's slot (exit 75), THEN the item's line shall say `slot-refused (not run)` so it is never read as passed.
- REQ-4: When a check does not fit the change or the project, the implementer shall skip it with the reason, under `Not run` or in the `## Self-Audit` section. · failure: none — 도구가 강제하는 경로가 아니라 문서 계약이고, 묵묵히 건너뛰면 감사 실패 클래스로 잡혀요.
- REQ-5: Where `agents/impl.md` and `references/chains/index.md` describe the implementer's handoff, they shall name `internal/self-check`.
- REQ-6: When `tests/self_check_test.py` runs, it shall pin the item list and its commands (whole `--bin` runs, never filtered), the keep-going rule, the skip and not_run rules, the heavy-slot refusal, the package file bound and the summary line, with stub `cargo`, `rustfmt`, `python3` and `moai` in a temporary repository with a temporary HOME. · failure: none — 고정 대상이 어긋나면 그 시험 자체가 빨개지는 것 외에 별도 거부 경로가 없어요.

## Acceptance

- AC-1 (REQ-1, REQ-5): self-check.md 가 19개 항목을 명령과 함께 나열하고, agents/impl.md 와 references/chains/index.md 가 internal/self-check 를 이름 올려요(mode-agent.md 는 리더 수정 2057 대로 이후).
  verify: `rg -c "^[0-9]+\\." skills/moai/references/internal/self-check.md`
  verify: `rg -n "internal/self-check" agents/impl.md skills/moai/references/chains/index.md`
- AC-2 (REQ-2, REQ-3): 한 명령이 기계 항목을 순서대로 돌리고 항목별 줄과 요약 줄을 내며, 슬롯 거부(exit 75)는 not run 으로 읽혀요.
  verify: `SELF_CHECK_DRY_RUN=1 tools/self_check.sh`
- AC-3 (REQ-4, REQ-6): 시험이 항목 목록·규칙들을 고정하고 전부 통과해요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p 'self_check_test.py'`

## Scope

- write: `.moai/specs/SPEC-20261119-self-check-diff/spec.md`
- read-only: `skills/moai/references/internal/self-check.md`, `tools/self_check.sh`, `tests/self_check_test.py`, `agents/impl.md`, `skills/moai/references/chains/index.md`, `tools/local_check.sh`(항목 대조), `skills/moai/references/source-manifest.json`
- forbidden: 실제 HOME 과 원장, 네트워크, 실제 원격 저장소, 키

## Plan

이 일은 병합돼 있어요(421b11a2c9 시험 구축, b9f2c07443 감사 N1-N6 수리); 이 라운드는 문서화·검증이에요.

1. 팁에서 인도물 확인: self-check.md 항목 수·명령, self_check.sh 항목 순서·dry-run, 언급 두 곳, 시험 32개.
2. 이 SPEC 작성·연결·PR(tier 0 — 문서만).

## Risks

- 기준 2 의 "PR Self-Audit 에 나열" 규약은 문서 계약이라 기계 검증이 없어요 — 감사가 몸으로 확인하는 영역이에요.
- 항목 목록이 self_check.sh 와 self-check.md 사이에서 어긋나면 시험이 잡지만, 새 항목 추가 시 두 곳과 시험을 같이 움직여야 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | mode-agent.md 대신 agents/impl.md·chains/index.md 가 이름 올림 | 리더 수정 2057(감사 B1): mode-agent.md 65,509/65,536 바이트로 자리 없음 | 자리가 나면 mode-agent.md 추가 |
| D-2 | SPEC 은 인도된 상태를 기록(새 항목 발명 없음) | SDD 뒤붙이기 라운드 — 코드 변화 없음 | 후속 카드로 항목 확장 |

## Result

- AC-1: met — 2026-11-19 이 worktree(팁 39fa580c)에서 `rg -c "^[0-9]+\\." skills/moai/references/internal/self-check.md` = 19; `rg -n "internal/self-check"` → agents/impl.md:7·52, references/chains/index.md:23.
- AC-2: met — `SELF_CHECK_DRY_RUN=1 tools/self_check.sh` 가 항목별 dry-run 줄(bin-moai…manifest·diff-check·python)과 요약 줄(`self-check: head=39fa580c… result=fail dirty=1 dry-run=1 failed=none not_run=none`)을 출력(untracked SPEC 으로 dirty=1 은 dry-run 기대 동작).
- AC-3: met — `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p 'self_check_test.py'` → Ran 32 tests, OK(같은 worktree).
