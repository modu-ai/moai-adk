---
id: SPEC-20261119-omp-host
title: "OMP as a full host - criterion-5 pre-check evidence and the observed skill discovery path (slice C)"
status: draft
revision: 1
created: 2026-11-19
parent: null
---

# OMP as a full host - criterion-5 pre-check evidence and the observed skill discovery path (slice C)

## Intent

카드 C-62의 마지막 열린 기준은 기준 5(사람이 실제 OMP 세션에서 moai 스킬과 agent 다섯 종의 로드를 관찰)예요. 조각 A(조사표, #3203)와 조각 B(론처, #3211)는 이미 병합됐고, 리더는 2026-11-19에 기준 5를 사용자 확인 대상으로 남겼어요(카드 노트 2380). 이 조각 C는 코드를 바꾸지 않고 두 가지를 해요.

첫째, agent-8이 2026-11-19에 무비용(모델 호출 없음)으로 실행한 사전 점검 세 가지를 원장 노트(note 2405)에서 이 추적되는 SPEC 으로 옮겨요:

1. `moai omp -l --print-command`(설치 build 7, 4.2647.7)가 argv `["omp","--approval-mode","yolo"]`, `executed:false`를 만드는 것 — 론처 구성 정상.
2. `omp skills list`(실제 OMP 18.18.7 탐색, 스킬 22개)가 moai 스킬을 `source claude:user`, 파일 `<홈>/skills/moai/SKILL.md`로 발견하는 것 — OMP가 Claude 사용자 skill 폴더까지 읽어요. 이것은 문서가 가정한 읽기 폴더(`.agents/skills`, `~/.omp/agent/skills`)보다 넓어서, 카드 노트 2405가 "문서 갱신 후보"로 남겼어요.
3. `~/.omp/agent/agents/`에 `moai-{audit,docs,impl,plan,research}.md` 다섯 파일이 있는 것(2026-11-18 17:04 기록).

둘째, 위 2의 관찰을 `docs/host-support-matrix.md`의 omp 행에 날짜와 함께 반영해요: skill 설치 gap 두 행의 폴더 가정 옆에 Claude 사용자 폴더 관찰을 더하고, 조사표 안내 문단의 `not_run` 문장을 세분화해요.

PR은 열지 않아요. 기준 5가 사람 관찰로 열려 있어 카드가 아직 닫히지 않으므로, 이 조각은 spec 가지(`spec/20261119-omp-host-observation`)에만 push 해요(리더 배차 2026-11-19).

### 기준 5 관찰 지침 (사람이 따르는 단계)

카드 노트 2405가 남긴 관찰 절차를 옮긴 것이고, 이 관찰이 이루어져야 기준 5가 판정돼요(그 전까지 not_run, 2026-11-19 기록):

1. 프로젝트 폴더에서 `moai omp -l`을 실행해 OMP 세션을 열어요.
2. 열린 세션에서 `/skill:moai`를 입력해 moai 스킬이 로드되고 응답하는지 봐요.
3. 하위 에이전트 호출 목록에서 `moai-plan`·`moai-impl`·`moai-audit`·`moai-research`·`moai-docs` 다섯 종이 보이는지 확인해요.

## Non-goals

- 실제 OMP 세션에서 `/skill:moai` 로드와 agent 다섯 종 노출을 관찰하는 것(기준 5, 사람 관찰 — 사용자 확인 대상).
- `src/**`의 코드 변경과 skill 설치 목록 구현(leader card, `src/update.rs`는 카드 C-47 범위).
- 나머지 gap 행(출력 스타일, 감사 독립성)과 리더 카드의 처리.
- 카드 노트 이외의 원장·결정 로그 편집.

## Requirements

- REQ-1: The SPEC shall record the three zero-cost pre-check observations for criterion 5 with their observation date.
- REQ-2: The SPEC shall mark the in-session load observation as not_run with the reason that a person must observe it.
- REQ-3: The SPEC shall record the criterion-5 observation guide (the steps a person follows) so the observation is repeatable.
- REQ-4: The survey rows about OMP skill discovery shall name the observed Claude-user discovery path (`source claude:user`) with its observation date, next to the folder assumptions they already carry.
- REQ-5: The survey's intro paragraph shall state that the zero-cost `omp skills list` probe observed the skill discovery path while the in-session load observation stays open for a person.
- REQ-6: The prose edits shall keep every omp `gap` row naming its closer and the table structure that `tests/omp_host_survey_test.py` asserts.
- REQ-7: The change shall carry its own SPEC at `.moai/specs/SPEC-20261119-omp-host/spec.md`, tracked by git in the same commit.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 이 SPEC이 사전 점검 세 가지와 관찰 지침을 2026-11-19 날짜와 함께 담고, 실제 세션 관찰을 not_run(사람 필요)으로 적어요.
  verify: `git grep -n "claude:user" .moai/specs/SPEC-20261119-omp-host docs/host-support-matrix.md`
- AC-2 (REQ-4, REQ-5, REQ-6): 문서의 안내 문단과 skill 설치 gap 두 행이 관찰을 더했고 구조 시험이 그대로 green이에요.
  verify: `python3 -I tests/omp_host_survey_test.py`; `python3 -I tests/host_matrix_test.py`
- AC-3 (REQ-7): SPEC이 git에 추적되고 SPEC 검사기가 받아들여요.
  verify: `git ls-files .moai/specs/SPEC-20261119-omp-host`; `moai-cli spec check SPEC-20261119-omp-host` (이 트리 빌드)
- AC-4: 배차가 시킨 push 문이 green이에요.
  verify: `cargo fmt --check`; `cargo clippy --offline --locked --all-targets -- -D warnings`; `python3 -m unittest tests.repo_scans_test tests.source_manifest_test`

## Scope

- write: `.moai/specs/SPEC-20261119-omp-host/spec.md`, `docs/host-support-matrix.md`
- read-only: `src/**`, `tests/**`, 카드 원장 노트
- forbidden: 실제 HOME(이 라운드는 이미 끝난 agent-8 기록을 옮기는 것), 네트워크, host 실행, PR

## Plan

1. `spec.md`를 쓰고(REQ 한 줄씩), `docs/host-support-matrix.md`의 조사표 안내 문단과 skill 설치 gap 두 행에 2026-11-19 관찰을 더해요.
2. 검사: `python3 -I tests/omp_host_survey_test.py`, `python3 -I tests/host_matrix_test.py`, 배차의 push 문 세 가지(가지당 한 번, 통과를 읽은 뒤 push).
3. `## Result`를 실제 출력으로 채우고 커밋한 뒤 spec 가지를 push 해요(PR 없음). 기준 5 판정은 사람의 관찰을 기다려요.

## Risks

- 산문만 바꾸는 조각이라 pinning 시험이 먼저 붉어지지 않는다(red=문서만): 표 구조 규칙은 `tests/omp_host_survey_test.py`의 통과로 지키고, 문장 자체는 감사와 리더의 읽음으로 검토한다.
- 관찰 기록은 agent-8의 원장 노트(note 2405)를 옮긴 것이라 재현 명령이 없다: `omp skills list`의 출력은 노트에 요약만 남고 세션·시각은 날짜뿐이다. 기준 5의 최종 증거는 사람의 관찰이다.
- `<홈>/skills/moai/SKILL.md` 표기는 개인 경로를 감추지만, OMP가 실제로 읽는 폴더의 전체 목록은 leader card가 확인할 때까지 가정과 관찀이 섞여 있을 수 있다.

## Decisions

| id | decision | why | revisit |
| --- | --- | --- | --- |
| D-1 | PR 없이 spec 가지만 push한다 | 기준 5(사람 관찰)가 열려 카드가 닫히지 않음; 리더 배차 2026-11-19가 "spec 가지 push 만, PR 없음" | 기준 5가 판정되면 그때의 절차(card pr → card done)를 따름 |
| D-2 | 조각 C는 코드를 바꾸지 않는다 | 남은 레인 작업은 기록과 문서 갱신뿐; 설치 목록은 leader card(`src/update.rs`는 카드 C-47 범위) | 리더 카드 착수 시 |
| D-3 | 관찰 기록의 파일 경로는 `<홈>/skills/moai/SKILL.md` 형태로 적는다 | 개인 경로를 저장소에 적지 않는다(AGENTS.local.md) | 없음 |

## Result

- AC-1: met — `git grep --cached -c "claude:user" .moai/specs/SPEC-20261119-omp-host docs/host-support-matrix.md`가 SPEC 5줄(의도의 사전 점검 2번, REQ-4, AC-1 verify 문장, 이 증거 문장, 수리 1 기록)과 docs 3줄(안내 문단 116행, skill 설치 gap 행 133·134행)을 센다. 사전 점검 세 가지는 Intent의 목록에, 관찰 지침(사람이 따르는 세 단계)은 Intent의 '기준 5 관찰 지침' 소절에, 실제 세션 관찰의 not_run(사람 필요)은 REQ-2와 Non-goals 첫 줄에 있다(2026-11-19 날짜와 함께).
- AC-2: met — `python3 -I tests/omp_host_survey_test.py` 9 tests OK, `python3 -I tests/host_matrix_test.py` 5 tests OK: 2026-11-19 관찰 문장을 넣은 뒤에도 표 구조·gap 닫는 곳·(list) 규칙이 그대로다. red=문서만(산문과 기록 추가라 pinning 시험이 먼저 붉어지지 않는다 — 계획 노트에 적은 대로).
- AC-3: met — `git ls-files .moai/specs/SPEC-20261119-omp-host` = spec.md; `moai-cli spec check SPEC-20261119-omp-host`(이 트리 빌드) errors 0 warnings 0. 첫 검사는 3 error였다: req_gears 2(REQ-1·REQ-5가 한 줄에 `shall` 둘 — GEARS는 하나만 허용)과 section_missing(Result 절). REQ를 나누고 Result를 채워 해소했고, 최종 검사는 통과다.
- AC-4: met — `cargo fmt --check` 통과(FMT_OK); `cargo clippy --offline --locked --all-targets -- -D warnings`(heavy-slot-acquire) Finished 49.19s 경고 0; `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.repo_scans_test tests.source_manifest_test` 68 tests OK. clippy의 첫 시도는 공유 캐시 `/tmp/moai-rc-product-cargo`의 `no matching package named cap-fs-ext`(알려진 결함)로 exit 101이었고, idle인 `/tmp/moai-rc-cargo-agent-22`(크레이트 548)의 registry를 `/tmp/moai-rc-cargo-agent-6`로 복사한 뒤 `cargo fetch --offline --locked`(FETCH_OK)로 검증하고 다시 돌려 통과했다.
- 돌리지 않은 것: 실제 OMP 세션 관찰(기준 5, 사람 — 사용자 확인 대상), host 실행, 전체 `cargo test`(CI가 인용), PR 열기(리더 배차 2026-11-19: spec 가지 push 만).
- 수리 1: 감사(agent audit-c62-38d5, head 38d5b95868) FAIL의 차단 B1 — REQ-3·AC-1이 약속한 관찰 지침이 SPEC에 없음 —을 Intent의 '기준 5 관찰 지침' 소절(note 2405의 세 단계, 2026-11-19) 추가로 닫고, 비차단 N1 — docs 134행이 `source claude:user` 토큰 없이 산문으로만 경로를 적음 —을 토큰 추가로 닫았다. 수리 뒤 fmt·구조 시험·스캔·spec check을 다시 돌려 통과했고, clippy는 diff가 md 전용이라 다시 돌리지 않았다(첫 head의 통과 기록을 인용).
