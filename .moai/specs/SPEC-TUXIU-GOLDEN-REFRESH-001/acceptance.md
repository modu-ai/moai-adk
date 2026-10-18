# Acceptance — SPEC-TUXIU-GOLDEN-REFRESH-001

## §1 Requirements → Criteria coverage

| REQ | AC | 비고 |
|-----|----|------|
| REQ-TGR-001 | AC-TGR-008 | 예외 문서화 |
| REQ-TGR-002 | AC-TGR-001 | postm4 리터럴 부재 |
| REQ-TGR-003 | AC-TGR-002, AC-TGR-008 | pre-logo 불변 + 예외 표기 |
| REQ-TGR-004 | AC-TGR-003 | 재생성 명령 존재·실행 |
| REQ-TGR-005 | AC-TGR-004 | 바이트 안정성 |
| REQ-TGR-006 | AC-TGR-005 | 재생성 후 4테스트 PASS |
| REQ-TGR-007 | AC-TGR-007 | mutant 판별(재발 차단) |
| REQ-TGR-008 | AC-TGR-005, AC-TGR-007 | 필터 정규화는 DATA 비교 성립과 mutant 판별로 간접 검증 |

## §2 Criteria

> 측정 환경 표기: 본 문서의 RED 셀 중 "측정됨"은 worktree t1681 @ base `03faddb705`, 2026-10-19, manager-spec(lane-11) 실행 분이다. run 단계는 시작 시점의 트리 SHA로 재고정(re-pin)한다. GREEN 경로 셀은 이를 뒤집는 마일스톤을 명시한다.

### AC-TGR-001 — postm4 골든의 제거 리터럴 부재 (REQ-TGR-002)

- **Given** 재생성이 완료된 작업 트리
- **When** `grep -rl "0 user customizations modified" internal/cli/testdata/tuxiu/postm4/` 실행
- **Then** 출력 없음, exit 1 (매치 0)

RED-now (측정됨, release-blocking): 동일 명령 → `internal/cli/testdata/tuxiu/postm4/update.nocolor.stdout.golden`, `.../update.notty.stdout.golden`, `.../update.tty.stdout.golden` 3행, exit 0. GREEN 경로: M3 재생성이 뒤집는다.

### AC-TGR-002 — pre-logo 골든 12파일 바이트 무변경 (REQ-TGR-003)

- **Given** run 시작 시 고정한 base 트리 SHA
- **When** `git diff <base-sha> -- 'internal/cli/testdata/tuxiu/*.golden' ':!internal/cli/testdata/tuxiu/postm4'` 실행
- **Then** diff 비어 있음 (pre-logo 12파일 불변)

RED-now: 해당성 없음 — 이 기준은 현재 충족 상태이며 회귀 가드(regression-guard)로 분류한다. M1-M3 전 과정에서 유지되어야 한다.

### AC-TGR-003 — 재생성 명령 존재·실행 (REQ-TGR-004)

- **Given** M2 완료 트리
- **When** `make tuxiu-golden-capture` 실행
- **Then** 명령이 자기완결적으로(시딩→빌드→pty 캡처→postm4 기록) 종료 exit 0, 저장소의 testdata 외 경로에 변경 없음

RED-now (측정됨, release-blocking): `grep -c "tuxiu-golden-capture" Makefile` → `0`. GREEN 경로: M2 타깃 추가가 뒤집는다.

### AC-TGR-004 — 재생성 바이트 안정성 (REQ-TGR-005)

- **Given** M2 완료 트리에서 첫 재생성 실행 후 커밋된(또는 staged) postm4 골든
- **When** 재생성 명령을 두 번째로 연속 실행한 뒤 `git status --porcelain -- internal/cli/testdata/tuxiu/` 실행
- **Then** 출력 없음 (두 번째 실행은 바이트 변화 없음)

RED-now: 하네스 부재로 측정 불가 — M2 완료 시점에 최초 측정하며, 이 기준은 M2/M3에서 release-blocking으로 채택한다. 비결정 요소(타임스탬프 등)가 남아 두 번째 실행이 변화를 만들면 FAIL이 정답이다.

### AC-TGR-005 — 재생성 후 특성화 가드 통과 (REQ-TGR-006, REQ-TGR-008)

- **Given** M3 완료 트리 (재생성된 postm4 + M1의 필터·단정 변경)
- **When** `go test ./internal/cli -run '^(TestInitUpdateTUXCharacterization|TestTUXChannelPartition|TestTUXDataValuesPreserved|TestTUXPostM4CarriesNewPresentation)$' -count=1` 실행
- **Then** `ok github.com/modu-ai/moai-adk/internal/cli` (4테스트 PASS — pre-logo/postm4 DATA 비교가 예외 정규화 하에서 성립)

RED-now: GREEN-now (측정됨 — base에서 `ok ... 1.252s`, 앵커된 셀렉터, sweep 19행). 현재 통과는 stale 골든 상태의 공신력 없는 통과이므로, 이 기준의 의미는 "재생성 후에도 통과"라는 회귀 가드다. M1 단정 타이트닝 직후·재생성 직전에는 stale 골든 + 타이트닝 단정 조합에서 일시 FAIL이 예상되며, M3 재생성이 이를 뒤집는다(순서 의존성 — M1 → M2 → M3).

### AC-TGR-006 — 재생성 diff의 범위 한정 (REQ-TGR-004, REQ-TGR-002)

- **Given** M3 실행 직전 트리와 실행 후 트리
- **When** `git diff --name-only -- internal/cli/testdata/tuxiu/` 실행
- **Then** 변경 파일이 모두 `internal/cli/testdata/tuxiu/postm4/` 이하이며, pre-logo 파일과 프로덕션 코드는 포함되지 않는다. 추가로 나머지 9파일(update stderr 3 + init 6)이 바이트 동일하게 유지되는지 실측 값을 기록한다(decision-index Q3 증거).

RED-now: M3 이전에는 관측 자체가 없다 — M3에서 최초 관측하며 release-blocking으로 채택한다.

### AC-TGR-007 — mutant 판별: 조작 골든에 대한 단정 FAIL (REQ-TGR-007)

- **Given** postm4 골든 하나(`update.notty.stdout.golden`)에 제거된 리터럴을 주입한 임시 상태 (M4 음성 제어 선례)
- **When** 4테스트 실행
- **Then** 타이트닝된 exact-line 단정이 주입 골든에서 FAIL을 보고한다
- **When** 주입을 복원하면
- **Then** PASS로 복귀

RED-now (구조적, release-blocking): 현재 접두사 `strings.Contains` 단정은 주입을 판별하지 못한다 — stale 접미사 포함 골든에서도 통과한다(§A.1의 vacuous pass 기제). GREEN 경로: M1 타이트닝이 뒤집는다. 채택 전 mutant probe(verification-completeness §2): "제거 리터럴을 지닌 골든도 통과하는 mutant"가 현재 상태 그 자체이므로, 타이트닝은 이 mutant를 배제하는 방향이어야 한다.

### AC-TGR-008 — 예외 문서의 근거 레코드 인용 (REQ-TGR-001, REQ-TGR-003)

- **Given** M1 완료 트리
- **When** `internal/cli/tuxiu_characterization_test.go` 헤더 문서를 읽는다
- **Then** (a) 제거 전 리터럴 `0 user customizations modified`, (b) 재판 레코드(SPEC-UPDATE-REPAIR-RC30-001 REQ-RC30-011 / AC-RC30-09), (c) stale 접미사를 유지하는 pre-logo 파일 3개의 파일명이 기록되어 있다

RED-now (구조적): 현행 헤더(:1-33)에는 U-08 예외 기록이 없다(실측: 헤더 읽음). GREEN 경로: M1 (1)이 뒤집는다.

## §3 Quality gates and Definition of Done

- TRUST 5: Tested(4테스트 + mutant 판별 + 패키지 회귀), Readable/Unified(프로젝트 포매터·린터 통과), Secured(해당 없음 — 테스트 인프라만), Trackable(Conventional Commits, `feat(SPEC-TUXIU-GOLDEN-REFRESH-001): M{N} ...`).
- 스펙 린트 0 오류, `go build ./...` 성공, `GOOS=windows GOARCH=amd64 go build ./...` 성공(하네스가 셸/pty 의존이면 빌드 태그·플랫폼 분기 검토 — manager-develop Section B1).
- 이 SPEC의 완료는 AC-TGR-001/003/004/005/006/007/008 GREEN + AC-TGR-002 유지를 요구한다.
