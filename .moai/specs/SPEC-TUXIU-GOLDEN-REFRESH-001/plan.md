# Plan — SPEC-TUXIU-GOLDEN-REFRESH-001

## §1 Approach

세 계층으로 나눠 실행한다. 순서는 결정 가역성(decision-reversibility) 기준 — 변경 가능성이 높은 계약 결정을 앞세우고 기계적 재생성을 뒤로 보낸다.

1. **계약 계층(M1)** — AC-016 데이터 불변 계약에 U-08 예외(제1 항목)와 archive 부분집합 제외(제2 항목, decision-index Q6)를 문서로 편입하고, 접두사 취약한 단정을 원시(raw) 라인 입력의 exact-line 단정으로 타이트닝한다. 타이트닝 단정의 입력에는 교차-기준선 제외가 적용되지 않는다 — 정규화·제외 계층과 단정 계층은 분리되며, 그렇지 않으면 조작 골든이 제외 계층을 통과해 mutant가 생존한다(plan-audit iter-1 D3). 이 계층의 결정(decision-index Q1/Q6)은 두 기준선 설계가 pre-logo 재캡처를, 부활 3 ID의 목록 제외가 기록 archive DATA의 재현을 구조적으로 막기 때문에 강제된 경로이며, 나머지 계층의 의미를 규정한다.
2. **도구 계층(M2)** — 유실된 스크래치패드 하네스를 자기완결형으로 재작성하고 `make tuxiu-golden-capture`로 노출한다. 기본안(decision-index Q2 기본안 적용): 환경변수 게이트형 go test + make 타깃 래퍼(`make tui-snapshot`의 `UPDATE_GOLDEN=1 go test` 선례). tty 변형은 실빌드 바이너리를 강제-트루컬러 pty로 실행해야 한다는 M4 기록의 제약을 하네스가 스스로 충족한다. 바이트 안정성(REQ-TGR-005)을 위해 캡처 시점에 타임스탬프를 정규화하거나 고정한다 — 방식 선택은 run 단계의 구현 재량이며, AC-TGR-004가 결과를 강제한다. 시딩은 decision-index Q6 기본안에 따라 `legacySkillIDs`(`internal/cli/update_archive.go:45-71`)에서 선정한 archive-대상 2 ID를 포함한다 — 기록 구골든의 `3 skills archived`(부활 3 ID)는 재현 불가이며, fresh 요약 리터럴 `total: 2 skills archived`는 M1의 테스트 기대값과 단일 상수로 공유된다.
3. **실행 계층(M3)** — postm4 집합 12파일을 재생성하고 전체 검증을 돌린다. 예상 diff는 update stdout 3파일(literal 접미사 소멸 + Q6 archive DATA 편입)이며, update stderr 3파일 + init 6파일의 바이트 동일성 실측이 decision-index Q3의 미결 증거를 소거한다.

## §2 File inventory (run phase)

| 파일 | 변경 | 비고 |
|------|------|------|
| `internal/cli/tuxiu_characterization_test.go` | 수정 | 헤더 예외 문서(제1·제2 항목, REQ-TGR-001), 교차-기준선 archive 부분집합 제외(REQ-TGR-008), `:201` 원시-라인 exact-line 타이트닝(REQ-TGR-007), pre-logo 3파일 예외 표기(REQ-TGR-003) |
| `Makefile` | 수정 | `tuxiu-golden-capture` 타깃 추가(REQ-TGR-004), `.PHONY` 갱신 |
| 신규 하네스(go test 게이트형 기본안) | 신규 | 위치·형식은 run 단계 결정(decision-index Q2); 시딩 레시피 내장, pty 캡처, 타임스탬프 정규화 |
| `internal/cli/testdata/tuxiu/postm4/*.golden` (12파일) | 재기록 | M3에서 `make tuxiu-golden-capture` 산출물; 내용 변경 예상은 update stdout 3파일 |
| `internal/cli/testdata/tuxiu/*.golden` (pre-logo 12파일) | 무변경 | REQ-TGR-003 — diff 0 검증 |

## §3 Milestones

### M1 — 예외 계약 확정 + 가드 타이트닝 (REQ-TGR-001/003/007/008)

결정 계층. 작업: (1) 테스트 헤더 문서에 예외 제1 항목(U-08 — 제거 전 리터럴, 재판 레코드 SPEC-UPDATE-REPAIR-RC30-001 REQ-RC30-011/AC-RC30-09)과 제2 항목(archive 부분집합 — 근거 레코드 SPEC-UPDATE-LEGACY-SKILL-LIST-001, decision-index Q6) 추가; (2) 교차-기준선 DATA 비교(`tuxDataLines` 경로 한정)에서 archive 부분집합(`✓ archive:` 항목 행 + `total: N skills archived` 요약 행) 제외; (3) `:201`의 `strings.Contains`를 원시(ANSI-stripped) 라인 exact-line 단정으로 교체 — 기대값은 요약 행 `total: 2 skills archived`와 시딩 ID 항목 행 2행(`✓ archive: <id> → .moai/archive/skills/v2.16/<id>`, Q6 시딩 계약 상수)이며, 허용 전처리는 pill-bracket 제거 + trim 같은 외양 정규화로 한정된다(접미사 처리 없음), 단정 입력에는 (2)의 제외가 적용되지 않는다(계층 분리 — 그렇지 않으면 mutant 생존); (4) pre-logo 3파일 예외 표기. 완료 판정: (a) mutant 조작 — postm4 update stdout 3변형 각각의 (i) 접미사 주입과 (ii) 시딩 항목 행 삭제(요약 행 유지) — 각각에 대해 타이트닝 단정 FAIL을 확인하고 주입을 복원한다(M4 음성 제어 선례); (b) 무조작 트리에서 가드 스위트는 예상 RED 상태로 종료한다 — 타이트닝 기대(요약 + 항목 행)가 미재생성 stale 골든(`total: 3 skills archived, 0 user customizations modified` + 부활 3 ID 기록)과 불일치하여 FAIL이며, 교차 비교(AC-TUXIU-016)는 (2)의 제외 편입으로 PASS를 유지한다. 예상 RED의 관측 행렬을 M1 종료 시 기록하고, GREEN 복귀는 M3 재생성에서 일어난다(acceptance.md AC-TGR-005의 순서 의존성 노트와 동일 기록). M1에서 '기존 4테스트 PASS 유지'를 요구하지 않는다 — PASS 유지는 M3 재생성 이후의 상태다.

### M2 — 재생성 하네스 + make 타깃 (REQ-TGR-004/005)

도구 계층. 작업: (1) throwaway 프로젝트 시딩 레시피 내장 — 골든·테스트에 기록된 환경 토큰(`Updated 26 files`, `Created 2 files`, 13 dirs)으로 역산한 시딩 규모 재현 + Q6 기본안에 따른 archive-대상 2 ID 시딩(`legacySkillIDs` 선정; 기록 `3 skills archived`의 부활 3 ID는 재현 불가); (2) 실빌드 바이너리 + pty 캡처(tty 변형), nocolor/notty 변형 캡처; (3) 타임스탬프 정규화/고정으로 바이트 안정성 확보; (4) `make tuxiu-golden-capture` 타깃. 완료 판정: 두 번 연속 실행 → `internal/cli/testdata/tuxiu/` 이하 `git status` 청결(AC-TGR-004). 이 시점에는 아직 골든을 덮어쓰지 않는다(드라이런 또는 임시 출력 대상으로 검증) — M3에서 실제 재기록한다. 덮어쓰기 유무 플래그가 필요하면 run 단계에서 정한다.

### M3 — 재생성 실행 + 전체 검증 (REQ-TGR-002/006)

기계적 계층. 작업: (1) `make tuxiu-golden-capture`로 postm4 12파일 재기록; (2) diff 검증 — 전체-저장소 관측(AC-TGR-006)에서 변경이 허용 목록으로 한정, pre-logo 12파일 무변경, 예상 내용 변경은 update stdout 3파일(literal 접미사 소멸 + Q6 archive DATA 편입); (3) grep 검증 — postm4에서 제거 리터럴 0매치; (4) 4테스트 `-count=1` PASS; (5) update stderr 3파일 + init 6파일의 바이트 동일성 실측 기록(decision-index Q3 증거 소거). 완료 판정: AC-TGR-001/002/005/006/008 GREEN.

## §4 Risks and mitigations

| 위험 | 완화 |
|------|------|
| pty 캡처 재현 난이도(tty truecolor, `tui.downsample` `sync.Once` 실빌드 제약) | M4 기록이 요구 조건을 문서화했다(spec.md §A.3). 하네스가 빌드→pty 실행을 한 명령 안에 수행하도록 자기완결 설계 |
| 시딩 레시피 유실로 환경 재현 불확실 | 골든 자체가 환경 토큰(파일 수·디렉터리 수)을 기록 — 역산 가능. archive DATA는 재현 목표가 아니라 Q6 기본안의 시딩 계약 대상(2 ID)이다. M3에서 update stderr 3 + init 6 파일의 바이트 동일성 실측이 재현 정확도를 검증 |
| 바이트 안정성 미달(정규화 누락 요소 발견) | AC-TGR-004가 강제 — 미달 시 정규화 방식을 보강해 재검증. 두 번째 실행 무변화가 유일한 합격 기준 |
| 카드 실행 중 업스트림 main 흡수로 골든 재차 변질 | M3 재생성은 흡수 이후 최종 트리에서 수행하고, 착지 직전 grep+4테스트를 재측정(landing remeasure discipline) |
| 타이트닝 단정이 pre-logo stale 접미사를 오탈 | 단정 대상을 postm4 집합(현행 기준선)으로 한정하고 pre-logo는 REQ-TGR-008의 정규화 계층이 처리 — AC-TGR-007 mutant 조작은 postm4 골든에만 적용 |

## §5 Verification scope

- 변경 영향 패키지: `internal/cli`(테스트·하네스)와 `Makefile`, `internal/cli/testdata/tuxiu/postm4/` 골든. 프로덕션 Go 코드 무변경이 원칙이며, `go build ./...` 회귀 없음을 확인.
- 4특성화 테스트 + `internal/cli` 패키지 회귀: `go test ./internal/cli -run '^(TestInitUpdateTUXCharacterization|TestTUXChannelPartition|TestTUXDataValuesPreserved|TestTUXPostM4CarriesNewPresentation)$' -count=1` 및 변경 범위 패키지 테스트.
- 스펙 린트: `moai spec lint` — MissingExclusions 포함 0 오류 목표.
- 바이트 안정성·diff 한정 검증은 AC-TGR-004/006 명령 그대로.

## §6 Open questions

- 본 계획서에는 미해결 확인 요청 항목이 없다. 기초 질문 Q1-Q6의 처리는 `decision-index.md` 참조 — Q1/Q4 DECIDED(앵커 기록), Q2/Q5/Q6 FOUNDER(implementation-level, 각 기본안 적용으로 plan close 시 DEFAULT-APPLIED 기입), Q3 EVIDENCE-NEEDED(M2/M3의 실측으로 소거).
- 하네스 유실은 열린 질문이 아니라 기록된 Gap이다(spec.md §A.3): 스크래치패드 내용을 유추해 복원하지 않고, 제약 조건만 이어받아 재작성한다.
