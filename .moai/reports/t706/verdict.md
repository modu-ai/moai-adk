# t706 verdict — factory.db 위치: 임시 루트 가드 뒤에서의 docs 주장 측정 + 형제 해석기 가드 불일치 수리

card: t706 (Class B — run) · branch: `WT-factory-queue-path` · base: local develop `6d5943052` · push: 금지 · 병합 순번: t658 → t584 → t706 (리드 확인)

## Claim

factory-mode.md:93 + README 4파일:80의 "레인 소유권은 `~/.moai/db/<project-key>/factory/factory.db`에 기록된다" 주장은 소스 판독+실험으로 측정한 결과 **guard가 실제 발동하는 클래스에서 거짓**이었고, 근원은 homestate 쪽 임시 판별식이 todo 큐 해석기의 앵커 집합보다 좁은 **형제 해석기 가드 불일치**(t687 교훈과 동일 결함류)였다. 수리 판정: **코드+문서 병행** — 코드(`internal/homestate/paths.go` 3개 지점의 앵커 집합을 todo 쪽 3종과 일치)로 split-brain을 해소한 뒤, 문서 8개 표면(4로케일 동시)에 todo 큐와 같은 carve-out을 추가했다. 리드 승인 조건 3개(패리티 테스트·전후 실측 기록·probe 삭제+문서 동일 커밋) 모두 이행.

## Evidence

### 결함 구조 (소스 판독)

- todo 큐 해석기 `kanban.StateDirForRoot`의 임시 가드는 `TempOriginReason`(`internal/kanban/temp_origin.go`): 앵커 3종 — `os.TempDir()` + `/tmp` + `/var/folders` (`TempRootsFn`/`defaultTempRoots`, REQ-THG-002; `/var/tmp`는 부팅 생존 사유로 의도적 제외).
- factory.db 해석기 `homestate.FactoryDBPath` → `ProjectDir`(`internal/homestate/paths.go`)의 임시 가드는 **`pathInside(canonical, os.TempDir())` 단일 앵커** — `/tmp`, `/var/folders` 어휘 앵커를 놓친다. 같은 단일 앵커 검사가 `EnsureProjectLayout` 내 2곳(:177, :194)에 더 있었다.
- `TempRootsFn` 주석 자체가 이 함정을 기록하고 있었다: "os.TempDir() alone … would NOT have been caught" (t203-probe 홈 오염 복구 사례).

### 수리 전 실측 (결함 상태, `d4c45bdcc`~`6d5943052` 트리, probe 빌드 Commit=6d5943052)

기본 환경(macOS, `TMPDIR=/var/folders/kt/.../T/`, `MOAI_HOME` 미설정):

- **케이스 A — 비-git 기점 `/tmp/t706-exp/tmpbaseA.FQAi69`** (split-brain 재현):
  ```
  kanban.StateDirForRoot  = /tmp/t706-exp/tmpbaseA.FQAi69/.moai/state/todo
  homestate.FactoryDBPath = /Users/goos/.moai/db/tmpbaseA.FQAi69-200d2aa8/factory/factory.db
  kanban.TempOrigin       = TEMP (/tmp)
  ```
  → 큐는 프로젝트 로컬, factory.db(레인 소유권)는 **홈** — 같은 프로젝트의 상태가 두 트리로 갈라짐. docs 주장은 문자상 참이나 시스템이 불일치.
- **케이스 B — git 기점 `/tmp/t706-exp/tmpbaseB.in3bsF`** (`git init` fixture): todo 로컬 + factory.db 홈 — A와 동일 split-brain.
- **케이스 D — 같은 A 기점 + `TMPDIR=/tmp/t706-exp/othertmp`**: todo 로컬 + factory.db 홈 — 판별식 불일치가 환경과 무관하게 재현.
- **케이스 C — 비임시 기점(워크트리 내 case-c-base)**: todo 홈 + factory.db 홈 — 일치, docs 주장 참.
- **케이스 E — `$TMPDIR` 안 기점(`mktemp -d` 기본값)**: `FactoryDBPath = <base>/.moai/db/<key>/factory/factory.db` **프로젝트 로컬** — guard가 여기서만 발동하며, **바로 이 클래스에서 docs 주장이 거짓**.

### 수리

- `internal/homestate/paths.go`: `tempRoots()`(3종 앵커, `/var/tmp` 제외 유지) + `insideTempRoots()` 헬퍼 추가, 3개 지점(ProjectDir :68, EnsureProjectLayout :177/:194)을 교체. kanban은 접촉 없음(t658 충돌 최소화).
- `internal/homestate/temp_parity_test.go` (신규, `package homestate_test` — homestate→kanban 사이클 회피): `kanban.TempOriginReason` 판정과 `FactoryDBPath` 프로젝트 로컬성이 앵커 4클래스(os.TempDir·/tmp·비임시·/var/tmp 제외) 전체에서 일치함을 단언 — 리드 조건 1(가드 갈라짐 재발 방지).

### 수리 후 실측 (probe 재빌드, 동일 환경)

- 케이스 A: `homestate.FactoryDBPath = /private/tmp/t706-exp/tmpbaseA.FQAi69/.moai/db/tmpbaseA.FQAi69-200d2aa8/factory/factory.db` — **프로젝트 로컬, todo 큐와 일치**.
- 케이스 D: 동일하게 프로젝트 로컬 — 일치.
- 케이스 E: 프로젝트 로컬 유지 — 일치.

### 검증 배치

- `go vet ./internal/homestate/...` → 통과
- `go test ./internal/homestate/...` → `ok … 16.808s` (패리티 테스트 포함)
- `go test ./internal/kanban/...` → `ok … 273.045s`
- `make build` → exit 0 (`Commit=6d5943052` 스탬프)
- docs 8파일(4로케일 동시): factory-mode.md ×4 + README ×4, 각 표면의 기존 문체 유지 — 리드 조건 3

## Baseline-attribution

측정 트리: `.claude/worktrees/t706` @ `WT-factory-queue-path`, 기점 `6d5943052`(= origin/develop tip, 이번 실행 `git rev-parse` 관측). 수리 전 실측은 수정 전 작업 트리에서 빌드한 probe(Commit=6d5943052), 수리 후 실측은 수정 적용 작업 트리에서 재빌드한 probe-fixed. 모든 출력은 이번 실행의 관측값. 실험 후 잔재 정리 완료: probe binaries(`/tmp/t706`), fixtures(`/tmp/t706-exp`, 케이스 E 디렉터리), 수리 전 `-materialize`가 실제 홈에 만든 빈 디렉터리 2곳(`~/.moai/db/tmpbaseA.FQAi69-200d2aa8`, `~/.moai/db/tmpbaseB.in3bsF-f06f84a1` — 내용 0 확인 후 삭제), `probe_t706/`(커밋 전 삭제 — 리드 조건 3).

## Gaps

- factory.db **쓰기 경로의 종단 간 실행 미수행** — `moai cc -f` 런처는 실제 세션을 띄우므로 실험에서 제외했고, factory.db 개방은 해석 함수 + `-materialize`(디렉터리 생성 관측) + 단위/패리티 테스트로 대체했다. 런치 레벨 스모크는 미측정.
- docs-site hugo 빌드 미실행 — 변경은 평문 문장 1개×4로 shortcode·mermaid 비접촉이나, 빌드 판정은 리드 게이트 몫.
- CI 판정 미측정 — push 금지. 리드 일괄 push 이후 통합 판정.
- `SearchDBPath`·`RunProjectDir` 자체에는 임시 가드가 없다(생성은 EnsureProjectLayout이 통제) — 이번 카드 범위 밖, 후속 후보.
- SPEC-DOCS-TODO-TEMP-GUARD-001 §2 진리표 1행("Git repository → home")은 git 저장소가 임시 디렉터리 안에 있는 경우를 다루지 않는다 — 케이스 B 실측에서는 git-in-/tmp도 guard가 발동해 프로젝트 로컬이었다(코드 동작 기준). 진리표 정밀화는 SPEC 소관 후속.

## Residual-risk

- homestate의 앵커 3종은 kanban과의 사이클 때문에 복제돼 있다 — 패리티 테스트가 불일치를 잡지만, t658(큐 단일 원천)이 착지하면 단일 원천으로 접는 후속이 권장된다.
- `internal/homestate/paths.go`는 t658의 예상 영역(큐 단일 원천)과 겹친다 — 변경 파일 목록에 명시했으며 병합은 t658 뒤(리드 지시). t658이 paths.go를 고쳤다면 병합 충돌 시 본 카드의 3지점 교체 + 헬퍼가 기준이다.
- 홈 `~/.moai/db/`에는 과거(수리 전)에 만들어진 임시 기점 프로젝트의 잔존 레인 레지스트리가 있을 수 있다 — 본 카드는 미청소(사용자 데이터 판별 불가), 필요 시 별도 카드.
