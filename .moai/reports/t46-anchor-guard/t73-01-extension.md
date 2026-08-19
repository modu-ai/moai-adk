# t73 — 앵컈 세션 가드, 나머지 제거 표면 3곳 적용

카드: t73 "t46 후속 ⑵ — 앵컈 세션 가드를 나머지 제거 표면 3곳에 적용" (Tier S)
브랜치: `worktree-t46-anchor-guard` (t46 `cf749fafe` 위에 추가 커밋)
작성: 2026-08-17 (run 레인, 칸반 tjv7iy)

## 적용 표면과 동작

| 표면 | 위치 | 동작 |
|---|---|---|
| (1) `moai worktree remove <path>` | `internal/cli/worktree/remove.go` | `filepath.Abs`로 경로 정규화 후 가드. 앵컈 세션 존재 시 `ANCHORED_SESSIONS_PRESENT:` 센티넬로 **거부(exit 1)** — `done`과 동일 문구. `--force`면 경고 후 제거 |
| (2) `moai worktree clean --stale` | `internal/cli/worktree/clean.go` `cleanStaleWorktrees` | 기존 `keepReason` 모델에 합류: "live session anchored in this worktree" — 미리보기/보존 보고에 자동 표시 |
| (2b) `moai worktree clean --merged-only` | `cleanMergedWorktrees` | dirty 검사가 없는 경로라 앵컈 가드가 유일한 보호 — "Keeping …: live session(s) anchored" 출력 후 skip |
| (3) PR-merge 자동정리 | `internal/cli/session_worktree_prmerge.go` (제거 직전, dirty guard와 같은 EC-11 위치) | "PR-merge cleanup skipped (live session anchored, N): worktree … preserved" 통지 후 skip — 기존 fail-open 시맨틱 유지 |

감지 로직은 t46의 `session.LiveAnchoredSessions` 재사용 (중복 0줄). `clean`의 기본 경로(인자 없음)는 `Prune()`만 하고 디렉터리를 삭제하지 않으므로 무해 — 미적용이 아니라 대상 아님.

## 정정 사항 (리드 디스패치 전제)

디스패치는 "(3)이 칸반 배치의 L1(.claude/worktrees) 폐기 경로와 직결"이라 했으나, prmerge 스윕은 `SessionWorktreeBranchPrefix = "WT-"`(`session_worktree.go:39`)로 시작하는 브랜치만 다룬다. L1 워크트리 브랜치는 `worktree-<name>` 형태라 **prmerge가 L1을 직접 지우지는 않는다** — (3)이 보호하는 것은 L2(~/.moai/worktrees의 WT-*) 세션 워크트리다. L1 폐기의 실제 경로는 `done`/`remove`(t46·t73 (1)로 이미 보호됨)이므로 방어망은 유효하다.

## 검증 (같은 디렉터리)

- `t73-02-red.txt` — 수정 전: remove 거부 없음 + `--force` 경고 없음(※ 아래 교정 참조) + clean 두 경로 모두 앵컈 트리 제거
- `t73-02b-red.txt` — 수정 전: prmerge가 앵컈 트리 제거
- `t73-03-green.txt` — prmerge 전체(기존 18종 + 신규 1종) PASS
- `t73-03b-green.txt` — worktree 패키지 앵컈 계열 전부 PASS (t46 4종 + t73 5종 + 기존 dirty/unmerged 동반 케이스)
- `go test ./internal/cli/worktree/ -count=1` → ok 12.458s (패키지 풀)
- `internal/cli` 패키지는 전체 스위트가 무거워(과거 지연 사고) 표적 실행으로 검증, 전량 판정은 CI
- `go vet` + `GOOS=windows go vet` (worktree/cli/session 3패키지) 통과, `golangci-lint` 0 issues

## 테스트 작성 중 발견한 함정 (재발 방지 기록)

`t.TempDir()` 경로에 테스트 함수명이 그대로 박힌다. 초안의 `TestRunRemove_ForceOverridesAnchorWarning`이 `strings.Contains(out, "Warning")`를 단언했는데, **경로 속 테스트 이름 자체에 "Warning"이 포함돼 수정 전에도 공허하게 통과**했다. 수정 전 RED 확인 단계에서 "force 경고" 테스트만 이상하게 초록이라 발견. 교정: `Warning` 같은 베어 단어가 아니라 구현이 실제로 출력하는 문구(`force removing`)로 단언. 같은 이유로 `done` 쪽 t46 테스트는 stderr만 검사해 무관.

## 파일

- `internal/cli/worktree/remove.go` — 가드 + 절대경로 정규화
- `internal/cli/worktree/clean.go` — 두 스윕 경로 가드 + `--help` 문서화
- `internal/cli/session_worktree_prmerge.go` — 제거 직전 skip 가드
- 신규 테스트 3파일: `remove_anchor_test.go`, `clean_anchor_test.go` (worktree), `session_worktree_prmerge_anchor_test.go` (cli)
