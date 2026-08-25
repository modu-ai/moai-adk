# acceptance — SPEC-ZONE-REGISTRY-HARDEN-001

> 판정 트리: 본 워크트리(HEAD = M3 완료 시점). 기준 baseline `db1362739`. 모든 AC는 명령 + 기대 출력으로 이진 판정 가능해야 한다. `SENT` 조건에서 측정하지 않은 항목은 Gaps로 보고한다(pass 아님).

## §D AC Matrix

| AC | 축 | 대응 REQ | 판정 기준(요약) |
|----|----|----------|------------------|
| AC-ZRH-001 | F1 clause | REQ-ZRH-001/002 | 신규 clause가 양쪽 미러에 1회, 핀된 원본에서 단일 행 유일 적중 1회 |
| AC-ZRH-002 | F1 rewrap 서식 전용 | REQ-ZRH-003 | 공백 정규화 후 ci-autofix-protocol.md(트윈 각각)가 base `db1362739` 대비 바이트 동일 |
| AC-ZRH-003 | 트윈·미러 동일 | REQ-ZRH-003/008 | 트윈 쌍·미러 쌍 `cmp` 바이트 동일 |
| AC-ZRH-004 | F2 가드 초록 | REQ-ZRH-004/008 | 가드 suite 초록 + mirror별 digest 단정 통과 로그 관측 |
| AC-ZRH-005 | F2 치환 차단 | REQ-ZRH-005 | 개수 보존 치환 4종(ID/zone/zone_class/canary_gate)이 digest 검사에서 실패 |
| AC-ZRH-006 | validator·CLI 축 | REQ-ZRH-008 | `make build` rc=0, `bin/moai constitution validate` exit=0(드리프트 0), 임베드 no-op |
| AC-ZRH-007 | F3 문서 정렬 | REQ-ZRH-007 | 발생 의미론 서실 제거 + 라인 수 의미론 기술 + 정정 출처 주석 존재 |
| AC-ZRH-008 | 품질 baseline | REQ-ZRH-008 | lint 0 issues·gofmt 빈·coverage ≥85%·windows vet rc=0 |
| AC-ZRH-009 | F2 실패 메시지 | REQ-ZRH-006 | digest 실패 메시지가 계산 digest + 같은-커밋 갱신 절차 문구를 문자 그대로 포함 |

### AC-ZRH-001 — clause 재선택 (단일 행 완결 문장)

- **Given** M1 완료 후 트리
- **When** 다음 3명령 실행:
  ```bash
  grep -c 'clause: "The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report."' .claude/rules/moai/core/zone-registry.md
  grep -c 'clause: "The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report."' internal/template/templates/.claude/rules/moai/core/zone-registry.md
  grep -c -F 'The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.' .claude/rules/moai/workflow/ci-autofix-protocol.md
  ```
- **Then** 세 출력 모두 정확히 `1` (레지스트리에는 1회 기입, 원본에는 단일 행으로 유일 적중 — 절단 clause `"…failure) MUST"` 는 더 이상 clause 값으로 존재하지 않음: `grep -c 'test assertion failure) MUST"' .claude/rules/moai/core/zone-registry.md` → `0`)

### AC-ZRH-002 — rewrap은 서식 전용 (단어 불변)

- **Given** M1 완료 후 트리와 baseline `db1362739`
- **When** 배포판 트윈에 대해(템플릿 원본도 동일 절차 반복):
  ```bash
  git show db1362739:.claude/rules/moai/workflow/ci-autofix-protocol.md | tr '\n' ' ' | tr -s ' ' > /tmp/zrh-old.n
  tr '\n' ' ' < .claude/rules/moai/workflow/ci-autofix-protocol.md | tr -s ' ' > /tmp/zrh-new.n
  cmp /tmp/zrh-old.n /tmp/zrh-new.n ; echo rc=$?
  ```
- **Then** `rc=0` (공백 정규화 텍스트 바이트 동일 — 개행 위치만 이동했고 어떤 단어도 바뀌지 않음)

### AC-ZRH-003 — 트윈·미러 바이트 동일

- **Given** M1 완료 후 트리
- **When**:
  ```bash
  cmp .claude/rules/moai/workflow/ci-autofix-protocol.md internal/template/templates/.claude/rules/moai/workflow/ci-autofix-protocol.md
  cmp .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md
  go test -run TestRegistrySyncMirrorsIdentical ./internal/constitution/
  ```
- **Then** `cmp` 2건 무출력(동일) + parity 테스트 `ok`

### AC-ZRH-004 — 가드 suite 초록 (digest 단정 포함)

- **Given** M1+M2 완료 후 트리
- **When**:
  ```bash
  go test ./internal/constitution/
  go test -run TestRegistrySyncGuard -v ./internal/constitution/ 2>&1 | grep -E 'tuple.?digest|literal buckets|evaluated'
  ```
- **Then** 패키지 `ok` + mirror별(`local`/`template`) 버킷 라인이 종전 규격 그대로: `clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries`, `once=97 zero=0 multi=0 retired_exempt=4 self_reference=0` (REQ-ZRH-008 비회귀) — digest 단정은 통과해야 하므로 실패 라인 없음

### AC-ZRH-005 — 개수 보존 치환 차단 (F2 핵심)

- **Given** M2 완료 후 트리
- **When**:
  ```bash
  go test -run TestRegistryTupleDigestRejectsSubstitution -v ./internal/constitution/
  ```
- **Then** `PASS` — 서브테스트가 개수 보존 변이 4종(ID 문자 삽입 치환 / `zone:` 값 치환 / `zone_class:` 값 치환 / `canary_gate:` 반전) 각각에 대해 산출 digest ≠ `wantTupleDigest`를 단정하고 전부 통과. RED 근거(구현 전 동일 변이가 가드를 통과함을 보인 출력)는 run-phase §E.2에 원문 보존

### AC-ZRH-006 — validator·CLI·임베드 축

- **Given** M1+M2 완료 후 트리
- **When**:
  ```bash
  make build && ./bin/moai constitution validate ; echo exit=$?
  git status --porcelain | grep -v '^??' | wc -l
  ```
- **Then** `exit=0`(드리프트 0) + 트래킹파일 변경은 본 SPEC 6개 대상 파일(+SPEC 산출물)뿐 — 임베드 재빌드가 추적 파일을 추가 오염하지 않음

### AC-ZRH-007 — plan.md 의미론 정렬 (F3)

- **Given** M3 완료 후 트리
- **When**:
  ```bash
  grep -c 'strings.Count(rawFileContent, clause)' .moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md
  grep -c -i '라인 수 의미론' .moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md
  grep -c 'SPEC-ZONE-REGISTRY-HARDEN-001' .moai/specs/SPEC-ZONE-REGISTRY-RESYNC-001/plan.md
  ```
- **Then** 첫 명령 `0`(발생 의미론 서술 제거), 둘째 `≥1`(실측 의미론 기술), 셋째 `≥1`(정정 출처 erratum 주석)

### AC-ZRH-008 — 품질 baseline

- **Given** 전 마일스톤 완료 후 트리
- **When**:
  ```bash
  golangci-lint run ./internal/constitution/...
  gofmt -l internal/constitution/
  go test -cover ./internal/constitution/
  GOOS=windows GOARCH=amd64 go vet ./internal/constitution/ ; echo rc=$?
  ```
- **Then** lint `0 issues` · gofmt 무출력 · `coverage: 8X.X%`(≥85) · `rc=0`

### AC-ZRH-009 — digest 실패 메시지 내용 (REQ-ZRH-006)

- **Given** M2 완료 후 트리
- **When**:
  ```bash
  grep -c -F 'update wantRegistryEntries and wantTupleDigest in the same change' internal/constitution/registry_sync_test.go
  ```
  그리고 스크래치 변이(로컬에서 `wantTupleDigest` 상수 1바이트 변경 — 미커밋)로 `go test -run TestRegistrySyncGuard -v ./internal/constitution/` 를 돌려 digest 실패 메시지 원문을 캡처한 뒤 원복
- **Then** grep → `≥1` (위 영문 문구는 본 AC 가 정식 토큰으로 규정하는 갱신 절차 — 구현 메시지에 문자 그대로 포함되어야 함), 스크래치 실행의 실패 메시지 원문이 **계산된 digest(hex 문자열)** 를 포함하고 해당 원문이 §E.2에 보존됨. 원복 후 해당 파일은 `git status --porcelain` clean

## §D.1 Edge Cases

- clause 문장이 원본에서 2회 이상 행 적중하는 경우 → AC-ZRH-001 셋째 명령이 `2+`로 FAIL (hitMulti 버킷도 가드에서 붉게 잡음 — 이중 방어)
- rewrap 후 validator 공백 정규화 containment 파괴(공백 삽입·삭제 동반) → AC-ZRH-006 validate exit≠0으로 포착
- digest 직렬화의 정렬 불안정(같은 입력 두 연산 상이) → sort 후 join이므로 발생 불가; 만약 관측되면 helper 버그로 FAIL
- 개수 보존 변이가 YAML 파싱까지 깨는 경우(ID에 따옴표/콜론 삽입) → LoadRegistry 에러 경로에서 테스트 실패 — 이것도 차단(파서 오류 역시 가드 영역)

## §D.2 Quality Gate / Definition of Done

- AC-ZRH-001..009 전부 이진 PASS, 각 판정에 명령+출력 원문 근거(VCI 5분할 보고)
- CI(main PR) 전 매트릭스 초록 — 로컬 전체 스위트는 돌리지 않는다(§4 로컬 규율: affected-package → push → CI 판독)
- RESYNC-001이 남긴 계약 전부 비회귀(REQ-ZRH-008): bucket 97/0/0·은퇴 4·anchor 101·미러 동일·validator.go 무편집
- sync-phase(manager-docs): CHANGELOG 1항 + progress.md §E.4 + 3-phase close
