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
        cli["internal/cli<br/>(264 non-test)<br/>60 root 등록"]
        tui["internal/tui"]
        statusline["internal/statusline"]
        web["internal/web"]
        version["pkg/version"]
    end
    
    subgraph B["Business/Domain Layer"]
        models["pkg/models"]
        foundation["internal/foundation"]
        spec["internal/spec"]
        workflow["internal/workflow"]
        loop["internal/loop"]
        harness["internal/harness"]
        navigator["internal/navigator<br/>(BAS 6 서브패키지)"]
        kanban["internal/kanban"]
        permission["internal/permission"]
        constitution["internal/constitution"]
        merge["internal/merge"]
    end
    
    subgraph I["Infrastructure Layer"]
        coreGit["internal/core/git"]
        coreProject["internal/core/project"]
        coreQuality["internal/core/quality"]
        template["internal/template"]
        config["internal/config<br/>(High Fan-in)"]
        defs["internal/defs"]
        paths["internal/paths"]
        atomicfile["internal/atomicfile"]
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
        settings["internal/settings<br/>(agentfm, yamlpatch)"]
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
    cli --> navigator
    cli --> kanban
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

## 팬-인 분석 (High)

측정: `go list -f '{{range .Imports}}'` 기반 직접 non-test 임포트 수 (2026-09-02 실측).

| 패키지 | 팬-인 | 이유 |
|--------|------|------|
| `internal/config` | 27 | CLI composition에서 모든 패키지에 주입 |
| `internal/defs` | 17 | 디렉토리 레이아웃 상수 (전 계층 참조) |
| `internal/paths` | 11 | `~/.moai` 경로 단일 해석점 |
| `internal/atomicfile` | 11 | 원자적 쓰기 프리미티브 (merge/manifest/config 공용) |
| `internal/lsp` | 10 | LSP 클라이언트 집합 (mx/hook/cli 소비) |
| `pkg/models` | 9 | Config 타입 중심 |
| `internal/tui` | 9 | 터미널 UI 컴포넌트 (cli/web/statusline) |
| `internal/template` | 9 | 배포 엔진 |
| `internal/harness` | 8 | 하네스 학습 (cli/hook 소비) |
| `internal/execerr` | 8 | 서브프로세스 실패 래퍼 |
| `internal/cli` | 팬-아웃 93 (내부 패키지 임포트) | Composition root |

---

## 계층 간 의존도

**Presentation → Business → Infrastructure**

- `cli` → 모든 business 계층 + infrastructure (내부 패키지 93개 임포트)
- `config` → models, defs (핵심 주입)
- `hook` → config, lsp, session, mx, graph
- `coreGit` → foundation
- `graph` → mx, navigator/{tiers,astx} (신선도 게이트·엣지 추출이 navigator AST 시임 사용)

---

## 신규 인프라 패키지 (2026-07 기준)

| 패키지 | 역할 |
|---|---|
| `internal/goal` | 목표 엔진 — 조건 선언형 에이전틱 루프 (`/moai goal`) |
| `internal/lockfile` | 크로스 플랫폼 잠금 (Unix `flock(2)` / Windows in-process mutex) |
| `internal/atomicfile` | 원자적 파일 쓰기 (write-temp + rename) |
| `internal/tokenusage` | 토큰 사용량 계수 (statusline 연동) |
| `internal/verify` | 검증 서브시스템 (공유 진단 스냅샷) |
| `internal/settings` | settings.json / settings.local.json 헬퍼 |
| `internal/graph` | 코드베이스 엣지 산출물(edges.jsonl) + 3-레이어 신선도 게이트(`moai graph check`)·쿼리 시점 갱신·인용 앵커·MCP 코드 쿼리 엔진 |
| `internal/graph/symbol` | 그래프 빌더의 astx 추출 시임 — navigator 계층 의존 없이 code-call/code-import 엣지 추출 (`go list -deps` 격리 검증) |
| `internal/codexwiring` | `moai init --agent codex`용 `.codex/hooks.json`·`config.toml` 생성기. `moai codex` 세 기동 형태(맨몸·`cli`·`app`)의 init-offer 게이트가 수락 시 `Wire()`를 위임받는다 |
| `internal/codexadapter` | codex 하네스 어댑터 — codexwiring이 소비하는 공통 계층 |
| `internal/ciwatch` | CI 감시 루프 엔진 (scripts/ci-watch 백엔드) |
| `internal/mcp` | 자체 호스팅 MCP 서버의 도구 카탈로그 단일 선언 |
| `internal/chain` | 훅 이벤트 체인 코어 |

> 이 패키지들은 `merge`/`manifest`/`session` 등 기존 인프라와 협력합니다. `goal`은 self-contained (의존성 없음), `lockfile`은 `session`이, `atomicfile`은 `merge`/`manifest`가 사용합니다.

## 신규 인프라 패키지 (2026-08 기준)

| 패키지 | 역할 |
|---|---|
| `internal/guardstate` | 가드 발화 이벤트를 8행 상태표로 분류하는 상태 모델 (SPEC-GUARD-STATE-MODEL-001) — 닫힌 7값 `Classification` 어휘, `Classify()`/`Evaluate()`/`Produce()` |
| `internal/guardliveness` | guardstate 판정을 운영자에게 표면화하는 계층 (SPEC-GUARD-LIVENESS-001) — 3-clause 계약만 소비, 어휘는 소유하지 않음. `internal/guardstate`가 이 패키지의 `Entry`/`Store`를 참조 |
| `internal/binlag` | 설치된 `moai` 바이너리의 commit lag 판정 (SPEC-BINARY-LAG-VISIBILITY-001) — `moai doctor`와 SessionStart 어드바이저가 같은 구현을 공유 |

> `guardstate` → `guardliveness` 참조는 `internal/hook/session_start_guard_liveness.go`가 양쪽을 조립하는 지점에서만 성립 — 두 패키지 사이에 직접 순환 의존은 없다 (`guardstate/produce.go`가 `guardliveness` 타입을 반환값으로 참조).

## 신규 패키지 (2026-09 기준)

| 패키지 | 역할 |
|---|---|
| `internal/navigator` (astx/detect/fix/route/sync/tiers) | BAS(Blueprint-Anchored Synchronization) 코드 탐색 — regen/audit/sync 3체인 통합, 4-tier 주소 맵 오버레이(tiers.json) |
| `internal/kanban` | 칸반 백로그 큐 엔진 — `moai todo` 명령 백엔드 (기계적 텍스트 분석기 포함) |
| `internal/epic` | 디스크 기반 에픽 진행 생산자 (`moai epic status`) |
| `internal/paths` | `~/.moai` 디렉터리 트리 단일 해석점 |
| `internal/execerr` | 서브프로세스 실패의 출력 보존 래퍼 |
| `internal/sessionmsg` | 단일 머신 세션 메시징 브로커 코어 (Claude↔Codex) |
| `internal/glmcred` | GLM 자격증명(`~/.moai/.env.glm`) 쓰기·판독 단일 구현 |
| `internal/feedback` | 피드백 제출 실패 리트라이 큐 |
| `internal/timing` | 테스트용 보정 지연 상한 (코드 측정 vs 머신 측정 구분) |
| `internal/report/planhtml` | plan-phase HTML 리포트 렌더러 |
| `internal/mirrornotice` | 템플릿 배포 스킬 미러 결과 공지 |
| `internal/template/agentemit` | 에이전트 출력 에밋 |
| `internal/config/toolpolicy` | 도구/권한 정책 SSOT (`moai tool-policy`) |

> 참고: `internal/bodp`(Branch Origin Decision Protocol)는 worktree 표면 리디자인(#1278)에서 코드베이스에서 제거되었다 — 현재 트리에 존재하지 않는다.

---

## 순환 의존성 검증

**결과**: 0개 순환 의존성 (검증됨 — `go build ./...` 성공이 기계적 증거; Go 컴파일러가 import cycle을 금지)

---

**생성**: `/moai codemaps --force`로 자동 생성
