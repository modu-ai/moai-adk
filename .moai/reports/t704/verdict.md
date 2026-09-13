# t704 verdict — template todo-queue-storage doc: temporary-origin carve-out

card: t704 · branch: `WT-queue-doc-truth` · base: local develop `7a7a08f20` · push: 금지(레인 push 없음)

## Claim

`internal/template/templates/.moai/docs/todo-queue-storage.md` 4행·106행의 무조건적 홈 경로 주장("한 개 SQLite DB는 `~/.moai/db/<project-key>/todo/backlog.db`")을 SPEC-DOCS-TODO-TEMP-GUARD-001(t575, develop 착지분) §2 진리표 기준으로 수정했다: 임시 디렉터리 기점(temp-origin) 비-git 프로젝트(절대 `MOAI_HOME` 오버라이드 없음)는 큐가 프로젝트 로컬 `<base>/.moai/state/todo/backlog.db`에 머문다. 로컬 `.moai/docs/` 사본도 동일하게 정합시켰다.

## Evidence

- `diff .moai/docs/todo-queue-storage.md internal/template/templates/.moai/docs/todo-queue-storage.md` → SYNCED (바이트 동일)
- `git diff --stat`:
  ```
   .moai/docs/todo-queue-storage.md                       | 18 ++++++++++++------
   .../templates/.moai/docs/todo-queue-storage.md         | 18 ++++++++++++------
   2 files changed, 24 insertions(+), 12 deletions(-)
  ```
- `make build` → exit 0 (agents-emit-check 선행, catalog.yaml 재생성, `go build -o bin/moai` 성공; Commit=7a7a08f20)
- 중립성 grep `grep -nE 'SPEC-[A-Z]+|t[0-9]{3}|20[0-9]{2}-...|[0-9a-f]{40}'` → 유일 적중 84행 `<SPEC-ID>` 플레이스홀더(사전 존재, 이번 diff 밖). 추가 문장에 SPEC-ID·카드 id·날짜·SHA 없음.
- `git status --short` → 위 2파일만 수정 (카탈로그 yaml 불변 확인)

## Baseline-attribution

측정 트리: `.claude/worktrees/t704` @ `7a7a08f20` (`WT-queue-doc-truth`, develop 팁 기점 — `git rev-parse --short HEAD` = `7a7a08f20`, 이번 실행, 이 트리). 진리표 근거: SPEC-DOCS-TODO-TEMP-GUARD-001 §2 (develop 워크트리 `.moai/specs/`에서 판독, `git -C .claude/worktrees/develop` 기점 `7a7a08f20`).

## Gaps

- CI 판정 미측정 — Class A의 "merge 될 head에서 CI green" 조건은 push 금지로 이 레인에서 관측 불가. 리드의 develop 일괄 push 이후 CI가 통합 판정을 만든다.
- make build 후 생성된 `bin/moai` 실행 스모크 미수행 (문서-only 변경이라 바이너리 동작과 무관하나 명시하지 않는다).
- 문서 111~114행(읽기전용 표면의 홈 DB 읽기 서술)은 배차 범위 밖 — temp-origin에서 홈 DB를 아예 resolve하지 않는다는 점에서 이 절에도 같은 carve-out이 필요한지 미판정, 후보로 남긴다.

## Residual-risk

- 진리표의 5행 중 "home unresolvable, non-temp base"(`state_dir.go:32-34`) 케이스는 이번 문장에 명시하지 않았다 — 배차가 4·106행 문장 수정으로 한정했기 때문. 필요 시 후속 카드.
- `docs-site` 쪽은 t575가 이미 수리했으므로 이 카드와 무관하지만, 두 표면의 문구가 정확히 일치하지 않을 수 있다(각자의 문체로 기술).
