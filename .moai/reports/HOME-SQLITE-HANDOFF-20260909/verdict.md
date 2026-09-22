# 홈 상태·SQLite·handoff 구현 검증 보고서

판정: **조건부 통과** — 요청한 구현과 범위 검증은 워크트리에 반영됐다. 다만 실사용
`~/.moai`의 강제 정리·전역 DB 전환과 설치 바이너리 교체는 실행하지 않았다.

## Claim

1. 프로젝트별 영속 상태 경로를 `~/.moai/db/<project-key>/` 아래로 통합했다.
   Todo는 `todo/backlog.db`, Factory와 두 종류의 handoff는
   `factory/factory.db`를 사용한다. linked worktree는 primary checkout에서 같은
   `project-key`를 얻는다.
2. Todo의 기존 프로젝트 로컬 DB/JSON은 원본 잠금 아래 논리 레코드로 복사하고,
   새 SQLite를 다시 읽어 동등성을 확인한다. 원본은 롤백 스냅샷으로 남기며 이미
   생성된 홈 큐를 후발 마이그레이터가 덮어쓰지 않는다.
3. `factory.db`는 workers, runs, cards, events, dead letters, resume handoff,
   SessionEnd memory handoff와 handoff 전이 이력을 분리해 저장한다. resume 소비는
   SQLite compare-and-swap으로 한 소비자만 선출한다. lane 등록도 한 IMMEDIATE
   트랜잭션에서 죽은 claim 정리, 빈 번호 선택, 등록을 끝낸다.
4. `moai clean --home`은 기본 dry-run이다. `--force`에서만 프로필 `projects/`
   180일, `debug/` 30일, 프로필 5 GiB 상한에 따른 후보를 삭제하고 프로필
   `~/.moai`의 모든 디렉터리 권한을 0700으로 고친다. 90일 미사용 프로필과 SHA-256으로 확인한
   중복 plugin 트리는 보고만 한다.
5. `rank.yaml`과 MoAI Rank의 현재 사용자 문서 표면을 제거했다. SPEC과 과거
   감사 보고서의 역사적 문자열은 증거 보존을 위해 수정하지 않았다.

## Evidence

### 변경 범위 테스트

명령:

```text
env GOCACHE=/private/tmp/go-build-home-sqlite go test -count=1 ./internal/homestate ./internal/hook/handoff ./internal/hook ./internal/profile ./internal/config ./internal/goal
```

관측 출력:

```text
ok  	github.com/modu-ai/moai-adk/internal/homestate	2.128s
ok  	github.com/modu-ai/moai-adk/internal/hook/handoff	9.850s
ok  	github.com/modu-ai/moai-adk/internal/hook	51.134s
ok  	github.com/modu-ai/moai-adk/internal/profile	2.027s
ok  	github.com/modu-ai/moai-adk/internal/config	4.275s
ok  	github.com/modu-ai/moai-adk/internal/goal	2.222s
```

명령과 관측 출력:

```text
env GOCACHE=/private/tmp/go-build-home-sqlite go test -count=1 ./internal/cli -run 'Test(CleanHome|ScanHomeCleanable|CheckHomeDisk|HomeDiskReport|SecureHomeDirectories|RunInit_(NonInteractive|DotArg_NonInteractive|NamedDirectoryArg|WithRootFlag_DeepPath)|Handoff|TodoHelp|DeprecatedPaths|TempOriginGuidance|ResolveFactoryWorkerName|CC_FactoryEntryThroughRunCC|GLM_FactoryWorkerEntry)'
ok  	github.com/modu-ai/moai-adk/internal/cli	7.770s

env GOCACHE=/private/tmp/go-build-home-sqlite go test -count=1 ./internal/web -run 'Test(ResolvedWatchPaths|LaneSection|FactoryLanes)'
ok  	github.com/modu-ai/moai-adk/internal/web	10.427s

env GOCACHE=/private/tmp/go-build-home-sqlite go test -count=1 ./internal/template
ok  	github.com/modu-ai/moai-adk/internal/template	26.667s

env GOCACHE=/private/tmp/go-build-home-sqlite go test -count=1 ./internal/kanban -run 'Test(RelocateQueueArtifacts|StateDir|ClaimFactoryWorkerName|FactoryRegistry|FactoryFreeSlots|PruneFactory)'
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.696s

env GOCACHE=/private/tmp/go-build-home-sqlite go test -count=1 ./internal/cli -run 'Test(ResolveFactoryWorkerName|CC_FactoryEntryThroughRunCC|GLM_FactoryWorkerEntry)'
ok  	github.com/modu-ai/moai-adk/internal/cli	3.294s

env GOCACHE=/private/tmp/go-build-home-sqlite go test -race -count=1 ./internal/kanban -run 'TestClaimFactoryWorkerNameConcurrentClaimsAreUnique|TestRelocateQueueArtifactsNeverOverwritesExistingHomeQueue'
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.944s
```

### 정적 검사와 빌드

명령:

```text
env GOCACHE=/private/tmp/go-build-home-sqlite go vet ./internal/homestate ./internal/kanban ./internal/hook/handoff ./internal/hook ./internal/cli ./internal/profile ./internal/web ./internal/config ./internal/goal ./internal/template
```

관측 출력: 없음, 종료 코드 0.

명령과 관측 출력:

```text
env GOCACHE=/private/tmp/go-build-home-sqlite go build -o /private/tmp/moai-home-state ./cmd/moai
go: writing stat cache: open /Users/goos/go/pkg/mod/cache/download/github.com/modu-ai/moai-adk/@v/v0.0.0-20260901084916-7ad9f8534dc4.info202646144.tmp: operation not permitted
```

종료 코드는 0이며 `/private/tmp/moai-home-state version`은 `moai-adk v3.1.3`을
출력했다. 위 한 줄은 사용자 module stat cache의 샌드박스 쓰기 거부 경고다.

`git diff --check`의 관측 출력은 없고 종료 코드는 0이었다.

### 임시 홈 종단 검증

실사용 홈과 분리한 `MOAI_HOME=/private/tmp/moai-home-e2e.D9Scmy`에서 빌드된
바이너리로 Todo 추가, resume handoff 저장, dry-run 정리를 차례로 실행했다.

관측 출력:

```text
t591 43
handoff saved: /private/tmp/moai-home-e2e.D9Scmy/db/moai-adk-go-1bd3d038/factory/factory.db
· nothing to clean under /private/tmp/moai-home-e2e.D9Scmy (retention 30d)
```

생성된 디렉터리는 모두 0700이었고 SQLite 파일은 모두 0600이었다.

```text
600 /private/tmp/moai-home-e2e.D9Scmy/db/moai-adk-go-1bd3d038/factory/factory.db
600 /private/tmp/moai-home-e2e.D9Scmy/db/moai-adk-go-1bd3d038/todo/backlog.db
dropped|27
picked|4
queued|43
pending|run|9cc4f7d4-a9c0-4b4b-8493-216a1782edef|resume e2e probe
```

이 검증은 프로젝트 로컬의 기존 큐를 읽어 임시 홈으로 복사했기 때문에 새 카드가
`t591`로 발급됐다. 쓰기 대상 DB는 임시 홈이었지만 프로젝트 로컬 원본의 전후
해시는 측정하지 않았으므로 원본 전체가 바이트 단위로 불변이었다고 주장하지 않는다.

### 현재 실사용 홈의 읽기 전용 확인

```text
legacy_root_backlog=absent
rank_yaml=absent
ok
dropped|27
picked|4
queued|42
archived|228
total|301
items	75
archived	213
total	288
legacy_ids|288
matched_in_db|288
missing_from_db|0
```

위 `ok`는 현재 프로젝트 로컬 SQLite에 대한 `pragma integrity_check` 결과다.
기존 JSON의 ID 288개가 SQLite에 모두 있고, SQLite에는 이후 추가된 레코드를
포함해 301개가 있다.

최종 빌드 바이너리의 `doctor --check 'Home Disk Usage' --export ...` 결과:

```text
status=ok
~/.moai 7.8G (claude-profiles 7.3G, backups 362.2M, releases 139.4M) — cleanable ~64B; ~/.claude 550.3M (report-only)
profiles: moai-adk 3.9G; mo.ai.kr 3.2G; moai-cowork 183.1M; __no_such_profile__ 67.2M; moai-code 20.0M
~/.moai directories not mode 0700: 10340 (repaired by 'moai clean --home --force')
cleanable estimate: ~64B under 30d retention — 'moai clean --home' (dry-run by default)
```

실사용 홈 dry-run:

```text
· [dry-run] Would delete [logs] logs/census-stderr.log (64B)
· 1 path(s), 64B eligible under 30d retention. Run with --force to delete.
```

현재 사용자 문서·코드의 Rank 문자열과 구 크기 기반 중복 함수 검색 계수:

```text
0
0
```

새 Todo 홈 경로의 현재 배포 표면 검색 계수:

```text
11
```

## Baseline-attribution

모든 위 검증은 이 실행에서 다음 워크트리와 커밋을 기준으로 측정했다.

```text
HEAD: 817990f67
branch: WT-home-sqlite-handoff
origin/main...HEAD: 0 2518
session: 9cc4f7d4-a9c0-4b4b-8493-216a1782edef
```

구현은 커밋하지 않은 워크트리 변경이다. primary checkout의 branch 상태는 바꾸지
않았다.

## Gaps

- 새 바이너리를 설치하지 않았고 실사용 `~/.moai/db/<project-key>/`로 Todo와
  Factory를 전환하지 않았다. 실제 전환은 새 바이너리의 첫 adopting 명령에서
  일어난다.
- `moai clean --home --force`를 실행하지 않았다. 따라서 현재 확인된 0700 위반
  디렉터리 10,340개와 삭제 후보 64B는 그대로다.
- `~/.moai/cache/search/<project-key>/sessions.db`는 경로와 디렉터리 계약만
  마련했다. 이 저장소에서 해당 DB를 쓰는 검색 인덱서 생산자는 확인되지 않아
  옮기지 않았다.
- `active-sessions`, `companions`, `leads` 같은 세션 런타임 JSON은 아직 프로젝트
  로컬 호환 경로에 남는다. Factory workers와 handoff만 이번 범위에서 SQLite로
  옮겼다.
- 전체 `internal/web` 테스트는 샌드박스의 소켓 bind 제한으로, 전체
  `internal/cli` 테스트는 Unix socket 경로 제한·Codex app-server handshake·TCP
  bind 제한으로 통과 판정을 내리지 못했다. 변경 관련 테스트는 위와 같이 따로
  통과했다.
- docs-site 빌드와 브라우저 렌더링은 실행하지 않았다.
- commit, push, PR은 수행하지 않았다.

## Residual-risk

- 구 버전 바이너리와 새 버전 바이너리를 동시에 오래 실행하면 구 버전은 남겨 둔
  프로젝트 로컬 큐에 계속 쓸 수 있다. 전환 시 모든 세션을 종료하고 한 버전만
  재기동하는 운영 절차가 필요하다.
- plugin 중복 검사는 프로필별 `plugins/` 트리 전체가 동일한 경우만 보고한다.
  트리 일부만 중복된 경우의 절감 가능 용량은 아직 집계하지 않는다.
- `factory.db`의 handoff 단일 소비자 성질은 동시 goroutine 테스트로 검증했지만,
  Claude lead와 Codex lane을 함께 띄운 실제 다중 프로세스 팩토리 종단 통신은 다음
  단계다.

## 다음 작업 제안

1. **High — 무중단 전환 명령**: `moai migrate home-state --dry-run`으로 source/target
   레코드 수·해시·활성 세션을 먼저 확인하고, 활성 세션 0일 때만 실제 전환한다.
2. **High — 런타임 레지스트리 SQLite화**: `active-sessions`, companions, leads를
   `factory.db`의 roster/run 관계로 합치고 프로젝트 로컬 JSON은 읽기 전용 호환으로
   내린다.
3. **High — 양방향 Factory transport 종단 테스트**: Claude lead ↔ Codex lanes와
   Codex lead ↔ Claude lanes를 같은 envelope/ack/dead-letter 계약으로 검증한다.
4. **Medium — search 소유자 결정**: `sessions.db`의 실제 생산자와 무효화 정책을
   먼저 정한 뒤 `~/.moai/cache/search/<project-key>/` 전환 여부를 결정한다.
5. **Medium — 운영자 승인 후 홈 정리**: dry-run 결과를 다시 읽고 별도 승인 뒤에만
   `moai clean --home --force`를 실행한다.
