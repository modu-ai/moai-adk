# t228 — scan-corpus 존재 여부 재측정 (v2, 정정본)

측정 트리: 워크트리 `.claude/worktrees/t228`, 브랜치 `WT-astgrep-16-langs`,
HEAD `294b4b6ab` (= `origin/main`, fetch 후 `0 0`).
**측정 위치: 워크트리 루트** — 이 문서의 v1 이 틀린 원인이 위치였습니다(아래 §3).

## 1. 결론

`internal/hook/security/testdata/scan-corpus/` 는 **`origin/main` 에 커밋된 상태로 존재합니다** —
파일 12개. 도입 커밋은 `a9eb896ce fix(security): restore the pre-write ast-grep deny capability
(t227) (#1637)` 로, t228 이 severity 원칙을 승계하는 그 카드의 산출물입니다.

`WT-security-scan-surface`(t217) 브랜치도 같은 12개를 가지며(상속),
그 브랜치가 해당 경로에 가하는 변경은 공집합입니다.

따라서: **t217 머지 의존 없음. 경로 충돌 위험 없음.** 코퍼스 확장은 이 카드 안에서 진행합니다.

## 2. 측정 (모두 cwd = 워크트리 루트)

| 명령 | 출력 |
|---|---|
| `git rev-list --count --left-right origin/main...HEAD` | `0	0` |
| `git ls-tree -r --name-only --full-tree origin/main \| grep -c scan-corpus` | `12` |
| `git ls-tree -r --name-only origin/main \| grep -c scan-corpus` | `12` |
| `git ls-tree -r --name-only --full-tree WT-security-scan-surface \| grep -c scan-corpus` | `12` |
| `git log --oneline origin/main -- internal/hook/security/testdata/scan-corpus` | `a9eb896ce fix(security): restore the pre-write ast-grep deny capability (t227) (#1637)` |
| `git diff --stat origin/main...WT-security-scan-surface -- internal/hook/security/testdata` | 출력 없음 |

파일 12개: `go_clean.go` · `go_deny_credential.go` · `go_deny_md5.go` · `go_warning_only.go` ·
`java_uncovered.java` · `js_clean.js` · `js_deny_exec.js` · `py_clean.py` ·
`py_deny_os_system.py` · `rs_uncovered.rs` · `ts_clean.ts` · `ts_deny_exec.ts`

## 3. 이 문서 v1 이 틀린 이유 — cwd 접두사 함정

v1 은 같은 ref 에 대해 0 을 보고했습니다. 원인은 ref 도 fetch 상태도 아니라 **측정 위치**입니다.

`git ls-tree -r --name-only <ref>` 는 인자로 준 ref 전체가 아니라 **현재 디렉터리 접두사 아래만**
나열합니다. v1 측정 당시 셸 cwd 는
`internal/template/templates/.moai/config/astgrep-rules/security` 로 드리프트해 있었고,
그 아래에는 `internal/hook/...` 이 없으므로 출력이 비었습니다.
`grep -c` 는 0 을 찍고 rc=1 을 냈으며, 그 0 이 부재의 근거로 잘못 채택됐습니다.

같은 계열로 `git log --oneline <ref> -- '*scan-corpus*'` 도 빈 출력이었습니다 —
glob pathspec 이 cwd 기준으로 해석돼 매치가 사라진 것입니다. 리포 상대경로
(`-- internal/hook/security/testdata/scan-corpus`)로 주면 정상적으로 커밋이 나옵니다.

이 함정은 **부재 주장에만 나타나고 침묵으로 실패합니다** — 오류도 경고도 없이 0 을 냅니다.

재발 방지: ref 조회는 리포 루트에서 돌리거나 `--full-tree` 를 붙이고, pathspec 은 glob 대신
리포 상대경로로 준다. 부재를 주장하기 전에 `pwd` 를 확인한다.

## 4. 함께 정정되는 앞선 주장들

| 주장 | 판정 |
|---|---|
| 리드 1차: "primary 에 없다" | 스테일 트리(`a1b1ca696`, origin/main 대비 259커밋 뒤) 관측 — 무효 |
| 이 레인 M5: "어느 ref 에도 없다" | cwd 접두사 함정 — 무효 |
| 리드 2차: "origin/main 에 12개 있다" | **확인됨** |
| 이 레인 v1: "t217 의 미커밋 워킹트리 파일" | 무효 — 커밋된 main 파일 |
| "t217 머지 의존 없음" | 유지 (근거는 교체: 부재가 아니라 이미 존재) |
| "경로 충돌 위험" | **철회** |
| 인용 기준 SHA `a1b1ca696` → `294b4b6ab` | 유지 (별개의 정정, 여전히 유효) |
