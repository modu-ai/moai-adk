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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-18T04:53:21Z SPEC-UPDATE-OWNERSHIP-REPAIR-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter4-delta.md
- 2026-10-18T04:53:26Z SPEC-UPDATE-OWNERSHIP-REPAIR-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865/.moai/reports/t1865/plan-audit-iter4-delta.md
