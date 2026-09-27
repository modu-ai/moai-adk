# t1236 판정서 — plan·plan-audit 단계

- 카드: t1236 (AUTONOMY-A3 계약 기반 자율 하네스 — 게이트 재배선, Class C)
- SPEC: `SPEC-AUTONOMY-GATE-REWIRE-001` v0.3.2, Tier L
- 워크트리: `.claude/worktrees/t1236`, 브랜치 `WT-contract-gate-rewire`, 기준 develop `ca1d5dc43`
- 배차 범위: `/moai plan` → plan-audit 까지. run 은 선행 카드 병합 뒤.

## 주장

plan 단계는 끝났고 plan-audit 은 Tier L 반복 상한(3회)에 닿았다. 최종 판정은 **FAIL 0.847** 로, 문턱 0.85 에 0.003 못 미친다. blocking 결함 4건(D43·D44 major, D45·D46 minor)이 남아 있다. 자동 반복은 더 없으므로 다음 처분은 리드·운영자 결정이다.

## 증거

| 회차 | 대상 커밋 | 판정 | 점수 | 보고서 |
|---|---|---|---|---|
| 1 | `62a65f709` (v0.2.0) | FAIL | 0.71 | `.moai/reports/t1236/plan-audit-1.md` |
| 2 | `1b071a573` (v0.3.1) | FAIL | 0.77 | `.moai/reports/t1236/plan-audit-2.md` |
| 3 | `710530d67` (v0.3.2) | FAIL | 0.847 | `.moai/reports/t1236/plan-audit-3.md` |

- v0.1 (`ca1e39f6f`) 감사는 범위가 추가돼 중단했다. 보고서는 만들어지지 않았다.
- 3회차 영역별 점수: Clarity 0.80 · Completeness 0.90 · Testability 0.80 · Traceability 0.90.
- MUST-PASS: MP-1·2·3·5·6·7 PASS, MP-4 해당 없음.
- 교차 모델 감사: codex 는 사용량 한도, glm 은 inconclusive. 세 회차 모두 plan-auditor 단독 판정이다.
- `moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001` → `✓ No findings — all SPEC documents are valid`, exit 0 (v0.3.2, 이 세션에서 실행).
- `grep -c '\[NEEDS CLARIFICATION'` 은 plan.md 에서 0.
- `go test ./internal/spec/ -count=1` → `ok ... 99.265s`, exit 0. 이 측정은 v0.2.0 `62a65f709` 기준이고, 이후 판에서는 다시 재지 않았다(Gaps 참조).

## 기준선 귀속

- 모든 측정은 이 워크트리 HEAD 에서 했다. 커밋 이력은 `ca1e39f6f → 62a65f709 → 781ddc355 → 1b071a573 → 710530d67` 이다.
- 선행 SPEC 인용 기준:
  - A1 0.5.2 `25283ebf8` (t1234, 작성자 에이전트가 `cmp` 로 워크트리 파일과 커밋본이 같음을 확인)
  - A2 는 리드가 확정한 최종 형식(현재 커밋 `d8926ff9a` 는 낡은 초안)
  - A2b (t1245) 는 SPEC 없음

## 남은 결함 — PASS 에 필요한 최소 변경 (plan-audit-3.md 원문 요약)

1. **D43 (major)**: kickoff-check 에서 사유가 여럿 성립할 때 우선순위가 정의되지 않았다. MUST-PASS AC-GR-016 의 기대값이 구현 순서에 따라 달라진다.
   - 수리: REQ-GR-007 에 「`--json` 은 성립하는 사유를 모두 담고, 대표 사유는 정해진 우선순위를 따른다」를 넣는다.
   - 픽스처 (12)(13) 에는 체인이 맞는 위조 사건 줄을 넣는다.
   - 픽스처 (6) 은 `not-signed-valid` 도 사유 목록에 있는지 단언한다.
2. **D44 (major)**: design.md:L170 은 「이 테스트는 매 커밋의 CI 에서 돈다」를 전제로 하지만, 리드 일괄 push 체제에서는 CI 가 head 에서만 돈다. 이 전제는 거짓이다.
   - 수리: 서명기 단계 (1) 이 doctrine 플래그를 주입받는다고 명시하고, `TestSignInterimRuleFollowsDoctrine` 이 매 head 에서 두 상태를 모두 주입해 검사하게 한다.
   - AC-GR-017 픽스처 5 를 이에 맞게 바꾼다.
3. **D45 (minor)**: acceptance.md 의 A1 교차 참조 `AC-CONTRACT-016` 두 곳에 `[REF]` 표지가 없다. 그래서 정본 카운터가 AC 를 26개로 센다(표지를 붙이면 live 25).
   - 스냅숏에 이 SPEC 행이 기록되기 전에 고쳐야 한다.
4. **D46 (minor)**: REQ-GR-012 와 design.md:L234 에 원천 없는 「moai 측 세션 식별자」가 남아 있다.
   - 수리: 원천을 정하거나 그 항목을 지운다.

선택 수정: D47 (`plan_artifact_hash:` 생산자가 없다는 사실을 활성 표에 명시), D48 (개수 문구), D49 (`refusals.go` 인용 갱신 — A1 브랜치 `7b0d20494` 이후 커밋됨).

## 처분 선택지 (리드·운영자)

- **(a) 명시적 iter-4**: 위 네 건을 고친 v0.3.3 으로 감사를 한 번 더 한다. 모두 기존 REQ·AC 문장과 픽스처 수정이라 REQ 상한 25 안에서 된다.
- **(b) PASS-WITH-DEBT**: D43·D44 가 MUST-PASS AC-016·017 의 판정 기준이므로, run 의 RED 설계 전에 반드시 닫는다는 조건을 붙인다.
- **(c) 범위 축소**: 예를 들어 Go 코드 층(revoke·decide·kickoff-check)과 문서 층을 두 SPEC 으로 나눈다.

## 미관측 (Gaps)

- Go 테스트는 v0.2.0 이후 다시 돌리지 않았다. SPEC 문서만 바뀌었고 코드는 바뀌지 않았다.
  - 다만 acceptance.md 를 개정했으므로, 병합 전에는 `./internal/spec` 을 병합 트리에서 다시 재야 한다.
- A2b SPEC 과 A2 개정본은 아직 없어서 읽지 못했다.
- `refusals.go` 관측(reject/human 소비자)은 t1234 작업본에서 한 것이다. 감사에 따르면 이후 커밋됐으므로 그 SHA 로 다시 재야 한다.
- 새 CLI 동사(revoke·decide·kickoff-check)는 아직 없다. RED 관측은 A1 병합 뒤 run 단계에서 한다.

## 잔여 위험

- 작성자 배제는 자기 신고한 신원에 기댄다. 거짓 신고는 막지 못한다.
- `plan_artifact_hash:` 생산자가 생기기 전까지 자율 Kickoff 결과는 늘 사람으로 간다. 안전한 방향이지만, 그동안 기능에는 도달할 수 없다.
- 선행 SPEC(A1·A2·A2b)이 바뀌면 재확인 표식이 붙은 항목을 다시 맞춰야 한다.
- AC 스냅숏: 이 SPEC 행은 아직 기록되지 않았다(커밋 가드 「unrecorded, report only」). D45 를 고치기 전에 기록되면 26 으로 굳는다.

## D50 처리 (plan-audit-4 이후)

- 4차 감사: PASS-WITH-DEBT 0.892 (`f87795581`, `.moai/reports/t1236/plan-audit-4.md`).
- 리드 판정 D50: 리드 중계 승인은 유효한 확인이 아님. design.md L54·L267 유지, §11.1 승인 기록을 「리드 중계, 레인 재확인 전」으로 표기 + M7b 착수 때 레인 운영자 재확인 줄 추가 → `604ee7b8b` 로 debt 종결.
- 선택 debt D51~D53 은 남김. plan 단계 종료, run 전제 t1234·t1235·t1245 병합 + t1175 흡수.

## plan-audit 결과 — PASS-WITH-DEBT (리드 결정 2026-09-27)

### 주장

plan-audit 은 PASS-WITH-DEBT 로 닫는다. 최종 6회차(`.moai/reports/t1236/plan-audit-6.md`, 감사 대상 `a02e7c02b`)는 **FAIL 0.868** 이었다. 점수는 Tier L 문턱 0.85 를 넘었지만, 리드 규칙(D47·D48·D49 중 하나라도 열려 있으면 FAIL)에 따라 D47 이 부분 미해결이라 FAIL 로 판정됐다. 남은 blocking 결함은 D56 하나다. 리드는 7차 감사를 열지 않고, D56 을 SPEC 한 곳의 좁은 편집과 기계적 측정 가능성 탐침으로 닫기로 결정했다. 선택 결함 D57·D58·D59 도 같은 커밋에서 고쳤다(SPEC v0.3.6).

- 5회차: FAIL 0.841 (`f40824c82`, `.moai/reports/t1236/plan-audit-5.md`) → v0.3.5 `a02e7c02b` 에서 D47~D55 수리.
- 6회차: FAIL 0.868 (`a02e7c02b`, `.moai/reports/t1236/plan-audit-6.md`) → D56 수리와 이 탐침.

### D56 수리 내용

`internal/constitution.Validate` 에는 `ZONE_UNREGISTERED`·`ANCHOR_NOT_FOUND` 항목을 만드는 코드 경로가 없다(6회차 감사 실측). 그래서 AC-GR-003 의 미등록 `[HARD]` 반증을 테스트 자체의 `[HARD]` 집합 비교로 바꿨다. 대상은 이 SPEC 이 편집하는 always-loaded 파일 세 개의 로컬·템플릿 사본 여섯 개이며, 파일마다 `[HARD]` 를 담은 줄의 중복 제거 집합이 현재 트리에서 BASE 의 부분집합이어야 한다. AC 본문에 「`constitution.Validate` 는 이 범주를 내지 않는다 — `[HARD]` 집합 비교가 판정이다」를 적었다.

### 측정 가능성 탐침 — 증거

- 트리: BASE `7fe658815eb0d4110b9acadad56e5a85bee3ed3f`(트리 객체 `1405c408103c9b7cf11728e36c3f899e1ecf5fee`, `git archive 7fe658815` 을 세션 스크래치 `base-7fe658815/` 에 푼 사본), 현재 트리 HEAD `a02e7c02b4a14b6d3edc422975f125b82b9acab4`(트리 객체 `386ac53305af4216f63e4c45421b5c1786c8471d`). `git diff --name-only 7fe658815 HEAD` 는 이 SPEC 디렉터리와 감사 보고서만 내므로 대상 여섯 파일은 두 트리에서 같다.
- 집합 정의(AC-GR-003 4번과 같음): 파일마다 `grep -F '[HARD]' <파일> | LC_ALL=C sort -u`, 비교는 `LC_ALL=C comm -13 <BASE 집합> <현재 집합>` 의 출력이 비어 있는지.
- 실행: 세션 스크래치의 `t1236-probe/probe.sh`(위 정의를 여섯 파일과 변이 사본에 적용). 명령 `sh .../t1236-probe/probe.sh > .../t1236-probe/probe-out.txt 2>&1`, 종료 코드 **1**(변이 단계가 RED 이면 1 을 내도록 작성). 출력 원문:

```text
== (a) BASE 7fe658815 vs current tree ==
GREEN CLAUDE.md base=3 cur=3 comm-13=empty
GREEN internal/template/templates/CLAUDE.md base=3 cur=3 comm-13=empty
GREEN .claude/rules/moai/core/askuser-protocol.md base=11 cur=11 comm-13=empty
GREEN internal/template/templates/.claude/rules/moai/core/askuser-protocol.md base=11 cur=11 comm-13=empty
GREEN .claude/rules/moai/workflow/goal-directive.md base=0 cur=0 comm-13=empty
GREEN internal/template/templates/.claude/rules/moai/workflow/goal-directive.md base=0 cur=0 comm-13=empty
(a) result: fail=0
== (b) mutant: askuser-protocol.md copy + one unregistered [HARD] line vs BASE ==
RED (expected) extra:
[HARD] Probe-only unregistered rule inserted by the t1236 falsifier.
```

- 단일 호출 재확인 두 건(스크립트가 만든 집합 파일에 대해):
  - (a) GREEN — `LC_ALL=C comm -13 .../t1236-probe/base__claude_rules_moai_core_askuser-protocol_md .../t1236-probe/cur__claude_rules_moai_core_askuser-protocol_md` → stdout 없음, exit 0.
  - (b) RED — `LC_ALL=C comm -13 .../t1236-probe/base__claude_rules_moai_core_askuser-protocol_md .../t1236-probe/mut__claude_rules_moai_core_askuser-protocol_md` → stdout `[HARD] Probe-only unregistered rule inserted by the t1236 falsifier.`, exit 0.
- 판정: (a) 는 GREEN, (b) 는 RED 다. 두 결과가 갈리므로 새 비교는 판정력이 있다.

### 남은 부채

- D51 잔여(plan.md §H 의 낡은 A2 표지)는 D57 로 이번에 고쳤다. 남은 D51 잔여는 없다.
- D57·D58·D59 는 모두 이번 커밋에서 고쳤다. 미수리 선택 결함은 없다.
- D47·D48 의 변이 제거 주장(판독기를 부르지 않는 kickoff-check·decide 변이가 AC-GR-016 (17)·AC-GR-018 (e) 에서 FAIL, 건너뛰기 변수 변이가 전제 단언에서 FAIL)은 아직 AC 문언에서 추론한 것이다. 해당 Go 테스트가 없으므로 run 단계에서 RED 를 실제로 관측할 때까지 판정이 아니라 가설이다.
- 이 탐침은 셸 동치다. run 단계의 `TestContractModeConstitutionDriftNotIncreased` 가 같은 비교를 Go 로 구현하고 반증 하위 테스트 RED 를 `progress.md §E.2` 에 남겨야 이 AC 가 완성된다.

### 잔여 위험

- 집합 비교는 중복을 없앤 집합이라, 블록 안에 **기존 `[HARD]` 줄과 글자까지 같은 줄**을 새로 넣으면 잡지 못한다. 미등록 규칙의 유입은 막지만 기존 규칙 문장의 복제는 이 비교 밖이다.
- `goal-directive.md` 두 사본은 BASE 에 `[HARD]` 줄이 0개라 어떤 `[HARD]` 줄이든 들어오면 실패한다. 이는 의도한 동작이다.

## run·sync 결과 — PASS-WITH-DEBT (2026-09-28)

적용 규칙: `verification-claim-integrity.md` §1(관측하지 않은 주장 금지), `verification-completeness.md` §1.1·§2. 아래 수치는 모두 `progress.md` §E.2·§E.3 에 남은 명령과 출력, 또는 이번 sync 에서 직접 실행한 명령에서 가져왔다.

### 주장

- run 단계는 AC 25/25 PASS 로 닫혔다. M8(자율 Kickoff 활성화)은 보류 부채로 종결했다. 리드는 run 결과를 PASS-WITH-DEBT 로 판정했다.
- sync 단계에서 CHANGELOG 항목, docs-site 4개 로케일의 `moai contract` 명령 문서, `spec.md` 상태 전이(in-progress → implemented → completed)를 한 커밋에 담았다.
- 자율 Kickoff 는 꺼진 채 출하된다(`autonomousKickoffEnabled = false`). 따라서 Implementation Kickoff Approval 사람 게이트는 실제 동작에서 바뀌지 않는다.

### 증거

**마일스톤 커밋**

| 단계 | 커밋 |
|---|---|
| M0 | `7e82f8b66` |
| M1 | `fdf274c42` |
| M2~M5 | `1e7e483f3` |
| M9 | `39ca88f50` |
| M6 | `5f8a67b78` |
| M7 | `0e1f2edb9` |
| M10 | `0fff55f64` |
| M7b | `185569ef3` |
| 커버리지 보강 | `62726e1e4` |
| revoke HEAD 판독 GIT_DIR 수리 | `8e504df95` |
| run_commit_sha 기입 | `a80c2941f` |
| SPEC v0.3.7 — sync 산출물 허용 목록(design §2 27행) | `0864b2a09` |
| `grAllowed` 미러 | `1a1746794` |

**AC 25/25 PASS**
- M10 행렬: `progress.md` §E.2 「M10 — AC matrix」(트리 `ec051a27b332a7e4bb3cee2715bb9762680092df`, 로그 `.moai/state/verify/t1236/ac/<AC>.txt`). 24건이 PASS, AC-GR-022 는 빈 선택으로 DEFERRED 였다.
- AC-GR-022 는 M7b(`185569ef3`) 뒤에 PASS 로 바뀌었다. 근거는 `progress.md` §E.2 「M7b / M8」에 기록된 `--- PASS: TestJevDoctrineAmendment`, `--- PASS: TestJevAmendmentLinkage` 이다.
- §E.3 기록: `ac_pass_count: 25`, `ac_fail_count: 0`.

**변이 제거 (FAIL 후 원복, HEAD 에서 PASS)**
- kickoff 판독기 우회: `check_test.go:147: pass=true reason="" reasons=[] state=signed-valid, want pass=false reason="revoked"` / `--- FAIL: TestKickoffCheck/17_revoke_record_only`. HEAD 에서 `ok …/internal/contract/kickoff`.
- decide 판독기 우회: `decide_test.go:274: outcome "human" reason "jev-doctrine-not-amended", want human "precondition:e"` / `--- FAIL: TestDecidePreconditions`. HEAD 에서 `ok`.
- 미등록 `[HARD]` 줄: `contract_mode_guided_test.go:422: … unregistered [HARD] line not present at the base: …` / `--- FAIL: TestContractModeConstitutionDriftNotIncreased`. HEAD 에서 `ok …/internal/template`.
- `specTier` 기본값 `"L"`→`"M"`: `units_test.go:62: specTier = "M", want "L"` / `--- FAIL: TestSpecTier`.
- 판독기 `.md` 필터 제거: `failures_test.go:179: Blocked = true, err parse …notes.txt: escalation: record has no YAML frontmatter; want false, nil` / `--- FAIL: TestBlockedReaderShapes`.
- 두 변이 모두 `cmp` 로 원복을 확인했고, 원복 뒤 HEAD 는 exit 0 이었다.

**GIT_DIR 누수 (재현 먼저)**
- RED: `failures_test.go:215: record head "fe9daa6f…", want the project's HEAD "cdf7e7e6…"` / `--- FAIL: TestRevokeHeadIgnoresAmbientGitEnv`.
- GREEN: `--- PASS: TestRevokeHeadIgnoresAmbientGitEnv (0.39s)` / `ok`. 수리는 `cmd.Env = gitenv.Env()` 한 줄이다.
- 형제 스윕: `internal/contract` 안의 테스트 외 `exec.Command` 는 모두 정리된 환경으로 실행된다.

**M7b 기록**
- 운영자가 레인 세션에서 AskUserQuestion 으로 §29 문구를 확인했다(선택지 「넣기 (권장)」).
- 연계 변경은 단일 커밋 `185569ef3` 로 들어갔다. 커밋된 트리에서 `--- PASS: TestJevAmendmentLinkage` 를 관측했다.

**M8 (원문 그대로)**
- "M8 보류 — design §7.1 6행 미충족(Frozen 문단에 비인간 결정자 서명 등가 부재), autonomousKickoffEnabled=false 유지, 3행 충족·4행 부분 확인"
- 리드 처분: 보류 부채로 종결.

**리드 결정**
- B1~B5(2026-09-27). B3 은 선택지 (a)로 정했다: 재개 차단은 A3 revoke 판독기가 맡고 A2 는 바꾸지 않는다. Jev 신뢰도 문턱은 0.80 이다.
- plan-audit 과 run 은 모두 PASS-WITH-DEBT 로 판정했다.

**허용 목록 개정 (레인 결정, 선택지 (a))**
- 처음 sync 초안을 넣었을 때 `TestContractModeChangeSetAllowlist` 가 `CHANGELOG.md` 와 docs-site 4개 파일을 허용 목록 밖으로 보고했다.
- 조치 순서: SPEC v0.3.7 design §2 에 27행 추가(`0864b2a09`), 테스트 `grAllowed` 에 같은 경로 반영(`1a1746794`).
- 이번 sync 에서 다시 측정한 결과는 아래 「sync 검증」에 있다.

**sync 검증 (이번 실행)**
- 결과는 커밋 메시지와 manager-docs 보고에 원문으로 남긴다. 명령:
  - `MOAI_GR_BASE=7fe658815 go test ./internal/template/ -run TestContractModeChangeSetAllowlist -count=1 -v`
  - `go run ./cmd/moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001`
  - `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json`

**문서 표면**
- `docs-site/content/{ko,en,ja,zh}/cli-reference/contract.md` 에 `kickoff-check`·`decide`·`revoke` 행과 설명 절을 넣었다. 네 로케일 모두 `##` 절이 4개로 같다.
- README 에는 `moai contract` 항목이 없어 고치지 않았다.

### 기준 귀속

- run 수치의 기준 트리: M10 행렬은 `ec051a27b…`(HEAD `0e1f2edb9`), AC-GR-022 는 `185569ef3`, 커버리지와 GIT_DIR 수치는 `62726e1e4`·`8e504df95` 트리에서 쟀다.
- BASE 는 `MOAI_GR_BASE=7fe658815` 이다.
- sync 검증은 싱크 커밋 직전의 작업 트리에서 실행했다. 그 트리는 `1a1746794` 에 이번 sync 편집을 더한 상태다.

### 미검증

- **BASE 파생 산출물**: `grBaseDriftIDs`, `grKickoffClasses`, EV-6 id 들. BASE 는 t1175 수리 이전 트리이므로, develop 을 흡수할 때 다시 생성해야 한다.
- **receipt 패키지 커버리지**: 단독 측정 68.1% 로 목표 85% 에 못 미친다. kickoff 87.4%, revoke 86.0% 는 목표를 넘었다.
- **M8 4행**: push 직렬화 쪽 테스트는 다시 돌리지 않았다. 같은 행의 sign/decide 가드 8건은 PASS 였다.
- **교차 플랫폼 빌드**: darwin/windows 빌드는 M7 에서 마지막으로 쟀고, M7b 이후에는 다시 돌리지 않았다.
- **헌법 검증**: `moai constitution validate` 의 DRIFT 9건은 BASE 에도 있던 것이다. 리드가 별도 카드로 처리한다.
- **sync 범위**: 전체 테스트 스위트는 돌리지 않았다. CI 가 판정한다.

### 잔여 위험

- 자율 Kickoff 경로는 꺼져 있어 실제 사용 환경에서 한 번도 돌지 않았다. 켜는 시점에 새 결함이 드러날 수 있다.
- 체인 저장소와 에스컬레이션 기록을 함께 지우는 행위자는 막지 못한다(design D21).
- 허용 목록이 sync 산출물을 받도록 넓어졌다. 이후 같은 경로에 무관한 변경이 들어와도 이 가드는 잡지 못한다.

## sync-audit 결과와 F1 처리 (2026-09-28)

- 감사: `.moai/reports/t1236/sync-audit.md` (첫 줄 `auditor-model: claude-opus-5-5[1m]`). 판정 FAIL, 원인은 blocking 결함 F1 한 건이다. Functionality 90, Security 85는 통과했고 가중 조화평균은 82.7이다. 감사관은 F1만 고치면 PASS-WITH-DEBT라고 판단했다.
- F1: `CHANGELOG.md:12`가 새 SSOT를 "always-loaded"로 적었으나, 실제 파일은 `paths:` 범위로 한정돼 있다(`contract-autonomy.md:3`). 이를 "path-scoped"로 고쳤다.
  - 수정 전 증거: 감사 보고서의 F1 항목.
  - 수정 후 증거: `grep -n 'always-loaded SSOT' CHANGELOG.md` → 출력 없음, exit 1. `grep -c 'path-scoped SSOT \`.claude/rules/moai/workflow/contract-autonomy.md\`' CHANGELOG.md` → `1`.
- F1 이후 레인이 재감사를 돌리지 않았다. F1 해소는 위 grep 증거로만 확인했고, 최종 판정은 리드에게 맡긴다.
- 선택 결함(부채로 남김):
  - F2: `receipt/store.go:147`의 prev 링크 검사를 무력화한 변이가 살아남았다(원본 동작은 정상). 회귀 테스트가 필요하며 receipt 커버리지 부채와 겹친다.
  - F3·F4: 자율 킥오프 활성화 전 정리 과제다. F3은 decide가 세 번에 나눠 쓰는 문제와 ErrIntegrity가 감싸기에서 사라지는 문제, F4는 kickoff-check가 decide 사건을 대조하지 않는 문제다.
  - F5: 서명기 `gitEnv()` 스크럽 범위가 좁다. BASE부터 있던 결함이다.
  - F6: kickoff-check·decide가 카드 id를 `ValidCard`로 검증하지 않는다.
  - F7: 이 판정서의 sync 검증 절에는 출력 원문이 없다. `--baseline` lint는 레인이 커밋된 트리 `1f0a99279`에서 다시 돌려 exit 0, `baseline: OK`를 얻었다.
  - F8: design §11.1의 승인 기록이 낡았다.
  - F9: BASE 참조 가드 6개는 `MOAI_GR_BASE`가 없으면 SKIP되므로 CI와 병합 이후에는 돌지 않는다.

## 리드 최종 판정 (2026-09-28)

리드는 PASS-WITH-DEBT로 판정하고 카드를 병합 대기열에 넣었다. 근거는 감사관이 sync-audit.md 105행에 처방한 재감사 범위이며, 리드가 이를 그대로 다시 쟀다. `git show af24dfbcc:CHANGELOG.md | grep -c 'always-loaded SSOT'`는 `0`을 냈고, CHANGELOG.md 12행은 "path-scoped SSOT"로 읽힌다. `contract-autonomy.md` 머리에는 `paths:`가 있다. F1 커밋이 바꾼 것은 CHANGELOG 1행과 reports 2파일뿐이다. 필수 두 차원은 이미 PASS였다(Functionality 90, Security 85).

병합 순서는 t1282 → t1099 → t1237 → t1226 → t1243 → t1236 → t1286이다. 흡수한 뒤 할 일은 두 가지다. BASE에서 파생한 산출물을 흡수한 트리에서 다시 만들고, `moai constitution validate`의 DRIFT가 0인지 확인해 "constitution DRIFT 9 기존" 줄을 그 결과로 고친다.

## 흡수 후 재측정 (병합 창, 2026-09-28)

- 흡수: 로컬 develop `5f5840ae4`를 흡수했다(흡수 커밋 `00fee4131`). 충돌은 CHANGELOG 하나였고 양쪽 항목을 모두 살려 해결했다. 이어서 BASE 파생 데이터를 다시 만들었다(`14ecb399c`). 새 BASE는 `5f5840ae4`이고, `grBaseDriftIDs`는 빈 집합이 됐으며 `grKickoffClasses`는 바뀌지 않았다.
- constitution: 트리에서 빌드한 바이너리로 `moai constitution validate`를 돌렸다. 결과는 `OK — no drift or violations detected (97 of 101 entries checked)`, exit 0, `drift_count: 0`이다. 앞 절의 "constitution DRIFT 9 기존" 부채는 t1175 수리가 흡수되면서 **해소**됐다. 9건은 옛 BASE `7fe658815` 시점의 측정값이다.
- 충돌 점검: t1243(frozen 집합에 AGENTS*.md 추가)과 t1237(closure 계약) 모두 실제 충돌이 없다. 명령 이름, 저장소 파일, frozen 경로가 겹치지 않는다. t1237의 `LoadA2Records`는 revoke 기록을 resolved로 읽으므로 호환된다.
- 병합 트리 재측정은 모두 exit 0이었고, SKIP과 빈 선택은 없었다.
  - `./internal/contract/...`: 6개 패키지 ok.
  - AC-GR 가드: 최상위 16개 PASS(`85 changed paths checked`).
  - `./internal/template` 전체: ok.
  - A2b 가드: 8개 PASS.
  - cli `^TestContract`: 3개 PASS.
  - `./internal/spec`: ok.
  - lint v2.1.6: 0 issues.
  - `agents-emit-check`와 `make build`: 통과.
- 새 부채(병합을 막지 않음): SPEC 본문의 EV-6 원장과 AC-GR-003 2단계 서술(acceptance.md L46·L51·L333 이하, plan.md L213)이 여전히 옛 BASE `7fe658815`의 DRIFT 9건을 전제로 적혀 있다. 과거 측정 기록으로서는 사실이지만, 새 BASE에서 테스트가 쓰는 전제(빈 집합)와는 문면이 어긋난다. manager-spec의 비전이 정정 대상이며 리드가 부채로 받는다. 또 `contract --help`의 긴 설명문이 새 동사를 반영하지 않는다(A1 소유 문구).

- 병합 후 gitenv 가드 적색 — kickoff TestMain 누락, t1237 창에서 수리 (`internal/contract/kickoff/activation_test.go:35`가 git을 직접 실행하는데 `gitenv.ScrubProcess()`를 부르는 TestMain이 없어 `internal/gitenv` TestFixturePackagesScrubProcess가 적색. 수리는 worker-62가 t1237 창에서 card t1236 귀속 커밋으로 진행. 카드 셀렉터가 다른 패키지의 가드를 놓친 사례.)

## 병합 후 CI 적색 수리 (2026-09-28, WT-gate-rewire-ci)

develop `e9577de4f` CI run 36333964365 적색(Test ubuntu + Race) 두 갈래 수리. base `088594d6b`.

### ① MOAI_HOME 쓰기 — 원인 정정
- 재현: `TestObserveInertUnderGuided` `detector_paths_test.go:25: guided mode wrote under MOAI_HOME: [d db/]`, `TestDetectorNeverAltersToolCall` `escalation_m5_test.go:232: … db/t9001-8d01049e/contract/events.jsonl`, `…/store.lock`.
- 원인: 감지기가 아니다. `escalation.Active` 는 contract 에서만 참이고 `Observe` 는 guided 에서 즉시 반환한다. 쓰는 쪽은 픽스처 준비 단계 `escalationtest.AddSpec → sign.Sign` 이며, 이 SPEC 이후 `Seams.RecordEvent` 기본값이 MOAI_HOME 저장소에 서명 이벤트를 남긴다. hook 테스트는 저장소 디렉터리명에 워크트리 해시가 붙어 두 실행의 준비 파일이 서로 다른 경로로 보였다.
- 수리: 두 단언을 관측 호출 전후 MOAI_HOME 차이(경로+크기)로 좁힘. 커밋 `3596dbeaf`.
- 변이 대조: `Active` 를 `return true` 로 바꾸면 `TestObserveInertUnderGuided` FAIL.
- **부채**: `TestDetectorNeverAltersToolCall` 은 같은 변이를 잡지 못한다(원래부터). escalation 경로 파일은 허용 목록이고, guided 실행이 contract 와 같은 출력을 내도 비교가 통과한다. 이 약점은 이번 수리 이전부터 있었다.

### ② 스킬 LOC 상한 — 블록 밖 산문 원문 이관
- 재현: `run.md has 207 LOC (ceiling: 200)`, `plan/spec-assembly.md has 616 LOC (ceiling: 600)`.
- 블록 원위치 앵커 방식(리드 1안)은 줄 수로 불가능: 이 SPEC 이전 LOC 199/600, 앵커 최소 3줄 × id 수 → 205/612.
- 리드 결정 (1): CI LOC 상한 때문에 블록 **밖** 산문을 하위 파일로 원문 그대로 옮김 — 계약 모드 블록은 원위치·같은 id 유지(emitter 배치 결정 불변).
  - `run.md` § 3. Autonomy invariants (9줄) → `run/phase-execution.md` 끝 「Run-phase Autonomy invariants (moved from run.md, verbatim)」. 원 자리에 한 줄 포인터. 199줄(템플릿 200).
  - `plan/spec-assembly.md` Step 2.3.3a Plan HTML Report Emission (21줄, [HARD] 1개 포함) → `plan/context-discovery.md` 끝. 원 자리에 한 줄 포인터. 596줄.
  - 로컬·템플릿 두 사본 동일 적용.
- 로드 경로: 출발지·도착지 모두 `paths:` 없는 `user-invocable: false` 하위 스킬로, 오케스트레이터가 `Read` 로 싣는다. 포인터 줄이 그 자리에서 도착지 `Read` 를 지시하므로 옮긴 [HARD] 조항은 원래와 같은 시점(해당 단계 진입 시)에 로드된다. 단 자동 로드가 아니라 포인터를 따르는 Read 에 의존한다 — 원래도 spec-assembly 자체가 Read 로 로드됐으므로 로드 방식은 같다.
- AC-GR-001 보존 검사: 이관 목록(`grRelocations`)에 한해 도착지 절을 포인터 자리에 되돌린 뒤 base 와 바이트 비교. 변이 대조: 이관된 한 줄(`Transcript-measurability`→`…biliti`)을 바꾸면 `TestContractModeGuidedPreservation` FAIL, 복원 확인.
- 허용 목록: 이관 도착지 2개와 ①의 테스트 2개 추가. `internal/template/catalog.yaml` 해시 재생성.

### 재측정 (이 트리)
- `go test ./internal/escalation/... ./internal/hook/... ./internal/contract/... ./internal/gitenv` → 19패키지 ok
- `go test ./internal/skills/... ./internal/template/...` → exit 0
- `MOAI_GR_BASE=5f5840ae4 go test -v -run TestContractMode ./internal/template/` → PASS 40, FAIL 1(`TestContractModeChangeSetAllowlist/tree`): 남은 6경로는 base 이후 다른 카드 변경 — t1286 codemaps 3, t1237 closure 2, t1225 codexadapter 1. 이 수리 경로는 0.
- `golangci-lint` v2.1.6 → 0 issues; `make agents-emit-check commands-emit-check` → ok
- CI 경로 확인: `MOAI_GR_BASE` 미설정(CI 와 같은 조건)에서 `go test -v -run TestContractModeChangeSetAllowlist ./internal/template/` → `--- SKIP: TestContractModeChangeSetAllowlist/tree` ("MOAI_GR_BASE is not set"), falsifier 하위 테스트만 PASS. 즉 이 base-ref 검사는 CI 에서 돌지 않는다 — 수동 수용 실행에서만 판정되며, base 이후 타 카드 경로로 적색이 되는 것은 F9 부채와 같은 축.
