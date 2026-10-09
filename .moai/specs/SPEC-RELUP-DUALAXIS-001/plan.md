---
id: SPEC-RELUP-DUALAXIS-001
title: "plan — release-update 하네스 CC+Codex 이중 축 정착"
created: 2026-10-09
author: manager-spec
tier: M
---

# plan: SPEC-RELUP-DUALAXIS-001

## §A Context

- **측정 트리**: `.moai/worktrees/t1579` (branch `WT-high-10-07`) @ `2aab5f797`. RED-now grep 앵커는 이 SHA에서 본 수리로 재실행해 확인했다(acceptance.md §D.3-d RED-summary). 검증 동사 2종의 RED는 M2 이전 관측이며 귀속 한계는 §D.3-d Gaps에 적었다. GREEN 관측은 측정 시점 HEAD(d36e97571)에 귀속한다.
- **카드**: t1579 (High·운영자 확장 지시 2026-10-07·builder-harness/SPEC 소관).
- **SPEC artifacts**: `.moai/specs/SPEC-RELUP-DUALAXIS-001/{spec,plan,acceptance,progress}.md` — Tier M 3-artifact 세트 + progress.md.
- **변경 표면 (3개 — 전부 사용자 소유 dev-only 네임스페이스)**:
  1. `.claude/agents/harness/hns-release-update-specialist.md` (280행 — codex 축 Phase + 상태 스키마 문서화 + BP 상시 섹션 + Phase 3 URL 세트)
  2. `.claude/workflows/hns-release-update-run.js` (175행 — codex 렌즈 fan-out + 커밋 복원 폴백 + 6테마 체크리스트)
  3. `.claude/commands/harness/release-update/manifest.json` (29행 — domain 문자열만)
- **Go 코드 변경: 0.** 템플릿 미러: 없음(불요 — §1.1 M9). `make build` 불요.
- **근거 연구**: `.moai/research/upstream-update-20261007.md` (1차 — CARD-1 정의 + 6테마 표 + Phase 7.5 findings 2건) / `upstream-update-20261008.md` (2차 — codex 0.161.0 승격 큐레이팅 + URL 세트 finding).
- **선례 SPEC**: SPEC-UPDATE-ADD-CODEX-001 (completed) — codex 배선 선례로 인용, 스코프 중복 없음(그쪽은 `moai update --add-codex` CLI 동사).

### §A.5 PRESERVE 목록 (건드리지 않는다)

| 표면 | 보존 이유 |
|---|---|
| `.moai/state/last-cc-version.json` (primary 체크아웃, 기계 로컬) | CC 축 상태 파일 — 스키마·내용 불변. run-phase가 편집 금지 |
| manifest.json `source_request` 필드 | 3-harness 분리의 역사 서술(프로비넌스) — 갱신 대상 아님 |
| manifest.json `sprint_contract.dimensions` + thresholds | 결정 D4 — Functionality/Consistency 0.85/0.80 불변 |
| manifest.json `entry_command` / `runner_workflow` / `specialists[]` (primitive·isolation) | 하네스 골격 불변 — `hns-release-update-run.js` 참조 1힛 유지 (AC-RDX-013) |
| 스페셜리스트 Phase 5 (인간 게이트) · Phase 6 (docs 4-locale) · Phase 7b (manager-git PR) · Phase 7.5 (REQ-HRR-006 findings) · Phase 8 | 기존 절차 전부 유지 — codex/BP 확장은 추가다 |
| Runner HARD 제약 (AskUserQuestion·gh pr 금지, Date.now()/Math.random() 금지, top-level 실행 + CommonJS export 가드) | AC-DHC-007a + 결정성 계약 |
| `.moai/research/upstream-update-*.md` 2건 | 읽기전용 연구 산출물 — 편집 금지 |
| 타 SPEC 디렉터리 · `.moai/reports/` | 스코프 밖 (B10) |

## §B Known Issues (관련 카테고리만)

- **B4 Frontmatter 스키마**: spec.md 12 필드 canonical (`created:`/`updated:`/`tags:` — snake_case 금지). 본 plan/acceptance는 status 축 stateless(§ Artifact Statelessness).
- **B6 Out of Scope 헤딩**: `## Out of Scope` h2 단독은 `MissingExclusions` lint ERROR — `### Out of Scope — <topic>` h3 + `-` 불릿으로 작성했다(spec.md §6).
- **B8 워킹 트리 위생**: 런타임 관리 파일(`.moai/harness/`, `.moai/state/`) 편집 금지. 커밋은 지정 pathspec만.
- **B10 범위 규율**: §A.5 PRESERVE 이외 무변경. 특히 Go 트리(`internal/`) 0변경 — AC-RDX-011이 지키는 회귀 가드다.
- **B11 사용자 질의 금지**: leaf worker — 열린 질문은 전부 spec.md §1.2 결정 기록으로 봉쇄했다. `[NEEDS CLARIFICATION]` 마커 0개.
- **상태 파일 특이사항**: `last-codex-version.json`은 gitignored 기계 로컬이라 CI가 판정할 수 없다 — AC는 본문 쓰기 지점을 측정면으로 삼는다(§5.2).
- **node --check 한계**: 러너 JS 파스 검증에 `node --check`는 무음 통과 한계가 있다(운영 교훈) — §E에서 CommonJS require() 스모크로 보강한다.

## §C Pre-flight (run-phase 진입 시 재실행)

```bash
git branch --show-current ; git rev-parse --short HEAD     # WT-high-10-07 이후 재확인
# RED-now 앵커 15종 재측정 (acceptance.md §D.3 원장의 명령 그대로 — 전부 단일 호출; 검증 동사 2종은 §E3-P3·P4; 육면 셀 6종(67–72행)은 AC-RDX-009 쌍; 회귀 가드 3종은 74–76행):
grep -c '"domain".*Codex CLI upstream change tracking' .claude/commands/harness/release-update/manifest.json   # 기대 0 (M4 전) — domain 필드 스코프 (CX-3)
grep -c '"domain".*best-practices axis' .claude/commands/harness/release-update/manifest.json                   # 기대 0 (M4 전) — domain 필드 스코프 (CX-3)
grep -c "selectCodexSweepTargets(args)" .claude/workflows/hns-release-update-run.js                              # 기대 0 (M2 전) — 착지 후 ≥2: 정의+top-level 디스패치 호출 (CX-2)
grep -c "CODEX_COMMITS_FALLBACK" .claude/workflows/hns-release-update-run.js                           # 기대 0 (M2 전)
grep -c "CODEX_THEME_CHECKLIST" .claude/workflows/hns-release-update-run.js                            # 기대 0 (M2 전)
grep -c "last-codex-version.json" .claude/agents/harness/hns-release-update-specialist.md              # 기대 0 (M1 전)
grep -c "rust-v0.161.0" .claude/agents/harness/hns-release-update-specialist.md                        # 기대 0 (M1 전)
grep -ci "best-practice" .claude/agents/harness/hns-release-update-specialist.md                       # 기대 0 (M3 전)
grep -c "code.claude.com" .claude/agents/harness/hns-release-update-specialist.md                      # 기대 0 (M3 전)
grep -c "HTML proposal report" .claude/agents/harness/hns-release-update-specialist.md                 # 기대 0 (M3 전)
grep -c "source-first" .claude/agents/harness/hns-release-update-specialist.md                          # 기대 0 (M3 전)
grep -c "7a-codex" .claude/agents/harness/hns-release-update-specialist.md                              # 기대 0 (M1 전) — Phase 7a 기록 단계 (CX-6)
grep -c "only the CC axis" .claude/agents/harness/hns-release-update-specialist.md                      # 기대 0 (M1 전) — Phase 2 축별 종료 (CX-9)
grep -c 'If no entries: emit "No new versions since vX.Y.Z" and stop' .claude/agents/harness/hns-release-update-specialist.md  # 기대 1 (M1 전) — 착지 후 0이 PASS (제거면, CX-10)
grep -c "docs.anthropic.com" .claude/agents/harness/hns-release-update-specialist.md                    # 기대 ≥1 (M3 전) — 착지 후 0이 PASS (제거면, CX-14)
git grep -c -h "docs.anthropic.com" 2aab5f797b75983e132af451da68f69e3426557b -- .claude/agents/harness/hns-release-update-specialist.md   # 핀 RED (B-01): 기대 6 / exit 0 — 2aab5f797 귀속
grep -cE 'https://code\.claude\.com/docs/en/hooks([^A-Za-z0-9_./-]|$)' .claude/agents/harness/hns-release-update-specialist.md   # 육면 셀 (B-03·B-04): 핀 2aab5f797 기대 0 (RED) · 착지 기대 1 (GREEN)
grep -cE 'https://code\.claude\.com/docs/en/sub-agents([^A-Za-z0-9_./-]|$)' .claude/agents/harness/hns-release-update-specialist.md   # 육면 셀 — 동일 패턴
grep -cE 'https://code\.claude\.com/docs/en/skills([^A-Za-z0-9_./-]|$)' .claude/agents/harness/hns-release-update-specialist.md   # 육면 셀 — 동일 패턴
grep -cE 'https://code\.claude\.com/docs/en/plugins([^A-Za-z0-9_./-]|$)' .claude/agents/harness/hns-release-update-specialist.md   # 육면 셀 — 동일 패턴
grep -cE 'https://code\.claude\.com/docs/en/mcp([^A-Za-z0-9_./-]|$)' .claude/agents/harness/hns-release-update-specialist.md   # 육면 셀 — 동일 패턴
grep -cE 'https://code\.claude\.com/docs/en/settings([^A-Za-z0-9_./-]|$)' .claude/agents/harness/hns-release-update-specialist.md   # 육면 셀 — 동일 패턴
# PRESERVE 앵커 3종:
grep -rn "last-codex-version" internal/   # 0힛 유지 (AC-RDX-011)
grep -c "last-cc-version.json" .claude/agents/harness/hns-release-update-specialist.md                 # ≥3 유지 (AC-RDX-012)
grep -c "hns-release-update-run.js" .claude/commands/harness/release-update/manifest.json              # 1 유지 (AC-RDX-013)
```

**시드값 재판정 (M1 착지 직전 — last-analyzed 의미론, plan-audit iter1 CX-1)**: `npm view @openai/codex version` + `gh api repos/openai/codex/releases?per_page=5` 재실행. 시드의 의미론은 **"마지막 분석 버전(last-analyzed)"**이다. 최신 비프리릴리즈가 `rust-v0.161.0`보다 새로워도 **그 델타가 아직 분석되지 않았다면 시드는 `rust-v0.161.0`에 고정**되고, 신규 안정 승격은 "다음 스윕의 분석 대상"으로 기록된다 — 시드를 미분석 버전으로 올리면 그 델타는 greater-than 필터에 영영 스킵된다. 시드를 새 값으로 올릴 수 있는 유일한 조건은 그 버전까지의 델타가 실제로 분석·큐레이팅된 경우뿐이며, 그때만 spec.md D1·REQ-RDX-002·AC-RDX-007을 동시 갱신한다(§3 계층 수정 규율).

## §D Constraints (앵커 고정 — run-phase 재량 금지)

### §D1 고정 앵커 (grep 판정면 — spec.md §1.2·§4와 쌍)

| 표면 | 고정 앵커 (리터럴) | AC |
|---|---|---|
| manifest.json `domain` | 필드 스코프 패턴 `'"domain".*Codex CLI upstream change tracking'` — domain 키 행만 매치, source_request 동 문구 불매 (CX-3 재앵커) | AC-RDX-001 |
| manifest.json `domain` | 필드 스코프 패턴 `'"domain".*best-practices axis'` — 동일 스코프 | AC-RDX-002 |
| runner | `selectCodexSweepTargets(args)` 출현 **≥2** — 제1 출현=정의, 제2 출현=top-level 런타임 블록 병합 지점. 병합 형태 고정: `const allTargets = ccTargets.concat(codexTargets);` + `parallel(allTargets` 호출식. codex 렌즈 라벨 접두사 `codex-release-notes:`(CC의 `cc-release-notes:`와 병렬 — E3-P3 판정 토큰). 정적 면 LED-003 + 동적 면 LED-016(§E3-P3 모의-런타임 — codex 라벨 agent 호출 실측). `module.exports`에 `selectCodexSweepTargets` 추가 필요 | AC-RDX-003 |
| runner | 상수 `CODEX_COMMITS_FALLBACK` — 본문 비어 있을 때의 커밋 API 복원 절차 문서 블록 앵커. 절차 내용: (1) 릴리즈 본문 1줄 제목만 관측되면 `gh api repos/openai/codex/commits`/`pulls` 주제 복원, (2) 복원 항목 전부 "commit-topic-derived" 라벨, (3) 잠재 티어1 후보는 PR 본문 확인으로 격상(#49713 정합 절차) | AC-RDX-004 |
| runner | 상수 `CODEX_THEME_CHECKLIST` — 6테마 리터럴 `thread` / `rollout` / `subagent` / `compaction` / `MCP` / `other`. 행 형식: 테마 키 + 관측 PR 번호 목록 + MoAI 노출면 | AC-RDX-005 |
| specialist | `last-codex-version.json` — Phase 0(판독·부재 기본값) + Phase 7a(쓰기)에 등장 | AC-RDX-006 |
| specialist | `rust-v0.161.0` — 시드 기술 | AC-RDX-007 |
| specialist | `Best-Practices` 상시 절차 섹션 (헤딩 문자열 대소문자 무관 `best-practice` 매치) | AC-RDX-008 |
| specialist | Phase 3 URL **6종 캐노니컬 전문 열거** — `code.claude.com/docs/en/hooks`·`...en/sub-agents`·`...en/skills`·`...en/plugins`·`...en/mcp`·`...en/settings` 각각 ≥1(LED-022..027 육면 열거면, CX-18: URL 전부 삭제 mutant 봉쇄) **+ 제거면 LED-021**: 구형 `docs.anthropic.com` 참조는 0 / exit 1(주변 서술 잔존분 포함 전부 제거 — CX-14; M3 항목 5, 이 카드 소관 — B-02). 육면 검사는 이스케이프 점 + 종결 경계 패턴(B-03) | AC-RDX-009 |
| specialist | `HTML proposal report` — BP 축 명명 산출물 | AC-RDX-010 |
| specialist | `rust-v0.161.0` 부재 기본값 — Phase 0 codex 블록(acceptance §D.3-c, 71–94행)이 부재 기본값(`rust-v0.161.0` + 경고)과 키군 4종을 포함해야 한다 (블록 스코프 게이트 — B-07·B-08) | AC-RDX-014 |
| specialist | `7a-codex` — Phase 7a 기록 단계 제목 리터럴(CC의 Step 7a와 병렬; codex 상태 기록 절차의 사이트 앵커 — CX-6). `last-codex-version.json` 출현 **≥2**: 제1=Phase 0 판독·부재 기본값 블록, 제2=Phase 7a 기록 단계(단일 사이트 mutant는 둘 중 하나에서 좌초) | AC-RDX-006 |
| specialist | `only the CC axis` — Phase 2 조기 종료 문장의 축별 재범위화 리터럴 (REQ-RDX-015, CX-9). **이중 면**: LED-018(신규 리터럴 ≥1) + LED-019 제거면 — 구형 무조건 문장 전문 `If no entries: emit "No new versions since vX.Y.Z" and stop`은 착지 후 **0**이어야 한다(주석 포함 어디에도 생존 금지 — 주석 mutant도 문장 생존 시 적색, fresh-run CX-10) | AC-RDX-017 |
| runner | alpha watch 규범 — CODEX_THEME_CHECKLIST 블록(acceptance §D.3-c, 79–92행)이 "alpha 테마는 watch 관찰목록, 안정 탑재 시에만 채택 판정" 규범 문장(81–84행)을 포함해야 한다 (블록 스코프 게이트 — B-08) | AC-RDX-015 |
| specialist | `source-first` — 원문 패치 선행 강제 리터럴 (REQ-RDX-013, mutant M-4의 기계 판정면) | AC-RDX-016 |

### §D2 결정 전파 (재논의 금지 — spec.md §1.2)

- D1 시드 `rust-v0.161.0` — 변경 시 spec.md D1 + REQ-RDX-002 + plan §C + AC-RDX-007 4곳 동시 갱신.
- D4 — manifest `sprint_contract` 무변경. 변경 제안은 별도 카드로.
- D7 — 스윕 실행·상태 파일 실생성 금지 (run-phase에서도).

### §D3 금지 목록

- `internal/`·`internal/template/templates/` 편집 금지 (Go 0변경 — AC-RDX-011).
- `--no-verify`·force-push 금지. 커밋은 Conventional + 카드 id + `🗿 MoAI` 트레일러.
- 상태 파일(`.moai/state/*.json`) 편집 금지.
- `[NEEDS CLARIFICATION]` 마커 신설 금지 — 판단 필요 시 결정 기록으로 자결.

## §E Self-Verification (run-phase 납품)

- **E1 AC 매트릭스** — acceptance.md §D 전 AC(17종: 블로킹 8 + 판정 보류 6 + 회귀 가드 3) PASS/FAIL/보류 + 검증 명령 + 실측 출력 (§E 삼중 귀속: 명령·출력·HEAD SHA).
- **E2 RED→GREEN 전수 재측정** — §C의 RED-now 앵커 15종(grep) + 검증 동사 2종(E3-P3·P4)이 대응 마일스톤 착지 후 뒤집혔는지 exit code 포함 재실행하고, AC-RDX-009 육면 셀(§C 67–72행)은 착지 GREEN 1을 확인한다.
- **E3 러너 런타임-형태 어댑터 스모크 (plan-audit iter1 CX-4 재설계)** — 러너는 ESM `export const meta`(1행)와 top-level `return`(153행)/`await`(143행)를 결합한 **하이브리드 형태**라 네이티브 Node 모듈 로딩이 어느 목표로도 파스하지 못한다(아래 근거). 워크플로 런타임은 본문을 함수-래핑 평가하므로, 스모크는 그 평가 형태를 재현한다: 행선지 스코프로 `export` 문만 적출한 뒤 AsyncFunction 본문으로 컴파일·실행하고 export 경로로 셀렉터를 실측 호출한다. 이 어댑터가 곧 top-level-형태 보존의 기계 면이다 — `return`/`await`가 함수 래핑 형태에서 벗어나면 컴파일이 즉시 적색으로 뒤집힌다.

  **E3-P1 (CC 축 — 현재 러너에서 관측 완료)**:

  ```bash
  node -e 'const fs=require("fs");const AsyncFunction=Object.getPrototypeOf(async function(){}).constructor;const src=fs.readFileSync(".claude/workflows/hns-release-update-run.js","utf8");const body=src.replace(/^export\s+/gm,"");new AsyncFunction(body);const mod={exports:{}};new AsyncFunction("module","exports","require",body)(mod,mod.exports,require);if(typeof mod.exports.run!=="function")throw new Error("run-missing");const cc=mod.exports.selectResearchSweepTargets({versionDeltas:["9.9.9"]});if(cc.length!==1)throw new Error("cc-count="+cc.length);for(const k of["purpose","agentType","isolation","label","prompt"]){if(!(k in cc[0]))throw new Error("missing:"+k)}console.log("adapter-ok run=fn cc=1 shape-ok")'
  ```

  **관측 (본 트리, Node v22.14.0, 2026-10-09)**: stdout `adapter-ok run=fn cc=1 shape-ok`, **exit 0** — M2 이전 현재 러너에서 관측. 거부된 형태의 근거(전부 본 실행 관측): (1) `node -e "require('./.claude/workflows/hns-release-update-run.js')"` → `SyntaxError: Illegal return statement` (run.js:153, exit 1) — Node v22 모듈 구문 탐지가 top-level await를 보고 ESM 목표로 파스, ESM에서 top-level return은 불법; (2) `node .claude/workflows/hns-release-update-run.js` 직접 실행 → 동일 SyntaxError, exit 1; (3) export 적출 없는 순수 AsyncFunction 래핑 → `export const meta`에서 `SyntaxError: Unexpected token 'export'`, exit 1 — 적출 단계가 어댑터의 필수 전제다.

  **E3-P2 (codex 축 — M2 종료 형태, M2 시점 관측)**: E3-P1 어댑터에 codex 단언을 추가한다 — 실행 직후 `const cx=mod.exports.selectCodexSweepTargets({codexDeltas:["rust-v0.161.0..rust-v0.162.0"]});if(cx.length!==1)throw new Error("codex-count="+cx.length);`와 동일 5키 형태 검사를 넣고 출력을 `adapter-ok run=fn cc=1 codex=1 shape-ok`로 확장한다. 기대: exit 0. 이 형태는 M2가 `selectCodexSweepTargets`를 정의하고 top-level 블록에 연결하며 **export 목록에 추가했을 때에만** 통과한다 — 미연결 정의·주석 mutant는 export 부재 또는 codex-count 단언에서 좌초한다(CX-2의 실질 생성 면, AC-RDX-003의 ≥2 앵커와 짝). M2 종료 시 이 verb의 exit 0 관측을 §E에 귀속한다.

  **E3-P3 (dispatch 병합 관측 — M2 종료 형태, plan-audit iter2 CX-5)**: E3-P1/P2가 export 경로의 셀렉터만 검증한다면, P3는 **러너 자신의 top-level 블록을 모의 런타임으로 실행**해 병합까지 관측한다 — `agent`/`parallel` 목업을 주입해 `agent`가 정의된 상태(런타임 평가 재현)로 블록을 돌리고, 시드 `codexDeltas`가 만든 codex target이 `parallel(...)` 병합을 흘러 `codex-release-notes:` 라벨의 agent 호출로 착지하는지 관측한다:

  ```bash
  node -e 'const fs=require("fs");const AsyncFunction=Object.getPrototypeOf(async function(){}).constructor;const src=fs.readFileSync(".claude/workflows/hns-release-update-run.js","utf8");const body=src.replace(/^export\s+/gm,"");const calls=[];const mockAgent=(p,o)=>{calls.push(o&&o.label?o.label:String(p).slice(0,30));return Promise.resolve("ok")};const mod={exports:{}};new AsyncFunction("agent","parallel","phase","log","args","module","exports","require",body)(mockAgent,(ts)=>Promise.all(ts.map(t=>t())),()=>{},()=>{},({versionDeltas:["9.9.9"],codexDeltas:["rust-v0.161.0..rust-v0.162.0"]}),mod,mod.exports,require).then(()=>{const cx=calls.filter(l=>String(l).startsWith("codex-release-notes:"));if(cx.length<1)throw new Error("no-codex-dispatch:"+calls.length);console.log("dispatch-ok codex="+cx.length+" total="+calls.length)});process.on("unhandledRejection",(e)=>{console.error("REJECTED:",e.message);process.exit(1)});'
  ```

  **관측 (본 트리, 2026-10-09 — M2 이전)**: stderr `REJECTED: no-codex-dispatch:1`, **exit 1** — 현재 러너의 디스패치는 CC 호출 1건뿐이고 codex 라벨 0건(올바른 RED — M2가 연결할 표면이 이 병합이다). **M2 GREEN 기대**: stdout `dispatch-ok codex=1 total=2`, exit 0 — 셀렉터를 정의·export·직접 호출하면서 병합에서 제외하는 mutant는 codex 호출 0건으로 이 verb에서 좌초한다(CX-5의 기계 면; AC-RDX-003의 LED-016).

  **E3-P4 (`run()` 공개 진입점 — M2 종료 형태, fresh-run CX-11)**: `run()`은 Node 소비자용 래퍼 export다 — top-level 블록만 고치면 run()이 CC 전용 target 구성으로 남아 이중 축이 반쪽이 된다. M2가 run()의 병합 경로 공유를 핀하면(위 M2-6), 본 verb가 공개 경로를 codex 전용 입력으로 직접 호출해 관측한다:

  ```bash
  node -e 'const fs=require("fs");const AsyncFunction=Object.getPrototypeOf(async function(){}).constructor;const src=fs.readFileSync(".claude/workflows/hns-release-update-run.js","utf8");const body=src.replace(/^export\s+/gm,"");const mod={exports:{}};new AsyncFunction("module","exports","require",body)(mod,mod.exports,require);const labels=[];const mockSpawn=async(p,o)=>{labels.push(o&&o.label?o.label:"");return "ok"};mod.exports.run(mockSpawn,{versionDeltas:[],codexDeltas:["rust-v0.161.0..rust-v0.162.0"]}).then(()=>{const cx=labels.filter(l=>l.startsWith("codex-release-notes:"));if(cx.length<1)throw new Error("no-codex-in-run:"+labels.length);console.log("run-ok codex="+cx.length+" total="+labels.length)}).catch(e=>{console.error("REJECTED:",e.message);process.exit(1)});'
  ```

  **관측 (본 트리, 2026-10-09 — M2 이전)**: stderr `REJECTED: no-codex-in-run:0`, **exit 1** — run()이 codex 전용 입력(`versionDeltas: []` + `codexDeltas` 1건)에서 agent 호출 0건. 구현은 존재하고 codex 연결이 없다는 실측 형태다. **M2 GREEN 기대**: stdout `run-ok codex=1 total=1`, exit 0 — run()이 병합 경로를 공유할 때만 통과한다(AC-RDX-003의 LED-020).
- **E4 JSON 파스** — `python3 -c "import json;json.load(open('.claude/commands/harness/release-update/manifest.json'))"` exit 0 (domain 문자열 편집 후).
- **E5 회귀 가드** — §C PRESERVE 앵커 3종 + sprint_contract 판독(LED-015 — dimensions·thresholds 출력이 기준선 `['Functionality', 'Consistency'] {'Functionality': 0.85, 'Consistency': 0.8}`와 일치; CX-3, internal/ 0힛 · last-cc-version.json ≥3 · runner_workflow 참조 1 포함).
- **E6 spec-lint** — `go run ./cmd/moai spec lint SPEC-RELUP-DUALAXIS-001` (또는 프로젝트 규약 형태) exit 0 — MissingExclusions·FrontmatterInvalid 0건 확인.
- **E7 의미론 검토면 (CX-12·CX-13 이관 — REQ-RDX-015·REQ-RDX-013의 유일한 구속 판정면)** — 기계 면이 paraphrase·반전 클래스에 우회됨이 실증됐으므로, run-exit E1 인간 검토가 의미론을 검증한다. **REQ-RDX-015 축별 종료**: (a) Phase 2 조기 종료가 CC 축 한정인지, (b) codex·BP 축의 계속 실행·기록이 같은 절차에 명시돼 있는지, (c) Phase 8 완료 게이트가 3축(CC/codex/BP) 실행 상태를 집계하는지. **REQ-RDX-013 source-first (리더 재개 CX-13 합류)**: (d) BP 인벤토리 항목이 제안 근거로 쓰일 때 원문 패치 선행이 절차상 실제 강제인지 — `source-first` 리터럴의 존재만으로 충분하지 않다(리터럴 유지·규칙 역전 mutant 실증). **run-phase 위임 프롬프트는 이 검토 항목들을 반드시 운반한다**(manager-develop-prompt-template §E 성격 — 누락 시 재위임 리스크).

## §F Milestones (결정 가역성 순 — 변동 가능성 높은 결정부터)

### M1 — 스페셜리스트 codex 축 절차 + 상태 스키마 (데이터 모델 결정 — 최상위)

파일: `.claude/agents/harness/hns-release-update-specialist.md`

1. Phase 0: codex 상태 파일 판독 절차 + 부재 시 기본값 `rust-v0.161.0` + 경고 (REQ-RDX-004). 스키마 블록 문서화 — CC 파일 키 계열 미러 + `rust-v0.161.0` 시드 명기 (REQ-RDX-001/002).
2. Phase 1: codex 수집로 신설 — `gh api repos/openai/codex/releases` (비프리릴리즈 판정) + `npm view @openai/codex version` + 본문 비었을 때 커밋 API 폴백 지시 (러너 §D1 앵커와 정합).
3. Phase 2: codex 티어 분류 — 러너 산출(테마별 관찰목록)을 받아 T1/T2/T3 큐레이션. alpha 테마 watch 규범 (REQ-RDX-009). **축별 종료 재범위화 (CX-9, REQ-RDX-015)** — Phase 2 조기 종료 문장("If no entries: emit \"No new versions since vX.Y.Z\" and stop.")을 `only the CC axis` 형태로 재작성: CC 널 델타는 CC 축만 중단하고 codex·BP 축은 같은 run에서 실행·기록. Phase 8 완료 요약은 3축(CC/codex/BP) 실행 상태를 집계 — 미실행 축이 남으면 완료 요약을 내지 않는다(단계 앵커 `only the CC axis` — AC-RDX-017).
4. Phase 7a: Step **7a-codex** 신설 — 단계 제목에 리터럴 `7a-codex`(CC의 Step 7a와 병렬, §D1 앵커)를 넣고 `last-codex-version.json` 병행 기록 (REQ-RDX-003). CC 단독 기록 금지. 본문 내 `last-codex-version.json` 출현은 Phase 0(판독·기본값) + Phase 7a(기록) 2곳이어야 한다(AC-RDX-006 ≥2 — CX-6).
5. **§C 시드값 재판정 실행 지점** — M1 착지 직전 npm/gh 재측정.

### M2 — Runner codex 렌즈 (신규 분석 면)

파일: `.claude/workflows/hns-release-update-run.js`

1. `selectCodexSweepTargets(args)` 신설 — CC 셀렉터와 병렬 형태 (REQ-RDX-006). codex 스윕 버전 창은 `args.codexDeltas` 주입 + 스크립트 본문 시드 상수(CC 셀렉터의 `CURRENT_SWEEP_VERSIONS` 패턴 계승 — args 불신뢰 교훈).
2. `CODEX_COMMITS_FALLBACK` 절차 블록 — §D1 (1)-(3) 내용 (REQ-RDX-007).
3. `CODEX_THEME_CHECKLIST` — 6테마 리터럴 + 행 형식 (REQ-RDX-008). 프롬프트 문자열에 체크리스트 주입.
4. top-level 실행부에 codex 렌즈 병렬 fan-out 편입 + 반환 형태에 codex 영향 표 추가. 불변식(§A.5) 유지 확인. 병합 형태 고정: `const ccTargets = selectResearchSweepTargets(args); const codexTargets = selectCodexSweepTargets(args);` 두 배열을 **단일 `parallel(...)` 디스패치로 합류**(`allTargets`) — codex 렌즈가 CC와 같은 agent() 호출 지점을 흐른다(AC-RDX-003 제2 출현의 위치 요건). codex target 라벨 접두사 `codex-release-notes:<window>`(CC의 `cc-release-notes:`와 병렬) — §E3-P3 관측면의 판정 토큰. **축별 독립성 (CX-9)** — codex 렌즈는 CC `versionDeltas` 공백과 무관하게 `codexDeltas`로 기동한다: 병합이 concat 형태라 CC 목록이 비어도 codex target은 디스패치된다. 구형 "empty versionDeltas makes this Runner a silent no-op" 주석의 적용 범위를 CC 축으로 한정하는 주석 갱신을 M2에 포함(REQ-RDX-015).
5. `module.exports` 확장 — `{ run, selectResearchSweepTargets, MANIFEST_PATH }`에 `selectCodexSweepTargets` 추가(plan §E3-P2 어댑터의 export 경로).
6. `run()` 공개 export 동기화 (fresh-run CX-11) — `run()`도 동일 target-구성 경로를 쓴다: run() 내부를 `selectResearchSweepTargets` 단독 호출에서 병합 경로(top-level 블록과 동일한 allTargets 구성 또는 그 위임 함수)로 교체 — 그렇지 않으면 Node 소비자 경로가 CC 전용으로 남는다(top-level만 고치는 mutant는 §E3-P4에서 좌초).

### M3 — 스페셜리스트 BP 상시 섹션 + Phase 3 URL 세트 (절차 영구화)

파일: `.claude/agents/harness/hns-release-update-specialist.md`

1. `Best-Practices` 상시 절차 섹션 신설: 스윕마다 공식 게시 면 스캔(Anthropic engineering/research·docs, OpenAI 블로그), 인벤토리 기록 (REQ-RDX-012).
2. `source-first` 원문 패치 선행 강제 — 검색 요약·2차 자료는 보고 전용 리드로만 (REQ-RDX-013). 2차 BP-1 게시일 오정보 정정 사례를 절차 근거로 인용.
3. `HTML proposal report` 명명 산출물 기록 (REQ-RDX-014 전반).
4. Phase 3 URL 세트 6종을 `code.claude.com/docs/en/*` 캐노니컬 형태로 갱신 (REQ-RDX-014 후반, 결정 D6). **6종 전문 열거 핀 (CX-18)** — `hooks`·`sub-agents`·`skills`·`plugins`·`mcp`·`settings` 6종 각각의 전문 URL(`https://` + 이스케이프 점 + 종결 경계)이 본문에 존재해야 한다(AC-RDX-009 육면 열거면 — 핀 2aab5f797 = 0 RED, 착지 d36e97571 = 1 GREEN 각각, 보존면).
5. **구형 `docs.anthropic.com` URL 제거 (B-02, 결정 기록 option (a))** — Phase 3 URL 블록과 산문(179행 포함)에서 구형 도메인 문자열을 전부 제거하고, 산문은 구형 도메인 없이 다시 쓴다. 게이트: 착지 커밋의 `grep -c "docs.anthropic.com"` = 0 / exit 1 (LED-021T, AC-RDX-009). 현재 d36e97571에서는 미착지(LED-021G = 1)이며 이 카드의 잔여 run-phase 작업이다.

### M4 — 매니페스트 domain 문자열 (기계적 — 최하위)

파일: `.claude/commands/harness/release-update/manifest.json`

1. `domain`만 갱신: `"moai-adk-go dev-only maintainer tooling — Claude Code + Codex CLI upstream change tracking and the best-practices axis"` (§D1 앵커 2종 내포 — AC-RDX-001/002).
2. `source_request`·`sprint_contract`·골격 불변 확인 (AC-RDX-013).

### 마일스톤 순서 근거

M1(상태 스키마·시드 — 데이터 모델, 0.162 승격 시 변동 가능성 최고) → M2(러너 렌즈 — 신규 면 설계) → M3(BP·URL — 절차 문안) → M4(매니페스트 문자열 — 기계적 1줄) 순으로, 인간 검토 집중도가 높은 결정을 앞세운다.

## §G Anti-Patterns

| 금지 | 이유 |
|---|---|
| 러너에 `gh`/`npm` 네트워크 호출 직접 삽입 | 러너는 읽기전용 fan-out 조정자다 — 네트워크 수집은 렌즈 에이전트(Explore)의 Read/WebFetch 몫 |
| codex 렌즈를 스페셜리스트 본문에만 두고 러너를 안 거치게 함 | 기존 CC 축 분업(러너=비대화형 스윕, 스페셜리스트=인간 게이트) 파괴 (결정 D3) |
| `grep -ci codex`를 AC로 쓰는 것 | 주석 mutant(M-2) 통과 — 식별자 앵커가 판정면이다 |
| sprint_contract 차원 추가 | 결정 D4 — 채점 의미론 변경은 요구 밖 |
| Go 라이터 신설 | 결정/REQ-RDX-005 — 하네스 계층 소유가 측정으로 확인된 구조 |
| 템플릿 미러 편집·make build | 미러 자체가 존재하지 않는다 (§1.1 M9) — 존재하지 않는 것을 만들면 스코프 위반 |

## §H Cross-References

- spec.md §1.2 설계 결정 기록 (D1-D7) / §4 REQ-RDX-001..015
- acceptance.md §D AC-RDX-001..017 + §D.2 게이팅 처분 + §D.3·§D.3-c·§D.3-d 증거 원장 (RED 2aab5f797 · GREEN d36e97571)
- `.moai/research/upstream-update-20261007.md` (1차: C1-C5·6테마 표·Phase 7.5 findings) / `upstream-update-20261008.md` (2차: codex 0.161.0 큐레이팅·URL 세트 finding)
- SPEC-UPDATE-ADD-CODEX-001 (codex 배선 선례) · SPEC-CC2219-UPSTREAM-ALIGN-001 (upstream 정렬 선례)
- `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell 규율) · `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 (관측 클레임)
- 카드 t1605 (GD-1 Haiku 5.5 docs 싱크) · CARD-4 (Codex 0.161.0 conformance) — 스코프 인접, 중복 없음
