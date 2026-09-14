# t654 AS-5 계획 개정 감사 보고서 — SPEC-MOAI-GATEWAY-001 v0.13.0

- 감사 대상: `.moai/specs/SPEC-MOAI-GATEWAY-001/` 개정 커밋 `15a3a21f5` (6파일, +326/−7), base `d416f8162` (= origin/develop)
- Iteration: 1/3 (Tier L, `plan_audit_tier_ceilings` 상한)
- **판정: CONDITIONAL PASS** — 종합 점수 0.90 / must-pass 7항목 전수 통과 / 경미 결함 3건이 조건부 수선 대상
- 감사자: plan-auditor (독립 감사, 작성자 추론 문맥 없음 — M1 준수)
- 측정 트리: worktree `.claude/worktrees/t654`, branch `WT-gateway-launchers`

---

## 1. Claim (주장)

SPEC-MOAI-GATEWAY-001 v0.13.0 개정은 카드 t654 AS-5 범위(세 launcher 생산 통합·구독/API 이중 검증·실제 PTY 실증·경로별 context·Windows CI·rc 로컬 배포 게이트)를 REQ-MG-027 / AC-MG-026 / AS-017~022 / A5-M1~M7로 과계수 없이 담고 있으며, must-pass 7기준을 통과한다. 단 3건의 경미 수선 조건(파일 목록 누락, (d) 전제 문구, design 11.1 진입 경로 서술)을 단다.

## 2. Evidence (증거 — 명령 + 관측 출력)

### Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-MG-[0-9]\{3\}' spec.md | sort -u` → REQ-MG-001..027 연속 27개, 공백·중복 0. 폐기 묘비 REQ-MG-007·020은 문서가 명시한 제외 2건이므로 실수 25 = §F·plan §I 선언과 일치. AC도 AC-MG-001..026 연속, 묘비 AC-MG-002 제외 25.
- **[PASS] MP-2 GEARS 형식 (요구사항 계층)** — 신설 REQ-MG-027 (spec.md:840)은 "The launcher shall …" / "When … the gateway shall …" / "Where … the launcher shall …" 복합 GEARS 형태. 기존 REQ-MG-001..026은 이 판에서 미변경(diff 상원 유지). Given-When-Then은 전부 AC 계층(acceptance.md)에만 존재하며 요구사항 계층 오염 없음 — MP-2 판정 대상은 REQ 계층임을 명시한다.
- **[PASS] MP-3 프론트매터** — spec.md:1-15에 12 정규 필드 전수 존재. `version: "0.13.0"` 인용 semver, `status: in-progress`는 8값 enum의 정식 값(spec-frontmatter-schema.md:86 확인), `updated: 2026-09-14` 실제 개정일과 일치, 거부 대상 snake_case 별칭 없음.
- **[N/A] MP-4 언어 중립성** — 단일 언어(Go) 프로젝트 대상 SPEC. 자동 통과.
- **[PASS] MP-5 D7 교차-SPEC 정합** — 참조 형제 4종(CG-RETIRE·PICKER·TEAMMATE·GPT-AUTH) 전부 `.moai/specs/`에 존재하며 `status: draft` — retired/superseded/archived 없음, 정합 의무 미발생. SPEC-MOAI-PROXY-001은 6회 참조되나 스펙 디렉터리에 부재 → D7-5 SHOULD(결함 D5, 경미). BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼** — `grep -n syscall spec.md` → 7건. 그중 spec.md:512가 `internal/cli/launch_exec_posix.go`의 `//go:build !windows` 제약을 명시("공통 파일에 무조건 syscall.Exec를 넣지 않는다"). `grep -cE '//go:build|EXCL.*syscall|cross-platform'` → 2. BLOCKING 없음.
- **[PASS] MP-7 clarification gate** — `grep -n '\[NEEDS CLARIFICATION' plan.md research.md` → 매치 0 (exit 1).

### Category Scores (0.0–1.0)

| 축 | 점수 | 근거 |
|---|---|---|
| 명료성 | 0.85 | REQ-MG-027 단일 해석 가능. (d) 전제 괄호 문구가 이중 독 가능한 1개소(결함 D2), design 11.1 서술 과장 1개소(결함 D3) |
| 완전성 | 0.90 | HISTORY·WHY·WHAT·HOW·REQ·AC·Out of Scope 전부 존재. 신설 Out of Scope — 원격 배포·통합 행위 (spec.md:961)가 §G(908)와 §H(978) 사이 소속 확인. plan A5-M1 예상 변경 목록 1파일 누락(결함 D1) |
| 검증가능성 | 0.90 | AC-MG-026 (a)~(c) 기계 판정 명령 명시, (d) 4단계 명령+출력 증거. AS-017~022 각각 증거 경로·검증 가능성·Gap 조건 명시. "다음 미사용 rc" 발견 명령 미정의(결함 D4) |
| 추적성 | 1.00 | REQ 25건 전수 acceptance.md 헤더 추적 괄호 등장(묘비 2건 제외, grep 계수 1 이상). AC-MG-026 → REQ-MG-027·019. AS-017~022은 기존 AC-MG-004/011/001/010/020/006에 부착 — 고아 AC 없음 |

### research.md §20 코드 상태 주장 재측정 (base `d416f8162` 기준, 본 트리에서 직접 실행)

| §20 주장 | 재측정 명령·관측 | 판정 |
|---|---|---|
| gpt.go launch 대기 오류 게이트 | `internal/cli/gpt.go:54` 리터럴 확인, `:56` `runClaudeEntry(…, "gpt", "gpt", kanban.BackendGPT, launch)` 확인 | 일치 |
| (보완 관측) 동일 리터럴 제2 위치 | `internal/cli/launcher.go:142` (`unifiedLaunchWithGateway`, `mode=="gpt" && binding==nil`) — §20 미기재 | **결함 D1** |
| NewRebaseLedger 비-테스트 호출자 0 | `grep -rn NewRebaseLedger internal/ --include='*.go' \| grep -v _test` → 정의·주석만 (compact.go:42-46) | 일치 |
| receipt/fork API 착지 위치 | core.go:174 `Manifest.ChainTo`, store.go:193 `Store.Rebase`, compact.go:87 `Manifest.Rebase`, family.go:289 `Manager.ForkAt` — 행 번호 전부 정확 | 일치 |
| product binding 공통 capability | gateway_product_binding.go:91 `Capabilities{ContextTokens: 1000000, Images: true, …}`, :94 Anthropic 한정 `RouteID = id + "[1m]"` | 일치 ("88 부근" 표기) |
| ci.yml Go test ubuntu 전용 | :46 `runs-on: ubuntu-latest`; :190 부근 "Moved to release-pr-multi-os.yml's windows leg at release time" 주석; :371-378 test-integration job matrix에 windows-latest 있으나 `integration` 빌드태그 하네스 | 일치 |
| release 워크플로 windows 레그·아티팩트 | release-pr-multi-os.yml:16 `workflow_dispatch`, :204 `go test -json -race -timeout 25m ./…` (integration 태그 없음), :225 `name: test-stream-release-verify-${{ matrix.os }}` | 일치 |
| CHANGELOG 계수 0 | `grep -c 'SPEC-MOAI-GATEWAY-001' CHANGELOG.md` → `0` | 일치 |
| gpt login/logout/status 닫힌 동사 집합 | gpt.go:18-33 확인 | 일치 |
| (보완 판정) 세 launcher 공유 진입 주장 | cc.go:126·gpt.go:56은 `runClaudeEntry`; glm은 glm.go:314 `unifiedLaunch(profileName, "glm", …)` — runGLM → unifiedLaunch → applyGLMMode 경로(glm.go:845 주석이 분기를 자증) | **결함 D3** |

### 지정 판정 쟁점 (의뢰 5항목)

1. **T22 경계 타당성 — 타당하다.** 카드 문구의 "실제 Claude PTY … 재개·모델전환"은 AS-019 (AC-MG-001 부착, REQ-MG-019·015)가 launcher 측(Claude/GLM 재개·`/model` 전환·provider 경계)으로 판정하고, AS-010/011/012의 GPT thread 실세션 양성은 0.12.0(T21)에서 이미 t844로 이관된 기존 경계를 유지한다(plan.md A5-M5가 "이 카드가 흡수하지 않는다"고 명시, T21 행 미수정 확인). 어휘 겹침에 대한 분리 선언이 문서 안에서 이중으로 이뤄져 있어(acceptance.md T22 행 + AS-019 본문 "t844 경계" 문단) 침묵 축소가 아니다. GLM launcher 재개까지 AS-019가 커버하는 것은 카드 문구 대상(Claude PTY)의 상위집합으로 축소가 아니다.
2. **rc.8 스테일 처리 — 수용.** 설치 바이너리는 이미 `v3.2.0-rc.10`(세션 MCP 서버 배너 실측, commit 8050369e5)으로 rc.8은 소비됐을 가능성이 실측상 높다 — 즉 개정이 경계한 상황이 실재한다. "다음 미사용 번호 + 발행 시점 표기와의 차이를 배포 보고서에 명시"(spec.md REQ-MG-027 4문단, plan A5-M7, AC-MG-026 (d) 1.)는 처리가 옳다. 다만 "미사용" 판발 명령이 어디에도 정의되지 않은 것이 결함 D4다.
3. **라이브 계정 Gap이 배포 게이트를 통과하지 않는가 — 통과하지 않는다.** AC-MG-026 말미: "실제 계정 실증(AS-017·AS-019·AS-021)이 세션 환경에서 불가하면 (d)는 창 대기 상태로 남고 PASS로 세지 않는다." 카드 제약 "검증 성공 후 rc 로컬 배포 판정"을 준수한다. 다만 전제 괄호의 "근거를 갖춘 Gap" 허용과 "단 (d) 자체의 전제는 검증 PASS"가 한 절에서 이중 독 가능한 것이 결함 D2다.
4. **appliedEpoch 복원·fork prefix 대조 구현 가능성 — 있다.** (b)는 `Store.Rebase`(store.go:193)가 이미 `transaction(ctx, …)` + `s.write` 원자 쓰기 구조를 갖추고 있어 같은 트랜잭션 epoch 기록이 자연스럽게 접합되고, 판독→`NewRebaseLedger(scope, epoch)` 주입은 시그니처상 그대로 맞는다(호출자 0이므로 시그니처 변경 부담도 없다). RED 전제("고정값 주입 현행 구현이 적색")는 생산 호출자 0 실측으로 성립한다. (c)는 `Manifest.ChainTo(boundary) → []Candidate`와 `Manager.ForkAt`의 chain 발행 구조(family.go:289-333)가 대조 접점을 제공하며, 자식 상태 생성 전 거절 지점 요구는 `forkAtCandidates`의 `os.MkdirAll` 이전 분기로 충족 가능하다. 시험명(TestRebaseLedgerRestoration / TestForkPrefixCrossCheck)이 plan·acceptance·design 3문서에서 일치한다.
5. **예산 25/25 정직성 — 정직하다.** REQ 27 고유 − 묘비 2 = 25, AC 26 − 묘비 1 = 25 (grep 계수 실측). AS-017~022은 새 AC 번호 없이 기존 AC의 하위 시나리오로 부착 — 이는 AS-001~016부터 이어진 이 문서의 확립된 관례이며 DoD(acceptance.md:488)와 §F(spec.md:901)의 24→25 갱신이 함께 이뤄졌다. 숨은 범위 없음.

### 기계 재측정 명령 목록 (baseline 귀속)

본 보고서의 모든 관측은 2026-09-14, worktree `.claude/worktrees/t654` (HEAD `15a3a21f5`)에서 다음 명령들로 직접 실행해 얻은 출력이다: `git show --stat 15a3a21f5`, `grep -rn 'awaiting transport verification' internal/cli --include='*.go'`, `grep -rn NewRebaseLedger internal/ --include='*.go' | grep -v _test`, `sed -n` 행 고정 판독(gpt.go/launcher.go/glm.go/cc.go/gateway_product_binding.go/receipt 3종/family.go), `grep -n runs-on .github/workflows/ci.yml`, release-pr-multi-os.yml `on:` 블록 판독, `grep -c 'SPEC-MOAI-GATEWAY-001' CHANGELOG.md`, REQ/AC/AS ID 전수 sort-uniq 계수, `grep '\[NEEDS CLARIFICATION' plan.md research.md`, 형제 SPEC 4종 `status:` 판독, `.claude/rules/moai/development/spec-frontmatter-schema.md` Status Enum 판독, `.moai/docs/version-management.md` Local RC Numbering + `.claude/rules/local/gitflow-lane-protocol.md` §9 판독.

## 3. Defects Found (구조화 결함 목록)

- **D1** — plan.md:75-78 (A5-M1 예상 변경 목록): 대기 오류 리터럴이 `gpt.go:54` 외에 `launcher.go:142`에도 존재하는데(`unifiedLaunchWithGateway`의 `mode=="gpt" && binding==nil` 분기) 목록에 launcher.go가 없다. AC-MG-026 (a)의 기계 판정이 internal/cli 전체에서 0건을 요구하므로 두 지점 모두 제거해야 GREEN이다 — 목록만 따라가면 판정 명령이 적색으로 낙오되는 재작업 함정. — Severity: minor — Class: optional (AC 기계 판정이 run-phase에서 스스로 잡는다) — Required fix: A5-M1 예상 변경 목록에 `internal/cli/launcher.go`(게이트 분기 제거)를 1행 추가.
- **D2** — acceptance.md AC-MG-026 (d) 전제 절: "PASS(또는 근거를 갖춘 Gap — 단 (d) 자체의 전제는 검증 PASS)" 괄호가 Gap 허용과 전제 PASS 예외를 한 절에 겹쳐 이중 돨 수 있다. 말미 문장(AS-017·019·021 불가 → (d) 창 대기, PASS 아님)이 운용 규칙을 분명히 하므로 카드 제약 위반은 아니다. — Severity: minor — Class: optional — Required fix: 강제 전제 집합을 명명 "(a)~(c) PASS + AS-017·019·021 PASS. Gap 허용은 그 외 항목(예: AS-020 실대형 입력 부분 Gap)으로 한정"으로 상세화.
- **D3** — design.md §11.1 (1011행 부근): "세 launcher는 이미 같은 진입 형태(`runClaudeEntry` → gateway launch plan)를 공유한다"는 측정 부정확 — cc·gpt만 runClaudeEntry이고 glm은 `unifiedLaunch` 직행(glm.go:314, :845 주석 자증)이다. 결합 결론 자체는 바인딩 계층 기준으로 유효하나 코드 상태 단언의 정밀도 결함이다. — Severity: minor — Class: optional — Required fix: 진입 형태 문장을 "cc·gpt는 runClaudeEntry 공유, glm은 unifiedLaunch 경로"로 정정.
- **D4** — acceptance.md AC-MG-026 (d) 1. / plan A5-M7: "다음 미사용 rc"의 발견 명령 미정의(version-management.md는 증분·리셋 정책만 소유). 설치본 `~/go/bin/moai version` 출력이 마지막 컷 후보의 실측 표면이다. — Severity: minor — Class: optional — Required fix: 배포 게이트 1단계에 발견 표면(설치본 version 출력 + 발행 이력 대조)을 명명.
- **D5** — spec.md 등 6개소의 SPEC-MOAI-PROXY-001: `.moai/specs/`에 부재(D7-5 SHOULD). 다만 전부 `.moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md`라는 역사 감사 보고서 경로 참조이며 각각 "옛 경로·옛 식별자 보존" 주석이 붙어 라이브 형제 의존이 아니다. — Severity: info — Class: optional — Required fix: 불요(주석이 이미 처분). D7 SHOULD로 기록만 남긴다.
- **D6** — 공식 기준 문서 URL(https://learn.chatgpt.com/docs/app-server, spec.md 참조·design §11.6): 이 세션 감사 도구셋에 웹 fetch가 없어 권위를 직접 검증 못 했다. — Severity: info — Class: optional — Required fix: 없음(감사 Gaps로 기록; 문서 내 일관성만 확인됨).

## 4. Baseline-attribution (귀속)

모든 증거는 위 2절의 명령 목록대로, 본 턴에서 본 워크트리(HEAD `15a3a21f5`, clean)에 대해 실행해 관측한 출력이다. 타 카드·타 시점 수치의 반입 없음. 참고로 감사 중 호출한 `mcp__moai__audit_multi`(교차모델 수렴)은 claude_verdict anchor 인자를 2회 전달했으나 서버(빌드 8050369e5 — HEAD의 조상, 바이너리 래그 실측됨)가 "anchor missing"으로 거부해 백엔드 판정 0건 — fail-open 설계대로 본 판정은 세션 내 Claude 감사로 확정하며, 교차모델 2차 의견 부재를 잔여 위험에 남긴다.

## 5. Gaps (미검증)

- AS-017~022의 런타임 행위(실제 PTY·실계정·GitHub CI 실행)는 계획 단계 감사 성격상 실행하지 않았다 — 이는 결함이 아니라 판정 대상이 run-phase 증거라는 뜻이다.
- REQ-MG-001..026 기존 본문 전수의 1행 단위 재심은 이번 delta 범위 밖이다(이 판에서 미변경 — diff로 확인). 선행 개정 감사의 판정을 계승한다.
- (b)·(c)의 구현 가능성은 코드 구조 판독에 근거한 판정이지 프로토타입 검증이 아니다.
- 외부 URL 권위 미검증(D6). 교차모델 백엔드 판정 미수합(Baseline-attribution 참조).

## 6. Residual-risk (잔여 위험)

- D1을 수선하지 않은 채 run에 들어가면 A5-M1 첫 GREEN 시도가 판정 명령에 낙오된다(낭비 1사이클, 판정 자체는 안전).
- (d) 전제 문구(D2)를 운영자가 관대하게 읽으면 검증 미완 상태 배포가 이론상 가능하나, 말미 문장과 카드 [HARD] 제약이 이중으로 막는다.
- t844(라이브 계측)가 착지하기 전까지 본 SPEC의 전체 기능 통과는 T21대로 보류 상태다 — 이는 0.12.0에서 유지된 기존 경계이며 이 판이 만든 위험이 아니다.
- 교차모델 2차 의견 없이 단일 모델 감사로 확정했다.

## 7. Recommendation

CONDITIONAL PASS. 세 가지 조건 정비(D1·D2·D3 — 각 1행 내외 문서 수선)는 다음 개정이나 run 시작 전 리뷰에서 반영하면 충분하며, plan-phase 반복 루프를 요구하는 등급이 아니다. D4·D5는 권고. run-phase 착수 시 A5-M1에서 launcher.go:142 제2 게이트 분기 제거가 반드시 A5-M1 범위에 포함되는지를 배차 지시가 명시하도록 권한다.
