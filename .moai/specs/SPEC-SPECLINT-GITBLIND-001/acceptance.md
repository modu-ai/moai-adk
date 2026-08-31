# Acceptance: SPEC-SPECLINT-GITBLIND-001

Tier M. 11개 AC 전부 명령과 관측 대상을 명시한다. 주관적 판단어("적절히", "잘")를 쓰지 않는다.

**공통 픽스처 규약** — AC-SLGB-001..007 은 `t.TempDir()` 안에 `git init` 으로 만든 저장소를 쓴다.
`internal/spec/` 에 선례가 있다(`drift_chore_skip_test.go`, `archive_git_test.go`, `closer_test.go`).
두 가지 구속이 모든 픽스처 테스트에 걸린다:

- **비병렬 [HARD]** — per-run git-query 캐시는 패키지 전역(`gitQueryCacheMu` / `gitQueryCacheV`,
  `internal/spec/gitquery_cache.go:21-23`)이다. 캐시를 건드리는 테스트는 `t.Parallel()` 을 쓰지 않는다.
  이 구속은 이 SPEC 이 새로 만드는 규칙이 아니다 — 기존 헬퍼의 주석
  (`internal/spec/drift_characterization_test.go:53`)이 이미 같은 문장을 담고 있다.
- **작업 디렉터리 — 기존 선례를 그대로 쓴다** — `cachedMainBranch` 는 `exec.Command` 에 `.Dir` 을
  설정하지 않으므로 프로세스 작업 디렉터리를 따른다(`:102`, `:113`). 이것은 **시그니처 변경을 요구하지 않는다**:
  같은 패키지의 `chdirForTest`(`drift_characterization_test.go:55` — `os.Chdir` + `t.Cleanup` 복원)가
  이미 이 성질을 다루고, `setupDriftCorpusFixture`(`:98`)가 마지막에 그것을 호출한다(`:103`).
  비-git 디렉터리로 진입하는 선례도 있다(`:280`). 새 픽스처는 이 헬퍼를 재사용한다.
- **관측 표면은 기본 표 출력 하나 [HARD]** — 판정은 `printTable`(`internal/cli/spec_lint.go:113`)이
  내보내는 기본 출력에서 읽는다. `--json` 과 `--sarif` 경로의 동작은 이 조사에서 **검증된 바 없으므로**
  (`Finding` 의 JSON 태그 `internal/spec/lint.go:33-45` 는 의도만 보인다),
  어떤 AC 도 그 위에 서지 않는다.

---

## M1 — 조용한 skip 의 관측 가능화

### AC-SLGB-001 — 기준 ref 가 없으면 Info 가 뜬다 (M1 의 핵심 판정)

- **Given** 로컬 `main` / `origin/main` / `master` / `origin/master` 중 **어느 것도 없는** git 픽스처와,
  frontmatter `status` 가 `terminalStatusEnum`(`superseded` / `archived` / `rejected` / `completed`)에
  속하지 **않는** SPEC 문서 1개
- **When** 그 저장소에서 `moai spec lint` 를 실행하고 **기본 표 출력**(`printTable`)을 읽으면
- **Then** 다음 둘이 **모두** 참이다:
  1. 출력에 `StatusGitUnreachable` code 를 가진 finding 이 **1건 이상** 나타난다.
  2. 출력에 `✓ No findings — all SPEC documents are valid` 줄이 **없다**.
- **[HARD] 픽스처 전제 (RED 기준선)**: 픽스처의 SPEC 문서는 다른 lint finding 을 내지 않는
  schema-valid 문서여야 한다(`fixtureSpecMD` 선례를 따른다). 그래야 **M1 이전**에는 finding 이 0건이라
  `printTable` 이 short-circuit(`internal/cli/spec_lint.go:115-118`)하여 위 2번의 줄을 실제로 찍는다.
  구현 전 그 줄이 찍히는 것을 관측한 뒤 구현한다 — 이것이 이 AC 의 RED 다.
  다른 finding 이 섞여 있으면 short-circuit 이 애초에 발화하지 않아 2번 단언이 공허해진다.

이 AC 가 실패하면 M1 은 미달이다 — 기준 ref 가 없는 체크아웃에서 Info 가 나오지 않는 것이 곧 결함이다.
2번 단언이 없으면 Info 를 `--json` 경로에만 내고 표의 거짓 전면 무결 선언은 그대로 두는 구현이 통과한다
(`spec.md` §1.2).

### AC-SLGB-002 — 메시지가 시도한 ref 이름을 담는다 (REQ-SLGB-002 추적)

- **Given** AC-SLGB-001 과 동일한 픽스처
- **When** lint 를 실행하고 `StatusGitUnreachable` finding 의 `Message` 를 읽으면
- **Then** 메시지는 해소를 시도한 후보 ref 이름을 담는다 — 최소한 문자열 `main` 과 `master` 가
  모두 나타나며, 조건이 저장소 전체에 걸린다는 사실이 메시지에 명시된다.

이 AC 가 없으면 구현이 ref 이름을 생략해도 나머지 AC 가 전부 통과한다.

### AC-SLGB-003 — 발화는 lint 실행당 정확히 1건 (REQ-SLGB-001 상한)

- **Given** AC-SLGB-001 의 픽스처에 비-terminal status 를 가진 SPEC 문서를 **10개** 둔 상태
- **When** `moai spec lint` 를 1회 실행하면
- **Then** `StatusGitUnreachable` finding 의 개수는 **정확히 1** 이다(10 이 아니다).

### AC-SLGB-004 — 얕은 저장소에서 shape ② 와 ③ 이 발화한다

`spec.md` §2.1 의 shape 표를 두 하위 픽스처로 나눠 검증한다. 두 픽스처 모두
`git rev-parse --is-shallow-repository` 가 `true` 이고 기준 브랜치는 해소된다.

- **AC-SLGB-004a (shape ②)** — Given 얕은 창 안에 해당 SPEC-ID 를 담은 커밋이 **하나도 없는** 픽스처,
  When lint 를 실행하고 기본 표 출력을 읽으면, Then `StatusGitUnreachable` 이 나타나고
  `✓ No findings — all SPEC documents are valid` 줄이 **없다**.
- **AC-SLGB-004b (shape ③)** — Given 얕은 창 안에 해당 SPEC-ID 를 담은 커밋은 있으나
  전부 분류 불가(`chore(spec): sweep` 형태)인 픽스처, When lint 를 실행하고 기본 표 출력을 읽으면,
  Then `StatusGitUnreachable` 이 나타나고 `✓ No findings` 줄이 **없다**.
- **[HARD] 픽스처 전제 (RED 기준선)**: AC-SLGB-001 과 동일 — 두 픽스처의 SPEC 문서는 다른 finding 을
  내지 않는 schema-valid 문서여야 하며, 구현 전 `✓ No findings` 줄이 실제로 찍히는 것을 관측한다.

004b 가 iter-1 D1 이 지적한 사각이다 — `drift.go:366` 의 창 소진 경로.
`✓ No findings` 부재 단언을 두 하위 픽스처에 모두 거는 이유는 §1.2 와 같다: 이 줄이 곧 눈감긴 상태의
출력이므로, 그것이 사라지는 것이 M1 의 실제 수리다.

### AC-SLGB-005 — 완전한 저장소에서는 침묵한다 (REQ-SLGB-004, 소음 방지 반대 방향 가드)

- **Given** `git rev-parse --is-shallow-repository` 가 `false` 이고 로컬 `main` 이 해소되는 픽스처와,
  frontmatter `status` 가 **`draft`**(비-terminal 임을 명시)인 SPEC 문서 — lifecycle 커밋은 없다
- **When** lint 를 실행하면
- **Then** `StatusGitUnreachable` finding 은 **0건**이다.
- **[HARD] 제약**: 픽스처의 `status` 는 반드시 비-terminal 이어야 한다.
  `Check` 는 `internal/spec/lint.go:1318-1320` 에서 `terminalStatusEnum` 인 문서에 대해
  git 에 닿기 **전에** `return nil` 하므로, `completed` 픽스처는 규칙이 한 번도 돌지 않은 채
  "0건"을 만족시킨다 — 즉 공허하게 통과한다.
- **[HARD] mutation 확인**: 구현에서 shallow 술어 가드를 제거해
  shape ②/③ 을 무조건 발화시키도록 바꾸면 이 AC 는 **반드시 실패**해야 한다
  (기대: 0건 → 1건). 실패하지 않으면 픽스처가 규칙에 닿지 않은 것이다.

### AC-SLGB-006 — Info 는 rc 를 바꾸지 않는다 (REQ-SLGB-005)

- **Given** AC-SLGB-001 의 픽스처
- **When** `moai spec lint --strict` 를 실행하면
- **Then** 종료 코드는 `0` 이고, `StatusGitUnreachable` finding 의 severity 필드 값은 `info` 다.

---

## M2 — 기준 ref 해소 체인

### AC-SLGB-007 — 체인이 순서대로 해소한다 (REQ-SLGB-006)

- **Given** 네 가지 픽스처 — (a) 로컬 `main` 만 존재, (b) `origin/main` 만 존재,
  (c) 로컬 `master` 만 존재, (d) 넷 다 없음
- **When** 각 픽스처에서 `cachedMainBranch` 를 호출하면
- **Then** 반환값은 차례로 `main` / `origin/main` / `master` / **해소 불가 신호**이며,
  (d) 에서 리터럴 문자열 `"master"` 는 반환되지 **않는다**.
- 픽스처 진입은 공통 규약의 `t.Chdir` + 비병렬 구속을 따른다.

### AC-SLGB-008 — 해소 결과가 실행당 1회만 계산된다 (REQ-SLGB-007)

- **Given** per-run 캐시가 활성인 상태(`startGitQueryCache()` 호출 후)에서
  `cachedMainBranch` 를 1회 호출한 뒤, 해당 브랜치 ref 를 삭제한 상태
- **When** 같은 실행 안에서 다시 호출하면
- **Then** 첫 호출과 **동일한 값**이 반환된다.
- **[HARD] mutation 확인**: `gitEnvCache.mainBranchSet` 조기 반환
  (`internal/spec/gitquery_cache.go:96-100`)을 제거하면 이 AC 는 **반드시 실패**해야 한다
  (두 번째 호출이 삭제된 ref 를 다시 조회해 다른 값을 낸다). 실패하지 않으면
  이 AC 는 오늘 트리에서 이미 통과하는 상속된 초록이며 아무것도 검증하지 않는다.
- **[HARD] 해소 불가 캐싱**: (d) 픽스처에서도 두 번째 호출이 재조회하지 않아야 한다 —
  해소 불가는 SPEC 마다 4회 `git rev-parse` 를 spawn 하지 않는다.

---

## M3 — CI 워크플로

### AC-SLGB-009 — 체크아웃이 완전한 이력과 `main` ref 를 갖춘다 (REQ-SLGB-008)

- **When** `.github/workflows/spec-lint.yml` 을 읽으면
- **Then** 다음 셋이 모두 참이다:
  1. `actions/checkout@v7` 단계가 `fetch-depth: 0` 을 담은 `with:` 블록을 가진다.
  2. `main` ref 를 fetch 하는 단계가 존재한다.
  3. 그 fetch 단계는 `go run ./cmd/moai spec lint` 실행 단계보다 **앞선다**.

### AC-SLGB-010 — trigger paths 가 규칙 소스와 워크플로 자체를 덮는다 (REQ-SLGB-009)

- **When** `.github/workflows/spec-lint.yml` 의 `on.pull_request.paths` 와 `on.push.paths` 를 읽으면
- **Then** 두 목록 모두 (a) `.moai/specs/**`, (b) lint 규칙이 사는 Go 소스
  (최소한 `internal/spec/**`), (c) 이 워크플로 파일 자체를 덮는 패턴을 담는다.
- 근거: 현재 두 트리거 모두 `paths: ['.moai/specs/**']` 단독이라,
  이 SPEC 의 M1/M2(전부 `internal/spec/` Go 변경)가 착지해도 잡이 재실행되지 않는다.
  그러면 AC-SLGB-011 이 영구히 닫히지 않는다.

### AC-SLGB-011 — 규칙이 실제로 돌았음을 CI 로그로 확인한다

- **When** 이 SPEC 착지 후 `SPEC Lint` 잡이 실제로 완주한 run 의 로그를 읽으면
- **Then** `StatusGitConsistency` finding 이 **1건 이상** 나타난다.
- 0건이면 규칙이 여전히 안 돌고 있다는 뜻이다 — "초록"과 "돌았음"을 가르는 유일한 관측이다.
- **[HARD] 판정 장소**: 이 AC 는 **로컬에서 판정 불가**하며 CI 로그로만 닫힌다.
  run-phase 에서 PASS 로 표시하지 않는다. 착지 후 조용한 head 의 잡 로그를 읽어 닫는다.

---

## 커버리지 표 (REQ → AC)

| REQ | 닫는 AC |
|---|---|
| REQ-SLGB-001 (발화 + 실행당 1건) | AC-SLGB-001, AC-SLGB-003 |
| REQ-SLGB-002 (ref 이름) | AC-SLGB-002 |
| REQ-SLGB-003 (shallow → shape ②③) | AC-SLGB-004a, AC-SLGB-004b |
| REQ-SLGB-004 (완전한 저장소 침묵) | AC-SLGB-005 |
| REQ-SLGB-005 (Info, rc 불변) | AC-SLGB-006 |
| REQ-SLGB-006 (해소 체인) | AC-SLGB-007 |
| REQ-SLGB-007 (실행당 1회 계산) | AC-SLGB-008 |
| REQ-SLGB-008 (체크아웃) | AC-SLGB-009 |
| REQ-SLGB-009 (trigger paths) | AC-SLGB-010 |
| — (반-공허 관측) | AC-SLGB-011 |

`spec.md` §4 의 `DetectDrift` 동작 변화는 **어떤 AC 도 덮지 않는다**. 의도된 공백이며
그 사실 자체가 §4 에 기록돼 있다 — run-phase 는 그 경로를 미검증 잔여 위험으로 안고 들어간다.

## Definition of Done

- AC-SLGB-001 … AC-SLGB-010 전부 통과, 각 테스트 함수를 `go test -run <name> -v ./internal/spec/...`
  로 직접 지목해 **실행 사실**을 확인한다(셀렉터 0건 매치 통과 방지).
- AC-SLGB-005 와 AC-SLGB-008 의 mutation 확인을 각각 수행하고, 심은 mutant 와
  그것이 만든 실패 출력을 progress.md §E.2 에 남긴 뒤 복원한다.
- `go test ./internal/spec/...` rc 0. **`go test ./...` 는 실행하지 않는다.**
- AC-SLGB-011 은 미해결로 명시한 채 run-phase 를 닫고, sync 이후 CI 로그로 닫는다.
