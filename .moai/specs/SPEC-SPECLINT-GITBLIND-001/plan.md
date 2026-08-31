# Plan: SPEC-SPECLINT-GITBLIND-001

Tier M · cycle_type=tdd · 대상 파일 6개(소스 4 + 테스트 2), 예상 300-400 LOC.

> Tier 상향 근거: iter-1 은 Tier S(대상 4파일 미만·AC 8건 이내)로 분류돼 있었다.
> iter-1 의 범위 추가(REQ-SLGB-009)와 결함 수리로 AC 가 11건이 되고 테스트 파일이 더해져
> 파일 수가 5개를 넘는다. Tier S 예산(REQ 8 / AC 8)에 맞추려면 AC 를 합쳐야 하는데,
> 그것은 iter-1 이 지적한 추적성 결함(D4)을 되살리는 방향이다. 예산에 맞추려 합치지 않고 상향한다.
> **결과: plan-auditor PASS 문턱이 0.75 → 0.80 으로 오른다.**

## §A 맥락

카드 t371. `SPEC Lint` CI 잡이 얕은 체크아웃 + `main` ref 부재 상태에서 lint 규칙 두 개를 눈감긴 채
초록을 내고 있다. 측정 근거와 결함 형태는 `spec.md` §1 에 있으며 여기서 반복하지 않는다.

기준 트리: `.claude/worktrees/t371` @ 브랜치 `WT-lint-shallow-clone`, develop `1e5199b88` 기준.
사전 A/B 측정은 develop `b9149857c` 에서 수행됐다.

## §B 알려진 이슈 / 전제

- **`cachedMainBranch()` 는 두 소비자의 공통 병목이다** — `internal/spec/drift.go:68` 의
  `branch: cachedMainBranch` 로 `DetectDrift` 도 이 헬퍼를 쓴다. M2 한 곳을 고치면 두 표면이 함께 움직인다.
  즉 **M2 의 회귀 표면은 lint 하나가 아니다**. 관측 가능한 결과와 그것을 덮는 AC 가 없다는 사실은
  `spec.md` §4 에 잔여 위험으로 기록돼 있다.
- **`StatusGitConsistencyRule` 에는 테스트 seam 이 없다.** `OwnershipTransitionRule` 이 쓰는
  `withFakeOwnershipLookup` 같은 주입점 없이 `getGitImpliedStatus` 를 직접 부른다.
  AC 는 실제 git 픽스처로 검증한다 — `internal/spec/` 에 선례가 있다.
- **per-run 캐시는 패키지 전역이다** — `gitQueryCacheMu` / `gitQueryCacheV`
  (`internal/spec/gitquery_cache.go:21-23`). 이를 건드리는 새 테스트는 **`t.Parallel()` 을 쓰지 않는다**.
  `t.Chdir` 도 프로세스 전역이라 같은 구속에 묶인다.
- **`cachedMainBranch` 는 `.Dir` 을 설정하지 않는다**(`:102`, `:113`) — 해소는 프로세스 작업 디렉터리를 따른다.
  이것은 **시그니처 변경을 요구하지 않는다**: `chdirForTest`(`drift_characterization_test.go:55`)와
  `setupDriftCorpusFixture`(`:98` → `:103`)가 이미 같은 성질을 다루는 확립된 선례다.
  M2 는 호출부를 건드리지 않는다(§F M2 단계 0).
- **`✓ No findings` short-circuit 이 곧 눈감긴 상태의 출력이다** — `printTable` 은 finding 0건이면
  `internal/cli/spec_lint.go:115-118` 에서 끊고 전면 무결을 선언한다(메시지 `:116`).
  M1 이 Info 를 내면 이 줄이 사라진다. 이것이 이 카드의 실제 수리 기제이며,
  AC-SLGB-001 / 004 가 그 부재를 단언한다(`spec.md` §1.2).
- **`--json` / `--sarif` 경로는 검증된 바 없다** — `Finding` 의 JSON 태그(`internal/spec/lint.go:33-45`)는
  의도만 보인다. 관측 표면은 기본 표 출력 하나로 고정하며, 어떤 AC 도 그 경로에 기대지 않는다.
- Info severity 는 `Report.HasErrors` 를 움직이지 않는다(`internal/spec/lint_test.go` 의
  "only info" 케이스가 이미 고정). REQ-SLGB-005 는 새 정책이 아니라 기존 성질의 확인이다.
- **요약 줄은 Info 를 세지 않는다** — `printTable`(`internal/cli/spec_lint.go:113-133`)은 모든 finding 을
  표로 찍지만 요약(`:136-142`)은 error/warning 만 센다. 어떤 AC 도 요약 줄에 기대지 않는다
  (`spec.md` §4 참조). 표에 행이 뜨는 것으로 충분하다.

## §C 사전 점검 (착수 전)

1. `git rev-parse --show-toplevel` → 이 워크트리 경로여야 한다.
2. `git branch --show-current` → `WT-lint-shallow-clone`.
3. `grep -rn "StatusGitUnreachable" internal .github | wc -l` → **0**. 새 code 가 아직 없음을 확인하는 RED 기준선.
4. `go build ./internal/spec/...` → rc 0 (기준선).
5. `go test ./internal/spec/...` → 기준선 기록. 이 패키지만 돈다. **`go test ./...` 는 금지**(CLAUDE.local.md §4/§6).

## §D 제약

- **[HARD] 전체 스위트 로컬 실행 금지.** 검증 범위는 건드린 패키지(`internal/spec`)와 워크플로 파일이다.
- **[HARD] `.github/workflows/ci.yml` 미접촉.**
- **[HARD] `.moai/specs/` 하위의 다른 SPEC frontmatter 미접촉** — `spec.md` §5 의 A/B 는 다른 레인의 파일이다.
- **[HARD] 캐시를 건드리는 테스트는 비병렬** (`t.Parallel()` 금지) — §B 참조.
- Go 코드/주석은 영어. 커밋 메시지는 Conventional Commits + 영어.
- 하드코딩 금지: 브랜치 후보 체인은 리터럴 산재가 아니라 한 곳의 정렬된 목록으로 둔다.

## §E 자가 검증

각 마일스톤 종료 시 `go test ./internal/spec/...` 를 돌리고 출력 전문을 progress.md §E.2 에 남긴다.
M3 종료 시 `.github/workflows/spec-lint.yml` 을 AC-SLGB-009 / 010 의 술어로 직접 읽는다.

**공허한 초록 가드 — 세 겹**

1. 새 테스트는 먼저 RED 를 보인다. 구현 전에 실패하는 것을 관측한 뒤 구현한다.
2. 각 테스트 함수를 `go test -run <name> -v ./internal/spec/...` 로 **직접 지목**해 실행 사실을 확인한다
   (셀렉터 0건 매치 통과 방지).
3. **AC-SLGB-005 와 AC-SLGB-008 은 오늘 트리에서 이미 통과한다** — 전자는 `StatusGitUnreachable` 이
   아직 존재하지 않아 0건이고, 후자는 `mainBranchSet` 이 이미 메모이즈하기 때문이다.
   두 AC 는 반드시 mutation 으로 RED 를 만든 뒤 닫는다. 심을 mutant 와 기대 실패는
   `acceptance.md` 의 해당 AC 에 적혀 있다. 심은 mutant 와 그 실패 출력을 §E.2 에 남기고 복원한다.

## §F 마일스톤

**순서의 성질을 정확히 적는다.**

- **M1 → M2 는 진짜 의존이다.** M2 의 "해소 불가" 결과가 M1 이 만든 Info 경로로 흘러야 한다.
  M1 없이 M2 를 하면 침묵이 사라지는 게 아니라 자리를 옮길 뿐이다.
- **M1 → M3 는 의존이 아니다.** run (3)은 **수정되지 않은 바이너리**로 18건을 냈다.
  즉 M3 만으로도 AC-SLGB-011 은 닫힌다. M3 를 뒤에 두는 것은 의존이 아니라 순서 선호다 —
  M1 이 먼저 있으면 M3 착지 후의 CI 출력이 "돌았다"인지 "여전히 감고 있다"인지 스스로 말해준다.

### M1 — `StatusGitUnreachable` Info finding (가장 되돌리기 어려운 결정)

새 finding 코드와 그 발화 조건을 정하는 일이라, 이후 두 마일스톤과 CI 판독이 전부 여기에 얹힌다.

1. `getGitImpliedStatus` 의 **세 error 모양**을 호출자가 구별할 수 있게 한다
   (sentinel error 또는 판별 가능한 타입 — 문자열 매칭 금지). 세 모양과 반환 지점은
   `spec.md` §2.1 의 표가 정본이다: `:312` ref 해소 실패 / `:316` 매치 0 / `:366` 창 소진.
2. per-run 캐시(`internal/spec/gitquery_cache.go`)에 shallow 술어를 추가한다
   (`git rev-parse --is-shallow-repository`). `cachedMainBranch` 와 동일한 캐시 수명·잠금 규약을 따른다.
3. per-run 캐시에 **발화 여부 플래그**를 추가한다 — `StatusGitUnreachable` 은 실행당 최대 1건
   (`spec.md` §2.2). 캐시 비활성 경로에서는 억제하지 않는다.
4. `StatusGitConsistencyRule.Check` 의 `err != nil → return nil`
   (`internal/spec/lint.go:1324-1327`) 자리를 분기로 바꾼다:
   - shape ① → 발화
   - shape ② 또는 ③ AND shallow → 발화
   - 그 외 → 기존대로 침묵
   - 이미 이번 실행에서 발화했으면 → 침묵
5. 메시지에 시도한 후보 ref 이름과 "저장소 전체에 걸린 조건"이라는 사실을 담는다
   (REQ-SLGB-002 / AC-SLGB-002).
6. **관측 표면 확인**: Info 가 기본 표 출력(`printTable`)에 실제로 나타나고 `✓ No findings` 줄이
   사라지는지 확인한다. `--json` 경로에만 내는 구현은 이 카드의 수리가 아니다(`spec.md` §1.2).

닫는 AC: AC-SLGB-001 / 002 / 003 / 004a / 004b / 005 / 006.

### M2 — `cachedMainBranch` 해소 체인

0. **시그니처는 그대로 둔다 — 결정 완료, 분기 없음.** `cachedMainBranch` 가 `.Dir` 을 설정하지 않는
   성질은 시그니처 변경을 요구하지 않는다. 같은 패키지의 `chdirForTest`
   (`internal/spec/drift_characterization_test.go:55`)가 이미 이 성질을 다루고 있고,
   `setupDriftCorpusFixture`(`:98`)가 마지막에 그것을 호출한다(`:103`). 새 테스트는 그 선례를 재사용한다.
   **M2 의 diff 는 `cachedMainBranch` 본문과 캐시 필드에 국한되며, 호출부는 건드리지 않는다.**
1. 후보를 정렬된 목록 하나로 둔다: `main` → `origin/main` → `master` → `origin/master`.
   각 후보는 `git rev-parse --verify <candidate>` 로 확인한다.
2. 넷 다 실패하면 **해소 불가**를 반환한다. 존재하지 않는 리터럴 `"master"` 를 반환하는 현재 동작을 없앤다.
   해소 불가는 M1 이 만든 Info 경로로 흘러야 하며, 새로운 조용한 실패를 만들지 않는다.
3. 캐시 필드(`mainBranch` / `mainBranchSet`)의 per-run 1회 계산 성질을 보존하고,
   **해소 불가도 캐시**한다(SPEC 마다 4회 spawn 금지).
4. 캐시 비활성 경로(`gitQueryCacheV == nil`, `DetectDrift` 직접 호출)도 동일 체인을 쓴다.

닫는 AC: AC-SLGB-007 / 008.

### M3 — `.github/workflows/spec-lint.yml`

1. `actions/checkout@v7` 단계에 `with: fetch-depth: 0` 을 넣는다.
2. lint 실행 단계 **앞에** `main` ref 를 명시적으로 가져오는 단계를 넣는다.
   중복 가능성을 감수하는 이유는 `spec.md` §2.3 에 있다 — 그 이유를 주석으로 파일에 남긴다.
3. **trigger `paths` 를 넓힌다.** 현재 `pull_request` 와 `push` 둘 다 `['.moai/specs/**']` 단독이라,
   이 SPEC 의 M1/M2(전부 `internal/spec/` Go 변경)가 착지해도 잡이 재실행되지 않는다 —
   그러면 AC-SLGB-011 이 영구히 닫히지 않는다. 두 트리거 모두 `internal/spec/**` 과
   이 워크플로 파일 자체를 덮도록 추가한다.
4. 워크플로가 이 리포의 다른 잡과 액션 버전 정렬을 유지하는지 확인한다(`ci.yml` 과 동일 `@v7`).

닫는 AC: AC-SLGB-009 / 010. AC-SLGB-011 은 착지 후 CI 로그로만 닫힌다 — 로컬에서 PASS 로 표시하지 않는다.

## §G 안티패턴

- **새 조용한 경로를 만드는 것**: M2 의 "해소 불가"가 M1 의 Info 로 이어지지 않으면
  결함을 한 칸 옮긴 것에 불과하다.
- **SPEC 마다 Info 를 찍는 것**: 이 카드가 고치려는 바로 그 CI 상태에서 수백 줄이 나오고
  진짜 신호가 묻힌다. `spec.md` §2.2 가 실행당 1건으로 못박았고 AC-SLGB-003 이 상한을 잰다.
- **완전한 저장소에서 Info 를 남발하는 것**: AC-SLGB-005 가 이 방향의 가드다.
- **terminal status 픽스처로 "0건"을 만족시키는 것**: `Check` 가 git 에 닿기 전에 반환하므로
  규칙이 한 번도 안 돌고 통과한다. AC-SLGB-005 가 비-terminal status 를 [HARD] 제약으로 못박았다.
- **오늘 이미 초록인 AC 를 그대로 닫는 것**: AC-SLGB-005 / 008 은 mutation 없이 닫지 않는다.
- **Info 를 `--json` 경로에만 내는 것**: 표의 `✓ No findings` 거짓 선언이 그대로 남으면 수리가 아니다.
  AC-SLGB-001 / 004 의 줄-부재 단언이 이 방향의 가드다.
- **`go test ./...` 로 확인하려는 것**: 이 리포에서 금지된다. 범위는 `internal/spec`.
- **CI 초록을 AC-SLGB-011 의 근거로 쓰는 것**: 초록은 "규칙이 돌았다"를 뜻하지 않는다.
  로그에서 `StatusGitConsistency` finding 을 **세어야** 한다.
- **`--strict` 를 붉게 만드는 것**: 이 카드의 목표가 아니다. rc 는 0 으로 유지된다(AC-SLGB-006).

## §H 상호참조

- `spec.md` §1 — A/B 측정 3회와 두 축의 건수 귀속
- `spec.md` §2.1 — error 3종 표와 모양별 발화 결정
- `spec.md` §2.2 — 실행당 1건 결정
- `spec.md` §4 — 잔여 위험(특히 `DetectDrift` 동작 변화, 미검증)
- `acceptance.md` — AC-SLGB-001..011 과 REQ→AC 커버리지 표
- `.moai/reports/t371/classification-18.md` — 18건 전수 분류(범위 밖 근거), C-2 가 shape ③ 형태
- `.moai/reports/t371/plan-audit-iter-1.md` — 이 개정이 닫는 결함 목록
- `internal/spec/lint.go:1306-1342` — `StatusGitConsistencyRule.Check`
  (terminal 조기반환 `:1299-1301`, err skip `:1305-1308`, emission `:1310-1319`)
- `internal/spec/drift.go:300-367` — `getGitImpliedStatus`, 세 error 반환 지점 `:312` `:316` `:366`
- `internal/spec/gitquery_cache.go:88-118` — `cachedMainBranch`(캐시 조기반환 `:96-100`)
- `internal/spec/lint_ownership.go:380-400` — 모델로 삼을 `OwnershipTransitionUnreachable` 의 모양
- `internal/cli/spec_lint.go:113` `printTable` — 관측 표면. zero-finding short-circuit `:115-118`(메시지 `:116`),
  요약 줄 `:136-142`(error/warning 만 집계)
- `internal/spec/drift_characterization_test.go:55` `chdirForTest` / `:98` `setupDriftCorpusFixture`(호출 `:103`)
  / `:280` 비-git 디렉터리 진입 선례 — 픽스처 규약의 근거
- `internal/spec/lint.go:33-45` — `Finding` JSON 태그(직렬화 의도만 보임; `--json` 동작 미검증)
