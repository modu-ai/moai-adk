# progress.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-10
artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M) + spec-compact.md — v0.1.3
card: t619
baseline: worktree `.claude/worktrees/t619`, branch `WT-toolpolicy-drift`, plan 작성 HEAD `c7b8d110b`, 1회차 감사 HEAD `b2cbfd207`, 2회차 `dc220b8fe`, 3회차 `7c96ed2e1`, `git merge-base HEAD develop` → `d1b61005d20967fdbd970ec7ec734c6d14f29dc3`
evidence_base: `.moai/reports/t619/verdict.md`
spec_id_check: `[[ "SPEC-TOOLPOLICY-DRIFT-GUARD-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS` (실행 출력), `ls .moai/specs | grep -c TOOLPOLICY-DRIFT-GUARD` → `0` (작성 전)
open_clarifications: 0 — 운영자 결정 6건(2026-09-10) 반영. 수리 방향·검사 위치·주장 정정(레인 세션 직접 수령), 집합 비교·주장 정정 전수·워킹 트리 판독(plan 제안 검토 후)
plan_measurements: 스크래치 build `allow=108 ask=0 deny=60 env_gated_skipped=5`, `diff` 종료 1 / 4 헝크 / 160 vs 166 줄, 커밋본 목록 `allow-NOT-C-sorted` `deny-NOT-C-sorted`, 커밋본 중복 0 / allow·deny 겹침 0
plan_audit: 1회차 FAIL 0.71 → 2회차 FAIL 0.79(Tier M 상한 도달, 운영자가 3회차 1회 승인) → 3회차 FAIL 0.89(blocking D21 1건) → 운영자 결정 "지금 고치고 진행", v0.1.3 수정
gap: **v0.1.3 의 D21-D23 수정은 독립 재감사를 받지 않았다.** 3회차가 최종 감사였고, 운영자 결정에 따라 수정 줄은 오케스트레이터가 직접 확인한다. 이 수정분에 대한 plan-auditor 판정은 존재하지 않는다.

### plan-audit 1회차 (2026-09-10) — FAIL 0.71 대응

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-1.md`
- blocking D1·D2·D3 반영, optional D4-D10 은 반영 또는 부분 반영(D8). 결함별 처분 표는 plan.md §H.
- 수리 회차 실측(HEAD `b2cbfd207`): env_gate 없음 + 네 경로 deny 질의 `12`, env-gated Write deny `1`, env_gate 항목 `5`, `ANALOGOUS drift class` `1`, `comment (audit surface) are generated` `1`, `tool-policy-drift-check` 두 파일 모두 `0`, `.claude/settings.json#permissions` `171`, `make -n build` 의 `TestToolPolicyDrift_` `0` / `TestGoldenCommittedArtifactsMatchEmission` `2`, `go_code:` 블록 14줄·`.moai/**` `1`·`.claude/settings.json` `0`, merge-base 기준 settings 두 파일 diff 출력 없음.
- D8 관련 가드 실측: `trap` 포함 명령, heredoc 으로 python/bash 에 스크립트를 넘기는 명령, 변수로 계산된 스크립트 경로를 python 에 넘기는 명령이 모두 워크트리 세션 가드에 거부됨. 리터럴 절대 경로의 스크래치 스크립트 호출은 통과.

### plan-audit 2회차 (2026-09-10) — FAIL 0.79 대응, 운영자 승인 3회차

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-2.md`
- Tier M 상한(2회) 도달. 운영자가 수정 후 3회차 감사 1회를 승인했다.
- blocking D11·D13·D14, optional D12·D15-D20 모두 반영. 처분 표는 plan.md §H 2회차.
- 수리 회차 실측(대상 파일 무수정 상태): env_gate 항목 정렬 JSON sha256 `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27`, `^build:.*tool-policy-drift-check` `0` / `^build:.*agents-emit-check` `1`, 기존 네 파일 역슬래시-u / Cf 모두 `(0, 0)`, 조건부 복원 사슬 일치 시 `restored=yes`·불일치 시 `restored=NO_stop`, `git status --porcelain --untracked-files=all > <파일>` 가드 통과.

### plan-audit 3회차 (2026-09-10) — FAIL 0.89, blocking D21 1건, 운영자 결정 "지금 고치고 진행"

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-3.md`
- D21(blocking)·D22·D23(optional) 모두 반영. 처분 표는 plan.md §H 3회차.
- 수리 회차 실측(대상 파일 무수정 상태, 스크래치):
  - 고정 fixture 네 쌍을 `moai tool-policy build --local-only` 로 재생성: 생성 고유 집합이 fixture 고유 집합과 모두 같음. `duplicate_settings_allow`·`_deny` `{"allow":["Read"],"ask":[],"deny":["Bash(rm -rf /:*)"]}`, `_ask` ask `["WebFetch"]`, `allow_deny_overlap` deny `["Bash(rm -rf /:*)","Read"]`. 중복 fixture 길이·고유 길이 `[2,1]`, 겹침 fixture allow∩deny `["Read"]`. 세 YAML fixture 로딩 성공.
  - `malformed_settings_json` 바이트 → `permissions object parse: invalid character '}' looking for beginning of object key string`, 종료 `1`.
  - `wrong_type_settings_list` 바이트(`"ask": 1`) → 생성기 build 종료 `0`, ask 사라짐(`settings_region.go:173` 오류 폐기).
  - `malformed_yaml` 모양 → `tool-policy parse … yaml: line 5: mapping values are not allowed in this context`, 종료 `1`.
  - 조건부 변이 사슬(백업에서 계산한 원본 sha, 리터럴 경로): 원본 상태에서 `mutate=DONE`·항목 173→172, 한 줄 덧붙인 불일치 상태에서 `mutate=STOPPED`·항목 173 유지. 가드 통과.
  - 가드 추가 실측: 셸 반복문 안에서 변수 인자로 `moai` 를 부르는 명령, 여러 heredoc 을 묶은 명령은 거부됨. 파일마다 `printf '%s\n' … > <리터럴 경로>` 로 나누면 통과.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
