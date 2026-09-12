# M5·PICKER 계획 변경분 감사

Iteration: 1 — M5 문서 결정과 PICKER 사전 측정 준비만 판정
Verdict: PASS
Overall Score: 0.86

Reasoning context ignored per M1 Context Isolation.

## Claim

일반 text·완료된 tool pair·SSE 변환의 작성된 계약은 구현을 시작할 수 있다. PICKER는 문서에 정한 사전 측정을 시작할 수 있다. opaque reasoning의 제품 경로와 PICKER의 최종 구현·호환성은 아직 통과하지 않았다. 이 구분 아래에서 이번 단계의 차단 결함은 찾지 못했으며 기존 검증을 구체화할 선택 개선 두 건을 남긴다.

읽기 전에 둔 실패 가설은 system 위치·optional schema 훼손, tool ID/이름의 요청 간 오염, 실패 SSE를 성공으로 닫는 판정, PROBE ONLY의 제품 활성화, picker overlay를 저장 격리 성공으로 착각하는 경우, Default·정책 제한의 미관측이었다.

## Must-Pass Results

- **MP-1 PASS:** PICKER REQ 001~009가 연속이고 AC 9개가 동일 번호 REQ를 추적한다(E2). 코어는 새 REQ/AC 번호를 도입한 감사가 아니다.
- **MP-2 PASS:** PICKER 요구사항 계층은 When 또는 The … shall 형태를 사용한다(`spec.md:30-48`). 검증 계층 AC는 Given-When-Then이다. 코어 M5는 기존 REQ-013·015의 설계/검증 구체화다(`design.md:361-409`, `acceptance.md:154-173`).
- **MP-3 PASS:** PICKER `spec.md:2-15`에서 canonical frontmatter 12필드, `version: "0.1.0"`, `status: draft`, `tier: M`을 직접 읽었다. 두 SPEC lint는 E1.
- **MP-4 N/A:** 단일 Go 저장소 CLI/gateway 계약이며 언어 공통 template 변경이 아니다.
- **MP-5 PASS, 변경분:** PICKER의 `related_specs`와 범위 문장은 현재 존재하는 코어를 참조하며 코어 계약을 다시 정의하지 않는다(`spec.md:15`, `:54-56`; `plan.md:5-6`). AUTH의 실제 충족은 `plan.md:65-66`에서 미완료로 남긴다. 형제의 구현 완료를 이번 판정 근거로 삼지 않았다.
- **MP-6 PASS, 변경분:** PICKER `spec.md`의 syscall 언급은 0개(E2). 코어의 이번 §4.2·4.3에는 syscall 도입이 없다.
- **MP-7 PASS:** PICKER plan의 unresolved clarification 표지는 0개(E2). 정한 사전 측정·PROBE ONLY 게이트는 미측정 상태 그대로다. 표지 부재를 그 게이트의 완료로 보지 않았다.
- **MP-8 N/A:** 이번 대상에는 release-blocking RED-now 명령 인용 셀이 없다. 예정된 실제 TUI·계정 시험을 재실행한 것으로 기록하지 않는다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 0.75 | 0.75 | 일반 변환과 PROBE ONLY는 분명함(`core design.md:375-409`). 혼합 정상/미완료 출력의 오류 우선순위를 더 직접 적을 여지(A1) |
| Completeness | 1.00 | 1.00 | system 위치·schema·종료·tool 복원과 picker 사전 측정/격리/정책 제한을 포함(`core design.md:367-388`; `picker plan.md:13-44`) |
| Testability | 0.75 | 0.75 | 응답·계수·파일·TUI 대조가 명시됨. 제한된 명시 startup argv의 독립 행은 없음(A2) |
| Traceability | 1.00 | 1.00 | PICKER 9/9 동일 번호 추적(E2), M5 기존 AC별 보강(`core acceptance.md:164-173`) |

조화평균 `4 / (1/0.75 + 1/1 + 1/0.75 + 1/1) = 0.857142…`, 표기 0.86. 코어 Tier L 0.85·PICKER Tier M 0.80 기준보다 높다. 이 감사는 새로운 제한 범위의 iter1이며 기존 전체 코어 감사 점수와 회귀 비교하지 않는다.

## Defects Found

D1. **M5P-A1 — 혼합 출력의 오류 우선순위와 fixture를 명확히 할 여지** — `SPEC-MOAI-GATEWAY-001/design.md:382-385`, `acceptance.md:165-167` — Severity: minor — Class: optional — Confidence: high (문서 판독).

종료 표에는 유효한 완료 text/tool이 있으면 정상 종료로 옮기는 행과 깨진/미완료 블록을 오류로 처리하는 행이 함께 있다. AC는 각각을 따로 시험한다. `완료 text + 미완료 function call`, 또는 `완료 tool call + 깨진 두 번째 call`처럼 섞인 응답도 오류 행이 우선한다는 문장과 fixture를 기존 AC에 붙이면 table-first-match 오해를 직접 잡을 수 있다. 현재 계약 자체가 미완료 블록의 성공 처리를 금지하므로 일반 변환기 착수를 막지는 않는다. 실제 upstream에서 이 혼합 사례를 관측했다는 주장은 아니다.

D2. **M5P-A2 — 정책에 막힌 명시 startup 모델을 별도 사전 측정 입력으로 추가할 여지** — `SPEC-MOAI-GATEWAY-PICKER-001/plan.md:35-41`, `acceptance.md:24-30` — Severity: minor — Class: optional — Confidence: high (공식 문서와 대상 계획 판독; 실제 client 미실행).

현재 표는 picker 행의 allowlist/managed 제한과 Default 후보 제한을 본다. 공식 model-config는 차단된 명시 `--model`도 시작 시 경고와 함께 기본 모델로 대체할 수 있다고 설명한다. 따라서 현재의 새 실행/재개 측정에 `--model=<정책상 제외된 GPT ID>`를 넣고 표시·첫 turn 모델·upstream 계수를 함께 기록하면 명시 argv만 확인하는 오판을 막을 수 있다. 기대 계약은 이미 PICKER REQ-004·008의 명시 거절/fallback 금지에 있으므로 새 REQ/AC나 새 gate를 만들 필요는 없다. 사전 측정을 시작하는 데 필요한 새 운영자 결정도 아니다. [공식 모델 설정 문서](https://code.claude.com/docs/en/model-config#restrict-model-selection).

## 범위별 판정

| 대상 | 이번 판정 | 근거와 한계 |
|---|---|---|
| 최상위 system 및 대화 내 system | 구현 착수 가능 | `core design.md:367-370`이 instructions 결합과 동일 위치 메시지를 구분하고 미지원 형태는 거절. 실제 provider 수용은 별도 Gap |
| optional schema·tool 이름·call pair | 구현 착수 가능 | `:375-378`이 strict false·필드 의미·요청별 매핑·순서/ID 보존을 정하고 `core acceptance.md:160-167`이 golden/동시 요청 판정을 연결 |
| SSE·stop reason | 구현 착수 가능 | `:380-388`과 AC-MG-007·008 보강이 정상/오류 terminal을 구분. A1을 반영한 혼합 fixture 보강 가능 |
| opaque reasoning 운반 | PROBE ONLY, 제품 BLOCKED | `:390-408`, `core acceptance.md:169-173`에 전체 carrier 유실 탐지 계약과 실제 왕복 게이트가 미결임을 명시. 이번 계획 PASS로 해제하지 않음 |
| PICKER overlay·discovery·Default | 사전 측정 착수 가능 | `picker plan.md:13-44`가 후보·공식 제한·positive/negative 조건·출력 위치를 명시. 실제 TUI 성공은 아직 없음 |
| 설정 저장 격리·인증 유지 | 사전 측정 대상, 미해결 | `picker plan.md:27-29`, `acceptance.md:32-33`이 읽기 overlay와 실제 쓰기 목적지를 구분하며 credential 복사 우회를 금지 |

## Evidence

### E1 — 이번에 직접 실행한 lint

두 명령의 작업 디렉터리는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`다.

```text
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
$ /tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-PICKER-001
✓ No findings — all SPEC documents are valid
```

각 exit code 0. 같은 HEAD에서 앞서 빌드된 바이너리를 사용했으며 이번 감사에서 재빌드하지 않았다. lint는 구조 확인이다.

### E2 — 정의·추적·표지·내용 고정

Read-only `python3 - <<'PY'`에서 PICKER `spec.md`의 `^\*\*REQ-GP-(\d{3})\*\*`, acceptance의 동일 AC 정의·헤더 REQ를 추출하여 순서와 집합을 대조했다. syscall·clarification·release-blocking 문자열 개수와 읽은 산출물 SHA-256을 출력했다. 관측 stdout(exit 0):

```text
picker_REQ=9 sequential=True
picker_AC=9 trace_equal=True
syscall=0 clarification=0 release_blocking=0
.moai/specs/SPEC-MOAI-GATEWAY-PICKER-001/spec.md 3a6a1acac946b86d4b837c9679e231bd089e52d7090ac7a59daf1804201adfda
.moai/specs/SPEC-MOAI-GATEWAY-PICKER-001/plan.md 0234d34577040ce668e1c97e56b5231186b1a00a5bc8a0f5d120f931f0389088
.moai/specs/SPEC-MOAI-GATEWAY-PICKER-001/acceptance.md f297813c30187ffa3bf80ad6fae9764ff3db758c47275e5db079dd437ab21ae3
.moai/specs/SPEC-MOAI-GATEWAY-001/design.md d2e6721a8826516900e68f075949f46f4815356a024339e195fe63e59e4eaf7c
.moai/specs/SPEC-MOAI-GATEWAY-001/acceptance.md ec87ebfa467f416f1b9877ea41e2d78a494958d1d7fd8ce9b9bf8be72f0df8f3
.moai/reports/SPEC-MOAI-GATEWAY-001/m5-design-decisions.md e11e95ef78cfcb4ed4769ddcbe6505147f86cd67398dda7627af1df74a242ba7
```

### E3 — 공식 문서 판독

- `/tmp/gateway-claude-settings-reference-20260911.md:1064-1110`을 직접 읽었다. modelPicker의 `--settings` 적용, 프로젝트/local 무시, built-in 교체 뒤 Default/현재 행 존속, 정책 제한에 따른 행 제거가 PICKER 계획에 반영되어 있다. [공식 설정 참조](https://code.claude.com/docs/en/settings-reference#modelpicker).
- 공식 function-calling 문서를 web 도구로 열고 strict 설명을 확인했다. 관측한 짧은 원문: `explicitly set strict: false`. Responses의 자동 schema 처리에서 벗어나 기존 optional 의미를 보존하려는 결정의 문서 근거다. 실제 API 수용 시험은 아니다. [OpenAI function calling](https://developers.openai.com/api/docs/guides/function-calling).
- 공식 gateway protocol을 web 도구로 열어 discovery가 ID의 claude/anthropic 부분문자열로 필터링됨을 확인했다. bare GPT discovery에 의존하지 않는 `picker plan.md:20-23`과 일치한다. [Gateway compatibility](https://code.claude.com/docs/en/llm-gateway-protocol#model-discovery).
- 공식 model-config를 web 도구로 열어 Default 변수의 정책·계정 제약과 blocked startup 대체를 확인했다. 관측한 짧은 원문: `Claude Code replaces the value at startup with a warning`. A2는 이 문서 동작을 실제 대상 버전에서 측정하라는 지적이며, 여기서 실제 대체를 관측하지 않았다. [Model configuration](https://code.claude.com/docs/en/model-config).

## Baseline-attribution

`git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse --short HEAD` → `81c1d58f9`, exit 0. 대상은 코어 0.8.0의 M5 변경 절·결정 보고서·해당 AC 보강과 새 PICKER 0.1.0의 세 문서다. 내용 해시는 E2. 기존 코어 전체를 재감사하지 않았다.

`git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified check-ignore .moai/reports/SPEC-MOAI-GATEWAY-001/m5-picker-plan-audit-iter1.md` → stdout 빈 출력, exit 1. 추적 가능 보고서 경로다. 이번 작업은 이 보고서만 작성했으며 제품 코드·SPEC·Git 상태를 변경하지 않았다.

## Gaps

실제 Claude TUI·실서비스·계정 가용성·Responses upstream·일반 변환기 코드·뮤턴트 시험을 실행하지 않았다. 공개 system 입력의 실제 API 수용은 여기서 확정하지 않았다. M0 INCONCLUSIVE는 유지되며 opaque carrier 부재 탐지와 picker 저장 격리는 미결이다. 사용자 지시의 Claude 시험 시각인 2026-09-11 19:00 Asia/Seoul 전 LIVE 호출을 하지 않았다.

## Residual-risk

일반 변환기 단위시험이 통과하더라도 reasoning 가능한 네 GPT 모델의 실제 도구 왕복을 보증하지 않는다. picker 행이 보여도 저장 격리·Default·정책 제한·재개 초기 모델이 맞는지는 별도 측정해야 한다. A1·A2를 기존 fixture와 측정 행에 구체화하면 이 두 오판 가능성을 줄일 수 있다.

## Recommendation

일반 translator 구현과 PICKER preflight 준비를 진행할 수 있다. opaque reasoning 제품 활성화와 PICKER 최종 완료는 해당 실측 게이트가 충족될 때 별도로 판정한다. 이번 PASS를 전체 GPT 목표의 완료 또는 실제 client 호환성으로 보고해서는 안 된다.
