auditor-model: claude-opus-5-5[1m]

# t1236 sync 감사 보고서 — SPEC-AUTONOMY-GATE-REWIRE-001

- 감사 대상: 워크트리 `.claude/worktrees/t1236`, 브랜치 `WT-contract-gate-rewire`, HEAD `1f0a99279`
- 변경 범위: `CARD_BASE=$(git merge-base develop HEAD)` → `7fe658815eb0d4110b9acadad56e5a85bee3ed3f`, `git diff --name-only 7fe658815..HEAD` 84개 파일
- 평가 프로필: `.moai/config/evaluator-profiles/default.md` (SPEC 에 `evaluator_profile` 없음, harness `default_profile: "default"`), 평면 가중 모드
- 적용 규칙: `verification-claim-integrity.md` §1·§2·§3, `verification-completeness.md` §1.1·§1.3

## 판정

**Overall Verdict: FAIL (델타 한정)** — 필수 통과 차원(Functionality·Security)은 모두 통과했다. FAIL 사유는 blocking 결함 F1 한 건이다. CHANGELOG 가 새 SSOT 를 "always-loaded" 라고 적었지만, 실제 파일은 `paths:` 로 범위가 한정돼 있다. 한 단어만 고치면 되며, 재감사 범위는 F1 델타뿐이다. 이 결함이 고쳐지면 부채를 안은 채 통과(PASS-WITH-DEBT)하는 조건이 갖춰진다.

### 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | AC 25/25. 이번 감사에서 다시 쟀다: `internal/contract/...` 6개 패키지 `ok`(RUN 693 · FAIL 0 · SKIP 0), `internal/cli -run '^TestContract'` `ok`(최상위 PASS 3 · 하위 포함 12), `internal/template -run TestContractMode`(`MOAI_GR_BASE=7fe658815`) PASS 15 · SKIP 0, `-run TestJev` PASS 1(하위 7), `internal/spec` `ok … 112.426s`. 변이 5개 중 4개가 제거됐다(아래) |
| Security (25%) | 85/100 | PASS | Critical·High 없음. Medium 3건(F3·F4·F5)과 Low 1건(F6)은 모두 자율 Kickoff 가 꺼져 있어(`autonomousKickoffEnabled = false`) 도달하지 않거나, BASE 에 이미 있던 결함이다 |
| Craft (20%) | 72/100 | FAIL (필수 통과 차원 아님) | `go test -cover`: contract 96.3% · kickoff 87.4% · revoke 86.3% · sign 90.1% · **receipt 68.1%**(패키지 교차 측정 `-coverpkg` 로도 79.5%). 체인 prev 링크 검사 변이가 살아남았다(F2). lint 는 `0 issues.`, vet 은 exit 0 |
| Consistency (15%) | 78/100 | PASS | 로컬과 템플릿의 추가 hunk 가 바이트 동일하다(47줄, 삭제 0). 4개 로케일 구조가 같다. CHANGELOG 오기(F1), verdict 증거 포인터 오류(F7), design §11.1 승인 표기 낡음(F8) |

가중 조화평균: 1 / (0.40/90 + 0.25/85 + 0.20/72 + 0.15/78) = **82.7**

## 기계 검증 — 원문 증거 (이 실행, 이 트리 `1f0a99279`)

| 명령 | 종료 코드 | 출력(발췌 원문) |
|---|---|---|
| `golangci-lint version` | 0 | `golangci-lint has version v2.1.6 built with go1.26.8` |
| `golangci-lint run ./internal/contract/... ./internal/cli/ ./internal/template/` | 0 | `0 issues.` |
| `go vet ./internal/contract/... ./internal/cli/ ./internal/template/` | 0 | (출력 없음) |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 | (출력 없음) |
| `MOAI_GR_BASE=7fe658815 go test ./internal/contract/... -count=1 -v` | 0 | `ok …/internal/contract 0.615s` · `ok …/kickoff 13.233s` · `ok …/receipt 2.138s` · `ok …/revoke 4.225s` · `ok …/sign 13.337s` · `ok …/sign/signtest 0.464s` |
| `MOAI_GR_BASE=7fe658815 go test ./internal/cli/ -run '^TestContract' -count=1 -v` | 0 | `--- PASS: TestContractDecide (2.97s)` · `--- PASS: TestContractKickoffCheck (0.55s)` · `--- PASS: TestContractRevoke (0.26s)` · `ok …/internal/cli 4.630s` |
| `MOAI_GR_BASE=7fe658815 go test ./internal/template/ -run TestContractMode -count=1 -v` | 0 | 최상위 `--- PASS` 15개, `--- SKIP` 0개, `ok …/internal/template 10.635s` |
| `MOAI_GR_BASE=7fe658815 go test ./internal/template/ -run TestJev -count=1 -v` | 0 | `--- PASS: TestJevDoctrineAmendment (0.00s)` · `ok … 0.176s` |
| `MOAI_GR_BASE=7fe658815 go test ./internal/template/ -count=1` (중립성·누출 가드 포함 전체) | 0 | `ok …/internal/template 73.580s` |
| `go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak$' -v` / `'TestTemplateNeutralityAudit$'` | 0 / 0 | `--- PASS: TestTemplateNoInternalContentLeak (0.65s)` / `--- PASS: TestTemplateNeutralityAudit (0.00s)` |
| `MOAI_GR_BASE=7fe658815 go test ./internal/spec/ -count=1` | 0 | `ok …/internal/spec 112.426s` |
| `go test ./internal/hook/ -run ContractSign -count=1 -v` (A2b 가드, 교차 패키지 확인) | 0 | PASS 42 · FAIL 0 · SKIP 0, `ok …/internal/hook 0.769s` |
| `make build` | 0 | `catalog.yaml updated successfully (13408 bytes)` → 이후 `git status --short` 출력 없음(카탈로그 해시가 커밋본과 일치) |
| `make agents-emit-check` | 0 | `ok …/internal/template/agentemit 0.497s` |
| `go run ./cmd/moai spec lint SPEC-AUTONOMY-GATE-REWIRE-001` / `SPEC-JEV-CORE-001` | 0 / 0 | `✓ No findings — all SPEC documents are valid` |
| `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` | 0 | `baseline: OK — .moai/spec-lint-baseline.json` |

선택이 비어 있지 않다는 확인: 위 네 선택 모두 `[no tests to run]` 가 0건이고, RUN 수는 contract 693 · cli 12 · template ContractMode 41 · Jev 7 이다.

### 변이 탐침 (임시 편집 → 관측 → `git checkout -- <그 파일>` 로 원복)

| # | 파일 | 변이 | 결과 |
|---|---|---|---|
| M1 | `internal/contract/kickoff/check.go:137` | revoke 판독기 호출을 `false, error(nil)` 로 바꿔 건너뜀 | **제거됨** — `check_test.go:147: pass=true reason="" reasons=[] state=signed-valid, want pass=false reason="revoked"` / `--- FAIL: TestKickoffCheck/17_revoke_record_only` |
| M2 | `internal/contract/revoke/revoke.go:86` | `cmd.Env = gitenv.Env()` 제거 | **제거됨** — `failures_test.go:215: record head "1d0c0a65…", want the project's HEAD "20d98fe5…" (ambient GIT_DIR HEAD is "1d0c0a65…")` / `--- FAIL: TestRevokeHeadIgnoresAmbientGitEnv` |
| M3 | `internal/contract/kickoff/kickoff.go:28` | `JevDoctrineAmended = true` → `false` | **제거됨** — `activation_test.go:250: partial amendment: present in .moai/specs/SPEC-JEV-CORE-001/spec.md, … CLAUDE.local.md; absent in internal/contract/kickoff/kickoff.go` / `--- FAIL: TestJevAmendmentLinkage/tree`. 참고: 같은 변이에서 `internal/cli -run TestContractDecide` 는 `ok` 였다. CLI 층은 상수의 실제 값을 검사하지 않고, 연동 가드만 이 변이를 잡는다 |
| M4 | `internal/contract/receipt/store.go:147` | 체인 `prev` 불일치 검사를 `false && …` 로 무력화 | **생존** — `go test ./internal/contract/...` 가 모두 `ok`. 원본 코드는 옳게 동작한다: 임시 탐침 테스트로 가운데 줄을 지우면 원본은 `events.jsonl line 2: prev does not match the preceding line` 을 내고, 변이에서는 `<nil>` 이 나온다. 이 동작을 지키는 회귀 테스트가 없다(F2). 탐침 파일은 지웠다 |
| M5 | `internal/contract/kickoff/decide.go:147` | 작성자 배제 조건에서 트레일러 집합 포함 검사 제거 | **제거됨** — `decide_test.go:411: outcome "approve" reason "", want human "author-decider-conflict"` / `--- FAIL: TestDecideAuthorExclusion/trailer_author` |

원복 확인: `git status --short` 출력 없음, `git rev-parse --short HEAD` → `1f0a99279`.

## Findings (structured defect-list)

- **F1** [Low 심각도, 확신 높음] **[blocking]** `CHANGELOG.md:12` — 항목이 "A new **always-loaded** SSOT `.claude/rules/moai/workflow/contract-autonomy.md`" 라고 적었다. 실제 파일 머리에는 `paths: "**/.moai/specs/**/contract.yaml,**/.moai/specs/**/kickoff-receipt.json,**/contract-autonomy.md"` 가 있어 조건부로 로드되고, `design.md:73` 과 본문 결정 D-6(상시 로드 대상은 블록 세 개)도 그렇게 정했다. 매 세션 문맥 비용에 대해 사용자를 잘못 안내하는 문장이다. — 필수 수정: "always-loaded" 를 "path-scoped (loaded when a `contract.yaml` / `kickoff-receipt.json` is in play)" 식으로 바로잡고, `grep -n 'always-loaded SSOT' CHANGELOG.md` 가 0행임을 보인다. docs-site 네 로케일에는 같은 표현이 없다(확인함).
- **F2** [Medium, 확신 높음] [optional] `internal/contract/receipt/store.go:147` — 가운데 줄 삭제·재배열을 잡는 유일한 검사(`prev` 링크)에 회귀 테스트가 없다(변이 M4 생존). `TestEventStore` 의 "(ii) edit a middle line" 은 `at` 필드를 바꾸므로 해시 검사가 먼저 잡는다. REQ-GR-012 가 말하는 변조 흔적의 핵심인 체인 결합이 무방비로 남아 있다. — 권장 수정: `TestEventStore` 에 가운데 줄 삭제 하위 테스트를 추가해 `ChainError{Line: 2, Why: "prev does not match…"}` 를 단언하고, M4 변이에서 FAIL 이 나는지 관측한다. receipt 커버리지 부채(68.1%)도 함께 줄어든다.
- **F3** [Medium, 확신 높음, 휴면] [optional] `internal/contract/kickoff/decide.go:226-247` — 결정 경로가 `AppendReceipt` → `AppendEvent` → `os.WriteFile` 순서로 세 번 따로 쓴다. 중간에서 실패하면 앞 단계의 기록이 남는데도 오류 exit 가 나고, REQ-GR-011 의 "아무것도 쓰지 않으면 exit 1/2" 전제와 어긋난다. 또 `AppendReceipt`·`AppendEvent` 오류를 `%w: %v` 로 `ErrUsage` 에 감싸서, 락 획득 뒤 드러난 체인 손상(`ErrIntegrity`)이 exit 1 이 아니라 exit 2 로 나간다. 호출 초입의 `st.Verify()` 가 대부분을 먼저 거르므로 현실에서 겪을 경로는 좁다. — 권장 수정: 오류 감싸기에서 `%w` 를 두 번 써서 `ErrIntegrity` 를 보존하고, 부분 기록을 결과에 명시하거나 잔여 위험으로 문서화한다.
- **F4** [Medium, 확신 중간, 휴면] [optional] `internal/contract/kickoff/check.go:146-160`, `receipt/events.go:127` — kickoff-check 는 영수증이 `receipts.jsonl` 에 있는지만 보고, 그 영수증 줄 해시를 가리키는 `decide` 사건이 있는지는 보지 않는다(codex 제기, 코드로 확인). 체인에는 키가 없어 같은 사용자 권한의 작성자는 체인이 맞는 영수증 줄을 덧붙일 수 있다. 설계가 "방지가 아니라 흔적"(REQ-GR-012, design D21)이라고 명시한 범위 안의 문제이고, 자율 Kickoff 가 꺼져 있어 지금은 도달하지 않는다. — 권장: 활성화(M8) 전에 영수증↔decide 사건 결합 검사를 활성 조건 표에 추가할지 결정한다.
- **F5** [Medium, 확신 높음, BASE 부터 있던 결함] [optional] `internal/contract/sign/defaults.go:94-106` — 서명기의 `gitEnv()` 는 대문자 `GIT_DIR`·`GIT_WORK_TREE`·`GIT_INDEX_FILE` 세 개만 지운다. 중앙 헬퍼 `gitenv.Env()` 는 11개 저장소 범위 변수를 지우고 Windows 대소문자 접기도 처리한다. `git show 7fe658815:internal/contract/sign/defaults.go` 에 같은 코드가 있으므로 이 카드가 만든 결함은 아니다. 다만 progress §E.2 의 "형제 스윕: 모든 `exec.Command` 가 정리된 환경으로 실행된다" 는 문장은 범위가 좁은 이 스크럽까지 같은 것으로 셌다. — 권장: 후속 카드에서 `gitenv.Env()` 로 통일한다.
- **F6** [Low, 확신 높음] [optional] `internal/contract/kickoff/check.go:50`, `decide.go:78` — revoke 는 `contract.ValidCard` 로 카드 id 를 검사하지만(`revoke.go:113`), kickoff-check 와 decide 는 비어 있지 않은지만 본다. kickoff-check 에서는 카드가 계약과 달라도 `judgeSignature` → `revokeReader(root, card, …)` 가 돌기 때문에, `../` 가 섞인 카드 id 로 `.moai/reports/<card>/escalation` 밖의 `*.md` 를 읽는 경로가 열린다. 읽기 전용이고 출력은 사유 코드뿐이라 정보 노출은 없다. — 권장: 두 입구에서 `contract.ValidCard` 를 적용한다.
- **F7** [Low, 확신 높음] [optional] `.moai/reports/t1236/verdict.md:194-198` — 「sync 검증 (이번 실행)」 절은 명령만 적고 "결과는 커밋 메시지와 manager-docs 보고에 원문으로 남긴다" 고 했다. 그러나 `git log -1 --format=%B c902d4ffd` 에는 그 출력이 없다(VCI §2 귀속 포인터가 가리키는 곳에 증거가 없음). 이 감사가 같은 세 명령을 다시 재서 모두 exit 0 을 관측했으므로 주장 자체는 참이다. — 권장: verdict 에 이 보고서의 증거 표를 인용하거나 원문 출력을 붙인다.
- **F8** [Low, 확신 중간] [optional] `design.md:292` 는 여전히 "운영자 승인 2026-09-26 (리드 중계, 레인 재확인 전)" 이다. progress §E.2:135 는 레인 세션 재확인(2026-09-27)을, verdict:178 은 AskUserQuestion 선택지 「넣기 (권장)」 를 적었다. 세 문서의 기술이 어긋난다. 또 `TestJevDoctrineAmendment/local-guide` 의 운영자 확인 검사는 progress.md 에 `operator confirmed the §29 line` 문자열이 있는지만 본다(문자열 존재 = 동의로 취급). — 권장: manager-spec 이 §11.1 에 재확인 사실을 한 줄로 추가한다. 테스트의 한계는 잔여 위험으로 둔다.
- **F9** [Low, 확신 높음] [optional] `internal/template/contract_mode_guided_test.go:32-38` — BASE 참조 가드 6개(`GuidedPreservation/tree`, `ChangeSetAllowlist/tree`, `InheritedDivergence`, `ConstitutionDriftNotIncreased`, `AlwaysLoadedBudget`, `EmitterSites`)는 `MOAI_GR_BASE` 가 없으면 `t.Skip` 한다. 환경 변수 없이 실행하면 `--- SKIP` 6건을 관측했다. CI 와 병합 이후에는 늘 건너뛰므로, 상시 로드 합계 1,500자 상한 같은 규칙은 병합 뒤 지켜 줄 장치가 없다(verification-completeness §1.3 지속 발화). 건너뛰기가 `--- SKIP` 로 드러나 조용하지는 않다. 블록당 상한(500/900)은 환경 변수 없이도 `TestContractModeBlocksWellFormed` 가 검사한다. — 권장: 설계가 의도한 동작이라면 그렇다는 사실을 SSOT 나 테스트 주석에 적는다.
- **F10** [Info] 문서 서술 — docs-site 네 로케일의 revoke 설명 "진행 중인 run 은 다음 단계 경계에서 멈춥니다" 는 contract 모드 오케스트레이터 절차(D-9)를 전제로 한다. guided 모드나 자율 Kickoff 가 꺼진 현재 상태에서는 조건부 사실이다. 조치할 필요는 없다.

## 확인 항목별 결과

| 항목 | 결과 | 근거 |
|---|---|---|
| Template-First 대칭 | PASS | 로컬 변경 16개 경로에 모두 템플릿 사본이 있다. 7개는 `cmp` 로 동일하고, 9개(`orchestration-mode-selection.md`, `SKILL.md`, `moai.md`, `plan.md`, `run.md`, `sync.md`, `delivery.md`, `doc-execution.md`, `workflow.yaml`)는 BASE 부터 갈라져 있던 파일이다. 이번 변경분만 따로 떼어 비교하면 로컬 47줄과 템플릿 47줄이 `diff` 무출력으로 같고 삭제는 0줄이다. 템플릿에만 있는 변경은 0건 |
| 중립성·누출 가드 | PASS | `internal/template` 전체 `ok 73.580s`, 개별 `--- PASS: TestTemplateNoInternalContentLeak`, `--- PASS: TestTemplateNeutralityAudit` |
| `make agents-emit-check` | 해당 없음(실행해서 통과) | 에이전트 파일 변경 0건(`git diff --name-only … -- .claude/agents internal/template/templates/.claude/agents` 0행; 같은 형태를 `internal/cli` 에 걸면 행이 나와 대조군이 성립한다). 검사는 exit 0 |
| docs-site 4개 로케일 | PASS | ko/en/ja/zh 모두 `##` 4개, `moai contract` 표 행 6개, callout 4개, Mermaid LR 0개. 문장은 각 언어 원어 문체이고 종료 코드 표기도 서로 맞다 |
| CHANGELOG 정확성 | **FAIL (F1)** | 나머지 주장은 모두 확인했다: 블록 20개·문서 13개(`grep -ro 'moai:contract-mode-start'` 합계 20), AC 25(`AC-GR-0NN` 고유 25, `AC-CONTRACT-016` 은 `[REF]` 표지), 서명 사건 기록이 파일 쓰기보다 먼저 일어남(`sign.go:250` `recordEvents` → `:253` `write`), `autonomousKickoffEnabled = false`(`kickoff.go:19`) |
| 부채 기록의 정직성 | PASS | receipt 68.1%(재측정 일치) · BASE 파생 산출물 · M8 보류 · M8 4행 push 직렬화 미재실행 · 헌법 DRIFT 9(가드 `ConstitutionDriftNotIncreased` PASS 로 증가 없음 확인)가 모두 verdict 「미검증」에 있다. 교차 플랫폼 빌드 미재실행 항목은 이번 감사가 HEAD 에서 쟀고 통과했다 |
| `CLAUDE.local.md §29` 문안 | PASS | 추가된 한 줄과 `design.md` §11.1 마커 사이 문안이 `cmp` 로 바이트 동일(`BYTE-IDENTICAL`). 위치는 「3등급 항목에서는 Jev 를 호출하지 않는다…」 문단 바로 뒤다 |
| doctor·hook 생성자 변경 여부 | 변경 없음 | `internal/hook`, `internal/cli/doctor*` 변경 0행. 교차 패키지 안전 확인으로 A2b 가드(`-run ContractSign`)를 돌려 PASS 42 를 관측했다. `TestBinaryLag` 는 해당 없음 |
| 금지 경로(REQ-GR-024) | PASS | 변경 목록 84개에서 `moai-constitution`, `zone-registry`, `/agents/`, `output-styles`, `agent-common-protocol`, `spec-workflow`, `internal/escalation` 등 0건 |
| 교차 모델 감사 | 불확정 | `mcp__moai__audit_multi`(project_root = 이 워크트리): claude `CLAUDE_DIFF_UNAVAILABLE`, glm `unknown review target`, codex 는 본문에 FAIL 소견 3건을 냈지만 판정 파싱 결과는 `inconclusive` 였다. codex 소견은 코드로 직접 대조해 F4·F5 와 F3 일부로 반영했고, "High" 등급은 휴면 상태와 설계가 명시한 잔여 위험을 근거로 Medium 으로 낮췄다 |

## 미검증 (Gaps)

- M8 4행의 push 직렬화 쪽 테스트는 이번에도 돌리지 않았다(sign/decide 가드 쪽 42건만 쟀다).
- 전체 테스트 스위트(`go test ./...`)는 돌리지 않았다. 판정은 CI 몫이다.
- 운영자의 §29 재확인은 이 감사가 관측할 수 없다. progress·verdict 의 기록만 읽었다.
- `MOAI_GR_BASE` 를 흡수 후 merge-base 로 바꾼 상태의 가드 결과는 재지 않았다(BASE 파생 부채).
- 교차 모델 감사에서 claude·glm 백엔드는 판정을 내지 못했다.

## 잔여 위험

- 자율 Kickoff 경로는 한 번도 켜진 적이 없다. 활성화(M8) 시점에는 F3·F4 가 실제로 도달하는 경로가 된다.
- 체인에 키가 없어서, 같은 사용자 권한을 가진 작성자는 체인이 맞는 줄을 덧붙이거나 파일 전체를 다시 쓸 수 있다(설계가 받아들인 위험).
- BASE 참조 가드는 병합 뒤 늘 SKIP 이라서, 상시 로드 1,500자 예산은 이후 변경에서 지켜지지 않는다(F9).
- `.moai/reports/t1236/sync-audit.md`(이 파일)는 커밋하지 않은 새 파일이며 `.gitignore:235` `.moai/reports/*` 규칙에 걸린다(`git check-ignore -v` 로 확인). 그래서 감사 종료 시점의 `git status --short` 는 출력이 없다. 트래킹하려면 `git add -f` 로 추가해야 한다(다른 `plan-audit-*.md`·`verdict.md` 가 그렇게 들어갔다).

## 권고 (처리 순서)

1. F1 수정(blocking) → 델타 재감사: `grep -n 'always-loaded SSOT' CHANGELOG.md` 0행 + CHANGELOG 항목 재독.
2. F2 회귀 테스트 추가(선택, 권장도 높음) — receipt 커버리지 부채와 변이 생존을 함께 해소한다.
3. F3·F4·F6 은 M8 활성화 전 선행 과제로 카드화할지 리드가 정한다. F5 는 별도 후속 카드 후보다.
4. F7·F8 은 문서 정합 수정(선택)이다.
