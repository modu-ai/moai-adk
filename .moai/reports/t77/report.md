# t77 — GLM 세션 과금 경로 실증 (전제검증 카드)

> 카드 원문 (backlog t77, picked): "GLM lead 과금 실증 [최우선·전제검증] — moai glm 세션에서 Agent 2~3개 스폰 후 z.ai 대시보드에서 토큰 출처 확인 + Claude 사용량 무증가 확인. 코드 변경 0. 코드 근거는 리드 확인 완료(glm.go:278 setGLMEnv가 ANTHROPIC_AUTH_TOKEN/BASE_URL을 프로세스 env 주입 → 스폰 Agent가 상속해 z.ai 라우팅) — 대시보드 실측이 최종 판정."
>
> 디스패치(tjv7iy 라운드4) 판정 기준: 측정 설계의 타당성 + 증거 재현성.

## 판정 요약

| 항목 | 결과 |
|---|---|
| 코드 근거 (setGLMEnv env 주입) | 확인 — `internal/cli/glm.go:336-364` |
| 런타임 env 실측 (live claude 프로세스) | 확인 — `ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic`, 토큰 SET(값 미출력) |
| 라이브 실험 (Agent 3개 × 2회) | 재현 성공 — 런 A/B 모두 3/3 완료, 사용량 전량 `modelUsage["glm-5.3"]`에 집계, claude-* 모델 항목 0건 |
| Claude측 무증가 | 레인 세션 0건(기계 확인) + 창구 내 lead 스폰 2건도 GLM 백엔드(lead 회신) — 최종은 콘솔 확인 |
| z.ai 대시보드 출처 확인 | 공개 usage API 부재(문서 확인) → **운영자 체크리스트로 잔존** (카드가 명시한 최종 판정) |

**결론(현 시점)**: 카드 전제("GLM 세션에서 스폰한 Agent는 z.ai 과금")를 반증하는 관측은 0건. 카드가 지정한 최종 판정 기준인 z.ai 대시보드 실측만 운영자 몫으로 남는다.

## 측정 설계

과금 경로는 "어느 API 엔드포인트·크리덴셜로 요청이 나가는가"로 실측한다. Claude Code의 서브에이전트는 별도 프로세스가 아니라 **같은 프로세스의 추가 API 호출**이므로, 세션 프로세스의 env(base URL + auth)가 곧 서브에이전트 트래픽의 라우팅이다. 이를 4겹으로 관측했다:

1. **코드**: setGLMEnv이 프로세스 env에 무엇을 심는지 (소스)
2. **런타임 env**: 살아 있는 claude 프로세스의 실제 환경 (`ps eww` — 비밀값은 존재만 확인)
3. **사용량 집계**: 세션 결과 JSON의 `modelUsage` — 서브에이전트 호출이 어느 모델 버킷에 합산되는지
4. **Claude측 대조**: 실험 창구에서 Claude 과금 세션의 Agent 스폰 부재

## 환경

- 실험장: `/tmp/t77t64-lab` — `moai init . --force` 후 운영 체크아웃의 `llm.yaml` 미러(GLM 전 슬롯 glm-5.3). CLAUDE.local.md §13(=dev 프로젝트에서 `moai glm` 금지) 준수를 위해 전 과정을 /tmp 고립 프로젝트에서 실행.
- moai 바이너리: **설치본 v3.1.0 (commit d6b80a01c, built 2026-08-16)** — t65 수정(225a51e24) 포함을 `git merge-base --is-ancestor`로 확인. 별도 빌드 없음 = 배포 상태 그대로 측정.
- claude CLI: v2.1.233. 측정 창구: 2026-08-17 03:44~04:25 KST.

## 절차 (재현)

```bash
# 실험장 (1회)
mkdir -p /tmp/t77t64-lab && cd /tmp/t77t64-lab
~/go/bin/moai init . --force
cp /Users/goos/MoAI/moai-adk-go/.moai/config/sections/llm.yaml .moai/config/sections/llm.yaml

# 본실험 (headless GLM 세션 — Agent 3개 스폰 지시)
cd /tmp/t77t64-lab && ~/go/bin/moai glm --print '<실험 프롬프트>' --output-format json
# raw/run-a.txt, raw/run-b.txt — 프롬프트 전문 포함

# 런타임 env 채집 (세션 실행 중)
lsof -p <pid> | awk '$4=="cwd"'        # cwd로 랩 세션 pid 특정
ps eww -o command= -p <pid> | grep -oE 'ANTHROPIC_BASE_URL=[^ ]+|...'   # 비밀키 외 키만 선택 출력
ps eww -o command= -p <pid> | grep -c 'ANTHROPIC_AUTH_TOKEN='           # 존재만 카운트(값 미출력)

# 트랜스크립트 기계 확인
grep -c '"name":"Agent","input"' <세션 트랜스크립트>.jsonl
grep -o '"subagent_type":"[^"]*"' <세션 트랜스크립트>.jsonl | sort | uniq -c
```

**호출 형태 주의(실측 발견)**: 디스패치 리터럴 `moai glm -p "<프롬프트>"`는 오동작한다. `-p`는 moai 자신의 **프로필 플래그**(`internal/cli/launcher.go:808`)이고, `--` 뒤에 붙여도 `--` 토큰이 claude argv에 새어 들어가 뒤 인자 전부가 위치 인자로 취급된다(증거: `raw/run0-dispatch-literal-form-attempt.txt` — 세션이 프롬프트를 받지 못하고 "-p"의 의미를 묻는 응답). 작동 형태는 **`moai glm --print '<프롬프트>'`** (claude 장평 플래그, moai 파서가 간섭하지 않음).

## 증거

### 1. 코드 근거

`internal/cli/glm.go:336-364` `setGLMEnv` — 프로세스 env에 `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`(기본 `https://api.z.ai/api/anthropic`, `internal/config/defaults.go:112`), `ANTHROPIC_DEFAULT_*_MODEL`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS` 등을 주입. 런처는 POSIX에서 `syscall.Exec`으로 claude를 exec하므로(launcher.go:614 주석) 자식 claude가 이 env를 그대로 상속한다.

### 2. 런타임 env (live claude 프로세스, pid 78144, cwd=/private/tmp/t77t64-lab)

```
ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic
ANTHROPIC_DEFAULT_FABLE_MODEL=glm-5.3
ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-5.3
ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.3
ANTHROPIC_DEFAULT_SONNET_MODEL=glm-5.3
CLAUDE_CODE_AUTO_COMPACT_WINDOW=1000000
CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1
CLAUDE_CODE_MAX_CONTEXT_TOKENS=1000000
ANTHROPIC_AUTH_TOKEN=<SET — 존재만 확인, 값 미출력>
Z_AI_API_KEY=<SET — 존재만 확인, 값 미출력>
```

### 3. 라이브 실험 — 런 A/B (각 3 Explore 에이전트, 단일 메시지 동시 스폰)

| | 런 A (session 7fc337cc) | 런 B (session 6929bac8) |
|---|---|---|
| 결과 | 391 / `/private/tmp/t77t64-lab` / OK — 3/3 완료 | 144 / 3 / OK — 3/3 완료 |
| 트랜스크립트 `Agent` tool_use | 3회 (전부 `subagent_type:"Explore"`) | 3회 (전부 `subagent_type:"Explore"`) |
| modelUsage 버킷 | **glm-5.3 단일** (input 51,053 + cache-read 493,504 + output 1,471) | **glm-5.3 단일** (input 80,282 + cache-read 187,520 + output 1,209) |
| claude-* 모델 항목 | **0건** | **0건** |
| costUSD | 0.5388 | 0.5254 |
| 세션 자가 보고 | "런타임은 z.ai GLM 백엔드로 감지됨, glm-5.3" | "glm-5.3으로 실행 중" |

원시 출력: `raw/run-a.txt`, `raw/run-b.txt` (프롬프트 전문 + JSON 전체). 프로브(`raw/probe-print-passthrough.txt`) 포함 실험 총 z.ai측 비용 ≈ **$1.55** (0.4858 + 0.5388 + 0.5254; 런0의 비용은 JSON 미출력으로 미계산).

### 4. Claude측 무증가

- 본 레인 세션(6f141c53): 트랜스크립트에서 `"name":"Agent"` **0건** (grep, 전 시간대).
- lead 세션(lead-tjv7iy, 창구 내 회신): 창구 내 Agent 스폰 **2건**(04:00:27 Worker D / 04:00:33 Worker E, task output birth time 기계 확인) — 단 **lead 세션 자체가 GLM 백엔드**(glm-5.3)이므로 이 2건 역시 z.ai 과금이며 Claude 소모 0. 기록 문구: "lead 세션 스폰 2건(04:00, 기계 확인) — lead 백엔드 GLM로 Claude 소모 0".
- `moai session list --json` = `[]` (03:44 실측) — 등록된 병행 세션 없음.

### 5. z.ai 출처 — 운영자 체크리스트 (공개 API 부재)

z.ai 개발자 문서(docs.z.ai)의 카탈로그를 열거함: **API Keys / Payment Method / Pricing / API Reference / Scenario Example / Coding Plan / Released Notes / Terms / Help Center** — 사용량 조회 공개 API는 없고 빌링은 콘솔 Billing Page뿐. 따라서 최종 판정은 다음 체크리스트로:

1. z.ai 콘솔 → Billing/사용량 페이지에서 **2026-08-17 04:00~04:15 KST** 부근 glm-5.3 호출 확인. 기대치: 합산 ≈ $1.55 (프로브 0.4858 + 런A 0.5388 + 런B 0.5254 + 런0 소액), 세션별 input ~80-97K + cache-read ~187-493K 토큰.
2. Anthropic/claude.ai 사용량 페이지에서 동일 창구 **실험 귀속 증가 부재** 확인 (find 2개 신규 파일 커밋 외 트래픽 없음).
3. 1~2가 카드 전제와 어긋나면 t77 FAIL → t78/t79/t84~t86 후속 전면 재검토 (카드 지시).

## Gaps (미검증)

- z.ai 콘솔·Anthropic 콘솔 직접 확인 — 운영자 몫(위 체크리스트). 이것이 카드가 규정한 최종 판정이다.
- headless 런 A/B의 claude 프로세스 env는 종료 후 채집 불가 — 인터랙티브 세션(pid 78144)에서 대표 채집. 런처 경로가 동일(binary·cwd·config)하므로 동등성은 추론이지만, 사용량 JSON의 glm-5.3 단일 집계가 이를 독립적으로 뒷받침.
- 런0(리터럴 형태 시도)의 정확한 비용 미계산(JSON 미출력).
- 세션이 스스로 "신뢰성 제한적"이라 단서를 단 **"15,000,000 tokens" 예산 표시**의 출처는 규명하지 않음(창 크기 1M과는 별개 필드, 표시 전용으로 관측 — t64 리포트 참조).

## 잔여 위험

- CC의 `modelUsage` 집계가 z.ai 청구 항목과 1:1이 아닐 수 있음(캐시 판독 과금 단가 등) — 체크리스트 1번에서 실제 청구액과 대조할 것.
- z.ai 동시성 제한(1-3 in-flight) — 본 실험 3동시 스폰은 통과했으나 t85 팩토리 N=8 구성에서는 오류 표면이 다를 수 있음(glm.go:91-95 경고 참조).
