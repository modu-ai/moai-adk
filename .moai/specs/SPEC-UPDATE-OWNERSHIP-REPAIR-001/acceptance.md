---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "acceptance criteria"
version: "0.3.1"
created: 2026-10-18
---

# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — 인수 기준

## 예산과 규칙

- Tier L 상한: 요구사항 25, 인수 기준 25. 이 파일은 **21** 개의 기준 (AC-OWR-01 ~ -21) 을 가진다. 요구사항은 17 개이며 각각 하나 이상의 기준을 가진다 (spec.md § E).
- 판정 명령은 모두 **단일 호출 형태**이다 (파이프, 리다이렉션, `&&` 없음). 시험 명령의 `TMPDIR` 는 `/private/tmp` 아래 카드 토큰이 없는 이름 `/private/tmp/own-fixture-d1` 이며, **이 디렉터리는 사전 점검 P0 (`mkdir -p /private/tmp/own-fixture-d1`, plan.md § C 와 M0 의 첫 단계) 가 만든다** — 명령 자체는 만들지 않는다 (단일 호출 형태). 모든 시험은 임시 HOME 과 `t.TempDir()` 만 쓴다. 실제 홈, mo.ai.kr, `moai update` 재실행, 바이너리 교체는 쓰지 않는다.

### 등급과 RED 셀 규약

이 규약이 이전 판본의 모순 — 본문 14행은 RED 셀을 run 단계로 미루면서 기준마다 "릴리스 게이트" 라고 적었다 — 을 닫는다. 근거: `.claude/rules/moai/development/verification-completeness.md` § 2 와 § 2.1.

1. **등급은 둘이다.** `release-blocking` 은 RED-now 셀의 네 요소 (명령, 그 명령의 원문 stdout, 종료 코드, 트리 SHA) 가 **이 트리에서 실측되어** 증거 원장 (§ L 또는 progress.md § E.2) 에 있는 기준이다. 그렇지 않은 기준은 `regression-guard` 이다. 재실행할 수 없는 RED 를 가진 기준은 release-blocking 자격을 잃고 regression-guard 로 분류되며 통과로 기록되지 않는다 (§ 2.1 의 미결정 처분). 같은 처리의 선례: SPEC-AC-COLLECTOR-ANCHOR-001 acceptance.md 「등급 분류」 — 실측된 RED 가 있는 기준만 release-blocking, 나머지는 「run 단계가 그 픽스처를 만들기 전에는 재실행할 RED 가 트리에 없다. 없는 RED 를 있다고 적지 않는다」.
2. **이 plan 시점에는 판정 시험이 트리에 없다** (새 파일은 plan.md § F 의 표). `-run '^TestX$'` 는 제로 매치로 `ok ... [no tests to run]` 종료 코드 0 을 낸다 — RED 도 통과도 아니다. 그래서 **21개 기준 모두 plan 시점에는 regression-guard 이며, 기준마다 적힌 「승격 지점」 에서 release-blocking 으로 승격된다.** § L 에는 이 트리에서 실제로 돌린 전제 셀과 기준선 셀만 있다 (이 트리의 기존 사실에 기대는 전제와 GREEN 가드). RED 셀은 한 칸도 미리 채우지 않았다.
3. **승격 규칙 (기준마다 승격 지점 = 해당 마일스톤).** (A) 그 기준의 시험 파일만 담은 커밋, (B) A 의 트리 SHA 에서 판정 명령을 돌려 명령·원문 stdout·종료 코드·트리 SHA·실패 이유를 progress.md § E.2 에 기록하는 커밋, (C) 구현 커밋 — 이 순서. 순서의 증거는 커밋 그래프이다 (`verification-claim-integrity.md` § 2.3: 같은 커밋에 묶인 기준선은 순서 주장을 영구히 검증 불가로 만든다). A 의 RED 는 시험이 존재하는 상태의 실패 (빌드 실패 또는 단언 실패) 여야 하며 「올바른 이유로 붉은가」 를 셀에 적는다. 승격된 기준만 완료 조건 (아래 Definition of Done) 에 산입된다.
4. **제로 매치 셀렉터는 통과도 RED 도 아니다.** `-run` 이 아무 시험도 고르지 못해 `[no tests to run]` 과 `ok` 를 내면 해당 기준의 실패이다. 판정은 `--- PASS: <시험 이름>` 줄이 보이고 쓸어낸 시험 수가 기대와 같을 때만 성립한다.
5. **변이 시험.** 보존 기준마다 "이 변이가 통과하면 기준이 얕다" 는 변이를 적었다. 변이는 가능한 곳에서 **구현을 주입하는 형태로 시험 안에서** 실행한다: 단언은 문제 목록을 돌려주는 순수 함수이고, 시험은 그 목록을 `t.Errorf` 로 보고하며, 메타 시험은 변이 입력 앞에서 목록이 비어 있지 않음을 단언한다 (시험이 변이의 적색을 직접 관측한다). 입력 쪽 변형만으로 하는 변이는 그렇게 적었다.
6. 공통 시험 헬퍼: `buildMigrationFixture`, `installMigrationUserCounterparts`, `profileReachSandbox`, `enableRemovalForTest`, `runUpdateFlowWithProgress`, `removalBackupCopy`, `runDryRunUpdateInProject`, `hashDryRunTree` (모두 이 트리에 있다), 그리고 plan.md M0 이 만드는 `migrate_ownership_helpers_test.go` (`runUpdateFlowForced` 포함).
7. **조정에 닿는 실행만 실행으로 센다 (공허 방지, 감사 N5).** 갱신은 같은 버전의 일반 재실행을 조정 단계 앞에서 건너뛴다 (`update_template_sync.go:1239-1248`; 이번 실행 측정 § L-13: 일반 실행 1 은 건너뛰지 않고 일반 실행 2 는 건너뛰며 `--force` 실행은 건너뛰지 않는다). 따라서 여러 번 실행하는 기준 (-10, -14, -17, -18) 은 `--force` 로 조정에 닿게 하는 헬퍼 `runUpdateFlowForced(t, root) (out string, skipped bool)` 로 구동하고, **매 실행마다** (i) `skipped == false`, (ii) 그 실행의 실행 백업에 새 `manifest.prior.json` 이 있고 바이트가 실행 전 매니페스트와 같다 (Backup 단계 도달), (iii) 잔존 행 개수가 픽스처가 기대하는 양수이다 (마이그레이션 도달) 를 단언한다. 그리고 쓸어낸 실행 수를 단언한다 (N 번 구동하면 위 세 조건을 만족한 실행이 정확히 N). 이 단언 없이 "실행했다" 로 읽지 않는다. 건너뛰는 실행도 운영 사실이므로 기준마다 **건너뛰는 실행 사례**를 따로 둔다 (`skipped == true`, 트리 해시·매니페스트 바이트·출력 불변).
8. **보존 기준은 도달 대조군을 가진다.** 보존을 단언하는 기준 (-02, -03, -04, -07, -14) 의 픽스처에는 같은 실행에서 **제거되어야 하는 형제 파일** — 모든 제거 전제를 갖춘 managed 사본 — 이 하나 이상 있고, 그 사본이 제거됨을 보존 단언보다 먼저 단언한다. 제거 암에 닿지 못해 보존이 공허하게 성립하는 일을 막는다.
9. **모든 기준의 변이 표는 § M** 이다. 변이가 시험을 붉게 하지 못하는 기준은 고쳤고 고친 내용을 표에 적었다.

## 인수 기준

### AC-OWR-01 — 판정 함수의 네 분류와 처분 *(maps REQ-OWR-001, REQ-OWR-002)* (항목 (1))

- **Given** 같은 스킬 루트 아래의 파일 매트릭스: (항목 있음 × `Provenance` ∈ {`manifest.TemplateManaged`, `manifest.UserModified`, `manifest.UserCreated`, `manifest.Deprecated`} × 바이트가 `TemplateHash` 와 같음/다름) 의 모든 조합과 항목 없음. 대표 넷은 m (등록 managed, 해시 같음), u (같은 등록, 바이트 다름), c (`user_created`), x (항목 없음). **항목 없음에는 두 변형이 있다: x1 (어떤 템플릿 파일과도 바이트가 다름), x2 (내장 템플릿의 실재 같은 경로 파일과 바이트가 같음).**
- **When** 판정 함수 `classifyOwnership` 가 매트릭스의 모든 파일을 판정한다.
- **Then** 분류는 m=`managed`, u=`user-modified`, c=`user-created`, x1=x2=`unregistered` 이다 (x2 가 `managed` 가 되는 구현 — 해시 일치로 승격 — 은 이 기준만으로도 붉다). 매트릭스의 모든 조합에서 돌려준 분류가 값으로 네 이름의 집합 안에 있고 네 값이 모두 한 번 이상 나온다 (상수 개수를 세지 않는다 — 다섯째 값은 어느 조합에서 돌려질 때 비로소 문제이며 매트릭스가 모든 조합을 먹인다). 처분은 spec.md § B 의 정의를 따른다: m 이 `removal-eligible` 이 되려면 (프로필 도달 유지 아님) ∧ (사용자 대응물 확인) ∧ (제거 스위치 켬, 픽스처에서) ∧ (루트 경계 안) 이 모두 필요하고, 그중 하나씩만 빼면 각각 `keep(<이유>)` 이다 (스위치 끔 → `keep(removal-gate-off)`). 나머지 분류는 항상 `keep(<이유>)` 이다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipVerdictClasses$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipVerdictClasses (…)`, 종료 코드 0.
- RED 이유 (A 커밋 뒤에 셀에 적을 것으로 예상): 판정 함수가 아직 없어 빌드 실패 (`undefined: classifyOwnership`, 종료 코드 1).
- 변이: x 를 `user-modified` 로 분류하는 구현, x2 를 `managed` 로 올리는 구현, 어떤 조합에서 다섯째 값을 돌려주는 구현, 스위치 조건을 판정에서 빼는 구현 — 각각 분류·처분 단언이 붉어진다 (§ M).
- 등급: regression-guard → 승격 지점 **M1a**.

### AC-OWR-02 — 보존 픽스처 1: 내장 템플릿의 실재 경로와 바이트가 같은 미등록 파일 *(maps REQ-OWR-002, REQ-OWR-003)* (항목 (6))

- **Given** 매니페스트 항목이 없고, 바이트가 **내장 템플릿에 실재하는 같은 경로의 파일**과 동일한 파일 두 개 — `.claude/skills/moai-ref-owasp-checklist/SKILL.md` (템플릿 9120 바이트) 와 `.agents/skills/moai-plan/SKILL.md` (템플릿 533 바이트). 두 경로가 이 트리의 내장 템플릿에 실재함은 § L-1 에서 측정했다 (종전 판본의 `moai-user-authored/SKILL.md` 는 § L-2 에서 어느 쪽 템플릿에도 없음이 측정되어 철회했다). 두 경로의 사용자 대응물을 `installMigrationUserCounterparts` 를 경로 지정으로 확장한 M0 도우미로 설치한다. 제거 스위치는 픽스처에서 켠다. 픽스처의 매니페스트는 이 두 경로의 항목을 담지 않는다 (`buildMigrationFixture` 가 등록하는 경로는 `migIdenticalSkill`, `migModifiedSkill` 뿐이다).
- **전제 단계 (시험의 첫 단언).** `fs.ReadFile(embedded, <경로>)` 가 두 경로 모두 성공하고, 시험이 프로젝트에 쓴 바이트가 그 결과와 같다. 이 단언이 실패하면 시험은 이후를 실행하지 않고 실패한다 — 템플릿 조회가 빗나가는 픽스처로 변이가 발화하지 않은 채 통과하는 일이 없다.
- **도달 대조군 (규칙 8).** 같은 프로젝트에 `buildMigrationFixture` 의 등록된 managed 사본 (`migIdenticalSkill`, 사용자 대응물 확인됨) 이 있고, 같은 실행에서 제거됨을 먼저 단언한다.
- **When** 마이그레이션이 두 번 실행된다: (A) 운영 증거 원천으로, (B) **음성 대조군** — 파일 자신의 바이트와 내장 템플릿 트리의 일치만으로 세 증거를 모두 "있다" 고 답하는 대역 `provenanceSource` (잘못된 승격 구현) 를 패키지 수준 변수로 꽂아 (design.md D-2).
- **Then** (A) 두 파일은 바이트 그대로 남는다. 분류는 `unregistered`, 이유는 `no-provenance-record` 이고 잔존 행의 빠진 증거 항목이 `deployed-version`, `original-deployed-path`, `prior-manifest-record` 셋 모두이다. 제거 목록과 실행 백업의 `removed-assets/` 에 두 파일이 없다. 매니페스트에 두 파일의 새 항목이 생기지 않는다. (B) 같은 보존 단언 함수가 **비어 있지 않은 문제 목록**을 돌려준다 (변이 구현은 이 픽스처에서 두 파일을 `managed` 로 승격해 제거하므로 붉게 잡힌다). 이 적색을 시험이 직접 관측한다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveByteEqualUserFile$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipPreserveByteEqualUserFile (…)` 와 서브 시험 줄 셋 — `    --- PASS: TestOwnershipPreserveByteEqualUserFile/Premise (…)`, `    --- PASS: TestOwnershipPreserveByteEqualUserFile/Preserve (…)`, `    --- PASS: TestOwnershipPreserveByteEqualUserFile/MutantIsCaught (…)` — 종료 코드 0.
- RED 이유: 이유 코드와 빠진 증거 항목이 아직 기록되지 않아 단언 실패 (파일 보존 자체는 현재도 성립한다 — `TestMigrationClassifiesUnregisteredMirrorCopy`, § L-4).
- 변이: 바이트 일치와 확인된 대응물만으로 `managed` 로 보고 제거하는 구현 — 위 (B) 가 시험 안에서 이를 주입해 적색을 관측한다.
- 등급: regression-guard → 승격 지점 **M1a**.

### AC-OWR-03 — 보존 픽스처 2: 사용자가 수정한 파일 *(maps REQ-OWR-002)* (항목 (1))

- **Given** 매니페스트에 `template_managed` 로 등록된 뒤 사용자가 바이트를 바꾼 스킬 파일 (`TemplateHash` 와 다름), 그리고 `user_modified` 로 등록된 에이전트 파일, 제거 스위치 켬. **도달 대조군 (규칙 8)**: 같은 프로젝트의 등록된 managed 사본 하나 (대응물 확인됨) 가 같은 실행에서 제거된다.
- **When** 마이그레이션이 실행된다. 같은 입력을 **현재 바이트를 무시하고 기록된 `Provenance` 만 보는 분류기를 주입한** 대조 실행이 한 번 더 있다.
- **Then** 도달 대조군 사본이 제거된 것이 먼저 확인된다. 두 파일은 바이트 그대로 남고, 분류는 `user-modified`, 이유는 `user-modified` 이며, 잔존 행에 오른다. 주입한 변이 분류기 앞에서는 보존 단언 함수가 **비어 있지 않은 문제 목록**을 돌려준다 (수정된 파일이 제거되어 붉어짐을 시험이 직접 관측).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveUserModifiedFile$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipPreserveUserModifiedFile (…)` 와 서브 시험 `Reach`, `Preserve`, `MutantIsCaught`, 종료 코드 0.
- RED 이유: 분류와 이유 코드가 잔존 행으로 기록되지 않아 단언 실패 (보존 자체는 `TestRF2F3_MigrationPreservesPostInstallEdits` 가 이미 지킨다).
- 변이: 현재 바이트 대신 기록된 `Provenance` 만 보고 `managed` 로 취급 — 위 대조 실행이 시험 안에서 이를 주입해 적색을 관측한다.
- 등급: regression-guard → 승격 지점 **M1a**.

### AC-OWR-04 — 보존 픽스처 3: 출처 불명 미등록 파일 *(maps REQ-OWR-003)* (항목 (6))

- **Given** 매니페스트 항목이 없고 어떤 템플릿 파일과도 바이트가 다른 파일들 — `.agents/skills/moai-custom-x/SKILL.md`, `.claude/agents/moai/custom-agent.md` — 그리고 템플릿 이름과 같은 경로이지만 바이트가 다른 `.agents/skills/moai-plan/extra.md`. **도달 대조군 (규칙 8)**: 등록된 managed 사본 하나 (대응물 확인됨) 가 같은 실행에서 제거된다.
- **When** 마이그레이션이 `--verbose` 와 함께 실행된다. 같은 입력을 **파일 이름이 템플릿 디렉터리에 있는 이름이면 `managed` 로 올리는 대역 원천**을 꽂은 대조 실행이 한 번 더 있다 (AC-OWR-02 (B) 와 같은 주입).
- **Then** 도달 대조군 사본이 먼저 제거된다. 세 파일은 보존된다. 각 파일의 잔존 행은 분류 `unregistered`, 이유 `no-provenance-record`, 빠진 증거 항목을 가진다. `--verbose` 출력이 각 파일 경로와 이유를 이름으로 댄다. 대역 원천 앞에서는 이유·분류 단언이 비어 있지 않은 문제 목록을 돌려준다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipUnregisteredUnknownOriginKeptWithReason$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipUnregisteredUnknownOriginKeptWithReason (…)` 와 서브 시험 `Reach`, `Reasons`, `MutantIsCaught`, 종료 코드 0.
- RED 이유: 현재 출력은 경로만 대고 이유 코드와 빠진 증거 항목이 없다.
- 등급: regression-guard → 승격 지점 **M1a**.

### AC-OWR-05 — 양성 픽스처 4: 증거가 완비된 파일만 managed *(maps REQ-OWR-002, REQ-OWR-015)* (항목 (6))

- **Given** 시험이 직접 공급하는 증거 원천 (`provenanceSource` 대역 — 운영 원천이 아니다). 이전 세대 매니페스트는 **실제 출처값 상수로 만든다** (`manifest.TemplateManaged`, `manifest.UserCreated` 등 `internal/manifest` 의 상수 — 문자열 리터럴 `"template_managed"` 를 손으로 쓰지 않는다; 이 트리의 상수는 § L-15). 양성 기준: 이전 세대가 경로 P 를 `manifest.TemplateManaged` 로, 배포 버전 V, 템플릿 해시 H 와 함께 기록하고, 현재 매니페스트에는 P 의 항목이 없으며 (항목 유실), 디스크의 P 는 해시 H 의 바이트를 가진다.
- **When** 판정 함수가 P 를 판정한다.
- **Then** 양성: 분류는 `managed` (증명된) 이고 증거 기록에 배포 버전 V, 원래 배포 경로 P, 이전 기록의 세대가 들어 있다. 기본 처분은 `keep(removal-gate-off)` 이고 `removal-eligible` 은 제거 스위치를 픽스처에서 켜고 사용자 대응물이 확인된 때만이다. **매니페스트는 바뀌지 않는다** (자동 채택 없음). 변이 일곱 가지와 운영 원천 한 건:
  - (a) 기록에서 배포 버전을 뺀 경우 → `unregistered` + `evidence-incomplete:deployed-version`,
  - (b) 같은 해시의 기록이 다른 경로 P' 키로만 있는 경우 (재배치된 파일) → `evidence-incomplete:original-deployed-path`,
  - (c) 이전 세대에 P 도, 같은 해시의 다른 경로 기록도 없는 경우 → `no-provenance-record`,
  - (d) 디스크 바이트가 H 와 다른 경우 → `record-hash-mismatch`,
  - (e) 이전 기록의 출처값이 `manifest.UserCreated` (버전·경로·해시는 양성과 모두 같음) → `unregistered` + `evidence-incomplete:prior-manifest-record`,
  - (f) 출처값이 `manifest.UserModified` → 같은 결과,
  - (g) 출처값이 `manifest.Deprecated`, `manifest.GeneratedManaged`, 빈 문자열, 알 수 없는 문자열 `"zzz"` 각각 → 같은 결과,
  - (h) **운영 증거 원천을 거친 두 건.** (h1) 보존: 실제 배포기가 매니페스트에 없는 기존 파일 P (템플릿과 바이트 같음) 를 만나 `UserCreated` 로 기록하고 건너뛴 상태 (시험이 그 기록의 `Provenance == manifest.UserCreated` 를 먼저 단언한다 — `deployer.go:265-267` 의 실제 값) 를 실행 백업의 `manifest.prior.json` 으로 운영 쓰기 함수로 남기고 현재 매니페스트에서 P 를 뺀 뒤 운영 이전 세대 원천이 읽은 결과로 판정 → `unregistered`. (h2) 양성: 실제 배포기가 P 를 새로 배포해 `TemplateManaged` + 이번 실행의 `deployed_version` 이 찍힌 항목을 운영 쓰기 함수로 남기고 같은 방식으로 항목을 뺀 뒤 운영 원천으로 판정 → `managed` (증명된), 처분 `keep(removal-gate-off)`.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipAttestedManagedOnlyWithFullEvidence$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence (…)` (서브 시험: 양성 1 + 변이 (a)-(g) 와 (h1) (h2)), 종료 코드 0.
- RED 이유: 증거 해석기가 없어 빌드 실패.
- 변이 (구현): 출처값을 무시하고 버전·경로·해시만 보는 해석기 — (e)(f)(g)(h1) 이 붉다 (감사 N2 의 codex 재현과 같은 변이).
- 비고: (a)-(g) 는 **함수 수준**이며 해석기와 인터페이스가 M1a 에 있으므로 Q2 의 기록된 대안 (나) 로 풀려도 남는다. (h) 는 운영 쓰기·읽기를 쓰므로 M1b 이후에만 도달 가능하다 (plan.md § G.1: (나) 에서는 삭제).
- 등급: regression-guard → 승격 지점 **M1a** ((h) 는 **M1b**).

### AC-OWR-06 — 배포 버전 기록, 이전 매니페스트 원본 보존, 이전 세대 원천 *(maps REQ-OWR-015, REQ-OWR-017)* (항목 (6))

- **Given** (a) `internal/manifest` 에서 `Track` 으로 항목을 기록하는 매니페스트 매니저, 그리고 `deployed_version` 필드가 없는 기존 JSON (옛 항목 둘 포함). (b) 갱신 흐름을 세 번 도는 픽스처: 실행 1 (일반, 버전 V1), 실행 2 (`--force`, 시험이 `pkg/version.Version` 을 V2 로 바꾸고 복원), 실행 3 (`--force`, 같은 V2). 사용자 수정 파일 하나가 있어 실행 2 가 그 파일을 다시 배포하지 않는다. 시작 매니페스트 파일은 **정규 직렬화와 다른 모양**(들여쓰기 4칸, 키 순서 뒤바뀜, 줄 끝 공백)으로 쓴다. (c) 실행 백업 디렉터리 여러 개에 보존된 `manifest.prior.json` 들 — 정상 사본, 형식이 깨진 사본, 읽을 수 없는 사본 (디렉터리가 그 자리에 있음), 상한 `priorManifestGenerations` (8) 보다 많은 세대 (10개).
- **When** (a) `Track`(배포 경로 기록) 과, 배포가 아닌 경로의 재추적(`Track` 을 부르는 사용자 파일·섹션 해시 재기록 경로) 을 부르고 저장·재적재한다. (b) 갱신 흐름이 위 세 실행을 거친다 (규칙 7 의 공허 방지 단언 포함). (c) 운영 이전 세대 원천이 증거를 묻는다.
- **Then** (a) 배포 경로로 기록된 새 항목은 배포 바이너리의 버전을 `deployed_version` 에 가진다. 배포하지 않은 항목(옛 JSON 의 항목, 재추적 경로의 항목) 은 필드가 지어지지 않고 이미 있던 값은 지워지지 않으며, 옛 JSON 을 적재해 다시 저장한 바이트가 처음 바이트와 같다. (b) 실행 2 뒤: 실행 2 가 배포한 항목은 `deployed_version == V2`, **배포하지 않은 항목(사용자 수정 파일, 옛 항목) 은 V1 값 또는 없음 그대로**이다 (V2 가 채워진 항목이 하나도 없다). 매니페스트 수준 `Version == V2` 이고 `DeployedAt` 이 바뀐다. 실행 3 (같은 V2) 뒤: 매니페스트 바이트가 실행 2 뒤와 같다. 각 실행의 실행 백업 디렉터리 `manifest.prior.json` 의 바이트는 **그 실행 시작 시점 매니페스트 파일의 바이트와 같다** (재직렬화가 아님 — 시작 파일이 비정규 모양이라 재직렬화하면 달라진다). 임시 파일 흔적이 성공 뒤 남지 않는다. (c) 원천은 최신 세대부터 최대 8개를 읽고 9번째 이후는 읽지 않으며 (읽은 파일 수를 센다), 깨졌거나 읽을 수 없는 사본은 건너뛴다 (증거 없음 → 보존, 오류로 실행을 죽이지 않는다).
- 판정 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/manifest/ -run '^TestManifestRecordsDeployedVersion$' -count=1 -v`
- 판정 명령 2: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestPriorManifestRetainedInRunBackup|TestPriorGenerationSourceBoundedAndTolerant)$' -count=1 -v`
- 기대: 명령 1 은 해당 `--- PASS`, 명령 2 는 두 시험의 `--- PASS`, 종료 코드 모두 0.
- RED 이유: 필드와 보존 단계와 원천이 없어 단언 실패 또는 빌드 실패.
- 변이: 모든 항목에 버전을 찍거나 옛 항목을 채우는 구현 (V2 가 채워진 비배포 항목으로 붉음), 사본을 재직렬화하는 구현 (바이트 불일치), 매 실행 `DeployedAt` 을 갱신하는 구현 (실행 3 바이트 불일치), 세대 상한 없이 읽는 구현 (읽은 수 > 8), 깨진 사본에서 오류를 내는 구현 (실행이 죽음) — 각각 위 단언이 붉어진다.
- 비고: 원자적 생성의 실패 주입·손상 세대·옛 JSON 왕복·다운그레이드·보존 기간은 AC-OWR-21. Q2 의 기록된 대안 (나) 에서는 이 기준이 삭제된다 (plan.md § G.1).
- 등급: regression-guard → 승격 지점 **M1b**.

### AC-OWR-07 — 프로필 지도 프로젝트: 대응 파일 없는 사본 *(maps REQ-OWR-004)* (항목 (2))

- **Given** 런치 원장이 프로젝트를 프로필 `work` 에 지도한 샌드박스 (`profileReachSandbox` + `mapProjectToWork`), `template_managed` 로 등록되고 바이트가 `TemplateHash` 와 같은 에이전트 파일 둘과 스킬 파일 하나, 사용자 루트 대응물은 없음 (임시 HOME 이 비어 있음), 제거 스위치를 픽스처에서 켬.
- **When** 마이그레이션이 실행된다.
- **Then** 세 파일은 제자리에 바이트 그대로 남고 제거 목록이 비어 있다. 세 항목의 `KeptBy` 는 `profile-reach` 이다 (Q1 의 결정 그대로). 잔존 행의 이유는 `profile-reach`, 사용자 루트 대응물 상태는 `absent`, 프로필 쪽 대응물 상태는 `not-checked` 이다 (리더 판정 2026-10-18: 프로필 쪽은 읽지 않는다). 어느 파일도 `removal-eligible` 이 아니다. 변형 1: 사용자 대응물이 확인되면 상태가 `confirmed` 이고 여전히 보존된다 (기존 `TestMigrationKeepsSharedCopyUnderProfileReach` 는 그대로 GREEN). 변형 2: 대응물이 있으나 해시가 다르면 `unconfirmed`.
- **도달 대조군 (규칙 8).** 같은 구성에서 프로젝트를 프로필에 지도하지 않은 쌍둥이 픽스처 (대응물 확인됨) 가 같은 입력으로 사본을 **제거**한다 — 보존이 프로필 지도 때문임을 보인다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestProfileMappedCopyKeptWithoutCounterpart$' -count=1 -v`
- 기대: `--- PASS: TestProfileMappedCopyKeptWithoutCounterpart (…)`, 종료 코드 0.
- RED 이유: 대응물 상태와 `not-checked` 가 이유에 기록되지 않아 단언 실패.
- 변이: 프로필 매핑을 무시하는 구현 — 대응물 확인 변형에서 사본이 제거되어 붉어진다. 프로필 쪽 상태를 `absent`/`confirmed` 로 적는 구현 (읽은 척) — `not-checked` 단언이 붉어진다.
- 등급: regression-guard → 승격 지점 **M2**.

### AC-OWR-08 — 프로필 루트 무접근: 호출 닫힘 전체의 심볼 단위 소스 스캔(호출 지점 동결) + 스캔 변이 대조군 + 정규화 차분 카나리 + 파일 불변 *(maps REQ-OWR-005)* (항목 (2))

종전 판본의 검증은 두 번 철회했다. iteration 1: 0o000 디렉터리 읽기 오류 (오류를 버리는 구현과 root 실행에서 무력). iteration 2 (감사 N1): 세 패키지·아홉 토큰 스캔은 갱신 경로의 `internal/profile` 을 보지 않았고 내보낸 도우미 `profile.GetProfileDir` 을 토큰에 갖지 않아서, 읽고 결과를 버리는 변이를 스캔·카나리·카드 이력 검사 모두 통과시켰다. 설계와 한계는 design.md D-8.

- **Given** (a) 소스 스캔: **갱신 경로의 호출 닫힘을 `go list -deps -f '{{.ImportPath}}' ./internal/cli` 로 시험이 도출한다** (손으로 쓴 패키지 목록이 없다; 이 트리에서 모듈 안 패키지 142개 — § L-11). 닫힘 안 모든 비시험 `.go` 파일에서 (L1) `internal/profile` 이 내보낸 모든 식별자의 `profile.<식별자>` 호출 지점, 문자열 토큰 `CLAUDE_CONFIG_DIR`·`claude-profiles`·`claude-skills@`·`claude-agents@`·`GetBaseDir`·`BaseDirOverride`·`EnvClaudeConfigDir`·`profilesDir`·`GetCurrentName` (상수 접기 뒤 비교), 상수 이름이 아닌 환경 변수 읽기를 (파일, 심볼) 별 개수로 센다. `internal/profile` 안에서는 함수 단위 파일시스템 동사 호출 개수를 센다. 기준선은 M0 에서 손대지 않은 트리로 측정해 `testdata/profile_callsite_baseline.txt` 에 얼린다 (측정 명령과 트리 SHA 를 파일 머리에 적는다). **이름 붙은 예외는 기준선의 한 행**이다: 패키지 `internal/profile` · 함수 `launchCandidateIsUsable` · `os.Stat` · 개수 1 (`profile.go:439` 는 이 트리의 좌표). (b) 스캔 변이 대조군 (L2): 스캔 함수는 파일 → 소스 맵을 받는 순수 함수이며 메타 시험이 실제 소스 맵에 변이 소스를 끼운다. (c) 차분 카나리 (L3): `CLAUDE_CONFIG_DIR` 을 카나리 프로필 디렉터리 (`profile.BaseDirOverride` 아래 `work/`) 로 지정하고 그 안 `skills/x/SKILL.md`, `agents/y.md` 의 상태를 세 가지 — 없음 / 템플릿과 바이트 같음 / 해시 다름 — 로 만든 같은 프로젝트 세 벌 (AC-OWR-07 과 같은 구성). (d) 파일 불변 (L4): 두 보호 파일과 `internal/profile/`.
- **When** (a) 스캔 시험이 돈다. (b) 변이 소스로 스캔이 돈다. (c) 세 벌 각각에 갱신 흐름이 `--force` 로 실행된다 (규칙 7). (d) 이 카드의 커밋 이력이 조회된다.
- **Then** (a) 기준선에 없는 (파일, 심볼) 행도, 기준선보다 늘어난 개수도 없다. 예외 행은 개수 정확히 1 로 일치한다. 스캔한 패키지 집합이 `go list -deps` 결과와 같고 `internal/profile`·`internal/config`·`internal/manifest`·`internal/template`·`internal/core/project`·`internal/userassets` 를 포함하며 0 개가 아니다 (공허 방지). 마이그레이션·미리보기·판정·갱신 동기화 파일 (`migrate_*.go`, `update_template_sync.go`, `update_dryrun_preview.go`, 새 `migrate_ownership*.go`, `update_prior_manifest.go`) 의 새 호출 지점은 0 이다 (기준선에 있는 `migrate_project_assets.go:185` 의 `ResolveLaunchProfileForProject` 한 행만 같은 개수로 남는다). (b) **변이마다 비어 있지 않은 문제 목록이 나온다 — 시험이 붉음을 직접 관측한다**: M-a 마이그레이션 파일에 `_, _ = os.ReadDir(filepath.Join(profile.GetProfileDir("work"), "skills"))` 추가 (읽고 버림, 내보낸 도우미), M-b `os.Getenv("CLAUDE_"+"CONFIG_DIR")` 추가 (토큰 접합), M-c `launchCandidateIsUsable` 에 **둘째 `os.Stat`** 추가 (같은 함수의 두 번째 호출 — 예외 개수 2), M-d `internal/profile` 의 다른 함수에 `os.ReadDir` 추가 (같은 파일의 새 함수), M-e 새 파일에서 `profile.HasClaudeConfig("work")` 호출, M-f `os.Getenv(name)` (상수 아닌 이름) 추가. 음성 대조: 변이 없는 실제 맵은 빈 목록이다. (c) 정규화한 관측 기록 (design.md D-8 L3: 임시 루트→`<ROOT>`, 백업 디렉터리 이름→`<RUN>`, 타임스탬프와 `deployed_at`→`<TS>`; 백업 목록은 `<RUN>` 아래 상대 배치) 의 제거 목록, 잔존 행 (프로필 쪽 상태는 세 벌 모두 `not-checked`), 저장된 매니페스트 바이트, 보고 줄, 표준 출력·오류, 백업 디렉터리 목록, 프로젝트 트리 해시가 세 벌 모두 같다. 정규화 규칙 자체를 먼저 단언한다 (같은 입력의 두 번 정규화 결과가 같고, 지워지는 것이 위 세 종류뿐이다). (d) 이 카드의 커밋 중 `internal/config/protected_zone.go`, `internal/config/userroots.go`, `internal/profile/` 아래 파일을 건드린 것이 없다.
- 판정 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestProfileEnvDoesNotReachProfileRoot|TestNoProfileRootAccessInUpdateSources|TestProfileScanCatchesMutants)$' -count=1 -v`
- 판정 명령 2: `git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- internal/config/protected_zone.go internal/config/userroots.go internal/profile`
- 판정 명령 3 (양성 대조군 — 명령 2 의 빈 출력을 읽기 전에 먼저 확인): `git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- .moai/specs`
- 기대: 명령 1 은 세 시험의 `--- PASS` 종료 코드 0. 명령 2 는 **빈 stdout, 종료 코드 0**. 명령 3 은 한 줄 이상 (이 트리에서 측정 가능한 값은 승격 B 커밋의 RED 셀이 적는다; § L-6 의 종전 측정은 2줄이었고 그 뒤 이 카드의 커밋이 더해져 지금은 더 많다), 종료 코드 0. 명령 3 이 비면 명령 2 의 빈 출력은 측정되지 않은 것이다.
- 정보용 (판정 아님): `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go` 는 이 트리에서 종료 코드 0 (§ L-6). main 병합이 t1702 의 정당한 변경을 들여오면 이 명령은 이 카드와 무관하게 붉어질 수 있어 판정 명령으로 쓰지 않는다.
- RED 이유: 시험 셋이 아직 트리에 없어 지금 실행하면 제로 매치 (`[no tests to run]`) 이고 이는 RED 가 아니다. A 커밋 뒤의 RED 셀에는 변이 입력 — 판정 코드에 `profile.GetBaseDir()` 를 읽어 `os.ReadDir` 하고 결과를 출력에 섞는 구현 (차분 카나리 적색), 위 M-a ~ M-f (스캔 적색) — 이 붉어지는 것을 기록한다.
- **한계 (거짓 안심 방지, 이름으로).** (1) 나열된 심볼·토큰을 하나도 쓰지 않고 조각난 문자열로 경로를 짓는 구현 (예: 실행 시점 값을 이어 붙임) 은 L1 도 L3 도 못 잡는다. (2) 기준선에 이미 있는 호출 지점의 인자만 바꾸는 구현 (개수 불변) 은 L1 이 못 잡는다. (3) 읽고 결과를 버리는 접근은 **런타임에 관측할 수단이 없다** (접근 기록 시임 없음; atime 은 근거가 못 된다) — 그 구현이 나열된 심볼·토큰을 쓰면 L1 이 잡고 안 쓰면 (1) 이다. 즉 차분 카나리는 "도우미를 통한 읽기를 관측한다" 는 약속을 하지 않고, 읽기 결과가 관측 가능한 곳에 새는 구현만 잡는다. (4) 정규화로 지워진 영역(타임스탬프, 백업 디렉터리 이름, 임시 루트) 은 비교되지 않는다. 접근 기록 시임은 채택하지 않았다 (decision-index.md 수용 부채에 (1) 을 리더가 이름으로 수락 또는 거부할 항목으로 올렸다).
- 등급: regression-guard → 승격 지점 **M2**.

### AC-OWR-09 — commands: 한 실행에 한 처분 *(maps REQ-OWR-006)* (항목 (3))

- **Given** 템플릿 commands 를 네 상태로 나눈 픽스처, 상태마다 관측 가능한 결과는 **하나**로 고정한다: (S1) 건강한 매니페스트 항목 (managed) — 실행 뒤 바이트 == 현재 템플릿 렌더. (S2) 사용자 수정 (템플릿 운반 + 항목 있음, 사용자가 줄 하나를 덧붙임) — 실행 뒤 파일에 **사용자가 덧붙인 줄이 그대로 존재한다**. (S3) 템플릿과 바이트가 같지만 항목 없음 (`migAbsentRecCmd` 모양) — 실행 뒤에도 파일이 존재하고 바이트 == 템플릿 렌더. (S4) 사용자 작성 `.claude/commands/moai-user-cmd.md` — 바이트 불변. 변형 **V1 같은 버전** (기록과 바이너리 버전이 같다), **V2 낡은 기록** (템플릿 버전 기록과 매니페스트 `Version`·`TemplateTreeHash` 가 바이너리보다 오래된 릴리스 값 — 현장은 rc.35 기록 + rc.36 바이너리였다).
- **When** V1 은 전체 갱신 흐름 (`runUpdateFlowWithProgress`) 으로, V2 는 `versionMatchSkipsSync` 를 거치는 `runUpdate` 진입으로 구동된다.
- **Then** reconcile 계획의 분류 목록 (template-owned, user-modified, user-owned, stale) 이 서로소이다 — 한 경로가 둘 이상의 처분에 나타나지 않는다. S1 ~ S4 가 위 결과를 갖는다. 실행 뒤 템플릿이 운반하는 모든 commands 파일이 존재한다. V2 는 처분의 서로소와 존재를 단언하고, V1 과 결과가 다르면 그 차이를 RED 셀에 그대로 적는다.
- **공허 방지 (규칙 7).** V1 의 실행 1 은 건너뛰지 않음을 단언한다 (`skipped == false`; 이 트리의 픽스처에서 측정, § L-13). V2 는 프로젝트 템플릿 버전·`TemplateTreeHash` 가 바이너리와 달라서 `versionMatchSkipsSync` 가 거짓인 구성이어야 하며, 시험은 출력에 "이미 최신" 표지가 없고 실행 백업에 새 `manifest.prior.json` 이 생겼음을 단언한다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCommandsSingleDispositionAndPresent$' -count=1 -v`
- 기대: `--- PASS: TestCommandsSingleDispositionAndPresent (…)` 와 서브 시험 V1, V2, 종료 코드 0.
- RED 이유: 현장의 "이전 제외 후 재배포" 가 이 트리에서 재현되는지가 이 시험의 첫 관측이다 (research.md G-2). 재현되면 서로소 단언이나 존재 단언이 붉어지고, **두 변형을 모두 돌려도** 재현되지 않을 때만 이 기준은 회귀 가드로 남아 RED 셀에 "재현 안 됨" (돌린 변형 포함) 을 적고 리더에게 보고한다. V2 를 기존 시임만으로 구동할 수 없으면 그 사실과 이유를 셀에 적는다 (갭 G-A).
- 등급: regression-guard → M6 의 측정 결과가 재현이면 승격 지점 **M6**, 아니면 회귀 가드로 유지.

### AC-OWR-10 — 중단 뒤에도 commands/agents 탐색 가능 *(maps REQ-OWR-006)* (항목 (3))

- **Given** 앞선 실행이 `.claude/commands/moai` 와 `.claude/agents/moai` 를 비운 채 죽은 상태 (픽스처가 만든다). harness 별로 claude, both 두 경우.
- **When** 갱신이 다시 실행된다.
- **Then** commands 와 agents 는 **형식이 달라 조건을 분리**한다. **commands** (`.claude/commands/moai/*`): 파일 이름 집합이 템플릿이 운반하는 17개와 같다 (`.md.tmpl` 은 `.md` 로 배포된 이름). 각 파일은 비어 있지 않고 frontmatter 에 비어 있지 않은 `description` 과 `allowed-tools` 가 있다 — 템플릿 감사 시험 (`commands_audit_test.go`) 이 템플릿에 요구하는 것과 같은 조건이다. **`name` 은 요구하지 않는다**: 템플릿 command 17개 모두 `name:` 이 없다 (§ L-3a). 바이트는 배포기가 같은 언어 문맥에서 렌더한 결과와 같다 (시험이 기대 바이트를 배포기 자신의 렌더로 얻는다 — 템플릿 원문과 직접 비교하지 않는다, `.md.tmpl` 이 렌더되기 때문이다). **agents** (`.claude/agents/moai/*.md`): 이름 집합이 템플릿의 12개와 같고 바이트가 템플릿과 같으며 frontmatter 의 `name` 과 `description` 이 비어 있지 않다 (12개 모두 `name:` 이 있다 — § L-3b). both 에서는 `.codex/agents/moai/*.toml` 도 템플릿 집합과 같다. 실행 백업에서 사라진 사용자 파일은 없다.
- **공허 방지 (규칙 7).** 복구 실행은 `--force` (`runUpdateFlowForced`) 로 구동하고 `skipped == false`, 새 `manifest.prior.json`, 기대한 양수의 잔존 행을 단언한다. 같은 상태에 대한 **건너뛰는 실행 사례**: 일반 같은 버전 재실행은 `skipped == true` 이고 비어 있는 상태가 그대로 남는다 — 이 SPEC 이 건너뛰기 관문을 바꾸지 않는다는 운영 사실의 기록이며, 복구 보증은 조정에 닿는 실행에 대해서만 성립한다 (REQ-OWR-006·-013, plan.md OQ-6).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCommandsAndAgentsDiscoverableAfterInterruptedRun$' -count=1 -v`
- 기대: `--- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun (…)` (서브 시험 2개), 종료 코드 0.
- RED 이유: 전용 시험이 없다 (research.md D(3)). 회귀 가드로 시작해 구현 변화가 이 집합을 깨는 변이로 붉어지는 것을 RED 셀에 기록한다.
- 변이: 복구 실행이 비어 있는 commands 디렉터리를 다시 배포하지 않는 구현 — 파일 이름 집합 단언이 붉다. 한 파일만 빈 내용으로 쓰는 구현 — 비어 있지 않음 단언이 붉다.
- 한계: Claude Code 가 명령을 무엇으로 발견하는지 (파일 경로 / `name`) 는 이 트리의 측정 사실이 아니다. 이 기준은 템플릿이 실제로 운반하는 형식만 단언하고 발견 방식을 단언하지 않는다.
- 등급: regression-guard → 승격 지점 **M6** (변이로 적색을 관측한 뒤).

### AC-OWR-11 — 대소문자 변형을 한 파일로 판정 *(maps REQ-OWR-007)* (항목 (4))

- **Given** 등록 키 `.claude/skills/moai-ref-owasp-checklist/SKILL.md` (`template_managed`, 바이트 == `TemplateHash`) 와 디스크의 같은 파일이 `skill.md` 철자로 존재하는 상태. 파일시스템 대소문자 모드를 주입하는 시험과, 실제 호스트 파일시스템에서 도는 시험 둘.
- **When** 판정 함수가 걷기 경로 `.../skill.md` 를 판정한다.
- **Then** 대소문자 무구분 모드 (또는 실제 호스트가 무구분일 때): 등록 항목의 판정이 적용되어 분류는 `managed`, 이유에 `case-variant` 가 붙고, 이 물리 파일의 행은 하나다. 대소문자 구분 모드: 두 철자는 별개 경로이다 — `skill.md` 는 `unregistered`, `SKILL.md` 는 자기 항목의 판정을 받는다. 어느 모드에서도 파일 이름은 바뀌지 않는다. 운영 코드는 파일시스템 모드를 따로 판정하지 않는다 (관측된 `os.SameFile` 이 근거, design.md D-4). 실제 호스트 시험은 **시험 자신의 임시 디렉터리에서** 대소문자 프로브를 해 그 결과에 맞는 기대를 단언하며 어느 쪽도 건너뛰지 않는다 (`probeCaseSensitiveFS` 는 비시험 호출자가 없고 프로젝트 루트에 쓰므로 운영 경로에서 쓰지 않는다 — § L-9).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestCaseVariantResolvesToRegisteredEntry|TestCaseVariantOnHostFilesystem)$' -count=1 -v`
- 기대: 두 시험의 `--- PASS` (첫 시험은 서브 시험 2개: 무구분/구분), 종료 코드 0.
- RED 이유: 현재 `skill.md` 는 매니페스트 맵 조회를 놓쳐 `unregistered` 로 센다 (research.md D(4); 이 호스트는 무구분이다).
- 변이: 철자를 무시하고 항상 별칭으로 합치는 구현 — 구분 모드에서 `skill.md` 가 `managed` 가 되어 붉어진다.
- 등급: regression-guard → 승격 지점 **M4**.

### AC-OWR-12 — 대소문자 변형: 사용자 매니페스트 키와 `.agents/skills` 철자 *(maps REQ-OWR-007)* (항목 (4))

- **Given** 사용자 매니페스트가 `claude-skills/<name>/SKILL.md` 를 기록하고 프로젝트 파일은 `skill.md` 철자인 대소문자 무구분 모드, 그리고 `.agents/skills/<name>/skill.md` 의 미등록 복사본.
- **When** 마이그레이션이 대응물을 확인하고 미등록 복사본을 판정한다.
- **Then** 대응물 상태가 `confirmed` 이다 (철자 때문에 `unconfirmed` 가 되지 않는다). 미등록 복사본은 하나의 행으로 `no-provenance-record` + `case-variant` 이유를 받고, 이름 바꾸기 없이 바이트 그대로 남는다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCaseVariantCounterpartAndMirrorSpelling$' -count=1 -v`
- 기대: `--- PASS: TestCaseVariantCounterpartAndMirrorSpelling (…)`, 종료 코드 0.
- RED 이유: 사용자 매니페스트 키 조회와 내장 트리 조회가 철자가 다르면 빗나간다 (`migrate_project_assets.go:400`, `:436`).
- 등급: regression-guard → 승격 지점 **M4**.

### AC-OWR-13 — 관리 파일만 있던 빈 디렉터리 정리 *(maps REQ-OWR-008)* (항목 (5))

- **Given** 이전 매니페스트 세대가 기록한 파일이 전부 `template_managed` 이고 지금은 전부 없는 빈 디렉터리 셋 — (A) 이전 실행이 비운 스킬 디렉터리, (B) 이번 실행이 마지막 파일을 제거해 비운 스킬 디렉터리 (제거 스위치는 픽스처에서 켬), (C) 하위 폴더만 남아 있고 그 하위 폴더도 비어 있는 중첩 디렉터리. 그리고 반례: (D) 이전 기록이 없는 빈 디렉터리, (E) 이전 기록에 `user_created` 파일이 있었던 빈 디렉터리, (F) 사용자 파일이 들어 있는 디렉터리, 스킬 루트, 비어 있는 `.claude/agents/moai`.
- **When** 마이그레이션이 실행된다.
- **Then** A, B, C 는 사라진다 (C 는 아래쪽부터). D 와 E 는 남고 잔존 행의 이유가 `empty-dir-no-provenance` 이다. F 는 내용이 그대로이다. 스킬 루트와 에이전트 디렉터리는 남는다. 정리 단계는 파일을 지우지 않는다 (`os.Remove` 는 빈 디렉터리만 지운다).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestManagedOnlyEmptyDirsPruned$' -count=1 -v`
- 기대: `--- PASS: TestManagedOnlyEmptyDirsPruned (…)` (서브 시험: A, B, C, D, E, F), 종료 코드 0.
- RED 이유: 걷기가 이미 비어 있던 디렉터리를 보지 않고 (`migrate_project_assets.go:195-197`) 정리 판단의 근거도 없다.
- 변이: 비어 있기만 하면 지우는 구현 — D 가 사라져 붉어진다.
- 비고 (Q2 (나), plan.md § G.1): 사례 A 와 C 는 이전 세대 기록에 의존하므로 삭제되고, B 는 t1668 의 `pruneEmptiedSkillDir` 가 이미 하는 일이라 새 코드 없이 회귀 가드가 된다. D·E·F 는 음성 대조군으로 남는다.
- 등급: regression-guard → 승격 지점 **M5**.

### AC-OWR-14 — 잔존 목록과 313개 보존 *(maps REQ-OWR-003, REQ-OWR-009)* (항목 (6), (7))

- **Given** 복사 폴백 거울을 흉내 낸 `.agents/skills` 아래 미등록 파일 313개. **각 파일의 바이트는 내장 템플릿에 실재하는 원본에서 읽는다**: `.claude/skills` 의 템플릿 파일 315개 (§ L-5) 에서 경로 정렬 순으로 313개를 골라 `.agents/skills/<같은 상대 경로>` 에 둔다. 전제 단계 (시험의 첫 단언): 313개 각각의 원본 경로가 `fs.ReadFile(embedded, …)` 로 조회되고 바이트가 같다, 개수가 313 이다. 확인된 사용자 대응물, 제거 스위치를 픽스처에서 켬. 별도로 사용자 수정 파일 몇 개와 프로필 지도 사본 몇 개가 섞인 두 번째 픽스처.
- **도달 대조군 (규칙 8).** 313개 옆에 모든 제거 전제를 갖춘 등록된 managed 사본이 하나 이상 있고 첫 실행에서 제거된다.
- **When** 갱신 흐름이 실행되고 (실행 1, 일반), 이어서 **`--force` 로** `--verbose` 와 함께 한 번 더 실행된다 (실행 2 — 일반 같은 버전 재실행은 건너뛴다, 규칙 7). 같은 픽스처에 AC-OWR-02 의 바이트 일치 승격 변이 원천을 꽂은 대조 실행이 한 번 더 있다.
- **Then** 313개 모두 바이트 그대로 남는다. 매니페스트는 바이트 단위로 바뀌지 않는다 (배포가 기록하는 항목 외에 새 항목이 없다). 잔존 행 313개가 이유 `no-provenance-record` 와 빠진 증거 항목을 가지며 경로 사전순으로 정렬된다. 기본 렌더는 이유별 개수 줄을 내고 개별 경로를 찍지 않으며, `--verbose` 렌더는 파일별 행을 낸다. 두 번째 픽스처의 행은 이유가 섞여 있고 각 행에 이유가 있다 (이유 없는 행이 없다). 변이 원천 앞에서는 보존 단언이 비어 있지 않은 문제 목록을 돌려준다 (시험이 적색을 직접 관측). 실행 2 는 규칙 7 의 세 조건을 만족한다 (`skipped == false`, 새 `manifest.prior.json`, 잔존 행 313 이상) — 그래서 "`--verbose` 의 파일별 행" 은 실제로 렌더된 출력에서 읽는다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$' -count=1 -v`
- 기대: `--- PASS: TestRemainingListKeeps313UnregisteredCopies (…)`, 종료 코드 0.
- RED 이유: 구조화된 잔존 행과 이유별 렌더가 없다 (현재는 보고 문자열과 사유 없는 경로 집합뿐 — `update_migrate.go:127`).
- 변이: 이유 열을 비운 행을 내는 구현 — "이유 없는 행 없음" 단언이 붉어진다. 313개를 기본 출력에 모두 찍는 구현 — 기본 렌더 단언이 붉어진다. 바이트 일치로 승격하는 구현 — 위 대조 실행이 붉어진다.
- 등급: regression-guard → 승격 지점 **M3**.

### AC-OWR-15 — 백업·복원 바이트 동일: 복원 영역과 계산된 나머지 차이 *(maps REQ-OWR-010)* (항목 (7))

종전 판본은 복원한 트리의 해시가 실행 전 트리 해시와 같다고 요구했다. 전체 갱신 흐름은 제거 말고도 `.moai/config/sections/llm.yaml` 을 다시 쓰고 템플릿 파일을 배포하고 디렉터리 mtime 을 바꾸므로 (`update_migrate_test.go` 의 `TestMigrationIdempotent` 가 같은 흐름에서 `config.ReadDeployMode == "local"` 을 단언; codex 가 같은 복원을 측정해 `differing-paths=529`), 제거 파일과 매니페스트만 되돌리는 절차로는 어떤 올바른 구현도 통과할 수 없었다 (감사 N4). 요구사항 REQ-OWR-010 이 말하는 것은 제거된 파일과 매니페스트의 바이트 동일이다.

- **Given** 제거 스위치를 픽스처에서 켠 전체 갱신 흐름 (실행 1, 일반) 과 실행 전 상태: 제거 후보 각각의 sha256, 실행 전 매니페스트 파일 바이트. 비정규 모양의 시작 매니페스트 (AC-OWR-06 (b)). **통제 실행**: 같은 픽스처의 독립 복사본에 제거 스위치를 **끄고** 같은 흐름을 돌린 트리 T0 (제거 말고 흐름이 하는 모든 재작성·배포를 담는다).
- **When** 흐름이 파일을 제거한 뒤 (트리 T1), 백업된 사본을 원위치에 되돌리고 `manifest.prior.json` 을 매니페스트 자리에 되돌린다 (되돌리기는 시험 전용 절차이며 운영 하위 명령이 아니다) → 트리 T2.
- **Then** (i) 제거된 각 파일의 백업 사본 sha256 이 제거 전 sha256 과 같다. (ii) `manifest.prior.json` 의 바이트가 실행 전 매니페스트 파일의 바이트와 같다. (iii) **복원 영역 R = {제거된 파일 전부, 매니페스트 파일}**: T2 의 R 의 각 파일 바이트가 실행 전 바이트와 같다. (iv) **나머지 차이의 센티널**: T2 와 통제 실행 트리 T0 를 비교한 차이 경로 집합 D 를 시험이 계산해 기대 허용 집합 E (`.moai/manifest.json`, `.moai-backups/**`, `.moai/logs/**`, `.moai/state/**`) 와 비교한다 — 정규화 (design.md D-8 L3 의 세 종류) 뒤 `D ⊆ E` 이고, `E` 밖의 차이가 하나라도 있으면 붉다. 이 센티널이 "흐름의 다른 재작성은 정상 차이이고 그 외는 없다" 를 단언한다. E 의 항목 중 실제로 차이가 나는 것과 나지 않는 것을 RED 셀이 측정해 적는다 (E 가 부풀려져 아무것도 못 잡는 일을 막기 위해 D 가 비어 있지 않음과 D ⊆ E 를 함께 단언하고, E 의 각 항목이 D 에 나타나는지를 보고한다). (v) 실행 1 이 조정에 닿았음 (`skipped == false`, 새 `manifest.prior.json`, 제거 목록 비어 있지 않음).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemovedAssetBackupRestoreByteIdentical$' -count=1 -v`
- 기대: `--- PASS: TestRemovedAssetBackupRestoreByteIdentical (…)`, 종료 코드 0.
- RED 이유: 매니페스트 사본이 실행 백업에 없어 (ii) (iii) 가 붉다 (파일 사본의 바이트 일치는 `TestMigrationBackupBeforeRemoval` 이 이미 지킨다, § L-4).
- 변이: 백업 사본에 줄 끝을 바꾸거나 잘라 쓰는 구현 (i 붉음), 매니페스트 사본을 재직렬화하는 구현 (ii 붉음 — 시작 매니페스트가 비정규 모양이라서), 제거와 함께 다른 파일(예: 사용자 파일)도 지우는 구현 (iv 에서 E 밖 차이로 붉음 — 사용자 파일은 T0 에 있고 T2 에 없다).
- 비고: 이 기준은 **성공 경로**이다. 사본 쓰기가 실패하는 경로는 AC-OWR-20. 파일 모드는 보존 범위 밖이다 (spec.md § J).
- 등급: regression-guard → 승격 지점 **M8**.

### AC-OWR-16 — `moai update --dry-run` 전체 진입의 무변경과 판정 일치 *(maps REQ-OWR-001, REQ-OWR-011)* (항목 (7))

- **Given** AC-OWR-02 ~ -05, -07 의 파일들이 섞인 프로젝트, 제거 스위치를 픽스처에서 켬 (그리고 별도로 끔). 실행 전에 모든 디렉터리의 mtime 을 고정된 과거 시각으로 맞춘다 (`os.Chtimes`).
- **When** `moai update --dry-run` 이 디스패치하는 분기 전체 — `runUpdate` 의 `dry-run` 분기 (`update.go:361`) — 를 기존 시험 헬퍼 `runDryRunUpdateInProject` 로 구동한다. `previewReconciliation` 한 단계가 아니라 명령이다 (REQ-OWR-011 은 명령에 걸린다).
- **Then** 트리 해시 (경로, 바이트, 모드, **디렉터리 mtime**) 가 실행 전후 같다: 디렉터리 안에서 항목이 생겼다 사라지면 그 디렉터리 mtime 이 고정 과거 시각에서 바뀌므로 "만들었다 지우는" 임시 쓰기도 잡힌다. `.moai-backups` 가 생기지 않는다. 출력이 "migration candidate" 로 알린 경로 집합이 판정 함수의 `removal-eligible` 집합과 정확히 같다. 프로필 지도 + `KeptBy` 항목은 후보에 없다. 스위치가 꺼져 있으면 후보가 없다 (기존 `TestPreviewOmitsRemovalCandidateWhileSwitchOff` 는 GREEN 유지). 별도 소스 스캔: 이 SPEC 이 더하는 판정·미리보기 코드 (`migrate_ownership*.go`, `update_dryrun_preview.go` 의 변경분) 에 `os.WriteFile`, `os.Create`, `os.Mkdir*`, `os.Remove*`, `os.Rename`, `probeCaseSensitiveFS` 호출이 없다.
- **공허 방지 (design.md D-9).** `previewReconciliation` 은 계획이 비면 아무것도 출력하지 않고 (`plan.Count() == 0`), 후보 알림은 `removalEnabled` 가 참이고 reconcile 이 `UserOwned` 로 분류한 공통 자산 파일에만 나온다. 그래서 시험은 집합을 비교하기 **전에** 단언한다: 출력에 "Dry-run: reconciliation preview" 머리가 있다; 픽스처의 `removal-eligible` 집합이 비어 있지 않다; 그 집합의 모든 경로가 reconcile 계획의 `UserOwned` 에 있다 (없는 경로가 있으면 그 경로 목록이 RED 셀에 적히는 측정 결과이고, 같음의 비교는 빈 집합끼리로 성립하지 않는다). 모든 디렉터리 mtime 고정은 `runDryRunUpdateInProject` 호출 직전에 한다.
- 판정 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestDryRunChangesNothingAndMatchesVerdict$' -count=1 -v`
- 판정 명령 2 (기존 분기의 같은 성질을 지키는 가드, GREEN 유지): `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestUpdateDryRun_ZeroMutation$' -count=1 -v`
- 기대: 명령 1 은 `--- PASS: TestDryRunChangesNothingAndMatchesVerdict (…)`, 명령 2 는 `--- PASS: TestUpdateDryRun_ZeroMutation (…)` 과 서브 시험 두 개, 종료 코드 모두 0.
- RED 이유: 미리보기 판정 (`update_dryrun_preview.go:120`) 은 `KeptBy` 와 대응물 확인을 보지 않아 프로필 지도 항목을 후보로 알린다 (코드 읽기 도출, research.md G-3 — 이 시험이 첫 관측이다).
- 변이: 미리보기가 자기 판정을 유지하는 구현 — 후보 집합 불일치로 붉어진다. 판정 경로가 `probeCaseSensitiveFS(projectRoot)` 를 부르는 구현 — 프로젝트 루트 mtime 이 바뀌어 붉어진다 (바이트 해시만 보면 못 잡는 변이).
- 한계: 프로젝트 밖(HOME 등) 으로의 쓰기와 모드만 바꿨다 되돌리는 쓰기는 이 방식이 보지 못한다. 기존 dry-run 단계 (worktree 이전 미리보기, 레거시 스킬 보관 미리보기, 재설치 계획) 는 이 SPEC 의 변경분이 아니며 가드 명령 2 가 지킨다 (design.md D-9).
- 등급: regression-guard → 승격 지점 **M7**.

### AC-OWR-17 — 반복 갱신 멱등 *(maps REQ-OWR-012, REQ-OWR-015)* (항목 (7))

- **Given** AC-OWR-14 의 첫 픽스처와 섞인 픽스처. 실행 1 은 일반 (프로젝트 버전이 바이너리와 달라 건너뛰지 않는다 — 이 트리의 픽스처에서 측정, § L-13), 실행 2·3 은 **`--force`** (`runUpdateFlowForced`, 규칙 7 — 같은 버전의 일반 재실행은 건너뛰어 아무것도 검증하지 못한다). 별도 사례: 실행 3 뒤의 **일반** 재실행 (4).
- **When** 실행 1, 2, 3 (그리고 사례 4) 이 끝난다.
- **Then** 실행 1-3 각각이 규칙 7 의 세 조건을 만족한다 (`skipped == false`, 새 `manifest.prior.json`, 양수의 잔존 행) 이고 만족한 실행이 정확히 3 이다. 실행 뒤마다 계산한 트리 해시 (제외: `.moai-backups/**`, `.moai/logs/**`, `.moai/state/**`) 가 실행 2 와 3 에서 같고 (실행 1 은 첫 배포가 있어 기준에서 제외 — 실행 1 과 2 의 차이는 RED 셀이 기록한다), 잔존 행 목록이 셋 모두 같고 경로 사전순 정렬 속성을 만족한다. 매니페스트 바이트가 실행 2 와 3 에서 같다 (`DeployedAt` 은 `Version` 이 바뀔 때만 바뀐다 — REQ-OWR-015, design.md D-3). 사례 4: 일반 재실행은 `skipped == true` 이고 트리 해시·매니페스트 바이트·출력이 바뀌지 않으며 잔존 목록을 새로 내지 않는다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRepeatedUpdateIsIdempotent$' -count=1 -v`
- 기대: `--- PASS: TestRepeatedUpdateIsIdempotent (…)` 와 서브 시험 `ReachedReconciliation`, `ForcedRepeats`, `PlainRepeatSkips`, 종료 코드 0.
- RED 이유: 잔존 행이 없어 단언 불가, 그리고 배포 단계가 매 실행 `DeployedAt` 을 갱신하는 구현(M1b)이 들어오면 붉어진다 — 이 시험이 그 변이를 잡는다 (`--force` 실행이라 변이가 실제로 발화한다).
- 변이: 매 실행 `DeployedAt` 을 갱신하는 구현 — 실행 3 바이트 불일치로 붉다. 잔존 행을 맵 순회 순서로 내는 구현 — 정렬 속성 단언이 붉다. 일반 재실행에서 조정을 다시 하는 구현 (건너뛰기 관문을 우회) — 사례 4 의 `skipped == true` 가 붉다.
- 등급: regression-guard → 승격 지점 **M9**.

### AC-OWR-18 — 중도 중단 → 복구 → 재실행 수렴 *(maps REQ-OWR-013, REQ-OWR-016)* (항목 (7))

- **Given** 제거 스위치를 픽스처에서 켠 프로젝트와 같은 프로젝트의 중단 없는 기준 실행 결과. 중단 모사 세 가지 (운영 코드에 끼어들기를 넣지 않는다): (a) 한 후보의 실행 백업 경로를 쓸 수 없게 만들어 그 실행의 제거가 막힌 상태를 만든다 (경로 자리에 **디렉터리를 두는** 방식으로 root 에서도 쓰기가 실패한다; REQ-OWR-016 에 따라 그 실행은 0건 제거), (b) 마이그레이션 뒤 매니페스트를 실행 전 바이트로 되돌려 저장 전 중단을 모사, (c) 정리된 빈 디렉터리를 다시 만들어 정리 전 중단을 모사.
- **When** 중단 상태에서 복구 (쓸 수 없던 경로를 풀거나 백업 사본을 되돌림) 한 뒤 갱신을 **`--force` 로** 다시 실행한다 (일반 같은 버전 재실행은 건너뛰어 (c) 의 디렉터리를 남긴다 — § L-13; 규칙 7). 사례마다 재실행이 규칙 7 의 세 조건을 만족함을 단언한다.
- **Then** 최종 트리 해시와 잔존 행이 중단 없는 기준 실행과 같다. 제거된 각 파일의 백업 사본이 존재한다. 제거된 경로의 매니페스트 항목이 되살아나지 않고, 비어 있는 관리 디렉터리가 남지 않는다. **(a) 에서는 중단된 첫 실행이 파일을 하나도 제거하지 않았음**(백업 실패 시 제거 0건, REQ-OWR-016) 을 먼저 단언하고, 복구 재실행에서 후보가 모두 제거되어 기준 실행과 같아진다. (c) 의 일반 재실행(`--force` 없이) 은 건너뛰며 디렉터리를 그대로 둔다는 운영 사실을 별도 사례로 단언한다 (plan.md OQ-6).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestInterruptedMigrationRecoversAndConverges$' -count=1 -v`
- 기대: `--- PASS: TestInterruptedMigrationRecoversAndConverges (…)` (서브 시험 a, b, c), 종료 코드 0.
- RED 이유: (c) 에서 이미 비어 있던 디렉터리가 정리되지 않는다 (AC-OWR-13 의 같은 원인). (a), (b) 는 현재 구조가 이미 수렴하는지의 첫 관측이다.
- 비고 (Q2 (나), plan.md § G.1): 하위 사례 (c) 는 이전 세대 기록에 의존하므로 삭제되고 (a)(b) 는 남는다.
- 등급: regression-guard → 승격 지점 **M9**.

### AC-OWR-19 — 제거 스위치는 꺼진 채 *(maps REQ-OWR-014)* (항목 (1), (7) 게이트)

- **Given** 이 트리의 비시험 Go 소스.
- **When** 소스 스캔 시험이 `removalEnabled` 대입을 찾고, 기존 게이트 시험이 기본 구성에서 돈다.
- **Then** 비시험 코드에서 `removalEnabled` 를 `true` 로 대입하는 곳이 없다 (선언의 `false` 만). `TestMigrationRemovalGate` 와 `TestMigrationDefaultKeepsConfirmedProjectCopy` 가 GREEN 이다 (§ L-4 에서 두 시험이 이 트리에서 GREEN 임을 측정했다).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestRemovalSwitchStaysOffInProduction|TestMigrationRemovalGate|TestMigrationDefaultKeepsConfirmedProjectCopy)$' -count=1 -v`
- 기대: 세 시험의 `--- PASS`, 종료 코드 0.
- RED 이유: 소스 스캔 시험이 아직 없다. 실패쪽 관측은 스위치 기본값을 `true` 로 뒤집은 변이에서 두 기존 시험과 새 시험이 붉어지는 것으로 기록한다.
- 변이: `init()` 안에서 `removalEnabled = true` 를 대입하는 구현, `removalEnabled = cfg.Enabled` 처럼 비상수 대입 — 소스 스캔은 `removalEnabled` 에 대한 **모든 대입**(선언의 `false` 외) 을 비시험 코드에서 찾으므로 둘 다 붉다.
- 등급: regression-guard (실패쪽은 변이로 관측) → 새 소스 스캔 시험의 승격 지점 **M10**.

### AC-OWR-20 — 백업 실패 시 그 실행은 파일을 제거하지 않는다 (운영 호출 경로 전체) *(maps REQ-OWR-016)* (항목 (7))

종전 판본(iteration 2 감사 N3)은 보존 함수와 시험이 지정한 실행 문맥만 검증해서, 운영 Backup 단계가 보존 실패를 버리고 문맥을 성공으로 남기는 구현이 두 단언을 만족하면서 파일을 계속 제거할 수 있었다. 이번 판은 **같은 운영 호출 경로**(`runTemplateSyncWithProgress` 의 Backup 단계 → `migrateProjectCommonAssets`, `update_template_sync.go:815-817`) 를 전체 흐름으로 구동한다.

- **Given** 제거 스위치를 픽스처에서 켜고 모든 제거 전제를 갖춘 managed 복사본 세 개 (대응물 확인됨) 가 있는 프로젝트. 패키지 수준 변수 `priorManifestWriter` (design.md D-3 항목 5) 에 **운영 쓰기 함수를 감싼 대역**을 꽂는다 (`t.Cleanup` 으로 복원): 대역은 인자로 받은 `runDir` 안 `manifest.prior.json` 자리에 **내용이 있는 디렉터리**를 만든 뒤 운영 함수를 부른다 (이름 바꾸기가 uid 와 무관하게 실패 — root 에서도 성립). 사례 (a) 이 구성. 사례 (b) `manifest.prior.json` 은 정상이고 한 후보의 `removed-assets/<rel>` 자리만 디렉터리가 되는 구성 (바이트 사본 쓰기 실패; 대역 `removalBackupWriter` 가 없으면 파일 시스템에서 직접 만든다 — 실행 백업 이름이 시각에서 나오므로 `ensureRunBackupDir` 직후 훅이 아니라 `runDir` 인자를 받는 대역으로 만든다). 사례 (c) 대조 쌍둥이: 대역 없이 같은 흐름 — 세 사본이 모두 제거된다.
- **When** 사례마다 전체 갱신 흐름이 `--force` 로 돌고 (`runUpdateFlowForced`) 규칙 7 의 조건을 만족한다 (`skipped == false`; 흐름이 Backup 단계와 마이그레이션에 도달했음).
- **Then** (a) **제거 목록이 비어 있고** 세 사본이 모두 바이트 그대로 남으며, 잔존 행의 이유가 `backup-failed` 이고, 경고 한 줄이 이를 이름으로 댄다. 매니페스트에서 어떤 항목도 떨어지지 않는다. 갱신의 나머지 (배포) 는 계속된다 — 템플릿이 운반하는 commands 파일이 존재하는 것으로 확인. 요약 보고가 "삭제 없음" 과 모순되지 않는다 (REQ-UPM-031). (b) **제거 목록이 비어 있다**(셋 다 보존, 리더 조건 (2): 백업이 실패하면 그 실행은 0개 제거) 이고 이유는 `backup-failed`. (c) 세 사본이 모두 제거된다 — (a)(b) 의 보존이 대조 가능한 결과임을 보인다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestBackupFailureRemovesNothingThroughProductionPath$' -count=1 -v`
- 기대: `--- PASS: TestBackupFailureRemovesNothingThroughProductionPath (…)` 와 서브 시험 a, b, c, 종료 코드 0.
- RED 이유: 보존 단계와 인계와 일괄 제거 구조가 없어 빌드 실패 또는 단언 실패. (b) 는 현재 t1668 코드가 실패한 파일만 보존하고 나머지는 제거하므로 단언 실패 — 올바른 이유로 붉다.
- 변이: **사본 실패를 삼키고 문맥을 성공으로 남기는 Backup 단계** (a 에서 제거 목록이 비지 않아 붉음), 파일 바이트 복사 실패를 파일별로만 처리하는 현행 구조 (b 에서 붉음), 사본 실패를 경고로만 처리하는 구현 (a 붉음), 실패 때 매니페스트에서 항목을 떨어뜨리는 구현 (a 의 매니페스트 단언 붉음).
- 한계: `priorManifestWriter` 를 통한 주입은 운영 쓰기 함수를 감싸므로 쓰기 함수 자체의 실패 경로는 AC-OWR-21 (a) 가 따로 주입한다. 파일 모드는 범위 밖이다.
- 비고: Q2 의 기록된 대안 (나) 에서는 이 기준이 삭제되지만 (사본 자체가 없다) 파일 바이트 복사 실패에 대한 (b) 의 일괄 보존은 남는다 (plan.md § G.1).
- 등급: regression-guard → 승격 지점 **M1b**.

### AC-OWR-21 — 이전 세대의 원자적 생성과 손상 허용, 호환, 다운그레이드, 보존 기간 *(maps REQ-OWR-015, REQ-OWR-017)* (항목 (6), (7))

리더 결정 d-20261018T040918Z-def2 의 조건 (3)(4)(5)(6) 을 시험으로 닫는다. 모든 시험은 임시 HOME 과 `t.TempDir()` 만 쓴다.

- **Given** (a) 이전 세대 **운영 쓰기 함수** (design.md D-3 항목 3 — 항목 5 의 `priorManifestWriter` 의 기본값) 와 그 안의 두 관측점: 패키지 수준 변수 `priorManifestTempWrite(dir, data) (tempPath, err)` (같은 디렉터리에 임시 이름으로 바이트를 쓰고 fsync 한 뒤 임시 경로를 돌려준다) 와 `priorManifestRename(oldPath, newPath) error` (임시 경로를 최종 이름으로 옮긴다). 시험은 두 변수에 **운영 함수를 감싼 기록용 대역**을 꽂고 `t.Cleanup` 으로 복원한다 (`removalEnabled` 와 같은 모양). 대역은 호출마다 (순서 번호, 인자, 호출 **시점의** `manifest.prior.json` 존재 여부) 를 기록한다. 요구되는 관측 가능 동작: 운영 쓰기 함수는 임시 쓰기와 이름 바꾸기를 이 두 변수로만 수행하고 `manifest.prior.json` 이라는 이름을 쓰기용으로 열지 않는다. 네 가지 구성은 모두 **실제 운영 쓰기 함수**를 거친다: (o) 주입 없는 정상 실행 (기록용 대역만), (p) 임시 쓰기 도중 실패 (임시 쓰기 대역이 임시 이름 파일에 바이트의 앞 절반만 쓰고 오류를 돌려준다), (q) **임시 쓰기와 이름 바꾸기 사이의 실패** (이름 바꾸기 대역이 이름을 바꾸지 않고 주입 오류를 돌려준다 — 이때 임시 파일에는 전체 바이트가 이미 있다), (r) 실제 이름 바꾸기 실패 (`manifest.prior.json` 자리에 내용 있는 디렉터리 — 디렉터리 충돌로 root 에서도 성립). (b) 읽는 쪽 입력: 정상 세대, 잘린 JSON, 빈 파일, JSON 이지만 매니페스트가 아닌 것, 읽을 수 없는 세대 (자리에 디렉터리), 임시 이름 파일 (`manifest.prior.json.tmp-*`) 만 있는 실행 백업. (c) 옛 JSON: `deployed_version` 이 없는 매니페스트 (항목 여럿). (d) 다운그레이드: `deployed_version` 을 모르는 구조체 (필드 없는 사본 타입을 시험 안에 두어 이전 바이너리를 모사) 로 적재·재저장한 JSON. (e) 보존 기간: 실행 백업 디렉터리 12개 (정책이 남기는 5개 + 오래된 7개) 가 있는 프로젝트와 이 카드의 변경 이력. (f) 같은 버전 반복: AC-OWR-06 (b) 의 실행 2 → 3.
- **When** (a) 쓰기 함수를 세 실패로 부른다. (b) 이전 세대 원천이 각 입력을 읽는다. (c) 옛 JSON 을 적재·저장한다. (d) 다운그레이드 JSON 을 해석기에 먹인다. (e) `PruneBackups` 호출 전후와 코드 이력을 본다. (f) 두 번 실행한다 (`--force`, 규칙 7).
- **Then** (a) **(o) 정상 실행**: 기록에 임시 쓰기 정확히 1건 뒤 이름 바꾸기 정확히 1건이 이 순서로 있고, 임시 쓰기의 대상 이름이 `manifest.prior.json` 과 다르되 같은 디렉터리이며, 이름 바꾸기의 원본이 그 임시 경로이고 대상이 `manifest.prior.json` 이며, **이름 바꾸기 호출 시점에 `manifest.prior.json` 이 존재하지 않았고** (최종 이름은 이름 바꾸기로만 생긴다), 끝난 뒤 최종 파일 바이트가 입력과 같고 임시 경로는 남지 않는다. **(p)(q)(r) 세 실패**: 모두 오류를 돌려주고 (인계 → AC-OWR-20), **쓰기 함수 호출 뒤 `manifest.prior.json` 자리에 정규 파일이 없으며** ((p)(q) 는 경로 자체가 없다; (r) 은 시험이 만든 디렉터리가 내용 그대로 남는다) 읽는 쪽이 그 실행을 유효한 세대로 세지 않는다. (q) 에서는 임시 파일에 전체 바이트가 남아 있으나 이름이 달라 읽는 쪽이 무시한다 — 완전해 보이는 임시 파일이 증거가 아님을 (q) 가 직접 보인다. (b) 정상만 증거이고 나머지는 모두 **증거 없음** — 해당 경로는 `unregistered` + 증거 부족 이유로 보존되며 오류로 실행이 죽지 않는다. (c) 옛 JSON 은 적재되고 항목이 그대로이며 저장한 바이트가 입력과 같고(필드를 지어내지 않음), 해석기는 그 항목을 `evidence-incomplete:deployed-version` 으로 읽는다. (d) 다운그레이드로 필드가 사라진 항목은 증거가 없는 것으로 읽혀 **보존**된다 (제거 방향 오류 없음). (e) 이 카드의 커밋은 `internal/cli/update/backup/` 아래를 건드리지 않고 (명령 + 양성 대조군), `PruneBackups` 를 거친 뒤 실행 백업 디렉터리 집합이 정책의 기존 결과와 같다 (이 카드가 더한 코드가 디렉터리를 지우지 않았음: 시험이 `manifest.prior.json` 을 가진 실행 백업과 가지지 않은 실행 백업이 같은 규칙으로 선택됨을 단언). `priorManifestGenerations` 는 조회 한도이며 어떤 삭제 호출도 이 상수를 읽지 않는다 (소스 스캔: 비시험 코드에서 이 상수의 사용처가 원천 읽기 함수 하나뿐). (f) 실행 3 뒤 매니페스트 바이트가 실행 2 뒤와 같다.
- 판정 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ ./internal/manifest/ -run '^(TestPriorManifestAtomicWriteFailures|TestPriorGenerationCorruptIsNoEvidence|TestLegacyManifestRoundTrip|TestDowngradeDropsEvidenceNotSafety|TestPriorGenerationsAreLookupBoundOnly)$' -count=1 -v`
- 판정 명령 2: `git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- internal/cli/update/backup`
- 판정 명령 3 (양성 대조군): `git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- .moai/specs`
- 기대: 명령 1 은 다섯 시험의 `--- PASS` (패키지별 `ok` 두 줄), 종료 코드 0. 명령 2 는 빈 stdout 종료 0, 명령 3 은 한 줄 이상.
- RED 이유: 원자적 쓰기·읽기 원천이 없어 빌드 실패 또는 단언 실패.
- 변이: **최종 이름에 직접 쓰는 구현 (`manifest.prior.json` 을 쓰기용으로 열어 바이트를 쓰고 이름 바꾸기가 없음)** — 붉어지는 단언 셋: (o) 의 「임시 쓰기 1건 뒤 이름 바꾸기 1건」 (두 관측점이 한 번도 불리지 않아 기록 0건), (p) 와 (q) 의 「오류를 돌려준다」 (주입한 오류가 이 구현에 닿지 못해 nil), (p) 와 (q) 의 「`manifest.prior.json` 자리에 정규 파일이 없다」 (전체 바이트가 이미 최종 이름에 있다). (r) 은 이 변이도 실패하므로 이 변이를 가르지 못한다 — (r) 의 몫은 실제 이름 바꾸기 실패 경로이다. 그 밖의 쓰기 변이: 임시 이름에 쓰되 이름 바꾸기 대신 최종 이름에 다시 쓰는 구현 (이름 바꾸기 기록 0건; (q) 에서 최종 이름 존재), 이름 바꾸기를 관측점 밖에서 부르는 구현 (이름 바꾸기 기록 0건), 최종 이름을 임시 이름으로 넘기는 구현 (임시 대상 이름 == 최종 이름), 최종 이름에 먼저 쓴 뒤 두 관측점을 장식으로 부르는 구현 (이름 바꾸기 호출 시점에 최종 이름이 이미 존재). 손상 세대에서 오류를 내거나 `managed` 로 올리는 구현, 옛 JSON 저장 시 빈 `deployed_version` 을 채우는 구현 (바이트 불일치), 다운그레이드 항목을 증거로 취급하는 구현, 세대 상한을 삭제 호출에 쓰는 구현 (소스 스캔 붉음), `internal/cli/update/backup` 의 보존 정책을 건드리는 커밋 (명령 2 붉음).
- 한계: 두 관측점을 정상으로 거쳐 임시 쓰기와 이름 바꾸기를 마친 뒤 같은 바이트를 최종 이름에 다시 직접 쓰는 구현은 이 방식이 가르지 못한다 (결과 바이트가 같고, 이름 바꾸기 이전 단계의 실패 주입은 그 쓰기에 닿지 못한다) — 중단 창을 여는 중복이지 최종 이름의 부분 노출이 아니라고 보아 수용한다. `fsync` 의 내구성 효과는 시험할 수 없다 (프로세스 충돌 모사는 이름 바꾸기 전 상태를 직접 만드는 것으로 갈음). 실제 보존 정책 센서스가 `manifest.prior.json` 을 해시 대상으로 삼아 정책 예산을 바꾸는지는 측정하지 않았다 (갭 G-K).
- 비고: Q2 의 기록된 대안 (나) 에서는 이 기준이 삭제된다 (plan.md § G.1).
- 등급: regression-guard → 승격 지점 **M1b**.

## 가장자리 사례

- 손상된 매니페스트: 읽기 전용 로더가 nil 을 돌려주면 모든 파일이 `unregistered` 이다 — 전부 보존되고 이유가 기록된다 (제거 없음).
- 사용자 매니페스트 없음 또는 HOME 해석 불가: 사용자 대응물 상태가 `unconfirmed` 이고 모든 사본이 보존된다.
- 루트 밖으로 나가는 심링크: 기존대로 제거를 거부하며 이유가 `symlink-escape` 로 기록된다.
- 읽기 실패 또는 `TemplateHash` 가 빈 항목: 보존, 이유 `read-failed` 또는 `no-template-hash`.
- Windows 경로 구분자: 판정 함수는 슬래시 경로를 키로 쓴다 (`isCommonAssetCleanTarget` 의 정규화와 같은 방향).

## § M. 변이 자기 점검 (iteration 3 전 — 기준마다 가장 싼 강한 변이)

규칙 9. 각 기준에 대해 요구사항을 어기면서 기준을 통과시킬 수 있는 가장 싼 구현을 적고, 기준의 단언이 그 변이에서 붉어지는지 확인했다. 붉어지지 않던 기준은 이번에 고쳤고 「변경」 열에 적었다. 모든 `moai update` 형 흐름은 조정에 닿는지(건너뛰기 관문·dry-run 조기 반환·매니페스트 없음 조기 반환) 를 코드로 확인했다 (이번 실행, § L-13, design.md D-10).

| AC | 가장 싼 변이 | 붉어지는 단언 | 흐름 도달 확인 | 변경 |
|----|-------------|---------------|----------------|------|
| 01 | 미등록 + 템플릿 바이트 일치를 `managed` 로 | x2 분류 | 순수 함수 | x2 변형 추가 |
| 02 | 해시·템플릿 일치로 승격 | (B) 주입 대역 | 마이그레이션 직접 호출; 제거 암 도달은 형제 사본 제거로 단언 | 도달 대조군 추가 |
| 03 | 현재 바이트 대신 `Provenance` 만 봄 | 수정 파일 제거 | 위와 같음 | 변이 주입과 도달 대조군 추가 (종전엔 변이가 서술뿐) |
| 04 | 이름이 템플릿 디렉터리에 있으면 `managed` | 이유 단언 | 위와 같음 | 변이 주입과 도달 대조군 추가 |
| 05 | 출처값 무시 (버전·경로·해시만) | (e)(f)(g)(h1) | 함수 + 운영 원천 (h) | 출처값 변이 3종과 운영 원천 2건 추가 (감사 N2) |
| 06 | 모든 항목에 버전 / 사본 재직렬화 / 매 실행 `DeployedAt` | 비배포 항목 값, 바이트 비교, 실행 3 바이트 | `--force` 3회 + 도달 단언 | 비정규 시작 파일, 비배포 항목 단언, `--force` 추가 |
| 07 | 프로필 매핑 무시 / 프로필 쪽 상태를 읽은 척 | 사본 제거, `not-checked` 단언 | 쌍둥이 픽스처 (비지도 → 제거) | 도달 대조군과 `not-checked` 변이 추가 |
| 08 | 읽고 버림 (도우미/토큰 접합/함수 안 새 stat) | L1 스캔 + L2 변이 M-a~M-f | 스캔은 호출 닫힘 도출, 카나리는 `--force` | 스캔 범위·심볼·정규화·변이 대조군 전면 재설계 (감사 N1, N6) |
| 09 | commands 재배포가 두 처분에 나타남 | 서로소 단언 | V1 도달·V2 비건너뜀 단언 | 공허 방지 추가 |
| 10 | 복구 실행이 디렉터리를 채우지 않음 | 이름 집합 / 비어 있지 않음 | `--force` 복구 + 건너뛰는 사례 분리 | 변이와 `--force` 추가 |
| 11 | 철자 무시 항상 별칭 | 구분 모드 단언 | 순수 함수 + 호스트 시험 | 변경 없음 |
| 12 | 사용자 매니페스트 키 조회가 철자 무시 안 함 | `confirmed` 단언 | 마이그레이션 직접 호출 | 변경 없음 |
| 13 | 비어 있기만 하면 삭제 | D 사례 | 마이그레이션 직접 호출 | 변경 없음 |
| 14 | 이유 열 비움 / 313 전부 출력 / 승격 | 이유 없는 행 없음 / 기본 렌더 / 대조 실행 | 실행 2 를 `--force` 로 구동 + 도달 단언 | `--force`, 도달 대조군, 규칙 7 추가 (감사 N5) |
| 15 | 사본 줄 끝 바꿈 / 재직렬화 / 사용자 파일도 삭제 | (i)(ii)(iv) | 전체 흐름 + 통제 실행 T0 + 도달 단언 | 전체 트리 동일에서 복원 영역 + 계산된 나머지 차이로 재진술 (감사 N4) |
| 16 | 미리보기가 자기 판정 / 프로브가 루트에 씀 | 후보 집합 불일치, 루트 mtime | `removal-eligible` 비어 있지 않음과 `UserOwned` 포함을 먼저 단언 | 공허 방지 추가 |
| 17 | 매 실행 `DeployedAt` / 일반 재실행이 조정 | 실행 3 바이트, `skipped` | `--force` 3회 + 일반 재실행 사례 | `--force`·건너뛰는 사례·쓸어낸 실행 수 추가 (감사 N5) |
| 18 | 복구 재실행이 건너뜀 / (a) 에서 첫 실행이 일부 제거 | 수렴 단언, (a) 제거 0건 | 재실행 `--force` + 도달 단언 | `--force`, (a) 제거 0건 단언 추가 |
| 19 | `init()` 이나 비상수 대입으로 켬 | 모든 대입 스캔 | 소스 스캔 | 변이 추가 |
| 20 | Backup 단계가 사본 실패를 삼키고 성공 인계 | a 의 제거 목록 | 운영 호출 경로 전체 (`--force`) + 쌍둥이 (c) | 인계 시임·전체 흐름 사례로 재설계 (감사 N3) |
| 21 | 최종 이름에 직접 쓰기 / 손상 세대를 증거로 / 필드 지어냄 / 삭제에 상한 사용 | 직접 쓰기: (o) 의 임시 쓰기→이름 바꾸기 기록, (p)(q) 의 오류와 최종 이름 부재. 나머지는 위 변이 목록 | 쓰기 함수를 두 관측점 (`priorManifestTempWrite`, `priorManifestRename`) 으로 직접 구동 + `--force` 반복 | 신설 (리더 조건 3-6); 감사 N12: 관측점 둘을 더해 직접 쓰기 변이가 붉어지게 고침 |

## 품질 게이트 (Definition of Done)

- AC-OWR-01 ~ -21 이 모두 위 규약의 승격을 거쳤다: 각 기준의 A·B·C 커밋 SHA 와 RED 셀(명령, 원문 stdout, 종료 코드, 트리 SHA, 이유), GREEN 원문이 progress.md § E.2 에 기록되어 있다. 제로 매치로 통과를 주장한 기준이 없다. (AC-OWR-09 가 두 변형으로도 재현되지 않으면 회귀 가드로 남는 사유가 셀에 있다.)
- 기준선 11개 시험과 t1668 시험 (`TestMigrationBackupBeforeRemoval`, `TestEmptiedSkillDirRemoved`, `TestMigrationRemovesSharedCopyWithoutProfileReach`, `TestMigrationRemovalGate`) 이 GREEN 이다. 단 REQ-OWR-016 의 일괄 보존(제거 0건) 이 파일별 계속 제거를 단언하던 시험을 붉게 하면 그 시험은 새 요구사항에 맞게 고친다 (이 트리에서 `backupRemovedAsset` 실패를 단언하는 시험은 없다 — § L-14).
- 터치한 패키지 `go test -cover` 85% 이상. `golangci-lint`, `go vet`, `gofmt` 클린.
- REQ-UPM-031 불변식: 제거가 일어난 실행의 요약이 삭제를 보고한다 (기존 `TestTallyCountsMigrationRemovals` 계승).
- 실제 홈, mo.ai.kr, `moai update` 재실행, 바이너리 교체가 없었다 (시험은 임시 HOME 과 `/private/tmp` 아래 `TMPDIR` 만 썼다).
- 갭 목록이 최신이고 빈 Gaps 선언은 사실일 때만이다.

## 갭 목록

- **G-A (AC-OWR-09, -10)**: 현장의 commands "이전 제외 후 재배포" 가 이 트리에서 재현되는지 미측정. M6 의 첫 관측이 두 변형(같은 버전, 낡은 기록) 으로 소거하거나 AC 를 회귀 가드로 낮춘다. 낡은 기록 변형을 기존 시임만으로 구동할 수 있는지도 미확인.
- **G-B (AC-OWR-11, -12)**: 대소문자 결론은 맵 조회 의미와 호스트 파일시스템 관측의 조합이다. 마이그레이션을 그 철자로 실행한 관측은 RED 셀이 처음 만든다.
- **G-C (AC-OWR-14)**: `.agents/skills` 313개가 심링크 거울 행인지 복사 폴백 파일인지 미확인. 시험은 두 모양을 모두 다룬다 (복사 폴백 313 + 심링크 거울 행).
- **G-D (AC-OWR-05, -06, -15, -20)**: Q2 미결. (나) 가 선택되면 plan.md § G.1 의 뺄셈이 적용된다.
- **G-E (REQ-OWR-005, AC-OWR-08)**: 프로필 쪽 대응물은 읽지 않는다 (Q5, 리더 판정). 프로필 쪽 상태는 항상 `not-checked` 이다. 무접근 검증은 소스 스캔 + 차분 카나리의 한계를 가진다 (AC-OWR-08 한계 항).
- **G-F (spec.md § J)**: 백업 사본의 파일 모드 정규화로 실행 가능 비트가 복원에서 빠질 수 있다. 바이트 동일만 검증한다.
- **G-G (AC-OWR-10)**: Claude Code 의 명령 발견 방식 (경로 대 `name`) 은 이 트리에서 측정하지 않았다. 기준은 템플릿이 운반하는 형식만 단언한다.
- **G-H (AC-OWR-08)**: 기존 런치 원장 판정이 프로필 설정 디렉터리 항목에 하는 `os.Stat` (§ L-8) 는 기준선에 포함되며 이 SPEC 이 바꾸지 않는다.
- **G-I (AC-OWR-16)**: 프로젝트 밖으로의 임시 쓰기와 모드만 바꿨다 되돌리는 쓰기는 디렉터리 mtime 방식이 보지 못한다.
- **G-J (AC-OWR-08)**: 조각난 경로 + 버림 읽기는 L1·L3 어느 쪽도 못 잡는다 (접근 기록 시임 없음). 리더가 이름으로 수락 또는 거부할 항목이다 (decision-index.md 수용 부채).
- **G-K (AC-OWR-21 (e))**: 기존 보존 정책의 센서스가 `manifest.prior.json` 을 해시 대상 파일로 세어 정책 예산(`pastKeep`) 안에서 오래된 실행을 더 남길 수 있는지는 코드 읽기로 가능성만 확인했고 측정하지 않았다. 남기더라도 기존 정책의 한도 안이다.
- **G-L (AC-OWR-20, M1b)**: 마이그레이션 제거 암을 「후보 수집 → 전체 백업 → 일괄 제거」 로 나누는 구조 변경은 t1668 코드와의 병합 충돌 면적을 키운다 (수용 부채, plan.md § D5).
- **G-M (REQ-OWR-013, plan.md OQ-6)**: 일반 같은 버전 재실행이 건너뛰기 관문에서 끝나므로 복구는 `--force` 또는 템플릿·버전 변화가 있는 실행에서만 닿는다. 이 SPEC 은 관문을 바꾸지 않는다. 운영자 확인이 필요하면 별 카드이다.

## § L. 증거 원장 — plan 시점에 이 트리에서 실제로 돌린 셀

트리 SHA: 모두 `15db036a3` (브랜치 `WT-update-ownership-repair`) 에서 측정했다. 판정 도구는 이 트리에서 `go build -o /private/tmp/own-fixture-bin/moai ./cmd/moai` 로 만든 빌드이며 경로로 호출했다 (`verification-claim-integrity.md` § 2.2). 아래는 전제 셀(이 트리의 기존 사실)과 기준선 가드 셀이며 RED 셀이 아니다 — RED 셀은 위 승격 규칙의 B 커밋이 만든다. 명령은 모두 저장소 루트에서 실행했다. `stdout` 은 실행에서 나온 원문이고 `(빈 출력)` 은 실제로 비었다는 뜻이다.

**L-1 — 내장 템플릿에 두 보존 픽스처 경로가 실재한다 (AC-OWR-02 전제)**

```text
$ git ls-files --error-unmatch internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md
stdout: internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md
exit: 0
$ git ls-files --error-unmatch internal/template/templates/.agents/skills/moai-plan/SKILL.md
stdout: internal/template/templates/.agents/skills/moai-plan/SKILL.md
exit: 0
```

**L-2 — 종전 픽스처 경로는 어느 쪽 템플릿에도 없다 (결함 D1 의 증거, 철회 근거)**

```text
$ git ls-files --error-unmatch internal/template/templates/.claude/skills/moai-user-authored/SKILL.md
stdout: (빈 출력)
stderr: error: pathspec 'internal/template/templates/.claude/skills/moai-user-authored/SKILL.md' did not match any file(s) known to git
        Did you forget to 'git add'?
exit: 1
$ git ls-files --error-unmatch internal/template/templates/.agents/skills/moai-user-authored/SKILL.md
stdout: (빈 출력)
stderr: error: pathspec 'internal/template/templates/.agents/skills/moai-user-authored/SKILL.md' did not match any file(s) known to git
        Did you forget to 'git add'?
exit: 1
```

**L-3a — commands 17개는 모두 `name:` 이 없다 (결함 D4 의 증거)**

```text
$ grep -L '^name:' internal/template/templates/.claude/commands/moai/*
stdout:
internal/template/templates/.claude/commands/moai/feedback.md.tmpl
internal/template/templates/.claude/commands/moai/clean.md.tmpl
internal/template/templates/.claude/commands/moai/gtd.md
internal/template/templates/.claude/commands/moai/codemaps.md.tmpl
internal/template/templates/.claude/commands/moai/plan.md.tmpl
internal/template/templates/.claude/commands/moai/fix.md.tmpl
internal/template/templates/.claude/commands/moai/e2e.md.tmpl
internal/template/templates/.claude/commands/moai/goal.md.tmpl
internal/template/templates/.claude/commands/moai/gate.md.tmpl
internal/template/templates/.claude/commands/moai/mx.md.tmpl
internal/template/templates/.claude/commands/moai/loop.md.tmpl
internal/template/templates/.claude/commands/moai/sync.md.tmpl
internal/template/templates/.claude/commands/moai/run.md.tmpl
internal/template/templates/.claude/commands/moai/project.md.tmpl
internal/template/templates/.claude/commands/moai/review.md.tmpl
internal/template/templates/.claude/commands/moai/harness.md.tmpl
internal/template/templates/.claude/commands/moai/todo.md
exit: 0
$ grep -L '^description:' internal/template/templates/.claude/commands/moai/*
stdout: (빈 출력)
exit: 1
```

**L-3b — agents 는 모두 `name:` 이 있다**

```text
$ grep -L '^name:' internal/template/templates/.claude/agents/moai/*
stdout: (빈 출력)
exit: 1
```

**L-4 — 기준선 시험 11개가 이 트리에서 GREEN (AC-OWR-02, -15, -19 의 가드 근거)**

```text
$ TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestMigrationClassifiesUnregisteredMirrorCopy|TestMigrationKeepsSharedCopyUnderProfileReach|TestMigrationRemovesSharedCopyWithoutProfileReach|TestEmptiedSkillDirRemoved|TestMigrationBackupBeforeRemoval|TestMigrationDefaultKeepsConfirmedProjectCopy|TestMigrationRemovalGate|TestMigrationIdempotent|TestMigrationPreservesExistingMirrorEntries|TestRF2F3b_MigrationPreservesUntracked|TestPreviewOmitsRemovalCandidateWhileSwitchOff)$' -count=1 -v
stdout (PASS 줄과 요약만 — 전체 -v 출력은 서브 시험 로그를 포함해 31줄, 승격 B 커밋에서 원문 전체를 기록한다):
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (3.40s)
--- PASS: TestMigrationKeepsSharedCopyUnderProfileReach (2.54s)
--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach (6.70s)
--- PASS: TestMigrationBackupBeforeRemoval (7.72s)
--- PASS: TestEmptiedSkillDirRemoved (8.38s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (2.66s)
--- PASS: TestMigrationRemovalGate (3.43s)
--- PASS: TestRF2F3b_MigrationPreservesUntracked (0.01s)
--- PASS: TestMigrationIdempotent (3.69s)
--- PASS: TestMigrationPreservesExistingMirrorEntries (4.62s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	44.908s
exit: 0
```

**L-5 — 내장 스킬 템플릿 파일 수 (정보용, 파이프 형태라 판정 셀이 아니다)**: `.claude/skills` 아래 315개, `.agents/skills` 아래 17개 (합 332). AC-OWR-14 의 313개는 앞의 315개에서 고른다. 시험의 전제 단계가 개수를 다시 단언한다.

**L-6 — 두 보호 파일의 불변 (AC-OWR-08 명령 2·3 과 정보용 diff)**

```text
$ git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- internal/config/protected_zone.go internal/config/userroots.go
stdout: (빈 출력)
exit: 0
$ git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- .moai/specs
stdout:
a8ad48c08 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): plan-phase artifacts (Tier L, 5 artifacts) (card t1865)
3e77fffd6 docs(SPEC-UPDATE-OWNERSHIP-REPAIR-001): read-only code investigation record (card t1865)
exit: 0
$ git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go
stdout: (빈 출력)
exit: 0
```

**L-7 — `claude-skills@` 는 이 트리의 Go 소스에 없다 (문자열 스캔이 공허한 이유)**

```text
$ grep -rn --include='*.go' 'claude-skills@' internal
stdout: (빈 출력)
exit: 1
$ grep -rn --include='*.go' 'claude-agents@' internal
stdout: (빈 출력)
exit: 1
```

**L-8 — 기존 런치 원장 판정이 프로필 설정 디렉터리 항목을 stat 한다**

```text
$ grep -n 'os.Stat(filepath.Join(baseDir, name))' internal/profile/profile.go
stdout:
439:	info, err := os.Stat(filepath.Join(baseDir, name))
575:	if info, err := os.Stat(filepath.Join(baseDir, name)); err != nil || !info.IsDir() {
exit: 0
```

439 는 `launchCandidateIsUsable` (읽기 경로, `ResolveLaunchProfileForProject` 가 부른다), 575 는 원장 기록 쪽의 존재 확인이며 갱신 경로가 아니다.

**L-9 — `probeCaseSensitiveFS` 는 비시험 호출자가 없다**

```text
$ grep -rn --include='*.go' --exclude='*_test.go' 'probeCaseSensitiveFS(' internal
stdout: internal/cli/update_cleanup.go:387:func probeCaseSensitiveFS(projectRoot string) (caseSensitive bool, _ error) {
exit: 0
```

**L-10 — 프로필 경로 구성 토큰의 기준선 파일 (AC-OWR-08 (a), 19개)**

```text
$ grep -rlE 'GetBaseDir|BaseDirOverride|EnvClaudeConfigDir|CLAUDE_CONFIG_DIR|claude-profiles|profilesDir|GetCurrentName' --include='*.go' --exclude='*_test.go' internal/cli internal/userassets internal/manifest
stdout:
internal/cli/memory.go
internal/cli/update_participation.go
internal/cli/doctor_mcp_provider.go
internal/cli/web.go
internal/cli/glm.go
internal/cli/harness_validate.go
internal/cli/cc.go
internal/cli/update.go
internal/cli/migrate_profiles.go
internal/cli/profile.go
internal/cli/clean_home.go
internal/cli/doctor_disk.go
internal/cli/profile_setup.go
internal/cli/launcher_model_source.go
internal/cli/downgrade_locale.go
internal/cli/init.go
internal/cli/ptycaptest/harness.go
internal/cli/worktree/launch_ledger.go
internal/cli/ptycaptest/homewatch.go
exit: 0
```

마이그레이션·미리보기·갱신 동기화 파일 (`migrate_project_assets.go`, `update_template_sync.go`, `update_dryrun_preview.go`) 은 이 목록에 없다 — 토큰이 없다. AC-OWR-08 의 토큰 집합은 위 명령의 7개에 `claude-skills@`·`claude-agents@` 를 더한 9개이며, 두 `@` 토큰은 L-7 에서 0건이라 9개 토큰으로 센 파일 수도 19개로 같다 (`grep -rlE` 9개 토큰 + `wc -l` 의 관측값, 정보용).

**L-11 — 갱신 경로 호출 닫힘 (AC-OWR-08 (a) 의 도출 대상; 트리 `5fb8920ee`, 이번 실행)**

```text
$ go list -deps ./internal/cli
(모듈 안 패키지만 센 결과) 142
핵심 여섯 패키지 포함 확인 (grep -E 'internal/(profile|userassets|manifest|template|config|core/project)$'):
internal/config, internal/profile, internal/manifest, internal/template, internal/core/project, internal/userassets
exit: 0
```

**L-12 — `internal/profile` 의 내보낸 함수 (AC-OWR-08 L1 의 ① 집합; 시험이 기계적으로 열거하므로 이 목록은 정보용)**

```text
$ grep -n '^func [A-Z]' internal/profile/profile.go internal/profile/preferences.go internal/profile/sync.go
profile.go: GetBaseDir, GetCurrentName, GetCurrentNameForProject, List, Delete, GetProfileDir, IsValidProfileName, EnsureDir,
            ResolveLaunchProfile, ResolveLaunchProfileForProject, HasClaudeConfig, RecordLastUsedProfile,
            RecordLastUsedProfileForProject, PruneStaleProjectEntries
preferences.go: IsValidPermissionMode, GetPreferencesPath, IsSetup, ReadPreferences, WritePreferences
sync.go: SyncToProjectConfig
갱신 경로의 기존 호출 지점 (grep 'profile\.' 비시험, 다섯 파일): migrate_project_assets.go:185 ResolveLaunchProfileForProject;
  update.go:693 GetCurrentNameForProject, :694 ReadPreferences, :698 SyncToProjectConfig; update_participation.go:49 ReadPreferences(GetCurrentName())
exit: 0
```

`GetProfileDir` (profile.go:193) 는 `filepath.Join(GetBaseDir(), name)` 을 돌려주는 내보낸 경로 생산 도우미이고 갱신 경로에 호출 지점이 없다 — 그래서 새 호출 지점이 곧 붉음이다. `launchCandidateIsUsable` 의 `os.Stat(filepath.Join(baseDir, name))` 은 `profile.go:439`.

**L-13 — 버전 일치 건너뛰기 관문의 실측 (AC-OWR-06, -10, -14, -17, -18 의 도달 근거; overlay 탐침, 트리 파일은 건드리지 않음)**

```text
$ go test -overlay=<임시 overlay> ./internal/cli/ -run '^TestProbeSkip$' -count=1 -v   (buildMigrationFixture 로 만든 같은 프로젝트에 runTemplateSyncWithProgress 를 순서대로 호출)
    force=false skipped=false      # 실행 1 (일반): 조정에 닿음
    force=false skipped=true       # 실행 2 (일반, 같은 버전): 건너뜀
    force=true  skipped=false      # 실행 3 (--force): 조정에 닿음
    force=false skipped=true       # 실행 4 (일반): 건너뜀
--- PASS: TestProbeSkip (1.38s)
exit: 0
```

관문 코드는 `update_template_sync.go:1239-1248`, 판정은 `versionMatchSkipsSync` (`update_template_sync_guard.go:55-68`). 이 측정은 plan-auditor 의 코드 읽기 결론과 codex 의 overlay 측정(`Up to date · Skipping sync`, 마이그레이션 추가 호출 0) 과 일치한다.

**L-14 — `backupRemovedAsset` 실패를 단언하는 기존 시험은 없다 (일괄 보존으로 바꿔도 깨지는 시험 없음)**

```text
$ grep -rn 'backupRemovedAsset\|backup-failed\|backup failed' --include='*_test.go' internal/cli
stdout: (빈 출력)
exit: 1
```

운영 코드의 해당 분기: `migrate_project_assets.go:310` (`removal refused (the run-backup copy failed)`).

**L-15 — 출처값 상수 (AC-OWR-05, N2)**

```text
$ grep -n 'Provenance = ' internal/manifest/types.go
TemplateManaged Provenance = "template_managed"   (:17)
UserModified Provenance = "user_modified"          (:21)
UserCreated Provenance = "user_created"            (:25)
Deprecated Provenance = "deprecated"               (:29)
GeneratedManaged Provenance = "generated_managed"  (:35)
exit: 0
```

배포기의 미등록 기존 파일 처리: `internal/template/deployer.go:265-267` (`m.Track(destRelPath, manifest.UserCreated, templateHash)` 뒤 건너뜀).
