# Plan — SPEC-TUXIU-GOLDEN-REFRESH-001

## §1 Approach

세 계층으로 나눠 실행한다. 순서는 결정 가역성(decision-reversibility) 기준 — 변경 가능성이 높은 계약 결정을 앞세우고 기계적 재생성을 뒤로 보낸다.

1. **계약 계층(M1)** — AC-016 데이터 불변 계약에 U-08 예외를 문서로 편입하고, DATA 필터의 정규화 계층에 예외 처리를 두며, 접두사 취약한 단정을 정확-라인 단정으로 타이트닝한다. 이 계층의 결정(decision-index Q1)은 두 기준선 설계가 pre-logo 재캡처를 구조적으로 막기 때문에 사실상 강제된 경로이며, 나머지 계층의 의미를 규정한다.
2. **도구 계층(M2)** — 유실된 스크래치패드 하네스를 자기완결형으로 재작성하고 `make tuxiu-golden-capture`로 노출한다. 기본안(decision-index Q2 기본안 적용): 환경변수 게이트형 go test + make 타깃 래퍼(`make tui-snapshot`의 `UPDATE_GOLDEN=1 go test` 선례). tty 변형은 실빌드 바이너리를 강제-트루컬러 pty로 실행해야 한다는 M4 기록의 제약을 하네스가 스스로 충족한다. 바이트 안정성(REQ-TGR-005)을 위해 캡처 시점에 타임스탬프를 정규화하거나 고정한다 — 방식 선택은 run 단계의 구현 재량이며, AC-TGR-004가 결과를 강제한다.
3. **실행 계층(M3)** — postm4 집합 12파일을 재생성하고 전체 검증을 돌린다. 예상 diff는 update stdout 3파일이며, 나머지 9파일의 바이트 동일성 실측이 decision-index Q3의 미결 증거를 소거한다.

## §2 File inventory (run phase)

| 파일 | 변경 | 비고 |
|------|------|------|
| `internal/cli/tuxiu_characterization_test.go` | 수정 | 헤더 예외 문서(REQ-TGR-001), DATA 필터 접미사 정규화(REQ-TGR-008), `:201` 단정 타이트닝(REQ-TGR-007), pre-logo 3파일 예외 표기(REQ-TGR-003) |
| `Makefile` | 수정 | `tuxiu-golden-capture` 타깃 추가(REQ-TGR-004), `.PHONY` 갱신 |
| 신규 하네스(go test 게이트형 기본안) | 신규 | 위치·형식은 run 단계 결정(decision-index Q2); 시딩 레시피 내장, pty 캡처, 타임스탬프 정규화 |
| `internal/cli/testdata/tuxiu/postm4/*.golden` (12파일) | 재기록 | M3에서 `make tuxiu-golden-capture` 산출물; 내용 변경 예상은 update stdout 3파일 |
| `internal/cli/testdata/tuxiu/*.golden` (pre-logo 12파일) | 무변경 | REQ-TGR-003 — diff 0 검증 |

## §3 Milestones

### M1 — 예외 계약 확정 + 가드 타이트닝 (REQ-TGR-001/003/007/008)

결정 계층. 작업: (1) 테스트 헤더 문서에 U-08 예외 항목 추가 — 제거 전 리터럴, 재판 레코드(SPEC-UPDATE-REPAIR-RC30-001 REQ-RC30-011/AC-RC30-09), 적용 라인 식별; (2) `tuxNormLine`/필터 계층에 요약 라인 접미사 정규화 추가; (3) `:201`의 `strings.Contains`를 정규화된 라인에 대한 exact-line 단정으로 교체; (4) pre-logo 3파일 예외 표기. 완료 판정: 기존 4테스트 `-count=1` PASS 유지 + AC-TGR-007 mutant 조작(임시 골든에 제거 리터럴 주입)에 대한 FAIL 확인 후 복원 — M4의 음성 제어(negative control) 선례를 따른다.

### M2 — 재생성 하네스 + make 타깃 (REQ-TGR-004/005)

도구 계층. 작업: (1) throwaway 프로젝트 시딩 레시피 내장 — 골든·테스트에 기록된 환경 토큰(`Updated 26 files`, `3 skills archived`, `Created 2 files`, 13 dirs)으로 역산한 시딩 규모 재현; (2) 실빌드 바이너리 + pty 캡처(tty 변형), nocolor/notty 변형 캡처; (3) 타임스탬프 정규화/고정으로 바이트 안정성 확보; (4) `make tuxiu-golden-capture` 타깃. 완료 판정: 두 번 연속 실행 → `internal/cli/testdata/tuxiu/` 이하 `git status` 청결(AC-TGR-004). 이 시점에는 아직 골든을 덮어쓰지 않는다(드라이런 또는 임시 출력 대상으로 검증) — M3에서 실제 재기록한다. 덮어쓰기 유무 플래그가 필요하면 run 단계에서 정한다.

### M3 — 재생성 실행 + 전체 검증 (REQ-TGR-002/006)

기계적 계층. 작업: (1) `make tuxiu-golden-capture`로 postm4 12파일 재기록; (2) diff 검증 — pre-logo 12파일 무변경, 변경은 `postm4/` 이하로 한정, 예상 내용 변경은 update stdout 3파일; (3) grep 검증 — postm4에서 제거 리터럴 0매치; (4) 4테스트 `-count=1` PASS; (5) 나머지 9파일 바이트 동일성 실측 기록(decision-index Q3 증거 소거). 완료 판정: AC-TGR-001/002/005/006/008 GREEN.

## §4 Risks and mitigations

| 위험 | 완화 |
|------|------|
| pty 캡처 재현 난이도(tty truecolor, `tui.downsample` `sync.Once` 실빌드 제약) | M4 기록이 요구 조건을 문서화했다(spec.md §A.3). 하네스가 빌드→pty 실행을 한 명령 안에 수행하도록 자기완결 설계 |
| 시딩 레시피 유실로 환경 재현 불확실 | 골든 자체가 환경 토큰(파일 수·스킬 수·디렉터리 수)을 기록 — 역산 가능. M3에서 9파일 바이트 동일성 실측이 재현 정확도를 검증 |
| 바이트 안정성 미달(정규화 누락 요소 발견) | AC-TGR-004가 강제 — 미달 시 정규화 방식을 보강해 재검증. 두 번째 실행 무변화가 유일한 합격 기준 |
| 카드 실행 중 업스트림 main 흡수로 골든 재차 변질 | M3 재생성은 흡수 이후 최종 트리에서 수행하고, 착지 직전 grep+4테스트를 재측정(landing remeasure discipline) |
| 타이트닝 단정이 pre-logo stale 접미사를 오탈 | 단정 대상을 postm4 집합(현행 기준선)으로 한정하고 pre-logo는 REQ-TGR-008의 정규화 계층이 처리 — AC-TGR-007 mutant 조작은 postm4 골든에만 적용 |

## §5 Verification scope

- 변경 영향 패키지: `internal/cli`(테스트·하네스)와 `Makefile`, `internal/cli/testdata/tuxiu/postm4/` 골든. 프로덕션 Go 코드 무변경이 원칙이며, `go build ./...` 회귀 없음을 확인.
- 4특성화 테스트 + `internal/cli` 패키지 회귀: `go test ./internal/cli -run '^(TestInitUpdateTUXCharacterization|TestTUXChannelPartition|TestTUXDataValuesPreserved|TestTUXPostM4CarriesNewPresentation)$' -count=1` 및 변경 범위 패키지 테스트.
- 스펙 린트: `moai spec lint` — MissingExclusions 포함 0 오류 목표.
- 바이트 안정성·diff 한정 검증은 AC-TGR-004/006 명령 그대로.

## §6 Open questions

- 본 계획서에는 미해결 확인 요청 항목이 없다. 기초 질문 Q1-Q5의 처리는 `decision-index.md` 참조 — Q1/Q4 DECIDED(앵커 기록), Q2/Q5 FOUNDER(implementation-level, 기본안 적용으로 plan close 시 DEFAULT-APPLIED 기입), Q3 EVIDENCE-NEEDED(M2/M3의 실측으로 소거).
- 하네스 유실은 열린 질문이 아니라 기록된 Gap이다(spec.md §A.3): 스크래치패드 내용을 유추해 복원하지 않고, 제약 조건만 이어받아 재작성한다.
