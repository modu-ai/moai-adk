# t751 판정서 — moai-todo.md 임시 루트 가드 문장 census

- 카드: t751 (Tier S, Class B — plan 생략, run → sync)
- 브랜치: `WT-todo-doc-tempguard` (워크트리 `.claude/worktrees/t751`)
- 기점: develop `4da5d1c4e` (배차 기점과 일치 — t658·t707 병합 포함)
- 변경: **product diff 없음** — 유일 산출물은 본 판정서. 수리의 반출 주체는 t575 커밋(아래).
- push 금지 · 병합 창은 리드 지명 대기

## Claim (주장)

1. **카드가 지목한 결함은 이미 수리돼 있다.** 카드 본문(t536 sync-audit F3, 2026-09-08 발행)이 가리킨 "git 메타데이터 없는 프로젝트는 `~/.moai/todo/<project-key>/`에 큐를 둔다" 문장 계열은 t575 — 커밋 `55adec04f` (docs(SPEC-DOCS-TODO-TEMP-GUARD-001), 2026-09-13 21:19) — 이 moai-todo.md 4로케일의 59·248행을 carve-out 문장으로 재작성하면서 소실됐다. 본 트리에서 그 구(舊) 모양은 0건이고, 현행 문장은 코드 동작과 일치한다(아래 census).
2. **census 결과: 같은 결함류의 잔존은 대상 표면에서 0건이다.** "가드 도입으로 거짓이 된 자리"는 moai-todo.md ×4 안에 더 없고, docs-site 전체·README ×4·factory-mode.md ×4에서 구 모양(`~/.moai/todo/`) 잔존도 0건이다.
3. 따라서 배차의 "4로케일 같은 커밋 수정" 단계는 **수리 대상이 소실된 상태**다 — 정직한 완료는 문장을 고치는 것이 아니라 census 근거를 판정서로 반출하는 것. 리드가 판독 뒤 별도 수정을 지시하면 후속으로 집행한다.

## Census 목록 (표면 × 확인한 주장 × 판정)

| # | 표면 | 확인한 주장 | 판정 |
|---|------|------------|------|
| 1 | moai-todo.md ×4 :59 | 큐 위치 — 홈 `~/.moai/db/<project-key>/todo/backlog.db`, 임시 기점(`os.TempDir()`, `/tmp`, `/var/folders`)이면 절대 `MOAI_HOME` 재정의 없는 한 `<base>/.moai/state/todo/backlog.db` | **코드와 일치** — `StateDirForRoot` (state_dir.go:26-41), `defaultTempRoots` 3종 앵커 (temp_origin.go:60-62), `paths.EnvHome = "MOAI_HOME"` (paths.go:30), 임시+비절대 재정의→로컬 분기 |
| 2 | moai-todo.md ×4 :248 | 워크트리→프라이머리 프로젝트 키 귀속, 비git도 같은 홈 구조, 임시 carve-out | **코드와 일치** — `ResolveTodoQueueRoot`→`primaryCheckoutRoot` (todo_root.go:81-99, 172-179) |
| 3 | moai-todo.md ×4 parity | t575 수정의 4로케일 균일성 | **동일** — `MOAI_HOME`=2 · `var/folders`=1 · `state/todo`=2 · `db/<project-key>`=1, 네 로케일 모두 같은 카운트 |
| 4 | moai-todo.md 전체 260행 (개요~관련 문서 전수 정독) | 기타 위치·동작 주장 | 위치 주장은 1·2가 전부. 동작 주장 6건 코드 대조 통과 — "여섯 가지 귀속 위치"=prlink_landed.go:16 "the six positional shapes" · WAL SQLite 트랜잭션(:120) · export-json=데이터베이스 옆 backlog.json(:244=state_dir.go:254-258 형제 산출물) · "커밋되지 않음"=템플릿 .gitignore:240 `.moai/state/` · 포인터 대상 `.moai/docs/todo-queue-storage.md`=템플릿에 존재(internal/template/templates/.moai/docs/) · UUID 계약(t648 문면 현행) |
| 5 | docs-site 전체 + README ×4 | 구 모양 `~/.moai/todo/<project-key>` 잔존 | **0건** |
| 6 | factory-mode.md ×4 + README ×4 (t706 수리 표면) | factory.db 임시 carve-out 유지 여부 | **유지** — 4로케일 카운트 1/1/1/1, en/ko :80 문면 대조(EN "temporary one (no absolute MOAI_HOME override)" = KO "임시 디렉터리면(절대 MOAI_HOME 오버라이드 없음)") |
| 7 | 런타임 (이 트리) | 문장이 서술하는 가드 동작의 실행 증명 | **5/5 PASS** — `go test ./internal/kanban/ -run 'TempOrigin\|StateDirForRoot' -count=1` → 선택자 일치 5 = 실행 5, `ok … 1.338s` |

## Evidence (증거 — 이번 실행, 이 트리에서 관측)

- **수리 반출 주체 확인**: `git show 55adec04f` 커밋 메시지 — "Line 59 and line 248 of ko/en/ja/zh utility-commands/moai-todo.md claimed the queue lives at ~/.moai/db/<project-key>/todo/backlog.db unconditionally — false for temporary-origin bases … Both sentences rewritten per locale with the carve-out; heading parity 29/29/29/29 preserved." diff에서 ko의 ±문장이 현행 59·248행과 대응하는 것을 직접 대조.
- **구 모양 소실**: `grep -rn "\.moai/todo/" docs-site/content/ README*.md` → 0행.
- **parity 카운트**: 위 census 3행의 네 로케일 마커 카운트 (이번 실행 grep -c 관측).
- **런타임**: census 7행의 테스트 명령과 출력 (선택자 수=실행 수 대조 포함).
- **정적**: census 4행의 코드 인용 위치 전부 이번 트리에서 Read/grep으로 판독.

## Baseline-attribution (baseline 귀속)

모든 측정은 이번 턴에 워크트리 `.claude/worktrees/t751` (브랜치 `WT-todo-doc-tempguard`, 기점 `4da5d1c4e`)에서 실행했고 출력은 그대로 옮겼다. t575 커밋의 내용은 `git show`로 이번 트리에서 재판독했다(기억 재사용 아님).

## Gaps (미검증)

- **다른 가드 계열은 census 밖이다.** 카드의 "가드"는 임시 루트 가드(SPEC-TODO-HOME-TEMP-GUARD-001 계열, t706 형제 포함)로 문맥상 한정돼 있고, 통합 잠금·settings-drift·branch guard 같은 다른 가드 계열의 문서 모순은 이번 census가 보지 않았다.
- **ja/zh 문면은 카운트·구조 대조**이지 문장별 정밀 독해가 아니다(en/ko는 문면을 나란히 대조했다). hugo 빌드 판정은 리드 게이트 몫(product diff가 없어 이번 변경으로는 영향 0).
- "여섯 가지 귀속 위치"의 여섯 셀 각각을 목록 순서까지 대조하지는 않았다 — 개수(6)와 존재 코드(landedSubjectForms)까지가 이번 대조 범위.
- **product diff가 없으므로** 배차의 "같은 커밋 수정"은 실행되지 않았다 — 리드가 census 결과와 다른 판단을 하면 후속 지시를 받는다.

## Residual-risk (잔여 위험)

- 가드의 앵커 집합이 다시 바뀌면(예: `/var/tmp` 포함) 같은 결함류가 재발한다 — 문장이 앵커 3종을 **리터럴로 열거**하고 있기 때문이다. 추적 주체는 이번 판정서와 t575 SPEC이며, 다음 앵커 변경 시 이 표면을 다시 열 것.
- 홈의 `moai-todo` 문서는 배포 시점에 고정된다 — 사용자가 구버전 템플릿을 쓰면 여전히 구 문장을 본다. 이는 배포 채널의 성질이며 본 카드 범위 밖.

## 동기화 메모

- 카드 본문이 지적한 "추적 주체 부재"는 해소됐다 — 결함의 수리는 t575 SPEC + 커밋 `55adec04f`가, 잔존 census의 판정은 본 문서가 운반한다.
- CHANGELOG: product diff가 없어 추가하지 않았다(변경 없음을 기록하는 항목은 넣지 않는다 — 기존 관례 준수).
