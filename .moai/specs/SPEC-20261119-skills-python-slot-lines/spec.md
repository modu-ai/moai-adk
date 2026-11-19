---
id: SPEC-20261119-skills-python-slot-lines
title: "The skill documents name the python slot and the --pre-ready list that SPEC-20261119-python-slot-pre-ready adds to the scripts"
status: draft
revision: 1
created: 2026-11-19
---

# The skill documents name the python slot and the --pre-ready list that SPEC-20261119-python-slot-pre-ready adds to the scripts

## Intent

SPEC-20261119-python-slot-pre-ready(C-114 조각 2)는 파이썬 전체 실행을 전용 슬롯(한도 2)에 넣고, READY 전에 돌릴 전체 트리 점검을 `tools/self_check.sh --pre-ready` 한 줄로 모았어요. 그런데 레인과 감사가 읽는 문서(`internal/self-check.md`, `workflows/mode-agent.md` 의 heavy-checks 단계, `workflows/verify.md` 의 "Heavy checks are serialised")는 아직 예전 글이에요: 파이썬 전체 실행을 어떻게 감싸는지와 READY 전에 무엇을 통과해야 하는지를 이름으로 적지 않아서, 레인은 스크립트가 있다는 것을 모르고 전체 `unittest discover` 를 맨몸으로 돌리거나 점검을 건너뛰어요(2026-11-19 에 한 묶음이 READY 에 몇 시간 못 올라간 일의 원인이에요, C-114 조각 1 계획). 카드 C-114 의 기준 1 뒷부분("lane push 체크리스트가 READY 전 전체 트리 점검을 이름으로 적는다")이 이 조각이에요. 리더가 조각 2 에서 `skills/` 를 뺐고(노트 2921), 조각 3 으로 이어 주었어요.

이 조각은 조각 2 의 head(5d72a44fbd) 위에 서요. 문서가 가리키는 `--pre-ready` 와 파이썬 슬롯은 그 head 에만 있으므로, 조각 2 가 먼저 묶음에 들어가야 해요.

## Non-goals

- `tools/self_check.sh`, `tools/local_check.sh` 와 그 시험, `docs/ci-rc-ci-spec.md` 바꾸기 — 조각 2 가 끝냈어요.
- READY 전 점검 항목을 늘리거나 줄이기, 슬롯 한도 바꾸기.
- `workflows/sync.md:23`, `AGENTS.local.md`, `.github/pull_request_template.md` 의 "전체 discover" 문장 — 다른 조각이에요(필요하면 리더가 배차).
- 규칙 가지치기(C-103)와 mode-leader.md.

## Requirements

- REQ-1: Where `skills/moai/references/internal/self-check.md` describes `tools/self_check.sh`, it shall name `--pre-ready`, its eight items in the order the script prints them, and the rule that a head passes the list before the lane reports it ready.
- REQ-2: Where `skills/moai/references/internal/self-check.md` describes whole Python runs, it shall name the python slot with its two variables `MOAI_PYTHON_SLOT_DIR` and `MOAI_PYTHON_SLOT_MAX` (default 2) and say that an item the slot refused is reported as not run.
- REQ-3: Where `skills/moai/workflows/mode-agent.md` describes the heavy-checks step, it shall tell the lane to run `tools/self_check.sh --pre-ready` on the head before its READY line and to run a whole Python suite only through the python slot.
- REQ-4: Where `skills/moai/workflows/verify.md` describes "Heavy checks are serialised", it shall name the python slot as the separate slot of whole Python runs, apart from the slots of the cargo runs.
- REQ-5: When a file under `skills/moai/` changes, `skills/moai/references/source-manifest.json` shall hold the sha256 that `tools/source_manifest.py write` writes for that file. · failure: none — `tools/source_manifest.py check` is the check that fails when a hash is stale.
- REQ-6: When `tests/self_check_test.py` runs, it shall read the three real documents and pin that each names what REQ-1 to REQ-4 require, so a later edit that drops a name turns a test red. · failure: none — a test that does not find the name fails, there is no other path.

## Acceptance

- AC-1 (REQ-1, REQ-2): `internal/self-check.md` names `--pre-ready`, its eight items in order, the READY rule, the python slot's two variables and the not-run rule.
  verify: `rg -n -e "--pre-ready" -e "MOAI_PYTHON_SLOT_MAX" -e "MOAI_PYTHON_SLOT_DIR" skills/moai/references/internal/self-check.md`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.self_check_test`
- AC-2 (REQ-3, REQ-4): `mode-agent.md` heavy-checks and `verify.md` "Heavy checks are serialised" name the READY list and the python slot.
  verify: `rg -n -e "--pre-ready" -e "python slot" skills/moai/workflows/mode-agent.md skills/moai/workflows/verify.md`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.self_check_test`
- AC-3 (REQ-5): the manifest is current and the package fits.
  verify: `python3 -I tools/source_manifest.py check`
  verify: `cargo run --offline --locked --quiet --bin moai -- package-inspect skills/moai`
- AC-4 (REQ-6): the document pins pass, and the tests that read these files still pass.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.self_check_test tests.skill_references_test tests.source_manifest_test tests.repo_scans_test`
  verify: `cargo test --offline --locked --test agent_defs_package`

## Scope

- write: `.moai/specs/SPEC-20261119-skills-python-slot-lines/spec.md`, `skills/moai/references/internal/self-check.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/verify.md`, `skills/moai/references/source-manifest.json`, `tests/self_check_test.py`
- read-only: `tools/self_check.sh`, `tools/local_check.sh`, `docs/ci-rc-ci-spec.md`, `tests/agent_defs_package.rs`, `tests/workflow_size_budget.rs`
- forbidden: 실제 HOME 과 원장, 네트워크, 키, 위 write 밖의 `skills/` 파일

## Plan

1. SPEC 을 쓰고 `spec check` 오류 0 을 확인한다.
2. 첫 RED: `tests/self_check_test.py` 에 세 문서가 이름을 적는지 읽는 시험을 먼저 쓰고, 지금 글로는 실패하는 것을 본다.
3. 세 문서에 한 문장씩(필요하면 두 문장) 더한다. 파일마다 `wc -c` 로 65,536 바이트 한도와 여유를 확인한다.
4. `python3 -I tools/source_manifest.py write` 뒤 `check`, 빌드한 바이너리로 `package-inspect`.
5. 점검: 위 AC 의 명령, `tests.agent_defs_package`(cargo), `tools/self_check.sh --pre-ready` 진짜 실행, 변이(문장 하나 지우기) 시험이 빨개지는지.

## Risks

- `mode-agent.md` 와 `source-manifest.json` 은 tip 도 바꾸고 있어서(C-99 의 flaky 단계 등) 병합이 어긋날 수 있어요: push 전에 `git merge-tree` 와 병합 결과의 manifest 검사로 확인해요.
- 글자 수를 못박는 시험(`workflow_size_budget`)이 mode-agent.md 를 읽어요: 더하는 글의 바이트를 재고 그 시험을 돌려요.
- 문서 문장이 코드보다 먼저 낡을 수 있어요: 항목 이름은 `tools/self_check.sh` 가 찍는 이름 그대로 시험이 읽어 맞춰 봐요(조각 2 의 시험이 쓰는 목록 상수와 같은 값).

## Decisions

| id | decision | why | reversible |
| --- | --- | --- | --- |
| D-1 | 세 문서와 manifest 와 시험 한 파일만 바꾼다 | 리더 배차("skills 쪽 줄")와 C-114 계획이 정한 범위 | 항목 추가 |
| D-2 | push 점검표의 자리는 mode-agent 의 heavy-checks 단계(7)다 | pr-gate(8)는 PR 개수 게이트이고 M-Flow 는 PR 을 열지 않아서, "푸시 전 무엇을 돌리나"를 말하는 곳은 7번이에요 | 문장 이동 |
| D-3 | 시험은 문서가 이름을 적는지만 읽는다(문장 전체 못박기 없음) | 문장을 다듬을 때마다 시험이 깨지는 것을 막으려고 이름 목록만 읽어요 | 시험 추가 |
| D-4 | manifest 해시는 이 가지의 마지막 커밋에서만 쓴다 | 리더 결정(C-59 노트 3081, 2026-11-19): `source-manifest.json` 을 C-59 가 먼저 쓰고, C-114 는 그 뒤에 다시 받아요 | 마지막 단계에서 `write` |

## Result

- AC-1: met — 2026-11-19 이 worktree 에서 `rg -n -e "--pre-ready" -e "MOAI_PYTHON_SLOT_MAX" -e "MOAI_PYTHON_SLOT_DIR" skills/moai/references/internal/self-check.md` 가 5번째 줄 하나를 찾아요(그 한 줄이 `--pre-ready`, 여덟 항목, 파이썬 슬롯 폴더와 두 환경 변수, 거절은 not run). `python3 -m unittest tests.self_check_test` 는 새 SkillDocsNameThePythonSlotTest 7개를 포함해 통과(`tests.self_check_test tests.skill_references_test tests.repo_scans_test` 를 파이썬 슬롯에서 한 번에 돌린 Ran 96 tests, OK; 감사 수리 뒤 head 532dd257c7).
- AC-2: met — `rg -n -o -e "--pre-ready" -e "python slot" skills/moai/workflows/mode-agent.md skills/moai/workflows/verify.md` 가 mode-agent.md:63 에서 둘 다, verify.md:45 에서 `python slot` 을 찾아요. 같은 시험(위 96개 안)이 통과해요.
- AC-3: open — `cargo run --offline --locked --quiet --bin moai -- package-inspect skills/moai` 는 exit 0, 세 문서는 65,536 바이트 한도 안(self-check.md 5,605, mode-agent.md 48,774, verify.md 25,400 바이트, head 532dd257c7). 그러나 `python3 -I tools/source_manifest.py check` 는 이 head 에서 exit 1 이에요(`changed`: self-check.md, mode-agent.md, verify.md): manifest 파일이 C-59 몫이라 이 가지에서 아직 쓰지 않았어요(D-4). 마지막 단계에서 `write` 한 뒤 이 줄을 `met` 로 바꿔요.
- AC-4: open — 처음 6개였던 시험은 문서를 고치기 전 5개 실패(RED 커밋 35931ca4ab, `failures=5`)에서 통과로 바뀌었어요. 감사 뒤 7개로 조였고(아래 감사 수리), 문서 한 곳씩 바꾸는 변이 9개(M1~M5: --pre-ready 이름 지움, python slot 이름 바꿈, 항목 두 개 순서 바꿈, 변수 이름 바꿈, READY 규칙 지움; H1~H3: 목록에서 `diff-check` 지움, `diff-check` 를 끝으로 옮기고 아홉째 항목 더함, "never as passed" 와 "(default 2)" 바꿈)가 각각 시험 하나만 빨갛게 만들었어요(원복은 `cmp` 로 확인, head 13776ffb34 의 시험으로). `cargo test --offline --locked --test agent_defs_package` 는 19개 통과, `--test workflow_size_budget` 는 4개 통과(head 532dd257c7). 다만 이 head 에서 `python3 -m unittest tests.source_manifest_test` 는 43개 중 1개(`test_this_repository_manifest_is_what_the_tool_writes`)가 manifest 해시가 낡았다고 실패하고, `product_cli skill_package::source_manifest_covers_every_package_file` 도 같은 이유로 실패해요(그 시험은 이번 라운드에 다시 돌리지 않았고, 앞 라운드(가지를 새 커밋 위로 옮기기 전) head e7b8b0b5cd 에서 `self-check.md: manifest sha256 is stale` 로 실패했어요). 마지막 단계의 `write` 뒤에 다시 돌려 이 줄을 `met` 로 바꿔요.
- 감사 수리: 독립 감사(새 `moai-audit` 에이전트, head 21b17bd40a — 가지를 조각 2 의 트레일러 수리 커밋 위로 옮기기 전 sha, 트리 460790c4b7 은 옮긴 뒤와 같아요 —, 권고 PASS, 막는 발견 없음)가 막지 않는 발견 5개를 적었고 모두 이 조각에서 고쳤어요. N1(`--pre-ready` 목록 시험이 파일 전체를 훑어 첫 항목 삭제·끝으로 옮김·아홉째 항목을 못 잡음, 변이 H1·H2)과 N2(`not run` 과 기본값 2 가 안 못박혀 H3 이 통과)와 N3(READY 규칙 정규식이 파일 전체에서 한 번만 맞으면 통과)는 시험을 조였어요: 목록은 `--pre-ready` 문장 안에서만 읽어 `PRE_READY_LABELS` 와 통째로 비교하고, 거절 규칙과 `(default 2)` 는 짧은 꼬리로 못박아요(커밋 13776ffb34; D-3 는 그대로: 꼬리만 읽고 문장 전체는 안 읽어요). N4(mode-agent.md 가 이 스킬에 정의가 없는 `LANE-READY-INTEGRATION` 이름을 썼음)와 N5(self-check.md 의 줄 모양·슬롯 항목 이름·"any other" 표현, verify.md 의 "설치돼 있을 때만 감싼다")는 문서를 고쳤어요(커밋 532dd257c7; 항목별 슬롯은 `tools/self_check.sh` 에서 읽었어요: diff-check·fmt·fmt-ops·manifest 는 `light`, py-source-manifest 와 py-repo-scans 는 파이썬 슬롯, skill-package 는 heavy). 감사가 따로 적은 verify.md:45 첫 문장 "at most one full build or full test run" 이 슬롯 개수와 어긋나는 것은 이 조각 전부터 있던 것이라 건드리지 않았어요(다른 조각 몫).
- 아직 안 한 것: manifest `write`(D-4). `tools/self_check.sh --pre-ready` 진짜 실행은 앞 라운드(가지를 새 커밋 위로 옮기기 전) head e7b8b0b5cd 에서 `diff-check`, `fmt`, `fmt-ops`, `py-repo-scans` 가 exit 0, `manifest`, `py-source-manifest`, `skill-package` 는 위 이유로 exit 1, `spec-check` 는 이 Result 가 없던 때 exit 1(`closure_result`)이었어요; 감사 수리 뒤 head 에서는 다시 돌리지 않았고(manifest 가 아직 낡아서 같은 세 항목이 빨강일 것이므로), manifest 단계 뒤에 한 번에 돌려요. 감사 수리 뒤 head 를 같은 감사 에이전트가 다시 읽어요(델타만).
