---
id: SPEC-20261119-audit-eval
title: "Audit-agent eval: replay incidents that mattered, score recall and over-flagging"
status: draft
revision: 1
created: 2026-11-19
---

# Audit-agent eval: replay incidents that mattered, score recall and over-flagging

## Intent

감사 에이전트(`moai-audit`)는 수십 개 PR 의 관문이지만, 무엇을 잘 잡는지는 사고 밖에서 측정된 적이 없어요.
이 저장소에는 실측 자료가 이미 있어요: 감사 실패 판정의 막는 항목(B-items)은 "리뷰 시점에 있던 진짜 버그"라는
이름표가 붙은 사례예요. 이 SPEC 은 그 사고를 재현하는 평가 세트를 만들어요 — 각 사고는 감사 댓글이 가리키는
리뷰-시점 head/base ref 와 한 문장 정답키로 기록되고, 통과·병합 PR 들이 크기를 맞춘 대조군이 돼요. 채점 도구는
감사 프롬프트를 조립해 돌려주고(dry-run), 결과물이 정답키의 버그를 실제로 서술했는지만 점수로 삼아요 — 원시
finding 회수율과 최종 판정 회수율을 나누고, 대조군에서의 과다지목을 함께 보고해요(오서먼 'Test your checks on
what actually matters', 리더 ORIGIN note 2341).

## Non-goals

- 감사 에이전트의 자동 실행(모델 호출) 없음: `run` 은 프롬프트를 만들어 저장하는 dry-run 이 기본이고, 시험은
  네트워크를 전혀 쓰지 않아요. 살아있는 실행은 별도 지시가 있을 때 호스트의 에이전트가 프롬프트를 받아 돌려요.
- 새 사고의 수집 자동화 없음: 사고 선정은 사람이 읽고 정답키를 한 문장으로 쓰는 일이에요(도구가 대신하지 않아요).
- 팁 push 로 붉어진 사고(macOS 전용 시험류) 가운데 이 세트에 못 들어간 것: 이 클래스는 CI-red-at-head 사고
  (#3199, #3212)로 대표하고, 별도 사례는 후속 카드가 목록화해요.

## Requirements

- REQ-1: When the eval set is built, `.moai/evals/audit/` shall hold at least ten incident cases — each naming its pull request, the audit comment, the review-time `head` and `base` refs of the repository, and a one-sentence answer key that states the bug a blocker of that audit found — split into separate tuning and held-out files with no case in both, and a size-matched control file of merged pull requests whose audits passed, at least one control per two incidents.
- REQ-2 (REQ-1): IF a case or control record lacks its pull request, audit comment, `head`, `base` or answer key (controls excepted), THEN the tool shall refuse, naming the file and the record, because a case that cannot name its evidence cannot be scored.
- REQ-3: When `tools/audit_eval.py run` is executed (dry-run by default), it shall build for each case the audit prompt an agent would receive — the exact head and base and where to read them — without the case's answer key, write the prompts to the output folder, and start no network call and no model.
- REQ-4 (REQ-3): IF the output folder cannot be written, THEN the command shall refuse, naming the folder, before writing any prompt.
- REQ-5: When `tools/audit_eval.py score` is executed on a findings file, it shall count, per case, only findings that describe the case's bug — a finding counts when at least half of the answer key's distinctive tokens appear in it — and report the raw-finding recall and the final-verdict recall separately, and the over-flagging rate on the controls.
- REQ-6 (REQ-5): IF the findings file is not a JSON list, or names a case id the chosen set does not hold, THEN the command shall refuse, naming the file and the id, having scored nothing.
- REQ-7: When `.moai/evals/audit/README.md` is read, it shall say how to add a case, how to run and score, and when to stop tuning — the next tuning step is overfitting when it raises the tuning score without raising the held-out score. · failure: none
- REQ-8: IF a case or control names a `head` or `base` the local repository does not hold, THEN `run` and `score` shall name that case and its missing ref and skip it, counted in the report as `missing_ref`, never silently.
- REQ-9: IF the tuning and held-out files share a case id, THEN `run` and `score` shall refuse, naming both files and the shared id, because a case scored twice measures nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2): 세트가 사고 12건(tuning 6 + held-out 6, 겹치는 id 없음)과 대조군 6건 이상을 들고, 각 사고에 pr·audit_comment·head·base·한 줄 answer_key 가 있으며, 필드가 모자란 기록은 거부돼요.
  verify: `python3 -m unittest tests.audit_eval_test`
- AC-2 (REQ-3, REQ-4, REQ-8, REQ-9): `run` 의 dry-run 출력 프롬프트에 head·base 가 들어가고 answer_key 는 들어가지 않으며, 로컬에 없는 ref 는 missing_ref 로 이름 나오고, tuning·held-out 에 같은 id 가 있으면 거부해요.
  verify: `python3 -m unittest tests.audit_eval_test`
- AC-3 (REQ-5, REQ-6): score 가 버그-서술 finding 만 세어 recall_raw·recall_verdict·대조군 과다지목률을 내고, 서술 없는 FAIL 판정은 어느 회수율에도 들어가지 않으며(허수 판정), 서술이 있어도 PASS 판정이면 recall_raw 에만 들어가요.
  verify: `python3 -m unittest tests.audit_eval_test`
- AC-4 (REQ-7): README 가 사고 추가·실행·채점·중단 규칙을 각각 한 단락씩 말해요.
  verify: `python3 -m unittest tests.audit_eval_test`

## Scope

- write: `.moai/evals/audit/**`, `tools/audit_eval.py`, `tests/audit_eval_test.py`, `.moai/specs/SPEC-20261119-audit-eval/**`, `.gitignore`(리더 결정 note 2415 (b) 로 추가)
- read-only: `agents/` (감사 에이전트 정의), `target/eval-research/`(사고 조사 스크래치, 커밋 안 함), GitHub 감사 댓글(사고 선정 시 읽음)
- forbidden: 네트워크 호출을 만드는 코드, 실제 모델 실행, `.moai/evals/` 외부의 평가 산출물, 실제 HOME·원장 쓰기

## Plan

1. 시험 먼저: `tests/audit_eval_test.py` 를 쓰고 RED 를 읽어요(도구·세트가 아직 없어 import 실패).
2. 사고 12건과 대조군을 선정해 `.moai/evals/audit/cases-tuning.json`·`cases-heldout.json`·`controls.json` 에 쓰고(정답키는 감사 댓글을 읽고 한 문장으로 손으로 씀), README 를 써요.
3. `tools/audit_eval.py`(list·run·score·report)를 만들어 GREEN 으로 가요.
4. 이웃 검사: `python3 -m unittest tests.repo_scans`(새 파일 등록 필요한지), 나머지 전체는 PR CI 로 인용.

## Risks

- 토큰 겹침 채점의 조악함: 한 줄 정답키의 고유 토큰 50% 임계값은 상수로 두고 README 에 한계를 적어요; 점수는
  비교용이지 진실이 아니에요.
- 사고 head ref 가 로컬에 없을 수 있어요: REQ-8 의 missing_ref 보고로 다루고, 선정 단계에서 cat-file 로 확인해요.
- 대조군 "크기 맞춤" 이 흐려질 수 있어요: 총 diff 줄 수 버킷(±50%)으로 정의하고 파일에 근거를 남겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 사고 12건 선정(#3216 #3215 #3212 #3211 #3204 #3203 #3199 #3198 #3197 #3102 #3093 #3084) | 유형 다양성(가드 우회·변이 생존·CI 레드·검증 명령 불가·문서 불일치·데이터 손실)과 댓글 가독성 | 사고 교체는 파일 편집으로 |
| D-2 | finding 은 정답키 고유 토큰 50% 이상 겹칠 때만 "버그 서술"로 인정 | 한 문장 키의 취지를 보존하는 가장 단순한 규칙; 임계값 상수 | 임계값 조정 |
| D-3 | tuning 6 / held-out 6 고정 분할 | 사고 12개의 균등 분할, id 중복 금지로 고정 | 재분할은 두 파일을 다시 쓰는 커밋으로 |
| D-4 | 팁-red 사고는 #3199·#3212(CI red at head)로 대표 | 팁 push 사고의 개별 목록은 PR-history 분석 후속 카드가 함 | 후속 카드에서 개별 사고 추가 |

## Result

- AC-1: met — `cases-tuning.json` 6건 + `cases-heldout.json` 6건(겹침 없음, 각 pr·audit_comment·head·base·한 줄 answer_key), `controls.json` 6건(각각 matched_case·size_bucket) — `python3 -m unittest tests.audit_eval_test` SetTest GREEN(2026-11-19, worktree moai-adk-102).
- AC-2: met — 같은 시험 RunTest GREEN: dry-run 프롬프트에 head·base 있고 answer_key 없음, 없는 ref 는 missing_ref 로 건너뛰고, tuning/held-out 중복 id·필드 누락 기록은 거부.
- AC-3: met — 같은 시험 ScoreTest GREEN: 서술 finding 만 recall_raw(2/4), 서술+FAIL 만 recall_verdict(1/4), 대조군 과다지목 1/2; TOKEN_FRACTION 0.5 상수 고정.
- AC-4: met — 같은 시험 ReadmeTest GREEN: 추가·실행·채점·중단(과적합) 네 단락.
- 이웃 검사: `tests.repo_scans_test` 25 OK, `tests.ci_scope_test` 19 OK(.gitignore 예외 줄 `!/.moai/evals/` 반영 뒤 재실행). 전체 파이썬 묶음 로컬 discover 은 30분 대기 후 정지(카드 note 2429) — 리더 통합 검증·CI 로 인용.
- 범위: 리더 결정(note 2415 (b))로 `.gitignore` 한 줄이 범위에 추가되었고 SPEC Scope 의 write 목록에 반영한다(커밋에 포함).
