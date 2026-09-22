# t836 verdict — template todo-queue-storage doc: read-only-surface paragraph carve-out

card: t836 · branch: `WT-queue-doc-carveout` · base: local develop `d4c45bdcc` (= origin/develop tip) · push: 금지(레인 push 없음)

## Claim

t704가 남긴 111~114행 후보(읽기전용 표면의 홈 DB 읽기 서술)에 대해 **carve-out이 필요하다**는 판정을 내리고 적용했다. 후보 문단은 t704 수정으로 +6행 밀려 현재 117~120행이다. 문단의 세 문장 — "홈 DB가 있으면 읽는다", "없으면 레거시 프로젝트 로컬 큐를 읽는다", "첫 채택 `moai todo` 명령이 검증 복사를 수행한다" — 은 temp-origin 클래스에 대해 거짓이다: resolver가 temp 분기로 프로젝트 로컬을 돌려주므로 홈 DB는 경로 위에 없고, 프로젝트 로컬 큐는 레거시가 아니라 그 클래스의 정위치이며, 홈으로의 채택은 일어나지 않는다. 템플릿+로컬 `.moai/docs/` 사본 쌍에 같은 문장을 추가해 바이트 동일을 유지했다. 결함류는 t704가 수리한 것과 동일(무조건적 홈 저장소 서술 + temp-origin carve-out 누락, SPEC REQ-003 원칙) — 진단 확정이므로 B로의 강등 불요.

## Evidence

- SPEC-DOCS-TODO-TEMP-GUARD-001 §2 진리표 5행 판독 (`.moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/spec.md:50-58`): temp-origin(비-git + `os.TempDir()`/`/tmp`/`/var/folders` 기점 + 절대 `MOAI_HOME` 없음) → `<base>/.moai/state/todo/backlog.db` 프로젝트 로컬.
- `internal/kanban/state_dir.go` 판독: `StateDirForRoot` temp 분기(:29-31)는 홈 해석 자체를 하지 않음. `resolveStateDir` 주석(:102-108)이 adopt=false PURE 경로를 "read-only surfaces' path (the console, the statusline)"로 명시. 채택 경로의 재배치는 legacy→프로젝트 로컬 rename뿐(:145-149) — 홈 복사 없음.
- 읽기 소비자 전원이 `BacklogPathForRoot`(PURE, :238-241) 경유 확인 — `git grep BacklogPathForRoot -- 'internal/**'`(비테스트) → `internal/cli/todo.go:114`, `internal/kanban/backlog_store.go:500,540`, `internal/kanban/factory_runtime.go:34,68`.
- `git diff --stat` (이번 실행, 이 트리):
  ```
   .moai/docs/todo-queue-storage.md                             | 4 +++-
   internal/template/templates/.moai/docs/todo-queue-storage.md | 4 +++-
   2 files changed, 6 insertions(+), 2 deletions(-)
  ```
  양쪽 blob 동일(`613a31bc4..e537f8925`) — 쌍 바이트 동일.
- `make build` → exit 0 (agents-emit-check/commands-emit-check 선행 통과, `go build -o bin/moai` 성공, ldflags `Commit=d4c45bdcc` 스탬프 관측).
- 중립성 grep `grep -nE 'SPEC-[A-Z]+|t[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|[0-9a-f]{40}'` → 유일 적중 84행 `<SPEC-ID>` 플레이스홀더(`moai todo next <n> --spec <SPEC-ID>` — 사전 존재, 이번 diff 밖; t704와 동일 관측). 추가 문장에 SPEC-ID·카드 id·날짜·SHA 없음.
- `git status --short` → 위 2파일만 수정 (catalog.yaml 불변 — `make build`의 "updated"는 바이트 동일 재생성).

## Baseline-attribution

측정 트리: `.claude/worktrees/t836` @ `d4c45bdcc` (`WT-queue-doc-carveout`, develop 팁 기점 — `git rev-parse --short HEAD` = `d4c45bdcc`, 이번 실행, 이 트리). 배차 기점 pin(로컬 develop `d4c45bdcc`)과 일치하며, t704 착지 커밋 `987b35204`가 그 조상임을 `git merge-base --is-ancestor`로 확인. 판정 대상 행 번호는 post-t704 파일 기준 — t704 verdict의 "111~114행"은 fixing 전 번호(내용 앵커: 읽기전용 표면 문단)이며 +6행 시프트로 현재 117~120행.

## Gaps

- CI 판정 미측정 — push 금지로 이 레인에서 관측 불가. 리드의 develop 일괄 push 이후 CI가 통합 판정을 만든다.
- `bin/moai` 실행 스모크 미수행 (문서-only 변경, 바이너리 동작과 무관).
- docs-site 표면(`utility-commands/moai-todo.md` 4개국어)의 동일 주제 문단은 배차 범위 밖 — t575가 그 표면을 수리했으나 이 문단의 문구에 같은 carve-out이 있는지는 미판독 (t704 verdict도 같은 residual-risk를 남김).
- temp-origin에서도 `resolveStateDir` 후보 스캔이 홈 *레거시* 디렉터리(`~/.moai/todo/<key>/…`, `legacyHomeStateDirsForRoot` :43-60)를 포함하는 에지(:129-135)는 문서화하지 않았다 — 본 카드의 결함류(홈 DB 무조건 서술)와 다른 층.

## Residual-risk

- 추가 문장의 "every surface, read-only or not, uses the project-local queue"는 위 Gaps 4번째 항목의 홈 레거시 후보 에지를 다루지 않는다 — 홈 레거시 디렉터리에 큐가 남아 있던 temp-origin 프로젝트라면 읽기전용 표면이 그쪽을 가리킬 수 있다. 신규 temp-origin 프로젝트에는 영향 없음. 필요 시 후속 카드.
- 진리표 5행 "home unresolvable, non-temp base"(`state_dir.go:32-34`)는 이번 문장에서 다루지 않았다 — t704와 동일하게 후속 카드 후보로 남는다.
- docs-site와 본 템플릿 문서의 문구가 정확히 일치하지 않을 수 있다(각자 문체로 기술 — t704 잔여 위험의 연장).
