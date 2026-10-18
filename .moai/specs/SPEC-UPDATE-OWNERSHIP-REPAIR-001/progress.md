# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- plan 산출물 작성 완료 (2026-10-18, manager-spec, card t1865): spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (decision gate on), 이 파일. 상태: `draft` (spec.md frontmatter).
- SPEC ID 사전 점검: `SPEC-UPDATE-OWNERSHIP-REPAIR-001` 정규식 검사 PASS (이번 실행, Bash 원문 출력: `PASS`). 카탈로그 중복 없음 (`ls .moai/specs | grep -c SPEC-UPDATE-OWNERSHIP-REPAIR-001` → `0`, 작성 전).
- 기준 트리: HEAD `cdd72a814` (브랜치 `WT-update-ownership-repair`), t1668 의 `e1e163320` 병합 포함. 기존 시험 11개 GREEN (research.md B).
- 열린 결정: decision-index.md Q2 (배포 버전 기록과 이전 매니페스트 보존, 제품 수준, 운영자 확인 대기 — 자율 Kickoff 차단, 이 수정판도 답하지 않음). Q5 (프로필 쪽 대응물 읽기) 는 리더 판정 2026-10-18 로 보수적 읽기가 확정되었다 (읽지 않음, `not-checked`, 프로필 지도 사본 무조건 보존; 라벨은 증거 대기 유지). Q3, Q4 는 기본값 적용 (2026-10-18T03:07:32Z).
- 다음 단계: plan-auditor 독립 감사 (Tier L 문턱 0.85), 그 뒤 Kickoff. run 단계 필수 독립 감사는 codex 게이트 (카드 완료 조건).

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-18
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (5, Tier L) + decision-index.md (decision gate on) + progress.md

### Plan audit iteration 1 repair (2026-10-18, manager-spec, card t1865)

- 입력: `.moai/reports/t1865/plan-audit-iter1.md` — FAIL, 점수 0.68 (Tier L 문턱 0.85), MP-8 실패, 차단 결함 5건, 필수 백엔드 codex 실패, 감사 대상 SHA `a8ad48c08`.
- 처리: 결함마다 이 트리 (`15db036a3`) 에서 자기 명령으로 확인하거나 반박한 뒤 고쳤다. 측정 원문은 acceptance.md § L (L-1 ~ L-10). 판정 도구는 이 트리에서 `go build -o /private/tmp/own-fixture-bin/moai ./cmd/moai` 로 만든 빌드이다.
- 수정판: spec.md `0.2.0` — 요구사항 16 (REQ-OWR-016 추가), 인수 기준 20 (AC-OWR-20 추가). plan_artifact_hash 는 이 문서가 계산하지 않는다 (lane 이 커밋 뒤 런타임 `ComputeHash` 로 다시 계산한다).

| 결함 | 처분 | 근거와 수정 위치 |
|------|------|-------------------|
| D1 AC-OWR-02 픽스처 경로가 템플릿에 없음 | 고침 (확인됨) | L-2: 종전 두 경로 모두 `git ls-files --error-unmatch` 종료 1. L-1: 실재 경로 `.claude/skills/moai-ref-owasp-checklist/SKILL.md`, `.agents/skills/moai-plan/SKILL.md`. AC-OWR-02 를 실재 경로 + 템플릿 조회 전제 단계 + 승격 변이 대조군(잘못된 `provenanceSource` 주입, design.md D-2) 으로 다시 썼다. AC-OWR-14 의 313개도 내장 원본에서 읽도록 명시했다 (L-5) |
| D2 AC-OWR-08 이 삼켜진 읽기·범위 밖 도우미를 못 잡음 | 고침 (확인됨, 한계 명시) | 0o000 방식 철회. 추가 발견: `claude-skills@` 는 Go 소스에 0건 (L-7) 이라 종전 문자열 스캔은 공허했다. 기준선 대비 패키지 전체 소스 스캔 (L-10 의 19개 파일) + 차분 카나리 + 카드 커밋 이력 기반 파일 불변 (L-6) 으로 교체. 접근 기록 시임은 채택하지 않음 (수용 부채, design.md D-8) |
| D3 RED-now 셀 없음 + 「게이트」 모순 (MP-8) | 고침 (확인됨, 선택지 a) | 판정 시험이 트리에 없어 RED 를 재실행할 수 없음. 20개 모두 regression-guard 로 분류하고 기준마다 승격 지점·승격 규칙 (시험 커밋 A → 셀 기록 커밋 B → 구현 커밋 C) 을 적었다. 선례 SPEC-AC-COLLECTOR-ANCHOR-001. 실측 가능한 전제 셀과 기준선 GREEN 셀 (L-1 ~ L-10) 만 원장에 넣고 RED 셀은 한 칸도 미리 채우지 않았다 |
| D4 AC-OWR-10 이 commands 에 `name` 요구 | 고침 (확인됨) | L-3a: commands 17개 모두 `name:` 없음, `description:` 은 모두 있음 (L-3a 둘째 명령). L-3b: agents 12개 모두 `name:` 있음. 기존 `commands_audit_test.go` 는 `description`·`allowed-tools` 를 요구한다. 조건을 commands/agents 로 분리하고 렌더 기준 비교를 명시했다 |
| D5 `manifest.prior.json` 쓰기 실패 뒤에도 제거 계속 | 고침 (확인됨) | design.md D-3 항목 3 을 fail-closed 로 바꾸고 REQ-OWR-016, AC-OWR-20 (root 에서도 성립하는 디렉터리 충돌 주입) 을 추가했다 |
| D6 Q2 「나」 낙하 목록 불완전 | 고침 (확인됨) | decision-index.md Q2 와 plan.md § G.1 에 접힘·축소·남음 전체 목록. 증거 해석기와 인터페이스를 M1b 에서 M1a 로 옮겨 AC-OWR-05 가 남는다. REQ-OWR-008, AC-OWR-13 (A·C), AC-OWR-18 (c), AC-OWR-17 변이, M5, M8 을 목록에 넣었고 운영자 가시 결과 (현장 증상 (5)(6) 일부 미이행) 를 적었다. Q2 는 열어 둔 채 |
| D7 AC-OWR-16 이 함수만 검증 | 고침 (확인됨) | `update.go:361` 의 dry-run 분기 전체를 기존 헬퍼 `runDryRunUpdateInProject` 로 구동. 디렉터리 mtime 을 트리 해시에 넣어 임시 쓰기를 잡는다 (design.md D-9). REQ-OWR-011 을 명령에 걸었다 |
| D8 `removal-eligible` 정의 불일치 | 고침 (확인됨) | spec.md § B 에 정의 한 곳, design.md D-1 과 AC-OWR-01 이 인용 |
| D9 M6 픽스처가 진단 원인을 재현 못 할 수 있음 | 고침 (확인됨, 한계 명시) | M6 과 AC-OWR-09 에 V1 같은 버전 / V2 낡은 기록 (rc.35 기록 + rc.36 바이너리 모양) 두 변형. 두 변형을 모두 돌린 뒤에만 「재현 안 됨」. V2 를 기존 시임으로 구동 못 하면 셀에 적는다 (갭 G-A) |
| D10 TMPDIR 디렉터리 생성 단계 없음 | 고침 (확인됨) | plan.md § C 의 P0 와 M0 첫 단계 `mkdir -p /private/tmp/own-fixture-d1`. acceptance.md 규칙 2번째 항이 P0 를 가리킨다 |
| D11 AC-OWR-06(b) 에 요구사항 문구 없음 | 고침 (확인됨) | REQ-OWR-015 에 매니페스트 수준 `Version` 갱신과 `DeployedAt` 변경 규칙을 넣었다 |
| D12 `initializer.go:607-608` 한 줄 어긋남 | 반박 (증거) | `sed -n 605,609p internal/core/project/initializer.go` 의 번호 매김으로 607 = `mf.Version = version.GetVersion()`, 608 = `mf.DeployedAt = …`. 이 트리에서 인용이 정확하다. 변경 없음 |
| D13 요구사항 문구의 구현 세부 | 일부 고침 + 수용 부채 | REQ-OWR-008 (코드 이름 제거), REQ-OWR-015 (필드 수준 서술 제거), AC-OWR-01 (상수 개수 단언을 모든 조합 매트릭스로) 고침. REQ-OWR-005 의 파일·디렉터리 이름은 리더 구속 제약 문구라 유지 (decision-index.md 수용 부채) |
| D14 AC-OWR-09 「또는」 결과 | 고침 (확인됨) | 상태마다 관측 결과를 하나로 고정 (S1 ~ S4) |
| D15 병합 충돌 노출 | 수용 부채 + 완화 | 새 판정 코드는 새 파일로, 두 파일의 변경은 호출 줄로 좁히고 M8 의 운영 코드 변경을 없앴다. 잔여 위험은 decision-index.md 수용 부채와 plan.md § D5 |

추가로 발견해 고친 것: (X1) plan.md M4 가 `probeCaseSensitiveFS` 재사용을 지시했으나 design.md D-4 는 모드 프로브 없는 별칭 확정을 정했고 그 함수는 비시험 호출자가 없다 (L-9) — M4 를 design.md 와 맞췄다. (X2) 위 D2 의 문자열 스캔 공허. (X3) 기존 런치 원장 판정이 프로필 설정 디렉터리 항목을 stat 한다 (L-8) — 이 SPEC 이 더하지 않는 기존 접근이며 plan.md OQ-5 로 리더에게 올렸다.

### Plan audit iteration 2 repair (2026-10-18, manager-spec, card t1865)

- 입력: `.moai/reports/t1865/plan-audit-iter2.md` — FAIL, 점수 0.76 (Tier L 문턱 0.85), 차단 결함 N1-N5, 필수 백엔드 codex 실패, 감사 대상 SHA `3c334dfdd`. 이 수정판은 spec.md `0.3.0` — 요구사항 17 (REQ-OWR-017 추가), 인수 기준 21 (AC-OWR-21 추가). 리더 지시 두 건을 같은 커밋에 실었다: OQ-5 의 좁은 예외 (보드 판정, 2026-10-18) 와 Q2 = (가) 조건부 (보드 d-20261018T040918Z-def2; 설계 결정일 뿐, 사용자 보안 승인·실제 홈 접근 승인·Kickoff 사전 승인 아님). plan_artifact_hash 는 이 문서가 계산하지 않는다 (lane 이 커밋 뒤 런타임 `ComputeHash` 로 다시 계산한다).
- 처리 원칙: 결함마다 이 트리 (`5fb8920ee`) 에서 자기 명령으로 확인한 뒤 고쳤고, 스무 개 넘는 기준 전체에 「가장 싼 변이」 점검을 돌려 같은 부류의 약점을 acceptance.md § M 표에 남겼다.

| 결함 | 처분 | 근거와 수정 위치 |
|------|------|-------------------|
| N1 AC-OWR-08 이 `internal/profile`·도우미를 못 봄 | 고침 (확인됨) | `go list -deps ./internal/cli` = 모듈 안 142개, 여섯 핵심 패키지 포함 (§ L-11). `internal/profile` 이 갱신 경로에서 불린다 (`migrate_project_assets.go:185`, `update.go:693-698`, `update_participation.go:49`; § L-12). `GetProfileDir` 는 `profile.go:193` 내보낸 경로 생산 도우미. AC-OWR-08 을 도출한 호출 닫힘 전체, `internal/profile` 의 내보낸 식별자 전부, 상수 접기 뒤 토큰, 비상수 환경 변수 읽기, `internal/profile` 함수 단위 파일시스템 동사의 호출 지점 동결로 재설계. 스캔 변이 대조군 M-a ~ M-f. 접근 기록 시임 없음 → 조각난 경로 + 버림은 못 잡음을 한계 (1)(3) 과 수용 부채에 이름으로 기록. 카나리는 도우미 읽기를 관측한다는 약속을 철회 (atime 은 근거가 못 됨). design.md D-8, acceptance.md AC-OWR-08, decision-index.md 수용 부채 |
| N2 이전 기록 출처값 미검증 | 고침 (확인됨) | `internal/template/deployer.go:265-267` 이 미등록 기존 파일을 `UserCreated` 로 기록 (관측). 출처 상수는 `internal/manifest/types.go:17-35` (§ L-15). REQ-OWR-002 와 spec.md § B 에 템플릿 관리 출처값 정의, AC-OWR-05 변이 (e)(f)(g) 와 운영 원천 경유 (h1)(h2), 이전 세대를 실제 상수로 만들도록 Given 수정. design.md D-2 |
| N3 Backup 단계 → 마이그레이션 인계 미검증 | 고침 (확인됨) | 호출 경로: `update_template_sync.go:815-817`. 패키지 수준 `priorManifestWriter` 시임 (운영 쓰기 함수를 감싼 대역, 목적지에 디렉터리 충돌로 실패) 을 도입하고 AC-OWR-20 을 `runTemplateSyncWithProgress` 전체 흐름 (`--force`) 으로 재설계: 제거 0건, 사본 유지, 이유 `backup-failed`, 배포 계속, 사본 실패를 삼키는 변이, 쌍둥이 대조 (c). REQ-OWR-016 은 파일 바이트 사본 실패까지 포함해 「0건 제거」 로 엄격해짐 (리더 조건 2). design.md D-3 항목 3-5, plan.md M1b |
| N4 AC-OWR-15 전체 트리 동일 불가 | 고침 (확인됨) | 전체 흐름이 `llm.yaml` 을 다시 쓰고 템플릿을 배포한다 (`update_migrate_test.go` 의 `TestMigrationIdempotent` 가 `ReadDeployMode == "local"` 단언). 단언을 복원 영역 R = {제거 파일, 매니페스트} 의 바이트 동일 + 스위치 끈 통제 실행 T0 와의 비교로 계산한 나머지 차이 집합 D ⊆ E 센티널로 재진술. 비정규 시작 매니페스트로 재직렬화 변이를 잡음 |
| N5 버전 일치 건너뛰기 | 고침 (확인됨, 실측) | `update_template_sync.go:1239-1248`. overlay 탐침: 일반 실행 1 `skipped=false`, 일반 실행 2 `skipped=true`, `--force` 실행 `skipped=false`, 이어 일반 실행 `skipped=true` (§ L-13). 메커니즘 선택: 여러 번 실행하는 기준은 `--force` 헬퍼 `runUpdateFlowForced` + 건너뛰는 일반 재실행 사례 분리 + 도달 단언 (규칙 7: `skipped==false`, 새 `manifest.prior.json`, 양수 잔존 행, 쓸어낸 실행 수). 영향: AC-OWR-06, -09, -10, -14, -17, -18, -20, -21. REQ-OWR-012/-013 이 건너뛰기를 명시. 관문 변경은 범위 밖 (OQ-6, 리더에게). design.md D-10 |
| N6 AC-OWR-08(b) 휘발 필드 | 고침 (확인됨) | 정규화 (임시 루트→`<ROOT>`, 백업 디렉터리 이름→`<RUN>`, 타임스탬프·`deployed_at`→`<TS>`) 를 먼저 단언하고 비교. 지워진 영역은 한계로 기록. design.md D-8 L3 |
| N7 Q2 (가) 가 현장에서 바꾸는 것 | 고침 | decision-index.md Q2 (가) 와 plan.md OQ-2 에 「지금 현장 313개에는 아무것도 바꾸지 않는다 (증거는 (가) 이후 배포분에만 생김, 복사 폴백 거울은 등록되지 않음)」 와 (가) 가 바꾸는 것을 명시 (리더 지시의 잔여 진술) |
| N8 REQ-OWR-005 의 OQ-5 예외 | 고침 (확인됨) | REQ-OWR-005 에 금지 동사 (open, read, enumerate, stat, walk, supply, write) 와 좁은 예외 하나 (`launchCandidateIsUsable` 의 `os.Stat`, `profile.go:439` — 카드 이전부터 있어 위반이 아님) 를 넣고, 새 호출 지점·다른 함수·파일·패키지는 예외 아님을 명시. AC-OWR-08 예외는 기준선의 한 행 (함수·`os.Stat`·개수 1), 변이 M-c (같은 함수에 둘째 `os.Stat`) 가 붉다. 리더 판정 (2026-10-18) 을 그대로 따름 |
| N9 낡은 개수 | 고침 | plan.md: 인수 기준 21, 요구사항 17, 시험 파일 16개 (§ F 시험 열을 센 값), AC-OWR-01 ~ -21 |
| N10 AC-OWR-17 분류 불일치 | 고침 | plan.md § G.1 의 남음 열에서 AC-OWR-17 을 뺐다 (축소 열에만) — decision-index.md 와 일치 |
| N11 research.md 버전 | 고침 | research.md 를 `0.3.0` 으로 올리고 기록임을 한 줄 명시 |

리더 조건 (Q2 (가)) 처리: (1) `deployed_version` 은 이번 실행이 배포한 항목에만, 채우기·추정 없음 (REQ-OWR-015, AC-OWR-06) (2) 원본 바이트 → `manifest.prior.json`, 백업 실패 시 0건 제거 (REQ-OWR-016, AC-OWR-20) (3) 원자성: `Save` 유지, 이전 세대도 임시 파일 + 이름 바꾸기, 손상 세대는 증거 없음 (design.md D-3 항목 2-3, AC-OWR-21 (a)(b)) (4) `priorManifestGenerations`=8 은 조회 한도, 기존 `PruneBackups(…, 5, …)` 와의 관계는 design.md D-3 항목 8, 보존 정책 파일 불변 (AC-OWR-21 (e)) (5) 옛 JSON·같은 버전 반복·다운그레이드 (REQ-OWR-017, AC-OWR-21 (c)(d)(f)) (6) 시험 목록은 AC-OWR-06, -20, -21 (7) 삭제·보호 완화·보안 권한·`removalEnabled`·t1702 가 섞인 항목 없음 — 해당 항목은 싣지 않았다.

수정 중 발견해 같이 고친 것: (Y1) 종전 AC-OWR-03/-04 의 「변이」 는 서술뿐이라 시험이 붉음을 관측하지 못했다 — 변이 주입 대조 실행을 넣었다. (Y2) 보존 기준 (-02, -03, -04, -07, -14) 은 제거 암에 닿지 못해도 공허하게 통과했다 — 도달 대조군 (규칙 8) 을 넣었다. (Y3) AC-OWR-16 은 계획이 비거나 후보가 `UserOwned` 가 아니면 빈 집합끼리 같아졌다 — 사전 단언을 넣었다 (design.md D-9). (Y4) AC-OWR-06 의 사본 바이트 단언은 시작 매니페스트가 정규 모양이면 재직렬화 변이를 못 잡았다 — 비정규 모양으로 쓰도록 했다.

잔여 (해결 못 한 것): Q2 는 결정되었으나 Kickoff 는 독립 계획 감사 뒤에 정한다. 접근 기록 시임이 없어 조각난 경로 + 버림 읽기는 감지 불가 (리더 수락 대기, decision-index.md 수용 부채). 일반 같은 버전 재실행의 복구 부재 (OQ-6, 리더에게). 보존 정책 센서스가 `manifest.prior.json` 에 미치는 효과는 미측정 (G-K).

### Ceiling correction commit (2026-10-18, manager-spec, card t1865)

- 입력: `.moai/reports/t1865/plan-audit-iter3.md` — FAIL 0.84 (문턱 0.85), 차단 결함 N12·N13, 필수 백엔드 codex 실패, 감사 대상 SHA `a2a7b8bb2`. 리더 판정 (보드 d-20261018T044408Z-3e9d, 상시 규칙 C): 정확히 한 번의 보정 커밋이 N12·N13 만 고치고, 그 뒤 그 커밋에 한정된 delta 감사 한 번. 이 판은 spec.md `0.3.1`.
- N12 고침: design.md D-3 항목 3 에 관측점 변수 `priorManifestTempWrite`·`priorManifestRename` 을 더했고, AC-OWR-21 (a) 를 정상 실행 (o) 과 세 실패 (p) 임시 쓰기 도중, (q) 임시 쓰기와 이름 바꾸기 사이, (r) 실제 이름 바꾸기 실패로 다시 썼다. 최종 이름에 직접 쓰는 변이는 (o) 의 기록 단언과 (p)(q) 의 오류·최종 이름 부재 단언에서 붉어진다. plan.md M1b 파일 목록·GREEN 과 acceptance.md § M 21행에 같은 관측점을 반영했다.
- N13 고침: plan.md M8 단계 1 을 AC-OWR-15 의 복원 영역 바이트 동일 + 통제 실행 센티널로, M9 단계 1 (a) 를 REQ-OWR-016 에 맞는 제거 0건·`backup-failed` 이유 나열로 바꿨다.
- N14-N21 과 N4 잔여: decision-index.md 에 수용 부채로 id 별 기록 (N16 은 리더 지정 문구). 요구사항·인수 기준 id 집합, 마일스톤 구성, Q2/Q5/OQ 결정은 바뀌지 않았다.
- plan_artifact_hash 는 이 문서가 계산하지 않는다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
