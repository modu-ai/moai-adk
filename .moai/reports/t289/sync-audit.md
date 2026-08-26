# sync-audit — SPEC-GLM-FLASH-DEFAULT-001 (card t289, Tier M)

- 감사 트리: `701bd64e0` (branch `WT-glm-flash-default`, worktree `.claude/worktrees/t289`)
- 비교 기준: `origin/main` = `da791eb0a` (merge 36f7b4e04 로 반입됨)
- 감사 범위: M1–M6 (`a5454a505`..`9e1bb9e3d`) + sync close (`f1208eba4`) + backfill (`701bd64e0`)
- 방식: 코드 대조 + 샘플 재실행(명령·출력 원문 기록) — Claude 단독 감사(`audit_model` 미설정, `.moai/config/` grep 0건)

## 1. 재실행한 기계 검증 (전부 이번 실행, 이 트리)

| # | 항목 | 명령 | 관측 출력 (원문) |
|---|------|------|------------------|
| V1 | M6 boot smoke | `go test ./internal/cli/ -run TestGLMFlashDefaultEnvInjection -v -count=1` | `--- PASS: TestGLMFlashDefaultEnvInjection (0.00s)` / `ok github.com/modu-ai/moai-adk/internal/cli 1.390s` |
| V2 | overlay mirror (양방향) | `go test ./internal/template/ -run 'TestCollapseClaudeEffortToGLMForModel\|TestResolveGLMReasoningForModel\|TestSessionGLMReasoningStateForModel\|TestIsGLMFlashModel' -v -count=1` | 4개 전부 `--- PASS` / `ok ... 0.514s` — flash × {low, medium, high, xhigh, max, bogus, ""} → max 행과 비-flash(glm-5.3/glm-5.1/"") × low → low 회귀 행이 **같은 테스트 안에** 존재 |
| V3 | closed set + 7 상수 | `go test ./internal/config/ -run 'TestDefaultGLMConstants\|TestNewDefaultLLMConfig_GLMTierMapping' -v -count=1` | `--- PASS: TestDefaultGLMConstants` + `defaults_test.go:432: ValidGLMModels() = [glm-5.3-flash glm-5.3 glm-5.1 glm-4.7 glm-4.5-air]` |
| V4 | web widget + i18n | `go test ./internal/web/ -run 'TestGLMModelSelectOptions\|TestGLMFlashOptionLabelsAllLocales' -v -count=1` | 둘 다 `--- PASS` / `ok ... 0.527s` |
| V5 | context window 테이블 | `go test ./internal/statusline/ -run 'TestGLMContextWindowsFlashDirectEntry\|TestResolveGLMContextWindow' -v -count=1` | `--- PASS` × 2 — 서브케이스에 `glm-5.3-flash direct table entry (divergence guard)`, `glm-5.3 retained entry`, `unregistered glm-5.3-* variant inherits 1M via substring` 포함 확인 |
| V6 | build | `go build ./...` | exit 0 (`BUILD_EXIT_0`) |
| V7 | vet | `go vet ./internal/config/... ./internal/template/... ./internal/statusline/...` | exit 0 (`VET_EXIT_0`) |
| V8 | catalog.yaml 불변 | `git diff origin/main..HEAD --name-only -- internal/template/catalog.yaml \| wc -l` | `0` — 바이트 불변 직접 확인 |
| V9 | subagent boundary | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/{config,template,statusline,web} --include='*.go' \| grep -v _test.go \| grep -v '// '` | `0` |
| V10 | i18n 8키 | `grep -c 'glm-5.3-flash' internal/web/assets/i18n.js` | `8` — 2개 키 패밀리 × en/ko/ja/zh, 전부 비어 있지 않은 현지화 라벨 |
| V11 | CHANGELOG 대조 | `grep -c 'SPEC-GLM-FLASH-DEFAULT-001' CHANGELOG.md` | `1` — 단일 항목, AC 13/13 명시, 카드 id는 t289뿐(t290 0건), 코드 사실(7 슬롯, 5원소 집합, DefaultGLM53 보존, overlay wire+display, statusline 직접 엔트리) 전부 소스와 일치 |
| V12 | AC 수 대조 | `grep -c "^### AC-" acceptance.md` | `13` — CHANGELOG 기술과 일치 |

## 2. 계약 요소 대조 결과

- **7개 tier 상수 + 템플릿 트윈**: `internal/config/defaults.go:154-176` — `DefaultGLM53Flash = "glm-5.3-flash"` 신규, `DefaultGLMHigh/Medium/Low/Fable` + `DefaultGLMHaiku/Sonnet/Opus` 전부 `DefaultGLM53Flash` (7개). 템플릿 `llm.yaml:183-186` 4슬롯 전부 `"glm-5.3-flash"`. ✓
- **glm-5.3 보존**: `DefaultGLM53 = "glm-5.3"` 명명 상수 + `closed_sets.go:83` `ValidGLMModels()` 명시적 멤버 — 베어 리타깃으로 빠지지 않음. F-3 교훈 반영 확인. ✓
- **flash max-only overlay**: `glm_effort_overlay.go` — `CollapseClaudeEffortToGLMForModel` / `ResolveGLMReasoningForModel` / `SessionGLMReasoningStateForModel` 세 함수 모두 `IsGLMFlashModel(model)` → `glmReasoningMax`, 비-flash는 기존 붕괴(`CollapseClaudeEffortToGLM`)에 무변경 위임. **거울 양방향**이 코드와 테스트(V2) 모두에서 확인됨 — 비-flash 세션 저평가 회귀 없음. ✓
- **display/wire threading 일치**: `internal/web/agentfm.go:314`, `internal/cli/model.go:117` — `ResolveGLMReasoningForModel(llm.GLM.Models.High, ...)`; wire는 `launcher.go:1198` → `glm.go:430` `SessionGLMReasoningStateForModel(glmHighModel, effort)` (`glmHighModel` = `resolveGLMBackendForLaunch`의 high 슬롯). 세 표면이 전부 high-슬롯 키잉으로 일치. 웹 표시-와이어 불일치를 잡는 `TestAgentFMGLMReasoningMapFlashPinsMax`도 통과. ✓
- **statusline**: `memory.go:34` `"glm-5.3-flash": 1_000_000` 명시 엔트리 + `"glm-5.3"` 유지 + 등록 시점 가이던스 주석. ✓
- **web**: 스키마는 `ValidGLMModels()` 파생(리터럴 재선언 없음), `glm_tier_test.go:54` 5원소 집합 핀. ✓
- **docs**: README 4개 로캘 + docs-site 20페이지 전부 브랜치 범위 안; en README:687 "glm-5.3-flash (the default)", ko/ja/zh 표본 각각 확인. ✓
- **3-phase close 무결성**: `f1208eba4` 파일 = `spec.md`(frontmatter `status: in-progress → completed` 한 줄 — 바디 무변경) + `progress.md` + `CHANGELOG.md` 딱 3개. backfill `701bd64e0`는 `sync_commit_sha: "pending-backfill" → "f1208eba4"` 1줄 — D3 면제 패턴 정확. ✓
- **템플릿 규율**: catalog 바이트 불변(V8 직접 측정), 중립성 grep 2건은 `1000000` 숫자 리터럴 오탐(규칙의 SHA 날짜 패턴에 우연히 걸림) — SPEC-ID/날짜/SHA 0건. `make build`는 감사자 재실행 안 함(트리 변경 연산) — 컴파일·임베드 등가검증은 V6 `go build ./...` exit 0, catalog는 V8 직접 diff. §E.2의 make-build 주장은 소비만 함(F3 gap).

## 3. Findings

- **F1 [Medium] [optional]** — per-slot 분기 residual이 스키마로 **도달 가능**하며 §E에 **미기록**. `schema_sections.go` `llmFields()`는 high/medium/low/fable **독립 select 4개**(옵션各 5원소)를 렌더하므로 web 콘솔에서 high=glm-5.3 + low=flash 같은 혼합 구성이 가능. 이때 reasoning env는 high 키잉이라 flash 슬롯 서브에전트가 저추론(낮은 thinking budget)될 수 있음 — shim이 top-level reasoning_effort를 무시한다는 t175 측정(t175 measurements §3) 때문에 하드 거부는 아니지만 저추론 방향은 실재. 배치 전제("§E.3/§E.2 residual notes에 out-of-scope로 기록됨")는 검증에 실패함: progress.md 어디에도 해당 residual 문구 없음(§E.4 gaps는 cli-suite/coverage 2건뿐). 수정: 폐쇄된 감사 추적에 residual 1줄 기록(또는 후속 SPEC에서 슬롯별 키잉) — 본 SPEC 요구사항 위반은 아님(REQ-004의 "resolved model" = 세션 모델 해석이 세 표면 일관).
- **F2 [Low] [optional]** — §E.3 `total_run_phase_files: 33` 오계. 실측 `git diff a5454a505^..9e1bb9e3d --name-only` = **44파일**(Go 17 + i18n.js + llm.yaml + docs 24 + spec.md frontmatter flip). 주석의 산술 자체(9+1+1+24)도 33과 안 맞음. 메타데이터 정확성 문제이며 AC에는 무영향.
- **F3 [Low] [gap]** — `make build` 미재실행(감사자가 트리 변경 연산 회피). 등가검증: V6 build exit 0 + V8 catalog diff empty. §E.2의 `make build` exit 0 / catalog 불변 주장 중 catalog는 독립 재측정으로 확인, build는 컴파일 등가로 대체 확인.
- **F4 [Low] [info]** — `IsGLMFlashModel` 부분문자열 의미론: "glm-5.3-flash"를 포함하는 미래 id(예: glm-5.3-flash-lite)도 flash로 취급 → max 고정. 과잉추론 방향이라 안전 쪽 실패이며 memory.go의 등록 시점 가이던스 관행과 상호보완됨. 미래 등록 시 알아둘 사항.
- **F5 [Low] [info]** — primary 체크아웃 병행 핫픽스(app.js `wireGLMFlashEffortLock`, "GLM Settings" 섹션명)는 의도적 미복제(§E.2.5). `git log origin/main -S "wireGLMFlashEffortLock"` 0건 — origin/main에도 없으므로 이 시점 기준 merge 조정 부채 없음. primary 로컬 발산은 orchestrator 소관으로 남음.

## 4. 미검증 (Gaps)

- `make build` 본체 재실행 (F3).
- 전체 `internal/cli` 스위트 — lane-local 규율상 CI 위임(§E.4 gaps와 동일).
- docs 4-로캘 전수 대조 대신 표본 검사(en 전수 grep + ko/ja/zh 표본 15건).
- live z.ai API 왕복 — AC-012가 env 수준 관측을 명시하므로 설계상 제외.

## 5. 잔여 위험 (Residual-risk)

- 혼합 슬롯 구성에서의 저추론 (F1) — web 콘솔로 도달 가능, 미기록 상태.
- flash로의 라이브 와이어 동작(shim의 thinking-budget 매핑 품질)은 본 감사 범위 밖 — t175 측정 간접 근거에 의존.
- primary 로컬 핫픽스와의 최종 병합 (F5).

## 6. 평가 (4차원, 0–1)

| 차원 | 점수 | 근거 |
|------|------|------|
| Functionality (40%) | **0.95** | 13/13 AC 재검증 통과(V1–V5, V10–V12 원문 출력); close·backfill 무결성 정확; mirror 양방향 코드+테스트 확인. 감점: F1 residual 미기록(보고 정확성), F2 오계. |
| Security (25%) | **0.95** | 신규 신뢰경계 없음; boundary grep 0(V9); vet/build 청정(V6/V7); closed-set 검증 유지(위반값 거부); substring 매칭은 과잉추론 방향으로 안전(F4). |
| Craft (20%) | **0.88** | RED→GREEN 사슬 이정표(§E.2.6), 이중 leg smoke, 표시-와이어 일치 테스트, overlay 함수 100% 커버리지. 감점: §E.3 파일수 오계(F2), config 패키지 80.6%(기존 부채, 명시됨). |
| Consistency (15%) | **0.93** | 상수 파생 closed set(리터럴 중복 없음), 템플릿 트윈·4-로캘 패리티, catalog 바이트 불변, 중립성 청정, Conventional Commits + 카드 id. |

**조화 평균 = 4 / (1/0.95 + 1/0.95 + 1/0.88 + 1/0.93) = 4 / 4.3169 ≈ 0.927**

## 7. 최종 판정

**Verdict: PASS** (0.927)

- must-pass 방화벽: Functionality PASS(8개 must-pass AC 전부 재실행 초록), Security PASS — 방화벽 미발동.
- blocking finding 0건 — F1–F5는 전부 optional/info 수준 (F1·F2는 기록 보강 권장).
- 3-phase close·backfill·CHANGELOG·catalog 불변 전부 독립 검증 완료.
