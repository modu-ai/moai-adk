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

## Binding run conditions

Copied by the plan→run Kickoff transition (lane-31, 2026-10-18T04:53Z) from `.moai/reports/t1865/plan-audit-iter4-delta.md` (verdict PASS-WITH-DEBT, audited_sha 073b3e1382b80f67e5ca9ac1f11ea4e1bab196e4, codex receipt rcpt-711bb9d54319e0e61892f763). Each is a run-phase condition the sync audit re-reads; an undisposed one caps the sync verdict at FAIL.

- debt: N22 dispose_in=run AC-OWR-21 When clause omits configuration (o)
- debt: N23 dispose_in=run plan.md M8 omits AC-OWR-15 (v) reach assertion
- debt: N24 dispose_in=run mutant M-F (hooks then direct rewrite of the final name) undetected by AC-OWR-21 (a), disclosed limit
- debt: N4-remainder dispose_in=run whole-tree equality withdrawn; the D subset-of-E sentinel cannot see a difference inside E or file modes
- debt: N14 dispose_in=run no precedence among prior manifest generations that disagree (design.md D-2, REQ-OWR-002)
- debt: N15 dispose_in=run attested managed path also accepts a DeployedHash match, looser than the registered path
- debt: N16 dispose_in=run profile-root read through a run-time-assembled path escapes AC-OWR-08; card t1702 hold scope, not bypassed
- debt: N17 dispose_in=run AC-OWR-17/-18 say "tree hash" without stating the compared hash is content-only
- debt: N18 dispose_in=run AC-OWR-20 (a)(b) told to satisfy rule 7 whose condition (ii) needs a new manifest.prior.json
- debt: N19 dispose_in=run no run directory when there is no .moai/config to back up; location of manifest.prior.json unstated
- debt: N20 dispose_in=run skipped UserCreated entry not asserted to carry no deployed_version
- debt: N21 dispose_in=run wording in REQ-OWR-005 and AC-OWR-20 (b) removalBackupWriter stand-in not defined by D-3 item 5

decision record: decided_by=lane-31 (card orchestrator, plan→run Kickoff, autonomous per auto-semantics §9.1) evidence_refs=.moai/reports/t1865/plan-audit-iter4-delta.md PASS-WITH-DEBT 0.92 audited 073b3e1382b80f67e5ca9ac1f11ea4e1bab196e4, codex pass rcpt-711bb9d54319e0e61892f763, plan_artifact_hash bd1f442c1bb2570a9f0747a875f0bae19de4374b3fc480d0a830d57f3fceeb2d unchanged at HEAD 508b32659c; Q2 (ga) conditional board d-20261018T040918Z-def2; leader rulings d-20261018T044408Z-3e9d (rule C delta round) ladder_path=gate §9.1 (factory stage plan-audit v12, kickoff v13, decide --decider audit → run v14)

## §E.2 Run-phase Evidence

기록 규약: 각 셀은 명령, 원문 출력, 종료 코드, 트리 SHA 를 갖는다. 출력 중 시각 잡음 한 줄 (`INFO SUNSET_CONFIG_DORMANT_NOTICE`) 은 시험이 아닌 설정 알림이며 판정과 무관하다. 모든 시험은 `TMPDIR=/private/tmp/own-fixture-d1` 아래 임시 디렉터리와 임시 HOME 만 썼다. 판정 도구는 이 트리의 `go test` (Go 1.26.8, darwin/arm64) 이며 바이너리 교체는 없다.

### M0 — 기준선과 공유 픽스처 (run chunk 1)

**P0 (사전 점검)**: `mkdir -p /private/tmp/own-fixture-d1` 실행 뒤 `ls -ld /private/tmp/own-fixture-d1` → `drwxr-xr-x@ 3 goos  wheel  96 Oct 18 13:11 /private/tmp/own-fixture-d1`, 종료 코드 0. 시작 트리 `785f2b6c48`, 브랜치 `WT-update-ownership-repair`, `git status --short` 깨끗.

**M0 커밋**: `e902867dea` (`internal/cli/migrate_ownership_helpers_test.go`, `internal/cli/testdata/profile_callsite_baseline.txt` 만 추가, 운영 코드 변경 없음). 기준선 데이터는 운영 코드가 손대지 않은 트리 `785f2b6c48` 에서 측정해 얼렸다 (`go test ./internal/cli/ -run '^TestProfileCallsiteBaselineMatchesTree$' -update-profile-baseline`, 156 행 / 스캔한 파일 1642개; `go list -deps` 닫힘은 GOOS darwin·linux·windows 합집합). 예외 행은 정확히 한 줄: `internal/profile/profile.go  fs:launchCandidateIsUsable:os.Stat  1  exception`. 마이그레이션·미리보기·갱신 동기화 파일의 행은 기존 `migrate_project_assets.go  profile.ResolveLaunchProfileForProject  1` 한 줄뿐이다.

**기준선 시험 11개 (GREEN, 회귀 가드 — 트리 `e902867dea`)**

명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestMigrationClassifiesUnregisteredMirrorCopy|TestMigrationKeepsSharedCopyUnderProfileReach|TestMigrationRemovesSharedCopyWithoutProfileReach|TestEmptiedSkillDirRemoved|TestMigrationBackupBeforeRemoval|TestMigrationDefaultKeepsConfirmedProjectCopy|TestMigrationRemovalGate|TestMigrationIdempotent|TestMigrationPreservesExistingMirrorEntries|TestRF2F3b_MigrationPreservesUntracked|TestPreviewOmitsRemovalCandidateWhileSwitchOff)$' -count=1 -v` — 쓸어낸 시험 수 11 (`--- PASS` 11줄), 종료 코드 0.

```text
=== RUN   TestMigrationClassifiesUnregisteredMirrorCopy
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (1.68s)
=== RUN   TestMigrationKeepsSharedCopyUnderProfileReach
--- PASS: TestMigrationKeepsSharedCopyUnderProfileReach (1.36s)
=== RUN   TestMigrationRemovesSharedCopyWithoutProfileReach
--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach (1.27s)
=== RUN   TestMigrationBackupBeforeRemoval
2026/10/18 14:14:34 INFO SUNSET_CONFIG_DORMANT_NOTICE spec="SPEC-V3R2-MIG-003 REQ-018" yaml_path=.moai/config/sections/sunset.yaml advice="SunsetConfig has no runtime hot path; activation requires a new SPEC"
--- PASS: TestMigrationBackupBeforeRemoval (1.67s)
=== RUN   TestEmptiedSkillDirRemoved
=== RUN   TestEmptiedSkillDirRemoved/emptied_dir_removed_with_its_last_file
=== RUN   TestEmptiedSkillDirRemoved/dir_with_user_content_stays
--- PASS: TestEmptiedSkillDirRemoved (2.56s)
    --- PASS: TestEmptiedSkillDirRemoved/emptied_dir_removed_with_its_last_file (1.63s)
    --- PASS: TestEmptiedSkillDirRemoved/dir_with_user_content_stays (0.93s)
=== RUN   TestMigrationDefaultKeepsConfirmedProjectCopy
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (0.92s)
=== RUN   TestMigrationRemovalGate
--- PASS: TestMigrationRemovalGate (0.94s)
=== RUN   TestRF2F3b_MigrationPreservesUntracked
--- PASS: TestRF2F3b_MigrationPreservesUntracked (0.00s)
=== RUN   TestMigrationIdempotent
moai-cli-test: userHomeDirFn redirected real home to sandbox /private/tmp/own-fixture-d1/moai-cli-home-3022242697
--- PASS: TestMigrationIdempotent (0.80s)
=== RUN   TestMigrationPreservesExistingMirrorEntries
--- PASS: TestMigrationPreservesExistingMirrorEntries (1.14s)
=== RUN   TestPreviewOmitsRemovalCandidateWhileSwitchOff
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	13.783s
exit=0
```

**M0 도우미 시험 (특성화, 도착 시 GREEN — 트리 `e902867dea`)**

명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestOwnershipHelperTreeSnapshotSeesDirMtime|TestOwnershipHelperForcedRunReachesReconciliation|TestOwnershipHelperUsesRealEmbeddedPaths|TestProfileCallsiteBaselineMatchesTree)$' -count=1 -v` — 종료 코드 0. 둘째 줄 묶음이 건너뛰기 관문 재측정 (acceptance.md § L-13 의 이 실행 재기록): 일반 1 `skipped=false`, 일반 2 `skipped=true`, `--force` `skipped=false`, 이어 일반 `skipped=true`.

```text
=== RUN   TestOwnershipHelperTreeSnapshotSeesDirMtime
--- PASS: TestOwnershipHelperTreeSnapshotSeesDirMtime (0.01s)
=== RUN   TestOwnershipHelperForcedRunReachesReconciliation
2026/10/18 14:14:50 INFO SUNSET_CONFIG_DORMANT_NOTICE spec="SPEC-V3R2-MIG-003 REQ-018" yaml_path=.moai/config/sections/sunset.yaml advice="SunsetConfig has no runtime hot path; activation requires a new SPEC"
    migrate_ownership_helpers_test.go:979: force=false run=1 skipped=false
    migrate_ownership_helpers_test.go:984: force=false run=2 skipped=true
    migrate_ownership_helpers_test.go:989: force=true  run=3 skipped=false
    migrate_ownership_helpers_test.go:994: force=false run=4 skipped=true
--- PASS: TestOwnershipHelperForcedRunReachesReconciliation (3.49s)
=== RUN   TestOwnershipHelperUsesRealEmbeddedPaths
--- PASS: TestOwnershipHelperUsesRealEmbeddedPaths (1.47s)
=== RUN   TestProfileCallsiteBaselineMatchesTree
--- PASS: TestProfileCallsiteBaselineMatchesTree (3.07s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	9.425s
exit=0
```

**M6 열린 질문 탐침 (commands 이전 제외 ↔ 재배포 순환이 재현되는가) — 결과: 재현되지 않음 (두 변형 모두)**

탐침은 던져 버리는 시험 파일 (`zz_m6_probe_test.go`, 커밋하지 않음, 삭제함) 이었고 판정 시험이 아니다. 같은 프로젝트 모양을 두 변형으로 `runTemplateSyncWithProgress` (일반 실행, `--force` 없음) 에 먹여 관측만 했다. 픽스처 네 상태: S1 `.claude/commands/moai/gtd.md` (등록 managed, 건강), S2 `.claude/commands/moai/fix.md` (등록 managed 로 순수 템플릿 바이트를 기록한 뒤 사용자가 `USERLINE: keep me` 한 줄을 덧붙임), S3 `.claude/commands/moai/todo.md` (템플릿과 바이트 같고 항목 없음), S4 `.claude/commands/moai-user-cmd.md` (사용자 작성). 변형 V1 같은 버전 (`system.yaml` 없음 → 프로젝트 버전 `0.0.0`, 바이너리와 다름), V2 낡은 기록 (`system.yaml` `template_version: v3.0.0`, 매니페스트 `Version: v3.0.0`, `TemplateTreeHash: stale-tree-hash`; 구동은 `runUpdate` 가 아니라 같은 관문 `versionMatchSkipsSync` 를 거치는 `runTemplateSyncWithProgress`). 원문 (관측 줄만 추림, 색상 제어 문자와 `total: 0 skills archived` 줄은 뺌, 트리 `785f2b6c48` + M0 파일):

```text
M6PROBE V1-same-version skipped=false
M6PROBE V1-same-version template-commands=17 missing=0
M6PROBE V1-same-version USERLINE found in .claude/commands/moai/fix.md
M6PROBE V1-same-version OUTLINE: Reconciliation: 1 refreshed, 0 merged, 1 conflict(s), 3 preserved, 0 archived-removed
M6PROBE V1-same-version OUTLINE:   conflict: .claude/commands/moai/fix.md (sidecar: .claude/commands/moai/fix.md.moai-new)
M6PROBE V1-same-version S2 fix.md has USERLINE=true
M6PROBE V1-same-version S2 sidecar fix.md.moai-new present
M6PROBE V1-same-version S1 gtd equals template=true
M6PROBE V1-same-version S3 todo equals template=true
M6PROBE V1-same-version S4 foreign unchanged=true
M6PROBE V1-same-version out-lines-mentioning .claude/commands/moai/gtd.md = 0
M6PROBE V1-same-version out-lines-mentioning .claude/commands/moai/fix.md = 1
M6PROBE V1-same-version out-lines-mentioning .claude/commands/moai/todo.md = 0
M6PROBE V1-same-version out-lines-mentioning .claude/commands/moai-user-cmd.md = 0
M6PROBE V2-stale-record skipped=false
M6PROBE V2-stale-record template-commands=17 missing=0
M6PROBE V2-stale-record USERLINE found in .claude/commands/moai/fix.md
M6PROBE V2-stale-record OUTLINE: Reconciliation: 1 refreshed, 0 merged, 1 conflict(s), 3 preserved, 0 archived-removed
M6PROBE V2-stale-record OUTLINE:   conflict: .claude/commands/moai/fix.md (sidecar: .claude/commands/moai/fix.md.moai-new)
M6PROBE V2-stale-record S2 fix.md has USERLINE=true
M6PROBE V2-stale-record S2 sidecar fix.md.moai-new present
M6PROBE V2-stale-record S1 gtd equals template=true
M6PROBE V2-stale-record S3 todo equals template=true
M6PROBE V2-stale-record S4 foreign unchanged=true
M6PROBE V2-stale-record out-lines-mentioning .claude/commands/moai/gtd.md = 0
M6PROBE V2-stale-record out-lines-mentioning .claude/commands/moai/fix.md = 1
M6PROBE V2-stale-record out-lines-mentioning .claude/commands/moai/todo.md = 0
M6PROBE V2-stale-record out-lines-mentioning .claude/commands/moai-user-cmd.md = 0
```

해석: 두 변형 모두 템플릿이 운반하는 commands 17개가 전부 존재하고 (`missing=0`), S1 은 템플릿 바이트, S2 는 사용자 줄 보존 + 충돌 사이드카 하나 (보고에 한 번만 나타남), S3 은 템플릿 바이트, S4 는 불변이다. 어떤 경로도 두 처분으로 보고되지 않았다. 따라서 이 트리에서 현장의 「commands 가 이전 대상에서 빠졌다가 재배포됨」 순환은 두 변형으로 재현되지 않았다. M6 의 AC-OWR-09 는 회귀 가드로 남을 가능성이 높고 (결정은 M6 chunk), V2 구동의 한계 — `runUpdate` 진입이 아니라 `runTemplateSyncWithProgress` 를 기존 시임으로 구동했고 낡은 기록은 `system.yaml`·매니페스트 `Version`·`TemplateTreeHash` 세 곳으로 모사했다 — 를 갭으로 남긴다. 참고 관측 (판정 아님): 탐침의 첫 시도는 S2 의 매니페스트에 편집 뒤 바이트의 해시를 배포 해시로 기록하는 잘못된 픽스처였고, 그 모양에서는 사용자 줄이 사이드카·백업 없이 「refreshed」 로 덮였다. 현실에서는 편집이 기록보다 나중이므로 (`ownAddRegisteredEdited` 가 그 순서로 만든다) 이 모양은 정상 경로가 아니지만, 매니페스트가 편집 뒤 상태를 배포 해시로 갖는 항목은 템플릿 소유로 읽혀 덮인다는 사실은 M6 chunk 가 알아 둘 만하다.

### M1a — 판정 함수, 이유 코드, 증거 해석기 (AC-OWR-01 ~ -05, run chunk 1)

**커밋 A (시험 파일만)**: `9b73a35938` — `internal/cli/migrate_ownership_test.go` (AC-OWR-01 ~ -04), `internal/cli/migrate_ownership_evidence_test.go` (AC-OWR-05 (a)-(g), 함수 수준). 이 트리에서는 판정 코드가 없어 `internal/cli` 시험 패키지 전체가 빌드되지 않는다 — 그래서 아래 다섯 셀은 모두 빌드 실패 RED 이고, 같은 컴파일러 첫 열 줄이 되풀이된다 (Go 는 열 개 뒤에서 `too many errors` 로 자른다). 이 시점에는 기준선 11개도 같은 이유로 돌릴 수 없다 (빌드가 풀린 C 커밋에서 다시 돌린다). 어떤 기준도 「올바른 이유로 붉은가」 를 위 이유 문장으로 답했다. 빌드 실패 RED 는 변이 관측 (시험이 변이 구현 앞에서 붉음을 직접 보는 단언) 을 이 시점에는 보여주지 못한다 — 그 관측은 C 커밋 GREEN 원문의 `MutantIsCaught` 서브 시험이 한다.

**AC-OWR-01 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipVerdictClasses$' -count=1 -v`; 트리 SHA: `9b73a359380e244a931e14ee10a6950e6d26c07e`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_test.go:30:36: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:32:41: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:35:74: undefined: provenanceSource
internal/cli/migrate_ownership_test.go:43:43: undefined: ownershipEnv
internal/cli/migrate_ownership_test.go:43:73: undefined: ownershipVerdict
internal/cli/migrate_ownership_test.go:51:58: undefined: remainingRow
internal/cli/migrate_ownership_test.go:64:10: undefined: ownershipClass
internal/cli/migrate_ownership_test.go:65:10: undefined: keepReason
internal/cli/migrate_ownership_test.go:66:12: undefined: evidenceItem
internal/cli/migrate_ownership_test.go:158:52: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:158:52: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: 판정 함수 `classifyOwnership`, 분류 상수, 처분 구조가 아직 없어 시험 패키지 빌드가 실패한다 (`undefined: ownershipEnv`, `ownershipVerdict`, `ownershipClass`, `keepReason` …). acceptance.md 가 예상한 이유 (`undefined: classifyOwnership`) 와 같은 부류이다. 제로 매치 아님 (`[build failed]`, 선택자는 존재하는 시험 이름).

**AC-OWR-02 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveByteEqualUserFile$' -count=1 -v`; 트리 SHA: `9b73a359380e244a931e14ee10a6950e6d26c07e`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_test.go:30:36: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:32:41: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:35:74: undefined: provenanceSource
internal/cli/migrate_ownership_test.go:43:43: undefined: ownershipEnv
internal/cli/migrate_ownership_test.go:43:73: undefined: ownershipVerdict
internal/cli/migrate_ownership_test.go:51:58: undefined: remainingRow
internal/cli/migrate_ownership_test.go:64:10: undefined: ownershipClass
internal/cli/migrate_ownership_test.go:65:10: undefined: keepReason
internal/cli/migrate_ownership_test.go:66:12: undefined: evidenceItem
internal/cli/migrate_ownership_test.go:158:52: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:158:52: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: 이유 코드·빠진 증거 항목·`provenanceSource` 대역 자리가 없어 빌드 실패. 파일 보존 자체는 현재 코드도 지키므로 (기준선 `TestMigrationClassifiesUnregisteredMirrorCopy`), 이 시험이 새로 요구하는 것은 이유·증거·변이 관측이다. 제로 매치 아님 (`[build failed]`, 선택자는 존재하는 시험 이름).

**AC-OWR-03 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveUserModifiedFile$' -count=1 -v`; 트리 SHA: `9b73a359380e244a931e14ee10a6950e6d26c07e`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_test.go:30:36: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:32:41: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:35:74: undefined: provenanceSource
internal/cli/migrate_ownership_test.go:43:43: undefined: ownershipEnv
internal/cli/migrate_ownership_test.go:43:73: undefined: ownershipVerdict
internal/cli/migrate_ownership_test.go:51:58: undefined: remainingRow
internal/cli/migrate_ownership_test.go:64:10: undefined: ownershipClass
internal/cli/migrate_ownership_test.go:65:10: undefined: keepReason
internal/cli/migrate_ownership_test.go:66:12: undefined: evidenceItem
internal/cli/migrate_ownership_test.go:158:52: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:158:52: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: 분류·이유가 잔존 행으로 기록되지 않아 (`remainingRow` 없음) 빌드 실패. 보존 자체는 `TestRF2F3_MigrationPreservesPostInstallEdits` 가 이미 지킨다. 제로 매치 아님 (`[build failed]`, 선택자는 존재하는 시험 이름).

**AC-OWR-04 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipUnregisteredUnknownOriginKeptWithReason$' -count=1 -v`; 트리 SHA: `9b73a359380e244a931e14ee10a6950e6d26c07e`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_test.go:30:36: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:32:41: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:35:74: undefined: provenanceSource
internal/cli/migrate_ownership_test.go:43:43: undefined: ownershipEnv
internal/cli/migrate_ownership_test.go:43:73: undefined: ownershipVerdict
internal/cli/migrate_ownership_test.go:51:58: undefined: remainingRow
internal/cli/migrate_ownership_test.go:64:10: undefined: ownershipClass
internal/cli/migrate_ownership_test.go:65:10: undefined: keepReason
internal/cli/migrate_ownership_test.go:66:12: undefined: evidenceItem
internal/cli/migrate_ownership_test.go:158:52: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:158:52: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: 현재 출력은 경로만 대고 이유 코드와 빠진 증거 항목이 없다. 이 시험은 새 타입을 참조하므로 빌드 실패로 먼저 붉다. 제로 매치 아님 (`[build failed]`, 선택자는 존재하는 시험 이름).

**AC-OWR-05 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipAttestedManagedOnlyWithFullEvidence$' -count=1 -v`; 트리 SHA: `9b73a359380e244a931e14ee10a6950e6d26c07e`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_test.go:30:36: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:32:41: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:35:74: undefined: provenanceSource
internal/cli/migrate_ownership_test.go:43:43: undefined: ownershipEnv
internal/cli/migrate_ownership_test.go:43:73: undefined: ownershipVerdict
internal/cli/migrate_ownership_test.go:51:58: undefined: remainingRow
internal/cli/migrate_ownership_test.go:64:10: undefined: ownershipClass
internal/cli/migrate_ownership_test.go:65:10: undefined: keepReason
internal/cli/migrate_ownership_test.go:66:12: undefined: evidenceItem
internal/cli/migrate_ownership_test.go:158:52: undefined: priorGeneration
internal/cli/migrate_ownership_test.go:158:52: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: 증거 해석기·`priorGeneration`·`priorRecord`·`evidenceIncompleteReason` 가 없어 빌드 실패 (acceptance.md 가 예상한 이유: 증거 해석기가 없어 빌드 실패). 제로 매치 아님 (`[build failed]`, 선택자는 존재하는 시험 이름).

### M1b — 배포 버전 기록, 이전 매니페스트 보존, 이전 세대 원천 (AC-OWR-06, -20, -21, -05 (h), run chunk 1)

**커밋 A (시험 파일만)**: `f9206f5242` (시험 네 파일: `internal/manifest/manifest_deployed_version_test.go`, `internal/cli/update_prior_manifest_test.go`, `internal/cli/update_prior_manifest_io_test.go`, 그리고 M0 도우미가 M1b 시점까지 쓰이지 않던 부분을 다루는 `TestOwnershipHelperFixtureBuilders` 를 더한 `migrate_ownership_helpers_test.go`), 이어 `da53289c5f` (이름 충돌 수정만: 내 도우미 `keysOf` 가 `update_deny_migration_test.go` 의 기존 도우미와 겹쳐 `priorKeysOf` 로 바꿈 — 시험 파일만, 붉은 이유는 그대로). RED 는 이 두 커밋 뒤의 트리 `da53289c5f` 에서 쟀다. 모든 셀이 빌드 실패이며 (production 심볼이 없음) 같은 컴파일러 첫 줄들이 되풀이된다. 이 시점에서 cli 시험 패키지 전체가 빌드되지 않으므로 M1a 의 다섯 시험과 기준선도 함께 못 돈다 (C 커밋에서 다시 돈다).

**AC-OWR-06 (a) — manifest 패키지 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/manifest/ -run '^TestManifestRecordsDeployedVersion$' -count=1 -v`; 트리 SHA: `da53289c5f2740bd0ff4ca990bc2bdaa98b36122`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/manifest [github.com/modu-ai/moai-adk/internal/manifest.test]
internal/manifest/manifest_deployed_version_test.go:50:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:54:37: entryOf(t, mgr, "a.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:62:37: entryOf(t, mgr, "b.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:69:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:76:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:85:40: entryOf(t, mgr, "kept.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:88:42: entryOf(t, mgr, "legacy.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:94:41: entryOf(t, mgr, "fresh.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:101:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:105:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:105:7: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/manifest [build failed]
FAIL
exit=1
```

실패 이유: `FileEntry.DeployedVersion` 와 `Manager.SetDeployingVersion` 이 아직 없어 manifest 시험 패키지가 빌드되지 않는다. 제로 매치 아님 (`[build failed]`).

**AC-OWR-06 (b)(c) — cli 패키지 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestPriorManifestRetainedInRunBackup|TestPriorGenerationSourceBoundedAndTolerant)$' -count=1 -v`; 트리 SHA: `da53289c5f2740bd0ff4ca990bc2bdaa98b36122`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/update_prior_manifest_io_test.go:52:26: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:52:50: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:53:21: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:53:45: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:56:41: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:59:2: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:72:2: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:90:65: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:97:10: undefined: newPriorGenerationSource
internal/cli/update_prior_manifest_io_test.go:115:13: undefined: writePriorManifest
internal/cli/update_prior_manifest_io_test.go:115:13: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: `priorManifestName`, `priorManifestTempWrite`, `priorManifestRename`, `newPriorGenerationSource`, `priorManifestGenerations`, `priorManifestWriter` 가 아직 없어 빌드 실패 (acceptance.md 가 예상한 이유: 필드와 보존 단계와 원천이 없어 빌드 실패). 제로 매치 아님 (`[build failed]`).

**AC-OWR-20 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestBackupFailureRemovesNothingThroughProductionPath$' -count=1 -v`; 트리 SHA: `da53289c5f2740bd0ff4ca990bc2bdaa98b36122`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/update_prior_manifest_io_test.go:52:26: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:52:50: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:53:21: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:53:45: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:56:41: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:59:2: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:72:2: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:90:65: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:97:10: undefined: newPriorGenerationSource
internal/cli/update_prior_manifest_io_test.go:115:13: undefined: writePriorManifest
internal/cli/update_prior_manifest_io_test.go:115:13: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: 보존 단계·인계 (`priorManifestWriter`, `migrationRunContext.priorManifest`)·일괄 제거 구조가 없어 빌드 실패 (acceptance.md 가 예상한 이유와 같다). (b) 하위 사례가 단언 수준에서 붉을 이유는 현행 t1668 코드가 실패한 파일만 보존하고 나머지는 제거한다는 점이며, 이 시점에는 빌드가 막아 단언까지 도달하지 못한다. 제로 매치 아님 (`[build failed]`).

**AC-OWR-21 (a)-(e) RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ ./internal/manifest/ -run '^(TestPriorManifestAtomicWriteFailures|TestPriorGenerationCorruptIsNoEvidence|TestLegacyManifestRoundTrip|TestDowngradeDropsEvidenceNotSafety|TestPriorGenerationsAreLookupBoundOnly)$' -count=1 -v`; 트리 SHA: `da53289c5f2740bd0ff4ca990bc2bdaa98b36122`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/manifest [github.com/modu-ai/moai-adk/internal/manifest.test]
internal/manifest/manifest_deployed_version_test.go:50:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:54:37: entryOf(t, mgr, "a.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:62:37: entryOf(t, mgr, "b.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:69:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:76:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:85:40: entryOf(t, mgr, "kept.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:88:42: entryOf(t, mgr, "legacy.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:94:41: entryOf(t, mgr, "fresh.md").DeployedVersion undefined (type FileEntry has no field or method DeployedVersion)
internal/manifest/manifest_deployed_version_test.go:101:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:105:7: mgr.SetDeployingVersion undefined (type Manager has no field or method SetDeployingVersion)
internal/manifest/manifest_deployed_version_test.go:105:7: too many errors
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/update_prior_manifest_io_test.go:52:26: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:52:50: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:53:21: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:53:45: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:56:41: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:59:2: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:72:2: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:90:65: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:97:10: undefined: newPriorGenerationSource
internal/cli/update_prior_manifest_io_test.go:115:13: undefined: writePriorManifest
internal/cli/update_prior_manifest_io_test.go:115:13: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL	github.com/modu-ai/moai-adk/internal/manifest [build failed]
FAIL
exit=1
```

실패 이유: 원자적 쓰기 관측점 두 변수와 읽기 원천이 없어 두 패키지 모두 빌드 실패 (acceptance.md 가 예상한 이유: 원자적 쓰기·읽기 원천이 없어 빌드 실패). 제로 매치 아님 (`[build failed]`).

**AC-OWR-05 (h) — 운영 원천 경유 RED 셀** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipAttestedManagedThroughOperatingSource$' -count=1 -v`; 트리 SHA: `da53289c5f2740bd0ff4ca990bc2bdaa98b36122`; 종료 코드: 1.

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/update_prior_manifest_io_test.go:52:26: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:52:50: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:53:21: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:53:45: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:56:41: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:59:2: undefined: priorManifestTempWrite
internal/cli/update_prior_manifest_io_test.go:72:2: undefined: priorManifestRename
internal/cli/update_prior_manifest_io_test.go:90:65: undefined: priorManifestName
internal/cli/update_prior_manifest_io_test.go:97:10: undefined: newPriorGenerationSource
internal/cli/update_prior_manifest_io_test.go:115:13: undefined: writePriorManifest
internal/cli/update_prior_manifest_io_test.go:115:13: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

실패 이유: 운영 쓰기 함수와 운영 이전 세대 원천이 아직 없어 빌드 실패 (plan.md § G.1: (h) 는 M1b 이후에만 도달 가능). 제로 매치 아님 (`[build failed]`).

### Run chunk 1 GREEN 기록과 정직한 회계 (M0 + M1a + M1b)

모든 GREEN 은 트리 `a10e4a65df5b0b69955275636125169079cfb191` 에서 한 번에 다시 쟀다 (판정 도구: 이 트리의 `go test`, Go 1.26.8 darwin/arm64, 바이너리 교체 없음; 빌드 `go build ./...`). 출력에서 `=== RUN` 줄과 설정 알림 한 줄 (`SUNSET_CONFIG_DORMANT_NOTICE`) 은 뺐고 나머지는 원문이다. 마지막 줄 `exit=` 는 명령의 종료 코드이다.

**커밋 지도 (기준 HEAD `785f2b6c48`)**

| 묶음 | A (시험만) | B (RED 셀) | C (구현) | 비고 |
|------|-----------|-----------|----------|------|
| M0 | `e902867dea` (도우미 + 얼린 기준선, 도착 시 GREEN 특성화) | `360f63dd31` | — | 운영 코드 변경 없음 |
| M1a AC-OWR-01 ~ -05 (a)-(g) | `9b73a35938` | `4764764ff0` | `3133329a64` | C 에 증거 시험 한 픽스처 수정이 들어 있다 (아래 정직한 회계 1) |
| M1b AC-OWR-06, -20, -21, -05 (h) | `f9206f5242` + `da53289c5f` (이름 충돌 수정) | `558dd9e6c2` | `f8936c2070` | 뒤따른 커밋: `3d6bc6512e` (죽은 코드 제거), `9b0a503651` (시계 초 해상도 시험 고침), `a10e4a65df` (어휘 특성화 시험), `7d8be68f6c` (spec.md status draft → in-progress) |

**승격 상태 (acceptance.md 「등급과 RED 셀 규약」)**: AC-OWR-01 ~ -05, -06, -20, -21 은 A → B → C 순서를 모두 마쳤다 (커밋 그래프가 순서를 증명). RED 셀 네 요소 (명령, 원문 stdout, 종료 코드, 트리 SHA) 가 위 셀에 있고 이유를 적었으므로 release-blocking 으로 승격 자격을 갖춘다. 단 RED 는 모두 빌드 실패 형이다 — 변이 구현 앞에서 시험이 붉어지는 직접 관측은 GREEN 출력의 `MutantIsCaught` 서브 시험과 아래 변이 점검 표가 맡는다. AC-OWR-05 (h) 는 M1b 에서 승격. 나머지 AC (07 ~ 19) 는 이 chunk 의 범위 밖으로 regression-guard 상태 그대로이다.

**AC-OWR-01 GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipVerdictClasses$' -count=1 -v`

```text
--- PASS: TestOwnershipVerdictClasses (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.226s
exit=0
```

**AC-OWR-02 GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveByteEqualUserFile$' -count=1 -v`

```text
--- PASS: TestOwnershipPreserveByteEqualUserFile (3.04s)
    --- PASS: TestOwnershipPreserveByteEqualUserFile/Premise (1.04s)
    --- PASS: TestOwnershipPreserveByteEqualUserFile/Preserve (1.13s)
    --- PASS: TestOwnershipPreserveByteEqualUserFile/MutantIsCaught (0.87s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.257s
exit=0
```

**AC-OWR-03 GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveUserModifiedFile$' -count=1 -v`

```text
--- PASS: TestOwnershipPreserveUserModifiedFile (3.37s)
    --- PASS: TestOwnershipPreserveUserModifiedFile/Reach (0.93s)
    --- PASS: TestOwnershipPreserveUserModifiedFile/Preserve (1.15s)
    --- PASS: TestOwnershipPreserveUserModifiedFile/MutantIsCaught (1.28s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.510s
exit=0
```

**AC-OWR-04 GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipUnregisteredUnknownOriginKeptWithReason$' -count=1 -v`

```text
--- PASS: TestOwnershipUnregisteredUnknownOriginKeptWithReason (2.05s)
    --- PASS: TestOwnershipUnregisteredUnknownOriginKeptWithReason/Reach (0.59s)
    --- PASS: TestOwnershipUnregisteredUnknownOriginKeptWithReason/Reasons (0.83s)
    --- PASS: TestOwnershipUnregisteredUnknownOriginKeptWithReason/MutantIsCaught (0.63s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.169s
exit=0
```

**AC-OWR-05 (a)-(g) GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipAttestedManagedOnlyWithFullEvidence$' -count=1 -v`

```text
--- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence (0.02s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/Positive (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/a-no-deployed-version (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/b-record-only-under-another-path (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/c-no-record-anywhere (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/c-no-generation-at-all (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/d-disk-bytes-differ-from-record (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/e-prior-provenance-user-created (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/f-prior-provenance-user-modified (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/g-prior-provenance-deprecated (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/g-prior-provenance-generated-managed (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/g-prior-provenance-empty (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/g-prior-provenance-unknown-string (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/n14-newest-incomplete-older-complete (0.00s)
    --- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence/n15-deployed-hash-match-only (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.322s
exit=0
```

**AC-OWR-05 (h) GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipAttestedManagedThroughOperatingSource$' -count=1 -v`

```text
--- PASS: TestOwnershipAttestedManagedThroughOperatingSource (0.02s)
    --- PASS: TestOwnershipAttestedManagedThroughOperatingSource/h1-existing-untracked-file-is-recorded-user-created-and-stays-unregistered (0.01s)
    --- PASS: TestOwnershipAttestedManagedThroughOperatingSource/h2-file-the-run-deployed-is-attested-managed (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.060s
exit=0
```

**AC-OWR-06 (a) GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/manifest/ -run '^TestManifestRecordsDeployedVersion$' -count=1 -v`

```text
--- PASS: TestManifestRecordsDeployedVersion (0.01s)
    --- PASS: TestManifestRecordsDeployedVersion/DeployPathStampsTemplateManagedEntries (0.00s)
    --- PASS: TestManifestRecordsDeployedVersion/NonDeployPathsNeitherSetNorClear (0.00s)
    --- PASS: TestManifestRecordsDeployedVersion/PersistsUnderTheDeployedVersionKeyAndOmitsEmpty (0.00s)
    --- PASS: TestManifestRecordsDeployedVersion/LegacyJSONRoundTripsByteForByte (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/manifest	0.260s
exit=0
```

**AC-OWR-06 (b)(c) GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestPriorManifestRetainedInRunBackup|TestPriorGenerationSourceBoundedAndTolerant)$' -count=1 -v`

```text
--- PASS: TestPriorManifestRetainedInRunBackup (4.10s)
--- PASS: TestPriorGenerationSourceBoundedAndTolerant (0.01s)
    --- PASS: TestPriorGenerationSourceBoundedAndTolerant/BoundedToTheNewestEight (0.01s)
    --- PASS: TestPriorGenerationSourceBoundedAndTolerant/DamagedGenerationsAreNoEvidenceAndDoNotKillTheRun (0.00s)
    --- PASS: TestPriorGenerationSourceBoundedAndTolerant/CurrentRunDirectoryIsExcluded (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	5.289s
exit=0
```

**AC-OWR-20 GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestBackupFailureRemovesNothingThroughProductionPath$' -count=1 -v`

```text
--- PASS: TestBackupFailureRemovesNothingThroughProductionPath (4.02s)
    --- PASS: TestBackupFailureRemovesNothingThroughProductionPath/a-prior-manifest-copy-fails (1.53s)
    --- PASS: TestBackupFailureRemovesNothingThroughProductionPath/b-candidate-bytes-copy-fails (1.14s)
    --- PASS: TestBackupFailureRemovesNothingThroughProductionPath/c-twin-without-injection-removes-all-three (1.35s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	5.035s
exit=0
```

**AC-OWR-21 GREEN** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ ./internal/manifest/ -run '^(TestPriorManifestAtomicWriteFailures|TestPriorGenerationCorruptIsNoEvidence|TestLegacyManifestRoundTrip|TestDowngradeDropsEvidenceNotSafety|TestPriorGenerationsAreLookupBoundOnly)$' -count=1 -v`

```text
--- PASS: TestPriorManifestAtomicWriteFailures (0.04s)
    --- PASS: TestPriorManifestAtomicWriteFailures/o-normal-run-goes-through-temp-then-rename (0.01s)
    --- PASS: TestPriorManifestAtomicWriteFailures/p-failure-during-the-temporary-write (0.01s)
    --- PASS: TestPriorManifestAtomicWriteFailures/q-failure-between-the-temporary-write-and-the-rename (0.01s)
    --- PASS: TestPriorManifestAtomicWriteFailures/r-real-rename-failure-onto-an-occupied-name (0.01s)
--- PASS: TestPriorGenerationCorruptIsNoEvidence (0.03s)
    --- PASS: TestPriorGenerationCorruptIsNoEvidence/intact (0.00s)
    --- PASS: TestPriorGenerationCorruptIsNoEvidence/truncated-json (0.00s)
    --- PASS: TestPriorGenerationCorruptIsNoEvidence/empty-file (0.00s)
    --- PASS: TestPriorGenerationCorruptIsNoEvidence/json-but-not-a-manifest (0.00s)
    --- PASS: TestPriorGenerationCorruptIsNoEvidence/json-object-without-files (0.01s)
    --- PASS: TestPriorGenerationCorruptIsNoEvidence/directory-in-place-of-the-file (0.00s)
    --- PASS: TestPriorGenerationCorruptIsNoEvidence/only-a-temporary-name-exists (0.00s)
--- PASS: TestLegacyManifestRoundTrip (0.01s)
--- PASS: TestDowngradeDropsEvidenceNotSafety (0.01s)
    --- PASS: TestDowngradeDropsEvidenceNotSafety/newer_binary's_generation_attests (0.00s)
    --- PASS: TestDowngradeDropsEvidenceNotSafety/after_the_older_binary_saved,_the_file_is_kept (0.00s)
    update_prior_manifest_io_test.go:544: G-K measurement: without prior manifests 5 run backups survive [20260101_000008 20260101_000009 20260101_000010 20260101_000011 20260101_000012]; with distinct prior manifests 8 survive [20260101_000005 20260101_000006 20260101_000007 20260101_000008 20260101_000009 20260101_000010 20260101_000011 20260101_000012]
--- PASS: TestPriorGenerationsAreLookupBoundOnly (0.52s)
    --- PASS: TestPriorGenerationsAreLookupBoundOnly/OnlyTheReaderUsesTheBound (0.38s)
    --- PASS: TestPriorGenerationsAreLookupBoundOnly/SameRuleWithAndWithoutPriorManifest (0.07s)
    --- PASS: TestPriorGenerationsAreLookupBoundOnly/MeasuredWhenEveryRunHoldsADifferentPriorManifest (0.08s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.242s
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/manifest	0.132s [no tests to run]
exit=0
```


**기준선 11개 + 마이그레이션·갱신 61시험 묶음 (회귀 가드, 트리 `a10e4a65df5b0b69955275636125169079cfb191`)**

```text
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (0.49s)
--- PASS: TestMigrationKeepsSharedCopyUnderProfileReach (0.66s)
--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach (0.56s)
--- PASS: TestMigrationBackupBeforeRemoval (1.02s)
--- PASS: TestEmptiedSkillDirRemoved (2.73s)
    --- PASS: TestEmptiedSkillDirRemoved/emptied_dir_removed_with_its_last_file (1.45s)
    --- PASS: TestEmptiedSkillDirRemoved/dir_with_user_content_stays (1.29s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (1.45s)
--- PASS: TestMigrationRemovalGate (2.35s)
--- PASS: TestRF2F3b_MigrationPreservesUntracked (0.01s)
moai-cli-test: userHomeDirFn redirected real home to sandbox /private/tmp/own-fixture-d1/moai-cli-home-1420669734
--- PASS: TestMigrationIdempotent (1.53s)
--- PASS: TestMigrationPreservesExistingMirrorEntries (1.56s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	13.206s
exit=0
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	50.524s
exit=0
```

(둘째 묶음은 `update_migrate_test.go`, `fix_round3_cli_test.go`, `migrate_profile_reach_test.go`, `migrate_removal_*_test.go`, `m0_doctor_red_test.go`, `update_dryrun_preview*_test.go`, `update_dry_run_reach_test.go`, `update_manifest_retrack_test.go`, `update_template_sync*_test.go` 의 모든 최상위 시험 61개를 앵커된 `-run` 으로 쓸어낸 것이다.)

**M0 도우미 시험 + 얼린 기준선 비교**

```text
--- PASS: TestOwnershipHelperTreeSnapshotSeesDirMtime (0.01s)
    migrate_ownership_helpers_test.go:980: force=false run=1 skipped=false
    migrate_ownership_helpers_test.go:985: force=false run=2 skipped=true
    migrate_ownership_helpers_test.go:990: force=true  run=3 skipped=false
    migrate_ownership_helpers_test.go:995: force=false run=4 skipped=true
--- PASS: TestOwnershipHelperForcedRunReachesReconciliation (3.65s)
--- PASS: TestOwnershipHelperUsesRealEmbeddedPaths (1.98s)
--- PASS: TestProfileCallsiteBaselineMatchesTree (6.79s)
--- PASS: TestOwnershipHelperFixtureBuilders (0.01s)
    --- PASS: TestOwnershipHelperFixtureBuilders/RegisteredUserCreatedAndTreeDiff (0.01s)
    --- PASS: TestOwnershipHelperFixtureBuilders/TemporaryHomeGuard (0.00s)
    --- PASS: TestOwnershipHelperFixtureBuilders/CanaryProfileStates (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	13.917s
exit=0
```

**품질 측정 (트리 `a10e4a65df5b0b69955275636125169079cfb191`)**

- `gofmt -l internal/cli internal/manifest` → 빈 출력, `exit=0`.
- `go vet ./internal/cli/ ./internal/manifest/` → `exit=0`. `go build ./...` → `exit=0`. `GOOS=windows GOARCH=amd64 go build ./...` → `exit=0` (새 코드는 `os`/`path/filepath` 만 쓴다).
- `golangci-lint run ./internal/manifest/ ./internal/cli/` (v2.1.6) → 3 issues, 모두 `internal/cli/factory_run_adopt.go:79:15`, `:91:14`, `:103:14` 의 `errcheck` (`fmt.Fprintf` 반환값). 그 파일은 기준 `cdd72a814` 와 바이트 동일하다 (`git diff --quiet cdd72a814 HEAD -- internal/cli/factory_run_adopt.go` → `exit 0`; 카드 t1707 의 기존 부채). 이 카드가 만든 파일에는 이슈가 0건이다. 범위 규율로 고치지 않았다.
- `go test ./internal/manifest/ -count=1 -cover` → `coverage: 89.7% of statements` (`exit=0`). `internal/cli` 는 패키지 전체를 돌리지 않는 제약 때문에 패키지 단위 커버리지를 재지 못했다 (Gaps). 대신 이 chunk 의 시험 묶음에 `-coverprofile` 을 걸어 새 코드의 함수별 커버리지를 쟀다: `classifyOwnership` 90.5%, `classifyUnregistered` 100%, `disposition` 100%, `reasonTag` 100%, `provenanceEvidence` 100%, `backUpRemovalCandidates` 100%, `writePriorManifest` 100%, `readPriorGenerations` 97.1%, `recordManifestDeployedVersion` 100%, `readStartManifest` 100%; 이 측정 뒤에 빈 가지를 `migrate_ownership_vocabulary_test.go` 로 메웠고, 그 두 시험만 따로 돌려 다시 쟀다 (`go test ./internal/cli/ -run '^(TestKeepReasonVocabularyIsClosed|TestPreservePriorManifestOutcomes)$' -count=1 -v -coverprofile=…`, 트리 `a10e4a65df`, 두 시험 `--- PASS`, `exit=0`): `valid` 100%, `isKnownKeepReason` 100%, `preservePriorManifest` 100%, `writePriorManifestTemp` 50.0% (임시 파일의 쓰기·동기화·닫기 실패 가지는 이 환경에서 유발할 수 없어 닿지 못했다 — Gaps).
- 서브에이전트 경계 grep (`AskUserQuestion` 을 새 비시험 파일 다섯에서 찾기) → 일치 없음, `exit=1`.
- `removalEnabled` 비시험 대입 스캔 (`grep -rn --include='*.go' --exclude='*_test.go' 'removalEnabled\s*=' internal`) → 선언 한 줄 (`internal/cli/migrate_project_assets.go:39:var removalEnabled = false`) 뿐, `exit=0`. 이 chunk 의 제거 관련 시험은 모두 `enableRemovalForTest` 로 스위치를 시험 안에서만 켠다.
- 보호 파일 불변: `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile` → `exit=0`. AC-OWR-08 명령 2 형식: `git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- internal/config/protected_zone.go internal/config/userroots.go internal/profile` → 빈 stdout, `exit=0`; 양성 대조군 `-- .moai/specs` → 한 줄 이상 (측정 시 10줄, `exit=0`). AC-OWR-21 (e) 명령 2: `... -- internal/cli/update/backup` → 빈 stdout, `exit=0`.

**변이 점검 (이 chunk 의 핵심 변이를 구현에 직접 주입해 시험이 붉어지는지 관측, 주입 뒤 원상 복구함)**

| 변이 | 붉어진 시험 |
|------|-------------|
| 이전 매니페스트를 최종 이름에 직접 쓰기 | `TestPriorManifestAtomicWriteFailures` o, p, q (r 은 AC 가 말한 대로 못 가른다) |
| 사본을 재직렬화 | `TestPriorManifestRetainedInRunBackup` |
| Backup 단계가 사본 실패를 삼키고 성공을 인계 | `TestBackupFailureRemovesNothingThroughProductionPath/a-prior-manifest-copy-fails` |
| 후보 파일 바이트 사본 실패를 건너뛰고 제거 계속 | `.../b-candidate-bytes-copy-fails` |
| `Track` 이 이미 있는 배포 버전을 지움 / 어디서나 지어냄 / `user_created` 도 도장 | `TestManifestRecordsDeployedVersion` 의 해당 하위 시험 |
| 세대 상한 없이 읽기 | `TestPriorGenerationSourceBoundedAndTolerant/BoundedToTheNewestEight` |
| 매 실행 `DeployedAt` 갱신 / 배포 저장에서 모든 항목에 버전 찍기 | `TestPriorManifestRetainedInRunBackup` |
| 이전 기록 출처값 무시 | `TestOwnershipAttestedManagedOnlyWithFullEvidence` e, f, g 와 n14 |
| 오래된 세대가 새 세대를 이김 | `.../n14-newest-incomplete-older-complete` |
| 바이트 일치 승격 대역 / 템플릿 디렉터리 이름 승격 대역 / 바이트 무시 분류기 | `MutantIsCaught` 서브 시험 (AC-OWR-02, -03, -04) |
| 두 관측점을 정상으로 거친 뒤 최종 이름을 다시 직접 쓰기 (M-F) | **붉어지지 않음** (`ok`) — 부채 N24 가 공개한 한계이며 실측으로 확인했다 |

**정직한 회계 (RED → GREEN 의 어긋남을 숨기지 않는다)**

1. M1a 의 증거 시험 (c) 경우의 픽스처는 A 커밋에서 틀렸다: 「어느 경로에도 기록 없음」 이라면서 무관한 키의 기록이 디스크 바이트와 같은 해시를 가져 사실은 (b) 모양이었다. 구현을 붙이자 이 시험이 붉어져 알았고, C 커밋 `3133329a64` 에서 시험 파일의 그 한 기록만 다른 해시로 고쳤다. 단언은 바뀌지 않았다.
2. M1b A 커밋의 도우미 `keysOf` 가 기존 `update_deny_migration_test.go` 의 `keysOf` 와 겹쳐 컴파일 오류가 하나 더 섞였다. RED 를 재기 전에 별도 커밋 `da53289c5f` 로 이름만 바꿨고 RED 셀은 그 뒤 트리에서 쟀다.
3. `TestPriorManifestRetainedInRunBackup` 은 실행 백업 디렉터리 이름이 초 해상도이므로 빠른 두 실행이 한 디렉터리를 공유하면 흔들렸다 (한 번 붉음을 직접 관측). 시험 쪽에서 실행 2 와 3 앞에 1.1초를 기다리게 고쳤다 (`f8936c2070` 의 실행 3 대기, `9b0a503651` 의 실행 2 대기). 운영 동작은 바꾸지 않았다. 참고: 같은 초의 두 실행이 한 실행 백업을 공유하면 둘째 실행의 `manifest.prior.json` 이 이름 바꾸기로 첫째의 것을 원자적으로 대체한다.
4. 변이 점검 중 `track-always-stamps` 변이 한 번은 컴파일 오류 (`prev` 미사용) 로 변이로서 성립하지 않아 다시 만들어 돌렸다 (표는 성립한 변이만 적었다).
5. 이 chunk 의 A 커밋 직후에는 `internal/cli` 시험 패키지가 통째로 빌드되지 않으므로 그 시점에는 기준선 11개를 돌릴 수 없었다. 회귀 가드는 C 이후 트리에서 돌렸다.
6. Manager 인터페이스에 메서드가 하나 늘었으므로 (`SetDeployingVersion`) 그 영향을 받는 패키지를 `go vet` 로 컴파일 확인했고 (`go vet ./internal/cli/ ./internal/cli/update/ ./internal/cli/update/backup/ ./internal/cli/update/merge/ ./internal/codexwiring/ ./internal/core/project/ ./internal/manifest/ ./internal/template/` → `vet-exit=0`, 트리 `f8936c2070`), `./internal/manifest/ ./internal/template/ ./internal/core/project/ ./internal/cli/update/backup/` 패키지 시험도 한 번 돌렸다 (모두 `ok`; `internal/template` 패키지 전체는 249초 걸렸다 — 무거운 실행 한 번이었음을 밝힌다).

**Binding run conditions 처분 (이 chunk 의 파일에 닿는 것)**

| 부채 | 처분 | 근거 |
|------|------|------|
| N14 세대 우선순위 없음 | 처분됨 (고침, 보존 방향) | 레코드를 가진 가장 새 세대가 결정한다. `provenanceEvidence`, 시험 `n14-newest-incomplete-older-complete` + 변이 `older-gen-outvotes` 붉음 |
| N15 `DeployedHash` 일치 허용 | 처분됨 (고침, 보존 방향) | 비교를 `TemplateHash` 한 가지로 좁혔다 (등록 경로와 같다). design.md D-2 의 「(또는 DeployedHash)」 문구와 다르게 더 엄격하다 — 리더가 알아야 하는 편차로 아래에 올렸다. 시험 `n15-deployed-hash-match-only` |
| N18 AC-OWR-20 (a)(b) 와 규칙 7 조건 (ii) | 처분됨 | (a) 에서는 새 유효 `manifest.prior.json` 이 만들어지지 않는 것이 시험의 요점이므로 도달을 「쓰기 함수 호출 1회 + 마이그레이션 도달」 로 단언했고, (b) 는 `rc.priorManifest == written` 으로 단언했다. 시험 주석에 적었다 |
| N19 `.moai/config` 없을 때 실행 디렉터리 | 처분됨 | `preservePriorManifest` 가 설정 백업이 디렉터리를 만들지 않았으면 `ensureRunBackupDir` 로 새 실행 디렉터리 (`<프로젝트>/.moai-backups/<타임스탬프>/manifest.prior.json`) 를 만든다. 시험 「a fresh run directory when the config backup made none」 |
| N20 건너뛴 `UserCreated` 항목에 `deployed_version` 없음 | 처분됨 | `Track` 은 `template_managed` 만 도장한다. 시험 `DeployPathStampsTemplateManagedEntries` 와 `h1-...` (실제 배포기로 만든 건너뛴 항목의 값이 비어 있음) |
| N21 AC-OWR-20 (b) 의 `removalBackupWriter` | 처분됨 (SPEC 문구 결함으로 보고) | D-3 항목 5 가 정의하는 것은 `priorManifestWriter` 하나뿐이라 (b) 는 그 대역이 실행 디렉터리 안에 막는 디렉터리를 미리 만드는 방식으로 구현했다. REQ-OWR-005 문구 부분은 M2 chunk 의 몫 |
| N22 AC-OWR-21 When 절의 구성 (o) 누락 | 처분됨 (시험이 (o) 를 가진다, SPEC 문구 결함으로 보고) | `TestPriorManifestAtomicWriteFailures/o-normal-run-goes-through-temp-then-rename` |
| N24 변이 M-F 미검출 | 확인됨, 못 닫음 (공개한 한계) | 위 변이 표: 두 관측점을 정상으로 거친 뒤 최종 이름을 다시 쓰는 구현은 `ok`. 접근 기록 시임 없이는 닫을 수 없다 |
| N23 (M8), N4 잔여 (M8), N16 (M2, 카드 t1702 보류 범위), N17 (M9) | 이월 | 각각 M8 / M2 / M9 를 맡는 chunk 로 넘긴다 |

**SPEC 쪽 결함·불일치로 보고하는 것 (고치지 않았다)**

- AC-OWR-21 의 판정 명령 1 은 `./internal/cli/ ./internal/manifest/` 를 함께 쓸지만 다섯 시험은 모두 `internal/cli` 에 있어 manifest 쪽은 `[no tests to run]` (`ok`) 이다 (규칙 4: 제로 매치는 통과가 아니다). 판정은 cli 줄의 다섯 `--- PASS` 로 성립한다.
- AC-OWR-21 (a) 의 When 절에 (o) 가 없다 (N22), AC-OWR-20 (b) 가 정의되지 않은 `removalBackupWriter` 를 가리킨다 (N21).
- AC-OWR-06 (b) 가 말하는 「사용자 수정 파일이 있어 실행 2 가 그 파일을 다시 배포하지 않는다」 는 이 트리에서 프로젝트 쪽 공통 자산 (스킬·에이전트) 이 프로젝트 배포기에 의해 배포되지 않는다는 사실 (사용자 홈 설치로 이전) 에 기대야 성립한다. 시험은 배포되지 않는 공통 자산 항목 두 개 (오래된 도장이 있는 것, 없는 것) 로 「배포하지 않은 항목은 지어지지도 지워지지도 않는다」 를 단언했다.
- design.md D-3 항목 1 은 「`Track` 이 찍는다」 고만 적었으나 `Track` 은 배포 경로와 재추적 경로가 공유하므로 둘을 가를 장치가 필요했다. 계획의 파일 목록 안에서 `Manager` 에 배포 창 설정 메서드 `SetDeployingVersion` 을 더하는 방식으로 풀었다 (`internal/template/deployer.go` 는 건드리지 않았다). 인터페이스 확장이 계획의 명시적 선택은 아니므로 리더가 승인 여부를 알아야 한다.
- 갭 G-K 를 실측했다 (acceptance.md 는 「측정하지 않았다」): 실행 백업 12개, `PruneBackups(…, 5, 0)`; `manifest.prior.json` 이 없으면 5개가 남고, 실행마다 내용이 다른 `manifest.prior.json` 이 있으면 8개가 남는다 (센서스가 서로 다른 버전을 보호 대상으로 세어 오래된 3개를 더 남김). 정책은 바꾸지 않았고 기존 `pastKeep` 한도 안이지만, 이 SPEC 이 실행 백업마다 파일 하나를 더하므로 디스크 사용이 정책의 상한까지 늘어난다.

### Run chunk 2 (M2, M3, M4) — 기록 시작

시작 HEAD `5eed5816ae`, 브랜치 `WT-update-ownership-repair`, 기록 규약은 chunk 1 과 같다 (명령, 원문 출력, 종료 코드, 트리 SHA; 시험은 `TMPDIR=/private/tmp/own-fixture-d1` 아래 임시 디렉터리와 임시 HOME 만 썼다; 판정 도구는 이 트리의 `go test`, 바이너리 교체 없음). 이 chunk 는 `internal/cli` 패키지 전체, `-race`, 자원 슬롯을 쓰지 않는다 (앵커된 `-run` 만).

**리더 결정 세 건 (lane, 이 chunk 의 입력으로 기록)**

1. `Manager.SetDeployingVersion` 인터페이스 추가 (chunk 1 의 M1b) 는 승인되었다.
2. N15 의 더 엄격한 비교 (`TemplateHash` 한 가지만 비교, 등록 경로와 같게) 는 승인되었다.
3. SPEC 문구 결함 네 건 — AC-21 판정 명령 경로, 구성 (o) 누락, `removalBackupWriter`, AC-06 (b) — 은 공개된 부채로 남고 SPEC 본문 (spec.md, plan.md, acceptance.md, design.md, research.md) 은 고치지 않는다.

#### M2 — 프로필 도달 유지 이유 (AC-OWR-07, AC-OWR-08)

**AC-OWR-07 — 커밋 A** `73ace0cc2e` (시험 파일만: `internal/cli/migrate_ownership_profile_test.go`). 시험 `TestProfileMappedCopyKeptWithoutCounterpart` 는 서브 시험 여섯: `NoCounterpart` (빈 사용자 루트), `ConfirmedCounterpartStillKept`, `DifferingCounterpartIsUnconfirmed`, `ReachControlUnmappedTwinRemoves` (프로필에 지도하지 않은 쌍둥이가 같은 입력으로 세 사본을 제거 — 규칙 8), 변이 `MutantIgnoringTheMappingIsCaught`, `MutantPretendingToReadTheProfileSideIsCaught`.

**AC-OWR-07 RED 셀 (트리 `73ace0cc2e`, A 커밋 직후, 작업 트리 깨끗)**

명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestProfileMappedCopyKeptWithoutCounterpart$' -count=1 -v`

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_profile_test.go:65:73: undefined: counterpartStatus
internal/cli/migrate_ownership_profile_test.go:96:11: row.Counterpart undefined (type remainingRow has no field or method Counterpart)
internal/cli/migrate_ownership_profile_test.go:97:95: row.Counterpart undefined (type remainingRow has no field or method Counterpart)
internal/cli/migrate_ownership_profile_test.go:99:11: row.ProfileSide undefined (type remainingRow has no field or method ProfileSide)
internal/cli/migrate_ownership_profile_test.go:99:26: undefined: profileSideNotChecked
internal/cli/migrate_ownership_profile_test.go:100:86: row.ProfileSide undefined (type remainingRow has no field or method ProfileSide)
internal/cli/migrate_ownership_profile_test.go:100:99: undefined: profileSideNotChecked
internal/cli/migrate_ownership_profile_test.go:128:52: undefined: counterpartStatusAbsent
internal/cli/migrate_ownership_profile_test.go:140:52: undefined: counterpartStatusConfirmed
internal/cli/migrate_ownership_profile_test.go:163:52: undefined: counterpartStatusUnconfirmed
internal/cli/migrate_ownership_profile_test.go:163:52: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

종료 코드 1. 이유: 잔존 행에 사용자 루트 대응물 상태와 프로필 쪽 상태를 담는 어휘와 필드가 아직 없어서 시험 패키지가 빌드되지 않는다 (시험이 존재하는 상태의 빌드 실패이며 제로 매치 셀렉터가 아니다). 이 RED 는 빌드 실패 형이다: 단언 수준의 실패 (`absent` 가 기록되지 않음 등) 는 이 시점에 관측되지 않는다. 변이 구현 앞에서 시험이 붉어지는 직접 관측은 C 커밋 GREEN 원문의 두 `Mutant...IsCaught` 서브 시험이 한다. 코드 읽기로 아는 사실 (측정 아님): `profile-reach` 이유와 `KeptBy` 표지는 chunk 1 의 코드가 이미 쓴다. 모자란 것은 두 상태 필드뿐이다.

**AC-OWR-08 — 커밋 A** `aa8037a7f2` (시험 파일만: `internal/cli/migrate_ownership_noprofileaccess_test.go`; C07 `cca0c9ead1` 뒤에 컴파일된다). 세 시험: `TestNoProfileRootAccessInUpdateSources` (L1: `go list -deps` 닫힘과 스캔 집합이 양방향으로 같음 + 여섯 핵심 패키지 포함, 기준선 대비 새 호출 지점·늘어난 개수 없음, 이 SPEC 이 더하거나 고친 `migrate_ownership*.go`·`update_prior_manifest.go`·`update_template_sync.go`·`update_dryrun_preview.go` 에는 프로필 호출 지점이 0 이고 `migrate_project_assets.go` 는 기준선의 한 행 `profile.ResolveLaunchProfileForProject` x1 만 남음), `TestProfileScanCatchesMutants` (L2: 변이 M-a ~ M-f 를 실제 닫힘 소스에 끼워 비어 있지 않은 문제 목록을 직접 관측; 변이 없는 맵은 깨끗), `TestProfileEnvDoesNotReachProfileRoot` (L3: 정규화 규칙 자기 단언, **같은 상태 두 번 구동의 동일성 대조군** — 정규화가 잡음을 지우는지 먼저 확인 — 그리고 상태 셋 `absent`/`template-equal`/`differing` 의 정규화 관측 기록 동일성, 그리고 프로필 디렉터리를 읽는 분류기 변이 앞에서 기록이 갈라짐).

**등급 정직 공개.** AC-OWR-08 의 시험은 운영 코드에 새 심볼을 요구하지 않고 (잔존 행 필드 `ProfileSide` 는 C07 이 이미 도입), 깨끗한 트리에서 도착 시 GREEN 이다. 이 기준은 「변이가 붉어짐」 으로 증명되는 가드이고 AC 의 RED 이유 문장도 그렇게 말한다 (변이 입력으로 적색을 기록). 그래서 AC-OWR-08 에는 구현 커밋 C 가 없다: 아래 셀이 도착 셀이자 GREEN 셀이다. 붉음은 세 층에서 직접 관측했다 — (1) 시험 안 변이 (아래 도착 출력의 `M-a`..`M-f` 와 `MutantReadingTheCanaryIsSeen` 서브 시험 PASS = 변이가 붉게 잡혔다는 뜻), (2) **실제 운영 코드에 임시로 끼운 변이 두 건** (R1: 스캔, R2: 카나리; 둘 다 되돌렸고 `git diff --quiet` 와 `git status --short` 빈 출력으로 확인), (3) 파일 불변 L4 명령.

**AC-OWR-08 도착 셀 (트리 `aa8037a7f2`, 작업 트리 깨끗)** — 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestProfileEnvDoesNotReachProfileRoot|TestNoProfileRootAccessInUpdateSources|TestProfileScanCatchesMutants)$' -count=1 -v`

```text
--- PASS: TestNoProfileRootAccessInUpdateSources (11.42s)
    --- PASS: TestNoProfileRootAccessInUpdateSources/ScannedSetIsTheGoListClosure (5.22s)
    --- PASS: TestNoProfileRootAccessInUpdateSources/NoNewCallSiteAgainstTheBaseline (1.99s)
    --- PASS: TestNoProfileRootAccessInUpdateSources/MigrationPreviewAndVerdictFilesAddNothing (0.80s)
--- PASS: TestProfileScanCatchesMutants (5.08s)
    --- PASS: TestProfileScanCatchesMutants/UnmutatedTreeIsClean (0.54s)
    --- PASS: TestProfileScanCatchesMutants/M-a_read-and-discard_through_an_exported_helper (0.49s)
    --- PASS: TestProfileScanCatchesMutants/M-b_spliced_token_read_of_the_config-dir_variable (0.98s)
    --- PASS: TestProfileScanCatchesMutants/M-c_a_second_os.Stat_in_launchCandidateIsUsable (0.50s)
    --- PASS: TestProfileScanCatchesMutants/M-d_a_new_function_in_internal/profile_that_lists_a_directory (1.34s)
    --- PASS: TestProfileScanCatchesMutants/M-e_a_new_file_calling_an_exported_profile_helper (0.75s)
    --- PASS: TestProfileScanCatchesMutants/M-f_an_environment_read_with_a_non-constant_name (0.49s)
--- PASS: TestProfileEnvDoesNotReachProfileRoot (21.14s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/NormalizationRules (0.00s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/ControlTwoRunsOfTheSameStateAreIdentical (4.80s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/ThreeProfileStatesLeaveTheSameRecord (9.30s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/MutantReadingTheCanaryIsSeen (7.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	39.215s
exit=0
```

**AC-OWR-08 명령 2·3 (L4 파일 불변, 같은 트리 `aa8037a7f2`)** — 명령 3 이 양성 대조군이다 (명령 2 의 빈 출력을 읽기 전에 먼저 확인: 한 줄 이상이어야 한다). 이 트리에서 측정한 값은 12줄.

```text
$ git log --oneline cdd72a814..HEAD --grep=(card t1865) -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
exit=0
$ git log --oneline cdd72a814..HEAD --grep=(card t1865) -- .moai/specs
36f638ef0a docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): chunk 2 start, lane decisions, M2 AC-OWR-07 RED cell (card t1865)
5eed5816ae docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): run chunk 1 GREEN record, mutation checks, honest accounting and debt dispositions (card t1865)
7d8be68f6c feat(SPEC-UPDATE-OWNERSHIP-REPAIR-001): M1 status draft to in-progress (card t1865)
558dd9e6c2 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): M1b RED cells for AC-OWR-06, -20, -21 and AC-OWR-05 (h) (card t1865)
4764764ff0 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): M1a RED cells for AC-OWR-01..05 (card t1865)
360f63dd31 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): M0 baseline evidence and M6 commands probe result (card t1865)
785f2b6c48 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): Kickoff transition record and binding run conditions (card t1865)
073b3e1382 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): ceiling correction for N12 and N13 (card t1865)
a2a7b8bb24 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): plan audit iteration 2 repair N1-N11, Q2 decided (a) conditional, narrow OQ-5 carve-out, v0.3.0 (card t1865)
3c334dfdd7 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): plan audit iteration 1 repair, v0.2.0 (card t1865)
a8ad48c08d docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): plan-phase artifacts (Tier L, 5 artifacts) (card t1865)
3e77fffd65 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): read-only code investigation record (card t1865)
exit=0
```

**R1 — 실제 운영 코드 변이 (스캔 적색, 트리 `aa8037a7f2` + 임시 변경)**: `internal/cli/migrate_ownership.go` 에 `internal/profile` import 와 `_, _ = os.ReadDir(filepath.Join(profile.GetProfileDir("work"), "skills"))` 를 담은 함수를 임시로 더했다. 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestNoProfileRootAccessInUpdateSources$' -count=1 -v`

```text
    migrate_ownership_noprofileaccess_test.go:163: NEW call site "internal/cli/migrate_ownership.go\tprofile.GetProfileDir" x1
    migrate_ownership_noprofileaccess_test.go:185: new judgment, sync or preview code carries a profile call site: "internal/cli/migrate_ownership.go\tprofile.GetProfileDir" x1
--- FAIL: TestNoProfileRootAccessInUpdateSources (17.78s)
    --- PASS: TestNoProfileRootAccessInUpdateSources/ScannedSetIsTheGoListClosure (7.98s)
    --- FAIL: TestNoProfileRootAccessInUpdateSources/NoNewCallSiteAgainstTheBaseline (2.09s)
    --- FAIL: TestNoProfileRootAccessInUpdateSources/MigrationPreviewAndVerdictFilesAddNothing (0.77s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	19.700s
FAIL
exit=1
```

되돌린 뒤 `git diff --quiet -- internal/cli/migrate_ownership.go` → 종료 코드 0 (`REVERTED-CLEAN` 출력), `git status --short` 빈 출력.

**R2 — 실제 운영 코드 변이 (카나리 적색, 트리 `aa8037a7f2` + 임시 변경)**: `classifyOwnership` 의 프로필 쪽 기록에 `os.ReadFile(CLAUDE_CONFIG_DIR/skills/x/SKILL.md)` 가 성공하면 `seen` 으로 적는 세 줄을 임시로 더했다. 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestProfileEnvDoesNotReachProfileRoot$' -count=1 -v` (이 변이는 스캔이 아니라 카나리와 도달 단언이 잡는다: 같은 상태 대조군은 PASS, 세 상태 시험이 붉음)

```text
    migrate_ownership_noprofileaccess_test.go:603: .claude/skills/moai-custom/SKILL.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/skills/moai-foundation-core/SKILL.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/skills/moai-workflow-tdd/SKILL.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/agents/moai/manager-develop.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/agents/moai/manager-docs.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/skills/moai-custom/SKILL.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/skills/moai-foundation-core/SKILL.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/skills/moai-workflow-tdd/SKILL.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/agents/moai/manager-develop.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:603: .claude/agents/moai/manager-docs.md: profile side "seen", want "not-checked"
    migrate_ownership_noprofileaccess_test.go:609: run 1 differs from run 0 in the remaining rows
    migrate_ownership_noprofileaccess_test.go:609: run 2 differs from run 0 in the remaining rows
--- FAIL: TestProfileEnvDoesNotReachProfileRoot (16.25s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/NormalizationRules (0.00s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/ControlTwoRunsOfTheSameStateAreIdentical (1.98s)
    --- FAIL: TestProfileEnvDoesNotReachProfileRoot/ThreeProfileStatesLeaveTheSameRecord (5.74s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/MutantReadingTheCanaryIsSeen (8.53s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	17.684s
FAIL
exit=1
REVERTED-CLEAN
```

#### M3 — 잔존 목록 (AC-OWR-14)

**AC-OWR-14 — 커밋 A** `167a6c0af4` (시험 파일만: `internal/cli/migrate_ownership_remaining_test.go`). 시험 `TestRemainingListKeeps313UnregisteredCopies` 는 서브 시험 일곱: `Premise` (내장 `.claude/skills` 파일 중 경로 순 앞 313개의 바이트를 읽어 `.agents/skills/<같은 상대 경로>` 에 두고 각각 바이트 일치 단언), `StructuredRowsKeepEveryCopy` (제거 가능한 등록 형제 `.claude/skills/moai-foundation-core/SKILL.md` 가 **같은 실행에서 먼저 제거됨**을 단언한 뒤 313개의 바이트·미제거·매니페스트 키 무증가·행 정확히 하나·분류 `unregistered`·이유 `no-provenance-record`·빠진 증거 셋·전체 행 경로 사전순·이유 폐쇄 집합), 변이 셋 (`MutantEmptyReasonRowsIsCaught`, `MutantPrintingEveryPathByDefaultIsCaught`, `MutantBytesEqualPromotionIsCaught` — 내장 `.claude/skills` 같은 상대 경로와 바이트가 같은 `.agents/skills` 파일에 세 증거를 모두 답하는 대역 원천), `RenderedByTheUpdateFlow` (실행 1: 일반, 기본 보기 / 실행 2: `--force` + `--verbose`, 규칙 7 의 도달 단언 — `skipped == false`, 새 `manifest.prior.json` 의 바이트가 실행 전 매니페스트와 같음 — 그리고 두 실행 뒤 313개 바이트 불변), `MixedReasonsEveryRowHasOne` (프로필 지도 사본, 사용자 수정, 사용자 작성, 미등록이 섞인 둘째 픽스처). 렌더러는 시험이 직접 부르지 않고 실제 갱신 흐름의 끝 터미널 블록 (`renderUpdateTerminalBlock`) 으로만 읽는다 — 그래서 이 시험은 렌더러가 없는 트리에서도 컴파일되고 단언 수준에서 붉다.

**AC-OWR-14 RED 셀 (트리 `167a6c0af4`, A 커밋 직후, 작업 트리 깨끗)**

명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$' -count=1 -v`

```text
    migrate_ownership_remaining_test.go:351: the remaining rows are not in path order: 1 occurrence(s), first: the list as returned
    migrate_ownership_remaining_test.go:419: default view: 313 planted paths are printed in the default standard output
    migrate_ownership_remaining_test.go:419: default view: no default line names "no-provenance-record" with a count of at least 313
    migrate_ownership_remaining_test.go:435: verbose view: a planted path has 0 verbose rows, want exactly 1: 313 occurrence(s), first: .agents/skills/moai-domain-backend/SKILL.md
--- FAIL: TestRemainingListKeeps313UnregisteredCopies (22.52s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/Premise (0.17s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/StructuredRowsKeepEveryCopy (1.05s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MutantEmptyReasonRowsIsCaught (1.10s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MutantPrintingEveryPathByDefaultIsCaught (0.14s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MutantBytesEqualPromotionIsCaught (12.44s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (3.61s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MixedReasonsEveryRowHasOne (4.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	23.580s
FAIL
exit=1
```

종료 코드 1. 이유 (단언 수준의 실패, 올바른 이유로 붉음): (1) 잔존 행이 경로 사전순이 아니다 — 걷기가 `.claude/skills` 뿌리를 `.agents/skills` 보다 먼저 돌아 행이 걷기 순서로 쌓인다. (2) **기본 표준 출력이 313개 경로를 전부 찍는다** — 갱신 흐름이 마이그레이션의 파일별 보고 줄 (`migration: unregistered mirror copy ... [reason: ...]: <경로>`) 을 그대로 `out` 으로 보낸다 (이번 측정으로 확인된 기존 동작; 코드 읽기 아님). (3) 기본 보기에 `no-provenance-record` 와 313 이상의 개수를 담은 이유별 한 줄이 없다. (4) `--verbose` 터미널 블록에 파일별 잔존 행이 없다 (313개 모두 0 행). 같은 출력의 통과 서브 시험이 보여주는 것: 구조 행의 나머지 (분류, 이유, 빠진 증거, 바이트, 제거 대조군) 는 chunk 1 의 코드가 이미 맞게 낸다 — 모자란 것은 정렬과 렌더뿐이다. 변이 서브 시험 셋의 PASS 는 각 변이 앞에서 단언이 비어 있지 않은 문제 목록을 돌려줬다는 뜻이다. 제로 매치가 아니며 (서브 시험 7개가 실제로 돌았다) 빌드 실패도 아니다.

#### M4 — 대소문자 변형 (AC-OWR-11, AC-OWR-12)

시험을 두 커밋으로 나눴다. 호스트 쪽 시험은 지금 있는 식별자만 쓰므로 컴파일되어 **단언 수준에서** 붉고, 파일시스템 대소문자 모드를 주입하는 시험은 새 API (`caseFS`, `resolveCaseAlias`, 판정의 `caseAlias`/`caseVariantOf`/`key`) 를 쓰므로 빌드 실패로 붉다. 한 커밋에 묶으면 호스트 쪽 단언 수준 RED 가 빌드 실패에 가려지기 때문이다.

**AC-OWR-11 호스트 쪽 — 커밋 A1** `5246f408cf` (시험 파일만: `internal/cli/migrate_ownership_case_host_test.go`). 시험 `TestCaseVariantOnHostFilesystem` 는 서브 시험 둘: 호스트의 대소문자 모드를 **시험 자신의 임시 디렉터리 안에서** 프로브하고 (프로젝트 루트에 쓰는 `probeCaseSensitiveFS` 는 쓰지 않는다) 그 결과가 요구하는 기대를 단언한다 — 건너뛰는 가지가 없다 (무구분 호스트: 물리 파일 하나, 등록 항목의 판정이 적용됨, 철자 차이가 보고됨, 디렉터리 목록 불변; 구분 호스트: 두 철자가 별개 파일, 소문자 쪽은 `unregistered`, 등록 쪽은 `managed`); 그리고 운영 소스에서 `probeCaseSensitiveFS` 호출식을 AST 로 세어 0 임을 단언하고 (비시험 호출자가 없다), 호출을 끼운 변이 소스 앞에서 스캔이 붉음을 직접 관측한다.

**AC-OWR-11 호스트 쪽 RED 셀 (트리 `5246f408cf`, 작업 트리 깨끗; 이 머신은 darwin, 대소문자 무구분)**

명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCaseVariantOnHostFilesystem$' -count=1 -v`

```text
    migrate_ownership_case_host_test.go:121: case-insensitive host: class "unregistered" for the lower-case spelling, want "managed" (the registered entry's verdict applies to the one file)
    migrate_ownership_case_host_test.go:124: case-insensitive host: the spelling difference is not reported: "  migration: unregistered mirror copy (kept — removal needs creation provenance or your approval) [reason: no-provenance-record; missing evidence: deployed-version, original-deployed-path, prior-manifest-record]: .claude/skills/moai-ref-owasp-checklist/skill.md"
--- FAIL: TestCaseVariantOnHostFilesystem (1.75s)
    --- FAIL: TestCaseVariantOnHostFilesystem/the_host's_own_case_mode_decides_the_expectation (1.43s)
    --- PASS: TestCaseVariantOnHostFilesystem/the_operational_code_never_runs_the_project-root_probe (0.32s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.941s
FAIL
exit=1
```

종료 코드 1. 이유 (단언 수준, 올바른 이유로 붉음): 무구분 호스트에서 등록 키 `SKILL.md` 와 같은 물리 파일을 걷기가 `skill.md` 로 만나면 맵 조회가 빗나가 `unregistered` 로 센다 (현장 증상 (4) 그대로); 철자 차이도 보고되지 않는다. 둘째 서브 시험 PASS 는 `probeCaseSensitiveFS` 의 비시험 호출자가 이미 없다는 사실의 도착 시 GREEN 이다 (가드이며 변이로 붉음을 관측한다).

**AC-OWR-11 주입 쪽, AC-OWR-12 — 커밋 A2** `ae8868f6fc` (시험 파일만: `internal/cli/migrate_ownership_case_test.go`). `TestCaseVariantResolvesToRegisteredEntry` 는 파일시스템 모드를 주입한다: 실제 디렉터리 내용 위에서 무구분/구분 `stat` 을 흉내 내 두 모드를 같은 시험에서 단언하고 (무구분: 한 파일, 등록 항목의 판정 — 제거 가능 판정까지 — 이 적용되고 사용자 수정은 사용자 수정으로 남음, 구분: `unregistered`, 키 없음), 정확한 철자를 먼저 찾고, 모호한 별칭은 별칭이 아니며, 항상 별칭으로 합치는 대역은 붉게 잡힌다. `TestCaseVariantCounterpartAndMirrorSpelling` (AC-OWR-12) 은 사용자 매니페스트가 `SKILL.md` 를 기록하고 프로젝트 파일은 `skill.md` 인 경우에 대응물 상태가 `confirmed` 이고, 철자가 `skill.md` 인 미등록 거울 복사본이 한 행으로 `no-provenance-record` 와 출하된 철자 `SKILL.md` 에 대한 변형 표지를 받으며 두 철자 모두 이름 바뀌지 않음을 단언하고, 구분 모드 대조군에서는 같은 프로젝트가 `confirmed` 가 아님을 단언한다.

**AC-OWR-11 주입 쪽 / AC-OWR-12 RED 셀 (트리 `ae8868f6fc`, 작업 트리 깨끗)**

명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestCaseVariantResolvesToRegisteredEntry|TestCaseVariantOnHostFilesystem|TestCaseVariantCounterpartAndMirrorSpelling)$' -count=1 -v`

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_case_test.go:37:30: undefined: caseFS
internal/cli/migrate_ownership_case_test.go:38:9: undefined: caseFS
internal/cli/migrate_ownership_case_test.go:58:28: undefined: caseFS
internal/cli/migrate_ownership_case_test.go:59:9: undefined: caseFS
internal/cli/migrate_ownership_case_test.go:100:68: undefined: caseFS
internal/cli/migrate_ownership_case_test.go:114:3: unknown field caseAlias in struct literal of type ownershipEnv
internal/cli/migrate_ownership_case_test.go:115:11: undefined: resolveCaseAlias
internal/cli/migrate_ownership_case_test.go:134:8: v.caseVariantOf undefined (type ownershipVerdict has no field or method caseVariantOf)
internal/cli/migrate_ownership_case_test.go:135:77: v.caseVariantOf undefined (type ownershipVerdict has no field or method caseVariantOf)
internal/cli/migrate_ownership_case_test.go:137:8: v.key undefined (type ownershipVerdict has no field or method key)
internal/cli/migrate_ownership_case_test.go:137:8: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

종료 코드 1. 이유: 파일시스템 모드를 주입하는 별칭 확정 함수와 그 판정 필드가 아직 없어서 시험 패키지가 빌드되지 않는다 (시험이 존재하는 상태의 빌드 실패이며 제로 매치가 아니다). 이 RED 는 빌드 실패 형이다: 이 두 시험의 단언 수준 실패는 이 시점에 관측되지 않는다 (호스트 쪽만 위 셀에서 단언 수준으로 관측). 변이 앞의 직접 관측은 C 커밋 GREEN 원문의 변이 서브 시험이 한다.

### Run chunk 2 GREEN 기록과 정직한 회계 (M2 + M3 + M4)

GREEN 셀의 트리: AC 판정 명령, 기준선·chunk 1 회귀, 61시험 묶음은 모두 트리 `91e5bbcfa5` (C11, 작업 트리 깨끗 — `git status --short` 빈 출력) 에서 쟀다. 그 뒤에 더해진 커밋 `abd9bbda34` 는 시험 파일 하나(`migrate_ownership_coverage_test.go`)뿐이어서 비시험 코드는 같고, 그 시험의 실행과 품질 측정(빌드, 윈도 빌드, gofmt, vet, 린트)은 `abd9bbda34` 에서 다시 쟀다. 판정 도구는 이 트리의 `go test` (Go 1.26.8, darwin/arm64), 린트는 설치된 golangci-lint v2.1.6 (이 저장소 도구가 아니라 외부 린터이며, 이 트리의 소스를 읽은 결과이다). 출력에서 `=== RUN` 줄과 설정 알림 한 줄(`SUNSET_CONFIG_DORMANT_NOTICE`)은 뺐고 나머지는 원문이다. `coverage: N% of statements` 줄은 `-coverprofile` 을 건 좁은 실행의 **패키지 전체 대비** 비율이라 의미 있는 측정이 아니다 (새 코드의 함수별 측정은 아래).

**커밋 지도 (기준 HEAD `5eed5816ae`)**

| 묶음 | A (시험만) | B (RED 셀) | C (구현) | 비고 |
|------|-----------|-----------|----------|------|
| M2 AC-OWR-07 | `73ace0cc2e` | `36f638ef0a` | `cca0c9ead1` | RED 는 빌드 실패 형 |
| M2 AC-OWR-08 | `aa8037a7f2` | `ae4032950e` | — | 가드: 도착 시 GREEN, 붉음은 변이로 관측 (구현 커밋 없음) |
| M3 AC-OWR-14 | `167a6c0af4` | `db68f30cf7` | `e6e3e8d9f2` | RED 는 단언 수준 (정직한 회계 3 의 시험 보정 포함) |
| M4 AC-OWR-11 호스트 쪽 | `5246f408cf` | `cbf9579e9a` | `91e5bbcfa5` | RED 는 단언 수준 |
| M4 AC-OWR-11 주입 쪽, AC-OWR-12 | `ae8868f6fc` | `cbf9579e9a` | `91e5bbcfa5` | RED 는 빌드 실패 형 |
| 특성화 시험 (C 뒤) | `abd9bbda34` | — | — | 시험만 |

**승격 상태 (acceptance.md 「등급과 RED 셀 규약」)**: AC-OWR-07, -11, -12, -14 는 A → B → C 순서를 마쳤고 (커밋 그래프가 순서를 증명) RED 셀 네 요소 (명령, 원문 stdout, 종료 코드, 트리 SHA) 와 이유가 있다. 단 AC-OWR-07 과 AC-OWR-11 주입 쪽 / AC-OWR-12 의 RED 는 빌드 실패 형이다 (단언 수준의 붉음은 GREEN 출력의 `Mutant...IsCaught` 서브 시험과 아래 변이 점검 표가 맡는다). AC-OWR-08 은 도착 시 GREEN 인 가드라서 「RED-now」 가 깨끗한 트리에는 없다: 그 붉음의 원문은 실제 운영 코드에 임시로 끼운 변이 R1, R2 의 셀과 시험 안 변이 서브 시험이다. release-blocking 으로 칠지 regression-guard 로 칠지의 분류는 감사의 몫이며 여기서는 사실만 적는다.

**AC-OWR-07 GREEN (트리 `91e5bbcfa5`)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestProfileMappedCopyKeptWithoutCounterpart$' -count=1 -v -coverprofile=...`

```text
--- PASS: TestProfileMappedCopyKeptWithoutCounterpart (9.93s)
    --- PASS: TestProfileMappedCopyKeptWithoutCounterpart/NoCounterpart (0.02s)
    --- PASS: TestProfileMappedCopyKeptWithoutCounterpart/ConfirmedCounterpartStillKept (2.88s)
    --- PASS: TestProfileMappedCopyKeptWithoutCounterpart/DifferingCounterpartIsUnconfirmed (2.75s)
    --- PASS: TestProfileMappedCopyKeptWithoutCounterpart/ReachControlUnmappedTwinRemoves (1.57s)
    --- PASS: TestProfileMappedCopyKeptWithoutCounterpart/MutantIgnoringTheMappingIsCaught (2.68s)
    --- PASS: TestProfileMappedCopyKeptWithoutCounterpart/MutantPretendingToReadTheProfileSideIsCaught (0.03s)
PASS
coverage: 5.3% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli	12.624s	coverage: 5.3% of statements
exit=0
```

**AC-OWR-08 GREEN (트리 `91e5bbcfa5`)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestProfileEnvDoesNotReachProfileRoot|TestNoProfileRootAccessInUpdateSources|TestProfileScanCatchesMutants)$' -count=1 -v -coverprofile=...`

```text
--- PASS: TestNoProfileRootAccessInUpdateSources (12.18s)
    --- PASS: TestNoProfileRootAccessInUpdateSources/ScannedSetIsTheGoListClosure (4.23s)
    --- PASS: TestNoProfileRootAccessInUpdateSources/NoNewCallSiteAgainstTheBaseline (2.26s)
    --- PASS: TestNoProfileRootAccessInUpdateSources/MigrationPreviewAndVerdictFilesAddNothing (1.49s)
--- PASS: TestProfileScanCatchesMutants (15.66s)
    --- PASS: TestProfileScanCatchesMutants/UnmutatedTreeIsClean (2.27s)
    --- PASS: TestProfileScanCatchesMutants/M-a_read-and-discard_through_an_exported_helper (3.38s)
    --- PASS: TestProfileScanCatchesMutants/M-b_spliced_token_read_of_the_config-dir_variable (1.93s)
    --- PASS: TestProfileScanCatchesMutants/M-c_a_second_os.Stat_in_launchCandidateIsUsable (1.85s)
    --- PASS: TestProfileScanCatchesMutants/M-d_a_new_function_in_internal/profile_that_lists_a_directory (2.23s)
    --- PASS: TestProfileScanCatchesMutants/M-e_a_new_file_calling_an_exported_profile_helper (2.89s)
    --- PASS: TestProfileScanCatchesMutants/M-f_an_environment_read_with_a_non-constant_name (1.12s)
--- PASS: TestProfileEnvDoesNotReachProfileRoot (23.33s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/NormalizationRules (0.00s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/ControlTwoRunsOfTheSameStateAreIdentical (8.78s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/ThreeProfileStatesLeaveTheSameRecord (8.13s)
    --- PASS: TestProfileEnvDoesNotReachProfileRoot/MutantReadingTheCanaryIsSeen (6.41s)
PASS
coverage: 7.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli	53.719s	coverage: 7.1% of statements
exit=0
```

**AC-OWR-14 GREEN (트리 `91e5bbcfa5`)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$' -count=1 -v -coverprofile=...`

```text
--- PASS: TestRemainingListKeeps313UnregisteredCopies (52.38s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/Premise (0.44s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/StructuredRowsKeepEveryCopy (4.59s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MutantEmptyReasonRowsIsCaught (4.02s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MutantPrintingEveryPathByDefaultIsCaught (0.58s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MutantBytesEqualPromotionIsCaught (30.89s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (9.15s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/MixedReasonsEveryRowHasOne (2.70s)
PASS
coverage: 7.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli	54.073s	coverage: 7.9% of statements
exit=0
```

**AC-OWR-11 GREEN (트리 `91e5bbcfa5`)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestCaseVariantResolvesToRegisteredEntry|TestCaseVariantOnHostFilesystem)$' -count=1 -v -coverprofile=...`

```text
--- PASS: TestCaseVariantOnHostFilesystem (5.57s)
    --- PASS: TestCaseVariantOnHostFilesystem/the_host's_own_case_mode_decides_the_expectation (4.05s)
    --- PASS: TestCaseVariantOnHostFilesystem/the_operational_code_never_runs_the_project-root_probe (1.53s)
--- PASS: TestCaseVariantResolvesToRegisteredEntry (0.03s)
    --- PASS: TestCaseVariantResolvesToRegisteredEntry/case-insensitive_mode:_one_file,_the_registered_verdict (0.01s)
    --- PASS: TestCaseVariantResolvesToRegisteredEntry/case-insensitive_mode:_a_user-modified_registered_file_stays_user-modified (0.00s)
    --- PASS: TestCaseVariantResolvesToRegisteredEntry/case-sensitive_mode:_two_files (0.00s)
    --- PASS: TestCaseVariantResolvesToRegisteredEntry/the_exact_registered_spelling_is_looked_up_first (0.00s)
    --- PASS: TestCaseVariantResolvesToRegisteredEntry/mutant:_an_always-alias_resolver_is_caught (0.01s)
    --- PASS: TestCaseVariantResolvesToRegisteredEntry/an_ambiguous_alias_is_no_alias (0.00s)
PASS
coverage: 5.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli	8.379s	coverage: 5.1% of statements
exit=0
```

**AC-OWR-12 GREEN (트리 `91e5bbcfa5`)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCaseVariantCounterpartAndMirrorSpelling$' -count=1 -v -coverprofile=...`

```text
--- PASS: TestCaseVariantCounterpartAndMirrorSpelling (4.68s)
    --- PASS: TestCaseVariantCounterpartAndMirrorSpelling/case-insensitive:_the_spelling_does_not_make_the_counterpart_unconfirmed (2.92s)
    --- PASS: TestCaseVariantCounterpartAndMirrorSpelling/case-sensitive_control:_the_same_project_reports_an_unconfirmed_counterpart (1.75s)
PASS
coverage: 5.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli	6.289s	coverage: 5.2% of statements
exit=0
```

**기준선 11개 + chunk 1 의 22개 회귀 (트리 `91e5bbcfa5`, 앵커된 `-run`, 최상위 `--- PASS` 만 추림)** — 쓸어낸 최상위 시험 수 33 (= 11 + 22)

```text
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (2.31s)
--- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence (0.24s)
--- PASS: TestOwnershipHelperTreeSnapshotSeesDirMtime (0.04s)
--- PASS: TestOwnershipHelperForcedRunReachesReconciliation (6.65s)
--- PASS: TestOwnershipHelperUsesRealEmbeddedPaths (2.19s)
--- PASS: TestProfileCallsiteBaselineMatchesTree (13.39s)
--- PASS: TestOwnershipHelperFixtureBuilders (0.12s)
--- PASS: TestOwnershipVerdictClasses (0.04s)
--- PASS: TestOwnershipPreserveByteEqualUserFile (8.86s)
--- PASS: TestOwnershipPreserveUserModifiedFile (6.76s)
--- PASS: TestOwnershipUnregisteredUnknownOriginKeptWithReason (6.12s)
--- PASS: TestKeepReasonVocabularyIsClosed (0.00s)
--- PASS: TestPreservePriorManifestOutcomes (0.05s)
--- PASS: TestMigrationKeepsSharedCopyUnderProfileReach (2.26s)
--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach (2.00s)
--- PASS: TestMigrationReachLeavesUserModifiedCopy (2.39s)
--- PASS: TestMigrationBackupBeforeRemoval (7.27s)
--- PASS: TestEmptiedSkillDirRemoved (6.80s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (3.20s)
--- PASS: TestMigrationRemovalGate (3.40s)
--- PASS: TestRF2F3b_MigrationPreservesUntracked (0.01s)
--- PASS: TestMigrationIdempotent (3.22s)
--- PASS: TestMigrationPreservesExistingMirrorEntries (3.60s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.00s)
--- PASS: TestPriorManifestAtomicWriteFailures (0.06s)
--- PASS: TestPriorGenerationCorruptIsNoEvidence (0.07s)
--- PASS: TestLegacyManifestRoundTrip (0.01s)
--- PASS: TestDowngradeDropsEvidenceNotSafety (0.04s)
--- PASS: TestPriorGenerationsAreLookupBoundOnly (0.67s)
--- PASS: TestPriorManifestRetainedInRunBackup (10.73s)
--- PASS: TestPriorGenerationSourceBoundedAndTolerant (0.02s)
--- PASS: TestBackupFailureRemovesNothingThroughProductionPath (13.68s)
--- PASS: TestOwnershipAttestedManagedThroughOperatingSource (0.04s)
PASS
coverage: 8.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli	108.166s	coverage: 8.1% of statements
exit=0
```

**마이그레이션·갱신 61시험 묶음 (chunk 1 과 같은 파일 목록의 모든 최상위 시험을 앵커된 `-run` 으로, 트리 `91e5bbcfa5`)** — 쓸어낸 최상위 시험 수 61 (기대 61). 이 묶음은 이번 chunk 가 갱신 흐름의 표준 출력(마이그레이션의 파일별 보류 줄을 `--verbose` 없이는 내지 않음)을 바꿨으므로 회귀 가드로 다시 돌렸다.

```text
--- PASS: TestFR3_1_MigrationPreservesOnUnverifiable (0.01s)
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (5.52s)
--- PASS: TestDoctorProjectManifestLoadFailureNotDisguised (0.00s)
--- PASS: TestDoctorProjectManifestPreservedOnLoadFailure (0.00s)
--- PASS: TestDoctorUserInstallHonestFailureAndReadOnly (0.00s)
--- PASS: TestDoctorLockCheckVacuousNotReportedMatch (0.01s)
--- PASS: TestDoctorWorkflowRootSelectedByRequiredFile (0.02s)
--- PASS: TestDoctorFallbackKeepsProjectScopeL1 (0.02s)
--- PASS: TestMigrationKeepsSharedCopyUnderProfileReach (6.11s)
--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach (5.21s)
--- PASS: TestMigrationReachLeavesUserModifiedCopy (5.15s)
--- PASS: TestFileEntryKeptBy (0.00s)
--- PASS: TestMigrationBackupBeforeRemoval (7.01s)
--- PASS: TestEmptiedSkillDirRemoved (4.88s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (3.42s)
--- PASS: TestMigrationRemovalGate (3.45s)
--- PASS: TestUpdateDryRun_EmitsCleanReinstallPlan (0.01s)
--- PASS: TestUpdateDryRun_EmitsResidueCleanupPlan (0.00s)
--- PASS: TestUpdateDryRun_ZeroMutation (0.02s)
--- PASS: TestUpdateDryRun_PreservesExistingOutput (0.01s)
--- PASS: TestPreviewPreservedLinesCountByDefault (0.01s)
--- PASS: TestRetrackManifestFilesResolvesDriftAndKeepsTemplateHash (0.00s)
--- PASS: TestRetrackManifestFilesLeavesUserOwnedEntriesAlone (0.00s)
--- PASS: TestRetrackManifestFilesSkipsAbsentAndUntracked (0.00s)
--- PASS: TestRetrackSectionFilesCoversSectionDirectory (0.00s)
--- PASS: TestRetrackSectionFiles_FailureReturnsErrorAndPrintsNothing (0.01s)
--- PASS: TestRunUpdateRestoreRetracksSections (0.06s)
--- PASS: TestUpdateMigratesLegacyProject (6.18s)
--- PASS: TestUpdateOptedOutMigrationPreservesManagedRootLocals (1.64s)
--- PASS: TestMigrationArchivesModifiedBeforeRemoval (0.05s)
--- PASS: TestMigrationIdempotent (3.18s)
--- PASS: TestUpdateForceKeepsPlantedCommonAssets (2.22s)
--- PASS: TestUpdateLocalModeKeepsFullScope (1.10s)
--- PASS: TestUpdateNeverFlipsModeRecord (2.86s)
--- PASS: TestForceUpdatePreservesUserCommonAssets (2.16s)
--- PASS: TestNotDemonstratedPreservationHoldsAtNextLocalUpdate (2.71s)
--- PASS: TestPreservedSectionSurvivesRestoreUntouched (1.06s)
--- PASS: TestMigrationArchiveRefusesSymlinkedArchiveDestination (0.22s)
--- PASS: TestUnknownSkillPreservedAcrossLocalUpdates (3.21s)
--- PASS: TestMigrationPreservesExistingMirrorEntries (3.07s)
--- PASS: TestUpdatePreservesUserModifiedPolicies (1.78s)
--- PASS: TestCommonAssetCleanFilterNormalizesSeparators (0.00s)
--- PASS: TestMigrationRefusesRemovalWhenUserInstallMissing (1.72s)
--- PASS: TestUserCounterpartMustBeCurrentVersion (0.01s)
--- PASS: TestCancelledUpdateKeepsProjectManagedAssets (2.88s)
--- PASS: TestRecordlessBranchTallyCountsReconcileRemovals (1.49s)
--- PASS: TestAgentsSkillsCounterpartMapsToClaudeSource (0.01s)
--- PASS: TestReconcileDropsStaleManifestRecord (1.47s)
--- PASS: TestPostRemovalUserFileSurvivesNextUpdate (4.40s)
--- PASS: TestAgentsOriginSkillMapsToItsOwnCatalogSource (0.01s)
--- PASS: TestPreviewFlagsMigrationRemovalCandidates (0.01s)
--- PASS: TestPreviewFlagsMigrationRemovalCandidates_VerboseListsControl (0.02s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.01s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy (0.01s)
--- PASS: TestTallyCountsMigrationRemovals (4.46s)
--- PASS: TestSkipDecisionUsesTreeHash (0.17s)
--- PASS: TestPreviewCountEqualsRunCount (0.02s)
--- PASS: TestTemplateSync_SymlinkedNamespaceBackupRootReportsSkippedPruning (6.55s)
--- PASS: TestVersionMatchSkipKeepsCorruptManifestInPlace (0.02s)
--- PASS: TestUpdateSkippedOnVersionMatchRejectsChangedTemplateTree (0.55s)
--- PASS: TestStampSkewWarning (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	97.831s
exit=0
```

**특성화 시험 (트리 `abd9bbda34`)**: `TestProfileReachMarkUsesTheRegisteredKeyUnderACaseAlias`, `TestAliasToAKeyTheManifestLacksIsNoAlias`, `TestCounterpartLocationAndStatusEdges`, `TestRemainingRenderEdges` — 모두 PASS (명령: `go test ./internal/cli/ -run '^(...)$' -count=1 -v`, 종료 코드 0). 첫 시험은 프로필 도달 표지를 **등록된 키** 로 쓰는 가지를 지킨다: 변이 (디스크 철자로 쓰기) 를 끼우면 마이그레이션이 `mark profile-reach keep ... entry not found` 로 중단되어 붉어짐을 관측했다.

**품질 측정 (트리 `abd9bbda34`)**

```text
HEAD=abd9bbda3410cc58a954cbd95b77f2352e3b7b0d
$ go build ./...
exit=0
$ GOOS=windows GOARCH=amd64 go build ./...
exit=0
$ gofmt -l internal/cli internal/manifest
exit=0
$ go vet ./internal/cli/ ./internal/manifest/
exit=0
$ golangci-lint run ./internal/manifest/ ./internal/cli/
internal/cli/factory_run_adopt.go:79:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
			fmt.Fprintf(warn, "factory leader resume: run %s (owner %s) still holds live records — adopting nothing; to adopt one deliberately, relaunch with `moai cc -f --factory-run %s`\n", c.RunID, c.Classification, c.RunID)
			           ^
internal/cli/factory_run_adopt.go:91:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
		fmt.Fprintf(warn, "factory leader resume: run %s is led by codex — a claude leader does not adopt it; retire it first or relaunch the codex leader deliberately\n", adopt.RunID)
		           ^
internal/cli/factory_run_adopt.go:103:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
		fmt.Fprintf(warn, "factory leader resume: run %s holds live legacy record %q from a binary before the leader/lane rename — adopting nothing; end its sessions and retire with 'moai factory runs --retire %s' first\n", adopt.RunID, value, adopt.RunID)
		           ^
3 issues:
* errcheck: 3
exit=1
```

린트 3건은 모두 `internal/cli/factory_run_adopt.go` 의 기존 `errcheck` (chunk 1 과 같은 3건; 그 파일은 기준 `cdd72a814` 와 바이트 동일 — `git diff --quiet cdd72a814 HEAD -- internal/cli/factory_run_adopt.go` → 종료 코드 0) 이고 이 카드가 만든 파일의 이슈는 0건이다. 범위 규율로 고치지 않았다.

**파일 불변·경계 확인 (트리 `abd9bbda34`)**

```text
HEAD=abd9bbda3410cc58a954cbd95b77f2352e3b7b0d
$ git log --oneline cdd72a814..HEAD --grep=(card t1865) -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
exit=0
$ git log --oneline cdd72a814..HEAD --grep=(card t1865) -- .moai/specs | wc -l (positive control: at least one line)
      15
$ git log --oneline cdd72a814..HEAD --grep=(card t1865) -- internal/cli/update/backup
exit=0
$ grep -rn --include='*.go' --exclude='*_test.go' 'removalEnabled\s*=' internal
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
exit=0
$ grep -n 'AskUserQuestion' <the five new non-test files>
exit=1
$ git diff --stat 5eed5816ae HEAD -- internal ":(exclude)*_test.go"
 internal/cli/migrate_ownership.go           | 181 +++++++++++++++++++++++++++-
 internal/cli/migrate_ownership_case.go      | 149 +++++++++++++++++++++++
 internal/cli/migrate_ownership_remaining.go | 155 ++++++++++++++++++++++++
 internal/cli/migrate_project_assets.go      |  29 ++++-
 internal/cli/update_template_sync.go        |   5 +-
 5 files changed, 508 insertions(+), 11 deletions(-)
```

보호 파일 둘과 `internal/profile` 은 이 카드의 커밋이 건드리지 않았다 (명령 2 빈 출력·종료 코드 0, 양성 대조군 15줄). `removalEnabled` 의 비시험 대입은 선언 한 줄뿐이다 (`= false`, `internal/cli/migrate_project_assets.go:40`). 서브에이전트 경계 grep 은 일치 없음 (종료 코드 1).

**새 코드의 함수별 커버리지 (이 chunk 의 AC 시험 + 회귀 묶음의 `-coverprofile` 을 합친 값, `go tool cover -func`; 패키지 전체를 돌리지 못하는 제약 때문에 패키지 단위 비율은 재지 못했다)**: 파일별 문장 커버리지 `migrate_ownership.go` 96.0% (119/124), `migrate_ownership_case.go` 96.6% (57/59), `migrate_ownership_remaining.go` 100% (64/64), `migrate_ownership_evidence.go` 100% (25/25), `update_prior_manifest.go` 87.5% (70/80); 100% 아닌 새 함수: `classifyOwnershipCore` 85.7%, `userCounterpartStatus` 90.9%, `shippedCaseSpelling` 94.1%, `canonicalCounterpartRel` 93.3%. 수정한 기존 파일 `migrate_project_assets.go` 는 이 좁은 실행에서 82.5% (212/257) 이다 — 기존 코드의 가지 (`verifyMigrationFinalState` 63.2% 등) 를 좁은 묶음이 다 밟지 않기 때문이며 85% 기준 확인은 CI 의 전체 실행 몫이다 (Gaps). `update_template_sync.go` 는 호출 줄 둘만 바꿨고 파일 단위로 재지 않았다.

**변이 점검 (실제 운영 코드에 임시로 변이를 끼워 시험이 붉어지는지 관측, 매번 바이트 복사본으로 되돌리고 `cmp` 로 동일 확인; 이 표는 성립한 변이만 적었다)**

| 변이 (임시) | 붉어진 시험 |
|------|-------------|
| R1 `migrate_ownership.go` 에 `profile.GetProfileDir` 로 프로필 뿌리를 읽고 버리는 함수 | `TestNoProfileRootAccessInUpdateSources` (새 호출 지점 + 판정 코드 영 규칙) |
| R2 프로필 쪽 기록이 `CLAUDE_CONFIG_DIR/skills/x/SKILL.md` 의 존재에 좌우됨 | `TestProfileEnvDoesNotReachProfileRoot/ThreeProfileStatesLeaveTheSameRecord` (같은 상태 대조군은 PASS) |
| M2-1 대응물 상태를 행에 기록하지 않음 | `TestProfileMappedCopyKeptWithoutCounterpart` 의 세 사용자 루트 상태 서브 시험 |
| M2-2 프로필 쪽 상태를 항상 `n/a` 로 기록 | 같은 세 서브 시험 |
| M3-1 잔존 행 정렬 제거 | `StructuredRowsKeepEveryCopy`, `RenderedByTheUpdateFlow` (경로 사전순) |
| M3-2 파일별 보류 줄을 항상 표준 출력으로 | `RenderedByTheUpdateFlow` (기본 표준 출력에 313개 경로) |
| M3-3 잔존 목록을 터미널 블록에 기록하지 않음 | `RenderedByTheUpdateFlow` (이유별 줄 없음, 파일별 행 0) |
| M3-4 기본 보기에도 파일별 행을 모두 기록 | `RenderedByTheUpdateFlow` (기본 블록에 313개 경로, 324줄) |
| M4-1 별칭이 파일 동일성을 보지 않음 | `TestCaseVariantResolvesToRegisteredEntry/case-sensitive_mode:_two_files` |
| M4-2 대응물 확인에서 대소문자 해소 제거 | `TestCaseVariantCounterpartAndMirrorSpelling/case-insensitive...` (대응물 `unconfirmed`) |
| M4-3 출하된 철자를 보고하지 않음 | 같은 서브 시험 (`case variant of ""`) |
| M4-4 실제 마이그레이션에서 별칭 조회 생략 | `TestCaseVariantOnHostFilesystem` (호스트: `unregistered`) |
| 프로필 도달 표지를 디스크 철자 키로 씀 | `TestProfileReachMarkUsesTheRegisteredKeyUnderACaseAlias` |
| 시험 안 변이: 스캔 M-a ~ M-f, 카나리 읽는 분류기, 항상 별칭, 비운 이유 열, 모든 경로 출력, 바이트 일치 승격 대역, 프로필 지도 무시, 프로필 쪽을 읽은 척 | 각각의 `Mutant...` / `M-x` 서브 시험이 비어 있지 않은 문제 목록을 직접 관측 |

**정직한 회계 (RED → GREEN 의 어긋남을 숨기지 않는다)**

1. AC-OWR-07 의 A 커밋은 새 어휘 (`counterpartStatus`, `ProfileSide` 필드) 를 요구해 RED 가 빌드 실패이다. 단언 수준의 「`absent` 가 기록되지 않음」 은 이 시점에 관측되지 않았다. 그 몫은 C 커밋 GREEN 출력의 변이 서브 시험과 변이 점검 표의 M2-1, M2-2 가 한다.
2. AC-OWR-08 의 시험은 처음 쓴 초안이 두 곳에서 틀렸고 A 커밋 **전에** 디버그 실행으로 고쳤다 (그래서 A 커밋은 최종 모양이다): (a) 「마이그레이션·미리보기·판정·갱신 동기화 파일」 을 acceptance.md 의 `migrate_*.go` 글롭으로 잡았더니 `moai migrate` 하위 명령의 파일 (`migrate_profiles.go`, `migrate_home_state.go` 등) 이 기준선 행을 이미 가지고 있어 붉었다 — 이 SPEC 이 더하거나 고친 파일 (`migrate_ownership*.go`, `update_prior_manifest.go`, `update_template_sync.go`, `update_dryrun_preview.go`) 은 프로필 호출 지점 0, `migrate_project_assets.go` 는 기준선의 한 행 정확히 하나로 좁혔다. (b) 같은 상태 두 번 구동의 **대조군** 이 정규화 부족을 드러냈다: 파일 시스템 안전 철자의 타임스탬프 (`update-2026-10-18T06-31-53Z`) 와 경로 키 안의 실행 디렉터리 이름이 지워지지 않았다. 정규화는 여전히 세 종류 (임시 루트, 실행 디렉터리 이름, 타임스탬프) 뿐이고 위 두 철자를 그 종류에 포함시켰다.
3. AC-OWR-14 의 시험은 C 커밋에서 한 번 보정했다: 실행 2 (`--force`) 에서 313개 중 두 개 (`.agents/skills/moai-foundation-core/SKILL.md`, `.agents/skills/moai-workflow-tdd/SKILL.md`) 의 이유가 `no-provenance-record` 가 아니라 `evidence-incomplete:original-deployed-path` 로 나왔다. 실행 1 의 시작 매니페스트가 이전 세대로 보존되고 거기에 같은 해시의 `.claude/skills` 등록 기록이 있기 때문이며, design.md D-2 가 정확히 그렇게 읽는다 (「같은 해시의 기록이 다른 키에만 있으면 original-deployed-path 가 빠진 것」). 구현이 아니라 시험의 기대가 틀렸다. 단언은 그 두 경로에 한해 좁은 이유를 기대하게 바뀌었고 나머지 311개는 `no-provenance-record` 그대로이다. A 커밋의 RED 이유 네 개 (정렬, 기본 출력 313 경로, 이유별 줄 없음, 파일별 행 없음) 는 이 보정과 무관하다.
4. AC-OWR-14 의 구현은 첫 판에서 chunk 1 의 `TestBackupFailureRemovesNothingThroughProductionPath` 를 붉게 했다: 흐름의 표준 출력이 `backup-failed` 를 이름 붙여야 하는데 파일별 보류 줄을 기본에서 빼면서 그 줄도 사라졌다. 실패는 백업 실패로 제거를 거부한 줄이 정상 상태의 보류가 아니라 운영자가 봐야 할 실패라서 예외로 두는 것으로 고쳤다 (`[reason: backup-failed]` 태그가 붙은 줄은 기본에서도 표준 출력으로 나간다; 개수는 그 실행이 제거할 파일 수로 한정). 이후 61시험 묶음과 chunk 1 회귀가 모두 GREEN.
5. AC-OWR-11 의 A 커밋을 둘로 나눴다 (호스트 쪽 `5246f408cf`, 주입 쪽 `ae8868f6fc`): 새 API 를 쓰는 시험과 같은 커밋에 두면 호스트 쪽의 단언 수준 RED 가 빌드 실패에 가려진다. 호스트 쪽은 현재 식별자만 쓰도록 쓰고 철자 차이 보고는 보고 줄의 `case-variant` 문자열로 단언했다.
6. 임시 변이는 `cp` 로 원본 복사본을 되돌리고 `cmp` 로 바이트 동일을 확인했다. 한 스크립트가 끝에 `git diff --quiet` (HEAD 대비) 로 `DIRTY` 를 적었는데, 그 시점에 해당 파일이 HEAD 보다 앞선 미커밋 구현이라서였고 복원 실패가 아니다 (같은 시점의 `cmp` 가 `BYTE-IDENTICAL-RESTORED`). 변이 M4-4 의 첫 형태는 미사용 변수·import 로 컴파일되지 않아 변이로 성립하지 않았고 (표에는 두 번째 형태만) 다시 만들어 돌렸다.
7. 측정 중 이 머신의 부하 평균이 약 65 였다 (`uptime`). `go test` 호출 몇 건이 도구의 120초 제한을 넘어 분리 실행했다. 시간은 증거로 쓰지 않는다.
8. AC-OWR-08 의 카나리는 외곽 진입 `runUpdateFlowForced` 가 아니라 `runTemplateSyncWithReporter` (`--force` 플래그 설정, 마이그레이션 실행 문맥을 시험이 쥠) 를 구동한다. 잔존 행·제거 목록을 구조로 관측하려는 선택이며, 버전 일치 건너뛰기 관문은 외곽 진입에만 있어 이 진입은 항상 조정에 닿는다. 규칙 7 의 (ii) 새 `manifest.prior.json` 의 바이트 == 실행 전 매니페스트, (iii) 잔존 행 양수 와 프로필 지도 사본 행 존재는 매 실행 단언한다.

**이 chunk 가 만진 부채의 처분 (Binding run conditions)**

| 부채 | 처분 | 근거 |
|------|------|------|
| N16 프로필 뿌리 읽기를 실행 시점 조각으로 조립한 경로가 AC-OWR-08 을 피함 | **열린 채 공개** (우회하지 않음) | 이 chunk 는 프로필 뿌리를 열거나 읽는 코드를 더하지 않았다: 호출 지점 스캔의 새 행 0, 판정·보고·동기화·미리보기 새 코드의 프로필 호출 지점 0 (`MigrationPreviewAndVerdictFilesAddNothing`), 보호 파일 둘과 `internal/profile` 불변 (L4). 조각 조립 + 결과 버림 읽기는 L1 도 L3 도 못 본다는 한계는 그대로이며 카드 t1702 보류 범위다. 프로필 쪽 파일을 읽어야만 충족되는 요구는 나오지 않았다 (Q5 중단 조건 미발동). |
| N21 중 REQ-OWR-005 문구 (「동결된 기존 접촉들」 대 「단일 기존 접촉」) | SPEC 문구 결함으로 **공개, 본문 미수정** | 리더 결정 3. 시험은 한 행·개수 1 의 예외를 기준선 행으로 단언한다. |
| N17 (M9), N23·N4 잔여 (M8) | 이월 (이 chunk 의 범위 밖) | M8, M9 를 맡는 chunk 로 |

**SPEC 쪽 결함·불일치로 보고하는 것 (고치지 않았다)**

- AC-OWR-08 이 「마이그레이션·미리보기·판정·갱신 동기화 파일 (`migrate_*.go`, ...)」 의 새 호출 지점이 0 이라고 쓰는데, `migrate_*.go` 글롭은 `moai migrate` 하위 명령의 파일도 포함하고 그 파일들은 기준선에 프로필 행을 가진다. 시험은 이 SPEC 의 새 파일·고친 파일로 좁혔다.
- AC-OWR-14 는 313개의 이유를 `no-provenance-record` 로 적는다. 첫 실행에서는 그렇다. 그러나 이전 세대 (`manifest.prior.json`) 가 쌓인 뒤에는 같은 해시의 `.claude/skills` 기록이 있는 복사본의 이유가 `evidence-incomplete:original-deployed-path` 가 된다 (design.md D-2 의 결과). 현장 313개 중 `.claude/skills` 짝이 매니페스트에 등록된 만큼은 두 번째 실행부터 그 이유로 바뀔 수 있다 — 보존이라는 결과는 같다. 리더가 알아야 하는 현장 가시 효과이다.
- AC-OWR-12 의 「미등록 복사본이 `case-variant` 이유를 받는다」 는 무엇에 대한 변형인지 적지 않았다. 미등록 파일에는 등록 키가 없으므로 **출하된 (내장 템플릿) 철자** 에 대한 변형으로 구현했다 (`remainingRow.CaseVariantOf`). 이유 집합 (design.md D-1) 의 `case-variant` 는 닫힌 이유가 아니라 표지로 둬서 `TestKeepReasonVocabularyIsClosed` 의 닫힌 집합을 바꾸지 않았다.
- AC-OWR-14 의 「기본 렌더는 개별 경로를 찍지 않으며」 는 갱신 흐름이 마이그레이션의 파일별 보고 줄을 기본에서 이미 313줄 찍고 있었다는 측정 (A14 의 RED) 으로 흐름 자체의 표준 출력까지 포함하게 읽었다. 이를 위해 갱신 흐름의 보고 콜백이 `--verbose` 없이는 파일별 보류 줄을 내지 않도록 호출 줄 하나를 바꿨다 (사용자 눈에 보이는 동작 변화: 보류 파일의 이름은 `moai update --verbose` 에서 터미널 블록의 Reference 줄로 나오고, 기본에서는 이유별 개수만 나온다).

**Gaps (관측하지 않은 것)**

- `internal/cli` 패키지 전체, `-race`, 저장소 전체 시험은 돌리지 않았다 (lane 제약). 저장소 전체 시험 판정은 통합 브랜치의 CI 가 소유하며 보고 시점에 **보류** 이다.
- 윈도·리눅스에서는 컴파일만 확인했다 (`GOOS=windows GOARCH=amd64 go build ./...`). 새 시험은 darwin 에서만 실행했다. 대소문자 시험은 두 모드를 흉내 낸 주입으로 어느 호스트에서나 같은 결과를 주지만, 호스트 시험의 **구분 모드 가지** (리눅스) 는 이 머신(무구분)에서 실행되지 않았다.
- 새 코드가 아닌 기존 파일 `migrate_project_assets.go` 의 파일 단위 커버리지는 좁은 묶음에서 82.5% 이다 (위). 85% 기준은 CI 의 전체 실행으로 확인해야 한다.
- 접근 기록 시임이 없어 「프로필 뿌리를 읽지 않았다」 는 호출 지점 동결 + 차분 카나리 + 변이 관측의 추론이다 (한계 (1)(3) 그대로; N16).
- 린트는 설치된 `golangci-lint` v2.1.6 이며 이 트리에서 빌드한 도구가 아니다.
- 시험 시간(부하 평균 약 65)은 증거가 아니다.
- 카나리의 구동 진입이 외곽 진입이 아니다 (정직한 회계 8).

**잔여 위험**

- 호출 지점 스캔의 한계 (조각난 경로 + 버림 읽기) 는 열려 있다 (N16, 카드 t1702 보류).
- 갱신 흐름의 표준 출력 변화 (기본에서 보류 파일별 줄이 사라짐) 를 기대하는 다른 소비자 (스크립트·문서) 가 있다면 영향을 받는다. 이 저장소 시험 중에는 61시험 묶음과 chunk 1 회귀에 그런 기대가 없었다.
- 현장의 313개는 두 번째 실행부터 일부가 다른 (더 좁은) 보류 이유로 나올 수 있다 (위 SPEC 쪽 불일치).
- `hostCaseFS` 는 패키지 변수라 병렬 시험이 바꾸면 서로 간섭한다. 이 chunk 의 시험은 `t.Parallel()` 을 쓰지 않고 `t.Cleanup` 으로 복원한다.

### Chunk 2 GREEN 정정: RenderedByTheUpdateFlow 간헐 실패 (lane 재검증, HEAD `644a7436bf`)

**관측 (lane).** `go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$' -count=1 -v` 를 연달아 세 번: 통과, 실패, 실패. 실패 서브 시험은 `RenderedByTheUpdateFlow`, 메시지는 `verbose view: a verbose row does not carry its reason "evidence-incomplete:original-deployed-path": 2 occurrence(s) ...` (실패한 실행에서는 두 복사본이 `no-provenance-record` 로 읽힌다). 위 chunk 2 GREEN 기록의 「한 번 통과」 는 결정성의 증거가 아니었다 — 이 정정이 그 기록을 대체하지 않고 덧붙는다.

**원인을 가르는 관측 (추측 아님).** 이 머신 (부하 평균 30-50) 에서 서브 시험만 앵커된 선택자로 `-count=14` 반복하면서, 임시로 (커밋하지 않음, 되돌림) 두 실행의 실행 백업 디렉터리 이름과 `manifest.prior.json` 개수를 시험 로그에 찍었다. 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$/^RenderedByTheUpdateFlow$' -count=14 -v` (트리 `644a7436bf` + 로그 두 줄). 원문 요약 (디렉터리 이름은 `DIAG` 로그 줄에서 그대로 옮김):

| # | 실행 1 의 실행 백업 | 실행 2 의 실행 백업 | 이름 | 프로젝트의 `manifest.prior.json` 개수 | 결과 |
|---|--------------------|--------------------|------|---------------------------|------|
| 1 | 20261018_165427 | 20261018_165428 | 다름 | 2 | PASS |
| 2 | 20261018_165429 | 20261018_165429 | 같음 | 1 | FAIL |
| 3 | 20261018_165431 | 20261018_165432 | 다름 | 2 | PASS |
| 4 | 20261018_165433 | 20261018_165434 | 다름 | 2 | PASS |
| 5 | 20261018_165435 | 20261018_165437 | 다름 | 2 | PASS |
| 6 | 20261018_165439 | 20261018_165439 | 같음 | 1 | FAIL |
| 7 | 20261018_165440 | 20261018_165441 | 다름 | 2 | PASS |
| 8 | 20261018_165442 | 20261018_165443 | 다름 | 2 | PASS |
| 9 | 20261018_165444 | 20261018_165444 | 같음 | 1 | FAIL |
| 10 | 20261018_165446 | 20261018_165446 | 같음 | 1 | FAIL |
| 11 | 20261018_165447 | 20261018_165448 | 다름 | 2 | PASS |
| 12 | 20261018_165449 | 20261018_165449 | 같음 | 1 | FAIL |
| 13 | 20261018_165451 | 20261018_165451 | 같음 | 1 | FAIL |
| 14 | 20261018_165452 | 20261018_165453 | 다름 | 2 | PASS |

14번 중 6번 실패, **실패한 실행은 예외 없이 두 실행이 같은 실행 백업 디렉터리를 쓴 실행이고 (`manifest.prior.json` 은 하나), 통과한 실행은 예외 없이 이름이 다른 두 디렉터리 (파일 둘) 였다.** 상관은 14/14 이다.

**원인.** 시험 쪽 시간 가정이다. 실행 백업 디렉터리 이름은 `time.Now().Format("20060102_150405")` (1초 해상도, `internal/cli/update/backup/backup.go:44`) 이고, `RenderedByTheUpdateFlow` 는 실행 1 과 실행 2 를 곧바로 이어서 돌린다. 두 실행의 백업 단계가 같은 시계 초에 걸리면 실행 2 는 실행 1 의 디렉터리를 재사용하고, `preservePriorManifest` 가 실행 2 자신의 시작 매니페스트 사본을 이름 바꾸기로 같은 이름 위에 놓아 실행 1 의 세대를 원자적으로 대체한다 (chunk 1 정직한 회계 3 이 이미 적은 동작). 증거 원천은 현재 실행의 디렉터리를 제외하므로 (`ps.exclude`) 읽을 이전 세대가 0 개가 되고, 같은 해시의 등록 기록이 보이지 않아 두 복사본은 `no-provenance-record` 로 남는다 (보존 방향, 승격 없음). 통과하는 경우 두 디렉터리가 있어 증거 원천이 실행 1 의 시작 매니페스트를 읽고 좁은 이유가 나온다. 제품 결함이 아님을 가르는 사실: (a) 디렉터리가 둘일 때 증거 원천은 세대를 항상 읽었다 (8번 모두 좁은 이유), (b) 디렉터리가 하나일 때 남은 파일은 실행 2 자신의 사본이라 제외 대상이며 읽을 세대가 애초에 없다 — 읽는 쪽의 누락이 아니다. 제품에 남는 사실은 **같은 초에 시작한 두 갱신은 이전 세대 하나를 잃는다**는 것이고 (잃는 방향은 증거 약화=보존 쪽, 제거 방향의 오류는 없음), 이는 `internal/cli/update/backup` 의 1초 해상도 디렉터리 이름이라는 기존 동작이며 이 카드가 건드리지 않기로 한 파일 (AC-OWR-21 (e)) 과 design.md D-3 항목 3 의 「정확한 이름만 읽는다」 규칙에 걸려 있어 여기서 고치지 않고 잔여 위험으로 공개한다.

**결정적 RED (커밋 A `814c6e044a`, 시험 파일만).** 시계를 기다리지 않고 같은 초 공유를 결정적으로 만든다: 실행 2 의 백업 단계가 `priorManifestWriter` 에 닿을 때 시험이 실행 1 의 사본을 그 디렉터리로 옮겨 놓고 진짜 쓰기 함수를 부른다 (공유 디렉터리의 종료 상태와 같다). 이 변형 `RenderedByTheUpdateFlowInOneSecond` 는 원래 기대 (두 복사본이 좁은 이유) 를 그대로 둬서 결정적으로 붉다 — 5번 중 5번, 메시지가 lane 의 것과 같다. 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$/^RenderedByTheUpdateFlowInOneSecond$' -count=5 -v` (트리 `814c6e044a`, 작업 트리 깨끗):

```text
    migrate_ownership_remaining_test.go:512: verbose view: a verbose row does not carry its reason "evidence-incomplete:original-deployed-path": 2 occurrence(s), first: · kept: .agents/skills/moai-foundation-core/SKILL.md — unregistered, no-provenance-record; missing evidence: deployed-version, original-deployed-path, prior-manifest-record
--- FAIL: TestRemainingListKeeps313UnregisteredCopies (2.94s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (2.94s)
    migrate_ownership_remaining_test.go:512: verbose view: a verbose row does not carry its reason "evidence-incomplete:original-deployed-path": 2 occurrence(s), first: · kept: .agents/skills/moai-foundation-core/SKILL.md — unregistered, no-provenance-record; missing evidence: deployed-version, original-deployed-path, prior-manifest-record
--- FAIL: TestRemainingListKeeps313UnregisteredCopies (2.85s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (2.85s)
    migrate_ownership_remaining_test.go:512: verbose view: a verbose row does not carry its reason "evidence-incomplete:original-deployed-path": 2 occurrence(s), first: · kept: .agents/skills/moai-foundation-core/SKILL.md — unregistered, no-provenance-record; missing evidence: deployed-version, original-deployed-path, prior-manifest-record
--- FAIL: TestRemainingListKeeps313UnregisteredCopies (1.94s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (1.94s)
    migrate_ownership_remaining_test.go:512: verbose view: a verbose row does not carry its reason "evidence-incomplete:original-deployed-path": 2 occurrence(s), first: · kept: .agents/skills/moai-foundation-core/SKILL.md — unregistered, no-provenance-record; missing evidence: deployed-version, original-deployed-path, prior-manifest-record
--- FAIL: TestRemainingListKeeps313UnregisteredCopies (1.44s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (1.43s)
    migrate_ownership_remaining_test.go:512: verbose view: a verbose row does not carry its reason "evidence-incomplete:original-deployed-path": 2 occurrence(s), first: · kept: .agents/skills/moai-foundation-core/SKILL.md — unregistered, no-provenance-record; missing evidence: deployed-version, original-deployed-path, prior-manifest-record
--- FAIL: TestRemainingListKeeps313UnregisteredCopies (2.35s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (2.35s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	12.924s
FAIL
exit=1
```

종료 코드 1.

**수정 (커밋 C `3d8045852c`, 시험 파일만 — 제품 코드 변경 없음).** 같은 초 공유는 `internal/cli/update/backup` 의 기존 동작이고 이 카드가 건드리지 않는 파일이므로 제품은 고치지 않았다. 시험이 시계 운에 기대던 두 가정을 없앴다, 단언은 약하게 하지 않고:

1. 실행 2 는 시작 전에 실행 1 의 실행 백업 이름이 가리키는 초보다 **뒤의 시계 초** 가 될 때까지 기다린다 (`awaitClockSecondAfter`: 고정 대기가 아니라 25ms 간격 폴링, 최대 5초, 이름이 시각이 아니거나 시계가 안 가면 큰 소리로 실패). 그러면 실행 2 의 백업이 실행 1 의 디렉터리를 쓸 수 없다.
2. 기대 이유가 「실행 1 의 세대가 보존되었는가」 라는 **관측된 사실** 을 따른다 (`priorGenerationRetained`: 현재 실행 자신의 것을 뺀 `manifest.prior.json` 중 실행 1 의 시작 매니페스트와 바이트가 같은 것이 있는가). 보존되었으면 두 복사본이 `evidence-incomplete:original-deployed-path`, 아니면 313개 모두 `no-provenance-record` 이고 프로젝트에 `manifest.prior.json` 이 정확히 하나다. 관측과 시험 변형 (`shareRunDirectory`) 이 어긋나면 디렉터리 이름 둘을 담은 메시지로 즉시 실패한다. 결정적 변형 `RenderedByTheUpdateFlowInOneSecond` 가 공유 쪽 가지를, 기본 변형이 별개 디렉터리 쪽 가지를 매번 정확히 단언한다.

**약해지지 않았음을 보이는 변이 (트리 `3d8045852c` + 임시 변경, 되돌림 `cmp` 확인)**: `newPriorGenerationSource.generations()` 가 항상 빈 세대를 돌려주게 하면 기본 변형이 붉다 (공유 변형은 PASS — 그 가지는 세대가 없는 것이 맞다).

```text
    migrate_ownership_remaining_test.go:534: verbose view: a verbose row does not carry its reason "evidence-incomplete:original-deployed-path": 2 occurrence(s), first: · kept: .agents/skills/moai-foundation-core/SKILL.md — unregistered, no-provenanc
--- FAIL: TestRemainingListKeeps313UnregisteredCopies (7.83s)
    --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (4.16s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (3.67s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	9.155s
FAIL
RESTORED
```

**증거 (트리 `3d8045852c`, 작업 트리 깨끗; 측정 시작 시점 `60.43 41.63 44.85` 의 부하 평균)**

연속 7회 — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$' -count=1 -v` (매 실행 한 번씩 따로, 줄은 원문에서 추림). 7/7 통과, 모든 rc 0:

```text
run 1 rc=0
--- PASS: TestRemainingListKeeps313UnregisteredCopies (19.33s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (3.17s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (3.43s)
ok  	github.com/modu-ai/moai-adk/internal/cli	20.856s
run 2 rc=0
--- PASS: TestRemainingListKeeps313UnregisteredCopies (23.09s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (2.75s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (3.02s)
ok  	github.com/modu-ai/moai-adk/internal/cli	24.287s
run 3 rc=0
--- PASS: TestRemainingListKeeps313UnregisteredCopies (20.25s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (2.19s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (3.52s)
ok  	github.com/modu-ai/moai-adk/internal/cli	21.290s
run 4 rc=0
--- PASS: TestRemainingListKeeps313UnregisteredCopies (16.12s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (2.15s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (2.17s)
ok  	github.com/modu-ai/moai-adk/internal/cli	17.108s
run 5 rc=0
--- PASS: TestRemainingListKeeps313UnregisteredCopies (15.33s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (2.20s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (2.37s)
ok  	github.com/modu-ai/moai-adk/internal/cli	16.247s
run 6 rc=0
--- PASS: TestRemainingListKeeps313UnregisteredCopies (22.96s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (2.15s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (5.14s)
ok  	github.com/modu-ai/moai-adk/internal/cli	24.200s
run 7 rc=0
--- PASS: TestRemainingListKeeps313UnregisteredCopies (22.24s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow (4.01s)
    --- PASS: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (3.17s)
ok  	github.com/modu-ai/moai-adk/internal/cli	23.510s
```

선택자 묶음 한 번 (기준선 11 + chunk 1 의 22 + chunk 2 의 새 최상위 시험 12 = 45개; 앞서 쓴 34개 선택자의 상위 집합) — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(<45개 앵커된 이름>)$' -count=1 -v`: 최상위 `--- PASS` 45개, `--- FAIL`/`--- SKIP` 0개.

```text
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	70.128s
exit=0
```

품질 (트리 `3d8045852c`): 빌드, 윈도 빌드, gofmt, vet 모두 `exit=0`; 린트는 이전과 같은 `factory_run_adopt.go` 의 기존 errcheck 3건뿐이다.

```text
HEAD=3d8045852c24ba9befa976639ea2445dc37dc8ca
$ go build ./...
exit=0
$ GOOS=windows GOARCH=amd64 go build ./...
exit=0
$ gofmt -l internal/cli internal/manifest
exit=0
$ go vet ./internal/cli/ ./internal/manifest/
exit=0
$ golangci-lint run ./internal/manifest/ ./internal/cli/
3 issues:
* errcheck: 3
exit=1
```

**시계 초 경계에 기대는 다른 시험 목록 (chunk 1, 2 에서 내가 쓴 것을 코드로 확인; 말없이 바꾸지 않았다)**

| 시험 / 도우미 | 시계 의존 | 처리 |
|---|---|---|
| `TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlow` | 두 실행의 실행 백업이 같은 초면 실행 1 의 세대가 대체됨 (이번 간헐 실패) | **이번에 고침** (위) |
| `TestPriorManifestRetainedInRunBackup` (`update_prior_manifest_test.go:220,260`) | 실행 2, 3 앞에 `time.Sleep(1100ms)`. 다음 실행의 백업은 이전 실행의 백업보다 최소 1.1초 뒤이므로 이름이 항상 달라진다 — 부하는 간격을 늘릴 뿐 줄이지 않는다 | **그대로 둠** (고정 대기가 하한이라 결정적이며 chunk 1 정직한 회계 3 에 기록됨). 폴링으로 바꾸면 의도는 같고 더 빠르나 말없는 변경이 되므로 목록만 적는다 |
| `runUpdateFlow` 도우미 (`migrate_ownership_helpers_test.go:453`, `NewPriorManifest`) | `began := now-1s` 보다 새로운 mtime 의 `manifest.prior.json` 중 이름순 마지막을 고름. 일반(건너뛰는) 실행 직후 1초 안에 있으면 앞 실행의 파일을 새것으로 잘못 보고할 수 있음 | **그대로 둠**: `NewPriorManifest` 를 단언하는 곳은 강제 실행 뒤뿐이며 (`runUpdateFlowForced` 가 `Skipped` 를 먼저 거부하고, 이름순 마지막이 항상 가장 새 디렉터리) 건너뛴 실행의 값을 단언에 쓰는 시험은 없다. 잠재적 약점으로 공개 |
| `TestOwnershipHelperForcedRunReachesReconciliation` | 실행 4 번을 이어 돌림 | 시계 초와 무관: 단언은 `skipped` 불리언뿐이다 |
| `TestProfileEnvDoesNotReachProfileRoot` 카나리·대조군 | 실행마다 별개 프로젝트, 타임스탬프·실행 디렉터리 이름은 `<TS>`/`<RUN>` 으로 정규화 | 초 경계에 걸려도 구조(디렉터리 수)가 달라지지 않는다 — 코드 읽기 + 정규화를 고친 뒤 변이 없는 상태로 이 시험을 6번 실행(한 실행은 1초보다 길어 경계를 반드시 넘는다)해 차이가 없었다는 관측. 별도의 경계 강제 실험은 하지 않았다 |
| 그 밖의 새 시험 (`StructuredRows...`, 사례 시험, 경우 시험, 특성화 시험 등) | 한 프로젝트에 한 번의 마이그레이션 또는 한 번의 실행 | 두 실행이 한 디렉터리를 공유할 수 없다 |
| `update_prior_manifest_test.go` 와 `..._io_test.go` 의 합성 세대 (`20260101_00000N`) | 시계 초가 아니라 **날짜** 에 의존: 진짜 실행의 이름 (오늘) 이 합성 이름 (2026-01-01) 보다 뒤에 정렬되어야 하는 가지가 있다 | 지금 날짜에서 참. 시계가 2026-01-01 이전이면 순서 가정이 깨진다 — 공개만 하고 바꾸지 않았다 |

**남는 제품 쪽 사실 (잔여 위험으로 공개, 이번 수정이 고치지 않음).** 같은 시계 초에 시작한 두 갱신은 이전 세대 하나를 잃는다. 방향은 보존 쪽이라 (증거가 줄어 `managed` 로 승격될 파일이 줄 뿐) 제거 방향의 오류는 없다. 두 세대를 모두 남기려면 실행 백업 디렉터리 이름의 해상도를 올리거나 (`internal/cli/update/backup`) 읽는 쪽이 한 디렉터리의 여러 세대를 읽도록 design.md D-3 항목 3 의 「정확한 이름만 읽는다」 를 고쳐야 하며 둘 다 이 카드의 범위 밖이다. 리더가 원하면 별도 카드로.

### Run chunk 3 (M5, M6, M7) — 기록 시작

시작 HEAD `78a09f1ce2`, 브랜치 `WT-update-ownership-repair`, 기록 규약은 chunk 1, 2 와 같다 (명령, 원문 출력, 종료 코드, 트리 SHA; 시험은 `TMPDIR=/private/tmp/own-fixture-d1` 아래 임시 디렉터리와 임시 HOME 만 썼다; 판정 도구는 이 트리의 `go test`, 설치된 바이너리 교체 없음). 이 chunk 는 `internal/cli` 패키지 전체, `-race`, 자원 슬롯을 쓰지 않는다 (앵커된 `-run` 만). **시계 초 비의존 규칙** (chunk 2 정정에서 배움): 실행 백업 디렉터리 이름은 1초 해상도라서 같은 초의 두 실행은 디렉터리를 공유한다 — 두 실행의 디렉터리가 달라야 하는 시험은 폴링으로 결정적으로 만들거나 `priorManifestWriter` 시임을 쓰고, 어느 시계 초에 실행이 놓이는지에 기대는 단언은 쓰지 않는다. M5 의 시험은 한 번의 마이그레이션 실행과 고정된 합성 이전 세대 파일 (`20260101_000001`) 만 쓰므로 시계에 닿지 않는다.

#### M5 — 관리 파일만 있던 빈 디렉터리 정리 (AC-OWR-13)

**커밋 A** `42854e32aa` (시험 파일만: `internal/cli/migrate_ownership_emptydir_test.go`) 와 **커밋 A2** `866a1f8103` (시험 파일만, 같은 파일): A 의 강도 점검에서 내가 찾은 약점 둘을 고쳤다. (1) 변이 서브 시험 둘이 「문제 목록에 해당 디렉터리가 언급되는가」 만 봐서 아무것도 정리하지 않는 구현 앞에서도 통과했다 → 이제 정확한 문제 문장 `<디렉터리>: removed, want it kept` 를 요구한다. (2) `RootsAndAgentDirectoriesAreNeverCandidates` 에 도달 대조군이 없었다 → 정리 단계가 같은 실행에서 돌았음 (사례 A 의 디렉터리가 사라짐) 을 먼저 단언한다. 이 RED 셀은 A2 트리에서 측정했다 (A 트리에서의 첫 측정은 약점 때문에 버렸다).

시험 `TestManagedOnlyEmptyDirsPruned` 의 서브 시험: `Premise` (픽스처가 빈 디렉터리 아홉 개, 이전 세대 파일, 시작 매니페스트의 기록을 실제로 갖는지), `PrunesManagedOnlyAndKeepsTheRest` (스위치 켬; 사례 A 이전 실행이 비움, A' 시작 매니페스트만 기억, B 이번 실행이 비움 — 도달 대조군 `migIdenticalSkill` 이 먼저 제거됨, C 중첩 세 단계, 스킬 미러 루트의 관리 빈 디렉터리 `.agents/skills/...` 정리 / D 기록 없음, E `user_created` 기록, 혼합 부모 (자식은 관리 기록뿐이라 정리, 부모는 `user_created` 형제 기록이 있어 유지) , 미러 쪽 기록 없음 유지 / F 사용자 파일이 든 디렉터리 내용 불변; 스킬 루트 둘과 `.claude/agents/moai` 가 남음; 이번 실행의 파일 손실이 보고된 제거 목록과 정확히 같음 = 정리 단계가 파일을 지우지 않음; 정리된 네 디렉터리를 보고 줄이 이름으로 댐), `IndependentOfTheRemovalSwitch` (스위치 끔: 제거 0건인데 A, A', C, 미러 정리 / D, E 유지), `RootsAndAgentDirectoriesAreNeverCandidates`, `RemainingRowsAreDirectoriesNotFiles` (행 네 개, 묶음 줄이 디렉터리를 「file(s)」 로 부르지 않음), 변이 둘 `MutantPruningAnyEmptyDirIsCaught` (비어 있기만 하면 지우는 구현 → D 가 사라짐) 과 `MutantIgnoringNonManagedRecordsIsCaught` (관리 기록이 하나라도 있으면 출처로 보는 구현 → 혼합 부모가 사라짐). 별도 표 시험 `TestEmptyDirProvenanceVerdict` (순수 판정의 경계 열한 가지: 기록 없음, 관리 기록만, 시작 매니페스트만, 두 원천 같은 경로, `user_created`/`user_modified` 혼합, `deprecated`, 빈 출처 값, 이름 접두 대 디렉터리 접두, 디스크에 아직 있는 기록 경로).

**AC-OWR-13 RED 셀 (트리 `866a1f8103`, A2 커밋 직후, 작업 트리 깨끗, 부하 평균 `29.60 42.23 45.79`)**

명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestManagedOnlyEmptyDirsPruned$' -count=1 -v`

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/migrate_ownership_emptydir_test.go:360:34: undefined: emptyDirEvidence
internal/cli/migrate_ownership_emptydir_test.go:360:60: undefined: emptyDirVerdict
internal/cli/migrate_ownership_emptydir_test.go:361:11: undefined: emptyDirVerdict
internal/cli/migrate_ownership_emptydir_test.go:376:37: undefined: emptyDirEvidence
internal/cli/migrate_ownership_emptydir_test.go:376:70: undefined: emptyDirVerdict
internal/cli/migrate_ownership_emptydir_test.go:380:13: undefined: emptyDirVerdict
internal/cli/migrate_ownership_emptydir_test.go:386:14: undefined: emptyDirVerdict
internal/cli/migrate_ownership_emptydir_test.go:390:11: undefined: emptyDirVerdict
internal/cli/migrate_ownership_emptydir_test.go:402:51: undefined: emptyDirEvidence
internal/cli/migrate_ownership_emptydir_test.go:402:77: undefined: emptyDirVerdict
internal/cli/migrate_ownership_emptydir_test.go:390:11: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

종료 코드 1. 이유: 정리 단계의 증거 모양 (`emptyDirEvidence`, `emptyDirVerdict`) 과 시험이 갈아 끼우는 판정 자리 (`emptyDirProvenanceFn`, `dirManagedProvenance`) 가 아직 없어 시험 패키지가 빌드되지 않는다 (시험이 존재하는 상태의 빌드 실패이며 제로 매치 셀렉터가 아니다).

**단언 수준 RED (같은 트리 `866a1f8103` + 임시 stub 파일, 커밋하지 않고 삭제)**. 빌드 실패만으로는 「올바른 이유로 붉은가」 를 알 수 없어서, 새 심볼 다섯 개의 선언만 담은 임시 파일 `internal/cli/zz_m5_stub.go` (판정은 항상 `unregistered`, **정리 단계 없음**) 를 얹고 같은 명령을 돌렸다. 인용에서 `=== RUN` 줄과 `no report line names the pruned directory` 네 줄은 뺐다. 종료 코드 1, 이 실행 뒤 stub 파일을 지웠고 `git status --short` 가 빈 출력이다.

```text
    migrate_ownership_emptydir_test.go:280: reach control .claude/skills/moai-ghost-a was not pruned: the pruning step did not run, so a keep below proves nothing
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-a: still present, want it removed
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-s: still present, want it removed
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-c: still present, want it removed
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-c/refs: still present, want it removed
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-c/refs/deep: still present, want it removed
    migrate_ownership_emptydir_test.go:280: .agents/skills/moai-ghost-m: still present, want it removed
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-d: kept without a remaining row
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-e: kept without a remaining row
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-cm: kept without a remaining row
    migrate_ownership_emptydir_test.go:280: .claude/skills/moai-ghost-cm/refs: still present, want it removed
    migrate_ownership_emptydir_test.go:280: .agents/skills/moai-ghost-n: kept without a remaining row
    (IndependentOfTheRemovalSwitch: 같은 열두 줄, 줄 번호 300)
    migrate_ownership_emptydir_test.go:311: the reach control .claude/skills/moai-ghost-a was not pruned: the keeps below would prove nothing
    migrate_ownership_emptydir_test.go:339: 0 rows carry "empty-dir-no-provenance", want 4 (.claude/skills/moai-ghost-d, .claude/skills/moai-ghost-e, .claude/skills/moai-ghost-cm, .agents/skills/moai-ghost-n)
--- FAIL: TestManagedOnlyEmptyDirsPruned (7.76s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/Premise (1.41s)
    --- FAIL: TestManagedOnlyEmptyDirsPruned/PrunesManagedOnlyAndKeepsTheRest (0.99s)
    --- FAIL: TestManagedOnlyEmptyDirsPruned/IndependentOfTheRemovalSwitch (1.04s)
    --- FAIL: TestManagedOnlyEmptyDirsPruned/RootsAndAgentDirectoriesAreNeverCandidates (0.85s)
    --- FAIL: TestManagedOnlyEmptyDirsPruned/RemainingRowsAreDirectoriesNotFiles (1.34s)
    --- FAIL: TestManagedOnlyEmptyDirsPruned/MutantPruningAnyEmptyDirIsCaught (0.97s)
    --- FAIL: TestManagedOnlyEmptyDirsPruned/MutantIgnoringNonManagedRecordsIsCaught (1.17s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	9.159s
FAIL
```

해석: 사례 B (`moai-foundation-core`, 이번 실행이 마지막 파일을 제거) 는 이 stub 상태에서도 붉지 않았다 — t1668 의 `pruneEmptiedSkillDir` 가 이미 하는 일이므로 B 는 새 코드 없이 회귀 가드이다 (plan.md § G.1, acceptance AC-OWR-13 비고와 같은 예측). 붉은 것은 A, A', C 의 세 단계, 미러 쪽 관리 디렉터리, 혼합 자식 (정리되지 않음) 과 D, E, 혼합 부모, 미러 D (잔존 행 없음) 이다. 이유: 걷기 뒤의 정리 단계가 없고 정리 판단의 근거 (시작 매니페스트 + 이전 세대 기록 해석) 가 없다 — 정확히 acceptance 의 RED 이유 문장이다.

**AC-OWR-13 — 커밋 C** `ba1ebac9a4` (`internal/cli/migrate_ownership_emptydir.go` 신규, `internal/cli/migrate_project_assets.go` 호출 한 줄 + 주석, `internal/cli/migrate_ownership_remaining.go` 묶음 줄의 명사). 정리 단계 `pruneManagedEmptyDirs` 는 걷기와 제거가 끝난 뒤, 매니페스트 항목 정리와 잔존 행 정렬 앞에서 한 번 돈다. 두 스킬 루트 아래 디렉터리마다 그 트리에 파일이 하나도 없을 때만 (파일, 심링크, 읽기 오류가 하나라도 있으면 그 트리는 후보가 아니다) 가장 깊은 디렉터리부터 판정한다. 판정 `dirManagedProvenance` 는 시작 시점 매니페스트와 보존된 이전 세대 두 원천에서 `<디렉터리>/` 접두의 모든 기록을 모아, 기록이 있고 전부 `template_managed` 이며 기록된 경로가 하나도 디스크에 없을 때만 「관리 파일만 있던」 으로 본다 (비관리 기록이 한 원천에라도 있으면 그 기록이 판정을 정한다: `user_created` > `user_modified` > 그 밖). 제거는 스킬 디렉터리 한 단계 아래는 t1668 의 `pruneEmptiedSkillDir` 를 그대로 재사용하고, 더 깊은 단계는 같은 경계 검사를 가진 `pruneNestedEmptyDir` (`os.Remove` 만) 가 한다. 스킬 루트 자체, 에이전트 디렉터리, 심링크 미러 루트는 후보가 아니다. 유지된 디렉터리는 이유 `empty-dir-no-provenance` 의 잔존 행이 된다 (분류는 판정이 정한다: 기록 없음 `unregistered`, `user_created` 기록 `user-created`).

**AC-OWR-13 GREEN (트리 `ba1ebac9a4`, 작업 트리 깨끗; 이 실행의 부하 평균은 측정 끝에서 `88.55 54.77 45.84`)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestManagedOnlyEmptyDirsPruned|TestEmptyDirProvenanceVerdict)$' -count=1 -v` (인용은 `--- ` 줄만, 서브 시험 이름은 원문 그대로):

```text
--- PASS: TestManagedOnlyEmptyDirsPruned (61.27s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/Premise (36.10s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/PrunesManagedOnlyAndKeepsTheRest (11.32s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/IndependentOfTheRemovalSwitch (4.41s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/RootsAndAgentDirectoriesAreNeverCandidates (1.88s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/RemainingRowsAreDirectoriesNotFiles (1.94s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/MutantPruningAnyEmptyDirIsCaught (1.94s)
    --- PASS: TestManagedOnlyEmptyDirsPruned/MutantIgnoringNonManagedRecordsIsCaught (3.67s)
--- PASS: TestEmptyDirProvenanceVerdict (0.00s)
    (서브 시험 열한 개 모두 PASS)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	62.801s
```

(이 실행의 종료 코드는 인용하지 않는다 — 파이프 뒤에서 읽어 값이 비었다. 종료 코드는 아래 연속 5회의 값이 증거이다.)

**연속 5회 (같은 선택자 `^TestManagedOnlyEmptyDirsPruned$`, 각 `-count=1 -v`, 각 실행의 종료 코드를 직접 읽음; 트리 `89731b5e59` + 아직 커밋하지 않은 구현 C — 같은 내용이 `ba1ebac9a4`)**

```text
run 1 rc=0 top-pass=1 sub-pass=7 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	8.515s
run 2 rc=0 top-pass=1 sub-pass=7 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	9.192s
run 3 rc=0 top-pass=1 sub-pass=7 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	8.072s
run 4 rc=0 top-pass=1 sub-pass=7 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	5.646s
run 5 rc=0 top-pass=1 sub-pass=7 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	8.518s
```

**실제 운영 코드 변이 둘 (트리 `89731b5e59` + 임시 변경; 둘 다 되돌렸고 `cmp` 로 원본과 동일함을 확인, 출력 `RESTORED-IDENTICAL`)**

- R1 정리 단계 호출 한 줄을 지움 → 서브 시험 여섯이 붉다 (`Premise` 만 PASS): `PrunesManagedOnlyAndKeepsTheRest`, `IndependentOfTheRemovalSwitch`, `RootsAndAgentDirectoriesAreNeverCandidates`, `RemainingRowsAreDirectoriesNotFiles`, 두 변이 서브 시험. 종료 `FAIL github.com/modu-ai/moai-adk/internal/cli 10.078s`.
- R2 `dirManagedProvenance` 가 비관리 기록 (`user_created`, `user_modified`, 그 밖) 을 보지 않게 함 → 서브 시험 셋 (`PrunesManagedOnlyAndKeepsTheRest`, `IndependentOfTheRemovalSwitch`, `RemainingRowsAreDirectoriesNotFiles`) 과 표 시험의 다섯 경우 (`a_user_created_record_beside_managed_ones`, `a_user_modified_record`, `a_deprecated_record_is_not_managed`, `an_empty_provenance_value_is_not_managed`, `a_user_created_record_in_the_older_of_two_sources_still_counts`) 가 붉다. 종료 `FAIL github.com/modu-ai/moai-adk/internal/cli 9.805s`.

**회귀 (기준선 11 + chunk 1, 2 의 시험 33 + 이번 M5 둘 = 앵커된 이름 46개, 트리 `89731b5e59` + 구현 C)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(<46개 앵커된 이름>)$' -count=1 -v` (이름 목록 = plan.md 의 기준선 11개 + `migrate_ownership*_test.go`, `update_prior_manifest*_test.go` 의 최상위 `func Test…` 전부, `sort -u` 로 46개임을 `wc -l` 로 확인). 출력 요약 (내 집계 한 줄과 `go test` 마지막 줄): `top-level PASS=46 FAIL=0 SKIP=0`, `ok  	github.com/modu-ai/moai-adk/internal/cli	90.840s`, `exit=0`. `internal/manifest` 의 `TestManifestRecordsDeployedVersion`: `--- PASS`, `ok  	github.com/modu-ai/moai-adk/internal/manifest	0.258s`. (chunk 2 의 기록은 「45개」 라 적었고 이 트리에서 `internal/cli` 에서 내가 센 앞선 시험은 11 + 21 + 12 = 44개이다; 이번 46 = 44 + 2 는 센 이름과 `--- PASS` 수가 일치한다. 앞선 「45」 는 `internal/manifest` 의 시험 하나를 합쳐 센 값으로 보이나 (44 + 1), 그 기록의 명령을 다시 돌려 확인하지 않았다 — 미검증.)

**M5 정직한 회계**

1. 사례 B (이번 실행이 마지막 파일을 제거) 는 새 코드 없이 t1668 의 `pruneEmptiedSkillDir` 가 이미 하는 일이다 — RED 의 stub 측정에서도 붉지 않았다. 이 케이스는 회귀 가드이며 이번 단계가 한 일이 아니다.
2. 판정 어휘는 닫힌 이유 집합을 늘리지 않았다 (`empty-dir-no-provenance` 는 chunk 1 에서 이미 어휘에 있었다).
3. 「스킬 디렉터리」 의 깊이: REQ-OWR-008 은 「스킬 루트 아래의 디렉터리」 라 하고 design.md D-6 은 중첩을 아래쪽부터 위로 평가한다고 한다. 이 구현은 스킬 루트 아래 모든 깊이의 디렉터리를 같은 규칙으로 평가한다 (스킬 루트 자신은 아님).
4. 심링크로 된 미러 루트 (`.agents/skills` 가 심링크) 는 후보가 아니다: 그 디렉터리는 대상 루트에서 판정된다. 실제 심링크 미러 루트를 둔 픽스처로 이 가지를 시험하지는 않았다 (코드 읽기와 `Lstat` 분기만) — 갭.
5. `--dry-run` 미리보기는 이 정리를 알리지 않는다 (요구사항에 없고 미리보기 개정은 M7 의 판정 일치 범위에 이 디렉터리 정리를 넣지 않았다).
6. 잔존 행에는 디렉터리 행이 생긴다. 묶음 줄은 이 이유에서만 「file(s)」 대신 「empty director(ies)」 를 쓴다 (`remainingSummaryLines` 의 명사 한 곳).
7. `userAssetsInstalled` 가 거짓이거나 프로젝트 매니페스트가 없는 실행은 걷기 앞에서 끝나므로 정리 단계에도 닿지 않는다 (기존 조기 반환을 바꾸지 않았다).

#### M6 — commands 한 처분과 탐색 보증 (AC-OWR-09, AC-OWR-10)

**chunk 1 탐침의 재수행 (정직 공개).** chunk 1 의 M6 열린 질문 탐침은 부분적이었다. (1) 변형 「같은 버전」 은 `system.yaml` 이 없어 프로젝트 버전이 `0.0.0` 이었고 바이너리 버전과 달랐다 — 같은 버전이 아니었다. (2) 변형 「낡은 기록」 은 `runUpdate` 가 아니라 `runTemplateSyncWithProgress` 를 직접 불렀고 `runUpdate` 의 사용자 자산 단계와 버전 일치 관문 앞 단계를 거치지 않았다. 그 판정은 근거로 쓰지 않고 아래 세 변형을 이번 chunk 에서 새로 구성해 처음부터 다시 쟀다. 모든 변형은 임시 HOME 과 `TMPDIR=/private/tmp/own-fixture-d1` 아래 임시 프로젝트에서만 돌았고 바이너리는 이 트리의 `go test` 가 빌드한 것이다 (설치된 바이너리 없음).

| 변형 | 기록 버전 / 바이너리 버전 | 기록된 템플릿 트리 | 구동 | 건너뛰기 아님을 보인 증거 |
|------|--------------------------|--------------------|------|---------------------------|
| V1 같은 버전 | `v3.2.0-rc.36` / `v3.2.0-rc.36` (`version.Version` 고정) | 낡음 (`stale-tree-hash`) — 그래서 `--force` 없이 조정에 닿는다 | `runTemplateSyncWithProgress` (`moai update` 의 동기화 진입) | `skipped == false`, 실행 백업에 새 `manifest.prior.json` |
| V2 낡은 기록 | `v3.2.0-rc.35` / `v3.2.0-rc.36` (현장은 rc.35 기록 + rc.36 바이너리) | 낡음 | **`runUpdate` 전체** (`--yes --templates-only`, 사용자 자산 단계와 버전 일치 관문 포함) | 출력에 「이미 최신」 표지 없음, 실행 백업에 새 `manifest.prior.json` |
| V3 매니페스트 유실 | `v3.2.0-rc.35` / `v3.2.0-rc.36` | 낡음 | `runUpdate` 전체 | 같음 |

V1, V2 의 픽스처는 명령 네 상태를 가진다: S1 `gtd.md` (등록 managed, 건강), S2 `fix.md` (등록 managed 로 배포기가 렌더한 바이트를 기록한 뒤 사용자가 `USERLINE: keep me` 한 줄을 덧붙임), S3 `todo.md` (템플릿 바이트, 매니페스트 항목 없음), S4 `.claude/commands/moai-user-cmd.md` (사용자 작성). V3 는 템플릿 command 17개 전부를 템플릿 바이트로 두고 항목을 하나도 만들지 않는다. 바이트의 기준은 같은 진입을 새 프로젝트·새 HOME 에서 돌린 **통제 실행** (`.tmpl` 명령의 바이트는 템플릿 원문이 아니라 배포기의 렌더이므로 그것만이 정직한 기준이다).

**커밋.** A `0ac54c59e2` (시험 파일만: `internal/cli/update_commands_discovery_test.go`), A2 `4e9564f790` (시험 파일만: 통제 실행 자신의 전제를 강화 — 카탈로그 core 계층의 모든 에이전트가 설치되었고 Claude 에이전트가 템플릿의 바이트 복사본임을 요구한다; 그렇지 않으면 설치기가 깨져도 통제 실행과 모든 변형이 함께 깨져 오라클 안에 숨는다), A3 `a7d2663e11` (시험 파일만: 변형 V3 추가). **구현 커밋 C 는 없다** — 아래 이유.

**AC-OWR-09 도착 셀 (트리 `a7d2663e11`, 작업 트리 깨끗, 부하 평균 `27.99 36.47 41.89`)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCommandsSingleDispositionAndPresent$' -count=1 -v`, 종료 코드 0. 인용은 관측 로그 줄과 `---`/결과 줄만 (원문 순서 그대로):

```text
    update_commands_discovery_test.go:529: V1: plan after the run (read-only classification): 384 files, 380 template-owned, 3 user-modified, 1 user-owned, 0 stale
    update_commands_discovery_test.go:529: V1: report line: ✓ 0 merged, 1 conflict(s), 0 preserved
    update_commands_discovery_test.go:529: V1: report line: + 416 add · ~ 5 update · ! 2 conflict
    update_commands_discovery_test.go:529: V1: report line: Reconciliation: 1 refreshed, 0 merged, 1 conflict(s), 0 preserved, 0 archived-removed
    update_commands_discovery_test.go:529: V1: report line: conflict: .claude/commands/moai/fix.md (sidecar: .claude/commands/moai/fix.md.moai-new)
    update_commands_discovery_test.go:557: V2: plan after the run (read-only classification): 384 files, 382 template-owned, 1 user-modified, 1 user-owned, 0 stale
    update_commands_discovery_test.go:557: V2: report line: ✓ 0 merged, 1 conflict(s), 0 preserved
    update_commands_discovery_test.go:557: V2: report line: + 416 add · ~ 5 update · ! 2 conflict
    update_commands_discovery_test.go:557: V2: report line: Reconciliation: 1 refreshed, 0 merged, 1 conflict(s), 0 preserved, 0 archived-removed
    update_commands_discovery_test.go:557: V2: report line: conflict: .claude/commands/moai/fix.md (sidecar: .claude/commands/moai/fix.md.moai-new)
    update_commands_discovery_test.go:557: V2: report line: ✗ 1 conflicting file(s) flagged by the 3-way merge — resolve them before the next template sync (see the backup at .moai-backups/20261018_174740)
    update_commands_discovery_test.go:594: V3: plan after the run (read-only classification): 383 files, 383 template-owned, 0 user-modified, 0 user-owned, 0 stale
    update_commands_discovery_test.go:594: V3: report line: ✓ 0 merged, 0 conflict(s), 0 preserved
    update_commands_discovery_test.go:594: V3: report line: + 402 add · ~ 19 update · ! 2 conflict
--- PASS: TestCommandsSingleDispositionAndPresent (3.26s)
    --- PASS: TestCommandsSingleDispositionAndPresent/Control (1.04s)
    --- PASS: TestCommandsSingleDispositionAndPresent/V1_SameVersion (0.39s)
    --- PASS: TestCommandsSingleDispositionAndPresent/V2_StaleRecord (0.82s)
    --- PASS: TestCommandsSingleDispositionAndPresent/V3_NoManifestRecords (1.00s)
    --- PASS: TestCommandsSingleDispositionAndPresent/V1AndV2ReachTheSameCommandsTree (0.00s)
    --- PASS: TestCommandsSingleDispositionAndPresent/MutantTwoDispositionsForOnePathIsCaught (0.00s)
    --- PASS: TestCommandsSingleDispositionAndPresent/MutantRedeployOverTheUserEditIsCaught (0.00s)
    --- PASS: TestCommandsSingleDispositionAndPresent/MutantRemovedCommandIsCaught (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.181s
```

**AC-OWR-10 도착 셀 (같은 트리)** — 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCommandsAndAgentsDiscoverableAfterInterruptedRun$' -count=1 -v`, 종료 코드 0:

```text
--- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun (9.98s)
    --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude (5.43s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude/Control (1.97s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude/RecoveryRun (1.41s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude/PlainSameVersionRerunIsSkipped (0.19s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude/MutantsAreCaught (1.69s)
    --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both (4.55s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both/Control (1.24s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both/RecoveryRun (1.59s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both/PlainSameVersionRerunIsSkipped (0.13s)
        --- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both/MutantsAreCaught (1.38s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	11.242s
```

**연속 5회씩 (같은 트리, 각 실행의 종료 코드를 직접 읽음; `sub-pass` 는 서브 시험 `--- PASS` 줄 수)**

```text
TestCommandsSingleDispositionAndPresent run 1 rc=0 top-pass=1 sub-pass=8 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	6.114s
TestCommandsSingleDispositionAndPresent run 2 rc=0 top-pass=1 sub-pass=8 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	7.744s
TestCommandsSingleDispositionAndPresent run 3 rc=0 top-pass=1 sub-pass=8 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	8.344s
TestCommandsSingleDispositionAndPresent run 4 rc=0 top-pass=1 sub-pass=8 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	12.117s
TestCommandsSingleDispositionAndPresent run 5 rc=0 top-pass=1 sub-pass=8 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	5.779s
TestCommandsAndAgentsDiscoverableAfterInterruptedRun run 1 rc=0 top-pass=1 sub-pass=10 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	8.522s
TestCommandsAndAgentsDiscoverableAfterInterruptedRun run 2 rc=0 top-pass=1 sub-pass=10 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	12.457s
TestCommandsAndAgentsDiscoverableAfterInterruptedRun run 3 rc=0 top-pass=1 sub-pass=10 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	10.366s
TestCommandsAndAgentsDiscoverableAfterInterruptedRun run 4 rc=0 top-pass=1 sub-pass=10 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	8.799s
TestCommandsAndAgentsDiscoverableAfterInterruptedRun run 5 rc=0 top-pass=1 sub-pass=10 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	16.748s
```

**결론: 두 변형이 아니라 세 변형과 두 harness 의 복구를 모두 돌렸고 현장의 순환은 이 트리에서 재현되지 않았다.** 근거: 건너뛰지 않은 실행 (V1 `skipped == false`; V2, V3 「이미 최신」 표지 없음; 모두 새 `manifest.prior.json`) 에서, 계획의 네 처분 목록이 서로소이고 (`cmdPlanProblems`; 계획이 비지 않음과 명령 경로가 계획에 있음을 먼저 단언), S1 은 배포기의 현재 렌더, S2 는 사용자 줄이 남고 충돌 사이드카 `fix.md.moai-new` 가 하나 생기며, S3 는 템플릿 렌더, S4 는 불변이고, 고정된 경로마다 그 경로를 말하는 보고 줄이 둘을 넘지 않으며, 템플릿이 운반하는 command 17개가 전부 존재하고 비어 있지 않으며 frontmatter 의 `description`, `allowed-tools` 가 비어 있지 않다. V1 과 V2 의 `.claude/commands` 트리 (경로, 모드, 내용 해시) 는 같다 (`V1AndV2ReachTheSameCommandsTree`). 그래서 **AC-OWR-09 와 -10 은 회귀 가드로 남고 (승격 없음, 구현 커밋 없음)**, REQ-OWR-006 의 존재 보증은 시험으로 닫았다. 리더에게 보고할 문구: 「재현 안 됨 — 돌린 변형: V1 같은 버전 (`runTemplateSyncWithProgress`), V2 낡은 기록 (`runUpdate`), V3 매니페스트 유실 (`runUpdate`); AC-OWR-10 복구: claude, both」.

**붉을 수 있음의 증거 (회귀 가드의 변이).** 시험 안의 입력 쪽 변이 (`Mutant...IsCaught` 서브 시험: 한 경로의 보고 줄 둘, 사용자 줄을 덮어쓴 재배포, 사라진 명령, 사라진 명령·빈 명령·이름 없는 에이전트·사라진 에이전트) 외에 **실제 운영 코드를 임시로 바꾼 변이 셋** 을 이 트리에서 돌렸다 (모두 되돌렸고 세 파일을 `cmp` 로 원본과 대조해 `RESTORED-IDENTICAL` 을 얻었다; 적색은 아래 원문). 명령은 모두 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestCommandsSingleDispositionAndPresent|TestCommandsAndAgentsDiscoverableAfterInterruptedRun)$' -count=1 -v` 의 줄 추림이다.

- **Ma** `internal/template/deployer_mode.go` 의 `isCommonAssetRoot` 목록에 `.claude/commands/moai/` 를 더함 (배포기가 commands 를 영영 쓰지 않음):

```text
    update_commands_discovery_test.go:506: control: .claude/commands/moai: open /private/tmp/own-fixture-d1/TestCommandsSingleDispositionAndPresentControl4235705273/002/.claude/commands/moai: no such file or directory
    update_commands_discovery_test.go:508: the control run did not complete; nothing below can be judged
--- FAIL: TestCommandsSingleDispositionAndPresent (2.56s)
    --- FAIL: TestCommandsSingleDispositionAndPresent/Control (2.56s)
    update_commands_discovery_test.go:751: control: .claude/commands/moai: open /private/tmp/own-fixture-d1/TestCommandsAndAgentsDiscoverableAfterInterruptedRunclaudeCont2631463020/002/.claude/commands/moai: no such file or directory
    update_commands_discovery_test.go:753: the control run did not complete; nothing below can be judged
    update_commands_discovery_test.go:751: control: .claude/commands/moai: open /private/tmp/own-fixture-d1/TestCommandsAndAgentsDiscoverableAfterInterruptedRunbothContro3797682394/002/.claude/commands/moai: no such file or directory
    update_commands_discovery_test.go:753: the control run did not complete; nothing below can be judged
--- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun (3.73s)
    --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude (1.73s)
        --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude/Control (1.73s)
    --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both (2.00s)
        --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both/Control (2.00s)
FAIL
```

- **Mb** `internal/userassets/install.go` 가 Claude 에이전트 본문 대상을 추가하지 않음 (설치기가 에이전트를 쓰지 않음):

```text
    update_commands_discovery_test.go:506: control: the core agent manager-develop was not installed at ~/.claude/agents/manager-develop.md
    update_commands_discovery_test.go:508: the control run did not complete; nothing below can be judged
--- FAIL: TestCommandsSingleDispositionAndPresent (2.23s)
    --- FAIL: TestCommandsSingleDispositionAndPresent/Control (2.23s)
    update_commands_discovery_test.go:751: control: the core agent manager-develop was not installed at ~/.claude/agents/manager-develop.md
    update_commands_discovery_test.go:753: the control run did not complete; nothing below can be judged
    update_commands_discovery_test.go:751: control: the core agent manager-develop was not installed at ~/.claude/agents/manager-develop.md
    update_commands_discovery_test.go:753: the control run did not complete; nothing below can be judged
--- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun (5.59s)
    --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude (2.47s)
        --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/claude/Control (2.47s)
    --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both (3.12s)
        --- FAIL: TestCommandsAndAgentsDiscoverableAfterInterruptedRun/both/Control (3.12s)
FAIL
```

- **Mc** `internal/cli/update/reconcile_classify.go` 의 계획 조립이 템플릿 소유 파일마다 한 번 더 `Stale` 에도 넣음 (한 경로 두 처분):

```text
    update_commands_discovery_test.go:519: plan: .claude/commands/moai/gtd.md: appears in 2 dispositions [stale template-owned]
    update_commands_discovery_test.go:546: plan: .claude/commands/moai/gtd.md: appears in 2 dispositions [stale template-owned]
--- FAIL: TestCommandsSingleDispositionAndPresent (7.92s)
    --- FAIL: TestCommandsSingleDispositionAndPresent/V1_SameVersion (0.78s)
    --- FAIL: TestCommandsSingleDispositionAndPresent/V2_StaleRecord (1.70s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	27.662s
FAIL
```

Ma 와 Mb 는 통제 실행의 전제 단언에서 붉다 (통제 실행도 같은 코드를 쓰므로 변형까지 가지 못한다; 그 단언이 「오라클 자체가 깨짐」 을 큰 소리로 알린다). Mc 는 V1, V2 의 계획 서로소 단언에서 붉다 (V3 는 모든 command 가 항목 없음이라 템플릿 소유가 아니어서 이 변이를 건드리지 않는다 — 변이 대상이 다르다).

**관측한 사실 (결함 판정 아님, 해석은 하지 않음).** 임시 진단 시험 (커밋하지 않고 삭제함) 으로 실행 전 계획을 읽은 원문 (변형 V1 과 V2, 같은 계획 함수):

```text
zz_m6_debug_test.go:26: V1 PRE user-modified: .claude/commands/moai/fix.md, .claude/commands/moai/todo.md, .moai/config/sections/llm.yaml, .moai/config/sections/system.yaml
zz_m6_debug_test.go:26: V2 PRE user-modified: .claude/commands/moai/fix.md, .claude/commands/moai/todo.md, .moai/config/sections/llm.yaml, .moai/config/sections/system.yaml
zz_m6_debug_test.go:26: V1 POST user-modified: .claude/commands/moai/fix.md, .moai/config/sections/llm.yaml, .moai/config/sections/system.yaml
zz_m6_debug_test.go:26: V2 POST user-modified: .claude/commands/moai/fix.md
```

즉 항목이 없는 채 템플릿과 바이트가 같은 command (`todo.md`, S3) 는 실행 전 계획에서 `user-modified` (3-way 병합 경로) 로 분류된다 — research.md G-2 의 후보 (c) 가 계획 수준에서 사실임을 관측했다. 그 병합의 결과는 템플릿 렌더 한 처분이고 (S3 의 보고 줄 0 개, 최종 바이트 == 렌더) 위 시험이 그것을 단언한다. 현장의 「이전 제외 후 재배포」 가 이 분류를 가리키는지는 확정하지 못했다 (분류는 관측했으나 그것이 두 처분이 되는 순환은 관측하지 못했다). V1 과 V2 의 실행 뒤 계획이 다른 부분 (`llm.yaml`, `system.yaml` 이 V1 에서만 `user-modified` 로 남음) 은 `runUpdate` 가 동기화 뒤에 하는 일의 차이로 보이나 원인을 확인하지 않았다 — commands 와 무관한 경로이고 두 변형의 commands 트리는 같다.

**SPEC 문구 결함 (공개된 부채, SPEC 본문은 고치지 않음; 리더에게 manager-spec 재위임 후보로 보고).**

1. **AC-OWR-10 와 REQ-OWR-006 의 에이전트 조건이 이 트리의 제품 구조와 맞지 않는다.** 배포기의 프로젝트 페이로드는 어떤 모드에서도 공통 스킬·에이전트 파일을 쓰지 않는다 (`internal/template/deployer_mode.go` 의 `isCommonAssetRoot` 와 `deployer.go` 의 걷기 건너뛰기 — SPEC-USER-ASSET-INSTALL-001 REQ-005). 에이전트는 사용자 자산 단계가 사용자 루트 (`~/.claude/agents/<name>.md`, 그리고 `~/.codex/agents/<name>.toml`) 에 설치한다. 그래서 AC-OWR-10 이 요구한 「`.claude/agents/moai/*.md` 이름 집합이 템플릿의 12개와 같다」 는 프로젝트 경로에서는 제품이 일부러 쓰지 않으므로 성립할 수 없다. 이 시험은 보증이 실제로 성립하는 곳 — 에이전트는 사용자 루트, commands 는 프로젝트 — 에서 단언한다.
2. **「12개」 는 기본 설치 집합이 아니다.** 기본 (L0, 카탈로그 core 계층) 설치는 에이전트 6개 (`manager-develop`, `manager-docs`, `manager-lead`, `manager-spec`, `plan-auditor`, `sync-auditor`) 만 쓰고 나머지 6개 (`builder-harness`, `e2e-tester`, `manager-design`, `manager-git`, `manager-todo`, `super-advisor`) 는 선택 묶음이다. 이 시험은 통제 실행이 설치한 집합 (전제로 core 계층 전부 포함을 요구) 을 기준으로 삼는다.
3. 같은 이유로 acceptance.md 가 말한 「`.codex/agents/moai/*.toml` 이 템플릿 집합과 같다」 도 프로젝트 경로에서는 쓰이지 않는다. 시험은 통제 실행이 사용자 루트에 설치한 파일과 바이트 일치를 단언한다.

**M6 정직한 회계 / 갭**

1. 구현 커밋이 없다. M6 의 산출은 시험 셋과 이 기록이다. plan.md M6 의 「재현되면 고친다」 가지에는 닿지 않았다.
2. 오라클이 통제 실행이다. 통제 실행이 설치기·배포기 전체와 함께 깨지면 변형이 그 깨짐을 못 본다. 그래서 통제 실행의 전제를 단언으로 묶었다 (commands 17개 이름, core 계층 에이전트 전부, Claude 에이전트의 템플릿 바이트 일치). Codex TOML 에이전트의 바이트는 통제 실행과만 비교한다 (템플릿과의 직접 비교는 하지 않음 — 설치기가 변환을 거치는 면이다).
3. V2, V3 는 `--templates-only` 로 이진 업데이트 단계를 건너뛴다. 실제 `moai update` 의 이진 단계, 재실행 (re-exec) 경로, 비영어 `conversation_language` 렌더, `--no-plugin` 모드, 프로필이 지도된 프로젝트는 구동하지 않았다.
4. 시험 환경의 `conversation_language` 는 기본 (영어) 이다. 다른 언어에서 `.tmpl` 렌더 바이트가 달라도 통제 실행이 같은 설정으로 돌기 때문에 같은 설정 안의 비교는 성립하지만, 다른 언어 설정의 재현 여부는 측정하지 않았다.
5. `PlainSameVersionRerunIsSkipped` 는 제품의 현재 사실 (같은 버전의 일반 재실행은 조정 앞에서 끝나 비워진 프로젝트 commands 가 비어 있는 채 남는다) 을 단언으로 기록한 것이지 요구하는 동작이 아니다. plan.md OQ-6 이 이미 리더에게 올린 항목이다. 같은 실행에서 사용자 루트 에이전트가 복구되는지는 단언하지 않고 관측도 기록하지 않았다.
6. 이 절의 도착 셀은 각각 `-count=1` 한 번의 실행이고 별도로 5회 반복이 있다. 측정 중 부하 평균은 대략 24에서 90 사이였고 반복은 10/10 통과했다.

#### M7 — 미리보기 판정 통일과 dry-run (AC-OWR-16)

**커밋 A** `2d6ec34053` (시험 파일만: `internal/cli/migrate_ownership_preview_test.go`). 시험 셋:

- `TestDryRunChangesNothingAndMatchesVerdict` — `moai update --dry-run` 이 디스패치하는 분기 전체를 `runDryRunUpdateInProject` (`runUpdate`) 로 구동한다. 실행 직전에 모든 디렉터리의 mtime 을 고정 과거 시각 (2020-01-02T03:04:05Z) 으로 맞추고 디렉터리 mtime 을 넣은 트리 해시를 전후 비교하며 `.moai-backups` 부재도 단언한다. 서브 시험: `UnmappedProjectAnnouncesExactlyTheEligibleSet` (AC-OWR-02 ~ -05 파일이 섞인 프로젝트, 제거 스위치 켬), `ProfileMappedKeptByProjectAnnouncesNothing` (프로필 지도 + `KeptBy` 표지, 후보 0), `SwitchOffAnnouncesNothing`, 변이 `MutantOwnJudgmentIsCaught` (미리보기의 종전 자체 판정을 그대로 옮긴 변이 집합 앞에서 비교가 붉음), `MutantCreateThenDeleteIsSeenByTheDigest` (`.claude/skills` 에 파일을 만들었다 지우면 해시가 그 디렉터리에서 갈라짐). 판정 함수의 `removal-eligible` 집합의 오라클은 둘이고 서로 독립이다: (1) **같은 픽스처의 복사본에서 진짜 마이그레이션을 돌려 실제로 제거되는 경로** (프로필 지도 프로젝트는 복사본도 같이 지도), (2) 판정 함수를 픽스처 전체에 직접 돌린 집합 (환경은 시험 안에서 기존 심볼로 따로 조립 — 운영 코드의 새 조립기를 쓰지 않는다). 비교 전 단언 (공허 방지): 두 오라클이 서로 같고 세 개 이상이며 (`manifest` 관리 사본 셋: `moai-foundation-core`, `moai-foundation-quality` 스킬과 `manager-spec` 에이전트), 그 경로 전부가 dry-run 의 reconcile 계획 `UserOwned` 에 있으며 (미리보기가 보는 유일한 집합), 출력에 「Dry-run: reconciliation preview」 머리가 있다. 함정 파일 `manager-git.md` 는 등록 managed + 템플릿 바이트 일치지만 사용자 대응물이 없는 (선택 묶음) 에이전트라 판정은 `counterpart-unconfirmed` 로 유지한다.
- `TestDryRunRunsNoBinaryUpdateStep` — REQ-OWR-011 은 「명령이 하는 모든 단계」 에 걸린다. 개발 빌드에서는 이진 업데이트 단계가 건너뛰어져 AC-OWR-16 의 시험이 이를 볼 수 없으므로, 릴리스 모양의 버전 (`v3.2.0-rc.36`) 과 설치를 거부하는 가짜 설치기 (호출 수만 센다; 거부하므로 재실행으로 이어질 수 없다) 로 dry-run 이 설치기에 닿는지를 단언한다.
- `TestVerdictAndPreviewCodeHasNoWritableCalls` — 판정·미리보기 코드에 쓰기 가능 호출이 없음을 AST 로 단언한다 (`os.WriteFile`, `Create`, `CreateTemp`, `Mkdir*`, `Remove*`, `Rename`, `Truncate`, `Chmod`, `Chtimes`, `Symlink`, `Link`, `OpenFile`, 매니페스트 관리자의 `Save`/`Track`/`SetKeptBy`/`SetDeployingVersion`, `probeCaseSensitiveFS`). 읽기만 하는 파일은 파일 전체를 (`migrate_ownership.go`, `_evidence.go`, `_case.go`, `_env.go`, `_remaining.go`, `update_dryrun_preview.go`), 쓰는 함수와 같은 파일에 있는 읽기 함수는 이름으로 (`migrate_project_assets.go` 의 `newOwnershipEnv`, `migrate_ownership_emptydir.go` 의 `dirManagedProvenance`, `startOfRunRecords`, `emptyDirTree`; M5 의 정리 단계는 일부러 `os.Remove` 를 하므로 파일 전체가 범위가 아니다) 훑고, 이름 붙은 함수가 실제로 있는지, 훑은 파일 수가 기대와 같은지를 먼저 단언한다. 일곱 가지 변이 소스 (쓰기 호출 다섯, 매니페스트 `Save`, `probeCaseSensitiveFS`) 앞에서 붉고 읽기 호출 앞에서는 붉지 않음을 시험이 직접 관측한다.

**AC-OWR-16 RED 셀 (트리 `2d6ec34053`, A 커밋 직후, 작업 트리 깨끗, 부하 평균 `48.98 49.91 46.04`)** — 판정 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestDryRunChangesNothingAndMatchesVerdict$' -count=1 -v`, 종료 코드 1 (인용은 시험 로그 줄과 `---`/결과 줄만, 원문 순서 그대로):

```text
    migrate_ownership_preview_test.go:314: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/agents/moai/manager-git.md
    migrate_ownership_preview_test.go:317: the dry run changed the tree (paths whose bytes, mode or directory mtime moved): [.moai/]
    migrate_ownership_preview_test.go:342: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/agents/moai/manager-git.md
    migrate_ownership_preview_test.go:342: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/agents/moai/manager-spec.md
    migrate_ownership_preview_test.go:342: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/skills/moai-foundation-core/SKILL.md
    migrate_ownership_preview_test.go:342: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/skills/moai-foundation-quality/SKILL.md
    migrate_ownership_preview_test.go:345: the dry run changed the tree: [.moai/]
    migrate_ownership_preview_test.go:356: the dry run changed the tree: [.moai/]
--- FAIL: TestDryRunChangesNothingAndMatchesVerdict (11.47s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/UnmappedProjectAnnouncesExactlyTheEligibleSet (3.22s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/ProfileMappedKeptByProjectAnnouncesNothing (1.65s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/SwitchOffAnnouncesNothing (1.03s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/MutantOwnJudgmentIsCaught (2.67s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/MutantCreateThenDeleteIsSeenByTheDigest (2.89s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	13.407s
FAIL
```

이유 (원문에서 읽은 것만): (1) 미리보기가 판정 함수가 유지하는 파일을 후보로 알린다 — 사용자 대응물이 확인되지 않은 `manager-git.md` (줄 314), 그리고 프로필 지도 + `KeptBy` 프로젝트에서는 판정이 유지하는 네 파일 전부 (줄 342). 이것이 research.md G-3 의 예측 (미리보기 판정이 `KeptBy` 와 대응물 확인을 보지 않는다) 의 첫 관측이다. 반대 방향 (판정이 제거 대상으로 보는데 알리지 않은 파일) 은 없었다. (2) **디렉터리 mtime 센티널이 `.moai/` 한 곳에서 붉다** — 세 서브 시험 (스위치 켠 두 프로젝트와 스위치를 끈 프로젝트) 모두에서 같은 하나의 경로 `[.moai/]` 만 갈라진다 (줄 317, 345, 356). 이 시험이 고치려는 판정/미리보기 코드와 무관하게 일어나므로 원인을 따로 쟀다 (아래 「관측」).

판정 명령 2 (기존 분기의 같은 성질을 지키는 가드, GREEN 유지) — `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestUpdateDryRun_ZeroMutation$' -count=1 -v`, 종료 코드 0:

```text
--- PASS: TestUpdateDryRun_ZeroMutation (0.08s)
    --- PASS: TestUpdateDryRun_ZeroMutation/v2.16.0 (0.07s)
    --- PASS: TestUpdateDryRun_ZeroMutation/3.0.1 (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.202s
```

**`TestDryRunRunsNoBinaryUpdateStep` RED (같은 트리), 종료 코드 1:**

```text
    migrate_ownership_preview_test.go:425: a dry run called the binary installer 1 time(s)
--- FAIL: TestDryRunRunsNoBinaryUpdateStep (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.190s
FAIL
```

**`TestVerdictAndPreviewCodeHasNoWritableCalls` RED (같은 트리), 종료 코드 1:**

```text
    migrate_ownership_preview_test.go:499: read migrate_ownership_env.go: open migrate_ownership_env.go: no such file or directory
    migrate_ownership_preview_test.go:532: migrate_project_assets.go: found 0 of the 1 functions the scan names
    migrate_ownership_preview_test.go:540: scanned 7 files, want 8
--- FAIL: TestVerdictAndPreviewCodeHasNoWritableCalls (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.188s
FAIL
```

(이유: 새 공유 파일 `migrate_ownership_env.go` 와 함수 `newOwnershipEnv` 가 아직 없다. 존재하지 않는 대상을 훑고 통과하는 일이 없도록 이름 붙은 대상의 존재와 훑은 파일 수를 먼저 단언한 결과이다.)

**관측: `moai update --dry-run` 자체가 REQ-OWR-011 을 두 곳에서 어긴다 (이 카드의 코드와 무관하게 이미 있던 동작; 이 카드가 지금 고치기로 한 것은 아래 C2, C3).** 둘 다 `internal/cli/update.go` 의 `runUpdate` 안, dry-run 분기 앞에 있다.

1. **업데이트 잠금.** `acquireUpdateLock` 이 `.moai/.update.lock` 을 만들었다가 반환 때 지운다. 항목이 만들어졌다 사라지면 `.moai/` 의 mtime 이 고정 과거 시각에서 벗어난다 — AC-OWR-16 의 센티널이 정확히 이런 「만들었다 지우는」 쓰기를 잡으려고 설계되었고, 임시 진단 시험 (커밋하지 않고 삭제) 으로 이 트리에서 `ownTreeDiff` 가 `[.moai/]` 하나임을 확인했다 (위 RED 줄 317 과 같은 값). 기존 `TestUpdateDryRun_ZeroMutation` 은 디렉터리 mtime 을 보지 않아 이를 못 본다 (위 가드 명령이 PASS 인 이유).
2. **이진 자체 업데이트.** 릴리스 모양의 버전에서 `shouldSkipBinaryUpdate` 는 `--templates-only`, `MOAI_SKIP_BINARY_UPDATE=1`, 개발 빌드만 건너뛰고 `--dry-run` 은 보지 않는다. 임시 진단 시험 (커밋하지 않고 삭제; 설치를 거부하는 가짜 설치기) 의 원문: `version=v3.2.0-rc.36 IsDevBuild=false dry-run=true err=<nil> UpdateAvailable-checks=1 Update()-calls=1`. 즉 `--dry-run` 이 설치기를 한 번 부른다. 성공하면 새 바이너리를 설치하고 재실행으로 이어진다 (`update.go` 의 「Step 1」 이 「dry-run 분기」 앞). 개발 빌드를 쓰는 기존 시험은 이 단계를 건너뛰므로 보지 못한다.

**SPEC 쪽 판정.** REQ-OWR-011 은 「명령 — 그것이 하는 모든 단계 — 은 어떤 파일이나 디렉터리도 만들거나 바꾸거나 지우지 않는다」 라고 쓴다. 잠금과 이진 설치는 그 문구를 어기는 현재 동작이므로 AC 를 약화하지 않고 (센티널에서 `.moai/` 를 면제하지 않고) 제품을 고치는 쪽을 택했다. 두 수정은 `update.go` 의 각각 한 줄이고 서로 별개의 커밋이라 리더가 따로 되돌릴 수 있다. 리더가 다른 판단을 하면: C2 를 되돌리면 AC-OWR-16 의 `.moai/` 줄이 다시 붉고, C3 를 되돌리면 `TestDryRunRunsNoBinaryUpdateStep` 만 붉다.

**커밋 (M7).** A `2d6ec34053` (시험만) → B `0065cb21ed` (위 RED 셀) → C1 `4683bed8e0` (미리보기가 판정 함수를 부른다) → C2 `d09d66c6f9` (`--dry-run` 이 업데이트 잠금을 잡지 않는다) → C3 `4767df4de4` (`--dry-run` 이 이진 설치 단계에 닿지 않는다). C 뒤에 시험만 더한 커밋 둘이 있다: `96cdad896d` (아래 「센티널 강화」), `84eaf40383` (아래 「fail-closed 가장자리」). 최종 HEAD `84eaf4038328ce95403e7593504a8108b34cdf82`.

**C1 — 운영 코드 변경.** `internal/cli/migrate_ownership_env.go` 신규 (`manifestView`, `viewOfManager`, `viewOfManifest`, `previewVerdicts`, `buildPreviewEnv`), `internal/cli/migrate_project_assets.go` (마이그레이션의 판정 환경 조립을 `newOwnershipEnv`/`ownershipEnvInputs` 로 옮김 — 프로필 원장을 읽는 `profile.ResolveLaunchProfileForProject` 호출은 같은 파일에 그대로 한 곳이고, 새 파일에는 프로필 심볼이 없다), `internal/cli/update_dryrun_preview.go` (자체 판정 `migrationRemovalCandidate` 삭제, `verdicts.removalEligible` 호출로 교체, 읽기 전용 로더 `update.LoadManifestReadOnly` 유지). 환경은 첫 질문에서 늦게 조립하고, 제거 스위치가 꺼져 있으면 (`removalEnabled &&` 가 앞서 거짓) 사용자 홈을 전혀 읽지 않는다. 홈, 번들 템플릿, 카탈로그, 사용자 매니페스트, 프로젝트 루트를 읽지 못하면 후보를 하나도 알리지 않는다 (제거를 짐작하지 않음).

**C1 만 적용한 상태 (트리 `4683bed8e0`, 저장된 원문)** — 후보 비교는 초록이 되고 `.moai/` 와 이진 설치만 붉다:

```text
    migrate_ownership_preview_test.go:317: the dry run changed the tree (paths whose bytes, mode or directory mtime moved): [.moai/]
    migrate_ownership_preview_test.go:345: the dry run changed the tree: [.moai/]
    migrate_ownership_preview_test.go:356: the dry run changed the tree: [.moai/]
--- FAIL: TestDryRunChangesNothingAndMatchesVerdict (9.33s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/UnmappedProjectAnnouncesExactlyTheEligibleSet (2.54s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/ProfileMappedKeptByProjectAnnouncesNothing (2.13s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/SwitchOffAnnouncesNothing (1.45s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/MutantOwnJudgmentIsCaught (2.26s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/MutantCreateThenDeleteIsSeenByTheDigest (0.94s)
    migrate_ownership_preview_test.go:425: a dry run called the binary installer 1 time(s)
--- FAIL: TestDryRunRunsNoBinaryUpdateStep (0.01s)
--- PASS: TestVerdictAndPreviewCodeHasNoWritableCalls (0.01s)
--- PASS: TestUpdateDryRun_ZeroMutation (0.01s)
    --- PASS: TestUpdateDryRun_ZeroMutation/v2.16.0 (0.01s)
    --- PASS: TestUpdateDryRun_ZeroMutation/3.0.1 (0.01s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.00s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.018s
FAIL
```

**최종 GREEN (트리 `84eaf4038328`, 작업 트리 깨끗, 부하 평균 `24.46 44.24 56.49`)**

판정 명령 1 — `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestDryRunChangesNothingAndMatchesVerdict$' -count=1 -v`, 종료 코드 0:

```text
--- PASS: TestDryRunChangesNothingAndMatchesVerdict (8.70s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/UnmappedProjectAnnouncesExactlyTheEligibleSet (1.18s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/ProfileMappedKeptByProjectAnnouncesNothing (1.28s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/SwitchOffAnnouncesNothing (1.63s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/MutantOwnJudgmentIsCaught (2.52s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/MutantRootProbeIsSeenByTheDigest (0.87s)
    --- PASS: TestDryRunChangesNothingAndMatchesVerdict/MutantCreateThenDeleteIsSeenByTheDigest (1.21s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	10.135s
```

판정 명령 2 (가드) — `... -run '^TestUpdateDryRun_ZeroMutation$' ...`, 종료 코드 0:

```text
--- PASS: TestUpdateDryRun_ZeroMutation (0.01s)
    --- PASS: TestUpdateDryRun_ZeroMutation/v2.16.0 (0.01s)
    --- PASS: TestUpdateDryRun_ZeroMutation/3.0.1 (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.096s
```

같은 트리의 나머지 둘 — `... -run '^(TestDryRunRunsNoBinaryUpdateStep|TestVerdictAndPreviewCodeHasNoWritableCalls)$' ...`, 종료 코드 0:

```text
--- PASS: TestDryRunRunsNoBinaryUpdateStep (0.01s)
--- PASS: TestVerdictAndPreviewCodeHasNoWritableCalls (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.012s
```

**연속 5회씩 (같은 트리, 각 실행의 종료 코드를 직접 읽음)**

```text
TestDryRunChangesNothingAndMatchesVerdict run 1 rc=0 top-pass=1 sub-pass=6 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	24.000s
TestDryRunChangesNothingAndMatchesVerdict run 2 rc=0 top-pass=1 sub-pass=6 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	16.977s
TestDryRunChangesNothingAndMatchesVerdict run 3 rc=0 top-pass=1 sub-pass=6 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	20.566s
TestDryRunChangesNothingAndMatchesVerdict run 4 rc=0 top-pass=1 sub-pass=6 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	26.012s
TestDryRunChangesNothingAndMatchesVerdict run 5 rc=0 top-pass=1 sub-pass=6 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	63.590s
TestDryRunRunsNoBinaryUpdateStep run 1 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	2.153s
TestDryRunRunsNoBinaryUpdateStep run 2 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	2.194s
TestDryRunRunsNoBinaryUpdateStep run 3 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.438s
TestDryRunRunsNoBinaryUpdateStep run 4 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.431s
TestDryRunRunsNoBinaryUpdateStep run 5 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.456s
TestVerdictAndPreviewCodeHasNoWritableCalls run 1 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.388s
TestVerdictAndPreviewCodeHasNoWritableCalls run 2 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.386s
TestVerdictAndPreviewCodeHasNoWritableCalls run 3 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.317s
TestVerdictAndPreviewCodeHasNoWritableCalls run 4 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.146s
TestVerdictAndPreviewCodeHasNoWritableCalls run 5 rc=0 top-pass=1 sub-pass=0 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	1.288s
```

**실제 운영 코드 변이 다섯 (트리 `96cdad896d` 기준, 임시 변경; 모두 되돌렸고 세 파일을 `cmp` 로 원본과 대조해 `RESTORED-IDENTICAL`; 명령은 `-run '^(TestDryRunChangesNothingAndMatchesVerdict|TestDryRunRunsNoBinaryUpdateStep|TestVerdictAndPreviewCodeHasNoWritableCalls)$' -count=1 -v` 의 줄 추림, 앞 여덟 줄)**

- **Pa** `previewVerdicts.removalEligible` 가 판정 없이 항상 참 (미리보기가 제 판정을 안 쓰는 최악의 경우):

```text
    migrate_ownership_preview_test.go:314: preview vs verdict: announced as a migration candidate but not removal-eligible: .agents/skills/moai-custom-x/SKILL.md
    migrate_ownership_preview_test.go:314: preview vs verdict: announced as a migration candidate but not removal-eligible: .agents/skills/moai-plan/SKILL.md
    migrate_ownership_preview_test.go:314: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/agents/moai/manager-git.md
    migrate_ownership_preview_test.go:314: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/skills/moai-custom/SKILL.md
    migrate_ownership_preview_test.go:314: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/skills/moai-ref-owasp-checklist/SKILL.md
    migrate_ownership_preview_test.go:314: preview vs verdict: announced as a migration candidate but not removal-eligible: .claude/skills/moai-workflow-tdd/SKILL.md
    migrate_ownership_preview_test.go:342: preview vs verdict: announced as a migration candidate but not removal-eligible: .agents/skills/moai-custom-x/SKILL.md
    migrate_ownership_preview_test.go:342: preview vs verdict: announced as a migration candidate but not removal-eligible: .agents/skills/moai-plan/SKILL.md
```

- **Pb** `previewReconciliation` 이 프로젝트 루트에 `probeCaseSensitiveFS(projectRoot)` 를 부름 (기준이 이름으로 든 변이):

```text
    migrate_ownership_preview_test.go:317: the dry run changed the tree (paths whose bytes, mode or directory mtime moved): [./]
    migrate_ownership_preview_test.go:345: the dry run changed the tree: [./]
    migrate_ownership_preview_test.go:356: the dry run changed the tree: [./]
--- FAIL: TestDryRunChangesNothingAndMatchesVerdict (14.06s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/UnmappedProjectAnnouncesExactlyTheEligibleSet (2.04s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/ProfileMappedKeptByProjectAnnouncesNothing (2.58s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/SwitchOffAnnouncesNothing (2.30s)
    migrate_ownership_preview_test.go:526: writable call in verdict/preview code: update_dryrun_preview.go: previewReconciliation calls probeCaseSensitiveFS
```

- **Pc** `--dry-run` 이 업데이트 잠금을 다시 잡음 (C2 를 되돌린 것과 같다):

```text
    migrate_ownership_preview_test.go:317: the dry run changed the tree (paths whose bytes, mode or directory mtime moved): [.moai/]
    migrate_ownership_preview_test.go:345: the dry run changed the tree: [.moai/]
    migrate_ownership_preview_test.go:356: the dry run changed the tree: [.moai/]
--- FAIL: TestDryRunChangesNothingAndMatchesVerdict (12.94s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/UnmappedProjectAnnouncesExactlyTheEligibleSet (2.18s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/ProfileMappedKeptByProjectAnnouncesNothing (2.02s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/SwitchOffAnnouncesNothing (2.01s)
FAIL
```

- **Pd** `buildPreviewEnv` 가 루트에 `.probe` 를 쓰고 지움:

```text
    migrate_ownership_preview_test.go:317: the dry run changed the tree (paths whose bytes, mode or directory mtime moved): [./]
    migrate_ownership_preview_test.go:345: the dry run changed the tree: [./]
--- FAIL: TestDryRunChangesNothingAndMatchesVerdict (20.10s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/UnmappedProjectAnnouncesExactlyTheEligibleSet (2.77s)
    --- FAIL: TestDryRunChangesNothingAndMatchesVerdict/ProfileMappedKeptByProjectAnnouncesNothing (2.95s)
    migrate_ownership_preview_test.go:526: writable call in verdict/preview code: migrate_ownership_env.go: buildPreviewEnv calls os.WriteFile
    migrate_ownership_preview_test.go:526: writable call in verdict/preview code: migrate_ownership_env.go: buildPreviewEnv calls os.Remove
--- FAIL: TestVerdictAndPreviewCodeHasNoWritableCalls (0.01s)
```

- **Pe** `shouldSkipBinaryUpdate` 가 `--dry-run` 을 보지 않음 (C3 를 되돌린 것과 같다):

```text
    migrate_ownership_preview_test.go:443: a dry run called the binary installer 1 time(s)
--- FAIL: TestDryRunRunsNoBinaryUpdateStep (0.02s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	16.343s
FAIL
```

**센티널 강화 (정직 공개, 커밋 `96cdad896d`, 시험만).** Pd 를 처음 돌렸을 때 소스 스캔만 붉고 트리 해시는 초록이었다 — 프로젝트 루트 디렉터리 바로 아래에 만들었다 지운 파일이 루트 디렉터리 자신의 mtime 만 바꾸는데, M0 의 `ownTreeSnapshot` 이 루트(`.`)를 건너뛰어 그것을 기록하지 않았기 때문이다. 기준 본문은 정확히 그 변이(`probeCaseSensitiveFS(projectRoot)` 가 프로젝트 루트 mtime 을 바꾼다)를 이름으로 든다. 그래서 `DirMtime` 일 때 루트 자신의 모드와 mtime 을 `./` 항목으로 기록하게 고치고, 도우미 자기 시험 `TestOwnershipHelperTreeSnapshotSeesDirMtime` 에 루트 직속의 만들었다-지움 사례를 더했고, AC-OWR-16 시험에 그 변이(`MutantRootProbeIsSeenByTheDigest`)를 더했다. 이 강화는 C1-C3 뒤에 들어왔다: 강화된 센티널로 AC-OWR-16 을 다시 쟀고 (위 최종 GREEN 이 강화 뒤 값) 루트 mtime 은 실제 `--dry-run` 에서 움직이지 않는다. 위 Pb, Pd 의 `[./]` 줄이 강화 효과의 원문이다.

**fail-closed 가장자리 (커밋 `84eaf40383`, 시험만).** `migrate_ownership_env.go` 의 범위가 85% 미만 (83.3%) 이어서 환경을 못 읽을 때 아무것도 알리지 않는 동작과 중첩 정리의 보호를 시험했다: `TestPreviewVerdictsFailClosed` (도달 대조군 — 모든 입력이 읽히면 알림; 홈 없음, 깨진 사용자 매니페스트, 풀리지 않는 프로젝트 루트, 매니페스트 없음 → 알림 없음; 환경은 한 번만 만들어 재사용), `TestViewOfManifestIsAReadOnlyCopy`, `TestPruneNestedEmptyDirGuards` (빈 디렉터리 제거, 내용 있는 디렉터리 유지, 루트 밖 빈 디렉터리 거절, 없는 디렉터리). 변이 둘을 돌렸다: Qb (중첩 정리가 루트 경계 검사를 버림) 는 붉다 (`migrate_ownership_env_test.go:176: pruned=true err=<nil>, want the directory outside the root refused`, `--- FAIL: TestPruneNestedEmptyDirGuards/RefusesAnEmptyDirectoryOutsideTheRoot`). Qa (깨진 사용자 매니페스트를 실패로 보지 않고 빈 매니페스트로 읽음) 는 **붉어지지 않았다 (`ok`)**: 빈 사용자 매니페스트는 어떤 대응물도 확인하지 못하므로 「실패 닫힘」 과 「빈 것으로 읽음」 이 이 입력에서 같은 결과(알림 없음)를 낸다 — 동등 변이이다. 그 서브 시험은 메커니즘이 아니라 결과를 단언한다.

**회귀 (앵커된 이름 60개, 최종 트리 `84eaf40383`)** — 기준선 11 + chunk 1, 2 의 시험 + M5 둘 (`TestManagedOnlyEmptyDirsPruned`, `TestEmptyDirProvenanceVerdict`) + M6 둘 + M7 셋과 새 시험 셋 + 가드 (`TestUpdateDryRun_ZeroMutation`, `TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy`) + `shouldSkipBinaryUpdate` 의 기존 시험 넷; 이름 수는 `sort -u | wc -l` 로 60. 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(<60개>)$' -count=1 -v -coverprofile=...`: `exit=0`, `top-level PASS=60 FAIL=0 SKIP=0`, `ok  	github.com/modu-ai/moai-adk/internal/cli	82.741s`. `internal/manifest` 의 `TestManifestRecordsDeployedVersion`: 직전 측정 (트리 `96cdad896d`) 에서 `--- PASS`.

**범위가 좁은 구문 커버리지 (위 회귀 한 번의 프로필; 패키지 전체가 아니라 앵커된 60개만이므로 그 파일을 덮는 다른 시험은 세지 않는다)**

```text
migrate_ownership_emptydir.go 93.3% (84/90 statements)
migrate_ownership_env.go 94.4% (34/36 statements)
update_dryrun_preview.go 86.1% (31/36 statements)
migrate_ownership_remaining.go 100.0% (67/67 statements)
migrate_project_assets.go 82.8% (216/261 statements)
update.go 35.7% (146/409 statements)
```

이 chunk 가 새로 만든 파일 넷은 모두 85% 이상이다. `migrate_project_assets.go` 와 `update.go` 는 대부분 이전에 있던 코드이고 이 60개 밖의 시험이 덮는 부분이 있어 85% 판정에 쓸 수 없다 (측정 갭; 패키지 전체 측정은 이 chunk 의 금지 범위이고 M10 과 CI 의 몫).

**품질 (최종 트리 `84eaf40383`, 각 명령의 종료 코드 직접 읽음)**

```text
go build ./... exit=0
GOOS=windows go build ./... exit=0
gofmt -l exit=0 (출력 없음)
go vet exit=0 (internal/cli, internal/manifest)
GOOS=windows go vet ./internal/cli/ exit=0 (시험 파일 포함)
golangci-lint run ./internal/manifest/ ./internal/cli/ exit=1:
		fmt.Fprintf(warn, "factory leader resume: run %s holds live legacy record %q from a binary before the leader/lane rename — adopting nothing; end its sessions and retire with 'moai factory runs --ret
		           ^
3 issues:
* errcheck: 3
```

린트가 보고한 셋은 모두 이 chunk 가 만지지 않은 `internal/cli/factory_run_adopt.go` 의 기존 errcheck 이다 (chunk 2 기록과 같은 셋). chunk 3 파일에 대한 린트 지적은 0건이다 (`grep` 으로 chunk 3 파일명을 찾아 출력 없음을 확인).

**하드 제약 확인 (최종 HEAD)**

```text
git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- internal/config/protected_zone.go internal/config/userroots.go internal/profile   → 빈 출력, exit=0
git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- .moai/specs | wc -l                                                                    → 22 (양성 대조군)
git diff --name-only cdd72a814..HEAD | grep -E "protected_zone|userroots|^internal/profile/"                                                      → none
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
removalEnabled = true 를 대입하는 비시험 코드                                                                                                       → 없음 (grep exit 1)
```

**M7 정직한 회계 / 갭 / 잔여 위험**

1. **C2, C3 는 `internal/cli/update.go` 의 제품 동작 변경이다.** M7 의 서술된 운영 파일 (`update_dryrun_preview.go`) 밖이다. REQ-OWR-011 의 「그 명령이 하는 모든 단계」 문구와 AC-OWR-16 의 센티널이 요구해서 했고, 한 줄씩이며 각각 별도 커밋이다. 되돌리면: C2 → `.moai/` 줄이 AC-OWR-16 에서 다시 붉다 (Pc), C3 → `TestDryRunRunsNoBinaryUpdateStep` 만 붉다 (Pe). 리더가 다른 판단을 하면 커밋 단위로 되돌릴 수 있다.
2. **C2 의 부작용 (코드 읽기로 안 사실, 별도 시험 없음):** 다른 업데이트가 잠금을 들고 있는 동안에도 `--dry-run` 은 이제 `ErrUpdateLockHeld` 로 거절되지 않고 진행한다 (읽기만 하므로 해가 없으나 미리보기가 다른 실행의 중간 상태를 읽을 수 있다). 잠금 자체의 시험 (`acquireUpdateLock` 직접 호출) 은 영향이 없다.
3. **C3 의 부작용:** `--dry-run` 은 더 이상 새 버전 확인 줄이나 이진 설치를 하지 않는다. 개발 빌드를 쓰는 기존 시험은 이 단계를 원래 건너뛰므로 영향이 없고, 릴리스 모양의 버전 시험은 위 새 시험뿐이다.
4. **디렉터리 mtime 방식의 한계는 기준 본문이 이미 적은 것과 같다:** 모드만 바꿨다 되돌리는 쓰기와 프로젝트 밖(HOME 등) 으로의 쓰기는 이 센티널이 보지 못한다. 미리보기는 사용자 홈과 프로필 원장을 읽기만 한다 (쓰기 가능 호출 없음을 스캔이 지킨다).
5. **미리보기가 보는 후보는 reconcile 계획의 `UserOwned` 중 공통 자산 청소 대상뿐이다** (`removalEnabled && isCommonAssetCleanTarget && 판정`). 판정이 제거 대상으로 보지만 계획이 `UserOwned` 로 분류하지 않는 파일은 알리지 않는다. 시험은 픽스처의 제거 대상이 전부 `UserOwned` 안에 있음을 먼저 단언한다 (이 픽스처에서는 성립); 다른 모양의 프로젝트에서 성립하는지는 측정하지 않았다.
6. 판정의 `removal-eligible` 은 실행 시점 조건 (백업 사본 쓰기 성공, OS 읽기·삭제 오류) 을 포함하지 않는다 (spec.md § B). 그래서 미리보기가 알린 후보와 실제 제거 목록은 실행 때 백업이 실패하면 다를 수 있다 (REQ-OWR-016 이 그때 제거 0건).
7. 제거 스위치가 운영에서 항상 꺼져 있으므로 (`removalEnabled = false`) 미리보기의 새 환경 조립은 운영에서 아직 한 번도 돌지 않는다; 시험 안에서만 스위치를 켠다.
8. C2, C3 의 단계별 중간 결과(C1 직후와 C3 직후의 시험 출력)는 터미널에서 봤으나 파일로 저장하지 않았다. 그 상태의 증거는 위 Pc (C2 되돌림) 와 Pe (C3 되돌림) 변이 실행이다.
9. 최종 GREEN 의 AC-OWR-16 판정 명령 1 은 `-count=1` 한 번과 별도 5회 반복이다. 이 머신의 부하 평균은 대략 24에서 90 사이였다.

### Run chunk 4 (M8, M9, M10) — 기록 시작

시작 HEAD `cf2581728e`, 브랜치 `WT-update-ownership-repair`, 작업 트리 깨끗. 기록 규약은 앞 chunk 와 같다. 이번 chunk 부터는 모든 셀을 저장한 출력 파일에서 스크립트로 옮겼고 손으로 다시 치지 않았다 (종료 코드는 각 실행에서 직접 읽음). 시험은 `TMPDIR=/private/tmp/own-fixture-d1` 아래 임시 디렉터리와 임시 HOME 만 썼다.

**lane 결정 (chunk 4 시작에 기록)**

1. chunk 3 이 더한 `internal/cli/update.go` 의 두 변경 — `--dry-run` 이 업데이트 잠금을 잡지 않는다 (`d09d66c6f9`), `--dry-run` 이 이진 설치 단계에 닿지 않는다 (`4767df4de4`) — 은 받아들인다. REQ-OWR-011 이 dry-run 은 어떤 파일이나 디렉터리도 만들거나 지우지 않는다고 요구하기 때문이다. sync 단계가 알아야 할 사용자 가시 부작용 둘: (a) 다른 업데이트가 잠금을 들고 있는 동안에도 `--dry-run` 이 `ErrUpdateLockHeld` 로 거절되지 않고 진행하므로 그 실행의 중간 상태를 읽을 수 있다. (b) `--dry-run` 은 새 버전 확인과 이진 자체 업데이트를 더 이상 하지 않는다.
2. REQ-OWR-006 / AC-OWR-10 의 에이전트 문구 틈 (프로젝트 배포기는 에이전트를 쓰지 않고 사용자 자산 단계가 쓴다) 은 공개된 부채로 둔다. SPEC 본문은 고치지 않는다.

#### M8 — AC-OWR-15 `TestRemovedAssetBackupRestoreByteIdentical`

**커밋 A `415644baff`** (`internal/cli/migrate_ownership_restore_test.go` 한 파일, 운영 코드 변경 없음). 작업 트리 깨끗, 부하 평균 35.63 35.87 34.58.

**도착 셀 — 이 기준은 도착 때 초록이다 (RED-now 없음).** acceptance.md 가 적은 RED 이유 (「매니페스트 사본이 실행 백업에 없어 (ii) (iii) 가 붉다」) 는 M1b 가 `manifest.prior.json` 을 만들면서 이미 닫혔다. 그래서 이 기준은 release-blocking 이 아니라 regression-guard 로 도착하고, 붉음은 변이에서 관측한다 (AC-OWR-08 과 같은 처리). 판정 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemovedAssetBackupRestoreByteIdentical$' -count=1 -v`, 종료 코드 0, 시험 로그 줄과 `---` 줄만 원문 순서대로:

```text
    migrate_ownership_restore_test.go:593: observation: manifest.prior.json has mode 600, the manifest it copies had 600 (the file modes of the restore area are asserted for the removed files only)
    migrate_ownership_restore_test.go:710: difference class manifest   : 1 path(s)
    migrate_ownership_restore_test.go:710: difference class run-backup : 93 path(s)
    migrate_ownership_restore_test.go:710: difference class logs       : 0 path(s)
    migrate_ownership_restore_test.go:710: difference class state      : 0 path(s)
    migrate_ownership_restore_test.go:710: difference class OUTSIDE-E  : 0 path(s)
--- PASS: TestRemovedAssetBackupRestoreByteIdentical (4.43s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/ReachedReconciliation (0.00s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/BackupCopiesAreByteIdentical (0.00s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/PriorManifestIsTheStartManifest (0.00s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/RestoreAreaIsByteIdentical (0.00s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/DifferenceSentinel (0.00s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/SharedClockNamesJudgeTheSame (0.13s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/NormaliserErasesOnlyClockDerivedText (0.00s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/ExpectedSetMembersThatDiffer (0.03s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantCRLFInBackupCopyIsCaught (0.17s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantTruncatedBackupCopyIsCaught (0.15s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantMissingBackupCopyIsCaught (0.18s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantBackupCopyModeNarrowedIsCaught (0.22s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantReserialisedPriorManifestIsCaught (0.14s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantStrayFileInRunBackupIsCaught (0.15s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantStrayRemovedAssetIsCaught (0.26s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantRunAlsoDeletesTheUsersFileIsCaught (0.14s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantRunAlsoRewritesAnotherFileIsCaught (0.13s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantRunNarrowsAFileModeIsCaught (0.24s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantRunLeavesAFileInTheNamespaceBackupChangedIsCaught (0.14s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantRunWritesAStateFileIsCaught (0.20s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantManifestNotRestoredIsCaught (0.14s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantSkippedRunIsCaught (0.08s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantUnrestoredTreeIsCaught (0.04s)
    --- PASS: TestRemovedAssetBackupRestoreByteIdentical/MutantControlRunThatRemovedFilesIsCaught (0.03s)
PASS
```

**연속 5회 (같은 트리 `415644baff`, 종료 코드는 각 실행에서 읽음)**

```text
TestRemovedAssetBackupRestoreByteIdentical run 1 rc=0 top-pass=1 sub-pass=24 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	5.284s
TestRemovedAssetBackupRestoreByteIdentical run 2 rc=0 top-pass=1 sub-pass=24 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	6.526s
TestRemovedAssetBackupRestoreByteIdentical run 3 rc=0 top-pass=1 sub-pass=24 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	7.674s
TestRemovedAssetBackupRestoreByteIdentical run 4 rc=0 top-pass=1 sub-pass=24 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	16.645s
TestRemovedAssetBackupRestoreByteIdentical run 5 rc=0 top-pass=1 sub-pass=24 fail=0 skip=0 ok  	github.com/modu-ai/moai-adk/internal/cli	6.891s
```

**시험 안의 변이 (순수 단언이 변이 입력 앞에서 붉음을 시험이 직접 관측, 위 PASS 줄의 `Mutant*IsCaught`)**: 백업 사본 줄 끝 변경, 절반 잘림, 사본 없음, 사본 모드 0600, 매니페스트 사본을 2칸 들여쓰기로 재직렬화, 실행 백업에 낯선 파일, 낯선 제거 자산, 사용자 파일도 같이 삭제, 다른 파일 내용 재작성, 다른 파일 모드 축소, 네임스페이스 백업 안의 파일 변경, `.moai/state` 에 파일 추가, 매니페스트 미복원 (모두 해당 절의 단언이 붉음). 도달 변이 셋: 건너뛴 실행 1, 복원하지 않은 트리, 제거를 한 통제 실행.

**운영 코드에 넣은 임시 변이 넷 (커밋 `415644baff` 기준, 변이 뒤 파일을 `cmp` 로 원본과 대조해 `RESTORED-IDENTICAL`; 위 시험 전체를 돌린 출력의 줄 추림, 같은 문제 줄은 한 번만)**

- **Q1** `preservePriorManifest` 가 사본을 쓰기 전에 매니페스트를 2칸 들여쓰기로 재직렬화한다:

```text
    migrate_ownership_restore_test.go:615: (ii) manifest.prior.json differs from the manifest as it was before the run (2004 vs 2174 bytes)
    migrate_ownership_restore_test.go:615: (iii) the restored manifest is not the pre-run manifest (err=<nil>)
--- FAIL: TestRemovedAssetBackupRestoreByteIdentical (17.19s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/PriorManifestIsTheStartManifest (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/RestoreAreaIsByteIdentical (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/SharedClockNamesJudgeTheSame (0.08s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	18.555s
rc=1
```

- **Q2** `backupRemovedAsset` 가 바이트의 앞 절반만 쓴다:

```text
    migrate_ownership_restore_test.go:615: (i) the backup copy of .claude/skills/moai-foundation-core/SKILL.md differs from the file as it was before the run (9971 vs 19943 bytes)
    migrate_ownership_restore_test.go:615: (i) the backup copy of .claude/skills/moai-ref-owasp-checklist/SKILL.md differs from the file as it was before the run (4560 vs 9120 bytes)
    migrate_ownership_restore_test.go:615: (i) the backup copy of .claude/skills/moai-workflow-testing/SKILL.md differs from the file as it was before the run (5398 vs 10797 bytes)
    migrate_ownership_restore_test.go:615: (iii) the restored .claude/skills/moai-foundation-core/SKILL.md differs from its pre-run bytes
    migrate_ownership_restore_test.go:615: (iii) the restored .claude/skills/moai-ref-owasp-checklist/SKILL.md differs from its pre-run bytes
    migrate_ownership_restore_test.go:615: (iii) the restored .claude/skills/moai-workflow-testing/SKILL.md differs from its pre-run bytes
    migrate_ownership_restore_test.go:615: (iv) .claude/skills/moai-foundation-core/SKILL.md differs between the restored tree and the control tree and is outside the expected set
    migrate_ownership_restore_test.go:615: (iv) .claude/skills/moai-ref-owasp-checklist/SKILL.md differs between the restored tree and the control tree and is outside the expected set
    migrate_ownership_restore_test.go:615: (iv) .claude/skills/moai-workflow-testing/SKILL.md differs between the restored tree and the control tree and is outside the expected set
--- FAIL: TestRemovedAssetBackupRestoreByteIdentical (15.05s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/BackupCopiesAreByteIdentical (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/RestoreAreaIsByteIdentical (0.00s)
```

- **Q3** 제거 암이 파일을 제거할 때 사용자 파일 `.claude/skills/mine/SKILL.md` 도 지운다:

```text
    migrate_ownership_restore_test.go:615: (iv) .claude/skills/mine/SKILL.md differs between the restored tree and the control tree and is outside the expected set
--- FAIL: TestRemovedAssetBackupRestoreByteIdentical (10.10s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/DifferenceSentinel (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/SharedClockNamesJudgeTheSame (0.13s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/ExpectedSetMembersThatDiffer (0.11s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/MutantRunAlsoDeletesTheUsersFileIsCaught (0.26s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.580s
rc=1
```

- **Q4** `backupRemovedAsset` 가 사본을 `0o600` 으로 쓴다:

```text
    migrate_ownership_restore_test.go:615: (iv) .claude/skills/moai-foundation-core/SKILL.md differs between the restored tree and the control tree and is outside the expected set
    migrate_ownership_restore_test.go:615: (iv) .claude/skills/moai-ref-owasp-checklist/SKILL.md differs between the restored tree and the control tree and is outside the expected set
    migrate_ownership_restore_test.go:615: (iv) .claude/skills/moai-workflow-testing/SKILL.md differs between the restored tree and the control tree and is outside the expected set
    migrate_ownership_restore_test.go:623: (i-mode) the backup copy of .claude/skills/moai-foundation-core/SKILL.md has mode 600, the file had 644
    migrate_ownership_restore_test.go:623: (i-mode) the backup copy of .claude/skills/moai-ref-owasp-checklist/SKILL.md has mode 600, the file had 644
    migrate_ownership_restore_test.go:623: (i-mode) the backup copy of .claude/skills/moai-workflow-testing/SKILL.md has mode 600, the file had 644
    migrate_ownership_restore_test.go:623: (iii-mode) the restored .claude/skills/moai-foundation-core/SKILL.md has mode 600, it had 644
    migrate_ownership_restore_test.go:623: (iii-mode) the restored .claude/skills/moai-ref-owasp-checklist/SKILL.md has mode 600, it had 644
    migrate_ownership_restore_test.go:623: (iii-mode) the restored .claude/skills/moai-workflow-testing/SKILL.md has mode 600, it had 644
--- FAIL: TestRemovedAssetBackupRestoreByteIdentical (28.32s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/DifferenceSentinel (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/SharedClockNamesJudgeTheSame (0.18s)
```

**구현 커밋 C 없음.** 남은 차이가 없다: 이 기준이 요구하는 동작은 M1b 가 만든 것으로 이미 성립한다.

**관측과 정직한 회계**

1. **SPEC 의 기대 차이 집합 E 에 빠진 시계 파생 위치 둘을 이 시험이 처음 측정했다.** 갱신 한 번은 `.moai-backups/<YYYYMMDD_HHMMSS>` 말고도 `.moai/backups/update-<UTC 시각>/` (안에 `.moai-run`, `.complete` 표식, 내용이 시각) 와 `.moai/archive/files/update-migration/<YYYYMMDD_HHMMSS>/` 를 만든다. 정규화 없이 비교하면 통제 실행과 이름이 같은 초면 `.moai-run` 내용만, 다른 초면 디렉터리 이름이 E 밖 차이로 나온다 — 시각에 따라 결과가 갈리는 시험이 된다. 그래서 design.md D-8 L3 의 정규화 (실행 백업 디렉터리 이름 → `<RUN>`, 시각 → 표식 값 가리기) 를 이 두 위치에 같은 방식으로 적용했다 (`restoreNormalize`). 정규화 규칙 자체를 시험이 먼저 단언한다: 시각만 다른 두 입력은 같아지고, 두 번 정규화한 결과가 같고, 이 두 위치 밖의 이름과 값은 그대로이고, 네임스페이스 백업 안의 내용 차이는 살아남으며, 한 트리에 두 디렉터리가 있으면 합치지 않고 보고한다. 시각 독립: 통제 실행은 실행 1 이 쓴 모든 시계 이름의 초가 지난 뒤에 시작하게 기다려 (폴링) 두 실행의 이름이 다르게 만들고 (전제로 단언), 같은 이름인 경우는 통제 트리의 시계 디렉터리를 실행 1 의 이름으로 바꾼 복사본으로 만들어 `SharedClockNamesJudgeTheSame` 이 같은 단언을 돌린다. 두 분기가 모두 초록이다.
2. E 의 나머지 항목은 이 흐름에서 이렇게 측정됐다 (시험 로그 `difference class`): 매니페스트 1, 실행 백업 93 (두 이름 분기의 합), 로그 0, 상태 0, E 밖 0. 로그와 상태는 차이가 하나도 없으므로 E 가 그 둘을 허용하지만 이 시험은 아무것도 숨기지 않는다. 그래도 둘의 경로가 한쪽에만 있는 경우는 붉게 만든다 (절 (iv-c), 변이 `RunWritesAStateFile`).
3. **N4 잔여 (E 안의 차이와 파일 모드) 의 처분: 시험이 닿는 곳까지 닫았다.** E 안의 차이: 매니페스트 파일은 복원 영역 R 에 들어 있어 (iii) 이 바이트로 비교한다; 실행 백업은 통제 실행의 백업 배치에 더해 정확히 `removed-assets/<제거한 경로>` 만 더 가진다 (절 (iv-b)); 로그/상태는 경로 집합 (iv-c). 파일 모드: 스냅샷 서술자에 모드가 들어 있어 R 밖의 모드 차이는 (iv) 가 잡고, R 의 제거된 파일은 백업 사본과 복원본의 모드가 실행 전 모드와 같아야 한다 (`(i-mode)`, `(iii-mode)`). **닫히지 않은 것(시험이 못 본다):** `.moai/logs/**` 와 `.moai/state/**` 안의 내용 차이, 디렉터리 mtime·xattr·소유자, 프로젝트 밖 (HOME 등) 쓰기, 그리고 복원 절차가 시험 전용 복사이므로 실제 운영자가 `cp` 로 복원할 때의 모드.
4. 관측: 매니페스트와 그 사본 `manifest.prior.json` 은 모드가 같다 (둘 다 `0600`, 시험 로그 `observation:` 줄). SPEC 은 파일 모드를 보존 범위 밖으로 둔다 (spec.md § J) 이므로 시험은 매니페스트 모드를 단언하지 않고 로그로만 남긴다. 복원 절차가 매니페스트를 `0644` 로 쓰는 것은 시험 전용 절차의 선택이며 T0 와의 차이 판정에서는 매니페스트가 E 에 들어 있어 영향이 없다.
5. **N23 (plan.md M8 이 AC-OWR-15 (v) 도달 단언을 빠뜨림) 처분: 닫음.** 시험 절 (v) 이 실행 1 의 `skipped == false`, 새 `manifest.prior.json`, 제거 목록이 비어 있지 않음 (제거 후보 셋과 정확히 같음), 통제 실행이 같은 단계에 닿고 아무것도 제거하지 않음을 단언하고, 변이 `MutantSkippedRunIsCaught`, `MutantControlRunThatRemovedFilesIsCaught` 가 이를 붉게 만든다.
6. N17 (어느 해시를 비교하는지 밝히지 않음) 처분은 M9 와 함께 마지막에 정리한다. 이 시험의 트리 비교는 내용만 본다 (경로, 바이트, 모드, 디렉터리 존재; 디렉터리 mtime 은 보지 않음) — 시험 파일 머리 주석에도 적었다.

#### M9 — AC-OWR-17 `TestRepeatedUpdateIsIdempotent`

**커밋 A `5614767764`** (`internal/cli/migrate_ownership_recovery_test.go` 한 파일, 운영 코드 변경 없음; AC-OWR-18 은 같은 파일에 뒤따르는 별도 커밋). 작업 트리 깨끗, 부하 평균 24.69 30.50 32.95.

**도착 셀 — 이 기준도 도착 때 초록이다 (RED-now 없음).** acceptance.md 의 RED 이유 (「잔존 행이 없어 단언 불가, 배포 단계가 매 실행 `DeployedAt` 을 갱신하는 구현이 들어오면 붉어진다」) 는 M1b 와 M3 가 이미 닫았다: `DeployedAt` 은 `Version` 이 바뀔 때만 움직이고 (`recordManifestDeployedVersion`), 잔존 행은 M3 가 만든다. 그래서 regression-guard 로 도착하고 붉음은 변이에서 관측한다. 판정 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRepeatedUpdateIsIdempotent$' -count=1 -v`, 종료 코드 0, 결과 줄과 시험 로그 줄만 (원문 순서):

```text
=== RUN   TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 2 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 41 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 315, run 4 lists 315; 0 row(s) differ between them (reason refinement only)
=== RUN   TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 315, run 4 lists 315; 0 row(s) differ between them (reason refinement only)
=== RUN   TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 8, run 2 lists 8; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 8, run 3 lists 8; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 8, run 4 lists 8; 0 row(s) differ between them (reason refinement only)
=== RUN   TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 8, run 2 lists 8; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 8, run 3 lists 8; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 8, run 4 lists 8; 0 row(s) differ between them (reason refinement only)
--- PASS: TestRepeatedUpdateIsIdempotent (25.03s)
    --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories (6.37s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories/ReachedReconciliation (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories/ForcedRepeats (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories/ManifestBytes (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories/RemainingRows (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories/RowOrder (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/DistinctRunDirectories/PlainRepeatSkips (0.00s)
    --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory (6.58s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory/ReachedReconciliation (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory/ForcedRepeats (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory/ManifestBytes (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory/RemainingRows (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory/RowOrder (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/ManyUnregisteredCopies/SharedRunDirectory/PlainRepeatSkips (0.00s)
    --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories (8.77s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories/ReachedReconciliation (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories/ForcedRepeats (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories/ManifestBytes (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories/RemainingRows (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories/RowOrder (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/DistinctRunDirectories/PlainRepeatSkips (0.00s)
    --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory (3.30s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory/ReachedReconciliation (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory/ForcedRepeats (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory/ManifestBytes (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory/RemainingRows (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory/RowOrder (0.00s)
        --- PASS: TestRepeatedUpdateIsIdempotent/Mixed/SharedRunDirectory/PlainRepeatSkips (0.00s)
PASS
```

시험의 구성: 실행 1 은 일반, 실행 2·3 은 `--force`, 그리고 **실행 4 (`--force`) 를 하나 더 둔다** — 아래 「관측 1」 때문에 정상 상태를 증명하려고 더한 것이며 기준 본문의 셋을 줄이지 않는다. 그 뒤 일반 재실행 둘. 픽스처는 둘 (313개 미등록 복사본 + 제거 대상 하나, 그리고 섞인 픽스처: 등록 managed 다섯, 사용자 작성 하나, 미등록 다섯), 시계 분기는 둘 (구별되는 실행 디렉터리: 이전 실행이 쓴 모든 시계 이름의 초가 지나길 폴링으로 기다림; 공유 실행 디렉터리: 이전 실행의 백업 내용을 다음 실행 디렉터리에 합치는 쓰기 대역으로 같은 초의 최종 상태를 재현). 네 조합이 모두 같은 단언을 돈다. 규칙 7: 센 실행은 (i) 건너뛰지 않음 (ii) 그 실행 직후 읽은 새 `manifest.prior.json` 이 실행이 시작한 매니페스트와 같음 (iii) 잔존 행이 양수인 것만이고, 센 수가 몬 수와 같아야 한다. 일반 재실행의 「아무것도 쓰지 않음」 은 헬퍼의 mtime 창 (`began = now-1s`) 이 직전 실행의 사본을 새것으로 읽을 수 있어 쓰지 않고, 재실행 전후 `.moai-backups` 목록이 같음으로 단언한다.

**비교하는 해시는 내용만이다 (부채 N17 처분).** 경로, 바이트, 모드, 디렉터리 존재. 디렉터리 mtime 과 시계 값은 들어 있지 않다. 두 다른 실행을 비교할 때 매니페스트의 `deployed_at` 값만 가린다. 실행마다 제 몫을 더하는 시계 파생 위치는 빼고 비교한다: `.moai-backups/`, `.moai/backups/` (네임스페이스 백업), `.moai/archive/files/update-migration/`, `.moai/logs/`, `.moai/state/`. 기준 본문의 제외 목록은 앞의 것, 로그, 상태 셋뿐이어서 둘이 빠져 있다 — 이 시험이 처음 측정했다 (M8 기록의 관측 1 과 같은 두 위치).

**연속 5회 (같은 트리 `5614767764`, 종료 코드는 각 실행에서 읽음)**

```text
TestRepeatedUpdateIsIdempotent run 1 rc=0 top-pass=1 all-pass=41 fail=0 skip=0 
TestRepeatedUpdateIsIdempotent run 2 rc=0 top-pass=1 all-pass=41 fail=0 skip=0 
TestRepeatedUpdateIsIdempotent run 3 rc=0 top-pass=1 all-pass=41 fail=0 skip=0 
TestRepeatedUpdateIsIdempotent run 4 rc=0 top-pass=1 all-pass=41 fail=0 skip=0 
TestRepeatedUpdateIsIdempotent run 5 rc=0 top-pass=1 all-pass=41 fail=0 skip=0 
```

**시험 안의 변이 (순수 단언, `Mixed/DistinctRunDirectories` 아래 `Mutant*IsCaught` 열둘 모두 통과):** 매 실행 `DeployedAt` 갱신, 행이 맵 순서, 행 하나 빠짐, 행 하나의 이유 바뀜, 실행 2 건너뜀, 실행 3 이 사본을 안 씀, 실행 3 의 사본이 시작 매니페스트가 아님, 실행 3 에서 트리가 바뀜, 일반 재실행이 조정함, 일반 재실행이 트리를 다시 씀, 일반 재실행이 실행 백업을 씀, 일반 재실행끼리 출력이 다름.

**운영 코드에 넣은 임시 변이 넷 (커밋 `5614767764` 기준, 파일을 `cmp` 로 원본과 대조해 `RESTORED-IDENTICAL`, 위 시험 전체를 돌린 출력의 줄 추림)**

- **R1** `recordManifestDeployedVersion` 이 `Version` 이 같아도 `DeployedAt` 을 매번 갱신한다 (구별되는 디렉터리 분기가 초를 갈라 놓으므로 관측된다; 공유 분기만으로는 같은 초에 떨어지면 못 본다 — 그래서 두 분기가 있다):

```text
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 2 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 41 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 315, run 4 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:513: (manifest) the manifest bytes changed between run 2 and run 3
    migrate_ownership_recovery_test.go:513: (manifest) the manifest bytes changed between run 3 and run 4
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 8, run 2 lists 8; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 8, run 3 lists 8; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 8, run 4 lists 8; 0 row(s) differ between them (reason refinement only)
```

- **R2** `sortRemainingRows` 가 아무것도 하지 않는다 (행이 루트 순서로 나옴):

```text
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 2 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 41 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 315, run 4 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:513: (order) run 1 lists its rows out of path order
    migrate_ownership_recovery_test.go:513: (order) run 2 lists its rows out of path order
    migrate_ownership_recovery_test.go:513: (order) run 3 lists its rows out of path order
    migrate_ownership_recovery_test.go:513: (order) run 4 lists its rows out of path order
```

- **R3** `versionMatchSkipsSync` 가 항상 거짓이다 (같은 버전의 일반 재실행이 건너뛰지 않음):

```text
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 2 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 41 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 315, run 4 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:513: (plain) plain re-run 1 was not skipped
    migrate_ownership_recovery_test.go:513: (plain) plain re-run 1 produced a remaining list
    migrate_ownership_recovery_test.go:513: (plain) plain re-run 1 changed the run backups: [20261018_192840/.moai-run]
    migrate_ownership_recovery_test.go:513: (plain) plain re-run 2 was not skipped
    migrate_ownership_recovery_test.go:513: (plain) plain re-run 2 produced a remaining list
```

- **R4** 마이그레이션이 매 실행 `.moai/idempotence-probe` 를 새 값으로 쓴다:

```text
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 2 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 41 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 3 lists 315, run 4 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:513: (tree) runs 2 and 3 left different trees: [.moai/idempotence-probe]
    migrate_ownership_recovery_test.go:513: (tree) runs 3 and 4 left different trees: [.moai/idempotence-probe]
    migrate_ownership_recovery_test.go:501: remaining rows: run 1 lists 315, run 2 lists 315; 0 row(s) differ between them (reason refinement only)
    migrate_ownership_recovery_test.go:501: remaining rows: run 2 lists 315, run 3 lists 315; 0 row(s) differ between them (reason refinement only)
```

**관측 1 — REQ-OWR-012 의 「동일한 잔존 목록」 은 313개 픽스처의 앞 실행에서 이유 열에서만 어긋난다 (공개된 한계, 구현을 바꾸지 않음).** 시험 로그가 센 결과: 313개 픽스처, 구별되는 디렉터리 — 실행 1 → 2 에서 2행, 실행 2 → 3 에서 41행의 이유가 `no-provenance-record` 에서 `evidence-incomplete:original-deployed-path` 로 바뀌고, 실행 3 → 4 에서는 0행이다. 같은 픽스처의 공유 디렉터리 분기와 섞인 픽스처의 두 분기에서는 어느 쌍도 0행이다. 원인: design.md D-2 는 다른 키 아래 같은 해시의 기록이 보존된 이전 세대에 있으면 이유를 더 구체적인 `evidence-incomplete:original-deployed-path` 로 가른다. 보존 세대는 실행마다 하나씩 늘고, 실행 3 이 읽는 세대에는 첫 배포가 끝난 뒤의 매니페스트 (실행 2 의 시작 매니페스트) 가 처음 들어와 41행이 더 가려진다. 시작 매니페스트가 안정되고 나면 (실행 3 → 4) 목록은 이유까지 같다. 파일 목록, 분류 (class), 처분 (모두 보존) 은 모든 쌍에서 같다. 이 정제를 없애는 것은 AC-OWR-14 (`RenderedByTheUpdateFlow` 가 정제를 단언한다) 와 design.md D-2 를 바꾸는 일이라 이 마일스톤의 「M1-M5 가 남긴 차이만 고친다」 범위 밖이다. 시험은 이렇게 단언한다: 모든 쌍에서 같은 파일·같은 분류, 이유는 같거나 이 한 가지 정제 (`no-provenance-record` → `evidence-incomplete:original-deployed-path`) 뿐이고, 마지막 쌍 (정상 상태) 은 완전히 같아야 하며, 정제가 일어날 수 없는 경우 (섞인 픽스처, 공유 디렉터리) 는 모든 쌍이 완전히 같아야 한다. 기준 본문의 「셋 모두 같다」 를 그대로 단언하면 313개 픽스처에서 붉다 — 이것이 이 한계의 직접 증거이다. 처분: 공개된 한계 (disclosed-limit), 리더에게 보고.

**구현 커밋 C 없음.** 구현이 남은 차이가 없다 (관측 1 은 의도된 설계 D-2 의 결과이고 SPEC 본문 수정 사안이다).

#### M9 — AC-OWR-18 `TestInterruptedMigrationRecoversAndConverges`

**커밋 A `ace2e8d84b`** (`internal/cli/migrate_ownership_recovery_test.go` 에 시험만 추가, 운영 코드 변경 없음). 작업 트리 깨끗, 부하 평균 29.59 33.27 33.95.

**도착 셀 — 이 기준도 도착 때 초록이다 (RED-now 없음).** acceptance.md 의 RED 이유 (「(c) 에서 이미 비어 있던 디렉터리가 정리되지 않는다 — AC-OWR-13 의 같은 원인」) 는 M5 가, (a) 의 제거 0건은 M1b 가 이미 닫았다. regression-guard 로 도착하고 붉음은 변이에서 관측한다. 판정 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestInterruptedMigrationRecoversAndConverges$' -count=1 -v`, 종료 코드 0, 결과 줄과 시험 로그 줄만:

```text
    migrate_ownership_recovery_test.go:905: compared tree hash: content-only (path, bytes, mode, directory presence); directory mtimes and clock values are not compared; deployed_at is masked between runs
--- PASS: TestInterruptedMigrationRecoversAndConverges (18.86s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/BackupPathOccupied/DistinctRunDirectories (2.99s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/BackupPathOccupied/SharedRunDirectory (2.29s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/ManifestRestoredToPreRunBytes/DistinctRunDirectories (1.64s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/ManifestRestoredToPreRunBytes/SharedRunDirectory (1.56s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/EmptiedDirectoriesRecreated/DistinctRunDirectories (1.13s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/EmptiedDirectoriesRecreated/SharedRunDirectory (1.09s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/PlainRerunLeavesRecreatedDirectories (0.76s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantEmptiedManagedDirectoryLeftIsCaught (0.46s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantRemovedManifestEntryResurrectedIsCaught (0.36s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantBackupCopyMissingIsCaught (0.41s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantRemovedFileStillOnDiskIsCaught (0.53s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantTreeDiffersFromTheUninterruptedRunIsCaught (0.26s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantRowsDifferFromTheUninterruptedRunIsCaught (0.23s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantRecoveryRunDidNotReachReconciliationIsCaught (0.22s)
    --- PASS: TestInterruptedMigrationRecoversAndConverges/MutantInterruptedRunRemovedAFileIsCaught (0.22s)
PASS
```

**구성.** 세 중단을 모두 픽스처로 만든다 (운영 코드에 끼어들기 없음): (a) 후보 하나의 실행 백업 사본 자리에 내용 있는 디렉터리를 둔다 — 실행 백업 디렉터리 이름이 시각이라 미리 알 수 없으므로 AC-OWR-20 이 이미 쓰는 `priorManifestWriter` 이음매의 대역이 `runDir` 인자 안에 만든다 (새 이음매 아님); (b) 첫 (완전한) 실행 뒤 매니페스트를 실행 전 바이트로 되돌린다 — 제거한 경로의 항목이 되살아난 상태를 전제로 단언한다; (c) 정리된 빈 디렉터리 셋을 다시 만든다. 각 중단은 두 시계 분기 (구별되는 디렉터리: 폴링 / 공유 디렉터리: 이전 실행의 백업 내용을 합치는 대역) 에서 복구 `--force` 실행을 한 번 한다. 규칙 7 은 중단된 실행과 복구 실행 모두에 적용했다 (센 조건 셋). **기준 실행은 같은 횟수 (두 번) 의 중단 없는 `--force` 실행이다.** 처음 초안은 한 번만 돈 기준과 비교했고 config 섹션 파일 여럿이 갈렸다 (저장한 첫 시도 출력의 줄):

```text
    migrate_ownership_recovery_test.go:925: (tree) the recovered tree differs from the uninterrupted one at [.moai/config/sections/archive.yaml .moai/config/sections/constitution.yaml .moai/config/sections/context.yaml .moai/config/sections/delegation.yaml .moai/config/sections/design.yaml .moai/config/sections/feedback.yaml .moai/config/sections/git-convention.yaml .moai/config/sections/git-strategy.yaml .moai/confi
```

첫 배포와 둘째 병합이 config 섹션을 다르게 쓰기 때문이며 (AC-OWR-17 이 실행 1 을 트리 비교에서 뺀 이유와 같다), 같은 횟수끼리 비교하자 갈림이 사라졌다. 비교하는 해시는 내용만이다 (부채 N17 처분, 시험 로그 첫 줄에도 남는다).

**연속 5회 (같은 트리 `ace2e8d84b`, 종료 코드는 각 실행에서 읽음)**

```text
TestInterruptedMigrationRecoversAndConverges run 1 rc=0 top-pass=1 all-pass=16 fail=0 skip=0 
TestInterruptedMigrationRecoversAndConverges run 2 rc=0 top-pass=1 all-pass=16 fail=0 skip=0 
TestInterruptedMigrationRecoversAndConverges run 3 rc=0 top-pass=1 all-pass=16 fail=0 skip=0 
TestInterruptedMigrationRecoversAndConverges run 4 rc=0 top-pass=1 all-pass=16 fail=0 skip=0 
TestInterruptedMigrationRecoversAndConverges run 5 rc=0 top-pass=1 all-pass=16 fail=0 skip=0 
```

**시험 안의 변이 (순수 단언, `Mutant*IsCaught` 여덟 모두 통과):** 비워진 관리 디렉터리가 남음, 제거한 경로의 매니페스트 항목이 되살아남, 백업 사본 없음, 제거된 파일이 디스크에 남음, 기준 실행과 트리가 다름, 기준 실행과 행이 다름, 복구 실행이 조정에 닿지 않음, 중단된 실행이 파일을 제거함.

**운영 코드에 넣은 임시 변이 셋 (커밋 `ace2e8d84b` 기준, `cmp` 대조 `RESTORED-IDENTICAL`, 출력 줄 추림)**

- **S1** 백업 실패가 제거를 막지 않는다 (`backUpRemovalCandidates` 의 결과를 무시): (a) 가 붉다 — 중단된 실행이 후보를 제거함.

```text
    migrate_ownership_recovery_test.go:905: compared tree hash: content-only (path, bytes, mode, directory presence); directory mtimes and clock values are not compared; deployed_at is masked between runs
    migrate_ownership_recovery_test.go:837: (interrupted) .claude/skills/moai-foundation-core/SKILL.md was removed or changed by the interrupted run (open /private/tmp/own-fixture-d1/TestInterruptedMigrationRecoversAndConvergesBackupPathOcc
    migrate_ownership_recovery_test.go:837: (interrupted) the manifest lost its entry for .claude/skills/moai-foundation-core/SKILL.md
    migrate_ownership_recovery_test.go:837: (interrupted) .claude/skills/moai-ref-owasp-checklist/SKILL.md was removed or changed by the interrupted run (open /private/tmp/own-fixture-d1/TestInterruptedMigrationRecoversAndConvergesBackupPat
    migrate_ownership_recovery_test.go:837: (interrupted) .claude/skills/moai-workflow-testing/SKILL.md was removed or changed by the interrupted run (open /private/tmp/own-fixture-d1/TestInterruptedMigrationRecoversAndConvergesBackupPathOc
    migrate_ownership_recovery_test.go:936: (backup) no run backup holds a byte-identical copy of .claude/skills/moai-foundation-core/SKILL.md
--- FAIL: TestInterruptedMigrationRecoversAndConverges (11.89s)
    --- FAIL: TestInterruptedMigrationRecoversAndConverges/BackupPathOccupied/DistinctRunDirectories (1.42s)
    --- FAIL: TestInterruptedMigrationRecoversAndConverges/BackupPathOccupied/SharedRunDirectory (0.70s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	13.563s
rc=1
```

- **S2** 걷기 뒤 빈 디렉터리 정리 단계 (`pruneManagedEmptyDirs`) 를 뺀다: (c) 가 붉다 — 두 분기 모두.

```text
    migrate_ownership_recovery_test.go:905: compared tree hash: content-only (path, bytes, mode, directory presence); directory mtimes and clock values are not compared; deployed_at is masked between runs
    migrate_ownership_recovery_test.go:936: (tree) the recovered tree differs from the uninterrupted one at [.claude/skills/moai-foundation-core/ .claude/skills/moai-ref-owasp-checklist/ .claude/skills/moai-workflow-testing/], want []
    migrate_ownership_recovery_test.go:936: (emptydirs) empty directories left under the common-asset roots: [.claude/skills/moai-foundation-core .claude/skills/moai-ref-owasp-checklist .claude/skills/moai-workflow-testing], want []
    migrate_ownership_recovery_test.go:936: (rows) kept directory .claude/skills/moai-foundation-core is not listed with reason empty-dir-no-provenance (row "")
--- FAIL: TestInterruptedMigrationRecoversAndConverges (11.38s)
    --- FAIL: TestInterruptedMigrationRecoversAndConverges/EmptiedDirectoriesRecreated/DistinctRunDirectories (1.64s)
    --- FAIL: TestInterruptedMigrationRecoversAndConverges/EmptiedDirectoriesRecreated/SharedRunDirectory (2.03s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	12.300s
rc=1
```

- **S3** 파일이 사라진 매니페스트 항목을 떨어뜨리는 단계를 끈다: (b) 가 붉다 — 되살아난 항목이 남는다.

```text
    migrate_ownership_recovery_test.go:905: compared tree hash: content-only (path, bytes, mode, directory presence); directory mtimes and clock values are not compared; deployed_at is masked between runs
    migrate_ownership_recovery_test.go:936: (manifest) the manifest entry for .claude/skills/moai-foundation-core/SKILL.md came back
    migrate_ownership_recovery_test.go:936: (tree) the recovered tree differs from the uninterrupted one at [], want [.claude/skills/moai-foundation-core/ .claude/skills/moai-ref-owasp-checklist/ .claude/skills/moai-workflow-testing/]
    migrate_ownership_recovery_test.go:936: (emptydirs) empty directories left under the common-asset roots: [], want [.claude/skills/moai-foundation-core .claude/skills/moai-ref-owasp-checklist .claude/skills/moai-workflow-testing]
    migrate_ownership_recovery_test.go:936: (rows) kept directory .claude/skills/moai-foundation-core is not listed with reason empty-dir-no-provenance (row "")
--- FAIL: TestInterruptedMigrationRecoversAndConverges (10.87s)
    --- FAIL: TestInterruptedMigrationRecoversAndConverges/BackupPathOccupied/DistinctRunDirectories (1.11s)
    --- FAIL: TestInterruptedMigrationRecoversAndConverges/BackupPathOccupied/SharedRunDirectory (1.16s)
```

**구현 커밋 C 없음.** 이 기준이 요구하는 수렴은 M1-M5 가 이미 닫았다.

**관측과 정직한 회계**

1. **시계 분기가 결과를 가르는 곳이 하나 있다 — (c) 의 공유 디렉터리 분기.** 구별되는 디렉터리: 복구 실행이 다시 만든 빈 디렉터리 셋을 정리하며 (M5 의 판정 근거는 시작 매니페스트와 보존된 이전 세대뿐인데, 첫 실행이 제거한 경로는 시작 매니페스트에서 이미 빠져 있으므로 남는 근거는 첫 실행의 시작 매니페스트 사본이다), 최종 트리와 행이 중단 없는 기준과 같다. 공유 디렉터리 (같은 초에 시작한 두 갱신): 둘째 실행의 사본이 같은 디렉터리의 첫 실행 사본을 대체하므로 그 근거가 없고, 디렉터리 셋은 그대로 남아 이유 `empty-dir-no-provenance` 로 잔존 목록에 올라온다. 시험은 두 분기를 정확히 단언한다 (공유 분기의 기대: 트리 차이 = 그 디렉터리 셋, 행 = 기준 행 + 그 셋). 파일은 삭제되지 않는다 (증거가 없으면 보존). **제품 쪽 사실로 열어 둔다: 같은 시계 초에 시작한 두 갱신은 더 오래된 이전 세대를 잃는다 — 증거 손실일 뿐이고 이 카드 밖이다 (실행 백업 디렉터리 이름의 초 단위 해상도는 `defs.BackupTimestampFormat`).** (a), (b) 는 두 분기 모두 수렴한다.
2. 일반 재실행 사례 (`PlainRerunLeavesRecreatedDirectories`): (c) 의 디렉터리를 다시 만든 뒤 `--force` 없이 돌린 같은 버전 갱신은 건너뛰고 (`Skipped == true`), 트리가 바뀌지 않으며 디렉터리 셋이 그대로 남는다 — plan.md OQ-6 의 운영 사실이 시험으로 확인됐다. 일반 재실행은 복구 경로가 아니다.
3. (a) 에서 실패한 실행은 갱신의 나머지를 계속한다: 그 실행의 배포는 진행되고 (매니페스트 저장, 템플릿 배포), 막히는 것은 제거뿐이다 (design.md D-3 항목 4). 시험은 후보 보존, 매니페스트 항목 유지, 사본 부재, 행 이유를 단언한다. 복구 실행 뒤 최종 트리가 기준과 같다는 것이 「배포가 이어졌다」 의 간접 증거이다.
4. 부채 N18 (AC-OWR-20 (a)(b) 가 규칙 7 조건 (ii) 를 요구받음) 처분은 E.3 에서 한다.

#### M10 — AC-OWR-19 `TestRemovalSwitchStaysOffInProduction`

**커밋 A `dd5a4b42b6`** (`internal/cli/migrate_ownership_gate_test.go` 한 파일, 운영 코드 변경 없음). 작업 트리 깨끗, 부하 평균 32.17 30.74 32.25.

**A 직전 트리 `886f5b2d82` 에서의 셀 (RED 도 통과도 아님).** 시험 파일을 치워 둔 채 같은 선택자를 돌렸다 (파일을 다시 가져다 놓고 `cmp` 로 `RESTORED-IDENTICAL` 확인). 제로 매치 규칙 (acceptance.md 규약 4) 대로 통과로 읽지 않는다 — 이 시점에는 소스 스캔 시험이 없었다는 사실만 기록한다:

```text
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.739s [no tests to run]
rc=0
```

**도착 셀 (트리 `dd5a4b42b6`).** 판정 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestRemovalSwitchStaysOffInProduction|TestMigrationRemovalGate|TestMigrationDefaultKeepsConfirmedProjectCopy)$' -count=1 -v`, 종료 코드 0. 새 시험은 비시험 코드에서 스위치를 바꿀 수 있는 문장이 없음을 단언하고, 기존 두 게이트 시험은 그대로 초록이다. 이 기준은 acceptance.md 가 정한 대로 regression-guard 이고 실패쪽은 운영 코드 변이로 관측한다:

```text
    migrate_ownership_gate_test.go:205: scanned 403 non-test files: 1 declaration, 3 mentions, 1 keyed hand-overs, 0 violations
--- PASS: TestRemovalSwitchStaysOffInProduction (0.13s)
    --- PASS: TestRemovalSwitchStaysOffInProduction/DefaultIsOff (0.00s)
    --- PASS: TestRemovalSwitchStaysOffInProduction/NoProductionStatementChangesTheSwitch (0.13s)
    --- PASS: TestRemovalSwitchStaysOffInProduction/ScanMutants (0.00s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (1.22s)
--- PASS: TestMigrationRemovalGate (0.56s)
PASS
```

**연속 5회 (같은 트리, 종료 코드는 각 실행에서 읽음; 세 시험을 한 호출로)**

```text
TestRemovalSwitchStaysOffInProduction+gate pair run 1 rc=0 top-pass=3 all-pass=18 fail=0 skip=0 
TestRemovalSwitchStaysOffInProduction+gate pair run 2 rc=0 top-pass=3 all-pass=18 fail=0 skip=0 
TestRemovalSwitchStaysOffInProduction+gate pair run 3 rc=0 top-pass=3 all-pass=18 fail=0 skip=0 
TestRemovalSwitchStaysOffInProduction+gate pair run 4 rc=0 top-pass=3 all-pass=18 fail=0 skip=0 
TestRemovalSwitchStaysOffInProduction+gate pair run 5 rc=0 top-pass=3 all-pass=18 fail=0 skip=0 
```

**스캔의 도달 단언.** 시험은 스캔이 패키지에 닿았음을 먼저 단언한다: 읽은 비시험 파일 수가 100 이상 (위 로그 403), 패키지 수준 `var removalEnabled = false` 가 정확히 하나, 변수를 쓰는 곳 (언급) 과 `ownershipEnv` 로 넘기는 키 필드 (migrate_project_assets.go 의 `removalEnabled: removalEnabled`) 가 있음. 시작 시점 값이 거짓임도 단언한다 (`DefaultIsOff`). 스캔이 보는 문장: 대입 (`=`, `:=`, 연산 대입), 범위 절의 대입, 증감, 주소 취하기, 필드 `removalEnabled` 에 대한 대입, 합성 리터럴의 필드 `removalEnabled` 가 패키지 변수의 읽기가 아닌 값을 받는 경우, 유일한 `= false` 선언 말고 다른 선언 (지역 가림 포함). **보지 못하는 것:** 리플렉션이나 `unsafe` 를 통한 쓰기, `go:linkname`, 이 패키지 밖의 프로세스가 스위치를 켜는 일. 설정 파일과 환경 변수는 위 문장 없이는 변수에 닿을 수 없다.

**시험 안의 변이 (스캔이 틀린 소스 앞에서 붉음, `ScanMutants` 열두 개 모두 통과):** init 에서 true 대입, 설정값(비상수) 대입, 선언 기본값 true, 선언 값이 함수 호출, 지역 `:=` 가림, 주소 취하기, 환경 필드 대입, 환경 필드 리터럴 true, 환경 필드 리터럴이 설정값, 연산 대입, 범위 대입, 패키지 수준 선언 둘. 출하 모양 (`var removalEnabled = false`, 읽기, 키 필드로 변수 넘김) 은 조용하다.

**운영 코드에 넣은 임시 변이 셋 (커밋 `dd5a4b42b6` 기준, `cmp` 대조 `RESTORED-IDENTICAL`, 세 시험을 한 호출로 돌린 출력의 줄 추림)**

- **T1** 기본값을 `var removalEnabled = true` 로 뒤집는다 — 새 시험과 기존 두 게이트 시험이 모두 붉다 (기준 본문이 요구한 실패쪽 관측):

```text
    migrate_ownership_gate_test.go:183: removalEnabled is true at the start of the test: the shipped default must be false
    migrate_ownership_gate_test.go:197: found 0 package-level `var removalEnabled = false` declarations, want exactly 1
    migrate_ownership_gate_test.go:203: production code can change the removal switch: migrate_project_assets.go:40:5: a declaration of removalEnabled other than the package-level `= false`
    migrate_ownership_gate_test.go:205: scanned 403 non-test files: 0 declaration, 3 mentions, 1 keyed hand-overs, 1 violations
--- FAIL: TestRemovalSwitchStaysOffInProduction (0.15s)
    --- FAIL: TestRemovalSwitchStaysOffInProduction/DefaultIsOff (0.00s)
    --- FAIL: TestRemovalSwitchStaysOffInProduction/NoProductionStatementChangesTheSwitch (0.15s)
    migrate_removal_switch_test.go:23: .claude/skills/moai-foundation-core/SKILL.md missing: stat /private/tmp/own-fixture-d1/TestMigrationDefaultKeepsConfirmedProjectCopy2661427888/001/.claude/skills/moai-foundation-core/SKILL.md
--- FAIL: TestMigrationDefaultKeepsConfirmedProjectCopy (0.75s)
    migrate_removal_switch_test.go:60: default-configuration migration removed [.claude/skills/moai-foundation-core/SKILL.md]; the gate must keep every removal-preconditioned copy
    migrate_removal_switch_test.go:62: .claude/skills/moai-foundation-core/SKILL.md missing: stat /private/tmp/own-fixture-d1/TestMigrationRemovalGate806506081/001/.claude/skills/moai-foundation-core/SKILL.md: no such file or dire
```

- **T2** `func init() { removalEnabled = true }` 를 더한다 (기본 선언은 `false` 그대로 — 선언만 보는 스캔이 놓칠 변이):

```text
    migrate_ownership_gate_test.go:183: removalEnabled is true at the start of the test: the shipped default must be false
    migrate_ownership_gate_test.go:203: production code can change the removal switch: migrate_project_assets.go:42:15: an assignment to removalEnabled
    migrate_ownership_gate_test.go:205: scanned 403 non-test files: 1 declaration, 4 mentions, 1 keyed hand-overs, 1 violations
--- FAIL: TestRemovalSwitchStaysOffInProduction (0.11s)
    --- FAIL: TestRemovalSwitchStaysOffInProduction/DefaultIsOff (0.00s)
    --- FAIL: TestRemovalSwitchStaysOffInProduction/NoProductionStatementChangesTheSwitch (0.11s)
    migrate_removal_switch_test.go:23: .claude/skills/moai-foundation-core/SKILL.md missing: stat /private/tmp/own-fixture-d1/TestMigrationDefaultKeepsConfirmedProjectCopy634228031/001/.claude/skills/moai-foundation-core/SKILL.md:
--- FAIL: TestMigrationDefaultKeepsConfirmedProjectCopy (0.54s)
    migrate_removal_switch_test.go:60: default-configuration migration removed [.claude/skills/moai-foundation-core/SKILL.md]; the gate must keep every removal-preconditioned copy
```

- **T3** `newOwnershipEnv` 가 `removalEnabled: true` 를 넘긴다 (변수는 그대로 `false`, 환경만 켜짐):

```text
    migrate_ownership_gate_test.go:200: the scan found no mention (2) or keyed hand-over (0) of the switch: it is not looking at the code that uses it
    migrate_ownership_gate_test.go:203: production code can change the removal switch: migrate_project_assets.go:196:3: a composite-literal field removalEnabled that is not a read of the package variable
    migrate_ownership_gate_test.go:205: scanned 403 non-test files: 1 declaration, 2 mentions, 0 keyed hand-overs, 1 violations
--- FAIL: TestRemovalSwitchStaysOffInProduction (0.39s)
    --- FAIL: TestRemovalSwitchStaysOffInProduction/NoProductionStatementChangesTheSwitch (0.39s)
    migrate_removal_switch_test.go:23: .claude/skills/moai-foundation-core/SKILL.md missing: stat /private/tmp/own-fixture-d1/TestMigrationDefaultKeepsConfirmedProjectCopy3315918169/001/.claude/skills/moai-foundation-core/SKILL.md
--- FAIL: TestMigrationDefaultKeepsConfirmedProjectCopy (1.84s)
    migrate_removal_switch_test.go:60: default-configuration migration removed [.claude/skills/moai-foundation-core/SKILL.md]; the gate must keep every removal-preconditioned copy
```

T3 에서는 기존 게이트 시험도 붉다 (환경이 켜져 제거가 일어남). T2 에서 새 시험의 `DefaultIsOff` 와 스캔이 모두 붉고, T1 에서는 선언 스캔과 시작 값 단언이 둘 다 붉다.

**구현 커밋 C 없음.** 구현이 필요한 차이가 없다 (스위치는 처음부터 꺼져 있다).

#### Run chunk 4 GREEN 기록과 정직한 회계 (M8 + M9 + M10)

**커밋 지도 (기준 HEAD `cf2581728e`)**

| 묶음 | A (시험만) | B (셀 기록) | C (구현) | 비고 |
|------|-----------|-----------|----------|------|
| M8 AC-OWR-15 | `415644baff` | `b896029228` | — | 도착 시 GREEN 가드, 붉음은 변이 (시험 안 열여섯 + 운영 코드 넷) |
| M9 AC-OWR-17 | `5614767764` | `7414777ee3` | — | 도착 시 GREEN 가드, 변이 열둘 + 운영 코드 넷 |
| M9 AC-OWR-18 | `ace2e8d84b` | `886f5b2d82` | — | 도착 시 GREEN 가드, 변이 여덟 + 운영 코드 셋 |
| M10 AC-OWR-19 | `dd5a4b42b6` | `5cebd5a3da` | — | 도착 시 GREEN 가드, 스캔 변이 열둘 + 운영 코드 셋 |
| 린트 후속 (시험만) | `8afbbe9436` | — | — | AC-OWR-15 시험의 쓰지 않는 필드 둘 제거, 동작 변화 없음 |

**구현 커밋이 하나도 없다.** M8-M10 의 기준 넷 모두 M1-M5 가 이미 닫은 동작을 도착 시 초록으로 확인하고, 붉음을 변이로 관측했다. 「M1-M5 가 남긴 차이만 고친다」 에 해당하는 남은 차이는 측정되지 않았다. 측정된 어긋남은 둘이고 모두 SPEC 본문의 모순이라 코드가 아니라 공개로 처리했다: REQ-OWR-012 의 「동일한 잔존 목록」 대 design.md D-2 의 이유 정제 (AC-OWR-17 기록의 관측 1), AC-OWR-15/-17 의 제외 목록과 기대 차이 집합이 시계 파생 위치 둘을 빠뜨림.

**회귀: 이 카드의 시험 전부 (앵커된 이름 66개, 트리 `8afbbe9436`, 작업 트리 깨끗)** — 기준선 11 + 이 카드가 만든 시험 + 가드 (`TestUpdateDryRun_ZeroMutation`, `TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy`, `TestShouldSkipBinaryUpdate_*` 여섯). 이름은 파일 `card_tests` 에서 나왔고 모두 패키지에 정확히 한 번 있음을 확인했다 (이름 점검 66/66). 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(<아래 66개>)$' -count=1 -v -coverprofile=<파일>` (패키지 전체 실행이 아니다), 종료 코드 0, 최상위 `--- PASS`=66 `--- FAIL`=0 `--- SKIP`=0, 부하 평균 시작 `26.37 29.51 31.73` 끝 `30.33 30.96 32.02`:

```text
PASS
coverage: 10.0% of statements
```

앵커된 이름:

```text
TestAliasToAKeyTheManifestLacksIsNoAlias
TestBackupFailureRemovesNothingThroughProductionPath
TestCaseVariantCounterpartAndMirrorSpelling
TestCaseVariantOnHostFilesystem
TestCaseVariantResolvesToRegisteredEntry
TestCommandsAndAgentsDiscoverableAfterInterruptedRun
TestCommandsSingleDispositionAndPresent
TestCounterpartLocationAndStatusEdges
TestDowngradeDropsEvidenceNotSafety
TestDryRunChangesNothingAndMatchesVerdict
TestDryRunRunsNoBinaryUpdateStep
TestEmptiedSkillDirRemoved
TestEmptyDirProvenanceVerdict
TestInterruptedMigrationRecoversAndConverges
TestKeepReasonVocabularyIsClosed
TestLegacyManifestRoundTrip
TestManagedOnlyEmptyDirsPruned
TestMigrationBackupBeforeRemoval
TestMigrationClassifiesUnregisteredMirrorCopy
TestMigrationDefaultKeepsConfirmedProjectCopy
TestMigrationIdempotent
TestMigrationKeepsSharedCopyUnderProfileReach
TestMigrationPreservesExistingMirrorEntries
TestMigrationRemovalGate
TestMigrationRemovesSharedCopyWithoutProfileReach
TestNoProfileRootAccessInUpdateSources
TestOwnershipAttestedManagedOnlyWithFullEvidence
TestOwnershipAttestedManagedThroughOperatingSource
TestOwnershipHelperFixtureBuilders
TestOwnershipHelperForcedRunReachesReconciliation
TestOwnershipHelperTreeSnapshotSeesDirMtime
TestOwnershipHelperUsesRealEmbeddedPaths
TestOwnershipPreserveByteEqualUserFile
TestOwnershipPreserveUserModifiedFile
TestOwnershipUnregisteredUnknownOriginKeptWithReason
TestOwnershipVerdictClasses
TestPreservePriorManifestOutcomes
TestPreviewOmitsRemovalCandidateWhileSwitchOff
TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy
TestPreviewVerdictsFailClosed
TestPriorGenerationCorruptIsNoEvidence
TestPriorGenerationSourceBoundedAndTolerant
TestPriorGenerationsAreLookupBoundOnly
TestPriorManifestAtomicWriteFailures
TestPriorManifestRetainedInRunBackup
TestProfileCallsiteBaselineMatchesTree
TestProfileEnvDoesNotReachProfileRoot
TestProfileMappedCopyKeptWithoutCounterpart
TestProfileReachMarkUsesTheRegisteredKeyUnderACaseAlias
TestProfileScanCatchesMutants
TestPruneNestedEmptyDirGuards
TestRF2F3b_MigrationPreservesUntracked
TestRemainingListKeeps313UnregisteredCopies
TestRemainingRenderEdges
TestRemovalSwitchStaysOffInProduction
TestRemovedAssetBackupRestoreByteIdentical
TestRepeatedUpdateIsIdempotent
TestShouldSkipBinaryUpdate_DevBuild
TestShouldSkipBinaryUpdate_DevBuild_Phase5_V2
TestShouldSkipBinaryUpdate_EnvVar
TestShouldSkipBinaryUpdate_EnvVar_Phase5
TestShouldSkipBinaryUpdate_Normal
TestShouldSkipBinaryUpdate_TemplatesOnlyFlag
TestUpdateDryRun_ZeroMutation
TestVerdictAndPreviewCodeHasNoWritableCalls
TestViewOfManifestIsAReadOnlyCopy
```

`internal/manifest`: 카드 시험 `TestManifestRecordsDeployedVersion` 만 (`-coverprofile`), 종료 코드 0:

```text
--- PASS: TestManifestRecordsDeployedVersion (0.01s)
coverage: 49.5% of statements
```

패키지 전체 (작은 패키지라 한 번 돌렸다; 이 줄은 카드 범위 수치가 아니라 패키지 전체 수치이다), 종료 코드 0:

```text

```

**범위가 좁은 구문 커버리지** — 위 회귀 한 번의 프로필 (internal/cli 는 앵커된 66개만이므로 그 파일을 덮는 다른 시험은 세지 않는다). 「바뀐 줄」 수치는 카드 diff (`git diff -U0 cdd72a814..HEAD`) 와 겹치는 블록의 구문 비율이며 그 블록 안의 바뀌지 않은 구문도 포함한다:

```text
internal/cli/migrate_ownership.go (created) 96.0% (119/124 statements)
internal/cli/migrate_ownership_case.go (created) 96.6% (57/59 statements)
internal/cli/migrate_ownership_emptydir.go (created) 93.3% (84/90 statements)
internal/cli/migrate_ownership_env.go (created) 94.4% (34/36 statements)
internal/cli/migrate_ownership_evidence.go (created) 100.0% (25/25 statements)
internal/cli/migrate_ownership_remaining.go (created) 100.0% (67/67 statements)
internal/cli/update_prior_manifest.go (created) 87.5% (70/80 statements)
internal/cli/migrate_project_assets.go (edited) whole-file 82.8% (216/261); card-changed lines 87.9% (94/107 statements in blocks touching the card diff)
internal/cli/update.go (edited) whole-file 35.7% (146/409); card-changed lines 100.0% (5/5 statements in blocks touching the card diff)
internal/cli/update_dryrun_preview.go (edited) whole-file 86.1% (31/36); card-changed lines 100.0% (8/8 statements in blocks touching the card diff)
internal/cli/update_template_sync.go (edited) whole-file 64.9% (276/425); card-changed lines 90.3% (28/31 statements in blocks touching the card diff)
internal/manifest/manifest.go (edited) [card test only] whole-file 50.0% (45/90); card-changed lines 100.0% (7/7)
internal/manifest/manifest.go (edited) [whole manifest package] whole-file 91.1% (82/90); card-changed lines 100.0% (7/7)
internal/manifest/types.go (edited) [card test only] whole-file 25.0% (1/4); card-changed lines 0.0% (0/0)
internal/manifest/types.go (edited) [whole manifest package] whole-file 100.0% (4/4); card-changed lines 0.0% (0/0)
```

85% 미만을 정직하게 표시한다: **`internal/cli/migrate_project_assets.go` 전체 82.8%, `internal/cli/update.go` 전체 35.7%, `internal/cli/update_template_sync.go` 전체 64.9%, `internal/manifest/manifest.go` 카드 시험만 50.0%** — 넷 모두 이 카드 이전의 코드가 대부분이고 (t1668 및 그 이전) 그 부분을 덮는 시험은 앵커된 66개 밖에 있다. 카드가 바꾼 줄은 각각 87.9%, 100%, 90.3%, 100% (패키지 전체 시험 기준 manifest.go 전체 91.1%). 카드가 만든 파일 일곱은 모두 85% 이상이다. 함수 단위로 85% 미만인 것은 둘: `update_prior_manifest.go` 의 `writePriorManifestTemp` 50.0% (임시 파일 만들기·쓰기·동기화·닫기 실패 가지 — 시험은 이 함수를 관측점 변수 `priorManifestTempWrite` 로 대체해 실패를 주입하므로 진짜 함수의 실패 가지는 닿지 않는다) 와 `readStartManifest` 75.0% (읽기 실패 시 nil). 카드가 만든 파일의 함수 단위 수치:

```text
internal/cli/migrate_ownership.go:45: valid 100.0%
internal/cli/migrate_ownership.go:80: evidenceIncompleteReason 100.0%
internal/cli/migrate_ownership.go:85: isKnownKeepReason 100.0%
internal/cli/migrate_ownership.go:177: counterpartOf 100.0%
internal/cli/migrate_ownership.go:218: manifestKey 100.0%
internal/cli/migrate_ownership.go:227: row 100.0%
internal/cli/migrate_ownership.go:243: rowKeptFor 100.0%
internal/cli/migrate_ownership.go:258: classifyOwnership 100.0%
internal/cli/migrate_ownership.go:272: classifyOwnershipCore 85.7%
internal/cli/migrate_ownership.go:322: classifyUnregistered 100.0%
internal/cli/migrate_ownership.go:345: disposition 100.0%
internal/cli/migrate_ownership.go:375: reasonTag 100.0%
internal/cli/migrate_ownership.go:397: userCounterpartLocation 100.0%
internal/cli/migrate_ownership.go:417: userCounterpartStatus 90.9%
internal/cli/migrate_ownership_case.go:44: resolveCaseAlias 100.0%
internal/cli/migrate_ownership_case.go:73: manifestKeysByDir 100.0%
internal/cli/migrate_ownership_case.go:85: shippedCaseSpelling 94.1%
internal/cli/migrate_ownership_case.go:116: canonicalCounterpartRel 93.3%
internal/cli/migrate_ownership_case.go:143: counterpartConfirmedCaseAware 100.0%
internal/cli/migrate_ownership_emptydir.go:63: dirManagedProvenance 100.0%
internal/cli/migrate_ownership_emptydir.go:114: startOfRunRecords 85.7%
internal/cli/migrate_ownership_emptydir.go:129: emptyDirTree 100.0%
internal/cli/migrate_ownership_emptydir.go:151: pruneNestedEmptyDir 88.9%
internal/cli/migrate_ownership_emptydir.go:171: pruneManagedEmptyDirs 91.7%
internal/cli/migrate_ownership_emptydir.go:195: pruneSkillDirTree 88.9%
internal/cli/migrate_ownership_env.go:32: viewOfManager 100.0%
internal/cli/migrate_ownership_env.go:38: viewOfManifest 100.0%
internal/cli/migrate_ownership_env.go:74: removalEligible 100.0%
internal/cli/migrate_ownership_env.go:87: buildPreviewEnv 87.5%
internal/cli/migrate_ownership_evidence.go:103: provenanceEvidence 100.0%
internal/cli/migrate_ownership_remaining.go:32: sortRemainingRows 100.0%
internal/cli/migrate_ownership_remaining.go:41: isRemainingRowLine 100.0%
internal/cli/migrate_ownership_remaining.go:51: remainingAwareReport 100.0%
internal/cli/migrate_ownership_remaining.go:64: remainingSummaryLines 100.0%
internal/cli/migrate_ownership_remaining.go:112: remainingDetailLines 100.0%
internal/cli/migrate_ownership_remaining.go:135: missingList 100.0%
internal/cli/migrate_ownership_remaining.go:146: recordRemainingList 100.0%
internal/cli/update_prior_manifest.go:75: writePriorManifestTemp 50.0%
internal/cli/update_prior_manifest.go:103: writePriorManifest 100.0%
internal/cli/update_prior_manifest.go:118: readStartManifest 75.0%
internal/cli/update_prior_manifest.go:131: preservePriorManifest 100.0%
internal/cli/update_prior_manifest.go:148: recordManifestDeployedVersion 100.0%
internal/cli/update_prior_manifest.go:172: newPriorGenerationSource 100.0%
internal/cli/update_prior_manifest.go:176: generations 100.0%
internal/cli/update_prior_manifest.go:199: readPriorGenerations 97.1%
```

**품질 (트리 `8afbbe9436` 의 Go 소스 = `8afbbe9436` 이후; 각 명령의 종료 코드를 직접 읽음)**

```text
go build ./... rc=0
GOOS=windows go build ./... rc=0
gofmt -l internal/cli internal/manifest rc=0 output-bytes=0
go vet ./internal/cli/ ./internal/manifest/ rc=0 output-bytes=0
GOOS=windows go vet ./internal/cli/ ./internal/manifest/ rc=0 output-bytes=0
```

`golangci-lint run ./internal/manifest/ ./internal/cli/` 첫 실행 (트리 `5cebd5a3da`), 종료 코드 1: 다섯 건 — 세 건은 이 카드가 만지지 않은 `internal/cli/factory_run_adopt.go` 의 기존 errcheck, **두 건은 내 시험 (`migrate_ownership_restore_test.go`) 의 쓰지 않는 필드** (`unused`). 두 건을 고쳐 별도 커밋 `8afbbe9436` 으로 냈다 (시험 다시 통과). 둘째 실행 (트리 `8afbbe9436`), 종료 코드 1:

```text
internal/cli/factory_run_adopt.go:79:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:91:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:103:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/migrate_ownership_restore_test.go:293:2: field postRun1Snapshot is unused (unused)
internal/cli/migrate_ownership_restore_test.go:294:2: field postRun1AssetFiles is unused (unused)
5 issues:
* errcheck: 3
* unused: 2
```

```text
internal/cli/factory_run_adopt.go:79:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:91:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:103:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
3 issues:
* errcheck: 3
```

카드 파일에 대한 린트 지적은 0건이다. (`factory_run_adopt.go` 는 `git diff --name-only cdd72a814..HEAD` 에 없다 — 아래 제약 점검.)

**하드 제약 점검 (최종 트리, 종료 코드는 명령마다 직접 읽음; 양성 대조군 포함)**

```text
--- protected files / profile (card commits)
exit=0
--- positive control (.moai/specs commits)
      27
--- diff names touching protected/profile/audit-receipts/backup policy/templates
grep-exit=1
--- removalEnabled = true in non-test code
grep-exit=1
40:var removalEnabled = false
--- AskUserQuestion in card production files
grep-exit=1
--- factory_run_adopt.go touched by the card?
grep-exit=1
--- spec bodies touched in chunk 4 (since cf2581728e)
grep-exit=1
```

**M8-M10 의 정직한 회계 / 갭 / 잔여 위험**

1. **네 기준의 RED-now 가 없다.** AC-OWR-15, -17, -18, -19 는 도착 시 GREEN 이다 (위 각 도착 셀). acceptance.md 가 적은 RED 이유는 M1b/M3/M5 가 이미 닫았다. 이 기준들은 regression-guard 이고 release-blocking 으로 승격되지 않는다. 붉음의 원문은 운영 코드 변이 열네 개 (M8 넷, AC-17 넷, AC-18 셋, AC-19 셋) 와 시험 안 변이 마흔여덟 개 (AC-15 열여섯, AC-17 열둘, AC-18 여덟, AC-19 열둘) 의 셀이다.
2. **시계 독립.** 이 chunk 의 어떤 단언도 두 실행이 같은 초에 떨어지는지에 달려 있지 않다. 두 분기를 모두 몬다: 구별되는 디렉터리 (폴링으로 결정적) 와 공유 디렉터리 (이전 실행의 백업 내용을 합치는 쓰기 대역, `recSameDir` 로 같은 디렉터리 판정). 첫 초안의 대역은 「같은 디렉터리」 판정을 문자열 비교로 해서 (상대 경로와 절대 경로) 사용 중인 실행 디렉터리를 지웠다 — 제품 결함이 아니라 시험 구성 오류였고 커밋 A 이전에 고쳤다.
3. **시험 구성의 의존.** 세 시험이 쓰는 이음매는 기존 것뿐이다: `priorManifestWriter` (AC-OWR-20 이 도입), `enableRemovalForTest`, 헬퍼 `runUpdateFlow*`. 운영 코드에 새 이음매나 끼어들기를 더하지 않았다.
4. **시험이 보지 못하는 것.** AC-OWR-15: `.moai/logs/**`·`.moai/state/**` 안의 내용 차이, 디렉터리 mtime/xattr/소유자, 프로젝트 밖 쓰기. AC-OWR-17/-18: 같은 항목들, 그리고 실제 같은 초의 두 실행 (공유 분기는 최종 상태를 재현할 뿐 경합 자체는 아니다). AC-OWR-19: 리플렉션/`unsafe`/`go:linkname`.
5. **실행하지 않은 것 (갭).** `internal/cli` 패키지 전체 시험과 패키지 전체 커버리지, `-race`, Windows 에서의 시험 실행 (빌드와 vet 만), 다른 OS 에서의 대소문자 구분 파일시스템 실행 — 이 chunk 의 금지 범위이거나 CI/리더의 착지 재측정 몫이다.
6. **착지 재측정이 볼 위험.** 이 카드의 시험은 `internal/cli` 에서 무겁다 (회귀 한 번 148초, 부하 평균 약 30). 패키지 전체 실행과 병렬 실행에서 시간 초과나 간헐 실패가 나타날 수 있다. 앞 chunk 에서 한 번 있었던 시계 초 의존 실패 (AC-OWR-14) 는 고쳤고, 이 chunk 의 시험은 처음부터 시계 독립으로 썼다. 이 chunk 의 새 시험 넷 모두 연속 5회 초록이다 (각 도착 셀).

#### Chunk 5 (card-review fixes)

코덱스 카드 리뷰(조언용, HEAD `7b758103d8`)의 P2 둘을 lane 이 코드로 확인했다. 둘 다 실제 결함이다. 규약은 앞 chunk 와 같다: 시험 커밋 A, 셀 기록 B, 수정 C. 셀은 저장한 출력 파일에서 스크립트로 옮겼다.

**커밋 A `73e37d3832`** (시험만: `internal/cli/update/backup/restore_prior_manifest_test.go`, `internal/cli/update_prior_manifest_restore_test.go`, `internal/cli/migrate_ownership_restore_test.go`). 부하 평균 13.03 20.91 26.68.

**발견 1 — `manifest.prior.json` 이 복원 때 `.moai/config` 로 복사된다 (이 카드가 만든 결함).** Backup 단계는 이 파일을 실행 백업 디렉터리 루트에 둔다. `RestoreMoaiConfigLegacy` 는 백업 루트 전체를 걸으며 `backup_metadata.json`, `.template-defaults`, 실행 표식, 비관리 설정 디렉터리만 건너뛰므로, 설정에 `sections/` 가 없던 프로젝트(백업에도 없음)에서는 이 파일이 `.moai/config/manifest.prior.json` 이 된다. `RestoreMoaiConfigRetained` 는 `sections/` 가 있으면 그 아래 YAML 만 걷고, 없으면 Legacy 로 넘어간다. 시험은 파일 이름 `manifest.prior.json` 을 리터럴로 적는다 (디스크에 실제로 쓰이는 이름을 고정; 수정이 쓰는 상수와 독립). 같은 이름이 하위 디렉터리에 있으면 사용자 내용이라 계속 복원해야 한다는 단언도 넣었다.

RED 셀 (트리 `73e37d3832`, 작업 트리 깨끗). 명령 1 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/update/backup/ -run 'PriorManifest' -count=1 -v`, 종료 코드 1, 시험 로그 줄과 결과 줄만 (원문 순서):

```text
    restore_prior_manifest_test.go:73: manifest.prior.json from the backup root was copied into the config tree (stat err: <nil>)
--- FAIL: TestRestoreLegacySkipsPriorManifest (0.00s)
    restore_prior_manifest_test.go:83: manifest.prior.json from the backup root was copied into the config tree (stat err: <nil>)
--- FAIL: TestRestoreRetainedWithoutSectionsSkipsPriorManifest (0.00s)
--- PASS: TestRestoreRetainedWithSectionsSkipsPriorManifest (0.00s)
    restore_prior_manifest_test.go:107: manifest.prior.json from the backup root was copied into the config tree (stat err: <nil>)
--- FAIL: TestRestoreFromBackupDirSkipsPriorManifest (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.165s
FAIL
```

(`TestRestoreRetainedWithSectionsSkipsPriorManifest` 는 초록이다: `sections/` 가 있는 백업은 구멍이 없다 — 특성화로 남긴다.) 명령 2 (업데이트 전체 흐름, 설정에 `sections/` 없는 프로젝트) `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestUpdateDoesNotRestorePriorManifestIntoConfig$' -count=1 -v`, 종료 코드 1. 도달 대조군(새 사본이 생겼고 설정 파일이 복원됨)은 통과한 뒤의 단언이 붉다:

```text
    update_prior_manifest_restore_test.go:36: manifest.prior.json was copied into .moai/config by the update (stat err: <nil>)
--- FAIL: TestUpdateDoesNotRestorePriorManifestIntoConfig (0.36s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.081s
FAIL
```

**발견 2 — 제거 후보의 백업 사본이 실행 비트를 잃는다 (잠복: 제거 스위치는 운영에서 꺼져 있다).** `backupRemovedAsset` 가 모든 사본을 `defs.FilePerm` (0644) 으로 쓴다. AC-OWR-15 의 모드 절 `(i-mode)`, `(iii-mode)` 은 있었지만 픽스처에 실행 파일이 없어 볼 수 없었다. 픽스처에 실제 내장 템플릿의 실행 스크립트 `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh` 를 등록된 managed 후보로 더하고 0755 로 만들었다 (사전 단언: 실행 전 모드가 755). 모드 절 둘이 보고되지 않던 구멍도 고쳤다: `(i-mode)`, `(iii-mode)` 이 이전에는 어느 서브 시험 묶음에도 속하지 않아 `(iv)` 로만 드러났다 — 이제 각각 `BackupCopyModes`, `RestoredModes` 로 보고된다. 변이 `BackupCopyLosesTheExecuteBit` 를 더했다.

RED 셀 (같은 트리), 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemovedAssetBackupRestoreByteIdentical$' -count=1 -v`, 종료 코드 1, 모드 줄과 실패 줄만 (같은 문제 줄은 한 번):

```text
    migrate_ownership_restore_test.go:629: (i-mode) the backup copy of .claude/skills/moai-workflow-project/scripts/navigator-audit.sh has mode 644, the file had 755
    migrate_ownership_restore_test.go:629: (iii-mode) the restored .claude/skills/moai-workflow-project/scripts/navigator-audit.sh has mode 644, it had 755
--- FAIL: TestRemovedAssetBackupRestoreByteIdentical (6.91s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/BackupCopyModes (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/RestoredModes (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/DifferenceSentinel (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/SharedClockNamesJudgeTheSame (0.18s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/ExpectedSetMembersThatDiffer (0.08s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	7.655s
```

**커밋 C `051ddc18ef`** (운영 코드: `internal/cli/update/backup/run_marker.go` 에 내보내는 상수 `PriorManifestName`, `internal/cli/update/backup/restore.go` 의 Legacy 걷기가 루트의 그 파일을 건너뜀, `internal/cli/update_prior_manifest.go` 가 쓰는 쪽 이름을 그 상수로 정의, `internal/cli/migrate_project_assets.go` 의 제거 후보가 파일 권한 비트를 실어 나르고 `backupRemovedAsset` 가 그 모드로 쓴 뒤 `chmod`). `RestoreMoaiConfigRetained` 는 고치지 않았다 — `sections/` 아래 YAML 만 걷고 `sections/` 가 없으면 Legacy 로 넘어가므로 같은 구멍이 아니다 (시험 `TestRestoreRetainedWithSectionsSkipsPriorManifest` 가 특성화로 지킨다). 상수는 Legacy 걷기에서 루트 상대 경로가 정확히 같을 때만 건너뛴다 (하위 디렉터리의 같은 이름은 사용자 내용이므로 계속 복원, 시험이 단언).

**GREEN (트리 `051ddc18ef`, 종료 코드는 각 실행에서 직접 읽음; 연속 5회).** 새 선택자: 복원 수준 `go test ./internal/cli/update/backup/ -run 'PriorManifest'`, 흐름 수준 `-run '^TestUpdateDoesNotRestorePriorManifestIntoConfig$'`, AC-OWR-15 `-run '^TestRemovedAssetBackupRestoreByteIdentical$'` (모두 `TMPDIR=/private/tmp/own-fixture-d1 ... -count=1 -v`):

```text
run 1: backup-level rc=0 pass=4 fail=0 | flow-level rc=0 pass=1 fail=0 | AC-OWR-15 rc=0 top-pass=1 all-pass=28 fail=0
run 2: backup-level rc=0 pass=4 fail=0 | flow-level rc=0 pass=1 fail=0 | AC-OWR-15 rc=0 top-pass=1 all-pass=28 fail=0
run 3: backup-level rc=0 pass=4 fail=0 | flow-level rc=0 pass=1 fail=0 | AC-OWR-15 rc=0 top-pass=1 all-pass=28 fail=0
run 4: backup-level rc=0 pass=4 fail=0 | flow-level rc=0 pass=1 fail=0 | AC-OWR-15 rc=0 top-pass=1 all-pass=28 fail=0
run 5: backup-level rc=0 pass=4 fail=0 | flow-level rc=0 pass=1 fail=0 | AC-OWR-15 rc=0 top-pass=1 all-pass=28 fail=0
```

`go test ./internal/cli/update/backup/ -count=1` (건드린 패키지 전체, 작은 패키지): 종료 코드 0, `ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.104s`.

**회귀: 이 카드의 시험 전부 (앵커된 이름 67개 — 앞 chunk 의 66개 + 새 흐름 시험 하나; 이름은 카드가 `cdd72a814` 이후 바꾼 `internal/cli/*_test.go` 에서 뽑고 기준선 11 + 가드를 더했다; `internal/cli` 패키지 전체 실행 아님)**, 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(<아래 이름>)$' -count=1 -v -coverprofile=<파일>`, 종료 코드 0, 최상위 `--- PASS`=67 `--- FAIL`=0 `--- SKIP`=0, 부하 평균 시작 `34.09 27.35 27.91` 끝 `36.52 34.25 30.88`:

```text
PASS
coverage: 10.0% of statements
```

```text
TestAliasToAKeyTheManifestLacksIsNoAlias
TestBackupFailureRemovesNothingThroughProductionPath
TestCaseVariantCounterpartAndMirrorSpelling
TestCaseVariantOnHostFilesystem
TestCaseVariantResolvesToRegisteredEntry
TestCommandsAndAgentsDiscoverableAfterInterruptedRun
TestCommandsSingleDispositionAndPresent
TestCounterpartLocationAndStatusEdges
TestDowngradeDropsEvidenceNotSafety
TestDryRunChangesNothingAndMatchesVerdict
TestDryRunRunsNoBinaryUpdateStep
TestEmptiedSkillDirRemoved
TestEmptyDirProvenanceVerdict
TestInterruptedMigrationRecoversAndConverges
TestKeepReasonVocabularyIsClosed
TestLegacyManifestRoundTrip
TestManagedOnlyEmptyDirsPruned
TestMigrationBackupBeforeRemoval
TestMigrationClassifiesUnregisteredMirrorCopy
TestMigrationDefaultKeepsConfirmedProjectCopy
TestMigrationIdempotent
TestMigrationKeepsSharedCopyUnderProfileReach
TestMigrationPreservesExistingMirrorEntries
TestMigrationRemovalGate
TestMigrationRemovesSharedCopyWithoutProfileReach
TestNoProfileRootAccessInUpdateSources
TestOwnershipAttestedManagedOnlyWithFullEvidence
TestOwnershipAttestedManagedThroughOperatingSource
TestOwnershipHelperFixtureBuilders
TestOwnershipHelperForcedRunReachesReconciliation
TestOwnershipHelperTreeSnapshotSeesDirMtime
TestOwnershipHelperUsesRealEmbeddedPaths
TestOwnershipPreserveByteEqualUserFile
TestOwnershipPreserveUserModifiedFile
TestOwnershipUnregisteredUnknownOriginKeptWithReason
TestOwnershipVerdictClasses
TestPreservePriorManifestOutcomes
TestPreviewOmitsRemovalCandidateWhileSwitchOff
TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy
TestPreviewVerdictsFailClosed
TestPriorGenerationCorruptIsNoEvidence
TestPriorGenerationSourceBoundedAndTolerant
TestPriorGenerationsAreLookupBoundOnly
TestPriorManifestAtomicWriteFailures
TestPriorManifestRetainedInRunBackup
TestProfileCallsiteBaselineMatchesTree
TestProfileEnvDoesNotReachProfileRoot
TestProfileMappedCopyKeptWithoutCounterpart
TestProfileReachMarkUsesTheRegisteredKeyUnderACaseAlias
TestProfileScanCatchesMutants
TestPruneNestedEmptyDirGuards
TestRF2F3b_MigrationPreservesUntracked
TestRemainingListKeeps313UnregisteredCopies
TestRemainingRenderEdges
TestRemovalSwitchStaysOffInProduction
TestRemovedAssetBackupRestoreByteIdentical
TestRepeatedUpdateIsIdempotent
TestShouldSkipBinaryUpdate_DevBuild
TestShouldSkipBinaryUpdate_DevBuild_Phase5_V2
TestShouldSkipBinaryUpdate_EnvVar
TestShouldSkipBinaryUpdate_EnvVar_Phase5
TestShouldSkipBinaryUpdate_Normal
TestShouldSkipBinaryUpdate_TemplatesOnlyFlag
TestUpdateDoesNotRestorePriorManifestIntoConfig
TestUpdateDryRun_ZeroMutation
TestVerdictAndPreviewCodeHasNoWritableCalls
TestViewOfManifestIsAReadOnlyCopy
```

**운영 코드 변이 둘 (트리 `051ddc18ef`, 파일을 `cmp` 로 원본과 대조해 `RESTORED-IDENTICAL`; 출력 줄 추림).** (1) 발견 2: `backupRemovedAsset` 가 모드와 무관하게 `defs.FilePerm` 으로 쓴다 — AC-OWR-15 가 붉다:

```text
    migrate_ownership_restore_test.go:629: (i-mode) the backup copy of .claude/skills/moai-workflow-project/scripts/navigator-audit.sh has mode 644, the file had 755
    migrate_ownership_restore_test.go:629: (iii-mode) the restored .claude/skills/moai-workflow-project/scripts/navigator-audit.sh has mode 644, it had 755
--- FAIL: TestRemovedAssetBackupRestoreByteIdentical (9.06s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/BackupCopyModes (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/RestoredModes (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/DifferenceSentinel (0.00s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/SharedClockNamesJudgeTheSame (0.06s)
    --- FAIL: TestRemovedAssetBackupRestoreByteIdentical/ExpectedSetMembersThatDiffer (0.08s)
```

(2) 발견 1: 건너뛰는 조건을 꺼 둔다 — 복원 수준 시험 셋이 붉고 `sections/` 있는 사례는 초록 그대로이다:

```text
--- FAIL: TestRestoreLegacySkipsPriorManifest (0.00s)
--- FAIL: TestRestoreRetainedWithoutSectionsSkipsPriorManifest (0.00s)
--- PASS: TestRestoreRetainedWithSectionsSkipsPriorManifest (0.00s)
--- FAIL: TestRestoreFromBackupDirSkipsPriorManifest (0.01s)
rc=1
```

**발견 1에서 더 점검한 것 (모두 `internal/cli/update/backup` 아래의 걷기·세는 코드, 코드 읽기와 필요한 곳의 실행으로).**

1. `RestoreMoaiConfigRetained` 의 걷기: `sections/` 만, `.yaml`/`.yml` 만 — 루트의 JSON 이 들어갈 수 없음. 시험으로 특성화.
2. `RestoreMoaiConfigLegacy`: 구멍 (고침). `RestoreFromBackupDir` (--restore 진입점, 두 번 적용) 와 `RestoreMoaiConfig` 도 같은 걷기를 거친다 — 시험 셋 모두 포함.
3. `RestoreUnmanagedConfigDirsSkipping`: `evaluator-profiles`, `astgrep-rules` 두 디렉터리만 — 해당 없음.
4. `base_loader.go`, `snapshot.go`, `backup.go` 의 걷기: 원본 설정과 템플릿 스냅샷 쪽을 걷는다 (백업 디렉터리의 루트가 아님) — 해당 없음.
5. 보존 정책 센서스 (`retention_prune.go` `takeCensus`/`isBookkeepingFile`): 부기 파일은 이름 셋 (실행 표식, `.complete`, `backup_metadata.json`) 뿐이라 `manifest.prior.json` 은 내용 키와 크기에 들어간다. **바꾸지 않았다:** REQ-OWR-017 이 보존 정책 변경을 금지하고 갭 G-K (센서스가 이 파일로 예산을 바꾸는지) 는 측정되지 않은 채이다 — 잔여 위험으로 남긴다.
6. **같은 부류로 새로 확인한 것 (고치지 않음, 최소 변경 지시): Legacy 걷기는 `removed-assets/**` 와 `in-memory-backups/**` 도 `.moai/config/` 로 복사한다.** 임시 시험(삭제함)으로 확인했다 (`removed-assets copied into config: true`, `in-memory-backups copied into config: true`). `in-memory-backups` 는 이 카드 이전부터이고 (UDS 카드), `removed-assets` 는 t1668 의 것이며 제거 스위치가 꺼져 있어 운영에서는 비어 있다. 같은 한 줄 부류의 수정이므로 리더가 원하면 후속 카드에서 고친다.

**품질 (트리 `051ddc18ef`)**

```text
go build ./... rc=0
GOOS=windows go build ./... rc=0
gofmt -l internal/cli internal/manifest rc=0 output-bytes=0
go vet ./internal/cli/ ./internal/cli/update/backup/ ./internal/manifest/ rc=0 output-bytes=0
GOOS=windows go vet (same three packages) rc=0 output-bytes=0
```

`golangci-lint run ./internal/cli/ ./internal/cli/update/backup/ ./internal/manifest/` 종료 코드 1: 카드 파일에 대한 지적 0건, 남은 셋은 이 카드가 만지지 않은 기존 errcheck:

```text
internal/cli/factory_run_adopt.go:79:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:91:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:103:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
3 issues:
* errcheck: 3
```

**하드 제약 점검 (최종 트리)**

```text
--- protected/profile (card commits)
exit=0
--- positive control
      29
--- diff names protected/profile/audit-receipts/templates
grep-exit=1
--- removalEnabled
grep-exit=1
40:var removalEnabled = false
--- SPEC bodies since 7b758103d8
grep-exit=1
--- files touched by chunk 5 so far
.moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/progress.md
internal/cli/migrate_ownership_restore_test.go
internal/cli/migrate_project_assets.go
internal/cli/update/backup/restore.go
internal/cli/update/backup/restore_prior_manifest_test.go
internal/cli/update/backup/run_marker.go
internal/cli/update_prior_manifest.go
internal/cli/update_prior_manifest_restore_test.go
```

**범위가 좁은 구문 커버리지 (위 회귀 한 번의 프로필 + `update/backup` 패키지 전체 시험 한 번; 앞 chunk 와 같은 계산)**

```text
internal/cli/migrate_ownership.go (created) 96.0% (119/124 statements)
internal/cli/migrate_ownership_case.go (created) 96.6% (57/59 statements)
internal/cli/migrate_ownership_emptydir.go (created) 93.3% (84/90 statements)
internal/cli/migrate_ownership_env.go (created) 94.4% (34/36 statements)
internal/cli/migrate_ownership_evidence.go (created) 100.0% (25/25 statements)
internal/cli/migrate_ownership_remaining.go (created) 100.0% (67/67 statements)
internal/cli/update_prior_manifest.go (created) 87.5% (70/80 statements)
internal/cli/migrate_project_assets.go (edited) whole-file 82.7% (220/266); card-changed lines 87.9% (102/116)
internal/cli/update.go (edited) whole-file 35.7% (146/409); card-changed lines 100.0% (5/5)
internal/cli/update_dryrun_preview.go (edited) whole-file 86.1% (31/36); card-changed lines 100.0% (8/8)
internal/cli/update_template_sync.go (edited) whole-file 64.9% (276/425); card-changed lines 90.3% (28/31)
internal/cli/update/backup/restore.go (edited, whole backup package tests) whole-file 89.1% (106/119); card-changed lines 100.0% (2/2)
internal/cli/update/backup/run_marker.go (edited): the card added one constant, no statements
```

`migrate_project_assets.go` 전체 82.7% 는 85% 미만이다 (이 카드 이전의 t1668 코드가 대부분; 카드가 바꾼 줄 87.9%). 새로 더한 `mode` 가지는 시험이 덮는다 (AC-OWR-15 의 실행 파일).

**정직한 회계 / 갭 / 잔여 위험 (chunk 5)**

1. AC-OWR-15 시험이 실행 파일 후보를 하나 더 가지게 되어 후보가 셋에서 넷이 되었다. 앞 chunk 의 AC-OWR-15 기록 수치(변이 개수 등)는 이 시험의 이전 판에 대한 것이고 변이 `BackupCopyLosesTheExecuteBit` 하나와 보고 묶음 둘 (`BackupCopyModes`, `RestoredModes`) 이 늘었다.
2. 발견 2 의 모드 보존은 `os.Stat` 의 권한 비트(`Mode().Perm()`) 이다. setuid/sticky 비트와 소유자는 보존하지 않는다 (공통 자산 파일에는 해당 없음). `os.Stat` 이 실패하면 `defs.FilePerm` 으로 되돌아간다 (제거 자체는 그 파일을 이미 읽었으므로 실패 가능성이 낮다).
3. 발견 1 의 복원 수준 시험은 실제 업데이트의 Restore 순서(`RestoreFromBackupDir` 두 번 적용)에서 두 번째 적용이 `notes/manifest.prior.json` 에 대해 「merge failed ... restoring backup」 경고를 낸다는 것을 보았다 — 이미 있는 비 YAML 대상에 대한 기존 동작이고 이 수정과 무관하다.
4. 실행하지 않은 것: `internal/cli` 패키지 전체 시험과 패키지 전체 커버리지, `-race`, Windows 시험 실행 (빌드와 vet 만). 착지 재측정 몫이다.

#### Chunk 6 (second card-review finding)

코덱스 카드 리뷰(조언용, HEAD `f9c4c76d4f`)의 새 P2 하나를 lane 이 코드로 확인했다: `classifyOwnershipCore` 의 등록 항목 경로(`os.ReadFile(abs)`)와 `classifyUnregistered` 가 공통 자산 걷기가 넘겨주는 모든 비디렉터리를 읽는다. 등록되지 않은 FIFO (소켓, 장치, 그것을 가리키는 심볼릭 링크도) 는 쓰는 쪽이 나타날 때까지 읽기가 막혀 제거 스위치가 꺼져 있어도 `moai update` 가 멈춘다. 기준 `cdd72a814` 는 미등록 경로를 읽지 않고 보존했고, 등록 경로의 읽기는 기준에도 있었다 (`migrate_project_assets.go` 의 기준판 236행) — 새로 생긴 퇴보는 미등록 쪽이고, 등록 쪽은 같은 결함을 이어받았다. 두 쪽을 판정 함수 한 곳에서 막는다.

**커밋 A `2eeb0559bd`** (시험만: `internal/cli/migrate_ownership_nonregular_test.go` — `//go:build !windows`, `syscall.Mkfifo`; `migrate_ownership_vocabulary_test.go` 의 닫힌 어휘 시험에 새 이유를 의도적으로 추가). 부하 평균 18.36 36.90 45.93.

시험 구성: 걷기 아래에 미등록 FIFO, 등록한 뒤 FIFO 로 바꾼 경로, FIFO 를 가리키는 심볼릭 링크를 두고, 곁에 평범한 미등록 파일, 그 파일을 가리키는 심볼릭 링크, 아무것도 가리키지 않는 심볼릭 링크, 제거 대상 managed 파일 (도달 대조군) 을 둔다. 마이그레이션을 고루틴에서 돌리고 **제한된 대기는 끝나지 않음을 알아내는 데에만 쓴다** (속도 단언 아님). 막혔으면 FIFO 를 반복해서 열었다 닫아 (쓰는 쪽을 잠깐 만들어 줌) 읽기를 끝내고 고루틴이 끝나길 기다린 뒤 실패로 보고한다 — 시험이 멈추지 않고 실패한다. 이유 어휘의 새 이름은 `non-regular-file` (시험은 리터럴로 적어 두어, 상수가 없는 A 시점에도 컴파일되고 동작 수준에서 붉다).

**기존 심볼릭 링크 동작 (조용히 바꾸지 않기 위해 수정 전 트리에서 따로 쟀다; 임시 시험, 삭제함).** `os.ReadFile` 은 링크를 따라가므로 가리키는 파일이 일반 파일이면 그 바이트로 판정되고, 끊어진 링크는 읽기 실패이다:

```text
    zz_scratch_test.go:20: SCRATCH .agents/skills/moai-nonregular/plain.md row=true class=unregistered reason=no-provenance-record
    zz_scratch_test.go:20: SCRATCH .agents/skills/moai-nonregular/link-to-plain row=true class=unregistered reason=no-provenance-record
    zz_scratch_test.go:20: SCRATCH .agents/skills/moai-nonregular/link-to-nothing row=true class=unregistered reason=read-failed
--- PASS: TestScratchSymlinkBaseline (0.69s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.930s
```

수정은 이 둘을 바꾸지 않는다 (시험이 같은 값을 단언). 새 이유는 가리키는 대상이 일반 파일이 아닌 경로에만 붙는다.

**RED 셀 (트리 `2eeb0559bd`, 작업 트리 깨끗).** 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestNonRegularFilesAreKeptWithoutBeingRead|TestKeepReasonVocabularyIsClosed)$' -count=1 -v`, 종료 코드 1, 시험 로그 줄과 결과 줄만 (원문 순서). 마이그레이션이 30초 안에 돌아오지 못했다 (FIFO 읽기에 막힘 — 결함의 직접 관측), 어휘 시험은 새 이유가 닫힌 집합에 없다고 붉다:

```text
    migrate_ownership_nonregular_test.go:227: the migration did not return: it is blocked reading a FIFO
--- FAIL: TestNonRegularFilesAreKeptWithoutBeingRead (31.38s)
    --- PASS: TestNonRegularFilesAreKeptWithoutBeingRead/MutantHungIsCaught (0.00s)
    migrate_ownership_vocabulary_test.go:32: defined reason "non-regular-file" is not in the closed set
--- FAIL: TestKeepReasonVocabularyIsClosed (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	32.594s
FAIL
```

**커밋 C `1d5ecbbc7c`** (`internal/cli/migrate_ownership.go`, `internal/cli/migrate_project_assets.go`). `classifyOwnershipCore` — 판정의 유일한 진입점 — 가 두 갈래 모두에서 읽기 전에 경로의 종류를 `os.Stat` 으로 묻는다 (`unreadableKind`; Stat 은 `ReadFile` 처럼 링크를 따라가고 경로를 열지 않으므로 막히지 않는다). FIFO, 소켓, 장치, 특수 파일, 또는 그것을 가리키는 심볼릭 링크는 새 닫힌 어휘 이유 `non-regular-file` 로 보존한다: 미등록 경로는 미등록 그대로, 등록 경로는 읽기 실패가 이미 그랬듯 `user-modified` 로. 디렉터리, 끊어진 링크, 조사할 수 없는 경로는 새 이유를 받지 않고 예전처럼 읽기에서 실패한다 (`read-failed`) — 일반 파일을 가리키는 링크는 가리키는 파일의 바이트로 판정된다 (수정 전 측정과 같음). 걷기의 보고 줄은 새 이유를 이름 붙여 설명한다.

**GREEN (트리 `1d5ecbbc7c`; 연속 5회, 종료 코드는 각 실행에서 읽음).** 새 선택자 `-run '^(TestNonRegularFilesAreKeptWithoutBeingRead|TestKeepReasonVocabularyIsClosed)$'` (`TMPDIR=/private/tmp/own-fixture-d1 ... -count=1 -v`; 상위 둘 + 변이 서브 시험 넷):

```text
run 1: rc=0 PASS-lines=6 FAIL-lines=0
run 2: rc=0 PASS-lines=6 FAIL-lines=0
run 3: rc=0 PASS-lines=6 FAIL-lines=0
run 4: rc=0 PASS-lines=6 FAIL-lines=0
run 5: rc=0 PASS-lines=6 FAIL-lines=0
```

**회귀: 이 카드의 시험 전부 (앵커된 이름 68개 — 카드가 `cdd72a814` 이후 바꾼 `internal/cli/*_test.go` 에서 뽑은 이름 + 기준선 11 + 가드; 새 시험 하나가 늘었다; 패키지 전체 실행 아님)**, 명령 `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(<아래 이름>)$' -count=1 -v -coverprofile=<파일>`, 종료 코드 0, 최상위 `--- PASS`=68 `--- FAIL`=0 `--- SKIP`=0, 부하 평균 시작 `37.71 37.52 43.50` 끝 `63.15 80.95 64.92` (이 머신이 평소보다 훨씬 바빴다: 실행 415초):

```text
PASS
coverage: 10.0% of statements
```

```text
TestAliasToAKeyTheManifestLacksIsNoAlias
TestBackupFailureRemovesNothingThroughProductionPath
TestCaseVariantCounterpartAndMirrorSpelling
TestCaseVariantOnHostFilesystem
TestCaseVariantResolvesToRegisteredEntry
TestCommandsAndAgentsDiscoverableAfterInterruptedRun
TestCommandsSingleDispositionAndPresent
TestCounterpartLocationAndStatusEdges
TestDowngradeDropsEvidenceNotSafety
TestDryRunChangesNothingAndMatchesVerdict
TestDryRunRunsNoBinaryUpdateStep
TestEmptiedSkillDirRemoved
TestEmptyDirProvenanceVerdict
TestInterruptedMigrationRecoversAndConverges
TestKeepReasonVocabularyIsClosed
TestLegacyManifestRoundTrip
TestManagedOnlyEmptyDirsPruned
TestMigrationBackupBeforeRemoval
TestMigrationClassifiesUnregisteredMirrorCopy
TestMigrationDefaultKeepsConfirmedProjectCopy
TestMigrationIdempotent
TestMigrationKeepsSharedCopyUnderProfileReach
TestMigrationPreservesExistingMirrorEntries
TestMigrationRemovalGate
TestMigrationRemovesSharedCopyWithoutProfileReach
TestNoProfileRootAccessInUpdateSources
TestNonRegularFilesAreKeptWithoutBeingRead
TestOwnershipAttestedManagedOnlyWithFullEvidence
TestOwnershipAttestedManagedThroughOperatingSource
TestOwnershipHelperFixtureBuilders
TestOwnershipHelperForcedRunReachesReconciliation
TestOwnershipHelperTreeSnapshotSeesDirMtime
TestOwnershipHelperUsesRealEmbeddedPaths
TestOwnershipPreserveByteEqualUserFile
TestOwnershipPreserveUserModifiedFile
TestOwnershipUnregisteredUnknownOriginKeptWithReason
TestOwnershipVerdictClasses
TestPreservePriorManifestOutcomes
TestPreviewOmitsRemovalCandidateWhileSwitchOff
TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy
TestPreviewVerdictsFailClosed
TestPriorGenerationCorruptIsNoEvidence
TestPriorGenerationSourceBoundedAndTolerant
TestPriorGenerationsAreLookupBoundOnly
TestPriorManifestAtomicWriteFailures
TestPriorManifestRetainedInRunBackup
TestProfileCallsiteBaselineMatchesTree
TestProfileEnvDoesNotReachProfileRoot
TestProfileMappedCopyKeptWithoutCounterpart
TestProfileReachMarkUsesTheRegisteredKeyUnderACaseAlias
TestProfileScanCatchesMutants
TestPruneNestedEmptyDirGuards
TestRF2F3b_MigrationPreservesUntracked
TestRemainingListKeeps313UnregisteredCopies
TestRemainingRenderEdges
TestRemovalSwitchStaysOffInProduction
TestRemovedAssetBackupRestoreByteIdentical
TestRepeatedUpdateIsIdempotent
TestShouldSkipBinaryUpdate_DevBuild
TestShouldSkipBinaryUpdate_DevBuild_Phase5_V2
TestShouldSkipBinaryUpdate_EnvVar
TestShouldSkipBinaryUpdate_EnvVar_Phase5
TestShouldSkipBinaryUpdate_Normal
TestShouldSkipBinaryUpdate_TemplatesOnlyFlag
TestUpdateDoesNotRestorePriorManifestIntoConfig
TestUpdateDryRun_ZeroMutation
TestVerdictAndPreviewCodeHasNoWritableCalls
TestViewOfManifestIsAReadOnlyCopy
```

**운영 코드 변이 하나 (트리 `1d5ecbbc7c`, 파일을 `cmp` 로 원본과 대조해 `RESTORED-IDENTICAL`).** `unreadableKind` 가 항상 거짓을 돌려준다 (가드 제거): 시험이 멈추지 않고 붉다 (마이그레이션이 30초 안에 돌아오지 못함):

```text
    migrate_ownership_nonregular_test.go:227: the migration did not return: it is blocked reading a FIFO
--- FAIL: TestNonRegularFilesAreKeptWithoutBeingRead (34.41s)
    --- PASS: TestNonRegularFilesAreKeptWithoutBeingRead/MutantHungIsCaught (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	36.249s
FAIL
rc=1
```

**점검한 것 (같은 읽기 결함이 다른 곳에 있는지).**

1. `migrate_ownership.go` 의 읽기 두 곳 (등록 경로 `ReadFile`, `classifyUnregistered` 의 `ReadFile`): 둘 다 가드 뒤에 있다. 등록된 `user_created` 는 원래 읽지 않고 먼저 돌아온다.
2. 미리보기 (`update_dryrun_preview.go`, `migrate_ownership_env.go`): 판정을 `classifyOwnership` 로 부르므로 같은 가드를 거친다. 파일 읽기는 따로 없다.
3. 빈 디렉터리 정리 (`migrate_ownership_emptydir.go`): `Lstat`, `ReadDir`, `os.Remove` 만 쓴다 — 파일을 열지 않는다. 디렉터리 `ReadDir` 은 FIFO 항목이 있어도 막히지 않는다.
4. 대소문자 별칭 (`migrate_ownership_case.go`): `os.Stat`/`os.SameFile` 만, 내장 템플릿은 `fs.ReadDir`.
5. 이전 세대 읽기 (`update_prior_manifest.go`): `Lstat` 뒤 `IsRegular()` 가 아니면 건너뛴다 (M1b 가 이미 막음). 시작 매니페스트 읽기 `readStartManifest` 는 매니페스트 파일 하나이고 `manifest.Load` 와 같은 대상이다 (이 카드 이전부터 같은 파일을 읽는다).
6. `userCounterpartConfirmed` (`migrate_project_assets.go` 의 사용자 홈 쪽 `ReadFile`): 기준 이전부터 있는 코드이고 읽는 대상이 프로젝트 안이 아니라 사용자 루트의 설치 파일이다. 건드리지 않았다.
7. 기준 `cdd72a814` 의 등록 경로 읽기 (기준판 `migrate_project_assets.go` 236행 `os.ReadFile(p)`) 도 가드가 없었다: 등록된 FIFO 는 기준에서도 막혔을 것이다. 퇴보는 미등록 쪽이지만 두 쪽을 함께 막았다.

**품질 (트리 `1d5ecbbc7c`)**

```text
go build ./... rc=0
GOOS=windows go build ./... rc=0
gofmt -l internal/cli internal/manifest rc=0 output-bytes=0
go vet ./internal/cli/ ./internal/cli/update/backup/ ./internal/manifest/ rc=0 output-bytes=0
GOOS=windows go vet (same three packages) rc=0 output-bytes=0
```

`golangci-lint run ./internal/cli/ ./internal/cli/update/backup/ ./internal/manifest/`: 카드 파일에 대한 지적 0건, 남은 셋은 이 카드가 만지지 않은 기존 errcheck:

```text
internal/cli/factory_run_adopt.go:79:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:91:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:103:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
3 issues:
* errcheck: 3
```

**하드 제약 점검**

```text
--- protected/profile (card commits)
exit=0
--- positive control
      31
--- diff names protected/profile/audit-receipts/templates
grep-exit=1
--- removalEnabled
grep-exit=1
40:var removalEnabled = false
--- SPEC bodies since f9c4c76d4f
grep-exit=1
--- files touched by chunk 6
.moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/progress.md
internal/cli/migrate_ownership.go
internal/cli/migrate_ownership_nonregular_test.go
internal/cli/migrate_ownership_vocabulary_test.go
internal/cli/migrate_project_assets.go
```

**범위가 좁은 구문 커버리지 (위 회귀 한 번의 프로필)**

```text
internal/cli/migrate_ownership.go: whole-file 96.2% (128/133); chunk-6 changed lines 100.0% (10/10 statements in blocks touching the chunk-6 diff)
internal/cli/migrate_project_assets.go: whole-file 82.8% (221/267); chunk-6 changed lines 100.0% (1/1 statements in blocks touching the chunk-6 diff)
```

**정직한 회계 / 잔여 위험 / 갭.**

1. **이유 어휘가 늘었다.** design.md D-1 은 닫힌 집합이 SPEC 개정으로만 늘어난다고 적는다. `non-regular-file` 은 시험과 닫힌 어휘 시험에 의도적으로 더했고 SPEC 본문은 고치지 않았다 — 필요하면 manager-spec 재위임 사안이다 (기준 AC-OWR-04/-14 의 일반 파일 이유 기대는 바뀌지 않았고 회귀에서 통과).
2. 종류 확인과 읽기 사이의 경합 (그 사이에 일반 파일이 FIFO 로 바뀜) 은 막지 않는다. 열기를 `O_NONBLOCK` 으로 하면 막을 수 있으나 최소 변경 범위를 넘는다.
3. `os.ModeIrregular` 와 소켓·장치 종류는 FIFO 와 같은 분기를 타지만 시험은 FIFO 만 만든다 (소켓과 장치 파일을 이식성 있게 만들 수 없다). 같은 비트 마스크 한 줄이다.
4. 이 머신에서 `syscall.Mkfifo` 가 실패하면 시험은 건너뛴다 (`t.Skipf`) — 이 파일 시스템 능력 확인 외에 시험 제외는 없다. 이 실행들에서는 건너뛰지 않았다 (SKIP=0).
5. 업데이트 흐름의 다른 앞 단계(조정 분석 등 이 카드가 만들지 않은 코드)가 FIFO 를 읽는지는 점검하지 않았다 — 기준에도 있던 코드이다.
6. 실행하지 않은 것: `internal/cli` 패키지 전체 시험과 패키지 전체 커버리지, `-race`, Windows 시험 실행 (빌드와 vet 만; 새 시험은 `!windows` 빌드 태그).

**Chunk 7 addendum (추가만).** 레인 검토가 이 chunk 의 `migrate_ownership_nonregular_test.go` 가 `syscall.Mkfifo` 실패 때 `t.Skipf` 하던 것을 짚었다: 건너뛴 시험은 아무것도 쓸어내지 않아 이 시험이 막으려는 멈춤이 mkfifo 를 거부하는 파일 시스템에서 보이지 않게 돌아올 수 있다. 커밋 `2b801c1971` 이 이를 `t.Fatalf` (같은 메시지 + 디렉터리 경로) 로 바꿨다. 위 「정직한 회계」 4번의 「건너뛴다 (`t.Skipf`)」 서술은 이 커밋으로 대체된다 (시험이 건너뛰지 않고 실패한다; 유일한 제외는 `//go:build !windows`). 저장한 출력: 해당 시험 연속 3회 rc=0 (각 PASS 5줄, FAIL 0, SKIP 0), 이 파일의 선택자 `^(TestNonRegularFilesAreKeptWithoutBeingRead|TestKeepReasonVocabularyIsClosed)$` 한 번 rc=0 (PASS 6줄, FAIL 0, SKIP 0, 부하 평균이 100 안팎인 머신), `gofmt -l internal/cli` 출력 없음, `go vet ./internal/cli/` 호스트·Windows 모두 rc=0. 이 카드가 `cdd72a814` 이후 바꾼 시험 파일의 `t.Skip` 전수 조사 결과: (1) `migrate_ownership_nonregular_test.go` — 위에서 고침; (2) `migrate_ownership_test.go:381` `t.Skipf("symlinks unavailable here: %v", err)` — 이 카드가 만든 파일(M1a, `9b73a35938`)의 심볼릭 링크 탈출 사례에서 `os.Symlink` 실패 때 건너뜀. 같은 규칙에 걸리지만 지시에 따라 이번에는 고치지 않았고 리더가 정하도록 보고한다 (고친다면 같은 모양의 `t.Fatalf`); (3) `migrate_ownership_helpers_test.go:1008` 은 `t.Logf` 의 `skipped=` 변수 출력이지 건너뛰기가 아니다. 그 밖의 파일에는 `t.Skip`, `SkipNow`, `testing.Short` 가 없다.

**Chunk 8 addendum (추가만).** 레인 결정: `migrate_ownership_test.go` 의 `t.Skipf("symlinks unavailable here")` 도 같은 규칙(건너뛴 시험은 아무것도 쓸어내지 않음)에 걸린다. 커밋 `4f970def44` 가 `TestOwnershipVerdictClasses` 의 뿌리 경계(심볼릭 링크 탈출) 사례를 새 시험 `TestOwnershipVerdictSymlinkEscape` 로 `migrate_ownership_symlink_test.go` (`//go:build !windows`) 로 옮겼고, 단언은 그대로이며 `os.Symlink` 실패는 경로와 오류를 담은 `t.Fatalf` 이다. 부모 시험은 나머지 사례를 그대로 가지며, 건너뛰기가 가로막던 뒤쪽의 폐기(deprecated) 사본 점검이 이제 항상 실행된다. 이동한 사례는 태그 없는 파일에서 참조되지 않는다 (`grep` 결과 없음); `GOOS=windows go vet ./internal/cli/` rc=0. 저장한 출력: 이동한 시험과 부모 시험의 앵커 선택자 `^(TestOwnershipVerdictSymlinkEscape|TestOwnershipVerdictClasses)$` 연속 3회 모두 rc=0, PASS 2줄, FAIL 0, SKIP 0; `gofmt -l internal/cli` 출력 없음; `go vet ./internal/cli/` 호스트 rc=0. 카드가 `cdd72a814` 이후 바꾼 모든 시험 파일의 `t.Skip`/`Skipf`/`SkipNow`/`testing.Short` 재조사: 호출 없음 (남은 일치는 `first.Skipped` 변수 이름과 새 파일 머리 주석의 문구뿐).

### Chunk 9 (sync-audit repair) — 기록 시작

독립 sync 감사 (HEAD `dfd7c40fb2`, 영수증 `rcpt-7770ccadc2f78594b050e777`) 가 FAIL 을 냈고 레인이 F1, F2, F3 을 받아들였다 (F4 는 거절, F5/F6 은 공개 사항). 근거 문서는 `.moai/reports/t1865/sync-audit.md` 의 Findings 절이며 이 chunk 는 그 파일을 고치지 않았다. 각 지적마다 A (시험만) → B (이 기록) → C (구현) 순서이고 순서의 증거는 커밋 그래프이다.

#### F1 — 등록 안 된 심볼릭 링크가 프로젝트 밖 파일을 읽는다 (REQ-OWR-005, AC-OWR-08 계열)

**A** `6315473594` — `internal/cli/migrate_ownership_unregsymlink_test.go` (`//go:build !windows`, `TestUnregisteredSymlinkOutsideProjectIsNotRead`). 판정 함수를 직접 부르는 절 (제거 스위치 끔·켬 각각)과 실제 걷기를 도는 절로 나뉜다. 걷기 절은 분류기를 감싸 각 판정이 실은 `v.data` 길이를 기록하고, 읽기가 일어났다면 그 길이가 0 이 아니게 된다. 통제: 프로젝트 안 링크는 대상 바이트로 판정(`no-provenance-record`), 끊어진 링크는 `read-failed`, 같은 걷기가 `migIdenticalSkill` 을 제거(도달 확인), 프로필 모양 대상 파일의 바이트 불변, 링크 자체 불변.

RED-now (명령 · 트리 SHA `6315473594` · 종료 코드 · 원문):

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestUnregisteredSymlinkOutsideProjectIsNotRead$' -count=1 -v
=== RUN   TestUnregisteredSymlinkOutsideProjectIsNotRead
    migrate_ownership_unregsymlink_test.go:109: switch=false .agents/skills/moai-unreg-symlink/to-profile: the verdict function read 34 bytes from outside the project
    migrate_ownership_unregsymlink_test.go:112: switch=false .agents/skills/moai-unreg-symlink/to-profile: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
    migrate_ownership_unregsymlink_test.go:109: switch=false .agents/skills/moai-unreg-symlink/chained: the verdict function read 34 bytes from outside the project
    migrate_ownership_unregsymlink_test.go:112: switch=false .agents/skills/moai-unreg-symlink/chained: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
    migrate_ownership_unregsymlink_test.go:109: switch=true .agents/skills/moai-unreg-symlink/to-profile: the verdict function read 34 bytes from outside the project
    migrate_ownership_unregsymlink_test.go:112: switch=true .agents/skills/moai-unreg-symlink/to-profile: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
    migrate_ownership_unregsymlink_test.go:109: switch=true .agents/skills/moai-unreg-symlink/chained: the verdict function read 34 bytes from outside the project
    migrate_ownership_unregsymlink_test.go:112: switch=true .agents/skills/moai-unreg-symlink/chained: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
    migrate_ownership_unregsymlink_test.go:148: .agents/skills/moai-unreg-symlink/to-profile: row class="unregistered" reason="no-provenance-record", want unregistered/"symlink-escape"
    migrate_ownership_unregsymlink_test.go:153: .agents/skills/moai-unreg-symlink/to-profile: the verdict carried 34 bytes read through the symlink
    migrate_ownership_unregsymlink_test.go:148: .agents/skills/moai-unreg-symlink/chained: row class="unregistered" reason="no-provenance-record", want unregistered/"symlink-escape"
    migrate_ownership_unregsymlink_test.go:153: .agents/skills/moai-unreg-symlink/chained: the verdict carried 34 bytes read through the symlink
--- FAIL: TestUnregisteredSymlinkOutsideProjectIsNotRead (0.95s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.436s
FAIL
rc=1
```

붉은 이유: 단언 실패이며 올바른 이유다. 프로필 모양 대상으로 가는 직접 링크와 사슬(링크 → 프로젝트 안 링크 → 밖) 둘 다 34 바이트를 읽었고(`read 34 bytes from outside the project`, `carried 34 bytes read through the symlink`) 이유가 `no-provenance-record` 로 나왔다 (`symlink-escape` 가 맞다). 통제 단언(프로젝트 안 링크, 끊어진 링크, 도달 확인, 대상 불변)은 한 줄도 실패하지 않았다.

#### F3 — 사용자 루트 대응물이 FIFO 이면 판정이 멈춘다

**A** `34c69c19fc` — `internal/cli/migrate_ownership_counterpartfifo_test.go` (`//go:build !windows`, `TestCounterpartFifoIsNotOpened`). 등록된 managed 파일의 사용자 루트 대응물을 FIFO 로 바꾼 뒤 (a) 판정 함수를 프로필 매핑 프로젝트와 일반 프로젝트로 각각 부르고 (b) 실제 걷기를 돈다. 멈춤은 `migrate_ownership_nonregular_test.go` 와 같은 방식 (고루틴 + 한정 대기는 멈춤 감지용일 뿐 속도 단언이 아님, 막힌 읽기는 FIFO 를 건드려 풀어 임시 트리보다 오래 살지 않게 함) 으로 감지한다. 기대: 대응물은 열리지 않고 `unconfirmed`, 프로젝트 사본 유지, 제거 0, 이유 `counterpart-unconfirmed` (프로필 매핑은 `profile-reach`).

RED-now (명령 · 트리 SHA `34c69c19fc` · 종료 코드 · 원문):

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCounterpartFifoIsNotOpened$' -count=1 -v
=== RUN   TestCounterpartFifoIsNotOpened
=== RUN   TestCounterpartFifoIsNotOpened/ProfileMappedProject
    migrate_ownership_counterpartfifo_test.go:112: the verdict did not return: it is blocked reading the user-root counterpart FIFO <tmp>/004/.claude/skills/moai-foundation-core/SKILL.md
=== RUN   TestCounterpartFifoIsNotOpened/OrdinaryProject
    migrate_ownership_counterpartfifo_test.go:112: the verdict did not return: it is blocked reading the user-root counterpart FIFO <tmp>/004/.claude/skills/moai-foundation-core/SKILL.md
=== RUN   TestCounterpartFifoIsNotOpened/MigrationReturnsAndKeepsTheProjectCopy
    migrate_ownership_counterpartfifo_test.go:123: the migration did not return: it is blocked reading the user-root counterpart FIFO <tmp>/004/.claude/skills/moai-foundation-core/SKILL.md
--- FAIL: TestCounterpartFifoIsNotOpened (92.33s)
    --- FAIL: TestCounterpartFifoIsNotOpened/ProfileMappedProject (30.98s)
    --- FAIL: TestCounterpartFifoIsNotOpened/OrdinaryProject (30.56s)
    --- FAIL: TestCounterpartFifoIsNotOpened/MigrationReturnsAndKeepsTheProjectCopy (30.79s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	93.691s
FAIL
rc=1
```

붉은 이유: 올바른 이유다. 세 절 모두 30 초 한정 대기 안에 반환하지 못했다 (판정 함수 두 절은 대응물 FIFO 의 읽기에서, 걷기 절은 같은 읽기에서 막힘). 원문의 임시 경로 접미사는 가독성을 위해 `<tmp>` 로 줄였다 (정규식 치환 한 줄; 나머지 바이트는 저장 파일과 같다).

#### F2 — 같은 시계 초의 두 실행이 `manifest.prior.json` 을 덮어쓴다 (REQ-OWR-017)

**조사 (관측한 것).**

1. **이름 규약과 읽는 쪽.** `defs.BackupTimestampFormat = "20060102_150405"` 는 초 단위다. `internal/cli/update/backup/backup.go:249` 의 `configBackupName = ^[0-9]{8}_[0-9]{6}$` 는 실행 표지(`.moai-run`)가 **없는** 디렉터리가 한 실행으로 세어지는 조건이다. `collectBackupRuns` 는 표지가 읽히면 이름 검사 없이 표지의 실행 id 로 묶고, 표지가 없고 이름도 안 맞으면 "세지도 지우지도 않는다" 이며 `scanUnmanagedConfigEntries` (retention_prune.go:322) 가 비관리 항목으로 보고만 한다. 그래서 접미사가 붙은 형제 디렉터리는 **같은 실행 표지를 가질 때만** 실행 수를 늘리지 않는다 (보존 정책 불변). `readPriorGenerations` (update_prior_manifest.go) 는 이름 정규식 없이 `.moai-backups` 의 모든 하위 디렉터리를 이름 역순으로 읽는다 — 고정 폭 접미사 `-NN` 이면 이름 순서가 곧 시간 순서다 (`-9` 대 `-10` 의 사전식 역전을 피하려고 두 자리 고정).
2. **이름을 해석하는 다른 곳.** `internal/core/project/validator.go:394` (매니페스트 치유) 는 `.moai-backups` 를 읽어 디렉터리 안의 `manifest.json` 만 보며 이름 형식을 해석하지 않는다. `clean_home.go:303` 은 다른 뿌리(`backups/removed-*`)다. 저장소 전체에서 `BackupTimestampFormat` 을 쓰는 곳은 `backup.go`, `update_disk_backup.go`, `deploy.go` (사전 정리), `reconcile.go`, `validator.go` 이고 모두 쓰는 쪽이다. 같은 초의 같은 이름 디렉터리를 여러 쓰는 쪽이 공유하는 것은 `deploy.go` 의 `pre-clean` 하위 디렉터리 등 이미 있는 관행이다.
3. **이 실행의 설정 백업 경로 자체도 같은 초에 공유되는가: 예.** `BackupMoaiConfig` 는 `MkdirAll(<ts>)` 이다. 측정 (덮어쓰기 탐침, 트리에 쓰지 않음; 저장 파일 `f2-probe.txt`): 같은 초에 두 번 부르면 반환 경로가 같고 (`same=true` 3회 중 3회), 설정 파일은 나중 실행의 것으로 덮이며, `WriteRunMarker` 는 나중 실행 id 로 표지를 바꾼다. 즉 설정 백업 쪽의 같은 초 공유는 이 카드 이전부터의 동작이고 이 수정은 건드리지 않는다. 이 카드가 만든 `manifest.prior.json` 만 이전 실행의 사본을 잃는 새 손실이었다.
4. **호출 흐름.** `update_template_sync.go:825` 는 `preservePriorManifest(projectRoot, configBackupPath, startManifest)` 가 돌려준 디렉터리를 `migrationRun.runBackupDir` 로 넘기고, `removed-assets/` 와 `ps.exclude` (현재 실행의 사본은 이전 세대로 읽지 않음) 가 그 디렉터리를 쓴다. 복원 쪽 걷기(`RestoreFromBackupDir`)는 `configBackupPath` 만 원본으로 쓰므로 형제 디렉터리는 복원 원본이 되지 않는다.

**선택한 가장 작은 안전한 수정.** `preservePriorManifest` 안에서, 의도한 디렉터리에 이미 `manifest.prior.json` 이 있으면 (한 실행은 한 번만 쓰므로 있다면 이전 실행의 것이다) 같은 부모 아래 `<이름>-02` ~ `-99` 중 `os.Mkdir` 이 성공하는 첫 이름의 새 디렉터리로 옮긴다. `MkdirAll` 이 아니므로 이미 있는 디렉터리 (다른 곳이 만든 것 포함)는 쓰지 않는다. 의도한 디렉터리의 `.moai-run` 표지가 있으면 그대로 복사해 보존 정책상 같은 실행으로 센다. 이름 형식을 파싱하는 다른 코드는 바꾸지 않았고 보존 정책에도 손대지 않았다. 이름을 못 얻으면 (99 개 소진, `Mkdir` 이 `EEXIST` 가 아닌 오류) `priorManifestFailed` — 그 실행은 파일을 하나도 지우지 않는다 (REQ-OWR-016).

**A** `ed58e64c3c` — `internal/cli/update_prior_manifest_collision_test.go` (`TestPriorManifestOfAnEarlierRunSurvivesALaterRun`, 시계를 읽지 않음: 같은 초 충돌은 같은 설정 백업 경로를 두 번 넘기는 것으로 직접 진술) + `migrate_ownership_recovery_test.go` 의 공유 디렉터리 모사 가지 (`SharedRunDirectory`: 앞 실행 내용을 뒤 실행 디렉터리에 합치는 모사로 손실을 기대값으로 단언했음)를 **무대기 연속 실행 가지** (`NoWaitConsecutiveRuns`)로 교체: 기다리지 않고 다음 실행을 시작하며 같은 초에 들든 아니든 단언은 구별 가지와 같다 (손실 기대 `keptDirs` 와 완화된 `strictRows` 제외 삭제).

RED-now (명령 · 트리 SHA `ed58e64c3c` · 종료 코드 · 원문; 임시 경로 접두사만 `<tmp>` 로 줄임):

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestPriorManifestOfAnEarlierRunSurvivesALaterRun$' -count=1 -v
=== RUN   TestPriorManifestOfAnEarlierRunSurvivesALaterRun
=== RUN   TestPriorManifestOfAnEarlierRunSurvivesALaterRun/TwoRunsSharingTheConfigBackupPath
    update_prior_manifest_collision_test.go:87: both runs wrote into <tmp>/.moai-backups/20261018_153045: the later copy replaced the earlier one
=== RUN   TestPriorManifestOfAnEarlierRunSurvivesALaterRun/ThreeRunsKeepNameOrderEqualToAgeOrderAndTheReaderSeesEveryEarlierOne
    update_prior_manifest_collision_test.go:112: directory names [20261018_153045 20261018_153045 20261018_153045] are not strictly increasing with age
    update_prior_manifest_collision_test.go:122: the reader saw [] (newest first), want [two one]
=== RUN   TestPriorManifestOfAnEarlierRunSurvivesALaterRun/ADirectoryTheRunDidNotCreateIsNeverWrittenInto
=== RUN   TestPriorManifestOfAnEarlierRunSurvivesALaterRun/NoFreeNameFailsClosedAndOverwritesNothing
    update_prior_manifest_collision_test.go:152: got ("<tmp>/.moai-backups/20261018_153045", 1, <nil>), want a failed state with an error: the run must remove nothing
    update_prior_manifest_collision_test.go:155: the earlier run's copy changed: {"files":{"two":{"provenance":"template_managed","template_hash":"sha256:two"}}}
=== RUN   TestPriorManifestOfAnEarlierRunSurvivesALaterRun/RunsWithoutAConfigBackupPathNeverShareACopy
    update_prior_manifest_collision_test.go:167: runs share a copy: <tmp>/.moai-backups/20261018_221827 <tmp>/.moai-backups/20261018_221827 <tmp>/.moai-backups/20261018_221827
=== RUN   TestPriorManifestOfAnEarlierRunSurvivesALaterRun/RetentionStillCountsTheSiblingWithItsRun
--- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun (0.10s)
    --- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/TwoRunsSharingTheConfigBackupPath (0.03s)
    --- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/ThreeRunsKeepNameOrderEqualToAgeOrderAndTheReaderSeesEveryEarlierOne (0.02s)
    --- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/ADirectoryTheRunDidNotCreateIsNeverWrittenInto (0.01s)
    --- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/NoFreeNameFailsClosedAndOverwritesNothing (0.02s)
    --- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/RunsWithoutAConfigBackupPathNeverShareACopy (0.01s)
    --- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/RetentionStillCountsTheSiblingWithItsRun (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.451s
FAIL
rc=1
```

붉은 이유: 올바른 이유다. 같은 설정 백업 경로를 두 번 넘기면 두 실행이 같은 디렉터리에 썼고 (`both runs wrote into ... the later copy replaced the earlier one`), 첫 사본은 둘째 것으로 바뀌었다 (`the earlier run's copy changed: ... "two"`), 세 번째까지 같은 이름이며 (`names ... not strictly increasing`), 읽는 쪽은 이전 세대를 하나도 보지 못한다 (`the reader saw []`). 99 개 소진 가지는 실패해야 할 때 성공으로 덮어썼다. `ADirectoryTheRunDidNotCreateIsNeverWrittenInto` 와 `RetentionStillCountsTheSiblingWithItsRun` 은 수정 전에도 통과하는 설계 가드다 (수정 뒤의 성질을 고정한다; RED 가 아니다). `RunsWithoutAConfigBackupPathNeverShareACopy` 는 세 호출이 한 초 안에 들어간 이 실행에서 붉었고, 다른 초에 들면 초록일 수 있는 시계 의존 RED 라서 결정적 RED 증거로는 앞의 절들을 쓴다.

#### Chunk 9 GREEN 기록과 정직한 회계

**등록 경로(registered path)의 결정.** F1 의 사전 읽기 문제는 등록된 파일 경로에도 있었다 (기준 커밋 `cdd72a814` 에서도 읽기가 경계 검사보다 앞섰음). 같은 보호 장치를 거기에 두면 AC-OWR-01 의 고정된 기대 — `TestOwnershipVerdictSymlinkEscape` 가 등록된 심볼릭 링크에서 class `managed` + 이유 `symlink-escape` 를 단언한다 — 를 바꾼다: class `managed` 는 대상 바이트가 템플릿 해시와 같다는 것을 읽어야 정해지기 때문이다. 레인 지시에 따라 AC 를 고치지 않고 등록 경로는 그대로 두며, **N16 계열의 공개된 부채(기준 시점부터 있었고 이 수정 밖)**로 기록한다: 등록된 managed 파일이 프로젝트 밖을 가리키는 심볼릭 링크이면 그 대상 바이트를 읽는다 (읽은 바이트는 해시 비교에 쓰이고, 제거는 `symlink-escape` 로 거부되므로 제거 후보도 백업 사본도 되지 않는다 — 코드 읽기, 이 경로를 실행으로 재측정하지는 않았다). 프로필 뿌리 읽기 금지(REQ-OWR-005)와 충돌하는 경로이므로 sync·리더가 알아야 한다.

**F4 거절 기록.** `pruneSkillDirTree` 가 살아 있는 형제 파일 옆의 빈 중첩 디렉터리를 두는 것은 REQ-OWR-008 이 스킬 디렉터리 자체가 비었을 때만 말하므로 SPEC 위반이 아니다. 레인 결정: 건드리지 않는다. F5/F6 은 공개 사항이며 할 일이 없다.

#### F1 — 수정 C `e0c2896d64` 와 증거

수정: `classifyOwnershipCore` 의 등록 안 된 분기에서 읽기 전에 `linkLeavesRoot(env.rootResolved, abs)` 를 부른다. 이 함수는 `Lstat` 으로 심볼릭 링크인지 보고, 링크이면 `EvalSymlinks` 로 끝까지 풀어 `withinRootBoundary` 로 뿌리 안인지 본다 — 대상을 열지 않는다. 뿌리 밖이면 class `unregistered`, 이유 `symlink-escape`, `v.data` 없음. 링크가 아니거나, 끊어졌거나, 조사할 수 없으면 종전대로 (끊어진 링크는 `read-failed`, 프로젝트 안 링크는 대상 바이트로 판정, FIFO 대상은 `non-regular-file`).

연속 5회 (`go test ./internal/cli/ -run '^TestUnregisteredSymlinkOutsideProjectIsNotRead$' -count=1 -v`, 저장 파일 `f1-green-1..5.txt`):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli	2.585s | PASS=1 FAIL=0 SKIP=0
run2: ok  	github.com/modu-ai/moai-adk/internal/cli	2.346s | PASS=1 FAIL=0 SKIP=0
run3: ok  	github.com/modu-ai/moai-adk/internal/cli	4.606s | PASS=1 FAIL=0 SKIP=0
run4: ok  	github.com/modu-ai/moai-adk/internal/cli	8.758s | PASS=1 FAIL=0 SKIP=0
run5: ok  	github.com/modu-ai/moai-adk/internal/cli	5.801s | PASS=1 FAIL=0 SKIP=0
```

차분 카나리 (감사가 쓴 overlay 시험 `TestT1865AuditUnregisteredProfileSymlink` 를 내 저장 위치로 복사해 `-overlay` 로 실행, 트리에 쓰지 않음; 수정 전 같은 시험은 `external-profile-bytes-read=true` 로 붉었다 — 감사 보고 인용):

```text
=== RUN   TestT1865AuditUnregisteredProfileSymlink
    zz_adversarial_review_test.go:42: gate=false class=unregistered reason=symlink-escape external-profile-bytes-read=false
--- PASS: TestT1865AuditUnregisteredProfileSymlink (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.629s
rc=0
```

실제 코드 변이 (`linkLeavesRoot` 가 뿌리 안 여부와 무관하게 모든 심볼릭 링크를 `true` 로 돌려줌; 그 뒤 `cmp` 로 복원 확인 `RESTORED-IDENTICAL`):

```text
    migrate_ownership_unregsymlink_test.go:118: switch=false in-project symlink: class="unregistered" keep="symlink-escape" data="", want unregistered/"no-provenance-record"/the target's bytes
    migrate_ownership_unregsymlink_test.go:118: switch=true in-project symlink: class="unregistered" keep="symlink-escape" data="", want unregistered/"no-provenance-record"/the target's bytes
    migrate_ownership_unregsymlink_test.go:167: .agents/skills/moai-unreg-symlink/to-inside: row {Rel:.agents/skills/moai-unreg-symlink/to-inside Class:unregistered Reason:symlink-escape Missing:[] Counterpart:n/a ProfileSide:n/a CaseVariantOf:} (found=true), want reason "no-provenance-record"
--- FAIL: TestUnregisteredSymlinkOutsideProjectIsNotRead (1.73s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.332s
FAIL
rc=1
RESTORED-IDENTICAL
```

붉은 이유: 올바른 이유다 — 프로젝트 안 링크 통제 단언이 붉어진다. 보호 장치 제거 변이는 위 RED 셀 자체다 (A 의 같은 시험이 수정 전 코드에서 붉었다).

이웃 시험 (같은 파일 시스템 영역, 저장 파일 `f1-neighbours.txt`):

```text
--- PASS: TestNonRegularFilesAreKeptWithoutBeingRead (0.76s)
--- PASS: TestNoProfileRootAccessInUpdateSources (9.09s)
--- PASS: TestProfileScanCatchesMutants (6.50s)
--- PASS: TestProfileEnvDoesNotReachProfileRoot (8.19s)
--- PASS: TestOwnershipVerdictSymlinkEscape (0.00s)
--- PASS: TestOwnershipVerdictClasses (0.00s)
--- PASS: TestKeepReasonVocabularyIsClosed (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	25.797s
rc=0
```

#### F3 — 수정 C `d8fa645c41` 와 증거

수정: `userCounterpartConfirmed` 가 사용자 루트 대응물을 읽기 전에 `unreadableKind(counterpart)` 를 묻는다. FIFO·소켓·장치 (그리고 그것을 가리키는 링크) 는 열리지 않고 "확인 아님" 이 된다 — 프로필 매핑 프로젝트의 상태 기록과 일반 프로젝트의 확인 질문이 같은 함수를 거치므로 두 순서가 모두 막힌다. 반환 이유와 고정된 어휘는 그대로 (`counterpart-unconfirmed`, `profile-reach`).

연속 5회 (`-run '^TestCounterpartFifoIsNotOpened$'`, 서브 시험 3개 + 부모, 저장 파일 `f3-green-1..5.txt`):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli	7.009s | PASS=4 FAIL=0 SKIP=0
run2: ok  	github.com/modu-ai/moai-adk/internal/cli	3.581s | PASS=4 FAIL=0 SKIP=0
run3: ok  	github.com/modu-ai/moai-adk/internal/cli	3.956s | PASS=4 FAIL=0 SKIP=0
run4: ok  	github.com/modu-ai/moai-adk/internal/cli	3.935s | PASS=4 FAIL=0 SKIP=0
run5: ok  	github.com/modu-ai/moai-adk/internal/cli	3.037s | PASS=4 FAIL=0 SKIP=0
```

실제 코드 변이 (`!unreadableKind(counterpart)` 로 뒤집음 — 정상 대응물이 확인되지 않음; 복원 `cmp` 확인):

```text
--- FAIL: TestNonRegularFilesAreKeptWithoutBeingRead (0.71s)
--- PASS: TestOwnershipVerdictClasses (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.025s
FAIL
rc=1
RESTORED-IDENTICAL
```

차분 카나리: 감사의 `TestT1865AuditProfileKeepOpensCounterpartFIFO` 는 수정 전 `blocked-until-FIFO-closed=true` 였고 수정 뒤 아래 overlay 실행에서 통과한다 (저장 파일 `overlay-all.txt`):

```text
=== RUN   TestT1865AuditSameSecondPriorGeneration
    zz_adversarial_review_test.go:28: could not place both calls in same second
--- FAIL: TestT1865AuditSameSecondPriorGeneration (0.12s)
=== RUN   TestT1865AuditUnregisteredProfileSymlink
    zz_adversarial_review_test.go:42: gate=false class=unregistered reason=symlink-escape external-profile-bytes-read=false
--- PASS: TestT1865AuditUnregisteredProfileSymlink (0.00s)
=== RUN   TestT1865AuditProfileKeepOpensCounterpartFIFO
--- PASS: TestT1865AuditProfileKeepOpensCounterpartFIFO (0.65s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.986s
FAIL
rc=1
```

(같은 실행의 `TestT1865AuditSameSecondPriorGeneration` 은 이제 "같은 실행 디렉터리에 두 호출을 놓을 수 없다" 며 실패한다 — 그 시험의 전제(`d1 == d2`)가 수정으로 더 이상 성립하지 않는다는 뜻이며, F2 의 결정적 시험이 대신 그 성질을 고정한다.)

#### F2 — 수정 C `fa26fcaf17` 와 증거

수정은 위 「선택한 가장 작은 안전한 수정」 그대로다 (`claimPriorManifestDir`, `copyRunMarker`; `update_prior_manifest.go` 한 파일, +60 줄). 보존 정책 (`backup.PruneBackups`)·이름 정규식·다른 파서는 바꾸지 않았다.

연속 5회 (`-run '^TestPriorManifestOfAnEarlierRunSurvivesALaterRun$'`, 서브 시험 6개 + 부모, 저장 파일 `f2-green-1..5.txt`):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli	1.546s | PASS=7 FAIL=0 SKIP=0
run2: ok  	github.com/modu-ai/moai-adk/internal/cli	1.402s | PASS=7 FAIL=0 SKIP=0
run3: ok  	github.com/modu-ai/moai-adk/internal/cli	1.161s | PASS=7 FAIL=0 SKIP=0
run4: ok  	github.com/modu-ai/moai-adk/internal/cli	1.256s | PASS=7 FAIL=0 SKIP=0
run5: ok  	github.com/modu-ai/moai-adk/internal/cli	1.294s | PASS=7 FAIL=0 SKIP=0
```

실제 코드 변이 (형제 디렉터리에 실행 표지를 복사하지 않음; 복원 `cmp` 확인):

```text
    update_prior_manifest_collision_test.go:101: the sibling's run marker is "" (open /private/tmp/own-fixture-d1/TestPriorManifestOfAnEarlierRunSurvivesALaterRunTwoRunsSharingT110383543/001/.moai-bac
    update_prior_manifest_collision_test.go:186: the sibling counts as 1 unmanaged entries
--- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun (0.14s)
    --- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/TwoRunsSharingTheConfigBackupPath (0.03s)
    --- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/ThreeRunsKeepNameOrderEqualToAgeOrderAndTheReaderSeesEveryEarlierOne (0.03s)
    --- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/ADirectoryTheRunDidNotCreateIsNeverWrittenInto (0.02s)
    --- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/NoFreeNameFailsClosedAndOverwritesNothing (0.04s)
    --- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/RunsWithoutAConfigBackupPathNeverShareACopy (0.02s)
    --- FAIL: TestPriorManifestOfAnEarlierRunSurvivesALaterRun/RetentionStillCountsTheSiblingWithItsRun (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.349s
FAIL
rc=1
RESTORED-IDENTICAL
```

붉은 이유: 올바른 이유다 — 형제의 실행 표지가 비고, 보존 정책이 형제를 비관리 항목 1건으로 센다 (보존 정책 중립성을 잃음). 보호 장치 제거 변이는 RED 셀 자체다.

AC-OWR-17/-18 의 흐름 시험: 공유 디렉터리 모사 가지를 무대기 연속 실행 가지로 바꾼 뒤의 결과는 아래 회귀 한 번에 포함된다 (`NoWaitConsecutiveRuns` 가지가 양 시험의 모든 픽스처·사례에서 통과). 이 가지는 두 실행이 같은 초에 드는지가 기계에 달려 있어 흐름 수준에서는 충돌 경로를 결정적으로 밟지 못한다 — 결정적 충돌 증거는 `update_prior_manifest_collision_test.go` 가 같은 설정 백업 경로를 두 번 넘기는 것으로 가진다.

#### 회귀와 점검 (저장 파일 `reg-*.txt`, `v-*.txt`)

변경된 시험 파일 (`git diff --name-only cdd72a814..HEAD -- '*_test.go'`, 27개 시험 파일) 의 최상위 시험 이름 집합을 패키지별로 도출했다 (`names-*.txt`): `internal/cli` 53, `internal/cli/update/backup` 4, `internal/manifest` 1. 앵커 선택자로 실행:

```text
internal/cli: ok  	github.com/modu-ai/moai-adk/internal/cli	123.208s rc=0  top-level PASS=53 FAIL=0 SKIP=0 (names=53)
internal_cli_update_backup: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.448s top-level PASS=4 FAIL=0 SKIP=0
internal_manifest: ok  	github.com/modu-ai/moai-adk/internal/manifest	0.259s top-level PASS=1 FAIL=0 SKIP=0
--- v-build:
rc=0
--- v-winbuild:
rc=0
--- v-vet (host):
rc=0
--- v-winvet:
rc=0
--- v-lint (golangci-lint run --new-from-rev=dfd7c40fb2 ./internal/cli/):
0 issues.
rc=0
--- invariants:
$ git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
invariant-exit=0
$ grep -rn "removalEnabled *=" internal --include=*.go (non-test)
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
$ gofmt -l internal/cli
gofmt-exit=0
$ git ls-files -o --exclude-standard
untracked-count=       0
$ git status --short
status-count=       0
$ git diff --name-only dfd7c40fb2..HEAD
.moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/progress.md
```

(일반 앵커 선택자의 최상위 PASS 수에는 서브 시험이 세어지지 않는다. 53/53, 4/4, 1/1 이다.)

**이번 chunk 의 한계 (정직한 회계).**

1. 회귀는 변경된 시험 파일의 이름 집합이며 `internal/cli` 패키지 전체 시험이 아니다. 패키지 전체 시험·커버리지·`-race`·Windows 시험 실행은 이전 chunk 와 같이 CI 와 착지 재측정 몫이다 (빌드와 vet 만 Windows 에서 했다).
2. F3 수정은 종류 확인과 읽기 사이의 경합 (그 사이에 일반 파일이 FIFO 로 바뀜) 을 막지 않는다 — 프로젝트 쪽 FIFO 에 대한 chunk 6 의 같은 공개 한계다.
3. F1 수정은 등록 안 된 경로만 고친다. 등록 경로의 사전 읽기는 위 결정대로 열린 부채다.
4. F2: 설정 백업 경로 자체의 같은 초 공유 (설정 파일 덮어쓰기, 실행 표지 교체) 는 이 카드 이전부터의 동작이고 건드리지 않았다. 형제 디렉터리는 복원 원본이 아니다 (`Restore Settings` 단계가 `configBackupPath` 를 쓴다 — 코드 읽기); 그 실행의 `removed-assets/` 는 형제에 있다. 설정 디렉터리가 없고 실행 표지도 없는 프로젝트에서 충돌이 나면 형제는 실행 표지가 없어 보존 정책이 비관리 항목으로 센다 (지우지 않음, 보고만).
5. 흐름 수준의 무대기 가지는 충돌 경로를 결정적으로 밟지 않는다 (위). 결정적 증거는 단위 수준이다.
6. 코덱스 게이트 재심사는 레인의 몫이다. 이 chunk 는 `.moai/reports/t1865/sync-audit.md`·`.moai/state/audit-receipts`·§E.4·CHANGELOG·SPEC 본문을 건드리지 않았다.

### Chunk 10 (card-review round 4) — 기록 시작

카드 리뷰 (advisory codex, HEAD `b5ffc48971` = `fa26fcaf17` 위의 문서 커밋) 가 P2 셋을 냈고 레인이 코드에서 확인했다. (A) 와 (B) 는 받아들여 시험 먼저 고치고, (C) 는 레인이 거절해 기록만 한다. 각 지적마다 시험 (RED) → 이 기록 → 수정 순서이며 순서의 증거는 커밋 그래프이다.

#### (A) 심볼릭 링크로 된 상위 디렉터리 — 등록 안 된 파일이 프로젝트 밖에서 읽힌다 (REQ-OWR-005)

chunk 9 의 `linkLeavesRoot` 는 경로의 **마지막 요소**가 링크일 때만 막았다. 프로젝트의 `.codex` (또는 `.claude`, `.agents`) 가 외부 디렉터리를 가리키면 그 아래 일반 파일은 이 검사를 통과해 `classifyUnregistered` 가 외부 파일을 읽었다.

**걷기 동작 (코드 읽기 + 시험으로 확인).** `filepath.WalkDir(projectRoot/<root>)` 는 걷기 뿌리에 `os.Lstat` 을 쓰며 이는 **상위** 링크는 따라가고 마지막 요소만 링크로 보고한다. 그래서 (i) 상위가 링크이면 걷기는 외부 `agents/moai` 를 내려가며 그 아래 모든 파일을 프로젝트 상대 이름으로 내놓는다 — 가드가 각 파일에서 전체 경로를 풀어야 모두 덮는다; (ii) 걷기 뿌리 자신이 링크이면 내려가지 않고 링크 하나를 비디렉터리 항목으로 내놓는다 (종전 마지막 요소 가드가 이미 덮던 형태; 시험이 대조군으로 둔다). 등록된 파일 경로가 이미 의존하는 `withinRootBoundary` 사용처 (`disposition`, 빈 디렉터리 정리 `migrate_ownership_emptydir.go:152`, `migrate_project_assets.go` 의 정리 경계) 는 모두 `EvalSymlinks` 로 **전체 경로**를 풀어 비교하므로 이미 전체 경로 해석이다 — 바꿀 곳이 없다. 등록 경로의 읽기가 경계 검사보다 앞서는 문제는 chunk 9 에서 기록한 N16 계열 부채 그대로이며, 상위 링크 형태에도 똑같이 적용된다 (AC-OWR-01 의 고정 기대 때문에 그대로 둔다).

**A (시험)** `780f7862f2` — `internal/cli/migrate_ownership_ancestorlink_test.go` (`//go:build !windows`, `TestUnregisteredFileBelowASymlinkedAncestorIsNotRead`). 판정 함수 절 (프로필 모양 내용을 가진 외부 디렉터리를 `.codex` 로 연결; 최상위와 한 단계 아래 파일; 제거 스위치 끔·켬), 대조군 절 (프로젝트 안을 가리키는 상위 링크는 내용으로 판정, 일반 파일, 끊어진 상위 링크 아래, 없는 파일), 흐름 절 둘 (상위가 링크인 걷기, 걷기 뿌리 자신이 링크인 걷기; 같은 걷기가 `migIdenticalSkill` 을 제거해 도달을 확인).

RED-now (명령 · 트리 SHA `780f7862f2` · 종료 코드 · 원문):

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestUnregisteredFileBelowASymlinkedAncestorIsNotRead$' -count=1 -v
=== RUN   TestUnregisteredFileBelowASymlinkedAncestorIsNotRead
=== RUN   TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/AncestorLeavesTheProject
    migrate_ownership_ancestorlink_test.go:93: switch=false .codex/agents/moai/profile-agent.md: the verdict function read 34 bytes from outside the project
    migrate_ownership_ancestorlink_test.go:96: switch=false .codex/agents/moai/profile-agent.md: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
    migrate_ownership_ancestorlink_test.go:93: switch=false .codex/agents/moai/sub/deep.md: the verdict function read 41 bytes from outside the project
    migrate_ownership_ancestorlink_test.go:96: switch=false .codex/agents/moai/sub/deep.md: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
    migrate_ownership_ancestorlink_test.go:93: switch=true .codex/agents/moai/profile-agent.md: the verdict function read 34 bytes from outside the project
    migrate_ownership_ancestorlink_test.go:96: switch=true .codex/agents/moai/profile-agent.md: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
    migrate_ownership_ancestorlink_test.go:93: switch=true .codex/agents/moai/sub/deep.md: the verdict function read 41 bytes from outside the project
    migrate_ownership_ancestorlink_test.go:96: switch=true .codex/agents/moai/sub/deep.md: class="unregistered" keep="no-provenance-record" eligible=false, want unregistered/"symlink-escape"/false
=== RUN   TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/ControlsAreJudgedAsBefore
=== RUN   TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/Flow/AncestorOfTheWalkRoot
    migrate_ownership_ancestorlink_test.go:185: .codex/agents/moai/profile-agent.md: row class="unregistered" reason="no-provenance-record", want unregistered/"symlink-escape"
    migrate_ownership_ancestorlink_test.go:190: .codex/agents/moai/profile-agent.md: the verdict carried 34 bytes read from outside the project
    migrate_ownership_ancestorlink_test.go:185: .codex/agents/moai/sub/deep.md: row class="unregistered" reason="no-provenance-record", want unregistered/"symlink-escape"
    migrate_ownership_ancestorlink_test.go:190: .codex/agents/moai/sub/deep.md: the verdict carried 41 bytes read from outside the project
    migrate_ownership_ancestorlink_test.go:195: .codex/agents/moai/profile-agent.md carried 34 bytes
    migrate_ownership_ancestorlink_test.go:195: .codex/agents/moai/sub/deep.md carried 41 bytes
=== RUN   TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/Flow/TheWalkRootItself
--- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead (1.99s)
    --- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/AncestorLeavesTheProject (0.57s)
    --- PASS: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/ControlsAreJudgedAsBefore (0.00s)
    --- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/Flow/AncestorOfTheWalkRoot (0.73s)
    --- PASS: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/Flow/TheWalkRootItself (0.69s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.115s
FAIL
rc=1
```

붉은 이유: 올바른 이유다. 상위가 링크인 두 형태 (판정 함수, 걷기) 모두에서 외부 파일의 34 와 41 바이트를 읽었고 이유가 `no-provenance-record` 로 나왔다 (`symlink-escape` 가 맞다). 대조군 (프로젝트 안 상위 링크, 일반 파일, 끊어진 상위 링크, 없는 파일)과 걷기 뿌리 자신이 링크인 형태는 수정 전에도 통과한다 — 바뀌면 안 되는 것의 고정이다.

#### (B) 실패한 이름 바꾸기가 남긴 임시 파일 `manifest.prior.json.tmp-*`

`writePriorManifest` 는 이름 바꾸기가 실패하면 임시 파일을 지우지 않고 돌아갔다. 읽는 쪽이 그 이름을 무시한다는 주석은 맞지만, `RestoreMoaiConfigLegacy` (그리고 섹션이 없는 백업의 `RestoreMoaiConfigRetained`, `RestoreFromBackupDir`) 가 백업 루트의 모든 파일을 `.moai/config` 로 복사해 임시 파일이 프로젝트에 들어간다.

**B (시험)** `9cc9b77a01` — 두 끝을 모두 고정한다. `internal/cli/update_prior_manifest_tmpcleanup_test.go` (`TestFailedPriorManifestRenameLeavesNoTemporaryFile`: 쓰는 쪽 임시 이름이 리터럴 `manifest.prior.json.tmp-` 로 시작함; 주입한 이름 바꾸기 실패와 실제 이름 바꾸기 실패 (자리를 차지한 디렉터리) 뒤에 임시 파일이 없고 오류가 종전 그대로 (`rename into place`, 주입 오류를 감쌈) 보고되며 최종 이름이 없음; 백업 단계 인수인계 — 실패한 상태에서 마이그레이션이 파일을 하나도 제거하지 않고, 같은 픽스처의 주입 없는 통제 실행은 `migIdenticalSkill` 을 제거함) 와 `internal/cli/update/backup/restore_prior_manifest_tmp_test.go` (복원 네 진입점; 백업 루트의 임시 파일은 복사되지 않고, 하위 디렉터리의 같은 이름 파일은 사용자 내용으로 복원됨; 이름은 리터럴). 기존 시험 `TestPriorManifestAtomicWriteFailures` 의 (q) 는 이름 바꾸기 뒤 임시 파일이 **남아 있다**는 것을 전제로 반환 후에 그 파일을 읽었다 — 그 전제는 이 수정이 의도적으로 바꾸는 동작이므로, 전제 (이름 바꾸기 시점에 임시 파일이 완전한 바이트를 가짐) 를 이름 바꾸기 이중 안에서 읽도록 옮겼다. 단언 (최종 이름 없음, 세대 없음, 오류 반환) 은 그대로다.

RED-now (트리 SHA `9cc9b77a01`, 종료 코드, 원문; 저장 파일 `b-red-cli.txt`, `b-red-backup.txt`):

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestFailedPriorManifestRenameLeavesNoTemporaryFile|TestPriorManifestAtomicWriteFailures)$' -count=1 -v
--- PASS: TestPriorManifestAtomicWriteFailures (0.05s)
    --- PASS: TestPriorManifestAtomicWriteFailures/o-normal-run-goes-through-temp-then-rename (0.02s)
    --- PASS: TestPriorManifestAtomicWriteFailures/p-failure-during-the-temporary-write (0.01s)
    --- PASS: TestPriorManifestAtomicWriteFailures/q-failure-between-the-temporary-write-and-the-rename (0.01s)
    --- PASS: TestPriorManifestAtomicWriteFailures/r-real-rename-failure-onto-an-occupied-name (0.02s)
    update_prior_manifest_tmpcleanup_test.go:74: temporary files remain after the failed rename: [manifest.prior.json.tmp-881594078]
    update_prior_manifest_tmpcleanup_test.go:94: temporary files remain after the failed rename: [manifest.prior.json.tmp-1722304425]
    update_prior_manifest_tmpcleanup_test.go:136: temporary files remain in the run backup: [manifest.prior.json.tmp-3588344372]
--- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile (0.85s)
    --- PASS: TestFailedPriorManifestRenameLeavesNoTemporaryFile/TheNameTheWriterUsesIsTheNameTheRestoreSkips (0.01s)
    --- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile/InjectedRenameFailure (0.00s)
    --- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile/RealRenameFailureOntoAnOccupiedName (0.01s)
    --- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile/TheRunStillRemovesNothing (0.84s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.858s
FAIL
rc=1
```

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/update/backup/ -run '^TestRestore(Legacy|RetainedWithoutSections|RetainedWithSections|FromBackupDir)SkipsPriorManifestTemporaryFile$' -count=1 -v
    restore_prior_manifest_tmp_test.go:58: manifest.prior.json.tmp-4242 from the backup root was copied into the config tree (stat err: <nil>)
--- FAIL: TestRestoreLegacySkipsPriorManifestTemporaryFile (0.00s)
    restore_prior_manifest_tmp_test.go:66: manifest.prior.json.tmp-4242 from the backup root was copied into the config tree (stat err: <nil>)
--- FAIL: TestRestoreRetainedWithoutSectionsSkipsPriorManifestTemporaryFile (0.00s)
--- PASS: TestRestoreRetainedWithSectionsSkipsPriorManifestTemporaryFile (0.00s)
Warning: merge failed for notes/manifest.prior.json, restoring backup
Warning: merge failed for notes/manifest.prior.json.tmp-4242, restoring backup
    restore_prior_manifest_tmp_test.go:87: manifest.prior.json.tmp-4242 from the backup root was copied into the config tree (stat err: <nil>)
--- FAIL: TestRestoreFromBackupDirSkipsPriorManifestTemporaryFile (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.337s
FAIL
rc=1
```

붉은 이유: 올바른 이유다. 실패한 이름 바꾸기 뒤 임시 파일이 남았고 (주입, 실제, 인수인계 세 곳), 복원 세 진입점이 백업 루트의 임시 파일을 `.moai/config` 로 복사했다. 섹션이 있는 백업의 `RetainedWithSections` 는 섹션 YAML 만 걸으므로 수정 전에도 통과하는 특성 기록이다 (chunk 5 와 같은 구성).

#### (C) 거절 기록 — 리뷰의 둘째 P2

리뷰는 레거시 백업 루트에 사용자가 글자 그대로 `manifest.prior.json` 이라는 설정 파일을 두었다면 이제 `--restore` 가 그것을 건너뛴다고 지적했다. 레인이 거절했다: (1) 그런 파일이 `.moai/config` 에 존재할 수 있는 유일한 경로는 chunk 5 가 고친 복원 결함 그 자체인데 그 결함은 출하된 적이 없다; (2) 이름은 REQ-OWR-015 와 design.md D-3 이 고정한다; (3) 복원 걷기는 그 둘을 구별할 방법이 없다. **공개된 한계**: 백업 루트의 이름이 `manifest.prior.json` (또는 `manifest.prior.json.tmp-` 접두사) 인 파일은 복원에서 제외된다. 하위 디렉터리의 같은 이름 파일은 계속 복원된다 (시험이 고정).

#### Chunk 10 GREEN 기록과 정직한 회계

#### (A) 수정 C `848bbc4707` 와 증거

수정: `linkLeavesRoot` 를 `pathLeavesRoot` 로 바꿨다. `Lstat` 지름길 (마지막 요소가 링크인지) 을 없애고 `EvalSymlinks(abs)` 로 전체 경로를 풀어 `withinRootBoundary(env.rootResolved, resolved)` 로 비교한다 — 상위 디렉터리가 링크여도 같다. 풀리지 않는 경로 (끊어진 링크, 없는 파일, 조사할 수 없는 경로)는 종전대로 보고하지 않아 `read-failed` 로 남는다. 프로젝트 안을 가리키는 링크 (파일이든 상위 디렉터리든) 는 종전대로 내용으로 판정한다. 등록된 파일 경로의 `withinRootBoundary` 사용처 (`disposition`, 빈 디렉터리 정리, 정리 경계) 는 이미 전체 경로 해석이라 바꾸지 않았다.

**확인한 것 (A).** (1) 걷기가 상위 링크 아래 파일을 어떻게 내놓는가 — 시험이 두 형태를 구동한다: 상위 (`.codex`) 가 링크이면 걷기는 외부 `agents/moai` 를 내려가며 모든 파일 (최상위와 한 단계 아래 둘 다)을 프로젝트 상대 이름으로 내놓고, 가드가 각 파일에서 전체 경로를 풀어 둘 다 덮는다; 걷기 뿌리 자신이 링크이면 내려가지 않고 링크 하나를 내놓아 가드가 그 링크를 덮는다. 즉 걷기가 내놓는 모든 파일이 가드를 거친다. (2) 대조군 — 프로젝트 안 상위 링크, 일반 파일, 끊어진 상위 링크 아래, 없는 파일 — 이 수정 전후 같은 결과. (3) 이웃 시험 53개 최상위 전부 통과 (`a-neighbours.txt`: PASS=53 FAIL=0 SKIP=0), `rootResolved` 가 빈 환경으로 등록 안 된 일반 파일을 분류하는 시험이 이 집합 안에는 없음을 이것으로 확인했다 (있었다면 모든 파일이 뿌리 밖으로 읽혀 붉어졌을 것이다; 집합 밖의 패키지 나머지는 CI 몫).

연속 5회 (`-run '^TestUnregisteredFileBelowASymlinkedAncestorIsNotRead$'`, 서브 시험 4개 + 부모, 저장 파일 `a-green-1..5.txt`):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli	3.471s | PASS=5 FAIL=0 SKIP=0
run2: ok  	github.com/modu-ai/moai-adk/internal/cli	2.156s | PASS=5 FAIL=0 SKIP=0
run3: ok  	github.com/modu-ai/moai-adk/internal/cli	2.045s | PASS=5 FAIL=0 SKIP=0
run4: ok  	github.com/modu-ai/moai-adk/internal/cli	2.511s | PASS=5 FAIL=0 SKIP=0
run5: ok  	github.com/modu-ai/moai-adk/internal/cli	2.121s | PASS=5 FAIL=0 SKIP=0
```

실제 코드 변이 (`pathLeavesRoot` 를 마지막 요소 검사로 되돌림: `Lstat` 이 링크가 아니면 `false`; `cmp` 로 복원 확인):

```text
--- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead (1.50s)
    --- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/AncestorLeavesTheProject (0.52s)
    --- PASS: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/ControlsAreJudgedAsBefore (0.00s)
    --- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/Flow/AncestorOfTheWalkRoot (0.51s)
    --- PASS: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/Flow/TheWalkRootItself (0.46s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.336s
FAIL
rc=1
RESTORED-IDENTICAL
```

붉은 이유: 올바른 이유다 — 상위 링크 두 형태 중 상위가 링크인 판정 함수 절과 걷기 절이 붉어지고 대조군과 걷기 뿌리 자신이 링크인 형태는 초록으로 남는다.

#### (B) 수정 C `d8a5f75053` 와 증거

수정 (두 끝): (1) `writePriorManifest` 가 임시 파일 생성 뒤 어느 실패 경로에서도 (임시 쓰기가 경로와 함께 오류를 돌려준 경우, 이름 바꾸기 실패) 임시 파일을 지운다 (최선 노력; 오류는 종전 그대로 `preserve prior manifest: ...` 로 보고). (2) 백업 패키지가 `PriorManifestTempPrefix` (`manifest.prior.json.tmp-`) 와 `IsPriorManifestBookkeeping(relPath)` 를 내보내고, 쓰는 쪽 `os.CreateTemp` 이 그 접두사를 쓰며 레거시 복원 걷기가 최종 이름과 같이 백업 루트의 임시 이름을 건너뛴다. 하위 디렉터리의 같은 이름 파일은 계속 복원된다.

연속 5회 (저장 파일 `b-green-cli-1..5.txt`: `-run '^(TestFailedPriorManifestRenameLeavesNoTemporaryFile|TestPriorManifestAtomicWriteFailures)$'`, 서브 시험 포함 PASS 10; `b-green-backup-1..5.txt`: 복원 네 진입점, PASS 4):

```text
run1: cli ok  	github.com/modu-ai/moai-adk/internal/cli	2.363s PASS=10 FAIL=0 SKIP=0 | backup ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.371s PASS=4 FAIL=0 SKIP=0
run2: cli ok  	github.com/modu-ai/moai-adk/internal/cli	2.653s PASS=10 FAIL=0 SKIP=0 | backup ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.205s PASS=4 FAIL=0 SKIP=0
run3: cli ok  	github.com/modu-ai/moai-adk/internal/cli	2.521s PASS=10 FAIL=0 SKIP=0 | backup ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.263s PASS=4 FAIL=0 SKIP=0
run4: cli ok  	github.com/modu-ai/moai-adk/internal/cli	2.339s PASS=10 FAIL=0 SKIP=0 | backup ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.310s PASS=4 FAIL=0 SKIP=0
run5: cli ok  	github.com/modu-ai/moai-adk/internal/cli	2.805s PASS=10 FAIL=0 SKIP=0 | backup ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.196s PASS=4 FAIL=0 SKIP=0
```

실제 코드 변이 둘 (각각 `cmp` 로 복원 확인):

```text
-- 변이 1: 이름 바꾸기 실패 뒤 임시 파일을 지우지 않음
    update_prior_manifest_tmpcleanup_test.go:74: temporary files remain after the failed rename: [manifest.prior.json.tmp-3412543388]
    update_prior_manifest_tmpcleanup_test.go:94: temporary files remain after the failed rename: [manifest.prior.json.tmp-902120179]
    update_prior_manifest_tmpcleanup_test.go:136: temporary files remain in the run backup: [manifest.prior.json.tmp-3431824611]
--- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile (0.96s)
    --- PASS: TestFailedPriorManifestRenameLeavesNoTemporaryFile/TheNameTheWriterUsesIsTheNameTheRestoreSkips (0.01s)
    --- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile/InjectedRenameFailure (0.01s)
    --- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile/RealRenameFailureOntoAnOccupiedName (0.01s)
    --- FAIL: TestFailedPriorManifestRenameLeavesNoTemporaryFile/TheRunStillRemovesNothing (0.93s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.926s
FAIL
rc=1
RESTORED-IDENTICAL-writer
-- 변이 2: 복원이 최종 이름만 건너뜀 (임시 접두사 미포함)
--- FAIL: TestRestoreLegacySkipsPriorManifestTemporaryFile (0.00s)
--- FAIL: TestRestoreRetainedWithoutSectionsSkipsPriorManifestTemporaryFile (0.00s)
--- PASS: TestRestoreRetainedWithSectionsSkipsPriorManifestTemporaryFile (0.00s)
--- FAIL: TestRestoreFromBackupDirSkipsPriorManifestTemporaryFile (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.365s
FAIL
rc=1
RESTORED-IDENTICAL-restore
```

붉은 이유: 올바른 이유다 — 변이 1 은 임시 파일이 남는 세 곳이, 변이 2 는 복원 세 진입점이 붉어진다 (섹션이 있는 백업의 한 시험은 섹션 YAML 만 걸으므로 변이와 무관하게 초록이다).

#### 회귀와 점검 (저장 파일 `reg-*.txt`, `v-*.txt`)

변경된 시험 파일 (`git diff --name-only cdd72a814..HEAD -- '*_test.go'`, 30개) 의 최상위 시험 이름 집합을 패키지별로 도출했다 (`names-*.txt`): `internal/cli` 55, `internal/cli/update/backup` 8, `internal/manifest` 1. 앵커 선택자로 실행:

```text
reg-cli: ok  	github.com/modu-ai/moai-adk/internal/cli	138.382s rc=0  top-level PASS=55 FAIL=0 SKIP=0
reg-internal_manifest: ok  	github.com/modu-ai/moai-adk/internal/manifest	0.259s rc=0  top-level PASS=1 FAIL=0 SKIP=0
reg-internal_cli_update_backup: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.220s rc=0  top-level PASS=8 FAIL=0 SKIP=0
--- v-build:
rc=0
--- v-winbuild:
rc=0
--- v-vet (host, internal/cli + internal/cli/update/backup):
rc=0
--- v-winvet:
rc=0
--- v-lint (golangci-lint run --new-from-rev=c194aca4de ./internal/cli/ ./internal/cli/update/backup/):
0 issues.
rc=0
--- invariants:
$ git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
invariant-exit=0
$ removalEnabled assignments (non-test)
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
$ skip grep over the 30 changed test files
internal/cli/migrate_ownership_helpers_test.go:1008:	t.Logf("force=false run=1 skipped=%v", first.Skipped)
internal/cli/migrate_ownership_helpers_test.go:1009:	if first.Skipped {
internal/cli/migrate_ownership_symlink_test.go:8:// platform without symlinks; a t.Skip would sweep nothing and read as a pass. A
skip-grep-exit=0 (1 = no t.Skip/SkipNow/testing.Short call)
internal/cli/migrate_ownership_helpers_test.go:425:	// Skipped is the flow's skipped result: true when the version-match gate
internal/cli/migrate_ownership_helpers_test.go:427:	Skipped bool
internal/cli/migrate_ownership_helpers_test.go:467:	run := forcedRun{Out: out.String(), Skipped: skipped, StartManifest: start}
$ gofmt -l internal/cli
gofmt-exit=0
untracked=0
```

(스킵 grep 의 세 줄은 이전 chunk 와 같다: 실행 결과 필드 `first.Skipped` 를 쓰는 두 줄과 `t.Skip` 을 설명하는 주석 한 줄이며 `t.Skip`·`SkipNow`·`testing.Short` 호출은 없다.)

**이번 chunk 의 한계 (정직한 회계).**

1. 회귀는 변경된 시험 파일의 이름 집합이며 패키지 전체 시험이 아니다. 패키지 전체 시험·커버리지·`-race`·Windows 시험 실행은 CI 와 착지 재측정 몫이다 (Windows 는 빌드와 vet 만).
2. (A) 는 등록 안 된 경로만 고친다. 등록된 파일 경로의 읽기가 경계 검사보다 앞서는 부채는 기준 시점부터 있었고 AC-OWR-01 의 고정 기대 때문에 그대로이며, 이제 상위 링크 형태에도 같은 부채가 적용됨을 위에서 밝혔다 (N16 계열). 가드는 `EvalSymlinks` 와 읽기 사이의 경합 (그 사이에 링크가 바뀜) 을 막지 않는다 — 종류 확인과 읽기 사이 경합과 같은 공개 한계.
3. (A) 는 `rootResolved` 가 빈 환경에서 모든 존재하는 경로를 뿌리 밖으로 보고 (fail-closed) 한다. 운영 경로는 `newOwnershipEnv` 가 항상 채운다 (못 채우면 오류); 실행한 시험 집합의 어느 시험도 빈 값으로 등록 안 된 일반 파일을 분류하지 않았다 (통과로 확인; 집합 밖은 CI).
4. (B) 임시 파일 제거는 최선 노력이다. 프로세스가 이름 바꾸기 실패와 제거 사이에 죽으면 파일이 남을 수 있고 이 경우에도 복원이 그 이름을 건너뛴다 (두 끝이 함께 막는 이유).
5. (B) 기존 시험 (q) 의 전제 위치를 옮긴 것은 수정이 바꾸는 동작에 맞춘 것이다. 단언은 약화하지 않았다.
6. 코덱스 카드 리뷰의 재확인은 레인의 몫이다. 이 chunk 는 `.moai/reports`, `.moai/state/audit-receipts`, §E.4, CHANGELOG, SPEC 본문을 건드리지 않았다.

### Chunk 11 (delta-audit repair) — 기록 시작

두 번째 독립 sync 감사 (delta, HEAD `38c8ca5376`, 영수증 `rcpt-aeb48f1d567c915a06f7b90f`) 가 FAIL (필수 codex 백엔드만) 을 냈다. 감사는 round 1 의 F1/F2/F3 이 고쳐졌음을 확인했고 코덱스의 P2 둘이 감사의 N1, N2 이다. 근거 문서는 `.moai/reports/t1865/sync-audit.md` (이 chunk 는 고치지 않았다). 레인 요구: 두 지적을 시험 먼저 고치고, 같은 chunk 에서 최근 라운드의 작은 도우미들을 공격자의 눈으로 다시 읽어 결과를 이 기록에 남기며, 경로 경계 반례를 한 표로 고정한다. 각 지적마다 시험 (RED) → 이 기록 → 수정 순서이며 순서의 증거는 커밋 그래프이다.

#### 공격자 시점 자체 검토 (적대적 자체 검토) — 결과 표

"실제 문제" 로 판정된 것은 이 chunk 에서 시험 먼저 고쳤다. 판정 근거는 열 「확인 방법」 이다.

| 도우미 | 공격 시나리오 | 결과 | 확인 방법 |
|---|---|---|---|
| `IsPriorManifestBookkeeping` | 접두사 일치에 구분자 검사가 없음 — `manifest.prior.json.tmp-notes/keep.txt` | **실제 문제 N1** — 수정 | 감사 재현 + 표 행 2개 + 단위 표 3행 RED |
| 〃 | Windows 경로 (`\`) | 수정은 `filepath.ToSlash` 로 구분자를 본다. Windows 실행은 하지 않았다 (빌드·vet 만) | 코드 읽기 |
| 〃 | 대소문자 구분 없는 파일 시스템 | 쓰는 쪽이 소문자 이름만 만든다. 같은 이름의 다른 대소문자 파일은 대소문자 구분 FS 에서는 다른 이름 (복원됨), 구분 없는 FS 에서는 같은 파일이라 공존 불가 | 단위 표 행 (「another letter case is another name」) |
| 〃 | 이름에 접두사가 들어 있을 뿐인 경우, 빈 이름 | `HasPrefix` 만 쓰므로 앞에 붙는 경우만 일치. 빈 이름과 `…json.tmp` (대시 없음) 는 거짓 | 단위 표 행 |
| 〃 | 접두사와 똑같은 루트 파일 이름 (`manifest.prior.json.tmp-`) | 루트 수준이므로 건너뛴다. `CreateTemp` 는 항상 숫자를 붙이므로 이 이름은 사용자 파일과만 충돌한다 — 수용 | 단위 표 행 |
| `claimPriorManifestDir` | 접미사 이름이 파일·끊어진 심볼릭 링크로 이미 있음 | `Mkdir` 가 `EEXIST` 를 돌려주므로 건너뛴다. 결과: 문제 없음 | 시험 `SuffixNamesTakenByAFileOrADanglingLinkAreSkipped` (도착 시 초록) |
| 〃 | `-99` 상한 | 소진하면 오류, 아무것도 덮어쓰지 않음 | chunk 9 시험 (재실행 초록) |
| 〃 | 끝에 구분자가 붙은 실행 디렉터리 (`<dir>/`) | **실제 문제** — 형제가 디렉터리 안 `<dir>/-02` 로 만들어진다 (실행 디렉터리의 형제가 아님) — 수정 (`filepath.Clean`) | 시험 `ABackupPathWithATrailingSeparatorStillGetsASibling` RED |
| 〃 | 실행 디렉터리가 심볼릭 링크 | 형제는 링크 옆의 실제 디렉터리로 만들어지고 사본은 거기 들어간다. `.moai-backups` 는 프로젝트 자신의 영역이며 루트가 링크면 보존 정책이 거부한다 (`OldBackupRootRefused`) — 새 경로 없음 | 코드 읽기 |
| `copyRunMarker` | 표지 복사 실패가 무시됨 | **실제 문제 N2** — 수정 | 시험 RED (아래) |
| 〃 | 표지에 줄이 더 있거나 CRLF | 첫 줄만 취하고 `TrimSpace` 가 `\r` 을 지운 뒤 한 줄로 쓴다 — 보존 정책이 같은 실행으로 센다 | 시험 `MarkerWithExtraLinesAndCRLF…` (도착 시 초록) |
| 〃 | 표지의 첫 줄이 실행 id 로 해석되지 않음 | 그대로 복사한다. 원본도 이름으로 센 독립 실행이고 형제는 비관리 항목이 된다 — 새 위험은 아니나 **남는 한계로 공개** (수정하지 않음: 해석 규칙이 백업 패키지의 비공개 함수다) | 코드 읽기 |
| `pathLeavesRoot` | `EvalSymlinks` 의 오류 (ELOOP, ENOTDIR, 권한) 가 읽기로 이어지는가 | 오류면 보고하지 않고 읽기가 같은 경로 해석에서 같은 이유로 실패한다 → `read-failed`, 읽은 바이트 없음. ELOOP 와 ENOTDIR 는 표 행으로 고정. 권한 오류는 root 실행에서 재현할 수 없어 시험하지 않았다 (추론) | 표 행 2개 (도착 시 초록) |
| 〃 | 빈 `rootResolved` | 모든 존재하는 경로를 뿌리 밖으로 보아 읽지 않는다 (fail closed) — 표 행 2개로 고정 (등록·미등록) | 표 행 (도착 시 초록) |
| 〃 | 업데이트 명령의 실제 뿌리는 `"."` (`update_template_sync.go:230`) | **실제 문제** — 절대 경로 대상의 프로젝트 안 링크는 `Rel(".", 절대경로)` 가 실패해 뿌리 밖으로 읽혀 `symlink-escape` 로 바뀐다 (기준 커밋에서는 내용으로 판정). 안전한 쪽(읽지 않음)이지만 운영 동작이 바뀐다 — 수정 (뿌리와 경로를 절대·해석된 형태로 맞춤) | 표 행 3개 RED, 걷기 수준 행 RED |
| `withinRootBoundary` | 형제 `proj-xxx-evil`, `proj-xxx2` 가 문자열 접두사로 안쪽 취급 | 안 된다 — `filepath.Rel` 로 요소 단위 비교한다 | 표 행 4개 (도착 시 초록) |
| 〃 | 최상위 항목 이름이 `..` 로 시작 (`..hidden`) | **실제 문제 (과잉 거부)** — `HasPrefix(rel, "..")` 가 안쪽 이름을 밖으로 본다. 공통 자산 뿌리 아래는 항상 `.claude/…` 로 시작해 닿지 않지만 도우미의 계약 위반 — 수정 | 표 행 2개 RED |
| `unreadableKind`, 대응물 확인 | `Stat` 대 `Lstat`, 링크 루프 | `Stat` 은 링크를 따라가 FIFO 로 가는 링크를 잡는다. 루프·끊어진 링크는 `Stat` 오류 → 거짓 → 읽기가 실패한다 (블록하지 않음). 대응물 FIFO 는 `Stat` 으로 막힘 | 표 행 (링크→FIFO, 루프, 대응물 FIFO; 도착 시 초록) |
| `writePriorManifest` 정리 | 임시 이름 충돌, 오류 감싸기 | `CreateTemp` 가 `O_EXCL` + 난수라 덮어쓰지 않는다. `os.Remove` 는 비어 있지 않은 디렉터리를 지우지 못한다 (최선 노력). 오류는 `%w` 로 감싸 `errors.Is` 가 통한다 (시험이 단언) | chunk 10 시험 + 코드 읽기 |

더 큰 수정이 필요한 문제는 찾지 못했다 (블로커 없음). 종류 확인과 읽기, `EvalSymlinks` 와 읽기 사이의 경합은 이전 chunk 에서 공개한 한계 그대로다.

#### N1 — RED 셀 (A `ef45c82cc9`)

`internal/cli/update/backup/restore_prior_manifest_tmp_test.go` 를 확장했다: 접두사 이름의 **디렉터리** 안 사용자 파일 (`manifest.prior.json.tmp-notes/keep.txt`), 최종 이름에 접미사가 붙은 디렉터리 (`manifest.prior.json.d/x`), 접두사 디렉터리 아래의 정확한 최종 이름은 복원되어야 하고, 루트 임시 파일은 건너뛴다. `IsPriorManifestBookkeeping` 의 단위 표 시험 `TestIsPriorManifestBookkeeping` 을 더했다.

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/update/backup/ -run '^(TestRestore(Legacy|RetainedWithoutSections|RetainedWithSections|FromBackupDir)SkipsPriorManifestTemporaryFile|TestIsPriorManifestBookkeeping)$' -count=1 -v
    restore_prior_manifest_tmp_test.go:75: manifest.prior.json.tmp-notes/keep.txt was not restored as user content (err=open <tmp>/.moai/config/manifest.prior.json.tmp-notes/keep.txt: no such file or directory, got "")
    restore_prior_manifest_tmp_test.go:75: manifest.prior.json.tmp-notes/manifest.prior.json was not restored as user content (err=open <tmp>/.moai/config/manifest.prior.json.tmp-notes/manifest.prior.json: no such file or directory, got "")
--- FAIL: TestRestoreLegacySkipsPriorManifestTemporaryFile (0.01s)
    restore_prior_manifest_tmp_test.go:83: manifest.prior.json.tmp-notes/keep.txt was not restored as user content (err=open <tmp>/.moai/config/manifest.prior.json.tmp-notes/keep.txt: no such file or directory, got "")
    restore_prior_manifest_tmp_test.go:83: manifest.prior.json.tmp-notes/manifest.prior.json was not restored as user content (err=open <tmp>/.moai/config/manifest.prior.json.tmp-notes/manifest.prior.json: no such file or directory, got "")
--- FAIL: TestRestoreRetainedWithoutSectionsSkipsPriorManifestTemporaryFile (0.01s)
--- PASS: TestRestoreRetainedWithSectionsSkipsPriorManifestTemporaryFile (0.01s)
Warning: merge failed for manifest.prior.json.d/x, restoring backup
Warning: merge failed for notes/manifest.prior.json, restoring backup
Warning: merge failed for notes/manifest.prior.json.tmp-4242, restoring backup
    restore_prior_manifest_tmp_test.go:104: manifest.prior.json.tmp-notes/keep.txt was not restored as user content (err=open <tmp>/.moai/config/manifest.prior.json.tmp-notes/keep.txt: no such file or directory, got "")
    restore_prior_manifest_tmp_test.go:104: manifest.prior.json.tmp-notes/manifest.prior.json was not restored as user content (err=open <tmp>/.moai/config/manifest.prior.json.tmp-notes/manifest.prior.json: no such file or directory, got "")
--- FAIL: TestRestoreFromBackupDirSkipsPriorManifestTemporaryFile (0.01s)
    restore_prior_manifest_tmp_test.go:134: a file below a directory named like the temporary prefix: IsPriorManifestBookkeeping("manifest.prior.json.tmp-notes/keep.txt") = true, want false
    restore_prior_manifest_tmp_test.go:134: a directory named like the temporary prefix: IsPriorManifestBookkeeping("manifest.prior.json.tmp-notes/") = true, want false
    restore_prior_manifest_tmp_test.go:134: the final name below a prefixed directory: IsPriorManifestBookkeeping("manifest.prior.json.tmp-notes/manifest.prior.json") = true, want false
--- FAIL: TestIsPriorManifestBookkeeping (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.486s
FAIL
rc=1
```

붉은 이유: 올바른 이유다. 접두사 디렉터리 아래 두 파일이 복원되지 않았고 (`was not restored as user content`), 단위 표에서 구분자가 있는 세 이름이 `true` 로 판정되었다. 섹션이 있는 백업의 시험은 섹션 YAML 만 걷기 때문에 수정 전에도 통과한다 (특성 기록).

#### N2 — RED 셀 (A `2cd11e520e`)

`internal/cli/update_prior_manifest_markercopy_test.go` (`TestMarkerCopyFailureFailsClosedWithoutResidue`). 표지 복사에 주입 이음매 `priorManifestMarkerCopy` 를 더했다 (동작 중립: 오류는 수정 전까지 무시됨; 다른 `priorManifest*` 이음매와 같은 방식, 운영 코드는 바꾸지 않음). 표지 복사가 실패하면 만든 형제가 남지 않아야 하고, 상태는 `priorManifestFailed`, 그 실행은 파일을 하나도 제거하지 않아야 한다 (주입 없는 통제 실행은 `migIdenticalSkill` 을 제거). 복사할 표지가 아예 없으면 종전 동작 (형제는 표지 없이 만들어짐). 표지에 줄이 더 있거나 CRLF 이면 한 줄로 정리되어 복사되고 보존 정책이 같은 실행으로 센다. 끝에 구분자가 붙은 경로의 형제.

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestMarkerCopyFailureFailsClosedWithoutResidue$' -count=1 -v
    update_prior_manifest_markercopy_test.go:58: got ("<tmp>/.moai-backups/20261018_153045-02", 1, <nil>), want a failed state carrying the injected error
    update_prior_manifest_markercopy_test.go:101: state=1 err=<nil>, want a failed state carrying the injected error
    update_prior_manifest_markercopy_test.go:104: the run removed [.claude/skills/moai-foundation-core/SKILL.md] although its prior manifest copy could not be placed
    update_prior_manifest_markercopy_test.go:147: the sibling is <tmp>/.moai-backups/20261018_153045/-02, want <tmp>/.moai-backups/20261018_153045-02 (a sibling of the run directory, not a directory inside it)
--- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue (1.61s)
    --- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue/NoSiblingIsLeftAndTheEarlierCopyIsUntouched (0.02s)
    --- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue/TheRunRemovesNothing (1.54s)
    --- PASS: TestMarkerCopyFailureFailsClosedWithoutResidue/NoMarkerToCopyKeepsTheBehaviourItAlwaysHad (0.02s)
    --- PASS: TestMarkerCopyFailureFailsClosedWithoutResidue/MarkerWithExtraLinesAndCRLFIsCopiedAsOneCleanLine (0.02s)
    --- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue/ABackupPathWithATrailingSeparatorStillGetsASibling (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.905s
FAIL
rc=1
```

붉은 이유: 올바른 이유다. 주입한 표지 복사 실패가 무시되어 상태가 성공 (`1` = `priorManifestWritten`, 오류 없음) 이었고 그 실행이 `SKILL.md` 를 제거했다. 끝에 구분자가 붙은 경로는 형제가 `<dir>/-02` (디렉터리 안) 로 만들어졌다. 나머지 두 절 (표지 없음, CRLF) 은 종전 동작을 고정하며 수정 전에도 통과한다.

#### 경로 경계 반례 표 — RED 셀 (A `d122d6794e`)

`internal/cli/migrate_ownership_boundary_table_test.go` (`//go:build !windows`, `TestPathBoundaryCounterexamples`) 한 시험 안의 표: 복원 (이름 필터), 경계 (`withinRootBoundary`, `pathLeavesRoot`; 문자열 접두사를 공유하는 형제 디렉터리 `proj-xxx-evil`, `proj-xxx2`; 업데이트 명령 형태의 상대 뿌리 `.`), 걷기 (업데이트 형태 실행), 심볼릭 링크 (프로젝트 안, 밖, 상위 안·밖, 사슬, 끊어진 링크, 루프, ENOTDIR, 없는 파일, 일반 파일), 종류 (FIFO 미등록·등록·링크, 디렉터리, 사용자 루트 대응물 FIFO), 빈 `rootResolved`. 행마다 `TABLE|그룹|경우|expected|observed` 한 줄을 로그에 남긴다. FIFO 가 있는 행은 고루틴 + 한정 대기로 멈춤을 감지하고 FIFO 를 건드려 풀며, 실패한 `os.Symlink`·`mkfifo` 는 `t.Fatalf` 이다.

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestPathBoundaryCounterexamples$' -count=1 -v   (표 43행 중 어긋난 행만)
    migrate_ownership_boundary_table_test.go:454: manifest.prior.json.tmp-notes/keep.txt: observed "skipped", want "restored"
    migrate_ownership_boundary_table_test.go:454: manifest.prior.json.tmp-notes/manifest.prior.json: observed "skipped", want "restored"
    migrate_ownership_boundary_table_test.go:454: withinRootBoundary: top-level entry whose name starts with two dots is inside: observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: pathLeavesRoot: file in a top-level directory named '..hidden': observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: root '.': in-project symlink whose target is an ABSOLUTE in-project path: observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: root '.' reached through a symlinked working directory: in-project absolute link target: observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: update-shaped run (root '.'): in-project symlink with an absolute target: observed "class=unregistered reason=symlink-escape removed=1", want "class=unregistered reason=no-provenance-record removed=1"
FAIL
rc=1
```

붉은 이유: 올바른 이유다. 어긋난 일곱 행: 접두사 디렉터리 둘 (N1), 최상위 `..hidden` 이름 둘 (`withinRootBoundary` 의 `HasPrefix(rel, "..")`), 상대 뿌리 `.` 에서 절대 경로 대상의 프로젝트 안 링크 셋 (경로 해석 도우미, 심볼릭 링크로 된 작업 디렉터리, 업데이트 형태 걷기 — 걷기에서는 `symlink-escape` 로 읽혀 `no-provenance-record` 가 아님). 나머지 표 행(경계 위반 형제 디렉터리, 모든 심볼릭 링크 행, 종류 행, 빈 뿌리)은 수정 전에도 기대와 같다 — 바뀌면 안 되는 것의 고정이다.

#### Chunk 11 GREEN 기록과 정직한 회계

수정 커밋: N1 `dcbafe568c`, N2 `90296319b4`, 경계 표 `1003a9edd4`. 시험 커밋: N1 `ef45c82cc9`, N2 `2cd11e520e`, 표 `d122d6794e`; 기록 `5aa3de1348`.

#### N1 — 수정 C `dcbafe568c`

`IsPriorManifestBookkeeping` 이 상대 경로에 구분자가 있으면 (`filepath.ToSlash` 뒤 `/` 포함) 거짓을 돌려준다. 루트 수준의 최종 이름과 임시 접두사 이름만 일치한다. 연속 5회 (`n1-green-1..5.txt`: 복원 네 진입점 + 단위 표, PASS 5):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.497s | PASS=5 FAIL=0 SKIP=0
run2: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.282s | PASS=5 FAIL=0 SKIP=0
run3: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.272s | PASS=5 FAIL=0 SKIP=0
run4: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.264s | PASS=5 FAIL=0 SKIP=0
run5: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.256s | PASS=5 FAIL=0 SKIP=0
```

실제 코드 변이 (구분자 검사를 지움; `cmp` 복원 확인):

```text
--- FAIL: TestRestoreLegacySkipsPriorManifestTemporaryFile (0.00s)
--- FAIL: TestRestoreRetainedWithoutSectionsSkipsPriorManifestTemporaryFile (0.00s)
--- PASS: TestRestoreRetainedWithSectionsSkipsPriorManifestTemporaryFile (0.00s)
--- FAIL: TestRestoreFromBackupDirSkipsPriorManifestTemporaryFile (0.01s)
--- FAIL: TestIsPriorManifestBookkeeping (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.417s
FAIL
rc=1
RESTORED-IDENTICAL
```

붉은 이유: 올바른 이유다 — 접두사 디렉터리 복원 두 진입점과 단위 표가 붉어진다.

#### N2 — 수정 C `90296319b4`

`copyRunMarker` 가 오류를 돌려준다. 원본에 표지가 **없으면** (없음, 일반 파일이 아님) 복사할 것이 없으므로 오류가 아니고 형제는 종전처럼 표지 없이 만들어진다 (이 동작은 시험이 고정한다: `NoMarkerToCopyKeepsTheBehaviourItAlwaysHad`). 표지가 있으나 읽거나 쓸 수 없으면 오류이고, `claimPriorManifestDir` 가 방금 만든 (아직 빈) 형제를 `os.Remove` 로 지운 뒤 (최선 노력) 실패를 돌려준다 → 상태 `priorManifestFailed`, 그 실행은 아무것도 제거하지 않는다. 복사 이음매 `priorManifestMarkerCopy` 는 다른 `priorManifest*` 이음매와 같은 방식 (변수, 시험만 바꿈). 같은 커밋이 `claimPriorManifestDir` 첫 줄에 `filepath.Clean` 을 더했다 (끝에 구분자가 붙은 경로의 형제 문제, 자체 검토 RED).

연속 5회 (`n2-green-1..5.txt`: `TestMarkerCopyFailureFailsClosedWithoutResidue` + chunk 9 의 `TestPriorManifestOfAnEarlierRunSurvivesALaterRun`, PASS 14):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli	3.136s | PASS=14 FAIL=0 SKIP=0
run2: ok  	github.com/modu-ai/moai-adk/internal/cli	3.640s | PASS=14 FAIL=0 SKIP=0
run3: ok  	github.com/modu-ai/moai-adk/internal/cli	2.941s | PASS=14 FAIL=0 SKIP=0
run4: ok  	github.com/modu-ai/moai-adk/internal/cli	2.379s | PASS=14 FAIL=0 SKIP=0
run5: ok  	github.com/modu-ai/moai-adk/internal/cli	3.232s | PASS=14 FAIL=0 SKIP=0
```

실제 코드 변이 셋 (각각 `cmp` 로 복원 확인; 저장 파일 `n2-mutant-*.txt`):

```text
-- 변이 1 (오류 무시):
--- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue (1.42s)
    --- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue/NoSiblingIsLeftAndTheEarlierCopyIsUntouched (0.02s)
    --- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue/TheRunRemovesNothing (1.34s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.429s
FAIL
rc=1
RESTORED-IDENTICAL-ignoreError
-- 변이 2 (형제를 지우지 않음):
--- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue (1.63s)
    --- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue/NoSiblingIsLeftAndTheEarlierCopyIsUntouched (0.03s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.858s
FAIL
rc=1
RESTORED-IDENTICAL-keepSibling
-- 변이 3 (Clean 제거):
--- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue (1.72s)
    --- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue/ABackupPathWithATrailingSeparatorStillGetsASibling (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.802s
FAIL
rc=1
RESTORED-IDENTICAL-noClean
```

붉은 이유: 올바른 이유다 — 변이 1 은 실패한 복사에도 형제가 성공으로 남아 그 실행이 제거했고, 변이 2 는 형제 잔여물이 남았고, 변이 3 은 끝 구분자 경로의 형제가 디렉터리 안에 만들어졌다.

#### 경계 표 수정 C `1003a9edd4`

(1) `withinRootBoundary` 가 `filepath.Rel` 결과의 **첫 요소**가 `..` 인지로 판정한다 (`rel == ".."` 또는 `../` 접두사). 종전의 `HasPrefix(rel, "..")` 는 최상위 이름 `..hidden` 을 밖으로 보았다 (과잉 거부). (2) `pathLeavesRoot` 가 상대 뿌리 (`"."`) 를 절대·해석된 형태로 바꾸고 경로도 절대로 만든 뒤 `EvalSymlinks` 로 푼다. 업데이트 명령은 뿌리 `"."` 를 넘기므로 (`update_template_sync.go:230`) 종전에는 절대 경로 대상의 프로젝트 안 링크가 `Rel(".", 절대경로)` 오류로 밖으로 읽혔다 (읽지 않는 쪽이라 안전하지만 기준 커밋과 다른 운영 동작). 빈 뿌리는 명시적으로 닫힌 쪽 (밖으로 봄). 등록된 파일의 `disposition` 은 바꾸지 않았다: 그쪽의 같은 상대 뿌리 동작은 기준 시점부터 있었고, 바꾸면 절대 경로 대상의 프로젝트 안 링크가 제거 가능으로 바뀔 수 있어 AC-OWR-01 의 고정 기대와 제거 범위를 건드린다 (N16 계열 부채로 기록).

연속 5회 (`table-green-1..5.txt`, `TestPathBoundaryCounterexamples`, 43행 + 부모 = PASS 44):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli	2.436s | PASS=44 FAIL=0 SKIP=0 rows=43
run2: ok  	github.com/modu-ai/moai-adk/internal/cli	1.846s | PASS=44 FAIL=0 SKIP=0 rows=43
run3: ok  	github.com/modu-ai/moai-adk/internal/cli	2.506s | PASS=44 FAIL=0 SKIP=0 rows=43
run4: ok  	github.com/modu-ai/moai-adk/internal/cli	2.056s | PASS=44 FAIL=0 SKIP=0 rows=43
run5: ok  	github.com/modu-ai/moai-adk/internal/cli	2.631s | PASS=44 FAIL=0 SKIP=0 rows=43
```

실제 코드 변이 둘 (`cmp` 복원 확인; 저장 파일 `c3-mutant-*.txt`):

```text
-- 변이 1 (HasPrefix(rel, "..") 로 되돌림):
    migrate_ownership_boundary_table_test.go:454: withinRootBoundary: top-level entry whose name starts with two dots is inside: observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: pathLeavesRoot: file in a top-level directory named '..hidden': observed "outside", want "inside"
RESTORED-IDENTICAL-dotdotPrefix
-- 변이 2 (상대 뿌리 정규화를 끔):
    migrate_ownership_boundary_table_test.go:454: root '.' (the update command's shape): relative in-project file: observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: root '.': in-project symlink whose target is an ABSOLUTE in-project path: observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: root '.' reached through a symlinked working directory: in-project absolute link target: observed "outside", want "inside"
    migrate_ownership_boundary_table_test.go:454: update-shaped run (root '.'): in-project symlink with an absolute target: observed "class=unregistered reason=symlink-escape removed=1", want "class=unregistered reason=n
RESTORED-IDENTICAL-rawRelativeRoot
```

#### 경계 반례 표 — 수정 뒤 관측 (`table-green-1.txt` 의 `TABLE|` 줄, 그대로)

형식: `그룹|경우|expected|observed`. 모든 행에서 expected 와 observed 가 같다 (같지 않으면 시험이 실패한다).

```text
    migrate_ownership_boundary_table_test.go:452: TABLE|restore|manifest.prior.json.tmp-notes/keep.txt|expected: restored|observed: restored
    migrate_ownership_boundary_table_test.go:452: TABLE|restore|manifest.prior.json.d/x|expected: restored|observed: restored
    migrate_ownership_boundary_table_test.go:452: TABLE|restore|manifest.prior.json.tmp-notes/manifest.prior.json|expected: restored|observed: restored
    migrate_ownership_boundary_table_test.go:452: TABLE|restore|manifest.prior.json|expected: skipped|observed: skipped
    migrate_ownership_boundary_table_test.go:452: TABLE|restore|manifest.prior.json.tmp-4242|expected: skipped|observed: skipped
    migrate_ownership_boundary_table_test.go:452: TABLE|restore|notes/manifest.prior.json|expected: restored|observed: restored
    migrate_ownership_boundary_table_test.go:452: TABLE|restore|notes/manifest.prior.json.tmp-4242|expected: restored|observed: restored
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|withinRootBoundary: sibling named root+'-evil' is outside|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|withinRootBoundary: sibling named root+'2' is outside|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|withinRootBoundary: a file under the root is inside|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|withinRootBoundary: the root itself is inside|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|withinRootBoundary: parent of the root is outside|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|withinRootBoundary: '..' escape through the root is outside|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|withinRootBoundary: top-level entry whose name starts with two dots is inside|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|pathLeavesRoot: real file in the sibling 'proj-xxx-evil'|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|pathLeavesRoot: in-project link to the sibling 'proj-xxx2'|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|pathLeavesRoot: regular file in the project|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|pathLeavesRoot: file in a top-level directory named '..hidden'|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|root '.' (the update command's shape): relative in-project file|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|root '.': in-project symlink whose target is an ABSOLUTE in-project path|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|root '.' reached through a symlinked working directory: in-project absolute link target|expected: inside|observed: inside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|root '.': symlink to an outside absolute path|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|boundary|root '.': symlink to an outside relative path|expected: outside|observed: outside
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|symlink to an in-project file: judged by its target's bytes|expected: class=unregistered keep=no-provenance-record read=content|observed: class=unregistered keep=no-provenance-record read=content
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|symlink to an outside file: kept unread|expected: class=unregistered keep=symlink-escape read=none|observed: class=unregistered keep=symlink-escape read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|symlinked ancestor directory outside: kept unread|expected: class=unregistered keep=symlink-escape read=none|observed: class=unregistered keep=symlink-escape read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|symlinked ancestor directory inside: judged by content|expected: class=unregistered keep=no-provenance-record read=content|observed: class=unregistered keep=no-provenance-record read=content
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|chain of links ending outside: kept unread|expected: class=unregistered keep=symlink-escape read=none|observed: class=unregistered keep=symlink-escape read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|chain of links staying inside: judged by content|expected: class=unregistered keep=no-provenance-record read=content|observed: class=unregistered keep=no-provenance-record read=content
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|dangling link: read-failed, nothing read|expected: class=unregistered keep=read-failed read=none|observed: class=unregistered keep=read-failed read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|link loop (ELOOP): read-failed, nothing read|expected: class=unregistered keep=read-failed read=none|observed: class=unregistered keep=read-failed read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|a regular file used as a directory component (ENOTDIR): read-failed|expected: class=unregistered keep=read-failed read=none|observed: class=unregistered keep=read-failed read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|file that does not exist: read-failed|expected: class=unregistered keep=read-failed read=none|observed: class=unregistered keep=read-failed read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|symlink|ordinary file: judged by content|expected: class=unregistered keep=no-provenance-record read=content|observed: class=unregistered keep=no-provenance-record read=content
    migrate_ownership_boundary_table_test.go:452: TABLE|flow|update-shaped run (root '.'): in-project symlink with an absolute target|expected: class=unregistered reason=no-provenance-record removed=1|observed: class=unregistered reason=no-provenance-record removed=1
    migrate_ownership_boundary_table_test.go:452: TABLE|kind|FIFO at an unregistered project path|expected: class=unregistered keep=non-regular-file read=none|observed: class=unregistered keep=non-regular-file read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|kind|FIFO at a registered path|expected: class=user-modified keep=non-regular-file read=none|observed: class=user-modified keep=non-regular-file read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|kind|link to a FIFO inside the project|expected: class=unregistered keep=non-regular-file read=none|observed: class=unregistered keep=non-regular-file read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|kind|directory where an unregistered file was expected|expected: class=unregistered keep=read-failed read=none|observed: class=unregistered keep=read-failed read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|kind|directory where a registered file was expected|expected: class=user-modified keep=read-failed read=none|observed: class=user-modified keep=read-failed read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|kind|FIFO as the user-root counterpart of a registered managed file|expected: keep=counterpart-unconfirmed counterpart=unconfirmed removable=false|observed: keep=counterpart-unconfirmed counterpart=unconfirmed removable=false
    migrate_ownership_boundary_table_test.go:452: TABLE|emptyroot|unregistered ordinary file, rootResolved empty: fail closed|expected: class=unregistered keep=symlink-escape read=none|observed: class=unregistered keep=symlink-escape read=none
    migrate_ownership_boundary_table_test.go:452: TABLE|emptyroot|registered ordinary file, rootResolved empty: never removable|expected: class=managed keep=symlink-escape read=content|observed: class=managed keep=symlink-escape read=content
```

#### 회귀와 점검 (저장 파일 `reg-*.txt`, `v-*.txt`)

변경된 시험 파일 (`git diff --name-only cdd72a814..HEAD -- '*_test.go'`, 32개) 의 최상위 시험 이름 집합을 패키지별로 도출했다: `internal/cli` 57, `internal/cli/update/backup` 9, `internal/manifest` 1. 앵커 선택자로 실행:

```text
reg-cli: ok  	github.com/modu-ai/moai-adk/internal/cli	148.255s rc=0  top-level PASS=57 FAIL=0 SKIP=0
reg-internal_manifest: ok  	github.com/modu-ai/moai-adk/internal/manifest	0.464s rc=0  top-level PASS=1 FAIL=0 SKIP=0
reg-internal_cli_update_backup: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.586s rc=0  top-level PASS=9 FAIL=0 SKIP=0
--- v-build:
rc=0
--- v-winbuild:
rc=0
--- v-vet (host):
rc=0
--- v-winvet:
rc=0
--- v-lint (golangci-lint run --new-from-rev=d116f05f44 ./internal/cli/ ./internal/cli/update/backup/):
0 issues.
rc=0
--- invariants:
$ git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
invariant-exit=0
$ removalEnabled assignments (non-test)
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
$ skip grep over the changed test files (t.Skip / SkipNow / testing.Short)
internal/cli/migrate_ownership_helpers_test.go:1008:	t.Logf("force=false run=1 skipped=%v", first.Skipped)
internal/cli/migrate_ownership_helpers_test.go:1009:	if first.Skipped {
internal/cli/migrate_ownership_symlink_test.go:8:// platform without symlinks; a t.Skip would sweep nothing and read as a pass. A
$ gofmt -l internal/cli
gofmt-exit=0
untracked=0
```

(스킵 grep 의 세 줄은 이전 chunk 와 같다: `first.Skipped` 결과 필드를 쓰는 두 줄과 `t.Skip` 을 설명하는 주석 한 줄이며 `t.Skip`·`SkipNow`·`testing.Short` 호출은 없다.)

**이번 chunk 의 한계 (정직한 회계).**

1. 회귀는 변경된 시험 파일의 이름 집합이며 패키지 전체 시험이 아니다. 패키지 전체 시험·커버리지·`-race`·Windows 시험 실행은 CI 와 착지 재측정 몫이다 (Windows 는 빌드와 vet 만; 경로 구분자 처리는 `filepath.ToSlash` 로 코드 읽기로만 확인).
2. 표지의 첫 줄이 실행 id 로 해석되지 않는 경우 (손상된 표지) 는 그대로 복사되고 형제는 비관리 항목이 된다. 원본도 이름으로 센 독립 실행이라 새 위험은 아니며, 해석 규칙이 백업 패키지의 비공개 함수여서 이번에 넓히지 않았다.
3. `EvalSymlinks` 와 읽기, 종류 확인과 읽기 사이의 경합은 이전 chunk 에서 공개한 한계 그대로다. 권한 오류 (`EACCES`) 가 읽기로 이어지지 않는다는 판단은 root 실행에서 재현할 수 없어 추론이다 (같은 경로 해석이 같은 이유로 실패).
4. 등록된 파일 경로의 읽기가 경계 검사보다 앞서는 부채와, `disposition` 의 상대 뿌리 동작은 기준 시점부터 있었고 그대로다 (N3, N16 계열). 본 chunk 의 상대 뿌리 정규화는 미등록 경로의 가드에만 적용했다.
5. 임시 파일 제거와 형제 제거는 최선 노력이다 (비어 있지 않은 디렉터리는 `os.Remove` 가 지우지 않는다).
6. 코덱스·감사의 재확인은 레인의 몫이다. 이 chunk 는 `.moai/reports`, `.moai/state/audit-receipts`, §E.4, CHANGELOG, SPEC 본문을 건드리지 않았다. N4 (E.4 caveat 문구) 는 N1 수정으로 서술이 사실과 일치하게 되었으나 E.4 는 manager-docs 몫이다.

### Chunk 12 (audit round 3 repair) — 기록 시작

세 번째 독립 delta 감사 (HEAD `1d9b06a737`, 영수증 `rcpt-665b06f8d214d9a97633b90c`) 가 필수 codex 백엔드에서 FAIL 이다. 감사는 기계 부하(155)로 시험을 돌리지 못했고 문서·시험을 다시 읽지 못해 판정은 코덱스의 세 지적에 기댄다. 근거 문서는 `.moai/reports/t1865/sync-audit.md` 의 Findings (고치지 않았다). 이 chunk 는 마지막 수정 라운드다. 리더의 자원 입장 규칙 (게시판 `d-20261018T151201Z-8c1e`): 새 Go 호출 전에 1분 부하 <= 32 이고 여유 메모리 >= 50% 일 때만 시작하고 처음 충족되면 30초 기다려 다시 확인한다. 아래 모든 Go 호출은 `admit.sh` 로 이 규칙을 지켜 한 번에 하나씩 돌렸고, 호출마다의 부하 값은 말미의 표에 있다 (저장 파일 `admit.log`).

#### 결정 기록

- **F-A (SHOULD-FIX, 수정).** 미등록 파일의 읽기를 `os.OpenRoot(프로젝트 뿌리)` 로 연 `Root` 를 통해 한다. 범위는 `classifyUnregistered` 와 그것이 쓰는 새 도우미 `readUnregisteredInsideRoot` (그리고 시험용 이음매 `readWindowHook`, 오류 `errReadEscapesRoot`) 에 엄격히 한정한다. 등록된 파일의 읽기 (`classifyOwnershipCore` 의 등록 분기), `disposition`, 걷기, 빈 디렉터리 정리, 복원 코드, 다른 호출자는 바꾸지 않았다. **등록 경로에는 적용하지 않았다.** 이유: 등록된 파일이 프로젝트 밖을 가리키는 링크인 경우의 AC-OWR-01 고정 기대는 class `managed` + `symlink-escape` 이고, `managed` 는 대상 바이트의 해시 비교로 정해진다. `Root` 로 읽으면 그 바이트를 읽을 수 없어 class 가 바뀌므로 두 줄로 안전하게 바꿀 수 없다. 등록 경로는 기준 시점부터의 N16 계열 공개 부채로 남는다.
- **F-B (SHOULD-FIX, 수정).** 표지 쓰기가 파일을 만든 뒤 실패하면 형제 디렉터리가 비어 있지 않아 `os.Remove` 가 실패했고 형제가 새어 나갔다. `claimPriorManifestDir` 가 방금 `Mkdir` 로 만든 바로 그 경로만 `os.RemoveAll` 한다 (공유 실행 디렉터리는 건드리지 않음). 청구는 여전히 닫힌 쪽으로 실패한다.
- **F-C (MINOR, 거절 — 공개된 한계).** 대소문자를 구분하지 않는 파일 시스템에서 실제 뿌리와 대소문자만 다르게 쓴 링크 대상은 문자열 비교에서 밖으로 보여 `symlink-escape` 로 읽지 않고 보존된다. 이유: (1) 닫힌 쪽(읽지 않음)이라 파일이 지워지거나 밖이 읽히지 않는다; (2) `disposition` 이 기준 시점부터 같은 문자열 비교를 쓴다; (3) `os.SameFile` 로 상위 사슬을 비교하려면 경계 비교 전체를 바꿔야 하는데 이는 이 마지막 라운드의 작은 수정 범위를 넘는다.

#### 시험 먼저 — RED 셀 (A `714adcc444`, 호출 시점 부하: load1=29.92 free=81%)

F-A: `TestPathBoundaryCounterexamples` 표에 경쟁 그룹 `race` 세 행을 더했다 (`migrate_ownership_boundary_table_test.go`). 판정 함수 안 읽기 창에 이음매 `readWindowHook(stage, abs)` 를 두고 (동작 중립; 수정 전에는 `check-to-read` 한 곳만 호출됨), 시험이 그 시점에 대상을 프로젝트 밖 파일로 가는 링크로 바꿔치기한 뒤 밖의 바이트가 읽히지 않고 `symlink-escape` 로 보존되는지 본다. 행: 경계 검사와 읽기 사이 (`check-to-read`), 경로를 푼 뒤 열기 전 (`resolve-to-open`), 통제 (프로젝트 안의 다른 파일로 가는 링크로 바꿔치기하면 여전히 내용으로 판정).
F-B: `TestMarkerCopyFailureFailsClosedWithoutResidue` 에 하위 시험 `FailureAfterTheMarkerFileWasCreatedLeavesNoSibling` 을 더했다 — 표지 복사 이중이 `.moai-run` 파일을 만든 **뒤** 실패를 돌려준다 (이전 이음매 시험은 파일이 생기기 전에 실패했다).

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestPathBoundaryCounterexamples|TestMarkerCopyFailureFailsClosedWithoutResidue)$' -count=1 -v
    migrate_ownership_boundary_table_test.go:504: target swapped for a link to an outside file between the boundary check and the read: observed "class=unregistered keep=no-provenance-record read=other-bytes swapped=true", want "class=unregistered keep=symlink-escape read=none swapped=true"
    migrate_ownership_boundary_table_test.go:504: target swapped for a link to an outside file after the path was resolved, before it is opened: observed "class=unregistered keep=no-provenance-record read=content swapped=false", want "class=unregistered keep=symlink-escape read=none swapped=true"
--- FAIL: TestPathBoundaryCounterexamples (4.36s)
    update_prior_manifest_markercopy_test.go:90: backup directories [20261018_153045 20261018_153045-02] remain, want only the intended 20261018_153045 (the sibling with its half-written marker must be removed)
    update_prior_manifest_markercopy_test.go:99: retention sees runs=1 unmanaged=1, want 1 and 0
--- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue (2.39s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	8.529s
FAIL
rc=1
```

붉은 이유: F-A 첫 행은 올바른 이유다 — 바꿔치기 뒤 밖의 바이트를 읽었다 (`read=other-bytes`) 이유 `no-provenance-record` (`symlink-escape` 가 맞다). 둘째 행(`resolve-to-open`)은 이 창이 수정 전 코드에 없어 이음매가 불리지 않아 (`swapped=false`) 붉다 — 이유가 「창이 없음」 이라 약한 RED 이며, 이 행의 실질 증거는 수정 뒤 초록과 변이다. 통제 행은 수정 전에도 통과한다. F-B 는 올바른 이유다 — 형제 `…-02` 가 남았고 보존 정책이 비관리 항목 1건으로 센다 (`unmanaged=1`).

#### Chunk 12 GREEN 기록과 정직한 회계

수정 커밋 C `dc32a94c9e`; 시험 A `714adcc444`; 결정·RED 기록 `53a6c88b09`.

#### 변경 범위 (F-A 는 엄격히 한정, `git diff --stat 1d9b06a737..HEAD -- internal`)

```text
 internal/cli/migrate_ownership.go                  | 77 +++++++++++++++++++++-
 .../cli/migrate_ownership_boundary_table_test.go   | 50 ++++++++++++++
 internal/cli/update_prior_manifest.go              |  9 ++-
 .../cli/update_prior_manifest_markercopy_test.go   | 31 +++++++++
 4 files changed, 163 insertions(+), 4 deletions(-)
```

- **F-A 로 바뀐 함수** (모두 `internal/cli/migrate_ownership.go`): `classifyUnregistered` (수정: `os.ReadFile(abs)` → `readUnregisteredInsideRoot(env.rootResolved, abs)`, 뿌리 밖이면 이유 `symlink-escape`), `readUnregisteredInsideRoot` (신규, 이 파일 안에서만 쓰임: 뿌리·경로를 절대·해석된 형태로 만들고 뿌리 기준 상대 경로를 `os.OpenRoot(뿌리).Open` 으로 열어 `io.ReadAll`; `Root` 가 뿌리 밖으로 나가는 링크를 거부하면 `errReadEscapesRoot`), 변수 `errReadEscapesRoot` 와 시험용 이음매 `readWindowHook` (신규; 이음매는 시험 커밋 A 에서 동작 중립으로 들어갔다). `pathLeavesRoot` 는 분류 사전 검사로 그대로 남아 있다. 바뀌지 않은 것: 등록된 파일의 읽기 (`classifyOwnershipCore` 의 등록 분기), `disposition`, 걷기, 빈 디렉터리 정리, 복원 코드, 다른 모든 호출자. 저장소 전체에서 다시 쓰이는 도우미를 도입하지 않았다.
- **F-B 로 바뀐 함수:** `claimPriorManifestDir` (`internal/cli/update_prior_manifest.go`: 방금 만든 형제 경로만 `os.RemoveAll`).
- `os.Root` 사용 가능성 점검: `go.mod` 는 `go 1.26.8` 이고 `os.OpenRoot` / `Root.Open` 은 Windows 를 포함해 컴파일된다 (`GOOS=windows go build ./...` rc=0, 아래). 파일/디렉터리 의미: `Root.Open` 이 디렉터리를 열면 `io.ReadAll` 이 실패해 종전처럼 `read-failed`. 종류 (FIFO 등) 는 열기 전에 `unreadableKind` 가 종전처럼 정한다. 표준 라이브러리는 뿌리를 벗어나는 링크를 비공개 오류 (`path escapes from parent`) 로 알리므로 그 문구가 유일한 손잡이이다 — 라이브러리 업그레이드로 문구가 바뀌면 이 분기는 `read-failed` 로 닫힌 쪽(읽지 않음)에 떨어지고 경계 표 행이 붉어진다.

#### F-A 증거

연속 5회 (`green-1..5.txt`: `TestPathBoundaryCounterexamples` 46행 + `TestMarkerCopyFailureFailsClosedWithoutResidue`, 서브 시험 포함 PASS 55):

```text
run1: ok  	github.com/modu-ai/moai-adk/internal/cli	7.613s rc=0  PASS=55 FAIL=0 SKIP=0 rows=46
run2: ok  	github.com/modu-ai/moai-adk/internal/cli	7.771s rc=0  PASS=55 FAIL=0 SKIP=0 rows=46
run3: ok  	github.com/modu-ai/moai-adk/internal/cli	5.557s rc=0  PASS=55 FAIL=0 SKIP=0 rows=46
run4: ok  	github.com/modu-ai/moai-adk/internal/cli	10.221s rc=0  PASS=55 FAIL=0 SKIP=0 rows=46
run5: ok  	github.com/modu-ai/moai-adk/internal/cli	6.679s rc=0  PASS=55 FAIL=0 SKIP=0 rows=46
```

경쟁 행 (`green-1.txt` 의 `TABLE|race|…` 줄, 그대로):

```text
    migrate_ownership_boundary_table_test.go:502: TABLE|race|target swapped for a link to an outside file between the boundary check and the read|expected: class=unregistered keep=symlink-escape read=none swapped=true|observed: class=unregistered keep=symlink-escape read=none swapped=true
    migrate_ownership_boundary_table_test.go:502: TABLE|race|target swapped for a link to an outside file after the path was resolved, before it is opened|expected: class=unregistered keep=symlink-escape read=none swapped=true|observed: class=unregistered keep=symlink-escape read=none swapped=true
    migrate_ownership_boundary_table_test.go:502: TABLE|race|control: target swapped for a link to ANOTHER INSIDE file is still judged by content|expected: class=unregistered keep=no-provenance-record read=other-bytes swapped=true|observed: class=unregistered keep=no-provenance-record read=other-bytes swapped=true
```

실제 코드 변이 둘 (각각 `cmp` 로 복원 확인):

```text
-- 변이 1: 다시 평범한 os.ReadFile(abs) (mutant-plainReadFile.txt)
    migrate_ownership_boundary_table_test.go:504: target swapped for a link to an outside file between the boundary check and the read: observed "class=unregistered keep=no-provenance-record read=other-bytes swapped=true", want "class=unregistered keep=symlink-escape read=none swapped=true"
    migrate_ownership_boundary_table_test.go:504: target swapped for a link to an outside file after the path was resolved, before it is opened: observed "class=unregistered keep=no-provenance-record read=content swapped=false", want "class=unregistered keep=symlink-escape read=none swapped=true"
--- FAIL: TestPathBoundaryCounterexamples (4.30s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	6.001s
FAIL
rc=1
RESTORED-IDENTICAL
-- 변이 2: Root.Open 대신 os.Open(filepath.Join(root, rel)) (mutant-plainOpen.txt) — 경로를 푼 뒤 열기 전 창은 os.Root 만이 닫는다
    migrate_ownership_boundary_table_test.go:504: target swapped for a link to an outside file after the path was resolved, before it is opened: observed "class=unregistered keep=no-provenance-record read=other-bytes swapped=true", want "class=unregistered keep=symlink-escape read=none swapped=true"
--- FAIL: TestPathBoundaryCounterexamples (1.70s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.962s
FAIL
rc=1
RESTORED-IDENTICAL
```

붉은 이유: 올바른 이유다. 변이 1 은 밖의 바이트를 읽는다 (`read=other-bytes`, 이유 `no-provenance-record`). 변이 2 는 첫 창 (경로를 다시 푸는 단계) 은 그대로 막지만 두 번째 창에서 밖의 바이트를 읽는다 — `os.Root` 가 없으면 이 창이 열린다는 증명이다.

#### F-B 증거

위 같은 5회 실행이 `FailureAfterTheMarkerFileWasCreatedLeavesNoSibling` 을 포함한다 (복사 이중이 `.moai-run` 을 만든 뒤 실패; 형제가 남지 않음, 공유 실행 디렉터리의 사본·표지 불변, 보존 정책 `runs=1 unmanaged=0`). 실제 코드 변이 (`os.RemoveAll` → `os.Remove`; `cmp` 복원 확인):

```text
    update_prior_manifest_markercopy_test.go:90: backup directories [20261018_153045 20261018_153045-02] remain, want only the intended 20261018_153045 (the sibling with its half-written marker must be removed)
    update_prior_manifest_markercopy_test.go:99: retention sees runs=1 unmanaged=1, want 1 and 0
--- FAIL: TestMarkerCopyFailureFailsClosedWithoutResidue (2.05s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.565s
FAIL
rc=1
RESTORED-IDENTICAL
```

#### 회귀와 점검 (저장 파일 `reg-*.txt`, `v-*.txt`; Go 호출은 한 번에 하나, 호출마다 입장 규칙 통과)

변경된 시험 파일 (32개) 의 최상위 시험 이름 집합을 패키지별로 도출했다: `internal/cli` 57, `internal/cli/update/backup` 9, `internal/manifest` 1.

```text
reg-cli: ok  	github.com/modu-ai/moai-adk/internal/cli	122.045s rc=0  top-level PASS=57 FAIL=0 SKIP=0
reg-internal_manifest: ok  	github.com/modu-ai/moai-adk/internal/manifest	0.240s rc=0  top-level PASS=1 FAIL=0 SKIP=0
reg-internal_cli_update_backup: ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.392s rc=0  top-level PASS=9 FAIL=0 SKIP=0
--- v-build:
rc=0
--- v-winbuild (os.Root compiles for Windows):
rc=0
--- v-vet (host):
rc=0
--- v-winvet:
rc=0
--- v-lint (golangci-lint run --new-from-rev=feefc4f1f1 ./internal/cli/ ./internal/cli/update/backup/):
0 issues.
rc=0
```

**호출 시점 부하 (`admit.log`, 1분 부하 / 여유 메모리):** 모든 Go 호출이 규칙 (부하 <= 32, 여유 >= 50%) 을 충족한 뒤 30초 재확인을 거쳐 시작됐다.

```text
A-red: go test TestPathBoundaryCounterexamples+TestMarkerCopyFailure... load1=29.92 free=81%
green-1: go test ^(TestPathBoundaryCounterexamples|TestMarkerCopyFailureFailsClosedWithoutResidue)$ load1=22.59 free=81%
green-2: go test ^(TestPathBoundaryCounterexamples|TestMarkerCopyFailureFailsClosedWithoutResidue)$ load1=21.05 free=81%
green-3: go test ^(TestPathBoundaryCounterexamples|TestMarkerCopyFailureFailsClosedWithoutResidue)$ load1=31.72 free=81%
green-4: go test ^(TestPathBoundaryCounterexamples|TestMarkerCopyFailureFailsClosedWithoutResidue)$ load1=30.10 free=81%
green-5: go test ^(TestPathBoundaryCounterexamples|TestMarkerCopyFailureFailsClosedWithoutResidue)$ load1=29.53 free=81%
mutant-plainReadFile: go test ^TestPathBoundaryCounterexamples$ load1=25.04 free=81%
mutant-removeNotRemoveAll: go test ^TestMarkerCopyFailureFailsClosedWithoutResidue$ load1=23.98 free=82%
regression internal/cli (57 names) load1=23.67 free=83%
regression internal/manifest load1=22.00 free=83%
regression internal/cli/update/backup load1=19.35 free=83%
go build ./... load1=19.14 free=83%
GOOS=windows go build ./... load1=19.65 free=83%
go vet host load1=20.51 free=83%
GOOS=windows go vet load1=19.32 free=83%
golangci-lint --new-from-rev=feefc4f1f1 load1=22.01 free=83%
mutant-plainOpen: go test ^TestPathBoundaryCounterexamples$ load1=22.98 free=83%
```

시간 초과·부하 의존 실패는 없었다 (모든 호출이 입장 규칙 안에서 끝났다).

불변식 점검 (`v-invariants.txt`; 스킵 grep 의 세 줄은 이전 chunk 와 같다: `first.Skipped` 필드를 쓰는 두 줄과 `t.Skip` 을 설명하는 주석 한 줄, 호출 없음):

```text
$ git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
invariant-exit=0
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
internal/cli/migrate_ownership_helpers_test.go:1008:	t.Logf("force=false run=1 skipped=%v", first.Skipped)
internal/cli/migrate_ownership_helpers_test.go:1009:	if first.Skipped {
internal/cli/migrate_ownership_symlink_test.go:8:// platform without symlinks; a t.Skip would sweep nothing and read as a pass. A
gofmt-exit=0 untracked=0
```

**이번 chunk 의 한계 (정직한 회계).**

1. 회귀는 변경된 시험 파일의 이름 집합이며 패키지 전체 시험이 아니다. 패키지 전체 시험·커버리지·`-race`·Windows 시험 실행은 CI 와 착지 재측정 몫이다 (Windows 는 빌드와 vet 만).
2. F-A 는 미등록 경로의 읽기만 닫는다. 등록된 파일 읽기는 N16 계열 부채로 남는다 (위 결정). `unreadableKind` 확인과 열기 사이에 일반 파일이 FIFO 로 바뀌는 경합은 여전히 막지 않는다 (`Root.Open` 이 FIFO 를 비차단으로 열지 않는다; 같은 이름의 플랫폼 전용 열기 플래그를 쓰지 않기로 했다).
3. `Root` 의 뿌리 이탈 오류는 문구로 식별한다 (위). 표준 라이브러리가 센티널을 내보내면 바꿔야 한다.
4. F-C 는 거절 (위 결정 기록).
5. 코덱스·감사의 재확인은 레인의 몫이다. 이 chunk 는 `.moai/reports`, `.moai/state/audit-receipts`, §E.4, CHANGELOG, SPEC 본문을 건드리지 않았다.

### Chunk 13 (shrink, leader B) — 기록 시작

네 번째 독립 sync 감사 (HEAD `aba63b1483`, 영수증 `rcpt-bc15a518bfe20c055b8f4bda`) 가 필수 codex 백엔드에서 FAIL 이다: N-A (확인 뒤에 프로젝트 루트를 경로 문자열로 `os.OpenRoot` 로 열어서, 루트 경로를 바꿔치기하면 밖의 바이트를 읽음), N-B (`os.RemoveAll(sibling)` 이 너무 넓음), N-C (바꿔치기한 절대 경로의 프로젝트 안 링크가 `symlink-escape` 로 남음). 리더는 (게시판 `d-20261018T174444Z-b5ea`, 처분 B) 고치지 않고 **보호하는 면을 줄이기로** 결정했다. 근거 문서는 `.moai/reports/t1865/sync-audit.md` (round 4) 와 `sync-hold.md` 이며 고치지 않았다. 자원 입장 규칙은 슬롯 조건이 더해졌다: 1분 부하 <= 32, 여유 메모리 >= 50%, `go-test-heavy` 슬롯이 착지·재측정 세션에 잡혀 있지 않음, 처음 충족되면 30초 뒤 재확인, Go 호출은 한 번에 하나. 모든 호출은 `admit.sh` 가 규칙을 확인한 뒤 시작했고 호출마다 시각·부하·메모리·슬롯 상태가 말미의 `admit.log` 표에 있다. 슬롯 점유가 첫 호출을 한참 늦췄다 (착지 재측정 세션 `lane-40` 이 점유 중이었다; 슬롯은 한 번도 잡지 않았다).

#### 결정과 범위

1. **형제 디렉터리 방식을 없앤다.** 같은 시계 초에 시작한 둘째 실행이 의도한 실행 백업 디렉터리에서 이미 `manifest.prior.json` 을 만나면: 덮어쓰지 않고, 다른 디렉터리도 만들지 않으며, 상태는 `priorManifestFailed` — 이미 있는 닫힌 실패 경로이므로 그 실행은 아무것도 제거하지 않고 (REQ-OWR-016) 보고가 이유를 말하며, 앞 실행의 사본은 바이트 그대로 남는다. 삭제: `claimPriorManifestDir` 의 형제 생성, `copyRunMarker`, 이음매 `priorManifestMarkerCopy`, `priorManifestSiblingMax`, `os.RemoveAll` 정리. 유지: 이름 바꾸기 실패 때의 임시 파일 정리, 복원이 백업 루트의 `manifest.prior.json` 과 `manifest.prior.json.tmp-*` 를 건너뛰는 것 (비판받지 않음).
2. **등록 안 된 파일은 경로 구성 요소에 심볼릭 링크가 하나도 없을 때만 읽는다.** 프로젝트 루트부터 파일까지 모든 구성 요소를 루트 핸들로 `Lstat` 하고, 링크가 하나라도 있으면 (안을 가리키든 밖을 가리키든) 읽지 않고 기존 이유 `symlink-escape` 로 둔다 (class unregistered, `v.data` 없음, 잔여 행). 끊어지거나 순환하는 링크는 `read-failed` 로 남는다 (`Root.Stat` 이 오류). 마지막 요소가 파이프·소켓·장치이면 열기 전에 `non-regular-file`. 이것이 등록 안 된 분기에서 `pathLeavesRoot` 를 대체하며, `pathLeavesRoot` 는 호출자가 없어져 삭제한다 (`withinRootBoundary` 는 다른 사용처가 있어 그대로).
3. **루트를 경로로 다시 열지 않고 한 번만 연다.** `newOwnershipEnv` 가 `os.OpenRoot(projectRoot)` 를 한 번 열어 환경(`ownershipEnv.root`)에 담고, 등록 안 된 파일의 읽기는 모두 그 핸들의 `Lstat`/`Open` 을 쓰며 `migrateProjectCommonAssets` 가 끝날 때 닫는다. 이 환경을 만드는 곳이 둘뿐이라 (마이그레이션과 `--dry-run` 미리보기) 큰 변경이 필요하지 않았다. 미리보기 환경 (`migrate_ownership_env.go`) 은 건드리지 않았다: 그쪽 핸들은 표준 라이브러리의 종료자 (`runtime.SetFinalizer`) 가 닫는다. 시험이 손으로 만든 환경 (`root` 가 nil) 은 그 호출 하나를 위해 경로 문자열로 연다 — 운영 경로는 항상 핸들을 가진다 (시험이 열기 횟수 1 로 고정).
4. **범위 가드.** 바뀌는 함수: `classifyOwnershipCore` (등록 안 된 분기 한 줄), `classifyUnregistered`, 새 `readUnregistered`/`escapesRoot`, `newOwnershipEnv`, `migrateProjectCommonAssets` (`defer env.closeRoot()` 한 줄), `preservePriorManifest` 와 그 이음매, 삭제한 `pathLeavesRoot`·`readUnregisteredInsideRoot`·`errReadEscapesRoot`·`claimPriorManifestDir`·`copyRunMarker`. 바꾸지 않는 것: 등록된 파일의 읽기, `disposition`, 걷기, 빈 디렉터리 정리, 복원 코드, dry-run 파일, manifest/types.

#### 바뀌는 기준선 (결정에 의한 변경)

chunk 6 에서 측정해 고정한 「프로젝트 안 일반 파일을 가리키는 심볼릭 링크는 대상의 바이트로 판정」 (`no-provenance-record`, 바이트를 읽음) 이 **「읽지 않고 `symlink-escape`」** 로 바뀐다. 같은 결정으로 「프로젝트 안 파이프를 가리키는 링크」 의 이유도 `non-regular-file` 에서 `symlink-escape` 로 바뀐다 (링크를 따라가지 않으므로 파이프를 건드리지 않는다). 끊어진 링크와 순환은 종전대로 `read-failed`, 일반 파일은 종전대로 내용으로 판정한다. 이름이 어긋나 보이는 점 (`symlink-escape` 가 프로젝트 안을 가리키는 링크에도 쓰임) 은 리더가 기존 이유를 쓰라고 정했기 때문이다. F-C (대소문자를 구분하지 않는 파일 시스템) 는 건드리지 않았다: 공개된 한계 그대로.

#### 삭제·바뀐 시험과 이음매 (이유와 함께)

| 대상 | 처분 | 이유 |
|---|---|---|
| `TestPriorManifestOfAnEarlierRunSurvivesALaterRun` (여섯 하위 시험: 같은 설정 백업 경로 두 실행, 이름 순서와 읽는 쪽, 직접 만들지 않은 디렉터리 건너뛰기, 이름 소진 닫힘, 설정 경로 없는 실행, 보존 정책이 형제를 같은 실행으로 셈) | **삭제** | 형제 방식이 없어졌다. 대신 `TestSecondClaimInTheSameDirectoryFailsClosed` |
| `TestMarkerCopyFailureFailsClosedWithoutResidue` (표지 복사 실패·실패 뒤 잔여물 없음·파일 생성 뒤 실패·표지 없음·CRLF·이름 점유·끝 구분자) 와 파일 `update_prior_manifest_markercopy_test.go` | **삭제** | 표지 복사와 형제가 없어졌다 |
| 이음매 `priorManifestMarkerCopy` | **삭제** | 위와 같음 |
| 이음매 `readWindowHook` | **유지**, 단계 이름 `check-to-read`/`lstat-to-open` | 두 창에서 바꿔치기하는 경쟁 행에 필요 |
| `migrate_ownership_unregsymlink_test.go`, `…_ancestorlink_test.go`, `…_nonregular_test.go` 의 프로젝트 안 링크 대조군 | **기대 변경** (`symlink-escape`, 읽지 않음) | 바뀐 기준선 |
| 경계 표 (`TestPathBoundaryCounterexamples`) 의 `pathLeavesRoot` 행 아홉 (경계 그룹 네 + 상대 뿌리 `.` 다섯) | **삭제** | 함수가 삭제됨. 같은 질문은 걷기 수준 행 (업데이트 명령 형태, 뿌리 `.`) 과 심볼릭 링크 행이 이제 판정 함수 수준에서 묻는다. 순수 `withinRootBoundary` 행 일곱은 유지 |
| 경계 표 심볼릭 링크 그룹 세 행, 걷기 그룹 한 행, 종류 그룹 한 행 | **기대 변경** | 바뀐 기준선 |
| 경계 표 경쟁 그룹 | **재작성** | 새 창 이름, 절대 링크 행 (`Root` 가 절대 링크를 거부 → N-C 가 같은 이유로 일관), 상대 링크 통제 행, 그리고 **루트 경로 바꿔치기 행** (N-A 재현) |
| `migrate_ownership_recovery_test.go` 의 `NoWaitConsecutiveRuns` 가지 | **삭제**, AC-OWR-18 에 사례 `PriorManifestCollided` 추가 | 시계에 맡긴 가지는 이제 결정적으로 충돌 결과를 단언할 수 없다. 새 사례는 실행이 들어갈 수 있는 모든 초의 실행 디렉터리에 앞 실행의 사본을 심어 충돌을 확실하게 만든다 |
| `migrate_ownership_remaining_test.go` 의 「한 초 안」 사례 | 주석만 변경 | 시험은 「앞 세대가 없는 잔여 목록」 의 렌더링 시험으로 남고 작성 방식은 쓰기 이음매 모사 그대로 |
| 새 시험 | **추가** | `TestSecondClaimInTheSameDirectoryFailsClosed`, `TestMigrationOpensTheProjectRootOnce` |

#### 시험 먼저 — RED 셀 (A `544c071c9e`, 호출 시각·부하는 `admit.log` 첫 줄)

```text
TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestSecondClaimInTheSameDirectoryFailsClosed|TestMigrationOpensTheProjectRootOnce|TestPathBoundaryCounterexamples|TestUnregisteredSymlinkOutsideProjectIsNotRead|TestUnregisteredFileBelowASymlinkedAncestorIsNotRead|TestNonRegularFilesAreKeptWithoutBeingRead|TestInterruptedMigrationRecoversAndConverges|TestRepeatedUpdateIsIdempotent)$' -count=1 -v -timeout 40m
    migrate_ownership_ancestorlink_test.go:137: switch=false in-project symlinked ancestor: class="unregistered" keep="no-provenance-record" data="# inside, through a symlinked ancestor\n", want unregistered/"symlink-escape"/""
    migrate_ownership_ancestorlink_test.go:137: switch=true in-project symlinked ancestor: class="unregistered" keep="no-provenance-record" data="# inside, through a symlinked ancestor\n", want unregistered/"symlink-escape"/""
--- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead (1.29s)
    --- FAIL: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead/ControlsAreJudgedAsDecided (0.00s)
    migrate_ownership_boundary_table_test.go:485: symlink to an in-project file: kept unread (changed baseline, chunk 13): observed "class=unregistered keep=no-provenance-record read=content", want "class=unregistered keep=symlink-escape read=none"
    migrate_ownership_boundary_table_test.go:485: symlinked ancestor directory inside: kept unread (changed baseline, chunk 13): observed "class=unregistered keep=no-provenance-record read=content", want "class=unregistered keep=symlink-escape read=none"
    migrate_ownership_boundary_table_test.go:485: chain of links staying inside: kept unread (changed baseline, chunk 13): observed "class=unregistered keep=no-provenance-record read=content", want "class=unregistered keep=symlink-escape read=none"
    migrate_ownership_boundary_table_test.go:485: update-shaped run (root '.'): in-project symlink with an absolute target is kept unread: observed "class=unregistered reason=no-provenance-record removed=1", want "class=unregistered reason=symlink-escape removed=1"
    migrate_ownership_boundary_table_test.go:485: update-shaped run (root '.'): in-project symlink with a relative target is kept unread: observed "class=unregistered reason=no-provenance-record removed=1", want "class=unregistered reason=symlink-escape removed=1"
    migrate_ownership_boundary_table_test.go:485: target swapped for a link to an outside file after the path was examined, before it is opened: observed "class=unregistered keep=no-provenance-record read=content swapped=false", want "class=unregistered keep=symlink-escape read=none swappe
    migrate_ownership_boundary_table_test.go:485: target swapped for an ABSOLUTE link to a file inside the project: the Root refuses absolute links, so it is kept unread: observed "class=unregistered keep=no-provenance-record read=content swapped=false", want "class=unregistered keep=symli
    migrate_ownership_boundary_table_test.go:485: control: target swapped for a RELATIVE link to another file inside the project is followed inside the Root: observed "class=unregistered keep=no-provenance-record read=content swapped=false", want "class=unregistered keep=no-provenance-reco
    migrate_ownership_boundary_table_test.go:485: the project root path swapped for a link to an outside tree after the environment was built: observed "class=unregistered keep=symlink-escape read=none swapped=true", want "class=unregistered keep=no-provenance-record read=content swapped=t
    migrate_ownership_boundary_table_test.go:485: link to a FIFO inside the project: a link is never followed: observed "class=unregistered keep=non-regular-file read=none", want "class=unregistered keep=symlink-escape read=none"
--- FAIL: TestPathBoundaryCounterexamples (1.93s)
    --- FAIL: TestPathBoundaryCounterexamples/symlink/symlink_to_an_in-project_file:_kept_unread_(changed_baseline,_chunk_13) (0.00s)
    --- FAIL: TestPathBoundaryCounterexamples/symlink/symlinked_ancestor_directory_inside:_kept_unread_(changed_baseline,_chunk_13) (0.00s)
    --- FAIL: TestPathBoundaryCounterexamples/symlink/chain_of_links_staying_inside:_kept_unread_(changed_baseline,_chunk_13) (0.00s)
    --- FAIL: TestPathBoundaryCounterexamples/flow/update-shaped_run_(root_'.'):_in-project_symlink_with_an_absolute_target_is_kept_unread (0.43s)
    --- FAIL: TestPathBoundaryCounterexamples/flow/update-shaped_run_(root_'.'):_in-project_symlink_with_a_relative_target_is_kept_unread (0.45s)
    --- FAIL: TestPathBoundaryCounterexamples/race/target_swapped_for_a_link_to_an_outside_file_after_the_path_was_examined,_before_it_is_opened (0.00s)
    --- FAIL: TestPathBoundaryCounterexamples/race/target_swapped_for_an_ABSOLUTE_link_to_a_file_inside_the_project:_the_Root_refuses_absolute_links,_so_it_is_kept_unread (0.00s)
    --- FAIL: TestPathBoundaryCounterexamples/race/control:_target_swapped_for_a_RELATIVE_link_to_another_file_inside_the_project_is_followed_inside_the_Root (0.00s)
    --- FAIL: TestPathBoundaryCounterexamples/race/the_project_root_path_swapped_for_a_link_to_an_outside_tree_after_the_environment_was_built (0.00s)
    --- FAIL: TestPathBoundaryCounterexamples/kind/link_to_a_FIFO_inside_the_project:_a_link_is_never_followed (0.00s)
    migrate_ownership_nonregular_test.go:231: .agents/skills/moai-nonregular/link-to-pipe: reason "non-regular-file", want "symlink-escape"
    migrate_ownership_nonregular_test.go:231: .agents/skills/moai-nonregular/link-to-plain: reason "no-provenance-record", want "symlink-escape"
--- FAIL: TestNonRegularFilesAreKeptWithoutBeingRead (0.70s)
--- PASS: TestRepeatedUpdateIsIdempotent (7.62s)
    migrate_ownership_recovery_test.go:902: compared tree hash: content-only (path, bytes, mode, directory presence); directory mtimes and clock values are not compared; deployed_at is masked between runs
    migrate_ownership_recovery_test.go:865: (interrupted) .claude/skills/moai-foundation-core/SKILL.md was removed or changed by the interrupted run (open <tmp>/.claude/skills/moai-foundation-core/SKILL.md: no such file or directory)
    migrate_ownership_recovery_test.go:865: (interrupted) the manifest lost its entry for .claude/skills/moai-foundation-core/SKILL.md
    migrate_ownership_recovery_test.go:865: (interrupted) .claude/skills/moai-foundation-core/SKILL.md has no row with reason backup-failed (row "")
    migrate_ownership_recovery_test.go:865: (interrupted) a complete backup copy of .claude/skills/moai-foundation-core/SKILL.md exists although nothing was removed
    migrate_ownership_recovery_test.go:865: (interrupted) .claude/skills/moai-ref-owasp-checklist/SKILL.md was removed or changed by the interrupted run (open <tmp>/.claude/skills/moai-ref-owasp-checklist/SKILL.md: no such file or directory)
    migrate_ownership_recovery_test.go:865: (interrupted) the manifest lost its entry for .claude/skills/moai-ref-owasp-checklist/SKILL.md
    migrate_ownership_recovery_test.go:865: (interrupted) .claude/skills/moai-ref-owasp-checklist/SKILL.md has no row with reason backup-failed (row "")
    migrate_ownership_recovery_test.go:865: (interrupted) a complete backup copy of .claude/skills/moai-ref-owasp-checklist/SKILL.md exists although nothing was removed
    migrate_ownership_recovery_test.go:865: (interrupted) .claude/skills/moai-workflow-testing/SKILL.md was removed or changed by the interrupted run (open <tmp>/.claude/skills/moai-workflow-testing/SKILL.md: no such file or directory)
    migrate_ownership_recovery_test.go:865: (interrupted) the manifest lost its entry for .claude/skills/moai-workflow-testing/SKILL.md
    migrate_ownership_recovery_test.go:865: (interrupted) .claude/skills/moai-workflow-testing/SKILL.md has no row with reason backup-failed (row "")
    migrate_ownership_recovery_test.go:865: (interrupted) a complete backup copy of .claude/skills/moai-workflow-testing/SKILL.md exists although nothing was removed
    migrate_ownership_recovery_test.go:869: the collided run wrote its own copy: <tmp>/.moai-backups/20261019_030526-02/manifest.prior.json
    migrate_ownership_recovery_test.go:878: the run does not say why it removed nothing: [{sev:1 msg:Hooks: running Claude Code sessions need a /hooks review (or a Claude Code restart) for new or changed hooks to take effect.}]
--- FAIL: TestInterruptedMigrationRecoversAndConverges (8.77s)
    --- FAIL: TestInterruptedMigrationRecoversAndConverges/PriorManifestCollided/DistinctRunDirectories (0.50s)
    migrate_ownership_rootonce_test.go:46: the project root was opened 0 times, want exactly once per run
--- FAIL: TestMigrationOpensTheProjectRootOnce (1.11s)
    migrate_ownership_unregsymlink_test.go:121: switch=false in-project symlink: class="unregistered" keep="no-provenance-record" data="# an unregistered file inside the project\n", want unregistered/"symlink-escape"/no bytes
    migrate_ownership_unregsymlink_test.go:121: switch=true in-project symlink: class="unregistered" keep="no-provenance-record" data="# an unregistered file inside the project\n", want unregistered/"symlink-escape"/no bytes
    migrate_ownership_unregsymlink_test.go:175: .agents/skills/moai-unreg-symlink/to-inside: row {Rel:.agents/skills/moai-unreg-symlink/to-inside Class:unregistered Reason:no-provenance-record Missing:[deployed-version original-deployed-path prior-manifest-record] Counterpart:n/a ProfileSi
--- FAIL: TestUnregisteredSymlinkOutsideProjectIsNotRead (0.80s)
    update_prior_manifest_collision_test.go:90: the second claim in the same directory: ("<tmp>/.moai-backups/20261018_153045-02", 1, <nil>), want a failed state with an error
    update_prior_manifest_collision_test.go:198: state=1, want failed
    update_prior_manifest_collision_test.go:201: the run removed [.claude/skills/moai-foundation-core/SKILL.md] although its prior manifest copy was not kept
    update_prior_manifest_collision_test.go:204: the report does not say why the run removed nothing:
--- FAIL: TestSecondClaimInTheSameDirectoryFailsClosed (1.58s)
    --- FAIL: TestSecondClaimInTheSameDirectoryFailsClosed/SameDirectory (0.01s)
    --- FAIL: TestSecondClaimInTheSameDirectoryFailsClosed/TheLaterRunRemovesNothingAndSaysWhy (1.54s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	24.703s
FAIL
rc=1
[...] 「the earlier copy in <dir> changed」 줄 175개 생략 (아래 주의)
```

붉은 이유: 올바른 이유다. (1) 둘째 청구가 같은 디렉터리에서 형제 `…-02` 를 만들어 성공했고 (`state=1`), 그 실행이 `SKILL.md` 를 제거했으며 보고가 이유를 말하지 않았다. (2) 프로젝트 안 링크 (직접, 상위 디렉터리, 사슬, 업데이트 형태 걷기 둘) 가 바이트를 읽고 `no-provenance-record` 로 판정됐다 (`symlink-escape` 가 결정). 파이프를 가리키는 링크는 `non-regular-file` 였다. (3) 새 창 `lstat-to-open` 은 수정 전 코드에 없어 이음매가 불리지 않았다 (`swapped=false`; 약한 RED — 실질 증거는 수정 뒤 초록과 변이). (4) 루트 경로 바꿔치기 행은 환경이 만든 핸들을 쓰지 않고 경로 문자열로 열어 `symlink-escape` 로 관측됐다 (기대: 핸들이 가리키는 원본 프로젝트의 내용을 읽음). (5) `TestMigrationOpensTheProjectRootOnce`: 열기 0회 (`newOwnershipEnv` 가 아직 열지 않음). (6) AC-OWR-18 의 `PriorManifestCollided` 가 형제 `…-02` 에 자기 사본을 썼다.

**주의 (정직한 회계).** 이 RED 실행의 `PriorManifestCollided` 는 첫 판본의 단언을 쓰고 있었다: 심은 모든 디렉터리에 앞 사본이 남아 있어야 한다고 단언했는데 실행 끝의 백업 보존 (기존 별개 경로) 이 오래된 실행 디렉터리를 정리해 대부분이 사라졌다 (그래서 「earlier copy … changed」 줄이 많다). 커밋하기 전에 단언을 「남은 디렉터리만 판정: 하나 이상 남음, 모두 심은 이름, 모두 앞 사본이 바이트 그대로」 로 고쳤다. 고친 단언은 아직 새 코드로 실행하지 않았다 — 수정 뒤 초록 실행이 그 첫 실행이다.

### Chunk 13 (shrink) — GREEN record

**커밋.** RED 시험 A `544c071c9e`, 기록 시작 `661bf4be08`, 구현 C1 `b2ad99d068` (형제 방식 삭제, 같은 초 충돌은 닫힌 실패), 구현 C2 `36f3315244` (미등록 읽기: 경로에 링크가 없을 때만, 루트 핸들 한 번 열기, 미리보기도 닫기), 시험 수정 C3 `7c3abe0e0d` (한 초 렌더링 사례). 이 기록 커밋은 그 뒤.

**이어받은 미검증 diff 에서 바꾼 것 (이유와 함께).** 이전 에이전트의 미커밋 diff (사본: `/private/tmp/own-fixture-d1/c13b/inherited-step-c.patch`) 는 빌드는 되었으나 세 가지가 맞지 않았다. (1) 커밋된 RED 시험 두 파일 (`migrate_ownership_ancestorlink_test.go`, `migrate_ownership_boundary_table_test.go`) 의 기대를 구현에 맞춰 고쳐 놓았다 (절대 경로 링크 대상을 상대 경로로 바꾸고 「끊어진 링크 + 절대 대상」 의 기대를 `read-failed` 에서 `symlink-escape` 로 바꿈). 시험을 구현에 맞추는 것은 허용되지 않아 그 시험 편집을 되돌렸고 (`git checkout`), 구현을 커밋된 기대에 맞췄다. (2) 링크를 만났을 때 이유를 가르는 확인이 `Root.Stat` 이었다. `Root` 는 절대 대상 링크를 평가하지 못해 끊어진 절대 링크가 `symlink-escape` 로 새었다. 메타데이터 확인 `os.Stat(abs)` (열지도 읽지도 않음, 바꿔치기가 바꾸는 것은 이유 하나뿐) 로 바꿔 커밋된 기대 (`read-failed`) 와 맞췄다. (3) 미리보기 환경 (`previewVerdicts`) 은 핸들을 열고 닫지 않았다: `previewVerdicts.close` 와 `defer verdicts.close()` 한 줄 (`update_dryrun_preview.go`) 을 더하고 시험 `TestPreviewVerdictsCloseTheRootTheyOpened` (파일 `migrate_ownership_previewclose_test.go`) 를 더했다. 나머지 (`readUnregistered` 의 구성 요소별 `Lstat`, `escapesRoot`, `newOwnershipEnv` 의 한 번 열기, `defer env.closeRoot()`, `preservePriorManifest` 의 닫힌 실패) 는 그대로 받았고 아래 시험으로 검증했다.

**전체 회귀가 찾은 것 (C3).** 변경된 시험 파일 이름 집합 전체를 돌리니 `TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond` 가 붉었다 (이전 에이전트의 선택 시험 집합에는 없었다). 이 사례는 한 초 충돌을 쓰기 이음매 모사로 흉내냈으나 시계를 기다리지 않아, 실제로 같은 초에 시작한 둘째 실행이 새 규칙대로 닫힌 실패를 만났다 (사본 없음, 첫 실행의 사본만 남음). 제품 결함이 아니라 시험의 전제였다. 두 사례 모두 다음 초까지 기다리고 (`awaitClockSecondAfter`) 모사는 청구 뒤 쓰기 이음매에서만 한다. 단언은 약해지지 않았다.

**닫히지 않은 것을 닫은 확인.** 위 RED 기록의 「고친 `PriorManifestCollided` 단언은 아직 새 코드로 실행하지 않았다」 에 대한 답: 수정 뒤 첫 실행 (green-1) 에서 `TestInterruptedMigrationRecoversAndConverges` (하위 `PriorManifestCollided/DistinctRunDirectories` 포함) 가 초록이었고 세 번 반복에서도 같다.

**바뀌는 함수 (범위 증명).** 제품 코드 diff `661bf4be08..HEAD` 의 함수: `classifyOwnershipCore` (미등록 가지의 두 확인 제거), `classifyUnregistered`, 새 `readUnregistered`, 새 `escapesRoot`, `readWindowHook` (단계 이름 바뀜), `newOwnershipEnv` (루트 한 번 열기), `migrateProjectCommonAssets` (`defer`), `preservePriorManifest`, 새 `previewVerdicts.close`, `previewReconciliation` (`defer` 한 줄). 삭제: `pathLeavesRoot`, `readUnregisteredInsideRoot`, `errReadEscapesRoot`, `claimPriorManifestDir`, `copyRunMarker`, 이음매 `priorManifestMarkerCopy`, 상수 `priorManifestSiblingMax`, `os.RemoveAll` 정리. `unreadableKind` 는 그대로 있고 등록된 파일 가지만 쓴다 (등록된 읽기·`disposition`·걷기·빈 디렉터리 정리·복원·manifest/types 는 건드리지 않음). `git diff --stat 661bf4be08 HEAD`: 7 files, 160 insertions, 197 deletions.

**삭제한 시험과 이음매의 처분.** RED 기록의 표 그대로 (이번 구현이 바꾼 것은 없음): `TestPriorManifestOfAnEarlierRunSurvivesALaterRun` 과 `TestMarkerCopyFailureFailsClosedWithoutResidue` (파일 `update_prior_manifest_markercopy_test.go` 포함) 는 형제와 표지 복사가 사라져 삭제, 이음매 `priorManifestMarkerCopy` 삭제, `readWindowHook` 유지.

**바뀐 기준선 (결정에 의한).** 프로젝트 안 일반 파일을 가리키는 심볼릭 링크 (직접, 상위 디렉터리, 사슬, 상대·절대 대상, 업데이트 형태 걷기의 상대 뿌리 `.`) 는 이제 바이트를 읽지 않고 `symlink-escape` 로 남는다 (chunk 6 의 「대상의 바이트로 판정」 이 바뀜). 프로젝트 안 파이프를 가리키는 링크의 이유는 `non-regular-file` 에서 `symlink-escape`. 끊어지거나 순환하는 링크는 종전대로 `read-failed`, 일반 파일은 종전대로 내용으로 판정한다.

**F-C 는 건드리지 않았다.** 대소문자를 구분하지 않는 파일 시스템 한계 (감사 F-C, 공개된 한계) 는 그대로이고 이번 chunk 의 코드·시험·기록에서 바뀐 것이 없다.

**GREEN 셀 (저장 파일에서 스크립트로 옮김; 파일은 `/private/tmp/own-fixture-d1/c13b/`).**

```text
green-1: rc=0  ok  	github.com/modu-ai/moai-adk/internal/cli	29.481s  PASS=101 FAIL=0 SKIP=0
green-2: rc=0  ok  	github.com/modu-ai/moai-adk/internal/cli	28.639s  PASS=101 FAIL=0 SKIP=0
green-3: rc=0  ok  	github.com/modu-ai/moai-adk/internal/cli	24.991s  PASS=101 FAIL=0 SKIP=0
selectors: ^(TestSecondClaimInTheSameDirectoryFailsClosed|TestMigrationOpensTheProjectRootOnce|TestPathBoundaryCounterexamples|TestUnregisteredSymlinkOutsideProjectIsNotRead|TestUnregisteredFileBelowASymlinkedAncestorIsNotRead|TestNonRegularFilesAreKeptWithoutBeingRead|TestInterruptedMigrationRecoversAndConverges|TestRepeatedUpdateIsIdempotent|TestPreviewVerdictsCloseTheRootTheyOpened)$
```

**변이 (복원은 모두 `cmp` 동일, 변이 뒤 트리 깨끗).**

```text
restore overwrites (claim check disabled): rc=1; red test(s): --- FAIL: TestSecondClaimInTheSameDirectoryFailsClosed (0.82s); RESTORED-IDENTICAL, TREE-CLEAN-AFTER-RESTORE
in-project symlink read allowed (link branch disabled): rc=1; red test(s): --- FAIL: TestPathBoundaryCounterexamples (3.84s); RESTORED-IDENTICAL, TREE-CLEAN-AFTER-RESTORE
per-file OpenRoot by string, opened BEFORE the check (fallback every file): rc=1; red test(s): --- FAIL: TestMigrationOpensTheProjectRootOnce (0.34s); RESTORED-IDENTICAL, TREE-CLEAN-AFTER-RESTORE
per-file OpenRoot by string, opened AFTER the check hook (the N-A shape): rc=1; red test(s): --- FAIL: TestPathBoundaryCounterexamples (1.52s); RESTORED-IDENTICAL, TREE-CLEAN-AFTER-RESTORE
migration never closes the root (defer removed): rc=1; red test(s): --- FAIL: TestMigrationOpensTheProjectRootOnce (1.39s); RESTORED-IDENTICAL, TREE-CLEAN-AFTER-RESTORE
preview never closes the root: rc=1; red test(s): --- FAIL: TestPreviewVerdictsCloseTheRootTheyOpened (0.60s); RESTORED-IDENTICAL, TREE-CLEAN-AFTER-RESTORE
```

변이 3 은 처음에 핸들 가져오기를 검사 전에 두어서 루트 바꿔치기 행이 붉어지지 않고 열기 횟수 시험만 붉었다. 바꿔치기가 일어나는 `check-to-read` 이음매 뒤에서 문자열로 여는 변이 4 를 더해 루트 바꿔치기 행 (`the project root path swapped ... read=other-bytes`: 밖의 바이트를 읽음) 이 붉음을 관측했다.

**회귀와 확인 (변경된 시험 파일 이름 집합, 패키지 전체 시험은 CI/착지 재측정 몫).**

```text
first changed-file regression of internal/cli (before the remaining-test fix): rc=1, PASS=347 FAIL=2 (--- FAIL: TestRemainingListKeeps313UnregisteredCopies (8.89s), --- FAIL: TestRemainingListKeeps313UnregisteredCopies/RenderedByTheUpdateFlowInOneSecond (1.76s))
final changed-file regression of internal/cli (after the fix): rc=0, ok  	github.com/modu-ai/moai-adk/internal/cli	75.372s, PASS=349 FAIL=0 SKIP=0
internal/manifest: rc=0, ok  	github.com/modu-ai/moai-adk/internal/manifest	0.260s, PASS=5 FAIL=0 SKIP=0
internal/cli/update/backup: rc=0, ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.422s, PASS=9 FAIL=0 SKIP=0
go build ./... : rc=0
GOOS=windows go build ./... : rc=0
go vet (cli, backup, manifest) host : rc=0
GOOS=windows go vet (same) : rc=0
golangci-lint run --new-from-rev=aba63b1483 (cli, backup, manifest): 0 issues. rc=0
$ gofmt -l internal/cli internal/manifest
gofmt-exit=0
$ git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
invariant-exit=0
$ grep -rn "removalEnabled\s*=" internal --include=*.go (non-test)
internal/cli/migrate_project_assets.go:44:var removalEnabled = false
$ skip-grep over the card test files
skip-grep-exit=1
$ git status --short (untracked)
untracked=0
```

**자원 입장 기록 (`admit.log`, 호출마다 시각·1분 부하·여유 메모리·슬롯 상태; 슬롯은 한 번도 잡지 않음).** 첫 `go build` 호출은 슬롯을 잡은 착지 재측정 세션 (`lane-36`) 때문에 입장이 오래 걸렸다 (입장 대기 중 도구 시간 제한으로 배경으로 넘어간 `admit.sh` PID 62479 를 내가 기록한 PID 로 직접 종료하고 대기 스크립트로 다시 시작했다; 그 사이 Go 호출은 없었다). `held-other` 는 착지·재측정 세션이 아닌 다른 세션의 점유.

```text
2026-10-18T22:53:51Z go build ./... load1=7.63 free=83% slot=free
2026-10-18T22:54:37Z green-1: go test new/changed selectors load1=13.11 free=83% slot=free
2026-10-18T22:55:54Z green-2: go test new/changed selectors load1=18.56 free=83% slot=free
2026-10-18T22:56:58Z green-3: go test new/changed selectors load1=15.41 free=82% slot=free
2026-10-18T22:58:21Z mutant-restoreOverwrite: go test ^TestSecondClaimInTheSameDirectoryFailsClosed$ load1=12.63 free=83% slot=free
2026-10-18T22:59:09Z mutant-allowInProjectLink: go test ^TestPathBoundaryCounterexamples$ load1=9.81 free=82% slot=free
2026-10-18T23:53:40Z mutant-perFileOpenByString: go test ^(TestPathBoundaryCounterexamples|TestMigrationOpensTheProjectRootOnce)$ load1=14.86 free=80% slot=free
2026-10-18T23:54:35Z mutant-migrationRootNeverClosed: go test ^TestMigrationOpensTheProjectRootOnce$ load1=15.02 free=80% slot=free
2026-10-19T01:00:18Z mutant-previewRootNeverClosed: go test ^TestPreviewVerdictsCloseTheRootTheyOpened$ load1=11.07 free=81% slot=free
2026-10-19T01:01:14Z regression internal/cli load1=17.80 free=80% slot=free
2026-10-19T01:03:15Z regression internal/manifest load1=13.29 free=81% slot=held-other
2026-10-19T01:03:46Z regression internal/cli/update/backup load1=13.72 free=80% slot=held-other
2026-10-19T01:04:18Z go build ./... load1=15.29 free=81% slot=held-other
2026-10-19T01:04:57Z GOOS=windows go build ./... load1=15.14 free=81% slot=held-other
2026-10-19T01:05:35Z go vet host load1=15.73 free=81% slot=held-other
2026-10-19T01:06:09Z GOOS=windows go vet load1=15.45 free=81% slot=held-other
2026-10-19T01:06:53Z golangci-lint --new-from-rev=aba63b1483 load1=12.67 free=81% slot=held-other
2026-10-19T01:09:32Z mutant-perFileOpenByStringAfterCheck: go test ^TestPathBoundaryCounterexamples$ load1=14.90 free=81% slot=held-other
2026-10-19T01:10:15Z final regression internal/cli (after the remaining-test fix) load1=14.44 free=81% slot=held-other
2026-10-19T01:12:17Z final go vet host load1=8.13 free=81% slot=held-other
2026-10-19T01:12:50Z final GOOS=windows go vet load1=8.16 free=81% slot=held-other
2026-10-19T01:13:29Z final golangci-lint --new-from-rev=aba63b1483 load1=7.41 free=81% slot=held-other
```

**정직한 회계.** (1) 회귀는 변경된 시험 파일의 이름 집합이다. 패키지 전체 시험·커버리지·`-race`·Windows 시험 실행은 CI 와 착지 재측정 몫이다 (Windows 는 빌드와 vet 만). (2) `go build ./...` 와 `GOOS=windows go build ./...` 는 C3 (시험 한 파일 수정) 이전의 트리에서 돌렸다; C3 는 `_test.go` 하나라서 빌드에 영향이 없고 vet·lint·전체 회귀는 C3 뒤에 다시 돌렸다. (3) 같은 호출들 중 일부 (`regression internal/manifest` 부터 `golangci-lint` 첫 실행까지, 그리고 C3 뒤 마지막 묶음) 는 `go-test-heavy` 슬롯이 다른 세션에 잡힌 채로 시작했다 (착지·재측정 세션은 아님, 규칙 충족). (4) `Root` 의 뿌리 이탈 오류는 여전히 문구로 식별한다 (chunk 12 의 한계 그대로). (5) 끊어짐/순환 판별은 `os.Stat(abs)` 메타데이터 확인이라 확인과 열기 사이 바꿔치기는 이유(`read-failed` 대 `symlink-escape`)만 바꾸고 바이트를 읽는 길을 열지 않는다. (6) 이 기록 뒤 감사는 레인 몫이다; 이 chunk 는 `.moai/reports`, `.moai/state/audit-receipts`, §E.4, CHANGELOG, SPEC 본문을 건드리지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-18
run_commit_sha: 7c3abe0e0d  # code-verification baseline: the last test/code commit (chunk 13 shrink: C1 b2ad99d068, C2 36f3315244, test fix 7c3abe0e0d; chunk 12 dc32a94c9e); the evidence commit that carries this block follows it
run_status: complete
ac_pass_count: 21
ac_fail_count: 0
preserve_list_post_run_count: 3  # internal/config/protected_zone.go; internal/config/userroots.go; internal/profile/** — diff-empty vs cdd72a814 over the card's commits (positive control: 27 spec commits)
l44_pre_commit_fetch: not-run (lane worktree on a card branch; HEAD and branch re-read immediately before every commit instead of a shared-branch fetch)
l44_post_push_fetch: not-applicable (the lane never pushes; the landing is the leader's batch act)
new_warnings_or_lints_introduced: 0  # one intermediate golangci-lint finding (two unused fields in the AC-OWR-15 test) found and fixed in 8afbbe9436; the 3 errcheck findings in internal/cli/factory_run_adopt.go predate the card and are untouched
cross_platform_build.darwin: PASS (go build ./... rc=0; go vet rc=0)
cross_platform_build.windows: PASS (GOOS=windows go build ./... rc=0; GOOS=windows go vet ./internal/cli/ ./internal/manifest/ rc=0)
total_run_phase_files: 49  # non-.moai files added or edited over the card, run phase only (CHANGELOG.md is the sync phase's): 33 test files, 15 Go sources, 1 data file (chunk 13 added 2 test files, rootonce and previewclose, deleted 1, markercopy, and edited Go sources already counted)
m1_to_mN_commit_strategy: per acceptance criterion A (tests only) -> B (cell recorded from saved output files) -> C (implementation, where something was open); explicit pathspecs only; no push; M8-M10 needed no C; chunk 5 card-review fixes: A 73e37d3832, B 9c00e81a2c, C 051ddc18ef; chunk 6 card-review fix: A 2eeb0559bd, B 60d38d1559, C 1d5ecbbc7c; chunk 9 sync-audit repair: F1 A 6315473594, B c607c315bc, C e0c2896d64; F3 A 34c69c19fc, B 2cbb1af355, C d8fa645c41; F2 A ed58e64c3c, B cc4ca5b239, C fa26fcaf17; chunk 10 card-review round 4: (A) A 780f7862f2, C 848bbc4707; (B) A 9cc9b77a01, C d8a5f75053; record 70cb8c0c56; chunk 11 delta-audit repair: N1 A ef45c82cc9, C dcbafe568c; N2 A 2cd11e520e, C 90296319b4; boundary table A d122d6794e, C 1003a9edd4; record 5aa3de1348; chunk 12 audit round 3 repair: F-A/F-B A 714adcc444, C dc32a94c9e; record 53a6c88b09; chunk 13 shrink: A 544c071c9e, record start 661bf4be08, C1 b2ad99d068, C2 36f3315244, test fix 7c3abe0e0d, GREEN record follows
```

**AC 행렬 (21개).** 등급은 RED-now 를 이 트리에서 실제로 관측했는지로 정했다 (acceptance.md 「등급과 RED 셀 규약」 1항): release-blocking = A 직후 트리에서 판정 명령이 붉음을 보였고 명령·원문·종료 코드·트리 SHA 가 셀에 있음; regression-guard = 도착 시 초록이고 붉음은 변이에서 관측. 이 분류는 사실 기록이며 최종 판단은 감사의 몫이다.

| AC | 등급 | RED-now | A | B | C |
|----|------|---------|---|---|---|
| AC-OWR-01 ~ -04 | release-blocking | 빌드 실패 형 | `9b73a35938` | `4764764ff0` | `3133329a64` |
| AC-OWR-05 (a)-(g) | release-blocking | 빌드 실패 형 | `9b73a35938` | `4764764ff0` | `3133329a64` |
| AC-OWR-05 (h) | release-blocking | 빌드 실패 형 | `f9206f5242` + `da53289c5f` | `558dd9e6c2` | `f8936c2070` |
| AC-OWR-06 | release-blocking | 빌드 실패 형 | `f9206f5242` + `da53289c5f` | `558dd9e6c2` | `f8936c2070` |
| AC-OWR-07 | release-blocking | 빌드 실패 형 | `73ace0cc2e` | `36f638ef0a` | `cca0c9ead1` |
| AC-OWR-08 | **regression-guard** | 도착 시 초록, 운영 코드 변이 둘과 시험 안 변이에서 붉음 | `aa8037a7f2` | `ae4032950e` | — |
| AC-OWR-09 | **regression-guard** | 세 변형 모두 재현 안 됨, 도착 시 초록 | `0ac54c59e2` (+ `4e9564f790`, `a7d2663e11`) | `f73dc001de` | — |
| AC-OWR-10 | **regression-guard** | 도착 시 초록 | `0ac54c59e2` (+ `4e9564f790`) | `f73dc001de` | — |
| AC-OWR-11 | release-blocking | 호스트 쪽 단언 수준, 주입 쪽 빌드 실패 형 | `5246f408cf` + `ae8868f6fc` | `cbf9579e9a` | `91e5bbcfa5` |
| AC-OWR-12 | release-blocking | 빌드 실패 형 | `ae8868f6fc` | `cbf9579e9a` | `91e5bbcfa5` |
| AC-OWR-13 | release-blocking | 단언 수준 | `42854e32aa` + `866a1f8103` | `89731b5e59` | `ba1ebac9a4` |
| AC-OWR-14 | release-blocking | 단언 수준 | `167a6c0af4` | `db68f30cf7` | `e6e3e8d9f2` |
| AC-OWR-15 | **regression-guard** | 도착 시 초록, 운영 코드 변이 넷과 시험 안 변이에서 붉음. chunk 5 에서 실행 파일 후보를 더하자 모드 절이 붉었다 (발견 2, 잠복 결함) | `415644baff` (+ chunk 5 `73e37d3832`) | `b896029228` (+ `9c00e81a2c`) | chunk 5 `051ddc18ef` |
| AC-OWR-16 | release-blocking | 단언 수준 | `2d6ec34053` | `0065cb21ed` | `4683bed8e0`, `d09d66c6f9`, `4767df4de4` |
| AC-OWR-17 | **regression-guard** | 도착 시 초록, 운영 코드 변이 넷과 시험 안 변이에서 붉음 | `5614767764` | `7414777ee3` | — |
| AC-OWR-18 | **regression-guard** | 도착 시 초록, 운영 코드 변이 셋과 시험 안 변이에서 붉음 | `ace2e8d84b` | `886f5b2d82` | — |
| AC-OWR-19 | **regression-guard** | 도착 시 초록, 기본값 뒤집기 포함 운영 코드 변이 셋에서 붉음 (기존 게이트 시험 둘도 같이) | `dd5a4b42b6` | `5cebd5a3da` | — |
| AC-OWR-20 | release-blocking | 단언 수준 (b), 빌드 실패 형 | `f9206f5242` + `da53289c5f` | `558dd9e6c2` | `f8936c2070` |
| AC-OWR-21 | release-blocking | 빌드 실패 형 | `f9206f5242` + `da53289c5f` | `558dd9e6c2` | `f8936c2070` |

release-blocking 14 (AC-OWR-01 ~ -07, -11 ~ -14, -16, -20, -21), regression-guard 7 (AC-OWR-08, -09, -10, -15, -17, -18, -19). 합 21.

**Binding run conditions 의 처분 (모든 항목, 각각 증거 위치).**

| 조건 | 처분 | 근거 |
|------|------|------|
| N4 잔여 (전체 트리 동일 요구 철회 뒤 센티널이 못 보는 것) | **처분됨 (시험이 닿는 데까지), 못 보는 것은 공개** | M8 기록 관측 3: E 안의 차이는 복원 영역 R (매니페스트) 의 바이트 비교, 실행 백업 배치 비교 (통제 실행 + 정확히 제거 자산 사본), 로그/상태 경로 집합 비교로, 파일 모드는 스냅샷 서술자와 R 의 `(i-mode)`, `(iii-mode)` 로 닫았다. 못 보는 것: 로그/상태 안의 내용 차이, 디렉터리 mtime·xattr·소유자, 프로젝트 밖 쓰기. 시험 안 변이 열여섯과 운영 코드 변이 넷으로 붉음을 관측 |
| N14 세대 간 우선순위 | 처분됨 (chunk 1) | 레코드를 가진 가장 새 세대가 결정한다 (보존 방향). 시험 `n14-newest-incomplete-older-complete` + 변이 `older-gen-outvotes` |
| N15 `DeployedHash` 일치 허용 | 처분됨 (chunk 1, 리더 승인) | 비교를 `TemplateHash` 하나로 좁혔다 (등록 경로와 같게). design.md D-2 의 「(또는 DeployedHash)」 문구보다 엄격한 편차로 이미 보고 |
| N16 실행 시점 조각으로 조립한 경로의 프로필 뿌리 읽기 | **열린 채 공개 (우회·재설정하지 않음)** | 카드 t1702 보류 범위. 이 카드는 프로필 뿌리를 열거나 읽는 코드를 더하지 않았다: 호출 지점 스캔 새 행 0, 보호 파일 둘과 `internal/profile` 불변 (위 제약 점검). 스캔·차분 카나리의 한계는 AC-OWR-08 기록에 그대로. **chunk 9 정정:** sync 감사 F1 이 사용자 심볼릭 링크를 통해 프로필 쪽 파일을 읽는 길 (등록 안 된 경로)을 찾았고 이 카드가 낸 것이어서 고쳤다 (`linkLeavesRoot`, 아래 기록). 등록 경로의 같은 사전 읽기는 기준 시점부터 있었고 AC-OWR-01 의 고정 기대 때문에 그대로 둔 공개 부채다. 이 부채는 상위 디렉터리가 링크인 형태에도 똑같이 적용된다 (chunk 10 (A) 가 등록 안 된 경로의 상위 링크 형태를 `pathLeavesRoot` 로 닫았다). 같은 부채가 업데이트 명령의 상대 뿌리 `"."` 에서 절대 경로 대상 링크의 `disposition` 해석에도 적용된다 (chunk 11 은 미등록 경로의 가드에만 상대 뿌리 정규화를 적용했다). chunk 12 는 미등록 파일의 읽기를 `os.Root` 로 옮겨 검사와 읽기 사이 경합을 닫았다 (등록 경로는 그대로, 위 결정). 즉 이 행의 「읽는 코드를 더하지 않았다」 는 F1 이전에는 참이 아니었다. |
| N17 「트리 해시」 가 내용만임을 밝히지 않음 | **처분됨** | AC-OWR-15/-17/-18 시험 파일 머리 주석 (AC-OWR-17/-18 은 시험 로그 첫 줄에도) 이 비교하는 해시를 내용만 (경로, 바이트, 모드, 디렉터리 존재) 이라고 밝히고, 디렉터리 mtime·시계 값은 비교하지 않으며 서로 다른 두 실행을 비교할 때 `deployed_at` 만 가린다고 적었다 |
| N18 AC-OWR-20 (a)(b) 가 규칙 7 조건 (ii) 를 요구받음 | **처분됨 (SPEC 문구 결함으로 공개, 본문 미수정)** | (a) 는 새 `manifest.prior.json` 이 실패하는 사례이므로 조건 (ii) 가 구조상 성립할 수 없다. 시험 `a-prior-manifest-copy-fails` 는 조건 (i)(iii) 에 해당하는 것으로 쓰기 대역이 정확히 한 번 불림 (Backup 단계 도달) 과 후보별 `backup-failed` 행을 단언하고, `b-candidate-bytes-copy-fails` 는 `rc.priorManifest == priorManifestWritten` 으로 (ii) 를 단언한다. 의도는 When 절의 괄호가 담는다 |
| N19 `.moai/config` 가 없을 때 사본 위치 | 처분됨 (chunk 1) | `preservePriorManifest` 가 설정 백업이 디렉터리를 만들지 않았으면 새 실행 디렉터리를 만든다. 시험 「a fresh run directory when the config backup made none」 |
| N20 건너뛴 `UserCreated` 항목에 `deployed_version` 없음 | 처분됨 (chunk 1) | `Track` 은 `template_managed` 만 도장한다. 시험 `DeployPathStampsTemplateManagedEntries` 와 `h1-...` |
| N21 REQ-OWR-005 문구와 AC-OWR-20 (b) 의 `removalBackupWriter` | **처분됨 (SPEC 문구 결함으로 공개, 본문 미수정)** | (b) 는 D-3 항목 5 가 정의하는 `priorManifestWriter` 대역이 실행 디렉터리 안에 막는 디렉터리를 만드는 방식으로 구현했다. REQ-OWR-005 의 「동결된 기존 접촉들」 대 「단일 기존 접촉」 은 시험이 한 행·개수 1 의 예외로 단언한다 |
| N22 AC-OWR-21 When 절에 구성 (o) 누락 | 처분됨 (chunk 1) | 시험이 (o) 를 가진다 (`o-normal-run-goes-through-temp-then-rename`); SPEC 문구 결함으로 공개 |
| N23 plan.md M8 이 AC-OWR-15 (v) 도달 단언을 빠뜨림 | **처분됨 (이 chunk)** | 시험 절 `ReachedReconciliation` 이 실행 1 의 `skipped == false`, 새 `manifest.prior.json`, 제거 목록이 비어 있지 않음 (후보 셋과 정확히 같음), 통제 실행의 도달과 무제거를 단언하고 변이 `MutantSkippedRunIsCaught`, `MutantControlRunThatRemovedFilesIsCaught` 가 붉다 |
| N24 변이 M-F (두 관측점을 정상으로 거친 뒤 최종 이름을 다시 쓰기) 미검출 | **확인됨, 못 닫음 (공개한 한계)** | 실측으로 `ok` 임을 확인했다 (chunk 1 변이 표). 접근 기록 이음매 없이는 닫을 수 없다 |

**닫힌 제품 쪽 사실 (chunk 9 F2).** 같은 시계 초에 시작한 두 갱신이 실행 백업 디렉터리를 공유해 뒤 실행의 `manifest.prior.json` 이 앞 실행의 것을 대체하던 손실은 `claimPriorManifestDir` 로 닫혔다 (각 실행의 사본이 다른 실행이 쓰지 않는 디렉터리에 놓임; 조사와 증거는 위 「F2」). 종전 AC-OWR-18 (c) 의 `SharedRunDirectory` 가지가 손실을 기대값으로 단언하던 것은 `NoWaitConsecutiveRuns` 가지가 두 세대 보존을 단언하는 것으로 교체되었다. 남은 것: 설정 백업 경로 자체의 같은 초 공유 (설정 파일 덮어쓰기, 실행 표지 교체) 는 이 카드 이전부터의 동작이다.

**chunk 5 (카드 리뷰 수정)로 바뀐 것.** 코덱스 카드 리뷰의 P2 둘을 고쳤다: 복원이 `manifest.prior.json` 을 `.moai/config` 로 복사하던 결함(이 카드가 만든 것)과 제거 후보 백업 사본이 실행 비트를 잃던 잠복 결함. 커밋 A `73e37d3832`, B `9c00e81a2c`, C `051ddc18ef`. 같은 부류로 확인했으나 고치지 않은 것: Legacy 복원 걷기는 `removed-assets/**`, `in-memory-backups/**` 도 `.moai/config` 로 복사한다 (앞의 것은 제거 스위치가 꺼져 있어 운영에서 비어 있고, 뒤의 것은 이 카드 이전부터). 보존 정책 센서스가 `manifest.prior.json` 을 내용으로 세는 문제(갭 G-K)는 REQ-OWR-017 때문에 건드리지 않았다.

**chunk 6 (둘째 카드 리뷰 지적)로 바뀐 것.** 판정 함수가 읽기 전에 파일 종류를 묻지 않아 공통 자산 걷기 아래의 FIFO 가 `moai update` 를 멈추게 했다. 판정의 유일한 진입점에서 종류를 먼저 확인하고 새 이유 `non-regular-file` 로 보존한다 (A `2eeb0559bd`, B `60d38d1559`, C `1d5ecbbc7c`). 이유 어휘가 하나 늘었다 — SPEC 본문 (design.md D-1) 개정이 필요한 사안으로 공개한다. (chunk 7: 같은 시험의 mkfifo 실패 때 건너뛰기를 `t.Fatalf` 로 바꿨다, `2b801c1971`. chunk 8: 마지막 남은 `t.Skipf` 였던 `migrate_ownership_test.go` 의 심볼릭 링크 탈출 사례를 `!windows` 파일로 옮기고 실패로 바꿨다, `4f970def44` — 카드가 바꾼 시험 파일에 건너뛰기 호출은 이제 없다.)

**SPEC 본문에 대한 공개 (본문은 고치지 않았다; 필요하면 manager-spec 재위임).**

1. REQ-OWR-006 / AC-OWR-10: 프로젝트 배포기는 에이전트를 쓰지 않고 사용자 자산 단계가 쓴다 (lane 결정 2, 공개된 부채로 유지).
2. REQ-OWR-012 / AC-OWR-17 「동일한 잔존 목록」 대 design.md D-2 의 이유 정제: 313개 복사본 픽스처에서 앞 실행의 이유 열이 어긋난다 (AC-OWR-17 기록 관측 1). 시험은 정상 상태부터 완전 일치, 그 전에는 한 가지 정제만 허용으로 단언한다.
3. AC-OWR-15 의 기대 차이 집합 E 와 AC-OWR-17 의 제외 목록이 시계 파생 위치 둘 (`.moai/backups/update-*`, `.moai/archive/files/update-migration/*`) 을 빠뜨렸다 (AC-OWR-15 기록 관측 1).
4. N18, N21, N22 의 문구 결함 (위 표).
5. N15: 증거 비교를 `TemplateHash` 하나로 좁힌 것은 design.md D-2 문구와 다르게 더 엄격하다 (리더 승인됨).

**sync 단계가 알아야 할 사용자 가시 부작용 (lane 결정 1, `update.go` 변경 둘).** (a) 다른 업데이트가 업데이트 잠금을 들고 있는 동안에도 `moai update --dry-run` 이 `ErrUpdateLockHeld` 로 거절되지 않고 진행하며 그 실행의 중간 상태를 읽을 수 있다. (b) `moai update --dry-run` 은 새 버전 확인과 이진 자체 업데이트 단계에 더 이상 닿지 않는다.

**이월한 갭 (CI 또는 리더의 착지 재측정 몫).** `internal/cli` 패키지 전체 시험과 패키지 전체 커버리지; `-race`; Windows 에서의 시험 실행 (빌드와 vet 만 했다); 대소문자 구분 파일시스템에서의 `TestCaseVariantOnHostFilesystem` 의 다른 가지; 이 카드가 만든 시험이 무거운 `internal/cli` 에서 병렬 실행 부하로 시간 초과를 내는지의 측정.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-18
sync_commit_sha: e99a990296dc8d84b48891a9d916b8379b9f4f3b  # the sync commit could not cite its own hash (it carried pending-backfill); this value is backfilled in the following commit
sync_status: complete
b12_self_test_a: pass  # pre-emission grep: `grep -c 'SPEC-UPDATE-OWNERSHIP-REPAIR-001' CHANGELOG.md` printed 0 before the entry and 1 after
b12_self_test_b: pass-with-disclosure  # ac_source=.moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/acceptance.md (tier L). The B12 counter printed 22 (live=22 excluded=0 ambiguous=0): the 21 AC-OWR-01..21 criteria (21 `### AC-OWR-` headings) plus AC-COLLECTOR-ANCHOR-001, an unmarked citation of another SPEC's criteria inside the criterion-grade paragraph (line 19). Marking it `[REF]` needs an acceptance.md body edit, which is not this commit's to make. The CHANGELOG says 21.
b12_self_test_c: pass  # every file path named in the CHANGELOG entry checked with ls (spec.md, plus the implementation files read for the entry)
changelog_entry_position: "[Unreleased] > Added, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed on this single sync commit; updated: 2026-10-18"
  plan_md_acceptance_md: "untouched (frontmatter carries no updated: field and no status:)"
canary_compliance_check: not-applicable
```

This sync commit edits `CHANGELOG.md`, this section, and the `spec.md` frontmatter `status:` only. It does not edit any SPEC body (`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`) and does not touch §E.1-§E.3.

**Caveats for the sync auditor and the leader (user-visible side effects and open limits).**

1. Default output: without `--verbose`, `moai update` no longer prints a per-file line for each kept file; the grouped Reference rows (one line per reason with a count) carry them. Source: `internal/cli/migrate_ownership_remaining.go` (`remainingAwareReport`, `recordRemainingList`).
2. Dry run and the update lock: `moai update --dry-run` takes no update lock (`internal/cli/update.go`), so a dry run is no longer refused while another update holds the lock and its result may then reflect that run's intermediate state. A one-line warning in the dry-run output was considered and is NOT implemented: carried as a disclosed follow-up/debt.
3. Same-second backup directory: CLOSED by `fa26fcaf17` (sync-audit finding F2). When the intended run backup directory already holds `manifest.prior.json`, the start-of-run copy goes to a sibling `<name>-02` .. `<name>-99` created with Mkdir (so it never writes into a directory another run owns), and the run marker is copied so retention still counts one run; if no sibling is free the run fails closed (removes nothing); a sibling that cannot be given its run marker is removed and the claim fails closed (the run removes nothing), so no unmanaged residue remains (`90296319b4`). What remains: the config backup path itself is still shared by two same-second runs (the later run's config files replace the earlier's and the run marker is rewritten); that behaviour predates this card and is untouched.
4. Legacy restore walk: `RestoreMoaiConfigLegacy` also copies `removed-assets/**` and `in-memory-backups/**` into `.moai/config` (found by the codex card review, not fixed here). `removed-assets` belongs to card t1668 and is empty while the removal switch is off; `in-memory-backups` predates this card.
5. Keep-reason vocabulary: the closed vocabulary gained `non-regular-file` (code, `internal/cli/migrate_ownership.go`). The design.md D-1 closed keep-reason list now contains `non-regular-file`, added by manager-spec in commit `0a33265120` (design.md only, one line). The earlier E.2 statements (the chunk-6 accounting near line 3103 and any other "design row pending" wording) were true when written and are superseded by this backfill commit; those E.2 lines are not edited (append-only record). The sync commit `e99a990296` itself edited no SPEC body.

   SPEC-body edit set for the audit brief: design.md only, by manager-spec, `0a33265120`. Since the plan phase (last plan-phase commit `073b3e1382`), spec.md changed only by status transitions (`7d8be68f6c` draft to in-progress, `e99a990296` to completed); plan.md, acceptance.md and research.md are unchanged since `073b3e1382`.
6. Uncovered races: the EvalSymlinks-to-read race on the UNREGISTERED read is closed by `os.OpenRoot` / `Root.Open` (`dc32a94c9e`). Still uncovered: a regular file swapped for a FIFO between the kind check and the open, and the REGISTERED-file read-before-boundary at base (N16 family debt). (§E.2 chunk 6 accounting item 3 covers the sibling kinds `os.ModeIrregular`, sockets and devices, which take the same branch but are not exercised by a test.)
7. `removalEnabled` is `false` in production: no project file is removed by this change.
8. Legacy restore skip: only ROOT-LEVEL backup entries named `manifest.prior.json`, or starting `manifest.prior.json.tmp-`, are skipped by the legacy restore, so a user config file with such a root-level name is not restored; files below a directory with such a name, and the same names nested in a subdirectory, are restored (`dcbafe568c`). Declined review finding; only the unshipped restore bug could have created such a root-level file, and the name is fixed by the SPEC; see §E.2 chunk 10.
9. Limit: the `os.Root` escape detection matches the standard library error text `path escapes from parent`. If that text ever changes, the read fails as `read-failed` (nothing is read) and the race test turns red.
10. Limit (audit finding F-C, declined): on a case-insensitive file system a link target spelled with different case than the real root compares as outside the project; this fails closed (the file is kept), and the disposition compared strings the same way at base.

**Sync-audit round 1 and the repair (record at the time of this commit).** The independent sync audit at HEAD `dfd7c40fb2` returned FAIL (required backend codex failed), receipt `rcpt-7770ccadc2f78594b050e777`. Findings: F1 BLOCKING (an unregistered symlink whose target is outside the project root was read before the boundary check) and F2, F3 SHOULD-FIX; the lane accepted all three and fixed each test-first: F1 by `e0c2896d64` (Lstat plus EvalSymlinks before any read; the link is kept unread with reason `symlink-escape`; in-project and dangling links unchanged), F3 by `d8fa645c41` (a non-regular user-root counterpart is never opened and counts as unconfirmed), F2 by `fa26fcaf17` (caveat 3). F4 (empty nested directories next to a live sibling file) was declined as outside REQ-OWR-008. The re-audit is pending at the time of this commit. Advisory card review round 4 (at HEAD `b5ffc48971`, three P2 findings) then led to two further fixes: `848bbc4707` (an unregistered file is kept unread with reason `symlink-escape` when its fully resolved path leaves the project, symlinked ancestor directories included; `pathLeavesRoot` resolves the whole path with EvalSymlinks) and `d8a5f75053` (a failed prior-manifest rename removes its temporary file, and the restore walks skip `manifest.prior.json.tmp-*` at the backup root like the final name, via `backup.PriorManifestTempPrefix` / `IsPriorManifestBookkeeping`). The third P2 was declined as a disclosed limit (caveat 8); the reasoning is in §E.2 chunk 10. Independent delta sync audit round 2 (HEAD `38c8ca5376`) returned FAIL (required backend codex failed), receipt `rcpt-aeb48f1d567c915a06f7b90f`; it verified F1/F2/F3 fixed and found N1 (the temp-name filter dropped user files below a directory named with the `manifest.prior.json.tmp-` prefix), N2 (a sibling that could not be given its run marker stayed out of retention) and N4 (caveat 8 understated N1). Fixes: N1 `dcbafe568c` (the bookkeeping-name filter matches root-level entries only), N2 `90296319b4` (such a sibling is removed and the claim fails closed), N4 this wording (caveat 8 above), plus boundary fixes found by the frozen counterexample table (`TestPathBoundaryCounterexamples`, 43 rows) in `1003a9edd4` (root containment compares path elements, not string prefixes; a relative project root such as the update command's "." is made absolute and resolved before comparing). N3 and N5 are carried as already-disclosed debt. Independent delta sync audit round 3 (HEAD `1d9b06a737`) returned FAIL (required backend codex failed), receipt `rcpt-665b06f8d214d9a97633b90c`; the auditor could run no tests at load 155 and did not re-check the docs. Findings: F-A (SHOULD-FIX, a link swapped between the boundary check and the read could make the unregistered read land outside the project), F-B (SHOULD-FIX, a marker write failing after the marker file existed left an empty `.moai-run`, so the sibling directory leaked) and F-C (MINOR, case-insensitive file system, caveat 10). F-A and F-B are fixed in `dc32a94c9e`: the unregistered file is read through `os.OpenRoot` on the project root plus `Root.Open`, confined to `classifyUnregistered`, `readUnregisteredInsideRoot`, `errReadEscapesRoot` and the test seam `readWindowHook` (the registered-file read is unchanged); `claimPriorManifestDir` removes the sibling it just created and never the shared run directory, and the claim still fails closed. F-C is declined as a disclosed limit. Tests and records: `714adcc444`, `53a6c88b09`, `28b2882688` (§E.2 chunk 12). Audit round 4 is pending at the time of this commit and is the last round before a hold under rule C. The final run code commit is `dc32a94c9e`.

**Disclosed debt (N16 family, carried from §E.2 chunk 9).** The registered-file read-before-boundary at the base commit is unchanged: a registered managed file that is a symlink out of the project has its target bytes read for the hash comparison (removal is refused as `symlink-escape`, so it is never a removal candidate or a backup copy). A pre-read guard there would change the pinned AC-OWR-01 symlink-escape class expectation (`managed` plus `symlink-escape`), and the lane directed that the AC not be changed. Reading the code is the only evidence for this path; it was not re-measured by execution.

**SPEC-wording debts for the sync auditor's brief (bodies not edited; disclosed in §E.3).**

- REQ-OWR-006 / AC-OWR-10: the project deployer never writes agents; the user-asset phase does, so AC-OWR-10 tests agents where they are delivered.
- AC-OWR-21: the command path, and the missing case (o) (the test carries `o-normal-run-goes-through-temp-then-rename`).
- AC-OWR-20 (b): `removalBackupWriter` is named, while D-3 item 5 defines `priorManifestWriter`; the test follows D-3.
- AC-OWR-06 (b), AC-OWR-08 glob, and AC-OWR-12 reason referent: wording points named in the lane's sync brief (the exact referents were not re-derived in this phase); the tests assert the intent.
- Also from §E.3: REQ-OWR-012 / AC-OWR-17 "same remaining list" versus the D-2 reason refinement; the expected-difference set of AC-OWR-15 and the exclusion list of AC-OWR-17 omit two clock-derived locations; N15 narrows the evidence comparison to `TemplateHash` (leader-approved).

**Binding-run-condition dispositions:** as recorded in §E.3 ("Binding run conditions 의 처분", N4 residual through N24); not re-judged here. N16 is open and disclosed, N24 is a confirmed limit.

**Honest gaps.**

- Go commands run between 2026-10-18T15:17Z and 16:50Z (the chunk 12 agent's calls and the lane's reproduction calls) were started after load and free-memory checks but WITHOUT a check of the landing slot condition of the factory admission rule (board d-20261018T151201Z-8c1e); their results are a functional record, not evidence of admission compliance; the lane's reproduction calls (16:33:28Z-16:43:14Z) fall inside a landing occupancy.
- Not run: the full `internal/cli` package test, package-wide coverage, and `-race` (carried to CI or the leader's landing remeasurement).
- Windows: build and vet only; no test execution on Windows.
- Gap G-K (the retention-policy census counting `manifest.prior.json` as content) is unmeasured and untouched because of REQ-OWR-017.
- Case-sensitive file system branch of `TestCaseVariantOnHostFilesystem` is not exercised on the host.
- Code-review findings and their fixes (from §E.3): chunk 5, two findings (restore copied `manifest.prior.json` into `.moai/config`; a removed executable's backup copy lost its mode), fixed in `051ddc18ef`; chunk 6, the FIFO stall (the verdict read before checking the file kind), fixed in `1d5ecbbc7c`, with the chunk 7 and chunk 8 follow-ups (`2b801c1971`, `4f970def44`) making the two tests fail instead of skip.
- Sync-phase checks were limited by instruction: no Go test package was run in this phase, only `go build` and `moai spec lint`; the judging build recorded `vcs.revision=478285ab4c1fc893b020bb464d14cb69c8422523` with `vcs.modified=true` against the tree HEAD `175407599b`, so that stamp is not the tree HEAD (the binary was built from this worktree's files).

## §G Override and Refusal Record

- 2026-10-18T04:53:21Z SPEC-UPDATE-OWNERSHIP-REPAIR-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter4-delta.md
- 2026-10-18T04:53:26Z SPEC-UPDATE-OWNERSHIP-REPAIR-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter4-delta.md
