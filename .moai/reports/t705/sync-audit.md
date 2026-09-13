# t705 sync-audit — 독립 검토 기록 (sync-auditor, 2026-09-14)

- 대상: card t705 · branch `WT-tempgit-refusal` @ 97cd3a562 (당시) · tree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t705`
- 이 문서는 세션 로그에만 존재하던 감사 보고를 리드 지시(2026-09-14, Class B sync 게이트 증거 요건)로 영속화한 것. F1 수리 커밋 8dcfa2470은 보고 이후 레인이 적용.

**종합 판정: PASS** — 6행 수리는 정확하고, 패리티 입증 + revert-check RED→GREEN. 병합 전 lint 1건(F1, 테스트 파일 1줄)만 창 전 해소 필요(→ 8dcfa2470으로 해소됨).

## 의무별 결과

**1. 재실행** — PASS 19/0. `go test -v ./internal/kanban/ -run 'TestTempOriginRefusalMatchesStateDirForRootForTempGit|TempOrigin|TodoQueueRoot' -count=1` → `exit=0`, `PASS=19 FAIL=0`, `ok github.com/modu-ai/moai-adk/internal/kanban 5.570s` — 본 워크트리 HEAD 97cd3a562에서 본 세션 측정.

**2(a) 비임시-git 케이스 불변** — 코드 대조로 입증(추정 아님). 삭제된 분기(`primaryCheckoutRoot` 조기 반환)는 git 응답 가능 집합만 커버. 수리 후 해당 base는 `TempOriginReason`(internal/kanban/temp_origin.go:75)으로 흐름 — 앵커 집합 `{os.TempDir(), "/tmp", "/var/folders"}`를 정규화(EvalSymlinks)·어휘 양쪽으로 비교, 성분별 포함(`pathWithin` — `/tmpfoo` 안전). 앵커 밖 git 저장소 → `isTemp=false` → `refused=false`: 판정 동일. 비-git base는 애초 git-first 분기를 안 탐. `explicitMoaiHome()` 면제 첫 분기 유지. 부수 효과: 거절 함수가 git 셸아웃을 안 하게 됨 — 문서화된 읽기전용 경로의 부작용 감소. 판정이 달라지는 유일 케이스 = 임시 앵커 안 git 저장소 — 의도된 수리 그 자체. macOS `/var`↔`/private/var` 정규화는 비교 양변에서 처리되며 GREEN darwin 실행이 temp-git 형태의 직접 증거.

**2(b) 호출자 의존** — 생산 호출자 정확히 1곳(grep 검증): `internal/cli/todo.go:95` `warnTempOriginQueueRefusal`, `PersistentPreRun`으로 1회 배선(todo.go:259-261). `refused=true`에서 stderr 안내 1줄 출력, exit 코드 변경 없음(temp-dir 스크립트 계속 동작 — todo.go:86-89에 의도 문서화). `refused=false`로 무언가를 건너뛰는 분기 없음. 핵심: `StateDirForRoot`(state_dir.go:26-41)는 수리 전부터 temp-git 큐를 프로젝트 로컬에 유지(git 무관) — 큐 위치가 바뀌는 사람은 아무도 없고, 안내 채널만 조용히-틀림→올바름으로 변경.

**2(c) doc comment 진실성** — 예. StateDirForRoot는 git을 보지 않음; temp 분기는 경로 기반 `TempOriginReason`만으로 발화; override 조건 `override == "" || !filepath.IsAbs(override)`(state_dir.go:29)가 `explicitMoaiHome()`과 양방향으로 정확히 대응(절대 override는 양쪽에서 가드를 이김). "Read-only" 주장도 수리 후 성립(Getenv + 경로 연산 + Lstat/EvalSymlinks만).

**3. revert-check** — 기계 실행, 트리 무손대: 부모 커밋 버전의 `todo_root.go`를 `go test -overlay /tmp/t705-audit/overlay.json`으로 주입 → 신규 테스트가 정확히 결함 서명으로 FAIL: `guidance side: TempOriginRefusal reported refused=false while StateDirForRoot kept the queue project-local (.../project/.moai/state/todo)` + `substitute = ""` + `named no matched temp root`. `todo_root_temp_git_refusal_test.go:53-55`의 단언이 refused=false 경로를 명시적으로 커버. 출력이 부수적으로 해석 측은 수리 전후 불변임을 재확인 — 결함은 안내 전용, 카드 프레이밍과 일치.

## Findings

- **F1** [low][blocking→해소] `todo_root_temp_git_refusal_test.go:50` `filepath.HasPrefix` 폐기 예정(SA1019) — `golangci-lint run ./internal/kanban/` → 1 issues 실측. 수리: 동일 패키지 `pathWithin(resolved, dir)` 1줄(커밋 8dcfa2470). 수리 후 재측정: `0 issues`, 테스트 PASS 재확인(레인, 2026-09-14).
- 생산 변경 자체의 차단 결함 없음; 기능/보안/일관성 축에서 추가 발견 없음(단일 호출자 표면, 입력 처리 변경 없음, 기존 코드 무접촉).

## Gaps (이 감사가 관측하지 않은 것)

- internal/kanban 전체 스위트와 호출자 패키지 internal/cli 미재측정 — "패키지 전체 ok 181.5s"는 레인 주장(범위 셀렉터만 재실행). 전체 판정은 develop push 시 CI 몫.
- GitHub CI가 internal/에 golangci-lint를 게이트하는지 미확인 — F1의 병합 귀결이 여기에 의존(해소 후 무관).
- 실제 temp-git 디렉터리에서의 엔드투엔드 `moai todo` 미수행(단위+호출부 판독만; PersistentPreRun 배선은 기존 코드).

## Residual-risk

- 유일 런타임 델타 표면: temp-git 디렉터리에서 `moai todo` 기동 시 stderr 안내 1줄 — 권고성, 큐 위치·exit 코드·스크립트 계약 불변. revert-check scratch는 /tmp/t705-audit/에 있으며 감사 트리 무접촉.
