# SPEC-CODEX-AUDIT-GATE-AXES-001 — 인수 기준

카드 **t686** · v0.2.0. 각 AC 는 명령과 관측 출력으로 판정한다. 축 (b) AC 는 §B.4 권장 표면 S1 기준으로 작성했다. Kickoff 에서 다른 표면이 선택되면 AC-CAG-009~012 의 **When** 절만 그 표면으로 바꾸고 Then 절은 유지한다.

## §D AC Matrix

### 축 (a) — 양방향 회귀

#### AC-CAG-001 — required + 바이너리 부재 → 차단
- **Given** 감사 대상 트리의 `workflow.yaml` 에 `workflow.audit.gates.codex: required` 가 쓰여 있고 codex 바이너리가 PATH 에 없다
- **When** `codex_audit` 를 그 트리의 `project_root` 로 호출한다
- **Then** 결과는 `isError: false`, `verdict == "fail"`, `gate_unmet` 비어 있지 않음, `summary` 에 required 게이트 미충족과 "codex binary not found" 원인이 함께 있다

#### AC-CAG-002 — required + RPC 실패 / 빈 출력 → 차단
- **Given** AC-CAG-001 과 같은 설정이고, codex 는 있으나 (i) RPC seam 이 fail-open inconclusive 를, (ii) 세션 스크립트가 공백뿐인 리뷰 본문을 돌려준다
- **When** 각 경우 `codex_audit` 를 호출한다
- **Then** 둘 다 AC-CAG-001 과 같은 차단 결과이며 `summary` 에 각 원인이 보존된다. `codex_blank_review_test.go` 의 AC-CBR-007 테스트는 `gate_unmet` 단언을 유지하고 verdict 단언만 `fail` 로 바뀐다

#### AC-CAG-003 — 비-required + inconclusive → 정규화 후 바이트 동일
- **Given** `gates.codex` 가 `off`, `advisory`, 키 부재, 설정 파일 부재, YAML 손상인 다섯 트리와, 변경 **전** 코드에서 캡처해 `build_commit`·`build_lag` 를 고정 자리표시자로 정규화한 골든 파일
- **When** 각 트리에서 바이너리 부재 조건으로 `codex_audit` 를 호출하고 같은 정규화를 적용한다
- **Then** 다섯 경우 모두 직렬화 JSON 이 골든과 바이트 동일하다(`verdict == "inconclusive"`, `gate_unmet`·영수증 필드 없음). 그리고 `git log --format=%H -- <골든 파일>` 의 첫 커밋이 축 (a) 구현 커밋의 조상이다(골든이 구현보다 먼저 커밋됨). 이 AC 는 축 (b) 착지 후에도 재실행해 통과한다

#### AC-CAG-004 — 배포 기본값은 opt-in 이 아니다
- **Given** `workflow.yaml` 에 `audit` 블록이 없는 트리
- **When** (i) 같은 트리에 대해 엔진 기본값이 적용된 codex 게이트를 조회하고, (ii) 바이너리 부재 조건으로 `codex_audit` 를 호출한다
- **Then** (i) 은 `required` 로 해석되고(전제 단언 — 기본값이 required 가 아니면 이 AC 는 무의미하므로 실패 처리), (ii) 는 `verdict == "inconclusive"` 이며 `gate_unmet` 이 없다

#### AC-CAG-005 — 실제 판정은 건드리지 않는다
- **Given** `gates.codex: required` 이고 codex 세션 스크립트가 각각 pass 와 fail 을 돌려준다
- **When** `codex_audit` 를 호출한다
- **Then** verdict 는 각각 `pass`, `fail` 이고 `gate_unmet` 은 비어 있다

#### AC-CAG-006 — 다중 경로·Stop-hook 판정 무변경
- **Given** 변경 전 기준 커밋 `f67d2193f` 와 테스트 파일 `required_gate_block_test.go`, `mcp_convergence_test.go`, `mcp_convergence_participant_test.go`, `mcp_audit_multi_test.go`, `multi_review_gate_test.go`, `codex_review_gate_test.go`
- **When** (i) `go test -list 'Converge_|RunMultiAudit_|AuditMulti_|ReviewGate_|MultiReviewGate|PerformCodexAudit_|LoadConvergenceResult_' ./internal/cli/` 을 실행하고, (ii) 같은 regex 로 `go test -run '<regex>' -count=1 ./internal/cli/` 을 실행하고, (iii) `git diff --stat f67d2193f..HEAD -- <위 6개 파일>` 을 실행한다
- **Then** (i) 의 목록에 `TestCodexAudit_RequiredGateUnmetRecordedOnInconclusive` 와 `TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput` 이 없고(이 둘은 AC-CAG-002 에서 의도적으로 바뀌는 단일 도구 테스트), (ii) 는 `ok`, (iii) 의 출력은 비어 있다

#### AC-CAG-007 — 서술 정합
- **Given** 변경 후 트리
- **When** `codex_audit` 도구 설명(`internal/cli/mcp_server.go`), 두 감사자 본문(로컬·템플릿), `moai-ref-cross-model-audit/SKILL.md`(로컬·템플릿)를 판독하고 `make agents-emit-check` 를 실행한다
- **Then** 일곱 곳 모두 "명시적 required 의 무판정은 단일 도구에서도 `verdict: fail` + `gate_unmet`"을 서술하고, 스킬 문서에 "convergence engine reuses" 류 문구가 없으며, `make agents-emit-check` 가 exit 0 이다

### 축 (b) — 영수증 (B-1, 표면 S1)

#### AC-CAG-008 — 영수증 기록과 필드 노출 조건
- **Given** required 트리 R 과 비-required(`advisory`) 트리 A
- **When** 각 트리에서 `codex_audit` 를 한 번, `audit_multi`(codex 참여)를 한 번 호출한다
- **Then** 두 트리 모두 영수증 저장소에 호출당 1건씩 레코드가 생기고(식별자·트리·시각·verdict·gate_unmet 포함), R 의 두 결과에는 그 식별자와 같은 영수증 필드가 있으며, A 의 두 결과에는 영수증 필드가 없다

#### AC-CAG-009 — 영수증 없는/모르는 PASS 거부
- **Given** required 트리에서, 저장소에 해당 감사자 시작 이후 영수증이 없는 상태
- **When** SubagentStop 입력(`agent_type: plan-auditor`, `stop_hook_active: false`)의 `last_assistant_message` 가 (i) 영수증 인용 없는 PASS, (ii) 저장소에 없는 식별자를 인용한 PASS, (iii) 다른 트리의 영수증을 인용한 PASS, (iv) 감사자 시작 이전에 생성된 영수증을 인용한 PASS 로 핸들러를 호출한다
- **Then** 네 경우 모두 출력이 `decision: "block"` 이고 `reason` 이 각각 해당 조건(인용 없음 / 저장소에 없음 / 트리 불일치 / 시작 이전)을 이름으로 밝힌다. `agent_type: sync-auditor` 로 바꿔도 같다

#### AC-CAG-010 — 실제 호출 영수증은 수용, 재진입은 기록+경고
- **Given** required 트리에서 실제 `codex_audit` 호출로 생긴 영수증
- **When** (i) 그 식별자를 인용한 PASS 로 SubagentStop 핸들러를 호출하고, (ii) 인용 없는 PASS 에 `stop_hook_active: true` 로 호출한다
- **Then** (i) 은 `decision` 필드가 없다(수용). (ii) 는 `decision: "block"` 이 없고, `systemMessage` 에 영수증 없는 PASS 는 수용할 수 없다는 경고가 있으며, 저장소에 거부 기록 1건이 생긴다

#### AC-CAG-011 — 비-required 무영향
- **Given** `gates.codex` 가 `off` / `advisory` / 부재인 세 트리
- **When** 영수증 인용 없는 PASS 로 plan-auditor·sync-auditor SubagentStop 핸들러를 호출한다
- **Then** 출력에 `decision`·`systemMessage` 가 없고 저장소에 거부 기록이 생기지 않는다

#### AC-CAG-012 — 판독 불가
- **Given** 영수증 저장소 디렉터리 부재, 레코드 JSON 손상, `last_assistant_message` 공백의 세 조건
- **When** required 트리와 advisory 트리에서 각각 PASS 로 핸들러를 호출한다
- **Then** advisory 는 무출력, required 는 수용하지 않으며 `reason` 또는 `systemMessage` 에 "증거 판독 불가"와 원인이 명시된다

### 축 (c) — 입력 대기 (C-1)

#### AC-CAG-013 — 관측 형태 characterization (RED 먼저)
- **Given** 리드가 확보한 제보자 구성 형태(값은 가짜 자리표시자)
- **When** 분류기를 고치기 **전** 트리에서 해당 테스트를 실행한다
- **Then** `auth_provider == "unknown"` 이 관측되어 테스트가 실패하고, 그 출력이 progress.md §E.2 에 인용된다. 형태를 확보하지 못하면 이 AC 와 AC-CAG-014 는 **기록된 gap** 으로 닫고 분류기는 고치지 않는다

#### AC-CAG-014 — 수리 후 분류
- **Given** AC-CAG-013 의 테스트
- **When** 최소 수리 후 실행한다
- **Then** 관측 형태에 맞는 provider 값(`chatgpt` 또는 `apiKey`)이 나온다

#### AC-CAG-015 — unknown 유지
- **Given** 판독 불가 JSON, 서로 다른 provider 를 말하는 두 줄, 자격증명 없는 파일
- **When** 분류한다
- **Then** 모두 `unknown` 이며 기존 `codex_auth_ladder_test.go` 표 사례가 전부 통과한다

#### AC-CAG-016 — 비밀값 비노출
- **Given** 자격증명 값에 식별 가능한 표식 문자열을 넣은 입력
- **When** `codex_setup` 결과와 분류 경로의 오류 문자열·로그를 수집한다
- **Then** 표식 문자열이 어디에도 나타나지 않는다

## §D.1 Edge Cases

- `gates.codex: "required "`(공백)나 `REQUIRED` — 판별은 정확 일치를 유지하며 비-required 로 취급함을 테스트로 고정.
- `project_root` 가 없고 서버 cwd 가 다른 트리 — 게이트와 영수증은 감사 대상 트리를 따른다.
- 같은 세션에서 `audit_multi` 가 codex 를 **제외**하고 돈 경우 — codex 영수증으로 치지 않는다.
- 감사자가 FAIL 을 보고한 경우 — 영수증 검사 대상이 아니다(PASS 만 검사).

## §D.2 Quality Gate

- `go vet` / `golangci-lint` 대상 패키지 0 건
- 신규·변경 코드 커버리지 85% 이상
- 템플릿 중립성 가드 통과, `make agents-emit-check` exit 0

## §D.3 Definition of Done

- AC-CAG-001 ~ 012 PASS
- AC-CAG-013 ~ 014 는 PASS 또는 C-1 에 따른 기록된 gap, AC-CAG-015 ~ 016 PASS
- §B.4 표면 확정 기록과 M1 페이로드 실측이 progress.md 에 있음
- 미러 쌍·방출본 동기화, `make build` 성공
- 이슈 댓글·종료는 DoD 아님(리드 몫)
