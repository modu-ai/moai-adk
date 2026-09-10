# t659 판정서 (1차) — `moai constitution amend` 적용 단계 미구현 재현과 설계 질문

- 카드: t659 · 레인: lane-6 · 브랜치: `WT-amend-apply` (워크트리 `.claude/worktrees/t659`)
- 기준: origin/develop `987eb7e40` 에서 생성 → 로컬 develop `ed71054d3` 를 `--no-ff` 흡수 → `5a066994b` (HEAD^2 = `ed71054d3`)
- 도구: `go version go1.26.8 darwin/arm64`
- 상태: 재현 완료, 설계 판단이 나와 멈춤. 코드 변경 없음(임시 관측 테스트는 실행 뒤 삭제).

## 1. Claim

1. 실제 적용 경로(`Execute`, dry-run 아님)는 5개 게이트와 승인을 통과한 뒤 `updateSourceFile` 에서 `not yet implemented` 로 끝난다. 등록부 갱신과 evolution-log 기록 줄에는 도달하지 않는다.
2. dry-run 경로는 적용 함수를 전혀 부르지 않고 성공을 보고한다. 미리 돌려 봐도 실제 적용 실패는 드러나지 않는다.
3. evolution-log 의 읽기·쓰기 키가 테스트 픽스처와 SPEC 표기(snake_case)와 다르다. snake_case 로 쓴 기록은 `RuleID`·`ApprovedAt` 가 비어 읽힌다.
4. 실제 `.moai/research/evolution-log.md` 는 코드의 `---` 분할 파서로 기록 0건으로 읽힌다.
5. 등록부의 살아 있는 항목 97개는 모두 원문 파일에 clause 가 정확히 한 번 나온다(폐기 표시 4개 제외).

## 2. Evidence

### 2.1 패키지 수준 재현 (`repro-package.txt`, 사전 확인 `precheck-repro.txt`)

사전 확인: 2026-09-10T18:43:21Z, load 11.78, 다른 레인 go test 1건(internal/hook).

```
unset … && go test ./internal/constitution/ -count=1 -v -timeout 300s -run '^(TestPipeline_Execute_DryRun_Success|TestPipeline_Execute_NonDryRun_AmendmentStubError|TestPipeline_applyAmendment_StubError|TestUpdateSourceFile_StubError|TestUpdateRegistryClause_StubError)$'
--- PASS: TestPipeline_Execute_DryRun_Success (0.00s)
--- PASS: TestPipeline_Execute_NonDryRun_AmendmentStubError (0.00s)
--- PASS: TestPipeline_applyAmendment_StubError (0.00s)
--- PASS: TestUpdateSourceFile_StubError (0.00s)
--- PASS: TestUpdateRegistryClause_StubError (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.357s
EXIT=0
```

이 5개는 미구현 상태를 고정하는 특성화 테스트다. 통과가 곧 실패 지점 관측이다.
- `NonDryRun_AmendmentStubError`: 격리 픽스처(`t.TempDir()` 등록부, 승인 대역, 락 경로 고정)에서 `Execute(dryRun=false)` 가 `amendment application error` 를 반환한다.
- `applyAmendment_StubError`: 도달 가능한 경로는 `source file update error` 뿐이다.

코드 근거(`5a066994b` 기준):

- `internal/constitution/pipeline.go:190-216` `applyAmendment` — 1단계 `updateSourceFile` 오류 시 즉시 반환
- `internal/constitution/pipeline.go:256-260` `updateSourceFile` → `fmt.Errorf("updateSourceFile: not yet implemented …")`
- `internal/constitution/pipeline.go:264-267` `updateRegistryClause` → `fmt.Errorf("updateRegistryClause: not yet implemented …")`
- `internal/constitution/pipeline.go:133-137` dry-run 분기: `createLogEntry` 만 반환, 적용 함수 미호출
- `internal/cli/constitution.go:542-544` CLI 는 오류를 `amendment failed: %w` 로 감싸 반환한다.

### 2.2 evolution-log 키 관측 (`probe-log-keys.txt`, 관측 테스트 사본 `tools/probe_log_keys_test.go.txt`, 실행 뒤 원본 삭제)

```
READ id="LEARN-20260428-001" ruleID="" clauseBefore="" approvedBy="" approvedAtZero=true
WRITTEN:
  id: LEARN-20260911-001
  ruleid: CONST-V3R2-003
  zonebefore: 0
  zoneafter: 0
  clausebefore: a
  clauseafter: b
  canaryverdict: ""
  contradictions: []
  approvedby: human
  approvedat: 2026-09-11T00:00:00Z
  rolledback: false
  rollbackreason: ""
  rollbackat: null
ROUNDTRIP id="LEARN-20260911-001" ruleID="CONST-V3R2-003" approvedAtZero=false
EXIT=0
```

`AmendmentLog`(`internal/constitution/amendment.go:192-220`)에는 yaml 태그가 없다. 그래서 키가 필드명 소문자 연결형이 되고, zone 은 정수로 기록된다. 기존 `TestLoadEvolutionLogs` 픽스처는 `rule_id` 등 snake_case 를 쓰지만 `ID` 만 단언해 이 불일치를 보지 못한다.

### 2.3 실제 파일 조사 (`clause-census.txt`, 도구 `tools/clause_census.py`, 읽기 전용)

```
registry_entries 101
buckets {'missing_file': 0, 'clause_0': 4, 'clause_1': 97, 'clause_2plus': 0}
anchor_slug_found 97
clause_0_ids ['CONST-V3R2-021', 'CONST-V3R2-022', 'CONST-V3R2-023', 'CONST-V3R2-024']
normalized_count CONST-V3R2-021 0 '[SUPERSEDED by worktree-opt-in policy — see CLAUDE.md §14 + '
(022·023·024 동일)
evolution_log_dash_split_parts 14
odd_segments_with_top_level_id_key 0
EXIT=0
```

- 등록부는 ```yaml 펜스 하나에 101항목이 있고, clause 는 모두 큰따옴표 한 줄 문자열이다.
- `anchor_slug_found` 는 이 도구의 근사 slug 규칙으로 센 값이다(검증기와 같은 규칙인지는 미확인).
- 실제 evolution-log 는 사람이 쓴 `## EVO-…` + ```yaml 블록 형식이다. 표 구분선 `|---|` 까지 `---` 분할에 잘려 `id:` 를 가진 조각이 0개다.

### 2.4 부재 확인 (대조군 포함)

```
grep -rn 'rejected-amendments' internal cmd pkg --include='*.go'   → 0줄   (REQ-CON-002-012 미구현)
grep -rn 'evolution-log' internal cmd pkg --include='*.go' | grep -v _test.go | wc -l → 13 (같은 범위 대조군)
grep -rln 'runConstitutionAmend|newConstitutionAmendCmd|"amend"' internal/cli --include='*_test.go' → 0개
grep -rln 'newConstitutionCmd|runConstitution' internal/cli --include='*_test.go' → 4개 (대조군)
grep -rn 'SentinelAnchorNotFound' internal --include='*.go' → validator.go:27·28 정의 2줄뿐 (사용처 0)
grep -rn 'SentinelDrift' internal --include='*.go' | wc -l → 6 (같은 파일 형제 상수 대조군)
```

SPEC `SPEC-V3R2-CON-002` frontmatter 는 `status: implemented` 인데, REQ-CON-002-011(세 파일 원자적 적용)은 위와 같이 구현되지 않았다.

## 3. Baseline-attribution

작업 트리 `5a066994b`(흡수 병합). 재현 테스트와 관측 테스트는 `internal/constitution` 패키지만 컴파일했다. `internal/cli` 는 컴파일하지 않았다. 관측 테스트 파일은 실행 뒤 삭제했고, 사본 sha256 `e05b230c…` 가 원본과 같다.

## 4. Gaps

- **CLI 수준 실행 미관측.** `moai constitution amend --dry-run` / 실적용을 격리 픽스처에서 바이너리로 돌리지 않았다. 바이너리 빌드는 `internal/cli` 컴파일이라 슬롯 승인이 필요하다. 실적용은 표준 입력 `Y` 가 필요하고, 락 파일이 현재 폴더 기준 상대 경로(`.moai/research/.amendment.lock`)라 cwd 를 임시 폴더로 둬야 한다.
- 따라서 CLI 출력 문구·종료 코드는 코드 판독뿐이다(`RunE` 가 오류를 반환하므로 비0 종료로 추정).
- 실제 저장소의 `zone-registry.md`·원문 규칙 파일에는 쓰지 않았고, 실제 경로로 적용을 시도하지 않았다.
- anchor slug 규칙은 조사 도구의 근사다. 검증기(`validator.go`)는 `SentinelAnchorNotFound` 를 정의만 하고 내보내지 않으며(2.4), clause 는 공백 정규화 후 파일 전체 부분 문자열로만 본다. 따라서 anchor 절 범위 판정에 재사용할 기존 규칙이 없다.

## 5. Residual-risk

- 구현 시 기존 특성화 테스트 5개가 반대로 뒤집혀야 한다. 그대로 두면 구현이 테스트를 깨거나, 테스트를 지우는 쪽으로 기울 수 있다.
- evolution-log 형식을 코드 형식으로 맞추면 사람이 쓴 기존 로그(EVO-HRN-002)와 한 파일에 두 형식이 섞인다. RateLimiter 는 지금도 실제 로그를 0건으로 본다.

## 6. 멈춤 — 리드 판단이 필요한 설계 질문

1. **원문 교체 규칙** — 살아 있는 97항목은 clause 가 파일에 정확히 1회라 "파일 전체에서 정확히 1회 일치할 때만 교체, 0회·2회 이상이면 거부"가 가능하다. 그러나 anchor 절 범위로 좁힐지, 공백 정규화 일치를 허용할지(검증기는 정규화 비교)는 정해야 한다.
2. **등록부 갱신 형식** — ```yaml 펜스 안의 해당 `clause:` 한 줄만 문자열로 바꿀지(주석·순서·따옴표 보존), YAML 을 다시 직렬화할지. 재직렬화는 101항목 파일 전체의 서식을 바꾼다.
3. **evolution-log 스키마** — (a) `AmendmentLog` 에 snake_case yaml 태그를 달아 SPEC·테스트 표기에 맞출지, (b) 현 소문자 연결형을 유지할지. 사람이 쓴 기존 로그 형식과 같은 파일에 공존시킬지, 파서가 어느 형식을 읽어야 하는지도 정해야 한다. zone 정수 직렬화도 같이 걸린다.
4. **원자성** — REQ-CON-002-011 은 세 파일 전부 또는 전무를 요구한다. 임시 파일에 셋 다 쓴 뒤 rename 하는 방식으로 할지, 실패 시 앞 단계 복구 방식으로 할지. rename 원자성은 파일 단위라, 세 파일 사이의 원자성은 추가 설계가 필요하다.
5. **범위** — REQ-CON-002-012(거부 기록)와 SPEC 상태 `implemented` 정정을 이 카드에 넣을지, 별도 카드로 뺄지.

## 7. 리드 판정 (재현 판독 뒤)

- **Q1 원문 교체**: 파일 전체에서 clause 문자열이 정확히 1회 일치할 때만 허용. 0회·2회 이상은 적용 전에 실패하고 원문은 바뀌지 않는다. 공백 정규화 금지(조용한 오일치 위험), anchor 절 범위 좁히기 불필요(살아 있는 97/97 이 이미 1회).
- **Q2 등록부 갱신**: yaml 펜스 안 해당 `clause:` 한 줄 문자열 치환(서식 보존) + 쓰기 직후 재파싱 검증(항목 수·해당 항목 clause 일치). 재직렬화 기각.
- **Q3 evolution-log 스키마**: snake_case yaml 태그 명시, 읽기는 기존 소문자 연결형 키도 호환. zone 표현은 등록부 형식을 따른다. 실제 파일(사람 작성 EVO-HRN-002 포함)을 파서가 읽도록 고치는 것은 이 카드 범위 — 기록만 되고 RateLimiter 가 못 읽으면 게이트가 무력하다.
- **Q4 원자성**: 세 파일 적용 전 백업 → 각 파일 임시 쓰기 → 원문·등록부·로그 순 rename, 어느 단계든 실패 시 백업으로 전부 복원. 실패 주입 테스트(2번째·3번째 rename 실패)로 복원을 고정. dry-run 은 적용 함수의 검증 단계(1회 일치·재파싱)까지 실제로 실행해 사전 점검에서 실패가 드러나게 한다.
- **Q5 범위**: 분리. REQ-CON-002-012 거부 기록 구현과 SPEC-V3R2-CON-002 상태 정정은 별도 카드 후보(큐 복구 뒤 발행).
- **CLI 슬롯**: 부여하지 않음. CLI 수준 동작은 run 단계에서 `internal/cli` 테스트(홈 시접)로 확인한다.
- 다음: plan(manager-spec).

### 7.1 별도 카드 후보

1. REQ-CON-002-012 — 안전 게이트 실패 시 `.moai/research/rejected-amendments/` 기록 구현(현재 코드 0건).
2. SPEC-V3R2-CON-002 frontmatter `status: implemented` 정정 — REQ-011 이 이 카드로 착지하기 전까지 사실과 다르다.

## 8. 리드 판정 — plan 설계 공백 (SPEC-CON-AMEND-APPLY-001 `7b4d1ac89` 판독 뒤)

- **G1 로그 해석 실패**: fail-closed 확인(Frozen 게이트는 편의보다 안전). 오류는 해석할 수 없는 항목을 파일·줄·키로 이름 대야 한다. AC 로 고정.
- **G2 새 clause 중복**: 적용 전 새 clause(after) 문자열이 원문에 0회여야 한다. 아니면 적용 거부, 원문 무변경.
- **G3 Before 검사**: `Execute` 에 넣는다(방어 심층). Before 가 현재 clause 와 다르면 `Execute` 가 거부, CLI 검사는 조기 안내로 유지.
- **G4 복원 실패**: 채택. 백업 삭제 금지, 백업 경로를 담아 오류 반환. 복원 실패 경로도 실패 주입 테스트 1개.
- **G5 CLI 검증**: 축소 승인. CLI 는 dry-run 만 테스트, 비-dry-run CLI 경로는 Gap(Execute 수준 테스트가 적용을 커버).
- **범위 추가 (a)**: 등록부 경로를 단일 해석기로 통일하고 AC 1개. 이 카드가 실제 쓰기를 켜는 순간, CLI 가 검증한 등록부와 `Execute` 가 쓰는 등록부가 다를 수 있다는 점은 잠재 결함이 아니라 쓰기 사고가 된다.
- **progress.md**: plan-audit 전에 만든다(run 증거 자리).
- 다음: SPEC 반영 → plan-auditor.

### 8.1 후속 카드 후보 추가 (큐 이전 뒤 발행)

3. dry-run 환경 변수 값 불일치 — SPEC-V3R2-CON-002 REQ-031·AC-07 은 `MOAI_CONSTITUTION_DRY_RUN=1`, `internal/cli/constitution.go:470` 은 `== "true"`.
4. EVO-HRN-002 항목 오기 — `const_registry_entry: CONST-V3R2-153` 인데 그 등록부 항목의 file 은 `.claude/rules/moai/workflow/session-handoff.md`(zone-registry.md:678-681), 로그의 `target_file` 은 `.claude/rules/moai/design/constitution.md`.
5. `MarkRolledBack` 손실 재직렬화 위험 — 파서가 사람 작성 항목을 읽게 되면 `rewriteEvolutionLog` 가 그 항목을 코드 형식으로 다시 써서 원래 필드를 잃을 수 있다(에이전트 판독, 프로덕션 호출자 0 — 미실측).
