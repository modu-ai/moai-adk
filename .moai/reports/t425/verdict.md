# t425 verdict — doctor 계열 9건 "로컬 적색·CI 미실행" 전제 반증

작업일: 2026-09-02 (lane-11) · 브랜치: WT-doctor-local-red · 기준 트리: origin/develop 9145806d8
카드 원문: "doctor 계열 9건 로컬 적색 — CI가 못 보는 상속 결함 규명·수리" (Class B, 원인 미확정)

## 판정

**전제 반증 3건 + 부수 실측 1건.** 카드의 세 전제(로컬 9건 적색 / CI 미실행 / 귀속 후보 t392 M4)가 모두 관측으로
반증됐고, 유일하게 살아남은 실측 결함(moai doctor 직접 실행의 Harness 5-Layer FAIL)은 go test와 무관한
개발 트리 상태 문제라 별도 소관이다. 권고 처분: **DROPPED (전제 반증)** — 수리할 테스트 적색이 존재하지 않음.

## Claim / Evidence

### C1. "TestRunDoctor_*/TestDoctorCmd_* 9건이 로컬 darwin에서 적색" → 반증

- 측정: `go test ./internal/cli/ -run 'TestRunDoctor|TestDoctorCmd' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 72.916s` (워크트리 9145806d8, 1회차)
- 측정: 동일 명령 `-v` → 27건(TestRunDoctor_* 15 + TestDoctorCmd_* 12) 전부 `--- PASS`, FAIL 0, SKIP 0
- 측정: `go test -C /Users/goos/MoAI/moai-adk-go ./internal/cli/ -run 'TestRunDoctor|TestDoctorCmd' -count=1` → `ok ... 56.453s` (primary checkout main 48239c7dc — origin/develop보다 877 커밋 뒤진 트리에서도 초록)
- 측정: `bin/moai` 존재 상태에서 재실행 → `ok ... 39.514s` (바이너리 유무 무관)
- 측정: 임시 바이너리 삭제 후 최종 재실행 → `ok ... 53.549s`
- **적색 0/4 트리·5회 실행.**

### C2. "CI(ubuntu) 초록 run에서 해당 이름 0회 → 조용히 미실행" → 반증

- 원인 규명: `.github/workflows/ci.yml:208` `go test -json -coverprofile=coverage.out -covermode=atomic ./... > test-stream.json` — **stdout이 파일로 리다이렉트**돼 go test -json의 모든 테스트 이벤트(실행·pass·skip 포함)가 콘솔 로그에 찍히지 않는다. CI 콘솔에서 "이름 0회"는 미실행의 증거가 아니다(관측 부재 ≠ 부재).
- 직접 증명: run 33529227921(fe121203e)의 artifact `test-stream-ci-test-ubuntu-latest`(test-stream.json.gz)를 `gh run download`로 회수 → grep 결과 **TestRunDoctor*/TestDoctorCmd* 이벤트 406건, Action=pass 고유 테스트 29개, Action=skip 0**.
- **CI에서 29건 전부 실행·PASS.** 스킵 계열(t346) 개입 없음.

### C3. lane-1 관측의 실체 — t349 로그 재인용으로 판정

- 카드 텍스트의 실패 사유 `runDoctor error: doctor: 1 check(s) failed (coverage_improvement_test.go:715 등)`과 9건 명단(:715/:737/:777/:4928/:5752/:5802 + doctor_test.go:76 + integration_test.go:176/:202)이 `.moai/reports/t349/doctor-base-attribution.log`(303행)와 **행번호까지 완전 일치**.
- 그 로그의 실패 원인: `✗ Error: no readable binary to judge at .../worktrees/t349/bin/moai (11 committed artifacts to compare)` — **Agent Emit Embed** check가 bin-absent를 FAIL로 처리하던 시점의 기록.
- 그 결함은 이미 수리됨: `f6c027fa0 fix(SPEC-CI-DOCTOR-BIN-001): M1+M2 — bin-absent embed check becomes an informational skip (t346)` (2026-08-28). fe121203e·9145806d8 모두 포함. 현재 트리에서 같은 상황은 `ok Agent Emit Embed "skipped: no readable binary ..."`로 판정됨(직접 실행 실측).
- **즉 lane-1이 재현했다는 9건 적색은 08-28 이전 측정 기록의 재인용이며, 현재 트리에서는 재현 불가(이미 수리된 결함).**

### C4. 귀속 후보 t392 M4(96bfa0c99) → 반증

- `git show 96bfa0c99` : "feat(t392): M4 — pin golden fixtures off the version predicate" (2026-09-01) — **TestDoctorGolden_/TestStatus_ golden 6건의 version-stamp 소관**. "each golden diff is exactly its one version line — doctor ok/warn aggregate counts unchanged" (커밋 메시지 명시).
- TestRunDoctor_*/TestDoctorCmd_*와 무관. golden 집계 수치 의존 문제는 t406 카드가 담당하는 별개 계열.

## 부수 실측 (별도 소관 권고 — 이 카드 범위 밖)

`go build -o /tmp/moai-t425 ./cmd/moai && /tmp/moai-t425 doctor` (워크트리 루트 cwd에서 실행) →
**`fail Harness 5-Layer: L1:FAIL L2:FAIL L3:FAIL L4:FAIL L5:FAIL L6:FAIL`** → `doctor: 1 check(s) failed`, exit 1.

- 원인: 개발 트리에 tracked harness 조각들이 축적돼 `harnessConfigured`가 참이 되어 L1~L6 배터리가 실시한다. 구체: `.moai/harness/{README.md,main.md}` (도입 2e27c14f8, #913), tracked `.claude/skills/hns-*` 10개(예: 002071e55, t259), `.claude/agents/harness/*.md` 6개(f55aefef3). 그러나 baseline은 미구성 — L1 hns-* 10개 triggers 섹션/키 부재, L2 workflow.yaml harness 섹션 부재, L3 CLAUDE.md 마커 0/0, L4 design.md import 부재, L5 baseline 5파일 부족, L6 에이전트 6개 `skills:` 키 부재.
- 성격: doctor의 진단 자체는 정확(트리가 실제로 harness 반구성). 다만 (a) go test는 cwd=internal/cli라 `.moai/harness`를 안 봐서 초록 — 테스트 초록은 이 배터리의 커버 밖, (b) 수리는 tracked 스킬/에이전트 frontmatter 보강 + `.moai/harness` baseline 완성 = 15+ 파일로 5+ 파일 위임 강제 대상이며 운영자/별도 카드 소관.
- 힌트: `runDoctor` 테스트(cwd=internal/cli)가 Harness 배터리를 우연히 피하는 구조 자체는 알아둘 가치가 있다 — internal/cli 하위에 `.moai/harness` 계열 파일이 생기는 날 테스트가 환경 의존 적색이 된다.

## Baseline-attribution

- 모든 측정은 본 run에서 직접 실행·관측함 (lane-11, 2026-09-02).
- 트리: 워크트리 `.claude/worktrees/t425` @ origin/develop **9145806d8** (배차 시점 head와 동일, `git merge-base --is-ancestor` 확인), primary checkout @ main **48239c7dc**.
- CI 원장: run **33529227921** @ **fe121203e**, artifact `test-stream-ci-test-ubuntu-latest` (다운로드: `/tmp/t425-ci/test-stream.json`).
- 참조 로그: `.moai/reports/t349/doctor-base-attribution.log` (primary 측 저장본), `.moai/reports/t346/verdict.md`.

## Gaps (관측하지 못한 것)

- lane-1 세션의 원본 명령·환경·출력: 세션이 이미 소멸해 관측 불가. C3의 재인용 판정은 행번호 완전 일치라는 간접 근거에 의존.
- run 33529227921 이후의 develop CI(9145806d8 head)는 검증하지 않았다 — 9145806d8의 신규 커밋 9건은 전부 t284(audit 계열)이고 doctor 파일 무변경(`git log --stat fe121203e..origin/develop` 실측)이라 판정에 영향 없다고 본다. 필요 시 리드가 9145806d8 head run의 census로 보강 가능.
- go test -json 원장의 skip 이벤트 전수 파싱(406 이벤트 중 skip 표기 테스트 0을 고유 이름 기준으로 확인) 외의 상세 분류는 하지 않았다.

## Residual-risk

- lane-1이 t349 로그와 무관한 실제 재현을 했을 가능성을 100% 배제하지 못한다. 다만 현재 트리에서 워크트리/primary × 바이너리 유무 × 5회 실행 전부 초록이므로, 재현 경로가 존재한다면 그 조건은 이 카드가 기술한 것("순수 origin/develop에서 재현")과는 다른 미기술 조건이다.
- Harness 5-Layer FAIL은 개발 트리에서 계속 재현된다(의도된 진단). `moai doctor`를 트리 루트에서 돌는 스크립트/훅이 있다면 exit 1을 만난다 — 이 리포트의 부수 실측 절을 읽고 별도 카드로 결정할 것.
- CI census(`scripts/ci-census/test-census.sh`)의 콘솔 요약이 run 33529227921에 어떤 행을 찍었는지는 로그에서 재확인하지 않았다(아티팩트 원장이 더 강한 증거라 생략).
