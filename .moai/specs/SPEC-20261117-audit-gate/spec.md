---
id: SPEC-20261117-audit-gate
title: CI audit-gate job — 판정 댓글로 감사 check 보고 (CI 재설계 1단계, D2)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

사용자 결정(2026-11-17): `moai/audit`는 사람이 손으로 올리는 status가 아니라, PR의 판정 댓글을 읽는 CI
`audit-gate` job이 보고해요. 이 SPEC은 CI 재설계 rollout 1단계예요: 판정 규칙을 하나의 `moai audit-gate`
명령으로 만들고(병합 관문의 읽는 함수를 그대로 씀), 그 명령을 돌리는 workflow job을 더해요. ruleset의 필수
check를 이 job으로 옮기는 일은 guarded 설정이라 이 SPEC이 하지 않고, 사용자가 고를 before/after JSON만 적어요.

## Non-goals

- ruleset 26791615 바꾸기(guarded; Decisions D3에 JSON만).
- PASS가 CI 실행을 인용하는지 보기(AGENTS.md workflow 7의 "cites that run"): 다음 조각. 출력에 `cites_run_checked: false`.
- merge_group trigger, merge queue(rollout 2·3단계).
- 판정 base의 조상 관계, mechanical delta(둘 다 git이 필요; 리더의 `gate-check`가 계속 봐요).
- `Moai-Session:` trailer를 커밋에 쓰게 하는 일(SDD-6).
- `forge-audit-status`·`gate-merge`가 `moai/audit`를 올리는 코드 지우기(ruleset을 옮긴 뒤).

## Requirements

- REQ-1: When the gate reads a pull request, it shall count only verdict comments whose writer is a repository member. · failure: none — 구성원이 아닌 댓글은 세지 않을 뿐이고, 못 읽는 경우는 REQ-9예요.
- REQ-2: While a posted FAIL verdict has no later PASS by the same auditor lane, the gate shall report `fail` at every audit tier.
- REQ-3 (REQ-2): IF a comment's first line starts with `AUDIT #` but does not read as a verdict, THEN the gate shall count it as a FAIL.
- REQ-4: Where the effective audit tier is 0, the gate shall report `pass` when no FAIL is open.
- REQ-5: Where the effective audit tier is 1 or 2, the gate shall report `pass` only for a PASS on exactly the checked head, with a 40-hex base, by an auditor lane that matches no `Moai-Session:` trailer value of the pull request's commits.
- REQ-6 (REQ-5): IF no commit of the pull request carries a `Moai-Session:` trailer or its commits cannot all be read, THEN the gate shall report `missing` with the rule `authors`.
- REQ-7: The gate shall take the effective tier from the body's `audit-tier:` line raised by the merge gate's path rule, with tier 2 when the files cannot be read. · failure: none — 못 읽은 파일 목록은 tier 2로 막는 쪽이에요.
- REQ-8: When the `audit-gate` workflow runs on a pull request, the job shall build the gate from the pull request's base commit and fail unless the gate reports `pass` for the pull request's head. · failure: none — 실패 경로는 REQ-9예요.
- REQ-9 (REQ-8): IF the gate cannot read the pull request or its comments from GitHub, THEN the job shall fail.
- REQ-10: The workflow shall use the `pull_request` event only, a read-only token, and actions pinned by commit SHA.

## Acceptance

- AC-1 (REQ-1…REQ-7): 기록된 입력으로 tier 0·1·2, 열린 FAIL과 같은 감사자의 PASS, 저자 PASS, 옛 head, trailer 없음, 구성원 아닌 댓글, 읽히지 않는 판정, 짧은 base, tier 없는 본문을 판정해요.
  verify: `cargo test --locked --offline --lib audit_gate::`; `cargo test --locked --offline --test audit_gate_cli`
- AC-2 (REQ-8, REQ-9): `--require-pass`는 `pass`가 아니면 실패 종료(exit 1)하고, 기록된 GitHub 응답에서 입력을 모아요.
  verify: `cargo test --locked --offline --test audit_gate_cli require_pass_fails_the_run_unless_it_passes`; `cargo test --locked --offline --lib audit_gate::tests::collect_reads_the_pr_its_files_commits_and_comments`
- AC-3 (REQ-8, REQ-10): workflow가 actionlint를 통과하고 `pull_request_target`·쓰기 권한이 없고 action이 SHA로 고정돼요.
  manual: `actionlint .github/workflows/audit-gate.yml` 실행과 파일 읽기
- AC-4 (REQ-1…REQ-10): 병합 관문 시험이 그대로 통과하고(읽는 함수를 공유), 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --test gate_check_cli`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`;
  `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests/source_manifest_test.py`

## Scope

- write: `src/audit_gate.rs`, `src/ops/audit_gate.rs`, `src/lib.rs`, `src/gate.rs`, `src/gate_collect.rs`, `src/help.rs`,
  `tests/audit_gate_cli.rs`, `tests/fixtures/help/*.txt`, `.github/workflows/audit-gate.yml`,
  `docs/design-review/appendix/operations/audit_gate.md`, `CONTRIBUTING.md`, `skills/moai/workflows/sync.md`,
  `skills/moai/references/source-manifest.json`, `.moai/specs/SPEC-20261117-audit-gate/spec.md`
- read-only: `.github/workflows/ci.yml`, `AGENTS.md`, ruleset 26791615(GET)
- forbidden: ruleset 바꾸기, `pull_request_target`, 쓰기 토큰, `install/**`, `src/headver.rs`, `src/update*.rs`, `src/language.rs`,
  `src/hook_engine.rs`, `src/agents.rs`, `src/card*`, `src/optimize.rs`, `src/lane_env.rs`, `src/heavy_slot.rs`

## Plan

1. 판정 규칙을 `src/audit_gate.rs`에(관문의 `verdict_line`·`parse_verdict`·`auditor`·`parse_body_tier`·`tier_of` 재사용), 기록된 입력 시험.
2. `moai audit-gate` 명령(stdin 입력 또는 `--pr` GET), `--require-pass`.
3. `.github/workflows/audit-gate.yml`: base 커밋에서 빌드, PR head 판정.
4. 손으로 올리는 status를 말하는 문서에 job 한 문장, ruleset before/after JSON.

## Risks

- 커밋에 `Moai-Session:` trailer가 아직 없어요(SDD-6 전) → tier 1·2 PR은 job이 `missing`이에요. 그래서 이 조각은 check 이름을
  `audit-gate`로 두고 필수로 만들지 않아요(D2); ruleset을 옮기는 건 trailer가 쓰인 뒤예요.
- 감사자는 판정 본문의 lane 이름(`agent-N`, `gate::auditor`), 저자는 `Moai-Session:` 값이에요. 둘이 같은 이름 체계여야 비교가
  뜻을 가져요 → SDD-6이 session 값의 형식을 정할 때 맞춰요(대소문자 무시 비교).
- 판정 댓글을 올린 뒤 job이 저절로 다시 돌지 않아요: `issue_comment` workflow는 기본 가지(`main`)의 파일로 돌아요 → 판정을 올린
  쪽이 `gh run rerun <run-id>`로 다시 돌려요(sync.md 8단계에 적음). merge_group(2단계)이 오면 대기열에서 다시 읽어요.
- `pull_request` workflow 파일은 PR이 바꿀 수 있어요 → `.github/`는 관문 경로 규칙상 tier 2(감사 필수)예요. 판정 코드는 base 커밋에서
  빌드하므로 PR이 자기 규칙을 바꿔 통과할 수 없어요.
- 이 job이 처음 들어오는 PR(이 SPEC의 PR)에서는 base에 `audit-gate` 명령이 없어 job이 실패해요(필수가 아니라 막지 않음).
- job이 base 커밋 전체를 빌드해요(캐시 복원만) → 몇 분이 걸려요. 속도는 rollout 5단계.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 판정 규칙은 Rust `moai audit-gate`(병합 관문의 읽는 함수 재사용, 새 해석기 없음); Python 스크립트 아님 | 조정자 지시 "reuse, do not duplicate" | Python으로 옮기기 |
| D2 | check 이름은 `audit-gate`(job 이름). 지금 ruleset은 `moai/audit`(app 고정 없음)를 요구하므로 이 job은 아직 필수가 아니고, 손으로 올리는 status가 계속 병합을 정해요. 이름을 `moai/audit`로 하면 trailer 없는 PR 모두를 바로 막고, 손 status와 이름이 겹쳐요 | 카드 D2 "required check is the job, pinned to the Actions app", Risks 1 | job `name:`을 `moai/audit`로 |
| D3 | ruleset 변경은 사용자 선택(guarded): 아래 before/after JSON | 사용자 결정(guarded 설정) | before JSON으로 PUT |
| D4 | 감사 판정 코드는 PR head가 아니라 base 커밋에서 빌드 | 자기 규칙을 바꿔 통과하는 길 막기 | `ref:` 빼기 |
| D5 | 올린 FAIL은 head와 상관없이 모든 tier를 막고, 같은 감사자 lane의 더 뒤 PASS만 풀어요(lane이 없는 FAIL은 풀리지 않음); 읽히지 않는 `AUDIT #` 줄도 FAIL | 조정자 규칙, G3 | head별로만 |
| D6 | "PASS가 실행을 인용"은 이 조각에서 보지 않음(`cites_run_checked: false`) | 조정자 지시(다음 조각) | 다음 조각에서 규칙 추가 |
| D7 | 다시 돌리기는 `gh run rerun`(issue_comment 없음): 기본 가지가 `main`이라 issue_comment workflow가 이 파일로 돌지 않아요 | GitHub 동작(issue_comment는 기본 가지의 workflow) | 기본 가지를 바꾼 뒤 issue_comment 추가 |

D3 ruleset JSON — 2026-11-17T09:23:27Z에 `gh api repos/modu-ai/moai-adk/rulesets/26791615`(GET)로 읽은 값이에요.

before(`required_status_checks` 규칙):

```json
{"type": "required_status_checks", "parameters": {"strict_required_status_checks_policy": false, "do_not_enforce_on_create": false, "required_status_checks": [{"context": "moai/audit"}, {"context": "ci", "integration_id": 15368}]}}
```

after(선택지 A, 권장: job을 GitHub Actions app(15368)에 고정해서 필수로, 손 status는 더 세지 않음):

```json
{"type": "required_status_checks", "parameters": {"strict_required_status_checks_policy": false, "do_not_enforce_on_create": false, "required_status_checks": [{"context": "audit-gate", "integration_id": 15368}, {"context": "ci", "integration_id": 15368}]}}
```

적용은 그 규칙만 바꾼 ruleset 전체(`name`, `target`, `enforcement`, `conditions`, `rules`의 나머지 셋 `pull_request`·`non_fast_forward`·`deletion`, `bypass_actors: []`)를
`gh api -X PUT repos/modu-ai/moai-adk/rulesets/26791615 --input after.json`으로 보내고 다시 읽어 확인해요. 되돌리기는 before로 같은 PUT.
선택지 B: job 이름을 `moai/audit`로 바꾸고 `{"context": "moai/audit", "integration_id": 15368}`로 고정(이름 유지, 손 status는 app이 달라 세지 않음).
선택지 C: 지금처럼 두기(job은 참고용). 어느 쪽이든 먼저 커밋이 `Moai-Session:` trailer를 써야(SDD-6) tier 1·2 PR이 막히지 않아요.

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 기준 `a0fe2ddd09`(C-89 PR head), 커밋 `d3b08738f0`, 이 worktree의
`target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는 한 번에 하나.

RED는 기록하지 않았어요: 새 명령이라 시험과 코드를 함께 썼어요. 대신 아래 mutation 3개가 시험이 규칙을 잡는다는 것을 보여요.

- AC-1: met — `--lib audit_gate:: gate:: gate_collect:: comment_authority::` 41 passed, `audit_gate_cli` 3
- AC-2: met — 같은 실행(`require_pass_fails_the_run_unless_it_passes`, `collect_reads_the_pr_its_files_commits_and_comments`)
- AC-3: met — `actionlint .github/workflows/audit-gate.yml .github/workflows/ci.yml` exit 0; 파일에 `pull_request_target`·`write` 권한 없음(grep 결과는 "없다"고 적은 주석 한 줄뿐), action 둘 다 40자 SHA
- AC-4: met — `gate_check_cli` 40, `gate_stacked_2859` 9, `audit_admission` 7, `--bin moai` 27, `--bin moai-cli operations_doc_matches_code fragment_mutations_are_detected` 2, `help_layout` 10(골든 4개에 `audit-gate` 한 줄씩만 더함), `product_cli` 75, `home_isolation_scan` 16, `spec_cli` 17, 저장소 scan 대상 모두 exit 0; fmt 0, clippy 0, Python 85 OK, `source_manifest.py check` ok, `moai spec check --all` ok

mutation 3개 모두 KILLED: 구성원 아닌 댓글도 세기, 다른 감사자의 PASS로 FAIL 풀기, 저자의 PASS 세기.

실행 안 함: GitHub에서 실제 job 실행(push·PR 없음), 전체 `cargo test`, Linux·Windows. 남은 일: ruleset 선택(D3), SDD-6 trailer,
"실행 인용" 규칙, merge_group trigger(rollout 2단계).
