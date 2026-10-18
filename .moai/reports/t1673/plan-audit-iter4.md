auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-PRECOMMIT-AFFECTED-SCOPE-001 (iteration 4, rule C delta: accce8df0 only)

verdict: PASS-WITH-DEBT
note: B1 resolved; the correction commit introduced nothing blocking; carried debt cited by id below.
audited_sha: accce8df084f57a2305457866e4f75a66248074d
overall_score: 0.875
must_pass_failed: 0
blocking_count: 0
plan_artifact_hash: ca50f13c937620eecddb59e5edcc9fa87de698d31e29e8c8e1530130d347144b
scope: delta
convergence_overall: pass
required_backend: codex pass
Iteration: 4/4 (rule C, single delta)
Auditor Version: plan-auditor/v1 (sonnet-5-5)

- debt: S1-a dispose_in=run AC-011 probe scenario demands a non-zero gate exit that REQ-PAS-011/AC-PAS-011 do not state (the gate config loader is fail-soft)
- debt: S1-b dispose_in=run LEDGER-PAS-016-A tree pin is a SHA plus a qualifier in some earlier text (now `576da7f19...`); re-pin to the measuring commit at run
- debt: S2 dispose_in=run AC-PAS-003 Given does not say the new packages carry test files
- debt: S3 dispose_in=run ledger commands carry a `<moai built from ...>` placeholder
- debt: S4 dispose_in=run AC-PAS-001's RED-now probe does not discriminate transitive from direct dependents

Reasoning context ignored per M1 Context Isolation. Scope limited to `git diff 576da7f19 accce8df0` under rule C; nothing else was re-examined and no optional findings are produced.

## 판정

B1(프로브가 기록된 `go test` 로그의 첫 줄만 판정)은 해소되었습니다. 수정 커밋이 이제 로그 전체와 호출 횟수(소스 변경 1 / 테스트만 변경 1 / docs만 변경 0)를 판정하고, 현재 트리에서는 여전히 RED이며, 올바른 호출만 하는 독립 스텁에서는 GREEN에 도달합니다. 같은 스텁에 `go test ./...`를 덧붙이면 RED가 됩니다. 필수 백엔드 codex는 pass이고(영수증 rcpt-2c4e2713d5407d9509dca0f1), 이 커밋이 도입한 시연 가능한 회귀는 없습니다. 이월 부채는 위 `debt:` 줄에 id로 인용했습니다(재감사 범위 밖, 판정 근거 아님).

judging build: HEAD accce8df0의 트리에서 `go build ./cmd/moai`한 `/private/tmp/t1673-audit/bin/moai4`(ldflags 없음). `git diff --stat 3642a3a9b accce8df0 -- internal cmd go.mod go.sum` 출력이 비어 Go 코드는 이전 라운드와 같습니다. MCP 서버와 설치본 `moai`는 빌드 `ee981a932`(HEAD의 조상)입니다.

## Must-Pass Results

범위가 `probe016.sh`, 레저 문단, DoD 2 한 줄이라 요구·AC 본문과 계획은 바뀌지 않았고 이전 라운드 결과가 그대로 유효합니다(MP-1/2/3/5/6/7/9 변경 없음, MP-4 N/A). MP-8: 이 커밋이 고친 셀(LEDGER-PAS-016-A)의 RED가 현재 트리에서 재현됨을 아래 E1로 확인했습니다. 해당 diff가 Go 코드와 SPEC 요구를 바꾸지 않았음을 `git diff` 출력으로 확인했습니다.

## Defects Found

없음. 수정 커밋 자체의 결함과 회귀를 찾지 못했습니다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

C1. 수정된 `scenario()`는 `go.log` 전체의 호출 횟수와 모든 줄을 판정한다.
C2. 프로브는 현재 트리에서 네 시나리오 모두 RED이고 레저가 적은 출력과 일치한다.
C3. 올바른 호출만 하는 독립 스텁 게이트는 네 시나리오 GREEN, 종료코드 0이다.
C4. 같은 스텁에 `go test ./...`를 덧붙이면 세 커밋 시나리오가 RED이다.
C5. DoD 2와 레저 문장이 프로브의 새 판정과 일치한다.

### Evidence

E0 (C1, C5). 명령 `git diff 576da7f19 accce8df0`을 읽은 결과: `scenario()`가 `NCALLS="$(grep -c . "$W/go.log")"`, `GOT="$(grep . "$W/go.log" | tr '\n' ';' | sed 's/;$//')"`, `WANTN`(none이면 0, 아니면 1)을 만들고 `[ "$NCALLS" -eq "$WANTN" ] && [ "$GOT" = "$WANT" ]`를 GREEN 조건에 넣는다. 로그가 비면 `GOT=none`. 래퍼 `go`는 변경 없음(`test`만 기록, 나머지는 실제 `go`로 위임). 레저 Command·Green path 문단과 DoD 2가 "whole recorded go test log and call count"로 수정됨. 헤더 주석에서 두 줄이 한 줄로 붙었으나(`makes it RED.# Usage: ...`) 주석 안이라 동작에는 영향이 없다.

E1 (C2). 명령 `env PROBE_DIR=/private/tmp/t1673-audit/pas-probe4 sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe016.sh /private/tmp/t1673-audit/bin/moai4`, 종료코드 1, 판정 줄 원문(게이트 배너 생략):
```
AC-016/001 change a: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./...')
AC-004 test-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a', got 'test ./...')
AC-006 docs-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 0), wanted go argv 'none', got 'test ./...')
AC-011 mode: bogus: RED  (gate exit=0, message names the key=no)
exit=1
```
레저가 기록한 네 줄과 문자 그대로 일치한다. 사유도 맞다(구현이 없어 게이트가 `./...`를 전달, bogus 모드 미거부).

E2 (C3, C4). 독립 스텁 `/private/tmp/t1673-audit/stubgate.sh`(이 감사가 작성: `gate` 인자를 받아, bogus 모드면 키 이름을 담은 메시지와 종료코드 1, 아니면 `go list`를 호출한 뒤 `git diff --cached --name-only --no-renames`로 스테이징 목록을 읽어 `a/a_test.go`면 `go test ./a`, `a/a.go`면 `go test ./a ./b`, 그 외엔 호출 없음, `EXTRA=1`이면 마지막에 `go test ./...` 추가)을 `sh /private/tmp/t1673-audit/run-controls.sh`로 프로브에 끼워 실행했다. 원문 출력:
```
=== control A: correct calls only
AC-016/001 change a: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./a ./b')
AC-004 test-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a', got 'test ./a')
AC-006 docs-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=0 (wanted 0), wanted go argv 'none', got 'none')
AC-011 mode: bogus: GREEN  (gate exit=1, message names the key=yes)
exit=0
exitA=0
=== control B: correct calls + extra go test ./...
AC-016/001 change a: RED  (commit exit=0, HEAD moved=yes, go test calls=2 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./a ./b;test ./...')
AC-004 test-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=2 (wanted 1), wanted go argv 'test ./a', got 'test ./a;test ./...')
AC-006 docs-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 0), wanted go argv 'none', got 'test ./...')
AC-011 mode: bogus: GREEN  (gate exit=1, message names the key=yes)
exit=1
exitB=1
```
스텁은 `go list`를 호출했으나 래퍼가 실제 `go`로 위임했고 로그에는 `test`만 남았다. 이로써 "그래프 질의를 먼저 하는 구현은 통과, 추가 `go test`는 RED"가 실증된다. 이전 라운드의 구멍(첫 줄만 판정)이었다면 B의 첫 두 시나리오는 GREEN이었을 것이다.

E3 (필수 백엔드). `mcp__moai__audit_multi`(project_root, card_id=t1673, 같은 target 형태, 범위를 이 커밋으로 한정한 focus) 결과: codex `gate: required`, `verdict: pass`, findings 비어 있음, `overall_verdict: pass`, `residual_risk_note: ""`, `audit_receipt: "rcpt-2c4e2713d5407d9509dca0f1"`, `plan_source: config`, `build_commit: ee981a932`. 다이제스트를 `.moai/state/audit-plan-result.json`에 쓰고 `moai verify audit-plan --project-root ... --result-file ...`를 돌리자 `enforced_required: ["codex"]`, `convergence_check: {"ok": true, "unmet": []}`.

### Baseline-attribution

- 트리: HEAD `accce8df084f57a2305457866e4f75a66248074d`. 관측은 모두 이 실행에서 이 트리에 대해 이뤄졌다. 레저가 핀한 `576da7f19...`는 부모 커밋이며, 프로브·스텁 실행은 HEAD의 프로브 파일로 했다.
- 도구 빌드(§2.2): E1은 HEAD에서 빌드한 `moai4`, E2는 스텁(모의 게이트)을 사용. MCP 호출과 `moai verify audit-plan`은 서버/설치본 빌드 `ee981a932`(HEAD의 조상)가 처리.
- `plan_artifact_hash`는 `ComputeHash` 알고리즘을 읽고 셸로 재구현한 값(acceptance·decision-index·plan·spec 네 파일).

### Gaps (미검증)

- G-a. 스텁은 이 감사가 만든 모의 게이트이며 실제 구현이 아니다. "실제 구현이 오면 GREEN에 도달한다"는 구조적 주장은 래퍼·로그 판정 경로를 스텁으로 통과시켜 본 것이지, 구현으로 확인한 것이 아니다.
- G-b. 범위 밖(규칙 C)이라 `probe009.sh`, 013-A, 추적 스크립트, CN-4, 린트는 이번 라운드에 다시 실행하지 않았다. 이전 라운드 결과와 "SPEC 요구·Go 코드 미변경"(빈 `git diff --stat`)에 의존한다.
- G-c. 이월 부채 S1-a, S1-b, S2~S4는 판정 근거가 아니며 재검토하지 않았다.

### Residual-risk

- 부채 S1-a(AC-011 시나리오의 종료코드 요구)로 구현이 정상이어도 프로브의 마지막 시나리오가 RED로 남을 수 있다. run 단계에서 정리해야 한다.
- Q1=B 결과(흔한 커밋의 affected set이 예산을 넘을 수 있음), 순수 단위 AC(002, 003)의 판별력 미확정(P1), R3는 SPEC이 수용한 위험이다.

AUDIT-VERDICT: PASS-WITH-DEBT spec=SPEC-PRECOMMIT-AFFECTED-SCOPE-001 receipts=rcpt-2c4e2713d5407d9509dca0f1
