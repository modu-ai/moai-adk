# t566 판정서 — codemaps fold 단위를 구속하는 검사 부재: 재현, 그리고 수리 위치는 설계 판단

- 카드: t566 (재현 먼저)
- 트리: `.claude/worktrees/t566`, 브랜치 `WT-fold-unit-guard`, HEAD `cc5006429`(두 번째 부모 = 로컬 develop `92bf71523`)
- 측정: 2026-09-10, lane-6. 모든 측정은 `.moai/reports/t566/repro/` 에 있다.

## 1. 주장 (Claim)

1. **재현됐다.** fold 로 판정된 단위에 codemaps 산문을 넣는 뮤턴트를 적용하면, fold 불변식("판정 뒤에도 히트 0")은 깨진다. 그런데 SPEC-CODEMAPS-REFRESH-002 의 검사 가운데 기계로 다시 돌릴 수 있는 것은 전부 뮤턴트 전과 같은 결과를 낸다. AC-CM2-004(omission 편입), AC-CM2-007 의 omission 규칙, 게이트(`moai graph check`)의 codemaps·citations 계층이 모두 그렇다.
2. **게이트도 이 위반을 보지 못한다.** `moai graph check` 의 codemaps 계층 값은 뮤턴트 전후 모두 `47`(임계 40)이다. 이 계층은 산문 내용이 아니라 스탬프 이후의 트리 변경을 세기 때문이다.
3. **수리 위치는 설계 판단이다.** SPEC-CODEMAPS-REFRESH-002 는 `status: completed` 라 그 AC 는 다시 평가되지 않는다. 결함이 문제가 되는 곳은 **다음** codemaps 갱신이다. 그런데 그 갱신이 어떤 산출물(새 SPEC, `/moai codemaps` 워크플로, 게이트 계층)의 규칙을 따를지는 이 카드가 정할 수 없다. 카드 지시("처방은 카드에서 판단")가 있지만, 셋 중 무엇을 고르느냐에 따라 산출물이 달라지므로 여기서 멈추고 보고한다.
4. **곁가지 두 층을 확인했다.** (1) AC 축의 id 정규식은 접미 문자 id 를 흡수한다(12 대 13). 이는 결함으로 재현된다. (2) REQ 추출 축에서는 접미 문자 id 정의 줄이 추출되지 않는다. 다만 이것은 코드 주석에 **의도된 잔여 형태**로 적혀 있다. 카드가 결함으로 부른 전제는 이 문서화 사실과 함께 판단돼야 한다.

## 2. 증거 (Evidence)

### 2.1 fold 뮤턴트 재현

집계 규약은 SPEC 자신의 것을 그대로 썼다. `census.py` 가 적용한 규약은 셋이다. 단위별 적중 행 수는 §A.3(a) 의 `grep -c -F` 이다. 파일 적중은 AC-CM2-004 의 `grep -rl -F` 이다. 적중 0 패키지 목록은 §A.3(a) 의 `go list` 모집합이다. fold 5·omission 15 단위는 `.moai/reports/t475/verdict.md` 관측 ② 의 최종 판정에서 가져왔다.

| 측정 | 뮤턴트 전 (`census-baseline`) | 뮤턴트 후 (`census-mutant`) |
|---|---|---|
| fold 단위 중 적중 0 | 5 / 5 | **3 / 5** (`internal/core/git` · `internal/kanban/prlink_landedref.go` 가 적중) |
| omission 단위 중 파일 적중 없음 (AC-CM2-004 FAIL 조건) | 0 | 0 → **AC-CM2-004 통과 조건 유지** |
| 적중 0 패키지 목록에 남은 omission 단위 (AC-CM2-007 FAIL 조건) | 0 | 0 → **AC-CM2-007 통과 조건 유지** |
| 적중 0 패키지 목록에 남은 fold 패키지 | `internal/core/git` | 없음 — 목록에서 사라졌지만 이를 요구하는 검사가 없다 |
| 적중 0 패키지 수 | 38 | 37 |
| `go list` 패키지 수 / exit | 137 / 0 | 137 / 0 |

- 뮤턴트: `mutant_apply.py` 가 `modules.md` 끝에 fold 단위 둘을 기술하는 문장 하나를 붙였다. 패키지 입도 하나(`internal/core/git`), 파일 입도 하나(`internal/kanban/prlink_landedref.go`)다. t475 첫 통과가 저지르고 되돌린 위반과 같은 모양이다.
- 게이트 (`go run ./cmd/moai graph check --json`, 이 트리에서 빌드):

| 계층 | 뮤턴트 전 | 뮤턴트 후 |
|---|---|---|
| codemaps | stale, 47 / 40 | stale, 47 / 40 |
| citations | fresh, 0 / 0 | fresh, 0 / 0 |
| mx-index · edges | absent (새 워크트리 예상 상태) | absent |
| exit | 1 | 1 |

  두 실행의 계층 값이 같다. codemaps 계층이 뮤턴트 전부터 stale 인 것은 t475 스탬프 이후 develop 변경 때문이며, 이 카드와 무관하다.
- 원복: `mutant_revert.py` 로 저장해 둔 원본을 되돌렸다. 원본 `d57e59f5…7c97`, 뮤턴트 `db1cb4ca…4d2f`, 복원 `d57e59f5…7c97`, `RESTORED_MATCHES_ORIGINAL True`(`mutant-revert.txt`). 원복 뒤 `git status` 에는 `.moai/reports/t566/` 만 새로 보인다.
- 기계로 다시 돌리지 않은 검사: AC-CM2-002(판정 행 수 = 후보 수)는 t475 증거 파일을 세는 명령이라 codemaps 산문과 무관하다. AC-CM2-006·008 은 증거 파일의 표를 대상으로 한다.

### 2.2 곁가지 (1) — AC id 정규식

`ac-id-regex-count.txt`: SPEC-CODEMAPS-REFRESH-002 `acceptance.md` 에서 서로 다른 id 의 수를 셌다. canonical `AC-([A-Z0-9]+-)*[0-9]+` → **12**, 접미 허용 `AC-([A-Z0-9]+-)*[0-9]+[a-z]?` → **13**. 차이는 `AC-CM2-003a` 하나다. canonical 형태는 `.claude/agents/moai/plan-auditor.md` Group A 검사의 세 번째 `Grep` 줄에 있다(develop 판독에서 227행. sync-audit 는 t475 시점의 228행으로 인용했다). sync-audit 가 적은 배포 템플릿·codex 방출본 사본도 같은 줄로 있다고 기록돼 있지만, 두 사본은 이번에 다시 읽지 않았다.

### 2.3 곁가지 (2) — REQ 추출

`req-extract-suffix.txt`: `internal/spec/lint_req_widen.go` 의 `reqLineWidePattern` 을 Python `re` 로 옮겨 적용했다(RE2 전용 문법 없음).

- 대조: `spec.md:305` 의 실제 `REQ-CM2-014` 정의 줄 → MATCH, 추출 id `REQ-CM2-014`.
- 같은 줄에서 id 만 `REQ-CM2-003a` 로 바꾸면 → **NO MATCH.**
- 같은 파일의 주석은 이 형태를 이렇게 적는다: "Known residual forms, deliberately OUT of scope for this widening (a line-based parser must not chase them): lowercase-token suffixes (`REQ-BDR-005b`) …".
- 첫 시도는 `REQ-CM2-014` 가 **언급된** 다른 줄을 대조로 골라 추출 id 가 `REQ-CM2-003` 으로 나왔다. 대조가 틀렸으므로 버리고, 정의 줄만 잡는 선택식으로 다시 쟀다(위 결과). 파일에는 두 번째 결과만 남아 있다.

## 3. 기준 귀속 (Baseline-attribution)

- 모든 측정은 이 워크트리에서, 로컬 develop `92bf71523` 을 병합한 HEAD `cc5006429` 위에서 했다. 게이트는 `go run ./cmd/moai` 로 이 트리에서 빌드했고, 설치된 바이너리는 쓰지 않았다.
- codemaps 파일은 뮤턴트 전후와 원복 뒤 해시로 추적했다(`modules-sha-before.txt`, `mutant-apply.txt`, `mutant-revert.txt`).

## 4. 전제 정정

- 카드는 근거를 "sync-audit **F2**"로 적었다. 그러나 t475 `sync-audit.md` 에서 fold 간극을 다룬 절은 **Flag 2**(142행)다. 같은 파일의 F2 는 AC-CM2-009 regression-guard 처분에 관한 다른 finding 이다(230행). 내용 인용은 맞고 절 이름만 다르다.
- 카드의 곁가지 (2)는 REQ 추출 실패를 "더 나쁜 쪽" 결함으로 부른다. 측정은 카드와 같지만(NO MATCH), 코드는 그 형태를 의도적으로 범위 밖에 두었다고 문서화한다. 이것을 결함으로 고칠지, 문서화된 한계를 유지하고 SPEC 작성 규칙에서 접미 문자 REQ id 를 막을지는 판단이 필요하다.

## 5. 미검증 (Gaps)

- 수리 방향을 고르지 않았으므로 새 검사를 만들지도, 그 검사의 RED/GREEN 을 관측하지도 않았다.
- AC id 정규식의 템플릿·codex 사본 두 곳은 이번 실행에서 다시 읽지 않았다.
- REQ 추출 실패가 lint 출력에서 실제로 조용한지(어떤 finding 도 나오지 않는지)는 `moai spec lint` 를 접미 id SPEC 픽스처로 돌려 확인하지 않았다. 확인한 것은 추출 정규식 한 줄의 매치 여부다.
- 뮤턴트는 파일 하나(`modules.md`)에 한 문장을 붙인 형태 하나만 썼다. 다른 문서나 입도(예: 부모 산문을 고치는 형태)는 재지 않았다.

## 6. 잔여 위험 — 리드가 정할 설계 판단

fold 불변식을 어디에 둘지 선택지는 셋이다. 이 카드는 고르지 않는다.

1. **다음 codemaps 갱신 SPEC 의 AC 로 둔다.** 저비용이지만, 다음 SPEC 작성자가 이 AC 를 기억해야 한다. t475 에서 불변식이 실행자 머릿속에만 있었던 것과 같은 구조가 한 층 위로 옮겨질 뿐이다.
2. **`/moai codemaps` 워크플로(스킬)의 검증 단계로 둔다.** 갱신 절차에 붙어 매번 돈다. 다만 fold 판정 목록이 어디에 기록되는지(판정 결과의 저장 위치)가 먼저 정해져야 기계 검사가 가능하다.
3. **`moai graph check` 에 계층을 추가한다.** 가장 강하지만, 판정 결과를 읽을 입력 형식이 필요하고 Go 코드 변경이다.

어느 쪽이든 카드가 요구한 **입도 결정**이 따라온다. 패키지 입도 fold 는 적중 0 패키지 목록에 나타나지만, 파일 입도 fold 4개는 그 목록에 원래 나타나지 않는다(이번 측정에서도 목록에 남은 fold 는 `internal/core/git` 하나였다). 따라서 검사는 목록이 아니라 판정 단위 전수에 대해 `grep -c -F` 적중 0 을 확인하는 형태여야 한다.

## 7. 리드 결정 이후 — 구현과 검증

- 트리: 같은 워크트리·브랜치. 커밋 순서는 RED `2131a21d5` → GREEN `3e6f2484a` → 뮤턴트 증거 `69cf6746c` → 판정 파일·실제 트리 증거·이 절(이 절을 담는 커밋).
- 모든 실행 출력과 종료 코드는 `.moai/reports/t566/` 와 `.moai/reports/t566/verb/` 에 있다.

### 7.1 리드 결정

- **② 를 골랐다.** fold 불변식은 `/moai codemaps` 워크플로의 검증 단계(Phase 4)에 실행 가능한 동사로 둔다.
- **① 은 기각했다.** 다음 codemaps 갱신 SPEC 의 AC 로 두면 불변식이 여전히 그 SPEC 작성자의 기억에 머문다. t475 에서 실행자 머릿속에 있던 구조가 한 층 위로 옮겨질 뿐이다.
- **③ 은 후속 후보로만 기록한다.** `moai graph check` 계층은 가장 강하지만 Go 코드 변경과 입력 형식 설계가 따라온다. 이번 카드에서는 만들지 않았다.
- 입도는 §6 의 결론을 따랐다. 검사는 적중 0 패키지 목록과 대조하지 않고, 기록된 판정 단위를 하나씩 센다. 파일 입도 단위는 패키지 목록에 원래 나타나지 않기 때문이다.

### 7.2 판정 파일을 생성기가 덮어쓰지 않는가

판정은 `.moai/project/codemaps/fold-judgments.txt` 에 둔다. 이 파일이 갱신 과정에서 사라지지 않는다는 근거는 셋이다.

1. `internal/cli/graph_stamp.go` 는 `provenance.json` 만 쓴다(임시 파일 뒤 rename). 이 사실은 lane-6 이 확인했다. 이번 실행에서는 같은 파일의 codemaps 디렉터리 상수(21행)와 스탬프 명령 설명(56행)만 다시 읽었고, 쓰기 경로 자체는 다시 읽지 않았다.
2. 워크플로 Phase 3 는 이름이 정해진 `.md` 문서 다섯 개를 쓴다. 이번에 추가한 문장도 생성 단계가 이 파일을 다시 쓰지 않는다고 명시한다.
3. citations 계층(`internal/graph/check_citations.go`)은 디렉터리를 읽은 뒤(57행) `.md` 접미어가 아닌 항목을 건너뛴다(67행). 이번 실행에서 두 행을 직접 읽었다.

같은 이유로 판정 파일은 `.txt` 다. 동사가 스캔하는 집합은 디렉터리 최상위의 `*.md` 뿐이라, 판정 파일이 그 안에 들어가 자기 단위를 모두 적중시키는 일이 생기지 않는다.

### 7.3 만든 것

1. **워크플로 문안** — 템플릿 사본을 먼저 고치고 같은 바이트를 로컬 사본에 썼다. 두 사본의 sha256 은 모두 `17c53295…a178c` 다(`tools/patch_workflow.py` 출력).
   - Phase 3: 판정 기록 형식(`fold <unit>` / `omission <unit>`, 빈 줄과 `#` 주석 무시)과 `.txt` 인 이유.
   - Phase 4: `### Fold and Omission Judgment Check` 제목 아래 bash 블록 하나와 읽는 법. 출력은 `COLLECTED: fold=N omission=M` 한 줄과 표지 줄 `UNCOVERED:` · `FOLD-PROSE: <unit> <행 수>` · `GAP:` 이다. PASS 를 출력하지 않고 항상 종료 코드 0 으로 끝나며, 표지 줄이 없을 때 통과다. 판정 파일이 읽히지 않거나, 유효한 판정이 0 개이거나, 형식에 맞지 않는 줄이 있거나, 생성 문서가 하나도 없으면 `GAP:` 을 낸다.
   - 템플릿 문안에는 특정 프로그래밍 언어의 경로·예시, SPEC ID, 카드 ID, 날짜, 커밋 SHA 를 넣지 않았다. 단위는 `<unit>` 자리표시자로만 적었다.
2. **Go 테스트** — `internal/cli/codemaps_fold_judgments_test.go`. 동사는 템플릿 사본에서 매 실행 추출한다(기존 `auditVerb` 재사용). 하위 테스트 일곱 개와 사본 동일성 테스트 하나다: `green_normal_state`, `red_fold_prose`, `red_uncovered_omission`, `gap_missing_judgments`, `gap_empty_judgments`, `gap_malformed_line`, `gap_no_documents`, `TestCodemapsFoldJudgments_VerbIdenticalAcrossCopies`. GREEN 픽스처는 판정 파일을 스캔 대상과 같은 디렉터리에 두므로, 이 셀의 FOLD-PROSE 0 줄이 판정 파일의 자기 적중이 없다는 증거를 겸한다(테스트 주석에 적었다). 표지 줄은 비어 있지 않은지가 아니라 정확한 줄 목록으로 비교한다. bash 가 없으면 건너뛴다.
3. **이 저장소의 판정 파일** — t475 판정서 관측 ② 의 판정 20 개(fold 5, omission 15).

### 7.4 RED · GREEN · 뮤턴트

| 단계 | 커밋 | 관측 | 증거 |
|---|---|---|---|
| RED | `2131a21d5` | 여덟 셀 모두 FAIL. 원인은 전부 `no bash block under "### Fold and Omission Judgment Check"` — 동사가 아직 없어서 실패했다. `EXIT=1` | `red-test.txt` |
| GREEN | `3e6f2484a` | 여덟 셀 모두 PASS, `EXIT=0` | `green-test.txt` |
| 뮤턴트 | `69cf6746c` | 템플릿 사본의 fold 분기를 `hits[i] > 0` → `hits[i] < 0` 으로 꺼서 다시 돌렸다. `red_fold_prose` 가 FAIL(기대 `[FOLD-PROSE: src/web/generated_view.ext 1]`, 실제 `[]`, 출력 `COLLECTED: fold=2 omission=2`), 사본 동일성 셀도 FAIL. 나머지 여섯 셀은 PASS. `EXIT=1` | `verb/verb-mutant-test.txt` |
| 원복 | — | 원본 `17c53295…`, 뮤턴트 `64beee37…`, 복원 `17c53295…`, `RESTORED_MATCHES_ORIGINAL True`. 해당 파일의 `git diff --stat` 출력 없음(`EXIT=0`). `git hash-object` 와 `HEAD:` 블롭이 모두 `bef5a2cfcb20312c3213213b56b69aad531f44d7`. 원복 뒤 여덟 셀 PASS | `verb/verb-mutant-revert.txt`, `verb/verb-mutant-restore-diffstat.txt`, `verb/verb-mutant-restore-blob.txt`, `verb/post-restore-test.txt` |

### 7.5 실제 트리 실행

동사를 템플릿 사본에서 추출해 이 워크트리 루트에서 bash 로 돌렸다(`tools/run_verb.py`).

| 실행 | 출력 | 원복 |
|---|---|---|
| 정상 상태 | `COLLECTED: fold=5 omission=15`, 표지 줄 없음, `EXIT=0` | — |
| fold 뮤턴트(§2.1 의 `mutant_apply.py`) | `FOLD-PROSE: internal/core/git 1`, `FOLD-PROSE: internal/kanban/prlink_landedref.go 1` — 뮤턴트가 건드린 두 단위 정확히 | `mutant_revert.py`: 원본·복원 `d57e59f5…7c97`, `RESTORED_MATCHES_ORIGINAL True` |
| omission 제거 뮤턴트(`tools/omission_mutant.py`) | `UNCOVERED: internal/settings/yamlpatch` 한 줄 | 세 문서 모두 복원 해시가 원본과 일치, `RESTORED_MATCHES_ORIGINAL True` |

- omission 뮤턴트는 판정 파일 순서대로 보며, 적중 줄에 다른 판정 단위가 함께 적히지 않은 첫 omission 단위를 고른다. 그 규칙으로 `internal/settings/yamlpatch` 가 뽑혔고, 이 단위를 적은 줄을 모두 지웠다: `data-flow.md` 1 줄, `dependencies.md` 2 줄, `modules.md` 1 줄. 앞 순서의 단위들이 왜 걸러졌는지는 따로 기록하지 않았다.
- 증거 파일: `verb/real-green.txt`, `verb/real-fold-mutant.txt`, `verb/fold-mutant-apply.txt`, `verb/fold-mutant-revert.txt`, `verb/real-omission-mutant.txt`, `verb/omission-mutant-apply.txt`, `verb/omission-mutant-revert.txt`.
- 두 원복 뒤 GREEN 커밋 직전의 `git status` 에는 codemaps `.md` 수정이 없었다. 뮤턴트가 커밋에 들어간 적은 없다.

### 7.6 카탈로그 해시와 방출 검사

- `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai`: `moai` 항목 해시 `fa683eb7fdb42944cc80a6461415cb371edb277a90c13e04e0ba079a68aa22b6` → `06701a7eefe06e66bd51ddfb2dc7055d119f761ab120c71f4fa7128343efb1b4`. `catalog.yaml` 의 diff 는 이 한 줄뿐이다(`catalog-hash-regen.txt`).
- `make commands-emit-check` `EXIT=0`, `make agents-emit-check` `EXIT=0`(`commands-emit-check.txt`, `agents-emit-check.txt`).
- `go test ./internal/template/ -run 'TestManifestHashFormat|TestCatalog|TestAllSkillsInCatalog'`: 여덟 테스트 PASS, `EXIT=0`. 알려진 plan-auditor `CATALOG_HASH_UNSTABLE` 실패는 이번 실행에 나타나지 않았다(`template-tests.txt`).
- `go vet ./internal/cli/` `EXIT=0`, `gofmt -l` 출력 없음(`vet.txt`), `golangci-lint run ./internal/cli/` `0 issues.`(`lint.txt`).

### 7.7 곁가지

1. **AC id 정규식**(§2.2): 기록만 한다. 이번 카드에서 바꾸지 않았다.
2. **접미 문자 REQ id**(§2.3): 코드 주석에 문서화된 의도된 한계로 유지한다. SPEC 작성 규칙에서 접미 문자 REQ id 를 막는 방안은 후속 후보다.

### 7.8 미검증 (Gaps)

- 동사는 darwin 의 bash·awk 에서만 돌렸다. Linux 의 gawk·mawk 에서의 동작은 CI 에서 처음 관측된다. bash 가 없는 Windows 실행기에서는 테스트가 건너뛰므로 그 환경에서 이 검사는 관측되지 않는다.
- `graph_stamp.go` 가 임시 파일 뒤 rename 으로 `provenance.json` 만 쓴다는 사실은 lane-6 의 확인에 기댄다(§7.2).
- 동사 뮤턴트는 fold 분기만 껐다. omission 분기를 끈 뮤턴트는 돌리지 않았다. omission 쪽은 `red_uncovered_omission` 셀과 실제 트리의 omission 제거 뮤턴트가 입력 쪽에서 확인했을 뿐이다.
- 판정 20 개가 현재 트리 기준으로도 옳은 판정인지는 다시 판단하지 않았다. 동사가 보는 것은 기록된 판정과 산문의 일관성이다.
- `--area` 로 만든 하위 디렉터리 지도는 스캔 범위 밖이다(문안에 명시). 영역별 판정을 검사하는 장치는 없다.
- RED 단계의 실패 메시지 앞머리는 재사용한 `auditVerb` 의 표지 `PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT` 로 찍힌다. 메시지 본문에 문서 경로와 제목이 함께 찍혀 원인은 읽히지만, 표지 이름은 이 검사와 맞지 않는다.

### 7.9 잔여 위험 (Residual-risk)

- **부분 문자열 일치.** 적중은 고정 문자열 포함 여부다. fold 단위가 서술된 더 긴 경로의 앞부분이면 산문이 없어도 `FOLD-PROSE` 가 난다(문안에 명시). 반대로 omission 단위가 다른 서술된 경로의 일부이면, 그 단위 자체의 설명이 없어도 적중으로 세어져 `UNCOVERED` 가 나지 않는다. 뒤쪽은 문안에 적지 않았다.
- **실행 강제력이 없다.** 이 검사는 오케스트레이터가 Phase 4 를 실제로 돌릴 때만 작동한다. 갱신이 이 단계를 건너뛰면 아무것도 막지 않는다. 강제는 ③ 게이트 계층의 몫이다.
- **판정 기록 누락.** 갱신이 판정을 하고도 파일에 적지 않으면 동사는 적힌 것만 본다. `COLLECTED` 수를 갱신 중 내린 판정 수와 비교하라고 문안에 적었지만, 그 비교는 사람의 몫이다.

### 7.10 공정 기록 — 승인 없는 `internal/cli` 컴파일 (lane-6)

- 테스트는 `internal/cli/codemaps_fold_judgments_test.go`(package `cli_test`)에 들어갔다. t524 의 `auditVerb` 헬퍼를 재사용하기 위해서다. 그런데 이 패키지의 컴파일은 리드 슬롯 승인 대상이다. 위 RED·GREEN 실행(`red-test.txt`, `green-test.txt`)과 `go vet` · `golangci-lint`(`vet.txt`, `lint.txt`)는 **승인 없이** 돌았다.
- 원인은 lane-6 의 첫 지시가 테스트 위치를 "t524 패턴과 일관된 곳"으로 열어 둔 것이다. "`internal/template` 에 두고 `internal/cli` 는 컴파일하지 말라"는 교정은 워커가 그 단계를 지난 뒤 수신함에 닿았다.
- **리드 판정: 이 실행들은 판정 근거로 쓰지 않는다.** 경합 조건이 기록되지 않았기 때문이며, GREEN 은 조기 신호로만 읽는다. 위치는 `internal/cli` 에 그대로 두기로 했다(`auditVerb` 추출 로직을 둘로 만들지 않기 위해).
- 정식 근거는 병합 창에서 리드가 승인한 슬롯으로 `go test ./internal/cli/ -run 'TestCodemapsFoldJudgments'` 를 한 번 돌려 얻는다. 이름별 PASS 와 두 사본의 동사 동일성 테스트를 함께 확인하고, 그 결과를 병합 트리 재측정 절에 적는다.
- 워커는 사용량 한도로 멈췄다. lane-6 은 메시지 없이 종료했고, 스테이징만 돼 있던 판정 파일·실제 트리 실행·뮤턴트 증거·이 판정서 §7 을 점검한 뒤 `242f2d8b7` 로 보존 커밋했다.
