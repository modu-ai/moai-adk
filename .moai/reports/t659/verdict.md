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

## 9. 리드 판정 — G6 (SPEC 0.1.1 `01af243ae` 판독 뒤)

- **G6 경로 경계**: (ii) 채택. `LoadRegistry` 가 projectDir 밖 등록부 경로를 거부하는 동작(`internal/constitution/loader.go:79-88`)을 의도된 경계로 명시한다.
  - REQ: 등록부·원문 규칙 파일·evolution-log 는 모두 같은 프로젝트 루트 안에서만 온다(불변식).
  - AC: `CLAUDE_PROJECT_DIR` 가 다른 트리를 가리키면 쓰기 전 적재 오류로 멈추고 어느 파일도 바뀌지 않는다. 대조: 같은 트리를 가리키면 통과한다.
  - (i) 원문·로그도 해석된 루트를 따르게 확장하는 안은 쓰기 대상 트리를 넓히는 방향이라 기각.
- **설치 바이너리 SIGKILL**(`/Users/goos/go/bin/moai`, 04:14 교체 뒤 `moai version` exit 137): 리드가 확인한다. 레인은 재설치하지 않는다.
- 순서: G6 SPEC 반영 → 바이너리 복구 확인 뒤 `moai spec lint` 1회 → plan-auditor.

## 10. SPEC 0.1.2 확인 · lint · 새 공백 G7

### 10.1 레인 직접 확인 (`e693f0583`)

- 커밋 `e693f0583` 은 SPEC 파일 4개만 바꿨다(69+/28-). 작업 트리 변경 0.
- `version: "0.1.2"`. REQ-CAA 20개(acceptance.md 에서 20/20 참조), AC 23개, 뮤턴트 25개(`grep -o … | sort -u | wc -l`).
- Cf 문자 네 파일 모두 0(대조 1).
- G6 반영 ID: REQ-CAA-020, AC-CAA-023(`divergent_root_real`·`divergent_root_dry_run`·`same_root_control`), M-20.

### 10.2 lint (`lint-0.1.2.txt`, 판정 바이너리 `lint-binary.txt`)

```
/Users/goos/go/bin/moai spec lint SPEC-CON-AMEND-APPLY-001
INFO  OwnershipTransitionUnmeasured  …/spec.md  1  … commit 7b4d1ac89… has no Authored-By-Agent trailer — ownership transition unmeasured
0 error(s), 0 warning(s)
LINT_EXIT=0
```

- 판정 바이너리: `v3.2.0-rc.7   moai_cp/20260910_130400-275-ged71054d3-dirty   built 2026-09-10T19:18:41Z`, `VERSION_EXIT=0`.
- 판정 트리: `e693f0583b3490620e9d40f82d9895f76ab594b3`.
- 귀속: `git merge-base --is-ancestor ed71054d3 92c8c3f36` → exit 0. `git diff --stat ed71054d3 92c8c3f36 -- '*.go' internal cmd pkg` → 출력 없음(같은 범위 전체 diff 는 11 files 로 대조). 따라서 lint 규칙 코드는 커밋 기준으로 이 트리와 같다. `-dirty` 는 빌드 트리에 미커밋 변경이 있었다는 뜻이며 그 내용은 관측하지 않았다(Gap).

### 10.3 새 공백 G7 — 에이전트 판독, 레인이 인용 줄을 확인

REQ-CAA-020 이 "다른 트리에는 읽기도 쓰기도 닿지 않는다"고 선언했지만, 현재 검사가 막지 못하는 모양이 두 가지다(실행 재현 없음).

1. 상대 등록부 경로 — `internal/constitution/loader.go:82` 는 `filepath.IsAbs(cleanPath)` 일 때만 경계를 검사한다. `MOAI_CONSTITUTION_REGISTRY` 나 `CLAUDE_PROJECT_DIR` 에 `../other` 같은 상대값이 오면 검사 없이 프로세스 cwd 기준으로 읽는다.
2. 등록부 항목 `file:` 이 절대경로이거나 `..` 를 포함 — `internal/constitution/pipeline.go:192-195` 는 상대경로면 `projectDir` 에 join 하고, 경계 검사가 없다.
   - 실제 등록부 `file:` 101줄 중 `/` 시작 0, `..` 포함 0(대조: `.claude/` 시작 87).

리드 판단 필요: (A) 두 모양 모두 REQ-CAA-020 으로 강제하고 AC·뮤턴트 추가, (B) REQ 를 절대 등록부 경로로 좁히고 두 모양은 §F 에 기록만.

## 11. 리드 판정 — G7

- **G7 경로 경계**: (A) 채택. 이 카드가 실제 쓰기를 켜는 순간 루트 밖 쓰기는 경로 순회 결함(보안)이 된다. 선언한 불변식을 반쪽만 강제하는 SPEC 은 만들지 않는다.
- **구현 규칙**: 경계 검사는 `filepath.Clean` + 절대화(`filepath.Abs`·심볼릭 링크 해석) 뒤 루트 접두 비교로, 상대값·`..`·절대값을 한 경로로 처리한다. 같은 검사를 다음 셋 모두에 적용한다.
  1. 등록부 경로(환경 변수의 상대값 포함)
  2. 등록부 항목 `file:` 을 join 한 결과
  3. evolution-log 경로
- **거부**: 쓰기 전 적재 오류. 원문·등록부·로그 무변경.
- **AC**: `../other` 상대 env · 절대 `file:` · `..` 포함 `file:` · 루트 안 대조, 각 1행. 실제 등록부 101줄(절대 0·`..` 0)에서 기존 동작 회귀가 없음을 AC 로.
- **뮤턴트**: (1) 검사 제거 (2) Clean 없이 문자열 접두 비교 — `/root-evil` 이 `/root` 접두를 통과하는 모양.
- 순서: SPEC 반영 → lint → plan-auditor.

## 12. SPEC 0.1.3 완성 확인 · lint

### 12.1 경과

- 0.1.3 첫 반영 에이전트가 세션 한도(HTTP 429, `resets 8am (Asia/Seoul)`, request id `req_011CevKfhxMhJPGoyGMrqYpd`)로 커밋 없이 중단했다. spec.md·plan.md·progress.md 만 고쳐져 있었고 acceptance.md 는 그대로였다(신규 ID 등장 0, 대조 AC-CAA-023 6·M-20 4). 레인이 `4e9273d0b`(`wip(t659)`, 첫 줄 INCOMPLETE AND INCONSISTENT)로 보존했다.
- 리드 재개 지시에 따라 새 manager-spec 이 acceptance.md 를 채우고 네 파일을 맞춰 `fa966740d` 로 커밋했다.

### 12.2 레인 직접 확인 (`fa966740d`)

- 커밋은 SPEC 파일 4개만 바꿨다(71+/15-). 작업 트리 변경 0.
- `spec.md` `version: "0.1.3"`. REQ-CAA 21(acceptance.md 에서 21/21 참조), AC 23→25, 뮤턴트 25→29(`grep -o … | sort -u | wc -l`). progress.md §E.1 `counts: 21 requirements, 25 acceptance criteria, 29 mutants` 와 일치.
- acceptance.md 안 등장 횟수(`grep -o -w`): AC-CAA-024 13 · AC-CAA-025 6 · M-21 6 · M-22 5 · M-23 4 · M-24 8 · REQ-CAA-021 9. 대조: AC-CAA-023 10 · M-20 7.
- Cf 문자 네 파일 모두 0(대조 1).

### 12.3 lint (`lint-0.1.3.txt`, 판정 바이너리 `lint-binary-0.1.3.txt`)

```
/Users/goos/go/bin/moai spec lint SPEC-CON-AMEND-APPLY-001
INFO  OwnershipTransitionUnmeasured  …/spec.md  1  … commit 7b4d1ac89… has no Authored-By-Agent trailer
0 error(s), 0 warning(s)
LINT_EXIT=0
```

- 판정 바이너리: `v3.2.0-rc.7   moai_cp/20260910_130400-275-ged71054d3-dirty   built 2026-09-10T19:18:41Z`, `VERSION_EXIT=0`.
- 판정 트리: `fa966740d`. 바이너리 기준 `ed71054d3` 는 이 트리의 조상이고, 그 사이 Go 코드 변경은 §10.2 에서 0 으로 확인했으며 이후 커밋도 SPEC·판정서 문서뿐이다. `-dirty` 내용은 미관측(Gap).

### 12.4 에이전트가 §11 범위보다 넓힌 결정 (리드 확인 대상)

1. AC-CAA-024 에 CLI dry-run 사례 추가 — REQ-CAA-021 의 "CLI 와 Execute 가 같은 검사를 쓴다" 조항을 확인하는 인수 조건이 없어서.
2. M-22 를 변형 둘로 분리 — `filepath.Join` 이 이미 Clean 을 수행해 "Clean 없는 접두 비교"가 join 경로에서는 드러나지 않으므로, (i) 경계 없는 접두 비교(`sibling_prefix_file` RED), (ii) 문자열 이어 붙이기 후 Clean·해석 없는 접두 비교(`dotdot_file` RED)로 나눔.
- 함께 정리: plan.md 의 `M-21a`/`M-21b` 를 M-21 하나의 변형 둘로 합쳐 뮤턴트 29 유지.

### 12.5 에이전트 확장 2건 — 리드 수용

- **AC-CAA-024 CLI dry-run 사례: 수용.** REQ-CAA-021 의 "CLI 와 Execute 가 같은 검사를 쓴다" 조항을 검증하는 AC 가 없던 구멍을 메운 것이라 요구사항 범위 안이다.
- **M-22 변형 둘 분리 + M-21a/b 를 M-21 하나로 합침(총 29 유지): 수용.** `filepath.Join` 이 이미 Clean 을 수행해 한 변형으로는 결함이 드러나지 않는 문제를 고친 뮤턴트 설계 개선이다.
- 둘 다 새 요구사항·범위가 아니므로 되돌리지 않는다. 심볼릭 링크를 만들 수 없는 환경(Windows)에서 M-23·M-24 가 skip 되는 것을 미관측 Gap 으로 적은 표기도 유지한다.

### 12.6 tier 예산 초과 — 감사에 명시해 넘김

- `spec-workflow.md` § SPEC Complexity Tier 의 REQ/AC 상한(각각 독립, M 16 · L 25)에 대해, frontmatter `tier: M` 인 이 SPEC 은 REQ 21 · AC 25 로 둘 다 M 상한을 넘는다(`fa966740d` 레인 계수).
- 레인은 리드에게 (A) tier L 상향 (B) 분리 (C) 감사 판정을 따름을 제시했고, 리드의 plan-auditor 진행 지시에 따라 (C)로 plan-auditor 에 이 초과를 명시해 넘긴다.

## 13. tier 결정 — 운영자: Tier L (§12.6 정정)

### 13.1 §12.6 정정

- §12.6 은 "(C)로 plan-auditor 에 이 초과를 명시해 넘긴다"고 적었다. 이는 실행되지 않았다. 리드는 `fa966740d` 에서 직접 계수(tier: M, REQ 21, AC 25)로 초과를 확인하고, (C)가 상한 규칙(초과는 tier 상향 또는 분리 신호)에 어긋난다며 금지했다. 운영자 결정이 올 때까지 plan-auditor 를 보류하라고 지시했다.
- 리드 참고: (B) 경로 경계만 분리해도 본 SPEC 은 REQ 18 · AC 21 로 여전히 M 상한을 넘는다.

### 13.2 중단된 감사

- 레인은 보류 지시가 도착하기 직전 `f2ba39c94` 커밋 뒤 plan-auditor(opus)를 띄웠다. 지시를 받은 즉시 TaskStop 으로 중단했다(`status: killed`).
- 산출물 없음(레인 확인): `git status --porcelain --untracked-files=all` 출력 없음, HEAD `f2ba39c94` 그대로, `find .moai/reports/t659 -maxdepth 1 -name 'plan-audit*'` 결과 0건(같은 명령의 대조 `verdict.md` 1건). 이 감사는 판정 근거가 아니며, 중단된 감사자에게 메시지를 보내지 않는다.

### 13.3 운영자 결정 — Tier L

- frontmatter `tier: L`, `design.md` · `research.md` 추가. 기존 결정(G1~G7·§11·§12)을 설계·조사 근거로 옮기되 요구사항·결정·범위와 REQ 21 · AC 25 · 뮤턴트 29 는 바꾸지 않는다. `progress.md` 에 산출물 수를 반영한다.
- 레인 확인 항목: 파일 6개 존재, lint 0/0, 신규 ID 등장 횟수와 대조군.
- 그 뒤 plan-auditor(Tier L 합격선 0.85)에 "AC 25 는 Tier L 상한과 같다"를 명시해 넘긴다. 판정 파일은 회차 번호를 붙여 `.moai/reports/t659/plan-audit-iter1.md` 로 받는다.

## 14. Tier L 반영 확인 · lint

### 14.1 레인 직접 확인 (`51454e5e5`)

- 커밋 `51454e5e5` 는 SPEC 폴더만 바꿨다: `design.md` 307+ · `research.md` 273+ 신규, `spec.md` · `plan.md` · `progress.md` 수정, `acceptance.md` 무변경(5 files, 597+/11-). 작업 트리 변경 0.
- 산출물 6개 존재: acceptance.md · design.md · plan.md · progress.md · research.md · spec.md.
- `spec.md`: `version: "0.1.4"`, `status: draft`, `tier: L`.
- 구별 ID 수(여섯 파일 전체): REQ-CAA 21, AC-CAA 25. 뮤턴트(acceptance.md) 29. 변경 전(§12.2)과 같다.
- 부속 문서 `^status:` 줄 수: plan.md 0 · acceptance.md 0 · design.md 0 · research.md 0.
- acceptance.md 등장 횟수(`grep -o -w`): AC-CAA-024 13 · AC-CAA-025 6 · M-21 6 · M-22 5 · M-23 4 · M-24 8 · REQ-CAA-021 9. 대조: AC-CAA-023 10 · M-20 7. §12.2 와 모두 같다.
- Cf 문자 여섯 파일 모두 0(대조 1).

### 14.2 lint (`lint-0.1.4.txt`, 판정 바이너리 `lint-binary-0.1.4.txt`)

```
/Users/goos/go/bin/moai spec lint SPEC-CON-AMEND-APPLY-001
INFO  OwnershipTransitionUnmeasured  …/spec.md  1  … commit 7b4d1ac89… has no Authored-By-Agent trailer
0 error(s), 0 warning(s)
LINT_EXIT=0
```

- 판정 바이너리: `v3.2.0-rc.7   moai_cp/20260910_130400-275-ged71054d3-dirty   built 2026-09-10T19:18:41Z`, `VERSION_EXIT=0`. 판정 트리 `51454e5e5`. 기준 커밋 이후 Go 코드 변경 0(§10.2), 이후 커밋은 SPEC·판정서 문서뿐. `-dirty` 내용은 미관측(Gap).

### 14.3 에이전트 보고 중 판정 출처 표시가 필요한 세부 (에이전트 판독)

design.md 에서 판정서 문구가 아니라 기존 SPEC 인코딩에서 온 것으로 표시된 세 가지 — 경로 구분자 경계(§11 두 번째 뮤턴트 `/root-evil` 차단 형태), 양쪽 경로 해석과 존재하는 가장 가까운 상위 경로 판정(REQ-CAA-021 · plan.md R-7), 소스 검사 순서(새 clause 0회 검사 먼저, plan.md M3).

### 14.4 범위 밖 쓰기 1건 (에이전트 자진 보고)

- manager-spec 이 워크트리 `.claude/agent-memory/manager-spec/` 에 `MEMORY.md` · `feedback_worktree_guard_plain_commands.md` 를 썼다. 지시(SPEC 폴더만 편집)를 벗어난 쓰기다.
- 그 경로는 `.gitignore` 로 무시되어 커밋·`git status` 에 나타나지 않는다. 처분은 리드 판단에 맡긴다.
- 레인 확인: 같은 파일이 primary 체크아웃의 공용 저장소에도 있다 — `/Users/goos/MoAI/moai-adk-go/.claude/agent-memory/manager-spec/feedback_worktree_guard_plain_commands.md`(1341 bytes, 11:22, 워크트리 사본과 크기·시각 동일). primary 의 `MEMORY.md`(27136 bytes)도 같은 11:22 에 수정됐다. 워크트리 에이전트 메모리를 primary 로 복사하는 쓰기 시점 미러가 동작한 것으로 보인다(판독 — 미러 로그는 확인하지 않았다). 여러 세션이 공유하는 저장소라 레인은 지우지 않았다.

## 15. plan-audit 1회차 — FAIL 0.78 (Tier L 합격선 0.85)

### 15.1 레인 확인 (`plan-audit-iter1.md`, 감사 HEAD `76144d40a`)

- 감사자가 쓴 파일은 보고서 하나뿐이었다(`git status` → `?? .moai/reports/t659/plan-audit-iter1.md`). 172줄, Cf 0.
- 보고서 4행 `Verdict: FAIL`, 5행 `Overall Score: 0.78`. 필수 항목 MP-1·2·3·5·6·7 PASS, MP-4 N/A. 범주 점수(감사자): 명확성 0.75 · 완결성 0.85 · 검증 가능성 0.70 · 추적성 0.80.
- 감사자는 `go test`·`go build` 를 돌리지 않았다. 코드 동작 주장은 모두 판독이다.

### 15.2 major 결함 4건 (감사자 판독)

- **D1** M-5b 는 원리상 죽일 수 없다 — AC-012 `third_rename_log_absent` 에서 세 번째 rename 이 위임 없이 실패하면 로그가 생기지 않아, 부재 파일을 지우지 않는 복원도 통과한다. 권고: 주입기를 "rename 에 위임한 뒤 오류 반환"으로.
- **D2** AC-024 CLI 사례가 CLI 의 검사 사용 여부를 가려내지 못한다 — `B/other` 가 `P` 등록부의 바이트 사본이라 검사 없는 CLI 도 `Execute` 거부로 같은 결과를 낸다. 권고: 사본 clause 를 다르게, 오류가 파이프라인에 닿지 않았음을 단언.
- **D3** 심볼릭 링크 해석을 로그 지점에서만 검증 — 등록부 경로·`file:` 지점에서 링크를 해석하지 않는 구현이 모든 AC 를 통과한다. 권고: AC-024 에 `symlinked_file`(가능하면 `symlinked_registry`) 행, M-23 을 지점별로.
- **D4** 공유 로더 `LoadRegistry` 변경의 영향 범위가 계획에 없다 — 리드 판단 필요(아래).

### 15.3 D4 근거 — 레인이 코드로 확인 (실행 안 함)

- `LoadRegistry(` 의 테스트 아닌 호출자 7곳(레인 grep): `internal/spec/lint.go:114`, `internal/cli/constitution.go:70`·`:160`·`:511`, `internal/cli/doctor.go:683`, `internal/constitution/validator.go:183`, `internal/constitution/pipeline.go:67`.
- `internal/spec/lint_test.go:14` `const testdataDir = "testdata"`, `:18-22` `testRegistryPath()` 는 상대 경로 `"../../.claude/rules/moai/core/zone-registry.md"` 를 반환. `:218-222` `TestLinter_AC08_DanglingRuleReference` 는 `RegistryPath: testRegistryPath()`, `BaseDir: testdataDir`.
- `internal/spec/lint.go:108-118` `NewLinter` 는 `projectDir := opts.BaseDir` 로 `LoadRegistry` 를 부르고, 적재 실패는 조용히 넘겨 `DanglingRuleReference` 검사를 건너뛴다.
- 판독: 현 `loader.go:82` 는 절대경로만 검사해 이 상대 경로를 받아들인다. §11 의 규칙(Clean + 절대화 + 루트 접두 비교)은 등록부를 저장소 루트의 `.claude/…` 로, 루트를 `internal/spec/testdata` 로 풀어 거부하게 된다 → 적재 오류가 삼켜짐 → finding 을 기대하는 AC08 테스트가 실패할 것으로 **예측**된다.
- 결정 필요: 이 거부를 §11 적용 범위의 의도된 결과로 볼지(호출자·테스트 갱신을 계획에 넣음), 경계 검사를 amend 경로에만 적용할지.

## 16. 2회차 수정(0.1.5) · D4 반영(0.1.6) 확인 · lint

### 16.1 0.1.5 — 한도로 끊긴 2회차의 완결 판단

- 2회차 manager-spec 이 세션 한도(`resets 1:40pm (Asia/Seoul)`, HTTP 429, request id `req_011Cevv1WBTDXx3PnKh5LeW2`)로 커밋 전에 중단했다. 약 7시간 뒤 리드 재개 지시로 레인이 미커밋분(acceptance·design·plan·progress·spec, 103+/79-)을 판독했다.
- 완결 판단 근거(레인 확인): version 0.1.5, REQ 21·AC 25·뮤턴트 29 가 HEAD `54ca2e3b6` 와 같음, Cf 0. D1 — AC-CAA-012 주입기 rename-then-fail 모드와 M-5b 변형별 kill map 문장이 끝까지 있음. D2 — AC-CAA-024 CLI 사례가 사본 clause 를 달리하고 `amendment failed`·`clause mismatch`·`Dry-run success` 부재를 단언, 이유 문장이 끝까지 있음. D3 — `symlinked_registry`·`symlinked_file` 행, M-23 지점별 변형. HISTORY 0.1.5 행이 D1–D3 과 D5–D16 을 각각 기록. plan.md R-8 diff 없음. 추가 108줄에 미완 표식(TODO·TBD·FIXME·XXX) 0.
- 레인이 카드 id 를 붙여 `193136a6a` 로 커밋했다.
- Gap: minor D5–D16 을 레인이 하나씩 대조하지는 않았다(HISTORY 기록에 의존). plan-audit 2회차의 확인 대상이다.

### 16.2 0.1.6 — D4 반영 (`564c370b5`, 새 manager-spec)

- 운영자 결정 D4: 경계 검사는 amend 경로(파이프라인 적용 단계·CLI `amend`)에만. `LoadRegistry` 자체와 amend 가 아닌 호출자는 현재 동작을 보존.
- **호출자 수 정정: 5곳이다.** §15.3 의 비테스트 호출 7곳 중 amend 경로는 `internal/cli/constitution.go:511`(`runConstitutionAmend`)·`internal/constitution/pipeline.go:67` 두 곳이고, 나머지는 `internal/spec/lint.go:114`·`internal/cli/constitution.go:70`·`:160`·`internal/cli/doctor.go:683`·`internal/constitution/validator.go:183` 다섯 곳이다. 리드 지시와 레인 보고의 "6호출자"는 틀린 수였다.
- 레인 확인: 커밋은 SPEC 파일 5개만 바꿈(79+/56-), 작업 트리 변경 0. version 0.1.6·tier L. 여섯 파일 전체 REQ 21·AC 25, 뮤턴트 29. 부속 4파일 `status:` 0. `plan.md:141` R-8 이 "The containment check is scoped to the amend path only (operator decision D4)." 로 시작. acceptance.md `loader_unchanged` 보존 사례(350행)와 `TestLinter_AC08_DanglingRuleReference` 보존 실행 존재.
- acceptance.md 등장 횟수(`grep -o -w`): AC-CAA-022 7 · AC-CAA-023 13 · AC-CAA-024 20 · AC-CAA-025 11 · M-19 8 · M-20 15 · M-24 11 · REQ-CAA-021 8. 미편집 대조: AC-CAA-001 3 · M-1 3. 에이전트가 보고한 편집 후 값과 모두 같다.
- 에이전트가 짚은 뮤턴트 조정(에이전트 판독): D4 로 로더의 절대경로 거부가 남으므로 M-20 (i) 는 AC-CAA-023 `divergent_root_*` 행을 더 못 죽인다 → 그 두 행의 담당을 M-19 로 옮김. M-24 의 AC-CAA-025 킬을 `load` 에서 `dry_run` 으로 옮김(`load` 는 `LoadRegistry` 직접 호출이라 containment 뮤턴트가 닿지 않음). 보존 행을 죽이는 뮤턴트는 새 변형 M-20 (iii)(검사를 `LoadRegistry` 안으로 옮김).
- Cf 문자 여섯 파일 모두 0(대조 1).

### 16.3 lint (`lint-0.1.6.txt`, 판정 바이너리 `lint-binary-0.1.6.txt`)

```
/Users/goos/go/bin/moai spec lint SPEC-CON-AMEND-APPLY-001
INFO  OwnershipTransitionUnmeasured  …/spec.md  1  … commit 7b4d1ac89… has no Authored-By-Agent trailer
0 error(s), 0 warning(s)
LINT_EXIT=0
```

- 판정 바이너리 `v3.2.0-rc.7   moai_cp/20260910_130400-275-ged71054d3-dirty   built 2026-09-10T19:18:41Z`, `VERSION_EXIT=0`. 판정 트리 `564c370b5`. 기준 커밋 이후 이 브랜치의 커밋은 SPEC·판정서 문서뿐(Go 코드 변경 0, §10.2). `-dirty` 내용은 미관측(Gap).

## 17. plan-audit 2회차 — FAIL 0.84 · 0.1.7 수정 · lint

### 17.1 2회차 판정 (`plan-audit-iter2.md`, 감사 HEAD `0085766cd`, 커밋 `1fe4a0289`)

- FAIL 0.84(Tier L 0.85). 1회차 D1–D16 전부 해결, MP 전부 통과. 새 결함 blocking N1(major, 숫자 단언이 임시 경로 숫자로 통과)·N2(darwin `/var` 링크)·N3(등록부 읽기 순서), optional N4–N6.

### 17.2 0.1.7 (`183bba916`, manager-spec)

- N1–N6 전부 반영(optional 포함 — 마지막 감사 전이라 한 번에). 새 REQ·AC 없음.
- 레인 직접 확인: `grep -cE '^- \*\*REQ-CAA-[0-9]+' spec.md` → 21, `grep -c '^### AC-CAA-' acceptance.md` → 25, `grep -c '^| M-' acceptance.md` → 29, `version: "0.1.7"`, 부속 파일 `^status:` 0, `\p{Cf}` 0건.
- 에이전트가 N4 명령에 `--first-parent` 를 더했다(흡수된 develop 개별 커밋 배제 목적, 에이전트 판독 — 실행 재현은 가드에 막혀 미관측). 3회차 감사가 판정할 항목.
- 에이전트 보고: `Authored-By-Agent` 트레일러가 마지막 줄 `🗿 MoAI` 때문에 트레일러 블록으로 인식되지 않음. 이 커밋은 상태 전이가 없어 lint 영향 없음. 전이 커밋(draft → in-progress)에서는 트레일러를 마지막 문단에 둘 것.

### 17.3 lint (`lint-0.1.7.txt`, 판정 바이너리 `lint-binary-0.1.7.txt`)

- `/Users/goos/go/bin/moai spec lint SPEC-CON-AMEND-APPLY-001` → `0 error(s), 0 warning(s)`, `LINT_EXIT=0`. 판정 트리 `183bba916`, 바이너리 `ed71054d3-dirty`(`-dirty` 내용 미관측, Gap).

## 18. run·sync 이후 — 리드 결정과 후속 후보

### 18.1 이 카드에서 처리한 것

- 운영자 결정(리드 전달, 2026-09-11): Kickoff 승인, 백업·임시 쓰기 실패 복원 Gap 공백 승인, X1–X4 재감사 없이 반영(`54b298476`).
- SPEC 밖 가드 G-A·G-B는 리드가 인정했다. 조건은 테스트 1개와 뮤턴트로 고정하는 것이다(`90ea2b26f`).
- G-B는 사람 승인 뒤에 거부되던 순서를 게이트 앞으로 옮겼다(`0893611ad` RED → `5801ebda0`).
- 컴파일 슬롯에서 CLI 칸을 재측정했다(`128a5ea52`). AC 25/25, 뮤턴트 42회 전부 잡힘.
- sync(`1f9946188`). sync-audit PASS 88.2를 리드가 수용했다(`e5a18feae`).
- 감사 후속
  - F5·F6: MX 태그 주석(`e8d16eaee`)
  - F1: G-B가 파일 동일성으로 비교(`2496053a8` RED → `bca8cf96a`)
  - F4: CHANGELOG 사실 오류 정정(`b36f50c4c`)
- 델타 감사 PASS 90.2(`cff2348a8`). 새 결함 D1·D2(Low, optional)는 처분을 리드에게 요청했다.

### 18.2 후속 카드 후보 (리드 결정 2026-09-12: 이 카드에서 구현하지 않고 기록만 한다)

| 후보 | 우선순위 | 내용 | 출처 |
|---|---|---|---|
| F2 | **High** (데이터 무결성) | 승인 뒤 레지스트리를 다시 읽을 때 대상 clause가 아직 `Before`인지 재확인하지 않는다. 승인을 기다리는 동안 들어온 다른 수정을 덮어쓸 수 있다 (`apply_transform.go:75-137`) | sync-audit F2 |
| F3 | 미정 (레인 제안 Medium) | 트리 안 심볼릭 링크 대상에 rename하면 링크가 일반 파일로 바뀐다. 현재 트리의 심볼릭 링크는 0개라 잠재 위험이다 (`apply_commit.go:127-131`) | sync-audit F3 |
| F7 | 미정 (레인 제안 Medium) | `.moai/research/`가 없고 락이 다른 곳에 있으면 첫 개정이 임시 쓰기에서 실패해, 검증되지 않은 복원 경로를 탄다. CLI 기본 경로는 해당하지 않는다 | sync-audit F7 |
| F8 | 미정 (레인 제안 Medium) | 공백만 있는 `After`, REQ-CAA-001/002/004/016 검증 실패가 사람 승인 뒤에야 드러난다. G-B를 옮긴 것과 같은 UX 결함이다 | sync-audit F8 |
| F9 | 미정 (레인 제안 Low) | 거부된 실제 모드 실행이 락 디렉터리를 남길 수 있다(기존 동작, SPEC §F 범위 밖) | sync-audit F9 |
| structure.md:78 | 미정 (레인 제안 Low) | `internal/constitution` "13 non-test files"는 이 카드 이전부터 틀렸다(BASELINE 14, HEAD 18) | manager-docs sync |
| MX WARN 3개 | 미정 (레인 제안 Low) | `LoadRegistry`(18), `rateLimiter.Admit`(17), `Validate`(22)는 복잡도 15 이상인데 WARN이 없다. PRESERVE 파일이라 손대지 않았다 | F6 작업 중 측정 |
| ANCHOR 강등 | 미정 (레인 제안 Low) | `LoadEvolutionLogs` ANCHOR의 프로덕션 fan_in이 2다. NOTE로 강등할지는 리드가 결정한다 | F5 작업 중 측정 |
| 비-amend 호출자 격리 | 미정 | D4로 제외한 `LoadRegistry` 비-amend 호출자 5곳의 containment check 확장 | §16 D4 (리드 기록) |
