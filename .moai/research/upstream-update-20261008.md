# Upstream Update Sweep — CC / Codex CLI / Best-Practices (2026-10-08)

> 확장 스코프 상시 스윕 2차 실행 (운영자 지시 2026-10-07 상시화 — CC+Codex 이중 체인지로그 축·BP 상시 조사 축).
> 실행 주체: release-update 하네스 스페셜리스트 (팀 리드 배차, 읽기전용 — 연구 산출물 외 무기록,
> 상태 파일 미갱신, PR/카드 미발행 — 인간 게이트 대기). 문서 싱크·PR은 후속 manager-git 패스 소관.

## Claim (요약 주장)

| # | 주장 | 판정 |
|---|------|------|
| C1 | CC: 신규 2버전 — 2.1.293 (npm/binary 현행) + 2.1.294 (CHANGELOG 헤드, 미출시 문서-선행) | 확정 (3-way + 헤딩 관측) |
| C2 | CC 델타의 실행 가능 드리프트는 1건 — **GD-1 Haiku 5.5 클러스터** (공식 문서 교차확인). 나머지 Tier 1/2는 positive/NO-OP/additive | 확정 (grep·공식 문서) |
| C3 | Codex: 0.161.0이 2026-10-07T15:58:45Z 안정 승격 (npm 0.161.0, 설치 바이너리 0.160.1). MoAI pin {gpt-6.1-sol, high}은 신기본 카탈로그와 정합 — 어댑터 conformance 재측정 카드 필요 | 확정 |
| C4 | Codex #49713 "repository-local guidance 제거"는 codex 저장소 자체 정리(PR 본문 확인) — AGENTS.md 탐색 동작 변화 아님 | 확정 (PR 본문 관측) |
| C5 | BP 축: 원문 2건 신규 패치 — "Scaling Managed Agents"(2026-04-08, 어제 리드의 날짜 '10월'은 오정보) + "Prompting Claude Opus 5.5" 공식 가이드. 어제의 "80% 시스템 프롬프트 축소" 주장은 공시 문서에서 미확인(2차 자료) — 보고 전용 유지 | 확정 |
| C6 | MoAI `prompting-best-practices.md`의 모델 목록(Opus 4.8/4.7·Sonnet·Haiku 4.5)이 공식 가이드 패밀리(Opus 5/5.5·Sonnet 5.5·Haiku 5.5) 대비 낡음 — GD-1에 합류 | 확정 (grep·공식 문서) |

## Evidence (증거)

### 축 1 — Claude Code 2.1.292..2.1.294

**버전 확인 (본 실행 관측)**:
- `claude --version` → `2.1.293 (Claude Code)`; `npm view @anthropic-ai/claude-code version` → `2.1.293`.
- `curl -fsSL https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` → 8561줄 / 954,183바이트,
  sha256 `b60a2efb867f7de45d3dc026bdeb29b0c03a70128bc3d290889bec96451522b9`
  (스냅샷 `CHANGELOG-upstream.md` 동봉). 헤딩: `## 2.1.294`(3행) → `## 2.1.293`(8행) → `## 2.1.292`(67행).
- **문서-선행**: CHANGELOG가 2.1.294까지 기재했으나 npm/binary는 2.1.293. 2.1.294 본문 2불릿은 게재본 기준 분석
  (시리즈 전례: CHANGELOG 재작성 가능 — Residual-risk).

**델타**: 2.1.293 = 57불릿 (9-66행), 2.1.294 = 2불릿 (4-7행), 계 59불릿.
티어 (불릿당 판정, 근사): **Tier 1 = 16 / Tier 2 = 3 / Tier 3 = 40**.

**Tier 1 (16) — 표면·판정**:

| # | 버전 | 항목 | 판정 | MoAI 표면 (실측) |
|---|------|------|------|------------------|
| 1 | 2.1.293 | Claude Haiku 5.5 (`claude-haiku-5-5`) 기본 Haiku 승격·1M | **GD-1 진드리프트** | 하단 GD-1 상세 |
| 2 | 2.1.293 | `subagentStatusLine` payload에 `agentType` 추가 | additive NO-OP | docs-site에 payload 표 미커버(grep 0힛); 공식 statusline 문서에 "Requires Claude Code v2.1.293" 명기 확인 |
| 3 | 2.1.293 | mods `$.tool.register`에 `isDeferred` | NO-OP | MoAI plugins 미출하 |
| 4 | 2.1.293 | HTTP MCP 접속 메모리 누수 수리 | positive NO-OP | moai mcp-server는 stdio |
| 5 | 2.1.293 | 도구 제거 세션에서 SendMessage 사용 지시 오류 수리 | positive | cross-session-messaging.md 역주장 없음 |
| 6 | 2.1.293 | 서브에이전트 "내장 도구 세션 전체 비활성" 오표 수리 | positive | worktree-integration.md tool-restriction 독트린과 정합 (자기 tools 목록 제외 ≠ 전역 비활성) |
| 7 | 2.1.293 | Bash 단일파일 cat/head/tail/sed/grep 읽기에 path-scoped 룰·중첩 CLAUDE.md 미적용 수리 | **positive (MoAI 직접 수혜)** | settings.json.tmpl:557-562 `Read(./secrets/**)`·`Read(~/.ssh/**)` 등 path-scoped deny 룰이 Bash 우회 경로에도 적용되는 보안 강화 |
| 8 | 2.1.293 | 2.1.281 auto-mode 거부 메시지("결과까지 커버") **revert** | audited NO-OP | `grep -rni "covers the outcome\|not only the exact command"` → 0힛 (MoAI 미흡수) |
| 9 | 2.1.293 | claude.ai 스킬 동기화 40분 간격 | NO-OP | MoAI 스킬은 로컬 |
| 10 | 2.1.293 | 에이전트 목록·MCP 서버 비ASCII 정렬 변경 | NO-OP | MoAI 에이전트명 전부 ASCII |
| 11 | 2.1.293 | bypass 동의 settings.local 저장 무시 수리 | positive NO-OP | 독트린 역주장 없음 |
| 12-14 | 2.1.293 | mods classic.* 훅 재시작 스킵 수리 / `claude plugin test` mock.session / `claude plugin eval` Docker Desktop | NO-OP | MoAI plugins 미출하 |
| 15-16 | 2.1.294 | "지시문 형태" prompt/agent 훅 차단 실패 수리 + Stop/SubagentStop 판정 개선 | NO-OP | MoAI 훅은 JSON 스키마 스크립트 훅 — hooks-system.md에 instruction-form 0힛 (grep exit=0 무출력) |

**Tier 2 (3)**: ① 컴팩션 직전 자기 작업 재수행 수리 (positive — Reduction Ladder 독트린 무영향) ② claude.ai 동기화 스킬 설명 전달 수리 (NO-OP) ③ Bash 편집 diff 노트 "명령 실행 중 변경, 타 프로세스 쓰기 포함" (positive — MoAI 다중세션 경합 독트린과 정합, 주석 기회).

**Tier 3 (40)**: Remote Control·클라우드 스트리밍·Claude Tag(Slack)×5·Code Review·vim×2·keybindings·/feedback·claude purge·logs/stop/kill 인증·footer/agents-view UI·/ultrareview×2·/model UI·/tui chrome·← 백그라운드×3·pasted text×2·Windows PID·데스크톱 정책 게이트웨이·Team/Enterprise 기동·Chrome×2·OTel at_mention·self-hosted runner·artifacts 핀·cloud /loop revert 등 — MoAI 표면 없음.

#### GD-1 상세 — Haiku 5.5 클러스터 (유일 실행 가능 드리프트)

공식 문서 교차확인 (`https://code.claude.com/docs/en/model-config`, 본 실행 WebFetch — webReader 500 후 폴백):
- ID `claude-haiku-5-5`, **1M 컨텍스트 전 플랜** (`[1m]` 접미사 불요), v2.1.293+ 필요, auto-compact 기본 ~967K.
- 별칭 해석: **Anthropic API → Haiku 5.5** / AWS·Bedrock·GCP Agent Platform·Foundry → Haiku 4.5 (200K 유지).
- effort: Haiku 5.5는 `low/medium/high/xhigh/max` **전체 지원**, 기본 **medium** (Opus 5.5·Sonnet 5.5·Haiku 5.5 공통).
  thinking 끌 수 없음·항상 adaptive reasoning. 100K 초과 프롬프트는 증액 과금.

MoAI 표면 (워크트리 t1538 @ 65e649d5f grep 실측; 편집 시점 재그레핑 필수):

| 표면 | 위치 | 갱신 내용 |
|------|------|-----------|
| 컨텍스트 창 임계 표 | `.claude/rules/moai/workflow/context-window-management.md:17` "Haiku (200K)\|200,000\|90%" | Anthropic API 경로 haiku=1M 반영 — Haiku 5.5 (1M) 행 분할 또는 주석 (provider별 해석 노트: 기존 "200K 세션 행 우선" 규칙과 정합되게) |
| 모델 정책 | `.claude/rules/moai/development/model-policy.md:83` ("Haiku 4.5 still ships 200K"), `:161` (effort 지원 문장 — "Haiku 4.5 supports neither"), `:18`/`:26` (별칭 설명) | 사실 유지 항목은 유효하나 현 라인업 프레이밍 갱신 + effort 행에 Haiku 5.5 (xhigh·max 지원·기본 medium) 추가 |
| 프롬프팅 BP | `.claude/rules/moai/development/prompting-best-practices.md:7` "(Opus 4.8/4.7, Sonnet (current generation), Haiku 4.5)" | 공식 가이드 패밀리(Opus 5·5.5, Sonnet 5.5, Haiku 5.5) 기준으로 갱신 + BP-2 원문 참조 |
| 템플릿 미러 | 위 3종의 `internal/template/templates/.claude/rules/...` 미러 | 라이브-미러 바이트 동일 동기화 + `make build` |
| docs-site (4-locale) | multi-llm/model-policy.md:39 · multi-llm/_index.md · advanced/token-budget.md:29 · cost-optimization/prompt-caching.md:155 · claude-code/context-memory/context-window.md:57 · claude-code/foundations/commands.md:94 · claude-code/foundations/how-claude-code-works.md:65 · claude-code/_index.md:21 (각 locale) | "Claude Haiku 4.5 \| `claude-haiku-4-5-20251001` \| 200K" 행 갱신/추가 — Haiku 5.5 (`claude-haiku-5-5`, 1M, 별칭 해석 provider별, v2.1.293+). docs-i18n-check.sh 실행 |
| Go 코드 | 없음 — 별칭 기반 (`internal/config/defaults.go:155` DefaultSpeedModel="haiku"), envkeys.go `ANTHROPIC_DEFAULT_HAIKU_MODEL` 유효 | 코드 변경 불필요 |

규모: **Tier S~M docs 단일 chore PR**. umbrella SPEC 불필요 (actionable=1 — 2026-08-27·2026-10-07 선례).

### 축 2 — Codex CLI 0.160.1 → 0.161.0 (안정)

**버전 확인 (본 실행 관측)**: `npm view @openai/codex version` → `0.161.0`; `codex --version` → `codex-cli 0.160.1`;
`gh api repos/openai/codex/releases` → `rust-v0.161.0` (prerelease=false, **2026-10-07T15:58:45Z** — 전일 스윕
약 7시간 뒤 승격). 0.162.0-alpha.20/18.1은 여전히 alpha. 릴리즈 본문 스냅샷 `codex-0.161.0-release-body.md` 동봉.

큐레이팅 본문 14불릿 분류: **Tier 1 = 4 / Tier 2 = 4 / Tier 3 = 6** (커밋 로그 #49246..#49713은 원자료로 열람).

| # | 항목 (PR) | 티어 | 판정·MoAI 표면 |
|---|-----------|------|----------------|
| C-1 | GPT-6.1 Sol 기본 카탈로그 (#49318/#49339) | 1 | positive 정합 — MoAI pin `{gpt-6.1-sol, high}` (closed_sets.go:96 DefaultCodexAuditModel, defaults.go:1647, SPEC-MODEL-MATRIX-UPDATE-001). 언핀 세션의 기본값만 변화 |
| C-2 | 인증 문서 keyring 반영 (#49361 + #49384/#49392 계측) | 1 | **watch** — mcp_codex.go:2212-2223이 `<CODEX_HOME>/auth.json` 직독(Stage 1). 신규 설치의 키링 저장 확산 시 감지면 영향 가능 → conformance 카드 항목 |
| C-3 | thread resume 최신 커밋 이력 포함 + SQLite 손상 조기감지·백업 (#49599/#49701/#49710) | 1 | positive — MoAI thread/rollout 소비면(codex_event_adaptation 계열) 수혜 |
| C-4 | approved filesystem escalation 의미 확장 (#49353) | 1 | watch — 샌드박스 쓰기 권한 의미 변화, MoAI 레인 독트린 역주장 없음(관측), 재검증 대상 |
| C-5 | 명시적 launch 권한 재접속 지속 (#49809) | 2 | watch — managed 세션 재접속면 |
| C-6 | 세션 인덱스·스레드명 배치 성능 (#49297/#49305/#49708/#49692-96) | 2 | positive |
| C-7 | hook matcher 컴파일 (#49379) | 2 | NO-OP (성능) |
| C-8 | npm alpha dist-tag 후퇴 방지 (#49704) | 2 | 추상 방법론 정합 (alpha 추적 신뢰성) |
| C-9~14 | Daybreak/Cyber×2·/mcp login TUI·voice 장치·Bedrock 멀티에이전트 V2/Ultra/GovCloud·Windows 샌드박스/UI 군 | 3 | MoAI 무관 (moai mcp-server는 로컬 stdio — /mcp login·enterprise MCP auth 전항 NO-OP) |

**#49713 검증 (전일 alpha 관찰목록의 잠재 Tier 1)**: "Remove repository-local Codex guidance, skills, and
environment config" — `gh api repos/openai/codex/pulls/49713` 본문 확인 결과 **codex 저장소 자체 정리**
("Delete the root `AGENTS.md`, the skills under `.codex/skills/` … and `.codex/environments/environment.toml`",
merged 2026-09-30). AGENTS.md 발견 체인 동작 변화 아님 → NO-OP. 커밋 주제만으로 판정하지 않고 PR 본문으로 확정한 사례.

**전일 alpha 테마 대조**: #51230/#51595/#51602(thread), #51329(서브에이전트 fork 제거), #51117(compaction 전체컨텍스트),
#51611(MCP elicitation) 등 51xxx번대는 0.162-alpha에 잔류 — 0.161.0 미탑재 확인 (관찰목록 승계).

**Conformance**: 설치 0.160.1 < 안정 0.161.0, testdata 핀 `codex-0.160.0` (managed_codex_tui_test.go 등),
전일 기준 리포 참조 0.160.1(t1531). → **어댑터 conformance 재측정 카드 제안** (fixtures 0.161.0 갱신 +
auth.json/키링 실측 + `resume --help` 출력 대조 + thread-resume 동작).

### 축 3 — Best-Practices (상시 축 2차)

**BP-1 (원문 패치 성공) — "Scaling Managed Agents: Decoupling the brain from the hands"**
`https://www.anthropic.com/engineering/managed-agents` (게시 **2026-04-08** — 전일 리드의 "2026-10"은 오정보,
원문 패치로 수정. 전일 Residual-risk의 "검색 리드 오정보 가능"이 실현된 사례).
- 뇌(모델+하네스)·손(샌드박스/도구)·세션 로그의 분리; "Harnesses encode assumptions that go stale as
  models improve" (하네스 휴리스틱은 부패성 자산); 하네스도 cattle — `wake(sessionId)` + 내구 로그 리플레이로 재부팅;
  p50 TTFT −60% / p95 −90%; "Session ≠ context window" (컨텍스트 밖 내구 이력 + 가역 변환);
  구조적 보안("tokens are never reachable from the sandbox") vs 좁은 스코핑; 손의 지연 프로비저닝.
- **MoAI 매핑**: release-update 주기 자체가 "부패성 휴리스틱 감지기"; lane watchdog·증거-디스크 부활
  (auto-semantics §5.1) = wake-replay 패턴의 저장소-디스크 등가물; branch-guard·훅 = 구조적 보안 원칙 정합.
- **갭**: "부패성 휴리스틱"을 클래스로 명명하는 독트린 없음 — 전일 CARD-3(BP 매핑 참조 문서) 범위에
  본 문서 매핑을 합류시키는 수정 제안.

**BP-2 (원문 패치 성공·신규) — "Prompting Claude Opus 5.5" 공식 가이드**
`https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-opus-5-5`.
- effort 기본 medium(Opus 5는 high), "Opus 5.5 at medium matches or exceeds Opus 5 at high", low 근접·저비용;
  **최상위 effort 값 변경은 프롬프트 캐시 무효 — per-message effort change(beta)로 회피**;
  xhigh·max는 실측 이득이 있는 곳에만.
- **unattended runs**: "Treat a text-only end of turn as a report rather than as proof the task is done" —
  체크리스트 유지, 잔여 항목 지적 후 계속, **2-3회 자동 계속 후 정지**(무한 계속 금지), 별도 소형 모델의
  종료조건 점검 패턴 공식화.
- **멀티에이전트 시간 신호**: `elapsed 340s / 1200s` 예산 라인이 팀 완료를 앞당김(단일 에이전트 대비,
  품질 유지) — effort 인하와 다르게 "병렬 유지" 효과.
- 채팅 프롬프트의 "think carefully"류 지시 제거 권고; pasted-text 마킹(무작위 id 태그 + 시스템 프롬프트 노트);
  progress-update thinking 블록(display: "updates").
- **MoAI 매핑**: goal-directive(평가자·정체 가드·연속 차단 상한 min(ceiling, cap))은 unattended 패턴의
  MoAI 등가물 — 공식 가이드가 같은 설계를 권고(강화 근거). cache-aware-execution directive 10(세션 내
  모델·effort 전환 캐시 무효)이 공식 문서로 재확인 + per-message effort beta 회피로직은 참조 각주 가치.
  "elapsed-time budget"은 팩토리 레인·팀 모드 실험 후보(카드급). Opus 5.5 Prompt Philosophy(방어 스캐폴딩
  제거)는 "think carefully 제거"·"재검증" 권고와 정합.

**BP-3 (승계) — Ref 2 "Context engineering: memory, compaction, and tool clearing"**: platform.claude.com
2026-03-20 게시 재확인(검색) — **정확 경로는 여전히 미확정** (CARD-2 저작 시 확정).

**BP-4 (승계·보고 전용) — "Claude 5 세대 시스템 프롬프트 80% 축소" 주장**: 공식 문서에서 직접 확인 불가
(2차 자료: developersdigest 등). 공식 근거는 Opus 5/5.5 가이드의 "remove instructions that stood in for
thinking"·"re-test mitigations" 방향성. 80% 수치는 비공식 — 인용 금지, 보고 전용 유지. t1469(상시 로드 예산)와의
연결은 방향성 근거로만.

**공식 프롬프팅 가이드 패밀리 URL (핀 완료)**:
- Opus 5: `.../prompt-engineering/prompting-claude-opus-5` / Opus 5.5: `.../prompting-claude-opus-5-5`
- 허브: `.../prompt-engineering/overview` (Opus 5·4.8·Sonnet 5.5·Sonnet 5·**Haiku 5.5** 가이드 존재 — GD-1 참조)

## Baseline-attribution (기준 귀속)

본 실행·본 트리 관측: (a) CC — baseline `.moai/state/last-cc-version.json` last_analyzed 2.1.292 (2026-10-07)
판독 + curl/npm/claude 3-way 본 실행 측정 + sha256 기록; (b) Codex — 전일 기준 0.160.1(안정)·리포 핀
0.160.0에서 npm/gh/codex --version 본 실행 측정; (c) 라이브 grep — **워크트리 t1538 @ 65e649d5f**
(WT-t1538-factory-recovery, origin/main ca721b91d 흡수 트리 — primary의 다른 진행 브랜치·미병합 워크트리 제외);
(d) 공식 문서 4건 WebFetch(model-config·statusline·managed-agents·prompting-claude-opus-5-5) 본 실행 패치.
GLM 라우팅: webReader 500 오류(2회 연속 실행) 후 orchestrator 승인 폴백으로 WebFetch 사용; WebSearch는
z.ai web_search_prime로 라우팅 관측(내장 직접 호출 아님).

## Gaps (명시적으로 관측 안 된 것)

- 2.1.294는 npm/binary 미배포 — CHANGELOG 게재본만으로 분석 (재작성 전례 상존). 다음 스윕 재확인.
- Tier 3 (CC 40·Codex 6)은 나머지 셈샘 — 벌크 내 Tier 1/2 후보 누락 배제 불가 (시리즈 관례).
- 공식 문서 6종 병렬 패치를 GD 판정 연결 4종으로 축소 (전일 선례) — skills/plugins/hooks/mcp 문서 미패치.
- Codex conformance는 **실행 없음** — 0.161.0 설치·fixture 대조는 카드 소관 (본 run은 읽기전용).
- docs-site en 로케일 Haiku 행은 ko/ja/zh grep 우선 — 편집 시점 4-locale 전수 재그레핑 필수.
- primary 체크아웃의 다른 활성 세션(main·t1538-resume·t1554) 미병합 작업 트리는 모든 라이브 주장에서 제외.
- 상태 파일 미갱신 (인간 게이트 대기). BP-3 정확 경로·BP-4 공식 근거 미확정.
- t1579(CARD-1) 착지 전까지 확장 스코프는 배차 메시지 의존 — 본 run도 재배차 메시지로 구동됨 (기억 + 배차문 이중 확인).

## Residual-risk (잔여 위험)

- 상류 CHANGELOG 재작성(2.1.294 내용 변동 가능). GD-1의 provider별 해석은 model-config 게시본 기준 —
  Anthropic API 외 경로(AWS 계열)에서 haiku=4.5 유지는 게시 시점 관측.
- Codex keyring 전환 여부는 문서화(#49361)만 관측 — 동작 전환은 미실측 (conformance 카드에서 확정).
- effort "기본 medium" 행이 Haiku 5.5에도 적용되는 것은 공식 표 관측 — MoAI effort 라우팅(harness.yaml)과의
  상호작용은 편집 시점 재확인.
- 티어 경계는 판단치 (T1-2 중 fail-closed 관리 항목이 T3로 읽힐 수 있음).

## Update Plan (Phase 4 — 플랜 초안, 미집행)

**권장: 옵션 A형 — GD-1 단일 chore PR + conformance/BP 카드 발행 (발행은 리더)**

1. **GD-1 docs 싱크** (manager-docs 위임, 4-locale 규율·URL 블랙리스트·Mermaid TD 준수):
   rules 3종 live+template 미러 + docs-site 8페이지×4 locale + `make build` + `scripts/docs-i18n-check.sh`.
   브랜치 `chore/cc-update-20261008`, 커밋 `chore(release-update): track CC 2.1.292..2.1.294 upstream changes (Haiku 5.5)`.
2. **카드 제안 (발행은 리더)**:
   - **CARD-4 (High)** Codex 0.161.0 conformance 재측정 — testdata/codex-0.160.0 핀 갱신, auth.json/키링 실측
     (mcp_codex.go:2212 Stage-1), `codex resume --help` 대조, thread-resume·escalation(#49353) 재검증.
   - **CARD-2 수정** (전일 승계, Medium) tool result clearing 럇 — BP-3 정확 경로 확정 포함.
   - **CARD-3 수정** (전일 승계, Medium) 멀티에이전트 실패 패턴 매핑 + **BP-1(managed-agents)·BP-2(Opus 5.5 가이드)
     매핑 합류** (elapsed-time budget 실험 포함).
   - **CARD-1 승계** (t1579 진행 중) — 본 run findings가 해당 카드 입력.
3. **상태 파일 갱신** (옵션 확정 후): last_analyzed 2.1.294 + analysis_history 항목 (codex 축은 note 필드에
   0.161.0 기록 — 스키마 확장은 t1579 소관).

umbrella SPEC 불필요 (actionable=1, 전례 준용).

## Phase 7.5 — Improvement findings (REQ-HRR-003/004 형식; confidence는 실행 시점 보수 추정치 — 하한 0.70 준수, learner 기본값 1.0 미사용)

```jsonc
"findings": [
  {
    "surface": ".claude/rules/moai/core/glm-web-tooling.md",
    "kind": "friction",
    "summary": "z.ai webReader 500 error on code.claude.com doc fetch, 2nd consecutive run (2026-10-07, 2026-10-08); the HARD routing table has no codified degradation path, forcing an orchestrator-sanctioned ad-hoc fallback to built-in WebFetch each time.",
    "confidence": 0.8,
    "suggested_tier": "rule"
  },
  {
    "surface": ".claude/agents/harness/hns-release-update-specialist.md (Phase 3 doc URL set)",
    "kind": "drift",
    "summary": "Specialist body's canonical doc-fetch URL set still lists docs.anthropic.com/en/docs/claude-code/* URLs; every fetched page canonicalizes to code.claude.com/docs/en/* (re-verified this run on model-config and statusline). t1579 owns the body — recorded as input to that card.",
    "confidence": 0.75,
    "suggested_tier": "rule"
  },
  {
    "surface": ".moai/state/last-cc-version.json (schema)",
    "kind": "gap",
    "summary": "State schema still has no codex-axis key; this run's codex baseline (0.161.0 stable) can only live in a free-text note field — repeats yesterday's CARD-1 structural gap (t1579 owns the fix).",
    "confidence": 0.7,
    "suggested_tier": "rule"
  }
]
```

## 다음 스윕 기준

- CC: CHANGELOG sha256 불일치 또는 npm/binary > 2.1.293 (2.1.294 배포 여부) 시 델타 분석.
- Codex: 안정 > 0.161.0 또는 0.162 승격 시 델타 + conformance(CARD-4 결과 연동).
- BP: BP-3 정확 경로 확정 + Claude 5 가이드 패밀리 신규 게시물 스캔 + elapsed-time budget 실험 결과 회수.
