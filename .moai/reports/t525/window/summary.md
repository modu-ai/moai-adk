# t525 통합 창 기록

창 보유: lane-8. 리드 지명(lane-5 반납 뒤) → `moai integration status` free 확인 → `moai integration acquire --name lane-8` exit 0.
관측: 창 기록의 `branch`·`worktree` 칸이 `develop`·`.claude/worktrees/develop` 으로 찍혔다. 세션의 실제 위치는 `.claude/worktrees/t525`(`git rev-parse --show-toplevel`), 브랜치 `WT-speclint-red` 였다. 보유자(lane-8, 세션 id)는 맞다.

## 흡수

- 흡수 대상: 로컬 develop `eb50af5a8ec51862e3df76c2e3b08377ce01c4d8` (리드: 로컬 = origin). 흡수 전 카드 HEAD `5b13d099d`.
- 사전 예측 `git merge-tree --write-tree --name-only eb50af5a8 HEAD` → `CONFLICT (content): Merge conflict in CHANGELOG.md` 1건.
- 실행 `git merge --no-ff develop` → exit 1, 같은 1건(`absorb-merge.log`).
- CHANGELOG: `## [Unreleased]` → `### Fixed` 머리에 양쪽이 항목을 하나씩 끼워 넣은 인접 삽입(카드 쪽 SPEC-SPECLINT-GATE-SIGNAL-001, develop 쪽 SPEC-REVIEW-SECRET-SCAN-REFS-001). 표지 세 줄만 지워 두 항목을 모두 살렸다. 문구 무변경.
  - 표지 0, `git diff --check` exit 0, `^## \[Unreleased\]` 1.
  - `git diff develop -- CHANGELOG.md`(`changelog-vs-develop.diff`): 추가 1줄(카드 항목)·삭제 0.
  - `git diff HEAD -- CHANGELOG.md`(`changelog-vs-head.diff`): 추가 2줄(develop 쪽 SPEC-REVIEW-SECRET-SCAN-REFS-001, SPEC-SYNC-GATE-FAILSTATE-001)·삭제 0.
- 흡수 병합 커밋 `05ba7bb36e7db73356d6c79dfca53dcb7585a648`, 부모 `5b13d099d` · `eb50af5a8`, 트리 `ddf0c3343c30d357a6e61bf724415473586a183b`.

## 재측정 범위 (리드 지시)

- `git diff --stat 5b13d099d 05ba7bb36 -- 'internal/cli/spec_lint*'` → 출력 0바이트(`spec-lint-files-delta.txt`). `internal/spec` 도 같은 범위에서 변화 없음. 전체 변경 파일 330.
- 따라서 패키지 테스트는 슬롯 결과로 대체한다: `.moai/reports/t525/f1-slot/summary.md`(RED 8/8 FAIL, GREEN 선택자 9/9·`-run TestSpecLint` 33/33, HEAD `25f689832`; 이후 `internal/cli/spec_lint*` 변경 0).

## spec lint 재측정 (흡수 트리)

- 판정 바이너리: 흡수 트리(HEAD `05ba7bb36`)에서 `go build -o <세션 스크래치>/moai-t525-window ./cmd/moai` 1회, exit 0. 직전 `pgrep` 로 다른 go test/build/vet·cli.test 0 확인.

| 실행 | 파일 | exit | 요약 |
|---|---|---|---|
| `spec lint` | `lint-default.txt` | 0 | `0 error(s), 3133 warning(s)` |
| `spec lint --baseline .moai/spec-lint-baseline.json` | `lint-baseline.txt` | 0 | `baseline: OK` · `inventory: 3133 warning(s) total (advisory included), 0 non-advisory tracked across 0 recorded rule(s)` · `recorded at: 4ac93f755 (2026-09-11)` |
| `spec lint --json` | `lint.json` | 0 | 발견 3335 = info 202 + warning 3133, **non-advisory warning 0**, `DuplicateAcceptanceID` 0 |

판독: 비-advisory 인구 0 → 0, 변화 없음(리드의 멈춤 조건 아님). 경고 총수 3133 그대로. info 가 203 → 202 로 1건 줄었다 — `--strict`·기준선 모두 info 를 게이트하지 않는다.

## 미검증

- 흡수 트리에서 `internal/cli` 테스트는 돌리지 않았다(리드 지시로 슬롯 결과 대체, 근거는 위 델타 판독).
- info 1건 감소의 원인은 가르지 않았다.
- CI(darwin·windows 매트릭스)와 AC-SLGS-011 의 CI 로그 절반은 develop push 뒤 리드가 읽는다.
