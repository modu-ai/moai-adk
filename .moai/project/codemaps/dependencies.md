# 패키지 의존도 분석

> 이 문서는 `/moai codemaps --force`로 자동 생성된 의존도 그래프입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4

---

## 의존도 그래프 (Mermaid)

```mermaid
graph TD
    cmd["cmd/moai<br/>main()"]
    
    subgraph P["Presentation Layer"]
        cli["internal/cli<br/>(261 non-test)<br/>61 root 등록"]
        tui["internal/tui"]
        statusline["internal/statusline"]
        web["internal/web"]
        version["pkg/version"]
    end
    
    subgraph B["Business/Domain Layer"]
        models["pkg/models<br/>(Very High Fan-in)"]
        foundation["internal/foundation"]
        spec["internal/spec"]
        workflow["internal/workflow"]
        loop["internal/loop"]
        harness["internal/harness"]
        permission["internal/permission"]
        constitution["internal/constitution"]
        merge["internal/merge"]
    end
    
    subgraph I["Infrastructure Layer"]
        coreGit["internal/core/git<br/>(High Fan-in)"]
        coreProject["internal/core/project"]
        coreQuality["internal/core/quality"]
        template["internal/template"]
        config["internal/config<br/>(Very High Fan-in)"]
        manifest["internal/manifest"]
        hook["internal/hook"]
        runtime["internal/runtime"]
        session["internal/session"]
        lsp["internal/lsp"]
        mx["internal/mx"]
        graph["internal/graph<br/>(edges + freshness + code queries)"]
        symbol["internal/graph/symbol<br/>(astx seam)"]
        codexwiring["internal/codexwiring"]
        codexadapter["internal/codexadapter"]
        ciwatch["internal/ciwatch"]
        mcp["internal/mcp<br/>(tool catalog)"]
        chain["internal/chain"]
    end
    
    cmd --> cli
    cli --> models
    cli --> config
    cli --> hook
    cli --> template
    cli --> session
    cli --> spec
    cli --> loop
    cli --> coreGit
    cli --> coreProject
    cli --> graph
    cli -->|codex init 게이트 위임| codexwiring
    cli --> ciwatch
    cli --> mcp
    cli --> chain
    
    config --> models
    template --> manifest
    spec --> constitution
    hook --> config
    hook --> graph
    graph --> mx
    graph --> symbol
    graph --> tiers["internal/navigator/tiers"]
    symbol --> mx
    symbol --> astx["internal/navigator/astx"]
    codexwiring --> codexadapter
    codexadapter --> hook
    workflow --> coreGit
    loop --> config
    harness --> config
    permission --> config
    coreGit --> foundation
    coreProject --> foundation
    lsp -.-> astgrep
    mx --> lsp
```

---

## 팬-인 분석 (Very High)

| 패키지 | 팬-인 수준 | 이유 |
|--------|----------|------|
| `pkg/models` | 45+ | Config 타입 중심 |
| `internal/config` | 48+ | CLI composition에서 모든 패키지에 주입 |
| `internal/cli` | (50+ import) | Composition root |
| `internal/core/git` | 35+ | workflow/spec/session 필수 |
| `pkg/version` | 30+ | 버전 출력 (CLI/help) |
| `internal/foundation` | 32+ | 언어 registry (모든 도메인) |

---

## 계층 간 의존도

**Presentation → Business → Infrastructure**

- `cli` → 모든 business 계층 + infrastructure
- `config` → models, defs (핵심 주입)
- `hook` → config, lsp, session, mx, graph
- `coreGit` → foundation

---

## 신규 인프라 패키지 (2026-07 기준)

| 패키지 | 역할 |
|---|---|
| `internal/goal` | 목표 엔진 — 조건 선언형 에이전틱 루프 (`/moai goal`) |
| `internal/lockfile` | 크로스 플랫폼 잠금 (Unix `flock(2)` / Windows in-process mutex) |
| `internal/atomicfile` | 원자적 파일 쓰기 (write-temp + rename) |
| `internal/tokenusage` | 토큰 사용량 계수 (statusline 연동) |
| `internal/verify` | 검증 서브시스템 |
| `internal/settings` | settings.json / settings.local.json 헬퍼 |
| `internal/graph` | 코드베이스 엣지 산출물(edges.jsonl) + 3-레이어 신선도 게이트(`moai graph check`)·쿼리 시점 갱신·인용 앵커·MCP 코드 쿼리 엔진 |
| `internal/graph/symbol` | 그래프 빌더의 astx 추출 시임 — navigator 계층 의존 없이 code-call/code-import 엣지 추출 (`go list -deps` 격리 검증) |
| `internal/codexwiring` | `moai init --agent codex`용 `.codex/hooks.json`·`config.toml` 생성기 (SPEC-CODEX-WIRING-001). `moai codex` 세 기동 형태(맨몸·`cli`·`app`)의 init-offer 게이트가 수락 시 `Wire()`를 위임받는다 (SPEC-CODEX-INIT-001) |
| `internal/codexadapter` | codex 하네스 어댑터 — codexwiring이 소비하는 공통 계층 |
| `internal/ciwatch` | CI 감시 루프 엔진 (scripts/ci-watch 백엔드) |
| `internal/mcp` | 자체 호스팅 MCP 서버의 도구 카탈로그 단일 선언 (28 도구) |
| `internal/chain` | 훅 이벤트 체인 코어 |

> 이 패키지들은 `merge`/`manifest`/`session` 등 기존 인프라와 협력합니다. `goal`은 self-contained (의존성 없음), `lockfile`은 `session`이, `atomicfile`은 `merge`/`manifest`가 사용합니다.

## 신규 인프라 패키지 (2026-08 기준)

| 패키지 | 역할 |
|---|---|
| `internal/guardstate` | 가드 발화 이벤트를 8행 상태표로 분류하는 상태 모델 (SPEC-GUARD-STATE-MODEL-001) — 닫힌 7값 `Classification` 어휘, `Classify()`/`Evaluate()`/`Produce()` |
| `internal/guardliveness` | guardstate 판정을 운영자에게 표면화하는 계층 (SPEC-GUARD-LIVENESS-001) — 3-clause 계약만 소비, 어휘는 소유하지 않음. `internal/guardstate`가 이 패키지의 `Entry`/`Store`를 참조 |
| `internal/binlag` | 설치된 `moai` 바이너리의 commit lag 판정 (SPEC-BINARY-LAG-VISIBILITY-001) — `moai doctor`와 SessionStart 어드바이저가 같은 구현을 공유 |

> `guardstate` → `guardliveness` 참조는 `internal/hook/session_start_guard_liveness.go`가 양쪽을 조립하는 지점에서만 성립 — 두 패키지 사이에 직접 순환 의존은 없다 (`guardstate/produce.go`가 `guardliveness` 타입을 반환값으로 참조).

---

## 순환 의존성 검증

**결과**: 0개 순환 의존성 (검증됨)

---

**생성**: `/moai codemaps --force`로 자동 생성
