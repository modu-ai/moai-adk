# t622 재고정 수리 증거 (SPEC 0.2.6)

감사: `.moai/reports/t622/plan-audit-reanchor.md` (FAIL 0.75, D1~D6). 측정 트리: 워크트리 `.claude/worktrees/t622`, HEAD `f2fa64e08bf4e2316cf95b7bf622b843b7e4ac11`, 로컬 develop `00ae57ad7ae74bc70ca1beb7d68c15fb4d96d1da` (`refs-at-measure.txt`).

## 흉내 낸 재흡수 (병합 없음, 참조 이동 없음)

| 명령 | 관측 |
|---|---|
| `git merge-tree --write-tree HEAD develop` | exit 0, 트리 `8d4d3919f573a51a6e94e177b831137938b88c6a` (`emu-merge-tree.txt`) |
| `git commit-tree <트리> -p HEAD -p develop -m …` | 매달린 커밋 `7fae4764e96b54d453f24bba4f746b4281c17f1b` (`emu-commit.txt`). 뒤이어 `git rev-parse develop HEAD` 가 위 두 값 그대로 |
| `git merge-base --all develop HEAD` | `f1f034bb4…` 1줄 (`card-base-all.txt`) |
| `git merge-base --all develop 7fae4764e…` | `00ae57ad7…` 1줄 |

## D1 범위 판정

| 판정 | 현재 트리 | 흉내 낸 흡수 | 파일 |
|---|---|---|---|
| 카드 커밋 `git log --no-merges --format=%H <tip> --not develop` | 23줄 | 23줄, `diff` exit 0 | `card-commits-*.txt`, `card-commits-current-vs-emulated.diff` |
| AC-016 `… <tip> --not develop -- <acp 두 사본>` | test -e 0, test -s 1 | test -e 0, test -s 1 | `ac016-judge-*.txt` |
| AC-016 양성 대조 `255f88eb0 --not b412f8a33 -- <acp 두 사본>` | 2줄 (97ef8e302, 6896eef37) | — | `ac016-control-notform.txt` |
| 리터럴 `$BASE..<tip>` 비병합 커밋 | 3줄 | 12줄 (늘어난 9줄 모두 develop 커밋) | `literal-base-range-*.txt`, `literal-base-range-emulated-foreign.txt` |
| 0.2.5 대응(BASE 를 흡수 병합으로) `7fae…..7fae…` | — | 0줄 | `old-remedy-moved-base-range.txt` |
| 범위 대조 `git diff --name-only develop...<tip>` | 192줄 | 192줄, `cmp` exit 0 | `card-range-names-*.txt` |
| 세 점 = 명시 merge-base (`git diff --name-only f1f034bb4 HEAD`) | `cmp` exit 0 | — | `card-range-names-current-explicit.txt` |
| AC-015 `git diff develop...<tip> -- internal/template/templates/` | test -e 0, test -s 1 | test -e 0, test -s 1 | `ac015-card-template-*.diff` |
| 리터럴 `git diff $BASE 7fae… -- internal/template/templates/` | — | 6 파일, 추가 줄 146 | `ac015-literal-base-emulated.diff` |
| AC-014 `git diff --name-only develop...<tip> -- …/.codex/agents/moai/` | 빈 파일 | 빈 파일 | `ac014-changed-*.txt` |
| AC-014 양성 대조 `ee99507fb...97ef8e302 -- …/.codex/agents/moai/` | 2줄 | — | `ac014-control-dr0911.txt` |
| AC-025 `git diff develop...<tip> -- <범위 18경로>` | test -e 0, test -s 1 | test -e 0, test -s 1 | `ac025-scope-*.diff` |
| AC-030 `git log --format=%H <tip> --not develop -- sync.md.tmpl` | 빈 목록 | 빈 목록 | `ac030-src-commits-*.txt` |
| AC-030 형태 양성 대조 `97ef8e302 --not ee99507fb -- <템플릿 acp>` | 1줄 | — | `ac030-control-dr0911-logform.txt` |
| 흡수를 거친 실제 카드 t645 `0db675bed...c7874923a` | 8 파일 (자기 기여) | — | `control-t645-card-names.txt` |
| 같은 카드를 흡수 전 분기점 `dae1b070b` 에 고정 | 19 파일 | — | `control-t645-literal-pinned-names.txt` |
| 같은 카드 커밋 `c7874923a --not 0db675bed` | 2줄 (둘 다 t645) | — | `control-t645-card-commits.txt` |
| 템플릿 `git diff --name-only f1f034bb4^1...f1f034bb4^2 -- internal/template/templates/` | 6 파일 | — | `control-dr0911-template-names.txt` |

## 스냅숏 신선도 점검

| 명령 | 관측 | 파일 |
|---|---|---|
| `git diff --name-only $BASE f1f034bb4 -- <스냅숏 21경로>` | test -e 0, test -s 1 | `snapshot-stale-current.txt` |
| `git diff --name-only $BASE 00ae57ad7 -- <스냅숏 21경로>` | test -e 0, test -s 1 | `snapshot-stale-emulated.txt` |
| 양성 대조 `git diff --name-only b412f8a33 $BASE -- <스냅숏 21경로>` | 8줄 | `snapshot-stale-control-b412.txt` |
| 미러 기준선용 `… $BASE f1f034bb4 -- internal/template/ .claude/rules/moai/` | 빈 파일 | `snapshot-stale-mirror-current.txt` |
| 같은 형태, `00ae57ad7` | 8줄 (`spec-workflow.md` 두 사본 포함) | `snapshot-stale-mirror-emulated.txt` |
| 명령 치환 형태 `git diff … "$(git merge-base develop HEAD)" …` | 워크트리 가드가 거부 — 그래서 `card-base.txt` 두 단계 형태 | — |

## D2 AC-GDP-002

판정기 `ac002/judge.awk` (acceptance.md 한 줄 판정기와 `cmp` exit 0 — `ac002/judge-program-from-acceptance.txt`). 결과 `ac002/judge-results.txt`(재실행 `judge-results-rerun.txt`, `diff` exit 0), 절 diff `ac002/section-results.txt`, 추가 시도 `ac002/pass-attempts.txt`.

| 픽스처 | (a) order | shape | (b) 절 diff | AC-GDP-002 |
|---|---|---|---|---|
| 현재 로컬 블록 / 절 | PASS | PASS | exit 0 | PASS |
| 현재 템플릿 블록 / 절 | PASS | PASS | exit 0 | PASS |
| i 옛 블록 | FAIL | FAIL | exit 1 | FAIL |
| ii `;` 결합 | PASS | FAIL | exit 1 | FAIL |
| iii 상태 검사 삭제 | FAIL | FAIL | exit 1 | FAIL |
| iv rev-list 먼저 | FAIL | FAIL | exit 1 | FAIL |
| v fetch `&` | FAIL | FAIL | exit 1 | FAIL |
| vi fetch `| tee` | FAIL | FAIL | exit 1 | FAIL |
| vii `exit_note=` | PASS | FAIL | exit 1 | FAIL |
| viii Lane B 직렬화 | PASS | PASS | exit 1 | FAIL |
| ix 절 밖 뒤집기 | PASS | PASS | exit 0 | PASS — 한계(AC-016·신선도 점검 몫, 파일 전체 diff exit 1) |

## D5

`diff reanchor-audit/post-sets.txt reanchor/mirror-baseline-sets.txt` exit 1, 둘을 `sort` 한 `d5-*.sorted` 끼리 `diff` exit 0.
