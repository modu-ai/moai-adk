---
id: SPEC-20261119-audit-eval-followups
title: "Audit-eval follow-ups: evals ignore secrets, the ref check pins the environment and names the missing ref"
status: draft
revision: 1
created: 2026-11-19
---

# Audit-eval follow-ups: evals ignore secrets, the ref check pins the environment and names the missing ref

## Intent

묶음 PR #3224 감사(AUDIT #3224 N1-N4)의 비차단 후속 중 audit-eval 쪽 조각 1(리더 배차 2026-11-19): 평가 레코드 폴더 `.moai/evals/` 가 비밀 모양 무시 규칙의 빠진 곳이고, `tools/audit_eval.py` 의 ref 확인이 부분 복제에서 lazy fetch 를 일으킬 수 있으며, 빠진 ref 를 사례 id 로만 알려주고(`head` 인지 `base` 인지 모름), 경로 구분자가 든 사례 id 가 프롬프트 파일 이름으로 흘러 출력 폴더 밖을 노릴 수 있어요. hooks 쪽(AC-3·AC-4)은 이 SPEC 의 범위 밖 — 같은 카드의 나중 조각이에요.

## Non-goals

- hooks 쪽 후속(AC-3·AC-4: 빈 hooks 키·symlink 레코드·project_update 주석) — 같은 카드의 두 번째 조각.
- ruleset·저장소 보호 설정 변경(리더 지시: 건드리지 않음).
- 평가 채점 로직(`describes`·`distinctive_tokens`) 변경.

## Requirements

- REQ-1: While a file under `.moai/evals/` carries a secret-shaped name (`*.key`, `id_rsa*`, `secrets.*`), the repository shall ignore it, so eval records never commit key material.
- REQ-2: When `tools/audit_eval.py` runs `git cat-file` to check a ref, it shall set `GIT_NO_LAZY_FETCH=1` in the child environment, so a partial clone reports the ref missing instead of fetching. · failure: none — 환경 변수 설정 자체에 거부 경로가 없고, 실패는 ref 누락 보고로 흘러요.
- REQ-3: When a case's head or base ref is missing, the tool's `missing_ref` output shall name which ref of which case is missing, not the case alone, and leave the case out of the score's cases, controls and findings rows as the module contract says.
- REQ-4 (REQ-3): IF a case id holds a path separator (`/` or `\`) or a colon (`:`) or is `.` or `..`, THEN the tool shall refuse the set before writing anything, so a case id never aims a prompt file outside the output folder and cannot forge another case's `id:which` entry.

## Acceptance

- AC-1 (REQ-1): `git check-ignore -v` 가 `.moai/evals/` 아래 `x.key`·`id_rsa_x`·`secrets.json` 을 무시 규칙과 함께 가리켜요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.audit_eval_test -k ignore`
- AC-2 (REQ-2, REQ-3, REQ-4): stub `git` 이 환경을 기록해 `GIT_NO_LAZY_FETCH=1` 을 못 박고, 빠진 head·base 가 `id:head`·`id:base` 로 이름 지어지며, `a/b`·`..` 사례 id 는 아무것도 쓰기 전에 거부돼요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.audit_eval_test`

## Scope

- write: `.gitignore`, `tools/audit_eval.py`, `tests/audit_eval_test.py`, `.moai/specs/SPEC-20261119-audit-eval-followups/spec.md`
- read-only: `tools/source_manifest.py`(도구 변경 매니페스트 대조 — tools/ 는 러스트 매니페스트 밖), `docs/design-review/appendix/operations/`(참조만)
- forbidden: 실제 HOME 과 원장, 네트워크(테스트는 stub git), 실제 원격 저장소, 키, ruleset·보호 설정

## Plan

1. RED 먼저: `tests/audit_eval_test.py` 에 (a) 임시 저장소에서 check-ignore 로 세 모양 커버, (b) stub `git` 이 받은 환경의 `GIT_NO_LAZY_FETCH` 기록, (c) missing_ref 가 `id:head`·`id:base` 이름, (d) 경로 구분자 사례 id 거부·무쓰기. 먼저 실패해요.
2. `.gitignore`: `.moai/evals/**` 앵커 규칙 세 줄.
3. `tools/audit_eval.py`: `ref_missing` 에 env, missing 수집을 `(id, which)` 로, `validate` 의 사례 id 검사.
4. 전체 `python3 -m unittest tests.audit_eval_test` + 소스 매니페스트(tools/ 변경 시 관행) → push → tier 1 자기 감사 → LANE-READY-INTEGRATION.

## Risks

- `missing_ref` 출력 형식 바뀜(`id` → `id:which`) — 소비자가 문자열을 그대로 화면에 보여주는 용도라 호환 위험 낮음; SPEC 에 기록.
- `.moai/evals/` 가 저장소에 아직 없어도 규칙은 무해(미래 파일 대비).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 무시 규칙은 `.moai/evals/**` 앵커 세 줄(`*.key`·`id_rsa*`·`secrets.*`) | 카드 AC-1 의 세 모양; check-ignore 로 검증 가능 | 전역 `*.key` 추가(거짓양성 위험 — #237에서 뺐음) |
| D-2 | missing_ref 형식 `id:which`(head/base) | 어느 ref 인지가 디버깅의 핵심; 정렬 가능 | 사전 형태 `{id: [which]}` |
| D-3 | 사례 id 검사는 `validate`(입구)에서 — `/`·`\`·`:`·`.`·`..` 거부 | 쓰기 전 한 곳에서 다 막음; `:` 는 `id:which` 형식과 충돌(감사 B-1) | 파일 이름 sanitize |
| D-4 | 채점의 ref 누락 제외는 missing id 집합 유도로 복원(감사 B-1 수리 — 형식 변경이 깬 계약의 복원이라 논리 변경 아님) | 모듈 독스트링·README 의 "reported and skipped" 계약 | 제외 제거(계약 위반) |

## Result

- AC-1: met — 2026-11-19 이 worktree(팁 061177768c)에서 `git check-ignore -v` 가 세 모양을 새 규칙으로 가림: `.gitignore:367:/.moai/evals/**/*.key → .moai/evals/x.key`, `:368 id_rsa* → id_rsa_fixture`, `:369 secrets.* → nested/secrets.json`; 시험 `test_the_ignore_rules_cover_the_evals_folder` 통과.
- AC-2: met — `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.audit_eval_test` 13 tests OK: stub `git` 이 기록한 환경에 `GIT_NO_LAZY_FETCH=1`, `missing_ref` 가 `["case-base:base", "case-head:head"]` 로 어느 ref 인지 이름, `../escape`·`a/b`·`a\\b`·`..` 사례 id 는 아무것도 쓰기 전에 "path separator" 거부(기존 형식 고정은 D-2 대로 갱신). 변이 2 KILLED — M1(env 제거) exit 1·M2(경로 검사 무효화) exit 1, 복원 `diff -q` 클린. RED 먼저: FollowupsTest 4 중 3 실패+1 오류.
- 이웃: `.moai/evals/audit/README.md` 의 missing_ref 표현은 형식 무관으로 그대로 유효(범위 밖 미변경).
