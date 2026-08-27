# 주요 데이터 흐름

> 이 문서는 `/moai codemaps --force`로 자동 생성된 데이터 흐름 설명입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4

---

## 1. 템플릿 배포 (moai init / moai update)

```mermaid
flowchart TD
    A["EmbeddedTemplates()"]
    B["Deployer.Deploy()"]
    C["Renderer.Render<br/>(strict missingkey=error)"]
    D["atomic write<br/>(temp+rename)"]
    E["Manifest.Track<br/>(3중 해시)"]
    F["3-way merge<br/>(사용자 커스터마이징 보존)"]
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
```

**흐름**:
1. `EmbeddedTemplates()` 파일시스템 로드
2. `Deployer.Deploy()` 렌더링 시작
3. `Renderer` TemplateContext와 함께 template 렌더 (엄격 모드)
4. 원자적 쓰기 (임시 파일 + 이름 바꾸기)
5. `Manifest.Track()` 3중 해시 기록 (template/deployed/current)
6. `update` 시에만: 3-way 병합 (사용자 변경사항 보존)

---

## 2. SPEC 라이프사이클 (moai plan/run/sync)

```mermaid
flowchart TD
    A["CLI /moai plan/run/sync"]
    B["spec.Linter<br/>(13+3 규칙)"]
    C["spec.ClassifyEra<br/>(H-1..H-6)"]
    D["spec.Audit<br/>(SyncStatusDrift)"]
    E["spec.ClassifyPRTitle<br/>(git 유추 상태)"]
    F["status enum<br/>transition"]
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
```

**흐름**:
1. CLI `/moai plan` → manager-spec 위임
2. CLI `/moai run` → manager-develop 위임 (TDD/DDD)
3. CLI `/moai sync` → manager-docs 위임
4. `Linter` frontmatter + ownership 검증 (13 단일 + 3 크로스 규칙)
5. `ClassifyEra()` grandfather (V2.x/V3R2-R4/V3R5) vs modern (V3R6)
6. `Audit()` drift 감지 (SyncStatusDrift 유일 차원)
7. `ClassifyPRTitle()` git 히스토리로부터 상태 추론 (50-commit 윈도우)
8. 상태 전환 기록 (status enum: draft→planned→in-progress→implemented→completed|superseded|archived|rejected)

---

## 3. 훅 이벤트 분배 (moai hook <event>)

```mermaid
flowchart TD
    A["Claude Code emits event"]
    B["handle-event.sh wrapper"]
    C["moai hook event<br/>(stdin JSON)"]
    D["Registry.Dispatch<br/>(30 EventType handlers)"]
    E["Handler chain<br/>(sequential)"]
    F["first 2-error<br/>short-circuit"]
    G["JSON + exit-code<br/>(stdout)"]
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
```

**흐름**:
1. Claude Code hook event 발생
2. `.claude/hooks/moai/handle-<event>.sh` 래퍼 실행
3. `moai hook <event>` stdin 통해 JSON 수신
4. `Registry.Dispatch()` 중앙 허브가 30개 EventType 핸들러로 라우팅
5. 핸들러 체인 순차 실행 (각각 exit 0 or 2)
6. 첫 번째 2-exit (차단) 시 short-circuit
7. JSON 결과 + exit-code stdout으로 반환

**PostToolUse 예시** (8개 핸들러):
- AST 스캔 (@MX 검증)
- 커버리지 계산
- cache telemetry
- harness observer
- 품질 게이트
- 권한 해석
- 진화 기록
- 생명주기 상태

---

## 4. Ralph 진단 루프 (moai loop)

```mermaid
flowchart TD
    A["LoopController.Start<br/>(상태 머신)"]
    B["FeedbackGenerator<br/>(go test + LSP)"]
    C["RalphEngine.Decide<br/>(우선순위)"]
    D["Continue<br/>또는"]
    E["Converge<br/>또는"]
    F["Abort<br/>또는"]
    G["HumanReview"]
    
    A --> B
    B --> C
    C --> D
    D --> B
    C --> E
    C --> F
    C --> G
```

**흐름**:
1. `LoopController.Start()` 진단 루프 시작 (상태 머신 + goroutine)
2. `FeedbackGenerator` 진단 정보 수집:
   - `go test ./...` 실행
   - LSP 진단 aggregate
   - `go vet`, 커버리지 분석
3. `RalphEngine.Decide()` 의사결정:
   - 우선순위: max_iter > perfect_gate > stagnation > human_review
   - Continue: 다음 반복 시작
   - Converge: 게이트 통과, 완료
   - Abort: 오류, 중단
   - HumanReview: 사용자 개입 필요

---

## 5. 권한 해석 (PreToolUse hook)

```mermaid
flowchart TD
    A["PreToolUse hook"]
    B["permission.Resolver.Resolve<br/>(8-tier)"]
    C["Tier 1: policy"]
    D["Tier 2-7: project→user→team<br/>→builtin→systemDefault<br/>→hookOverride"]
    E["Tier 8: builtinDeny"]
    F["allow/deny/ask"]
    G["bubble mode<br/>→ parent"]
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
```

**흐름**:
1. Claude Code PreToolUse 훅 발동
2. `permission.Resolver.Resolve()` 호출 (8-tier 스택)
3. Tier 순서 적용:
   - policy (프로젝트 정책)
   - project (.claude/settings.json)
   - user (~/.claude/settings.json)
   - team (팀 설정)
   - builtin (기본 정책)
   - systemDefault (OS 기본)
   - hookOverride (훅 오버라이드)
   - builtinDeny (최종 거부)
4. 첫 번째 "allow" 또는 "deny" 반환
5. bubble 모드: 자식 세션이 부모 세션으로 에스컬레이션

**5 모드**:
- default: 각 tool 별로 사용자에게 묻기
- acceptEdits: 모든 편집 자동 허용
- bypassPermissions: 모든 권한 자동 허용
- plan: 읽기만 허용
- bubble: 부모에게 에스컬레이션

---

## 6. 다중 세션 조율 (session.Registry)

```mermaid
flowchart TD
    A["SessionStart hook"]
    B["session.Registry.Register<br/>(active-sessions.json)"]
    C["Heartbeat"]
    D["PreToolUse<br/>race check"]
    E["ListActive"]
    F["SessionEnd hook"]
    G["Deregister"]
    H["PurgeStale<br/>(zombies)"]
    
    A --> B
    B --> C
    C --> D
    E --> D
    F --> G
    G --> H
```

**흐름**:
1. SessionStart 훅 → `Registry.Register()` 활성 세션 기록
2. active-sessions.json 에 진입
3. Heartbeat 주기적 갱신 (stale 방지)
4. PreToolUse: ListActive 쿼리, 병렬 세션 race 감지
5. SessionEnd 훅 → `Deregister()` 제거
6. `PurgeStale()` 좀비 세션 정리

**다중 세션 race 감지**:
- 같은 SPEC 작업 중인 다른 세션 감지
- git worktree base mismatch 경고
- advisory lock (flock/Windows mutex)

---

## 7. 그래프 신선도 (moai graph check / query-time refresh)

```mermaid
flowchart TD
    A["moai graph check"]
    B["codemaps<br/>endpoint git-diff"]
    C["mx-index<br/>inventory hash-diff"]
    D["edges.jsonl<br/>source fingerprint"]
    E["verdict fresh/stale/absent<br/>exit 0/1/2"]
    F["moai mx/graph query"]
    G["stale 판정 시<br/>changed-files-only refresh"]
    H["provenance 재스탬프<br/>tree root + commit"]

    A --> B
    A --> C
    A --> D
    B --> E
    C --> E
    D --> E
    F --> G
    G --> H
```

**흐름**:
1. `moai graph check` 3개 게이트 레이어(codemaps / mx-index / edges.jsonl)의 신선도를 수치로 보고
2. codemaps: 스탬프된 생성 커밋 대비 described-source 파일의 endpoint diff (revert된 churn은 0으로 계산)
3. mx-index: 스캔 인벤토리의 per-file content hash 대비 드리프트 수
4. edges.jsonl: 4개 소스 세트(codemaps/mx-index/specs/reports) fingerprint 불일치 수
5. mtime은 어떤 레이어에서도 신선도 신호로 쓰지 않음 (fresh worktree checkout이 모든 mtime을 초기화)
6. `moai mx query` / `moai graph query`는 답변 전 stale 레이어를 changed-files-only로 갱신 (LLM/네트워크 없음)
7. provenance 블록이 트리 루트 + 커밋(dirty면 fingerprint)을 기록 — 잘못된 트리의 인덱스는 절대 증분 신뢰하지 않음

---

## 8. 프리커밋 훅 보존 (moai init / update)

```mermaid
flowchart TD
    A["moai init / update"]
    B["설치 분류기<br/>(3-way)"]
    C["SHA-256 사이드카<br/>.git/hooks/.moai-pre-commit.sha256"]
    D["사용자 수정 훅"]
    E["백업 pre-commit.bak.<UTC>"]
    F["교체 + 백업 경로/stderr 공개"]

    A --> B
    B --> C
    C -->|불일치| D
    D --> E
    E --> F
```

**흐름**:
1. `moai init`/`moai update`가 기존 pre-commit 훅 발견 시 설치 분류기가 3-way 귀속 (설치본 vs 기록 다이제스트 vs 교체본)
2. 기록된 다이제스트와 다른 훅은 사용자 수정본으로 분류 — 자동 덮어쓰지 않음
3. 수정본은 `pre-commit.bak.<타임스탬프>`로 백업 (Windows 생성 가능한 콜론 없는 UTC 형식, 이름 충돌 시 형제 접미사)
4. 백업 경로와 교체 공지가 양 호출 지점의 stderr에 출력; 백업 실패 시 훅은 그대로 유지

---

## 9. codex 런치 게이트 (moai codex cli/app)

```mermaid
flowchart TD
    A["moai codex cli / app"]
    B["classifyCodexWiring<br/>(런처 배선 판정 소비)"]
    C{"wired?"}
    D["직접 기동 또는<br/>tmux spawn 기동"]
    E["상태·처방 보고"]
    F{"프롬프트 가능?"}
    G["보고 후 종료 rc 1<br/>(프롬프트 발행 없음)"]
    H["초기화 제안 프롬프트"]
    I{"수락?"}
    J["기록 없이 종료 rc 130<br/>(취소 — 오류와 구분)"]
    K["codexwiring.Wire<br/>(배선 생성, agent=codex, 1회)"]
    L["지시 계약 확보<br/>(AGENTS.md ↔ CLAUDE.md 링크)"]
    M["기동"]

    A --> B
    B --> C
    C -->|wired| D
    C -->|불완전| E
    E --> F
    F -->|불가| G
    F -->|가능| H
    H --> I
    I -->|거절| J
    I -->|수락| K
    K --> L
    L --> M
```

**흐름**:
1. 두 기동 동사(`cli`, `app`)가 기동 직전 같은 게이트 함수(`codexInitOfferGate`)를 통과 — 게이트는 `--spawn` 인자를 받지 않아 spawn 우회 경로가 존재하지 않음
2. 배선 판정은 런처의 단일 분류기(`classifyCodexWiring`, `internal/cli/codex_readiness.go`) 반환값을 소비 — 게이트가 디스크를 재판정하지 않음
3. `wired`면 즉시 기동, 그 외 상태는 상태와 처방을 stderr로 보고
4. 비대화형 세션은 프롬프트를 발행하지 않고 보고 후 종료 (대답할 수 없는 프롬프트가 자동화를 매다는 것을 방지)
5. 거절 시 아무것도 쓰지 않고 종료 코드 130 — 취소이지 오류가 아님
6. 수락 시 기존 배선 생성기(`codexwiring.Wire`)를 정확히 1회 호출한 뒤 지시 계약 확보 — 생성기 실패 시 계약 단계도, 기동도 없음
7. 지시 계약은 연결 전용: 없는 파일은 최소 본문으로 생성, 있는 파일은 바이트 단위 보존 + 링크 지시 1행 추가(이미 링크됐으면 무변화 — 코드펜스/주석/인용구 밖의 실행 import만 계수). 모든 경로는 봉쇄 판정(기존 컴포넌트는 일반 디렉터리, 심볼릭 링크/`..` 해석 결과가 프로젝트 루트 안)을 **읽기·쓰기보다 먼저** 통과하며, 쓰기는 per-file temp+rename에 전체 스테이징 → 전체 rename 순서

---

## 주요 인터페이스 계약

### Handler (Hook System)
```go
type Handler interface {
    Handle(ctx context.Context, input []byte) (output []byte, exit int, err error)
}
```

### Resolver (Permission)
```go
type Resolver interface {
    Resolve(ctx context.Context, tool string) (Allow, Deny, Ask)
}
```

### Deployer (Template)
```go
type Deployer interface {
    Deploy(ctx context.Context, dest, version string) error
}
```

### Registry (Session)
```go
type Registry interface {
    Register(spec, branch string) error
    Heartbeat(spec string) error
    ListActive(spec string) ([]Session, error)
    Deregister(spec string) error
}
```

---

**생성**: `/moai codemaps --force`로 자동 생성
