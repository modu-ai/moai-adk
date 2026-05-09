/* moai-adk CLI 터미널 씬 모음 — 각 명령어별 출력 디자인 */
/* eslint-disable */

// ── 1. install.sh ─────────────────────────────────────────
function SceneInstallSh() {
  const t = useTok();
  return (
    <Term title="bash — install.sh" cols={92} footer={<><span>install.sh · MoAI-ADK Installer</span><span>SHA256 verified · macOS · arm64</span></>}>
      <Prompt path="~" branch="" cmd={<><span style={{ color: t.fg }}>curl -fsSL </span><span style={{ color: t.info }}>https://moai-adk.dev/install.sh</span><span style={{ color: t.fg }}> | bash</span></>} />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700 }}>{`╔══════════════════════════════════════════════════════════════╗`}</div>
      <div style={{ color: t.accent, fontWeight: 700 }}>{`║          MoAI's Agentic Development Kit Installer            ║`}</div>
      <div style={{ color: t.accent, fontWeight: 700 }}>{`╚══════════════════════════════════════════════════════════════╝`}</div>
      {"\n"}
      <div><Tag kind="info">INFO</Tag>    Detecting platform…</div>
      <div><Tag kind="ok">SUCCESS</Tag> Detected platform: <C c="fg" b>darwin_arm64</C></div>
      <div><Tag kind="info">INFO</Tag>    Querying GitHub releases…</div>
      <div><Tag kind="ok">SUCCESS</Tag> Latest Go edition version: <C c="accent" b>3.2.4</C></div>
      <div><Tag kind="info">INFO</Tag>    Downloading from: <C c="info">github.com/modu-ai/moai-adk/releases/download/v3.2.4/...</C></div>
      <div style={{ color: t.dim }}>{`        ████████████████████████████████████████████  100%   18.4 MB / 18.4 MB`}</div>
      <div><Tag kind="ok">SUCCESS</Tag> Download completed</div>
      <div><Tag kind="info">INFO</Tag>    Verifying checksum…</div>
      <div><Tag kind="ok">SUCCESS</Tag> Checksum verified  <C c="dim">sha256: 7a9f…0c2e</C></div>
      <div><Tag kind="info">INFO</Tag>    Extracting archive…</div>
      <div><Tag kind="ok">SUCCESS</Tag> Extraction completed</div>
      <div><Tag kind="info">INFO</Tag>    Installing to: <C c="fg" b>~/.local/bin/moai</C></div>
      <div><Tag kind="ok">SUCCESS</Tag> Installed to: ~/.local/bin/moai</div>
      {"\n"}
      <Rule label="Installation complete" n={88} />
      {"\n"}
      <div><Tag kind="primary">moai</Tag>{`  `}<C c="fg" b>moai-adk 3.2.4</C>  <C c="dim">·  commit a1b2c3d  ·  built 2026-04-25</C></div>
      {"\n"}
      <div style={{ color: t.dim }}>To get started, run:</div>
      <div>  <C c="accent" b>moai init</C>      <C c="dim">→ Initialize a new project</C></div>
      <div>  <C c="accent" b>moai doctor</C>    <C c="dim">→ Check system health</C></div>
      <div>  <C c="accent" b>moai update</C>    <C c="dim">→ Update project templates</C></div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 2. moai (no args) ─ banner + help ───────────────────
function SceneBanner() {
  const t = useTok();
  return (
    <Term title="moai — Agentic Development Kit" cols={92} footer={<><span>$ moai</span><span>v3.2.4 · darwin_arm64</span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai" />
      {"\n"}
      <Banner version="v3.2.4" />
      {"\n"}
      <div style={{ color: t.fg, fontWeight: 700, letterSpacing: "-0.02em", fontFamily: '"Pretendard", sans-serif', fontSize: 14 }}>
        MoAI-ADK <C c="dim" style={{ fontWeight: 400 }}>·</C> Agentic Development Kit for Claude Code
      </div>
      <div style={{ color: t.dim }}>한국어 우선의 AI 개발 워크플로우 — <C c="accent">모두를 위한 AI를, 모두와 함께</C></div>
      {"\n"}
      <Rule label="Launch Commands" n={88} />
      <div>  <C c="accent" b>moai cc</C>      <C c="dim">Claude Code 런처 · 세션 자동 부트스트랩</C></div>
      <div>  <C c="accent" b>moai cg</C>      <C c="dim">Claude · GLM 콤보 런처</C></div>
      <div>  <C c="accent" b>moai glm</C>     <C c="dim">GLM 단일 모델 런처</C></div>
      {"\n"}
      <Rule label="Project Commands" n={88} />
      <div>  <C c="fg" b>moai init</C>     <C c="dim">새 프로젝트 초기화 (대화형 위저드)</C></div>
      <div>  <C c="fg" b>moai update</C>   <C c="dim">템플릿/SKILL/Hook 동기화</C></div>
      <div>  <C c="fg" b>moai doctor</C>   <C c="dim">시스템 진단 · 의존성 점검</C></div>
      <div>  <C c="fg" b>moai status</C>   <C c="dim">현재 프로젝트 상태</C></div>
      <div>  <C c="fg" b>moai spec</C>     <C c="dim">SPEC 라이프사이클 (lint · view · status)</C></div>
      <div>  <C c="fg" b>moai loop</C>     <C c="dim">자율 개발 루프 (start · status · pause)</C></div>
      <div>  <C c="fg" b>moai worktree</C>  <C c="dim">Git worktree (new · switch · sync · done)</C></div>
      {"\n"}
      <Rule label="Tools" n={88} />
      <div>  <C c="fg">moai astgrep</C>   <C c="dim">AST-grep 패턴 검색</C></div>
      <div>  <C c="fg">moai mx</C>        <C c="dim">@MX 앵커 · 의존성 그래프</C></div>
      <div>  <C c="fg">moai brain</C>     <C c="dim">Knowledge Graph & Memory</C></div>
      <div>  <C c="fg">moai constitution</C> <C c="dim">CX 7원칙 검증</C></div>
      <div>  <C c="fg">moai harness</C>   <C c="dim">테스트 하네스 / 품질 게이트</C></div>
      <div>  <C c="fg">moai telemetry</C> <C c="dim">로컬 텔레메트리 리포트</C></div>
      <div>  <C c="fg">moai version</C>   <C c="dim">버전 정보</C></div>
      {"\n"}
      <div style={{ color: t.dim }}>Use "<C c="accent" b>moai [command] --help</C>" for more information about a command.</div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 3. moai init wizard ─────────────────────────────────
function SceneInit() {
  const t = useTok();
  return (
    <Term title="moai — init wizard" cols={92} footer={<><span>moai init my-app</span><span>↑↓ 선택  ↵ 확인  ⌃C 취소</span></>}>
      <Prompt path="~/work" branch="" cmd="moai init my-app" />
      {"\n"}
      <Banner version="v3.2.4" />
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14 }}>
        Welcome to MoAI-ADK Project Initialization!
      </div>
      <div style={{ color: t.dim }}>이 위저드가 프로젝트 설정을 안내합니다. 언제든 ⌃C 로 취소할 수 있어요.</div>
      {"\n"}
      <Box width={88} title="Step 1 of 5 — 프로젝트 기본 정보" accent>
        <div><C c="rule">│</C>  <C c="dim">프로젝트 이름</C>      <C c="fg" b>my-app</C></div>
        <div><C c="rule">│</C>  <C c="dim">위치</C>             <C c="fg">/Users/yuna/work/my-app</C></div>
        <div><C c="rule">│</C>  <C c="dim">기본 언어</C>         <C c="accent" b>{`▸ TypeScript`}</C> <C c="faint">  Go   Python   Rust   기타</C></div>
        <div><C c="rule">│</C>  <C c="dim">프레임워크</C>        <C c="fg">Next.js 15 (감지됨)</C></div>
        <div><C c="rule">│</C></div>
      </Box>
      {"\n"}
      <Box width={88} title="Step 2 of 5 — 개발 모드" accent={false}>
        <div><C c="rule">│</C>   <C c="accent" b>{`◉ TDD`}</C>  <C c="fg" b>Test-Driven Development</C> <C c="dim">— 테스트가 먼저 (권장)</C></div>
        <div><C c="rule">│</C>   <C c="dim">{`○ DDD`}</C>  <C c="dim">Doc-Driven Development — 문서가 먼저</C></div>
        <div><C c="rule">│</C></div>
      </Box>
      {"\n"}
      <Box width={88} title="Step 3 of 5 — Git 워크플로우" accent={false}>
        <div><C c="rule">│</C>   <C c="dim">{`○ manual`}</C>    <C c="dim">로컬 git만 — 가장 단순</C></div>
        <div><C c="rule">│</C>   <C c="accent" b>{`◉ personal`}</C>  <C c="fg">개인 GitHub — 자동 PR</C></div>
        <div><C c="rule">│</C>   <C c="dim">{`○ team`}</C>      <C c="dim">팀 워크플로우 — 보호 브랜치 + 리뷰</C></div>
        <div><C c="rule">│</C></div>
      </Box>
      {"\n"}
      <Box width={88} title="Step 4 of 5 — Claude Code 통합" accent={false}>
        <div><C c="rule">│</C>   <C c="success" b>✓</C>  Skills 디렉토리 (.claude/skills)</div>
        <div><C c="rule">│</C>   <C c="success" b>✓</C>  Hooks (pre-push · spec-status · CX-guard)</div>
        <div><C c="rule">│</C>   <C c="success" b>✓</C>  Constitution / CX 7원칙 가드</div>
        <div><C c="rule">│</C>   <C c="dim">▢</C>  GLM 듀얼 모델 (선택)</div>
        <div><C c="rule">│</C></div>
      </Box>
      {"\n"}
      <div>{`  `}<C c="dim">[ ←  뒤로 ]</C>{`     `}<span style={{ background: t.accent, color: "#fff", padding: "2px 14px", borderRadius: 4, fontWeight: 700 }}> 다음 → </span>{`     `}<C c="dim">⌃C  취소</C></div>
      {"\n"}
      <div style={{ color: t.faint }}>{`  ──────────────────────  STEP  ●●○○○  ──────────────────────`}</div>
    </Term>
  );
}

// ── 4. moai doctor ──────────────────────────────────────
function SceneDoctor() {
  const t = useTok();
  const row = (label, status, value, hint) => (
    <div>
      {"  "}
      {status === "ok" && <C c="success" b>✓</C>}
      {status === "warn" && <C c="warning" b>!</C>}
      {status === "err" && <C c="danger" b>✗</C>}
      {status === "info" && <C c="info" b>·</C>}
      {"  "}
      <span style={{ color: t.fg, display: "inline-block", width: 220 }}>{label}</span>
      <span style={{ color: t.dim }}>{value}</span>
      {hint && <span style={{ color: t.faint }}>  {hint}</span>}
    </div>
  );
  return (
    <Term title="moai — doctor" cols={92} footer={<><span>moai doctor</span><span>17 checks · 0.42s</span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai doctor" />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14, letterSpacing: "-0.02em" }}>
        MoAI Doctor — 시스템 진단
      </div>
      <div style={{ color: t.dim }}>설치 · 의존성 · 프로젝트 상태를 17개 항목으로 점검합니다.</div>
      {"\n"}
      <Rule label="Runtime" n={88} />
      {row("MoAI-ADK 바이너리", "ok", "v3.2.4 (~/.local/bin/moai)")}
      {row("Go 런타임", "ok", "go1.23.4 darwin/arm64")}
      {row("Claude Code", "ok", "v1.4.2", "PATH: /usr/local/bin/claude")}
      {row("Git", "ok", "git 2.46.0")}
      {row("GitHub CLI (gh)", "ok", "v2.55.0  ·  authenticated as @yuna-kim")}
      {"\n"}
      <Rule label="Project (.moai)" n={88} />
      {row("프로젝트 구조", "ok", ".moai/ · .claude/ · 모두 존재")}
      {row("manifest.yaml", "ok", "v3.2.4  · 47 entries", "마지막 동기화 5시간 전")}
      {row("Skills (.claude/skills)", "ok", "12개 등록")}
      {row("Hooks", "ok", "pre-push · post-merge · 4개 활성")}
      {row("Constitution / CX 7원칙", "ok", "통과 (7/7)")}
      {"\n"}
      <Rule label="Quality Gates" n={88} />
      {row("Test harness", "ok", "go test ./...  · 218 tests · 92.4% coverage")}
      {row("AST-grep rules", "ok", "31 rules · 0 violations")}
      {row("LSP (gopls)", "ok", "준비됨")}
      {row("Telemetry", "info", "local-only · ~/.moai/telemetry.db")}
      {"\n"}
      <Rule label="Warnings" n={88} />
      {row(".moai-backups 사이즈", "warn", "1.8 GB", "‘moai update --prune’ 권장")}
      {row("LSP 캐시 오래됨", "warn", "마지막 갱신 14일 전", "‘moai lsp doctor --refresh’")}
      {"\n"}
      <Box width={88} title="요약" accent>
        <div><C c="rule">│</C>  <C c="success" b>15</C> 정상   <C c="warning" b>2</C> 경고   <C c="danger" b>0</C> 에러   <C c="info" b>1</C> 정보</div>
        <div><C c="rule">│</C>  <C c="dim">전체적으로 양호합니다. 경고 항목은 자동으로 수정할 수 있어요.</C></div>
        <div><C c="rule">│</C>  <C c="accent" b>→</C> <C c="fg" b>moai doctor --fix</C> <C c="dim">로 자동 수리</C></div>
      </Box>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 5. moai status ──────────────────────────────────────
function SceneStatus() {
  const t = useTok();
  return (
    <Term title="moai — status" cols={92} footer={<><span>moai status</span><span>SPEC: AUTH-001 · loop: idle</span></>}>
      <Prompt path="~/work/my-app" branch="feat/auth-flow" dirty cmd="moai status" />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14, letterSpacing: "-0.02em" }}>my-app  <C c="dim" style={{ fontWeight: 400, fontSize: 12 }}>· TypeScript · Next.js 15 · TDD · personal</C></div>
      {"\n"}
      <Rule label="Active SPEC" n={88} />
      <div>  <C c="primary"><Tag kind="primary">SPEC-AUTH-001</Tag></C>  <C c="fg" b>OAuth 로그인 플로우 (Google · GitHub)</C></div>
      <div>  <C c="dim">단계</C>     <C c="accent" b>{`[●●●○○]`}</C>  <C c="fg">RED → GREEN → REFACTOR  ·  현재: GREEN</C></div>
      <div>  <C c="dim">진행률</C>   <C c="success" b>62%</C>  <C c="dim">12 of 19 acceptance criteria</C></div>
      <div>  <C c="dim">담당</C>     <C c="fg">@yuna-kim</C>{`   `}<C c="dim">갱신</C> <C c="fg">8분 전</C></div>
      {"\n"}
      <Rule label="Loop & Worktree" n={88} />
      <div>  <C c="dim">Loop 상태</C>           <C c="success" b>idle</C> <C c="dim">— 마지막 사이클 14:32 (3분 18초)</C></div>
      <div>  <C c="dim">현재 worktree</C>       <C c="fg" b>feat/auth-flow</C> <C c="dim">  →  ~/work/my-app.wt/auth-flow</C></div>
      <div>  <C c="dim">활성 worktree</C>       <C c="fg">3</C> <C c="dim">  ·  feat/auth-flow · fix/csrf-token · spike/passkeys</C></div>
      {"\n"}
      <Rule label="Quality Gates" n={88} />
      <div>  <C c="success" b>✓</C> Tests       <C c="fg">218 / 218 통과</C>{`   `}<C c="dim">12.4s</C></div>
      <div>  <C c="success" b>✓</C> Coverage    <C c="fg">92.4%</C>{`   `}<C c="dim">목표 90% 이상</C></div>
      <div>  <C c="success" b>✓</C> AST-grep    <C c="fg">0 위반</C>{`   `}<C c="dim">31 규칙</C></div>
      <div>  <C c="warning" b>!</C> Lint        <C c="fg">3 경고</C>{`   `}<C c="dim">internal/auth/oauth.go</C></div>
      <div>  <C c="success" b>✓</C> Constitution <C c="fg">CX 7원칙 통과</C></div>
      {"\n"}
      <Rule label="Recent" n={88} />
      <div>  <C c="dim">14:32</C>  <C c="success" b>✓</C>  loop cycle complete   <C c="dim">3 files · +142 / −18</C></div>
      <div>  <C c="dim">14:21</C>  <C c="info" b>·</C>  spec view AUTH-001</div>
      <div>  <C c="dim">13:58</C>  <C c="success" b>✓</C>  pre-push hook  <C c="dim">testify · vet · ast-grep</C></div>
      <div>  <C c="dim">13:45</C>  <C c="info" b>·</C>  worktree new feat/auth-flow</div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 6. moai version ─────────────────────────────────────
function SceneVersion() {
  const t = useTok();
  return (
    <Term title="moai — version" cols={92} footer={<><span>moai version</span><span></span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai version" />
      {"\n"}
      <Box width={64} title="moai-adk 3.2.4" accent>
        <div>{`│  `}<C c="dim">commit</C>     <C c="fg">a1b2c3d4e5f6</C></div>
        <div>{`│  `}<C c="dim">built</C>      <C c="fg">2026-04-25 09:14 KST</C></div>
        <div>{`│  `}<C c="dim">go</C>         <C c="fg">go1.23.4</C></div>
        <div>{`│  `}<C c="dim">platform</C>   <C c="fg">darwin/arm64</C></div>
        <div>{`│  `}<C c="dim">channel</C>    <C c="accent" b>stable</C>{`  `}<C c="dim">· 다음 릴리즈 v3.3.0</C></div>
      </Box>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 7. moai cc launcher ─────────────────────────────────
function SceneCC() {
  const t = useTok();
  return (
    <Term title="moai — cc launcher" cols={92} footer={<><span>moai cc</span><span>session: ck_2k4mq8 · sonnet-4.5</span></>}>
      <Prompt path="~/work/my-app" branch="feat/auth-flow" dirty cmd="moai cc" />
      {"\n"}
      <Banner version="v3.2.4" />
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 13 }}>
        Claude Code 런처{`  `}<C c="dim" style={{ fontWeight: 400 }}>· 세션 자동 부트스트랩</C>
      </div>
      {"\n"}
      <div>  <C c="success" b>✓</C> CLAUDE.md 로드     <C c="dim">2.1 KB · 17 sections</C></div>
      <div>  <C c="success" b>✓</C> Skills 12개 활성   <C c="dim">.claude/skills/</C></div>
      <div>  <C c="success" b>✓</C> Hooks 4개 등록     <C c="dim">pre-tool · post-tool · pre-push · spec-status</C></div>
      <div>  <C c="success" b>✓</C> Constitution 가드  <C c="dim">CX 7원칙 · 위반 자동 차단</C></div>
      <div>  <C c="success" b>✓</C> MX 앵커 인덱스     <C c="dim">412 anchors · graph 준비됨</C></div>
      <div>  <C c="info" b>·</C> Telemetry          <C c="dim">local-only · 익명</C></div>
      {"\n"}
      <Box width={88} title="세션" accent>
        <div>{`│  `}<C c="dim">model</C>      <C c="fg" b>claude-sonnet-4.5</C>{`     `}<C c="dim">temp 0  ·  max-out 8K</C></div>
        <div>{`│  `}<C c="dim">profile</C>    <C c="fg">my-app · personal</C></div>
        <div>{`│  `}<C c="dim">SPEC</C>       <Tag kind="primary">SPEC-AUTH-001</Tag>{`  `}<C c="fg">OAuth 로그인 플로우</C></div>
        <div>{`│  `}<C c="dim">cwd</C>        <C c="fg">~/work/my-app.wt/auth-flow</C></div>
      </Box>
      {"\n"}
      <div>  <C c="accent" b>→</C> Claude Code 시작 중…{`  `}<span style={{ color: t.dim, animation: "pulse 1.6s ease-in-out infinite" }}>{`▰▰▰▱▱▱`}</span></div>
      <div>  <C c="dim">  ⌃D 또는 'exit' 으로 종료 · 진행 상황은 텔레메트리에 자동 기록됩니다.</C></div>
      {"\n"}
      <div style={{ color: t.fg }}><C c="accent" b>›</C> Claude Code <C c="dim">— SPEC-AUTH-001 · sonnet-4.5</C></div>
      <div style={{ color: t.dim }}>  안녕하세요, 유나님. <C c="fg">OAuth Google 콜백 핸들러</C>의 GREEN 단계가 끝났습니다.</div>
      <div style={{ color: t.dim }}>  이어서 <C c="accent" b>REFACTOR</C> 단계를 진행할까요? <C c="faint">[y/N/show diff]</C></div>
      {"\n"}
      <div>  <C c="accent" b>›</C> <span style={{ color: t.cursor, animation: "blink 1.05s steps(1) infinite" }}>▌</span></div>
    </Term>
  );
}

// ── 8. moai loop start ──────────────────────────────────
function SceneLoop() {
  const t = useTok();
  return (
    <Term title="moai — loop" cols={92} footer={<><span>moai loop start --max-cycles 6</span><span>cycle 3 of 6 · 1m 47s</span></>}>
      <Prompt path="~/work/my-app" branch="feat/auth-flow" dirty cmd="moai loop start --max-cycles 6" />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14, letterSpacing: "-0.02em" }}>
        자율 개발 루프{`  `}<C c="dim" style={{ fontWeight: 400, fontSize: 12 }}>· SPEC-AUTH-001 · TDD</C>
      </div>
      {"\n"}
      <div>  <C c="success" b>●</C> <C c="dim">cycle 1</C>{`  `}<C c="success">RED → GREEN</C>{`         `}<C c="fg">테스트 7개 추가 · 모두 통과</C>{`  `}<C c="dim">42s</C></div>
      <div>  <C c="success" b>●</C> <C c="dim">cycle 2</C>{`  `}<C c="success">REFACTOR</C>{`            `}<C c="fg">handler 분리 · 중복 제거</C>{`        `}<C c="dim">31s</C></div>
      <div>  <C c="accent" b>●</C> <C c="fg" b>cycle 3</C>{`  `}<C c="accent" b>{`▸ RED`}</C>{`               `}<C c="fg">CSRF 토큰 검증 테스트 작성 중…</C>{`  `}<C c="dim">34s</C></div>
      <div>  <C c="dim">○ cycle 4</C>{`  `}<C c="faint">대기</C></div>
      <div>  <C c="dim">○ cycle 5</C>{`  `}<C c="faint">대기</C></div>
      <div>  <C c="dim">○ cycle 6</C>{`  `}<C c="faint">대기</C></div>
      {"\n"}
      <Box width={88} title="현재 단계 — RED (실패하는 테스트 작성)" accent>
        <div>{`│  `}<C c="dim">목표</C>       <C c="fg">CSRF 토큰이 없으면 401을 반환해야 한다</C></div>
        <div>{`│  `}<C c="dim">파일</C>       <C c="info">internal/auth/csrf_test.go</C></div>
        <div>{`│  `}<C c="dim">진행</C>       <span style={{ color: t.accent }}>{`██████████████░░░░░░░░░░  58%`}</span></div>
        <div>{`│  `}<C c="dim">에이전트</C>   <Tag kind="primary">spec-builder</Tag>{` `}<Tag kind="info">code-builder</Tag>{` `}<Tag kind="ok">tag-agent</Tag></div>
      </Box>
      {"\n"}
      <Rule label="Live log" n={88} />
      <div><C c="dim">14:38:12</C> <Tag kind="info">SPEC</Tag> AUTH-001 acceptance criterion 13/19 로딩</div>
      <div><C c="dim">14:38:15</C> <Tag kind="primary">PLAN</Tag> 테스트 시나리오 3개 도출</div>
      <div><C c="dim">14:38:21</C> <Tag kind="ok">EDIT</Tag> internal/auth/csrf_test.go  <C c="success">+47</C> <C c="danger">−0</C></div>
      <div><C c="dim">14:38:24</C> <Tag kind="info">RUN</Tag> go test ./internal/auth/...  <C c="warning">FAIL (예상됨)</C></div>
      <div><C c="dim">14:38:25</C> <Tag kind="ok">GATE</Tag> RED 단계 통과 — GREEN 으로 전환</div>
      {"\n"}
      <div style={{ color: t.dim }}>  <C c="accent" b>p</C> 일시정지   <C c="accent" b>r</C> 재개   <C c="accent" b>x</C> 취소   <C c="accent" b>?</C> 도움말</div>
      <div style={{ color: t.faint }}>  ──────  cycle 3 / 6  ●●●○○○  진행 시간 1m 47s · 추정 잔여 4m 12s  ──────</div>
    </Term>
  );
}

// ── 9. moai spec view ───────────────────────────────────
function SceneSpec() {
  const t = useTok();
  return (
    <Term title="moai — spec view AUTH-001" cols={92} footer={<><span>moai spec view AUTH-001</span><span>.moai/specs/AUTH-001.md</span></>}>
      <Prompt path="~/work/my-app" branch="feat/auth-flow" dirty cmd="moai spec view AUTH-001" />
      {"\n"}
      <Box width={88} title="SPEC-AUTH-001 · OAuth 로그인 플로우" accent>
        <div>{`│  `}<C c="dim">상태</C>       <C c="success" b>● ACTIVE</C>{`   `}<C c="dim">단계</C> <C c="accent" b>GREEN</C>{`   `}<C c="dim">진행</C> <C c="success" b>62%</C></div>
        <div>{`│  `}<C c="dim">생성</C>       <C c="fg">2026-04-19 · @yuna-kim</C></div>
        <div>{`│  `}<C c="dim">우선순위</C>   <Tag kind="warn">P1</Tag>{`   `}<C c="dim">예상 사이클</C> <C c="fg">8</C>{`   `}<C c="dim">완료</C> <C c="fg">5</C></div>
        <div>{`│  `}<C c="dim">의존성</C>     <C c="info">SPEC-DB-002</C>{`   `}<C c="info">SPEC-SESSION-001</C></div>
      </Box>
      {"\n"}
      <Rule label="목표 (Goal)" n={88} />
      <div style={{ color: t.fg, fontFamily: '"Pretendard", sans-serif', letterSpacing: "-0.025em" }}>
        Google · GitHub OAuth 로그인을 PKCE 흐름으로 구현하고, 세션 토큰을
      </div>
      <div style={{ color: t.fg, fontFamily: '"Pretendard", sans-serif', letterSpacing: "-0.025em" }}>
        HttpOnly 쿠키로 안전하게 발급한다. <C c="dim">검증된 한국어 흐름 — 모두를 위한 인증.</C>
      </div>
      {"\n"}
      <Rule label="수용 기준 (Acceptance Criteria)" n={88} />
      <div>  <C c="success" b>✓</C> <C c="fg">AC-01</C>  /auth/google → state · code_verifier 생성 후 리디렉션</div>
      <div>  <C c="success" b>✓</C> <C c="fg">AC-02</C>  콜백에서 state 불일치 시 403</div>
      <div>  <C c="success" b>✓</C> <C c="fg">AC-03</C>  토큰 교환 실패 시 사용자에게 한국어 메시지 노출</div>
      <div>  <C c="success" b>✓</C> <C c="fg">AC-04</C>  최초 로그인 시 users 테이블에 upsert</div>
      <div>  <C c="success" b>✓</C> <C c="fg">AC-12</C>  세션 쿠키 SameSite=Lax · Secure · HttpOnly</div>
      <div>  <C c="accent" b>▸</C> <C c="fg" b>AC-13</C>  CSRF 토큰 미존재 시 401 (현재 작업 중)</div>
      <div>  <C c="dim">○ AC-14  세션 만료 시 자동 갱신</C></div>
      <div>  <C c="dim">○ AC-19  Passkeys 폴백 (스파이크)</C></div>
      <div>  <C c="dim">  ... 7 hidden  · </C><C c="accent">전체 보기 ↵</C></div>
      {"\n"}
      <Rule label="MX 앵커" n={88} />
      <div>  <C c="info">@MX:ANCHOR</C>  <C c="fg">internal/auth/oauth.go:42</C>{`     `}<C c="dim">HandleGoogleCallback</C></div>
      <div>  <C c="info">@MX:ANCHOR</C>  <C c="fg">internal/auth/csrf.go:18</C>{`      `}<C c="dim">VerifyCSRFToken (in progress)</C></div>
      <div>  <C c="info">@MX:REASON</C>  <C c="dim">fan_in=4 · 보호 필요한 진입점</C></div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 10. moai update ─────────────────────────────────────
function SceneUpdate() {
  const t = useTok();
  return (
    <Term title="moai — update" cols={92} footer={<><span>moai update --project</span><span>3 changed · 0 conflicts</span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai update --project" />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14, letterSpacing: "-0.02em" }}>프로젝트 템플릿 동기화</div>
      <div style={{ color: t.dim }}>upstream <C c="fg">v3.2.4</C> ↔ 내 프로젝트 <C c="fg">v3.2.1</C> 차이 비교</div>
      {"\n"}
      <Rule label="비교 결과 (3 변경 · 0 충돌)" n={88} />
      <div>  <C c="success" b>+</C>  <C c="fg">.claude/skills/spec-lint.md</C>{`             `}<Tag kind="ok">새 파일</Tag>{`  `}<C c="dim">+248</C></div>
      <div>  <C c="warning" b>~</C>  <C c="fg">.moai/templates/spec.md</C>{`                  `}<Tag kind="warn">병합</Tag>{`  `}<C c="dim">+12 / −4</C></div>
      <div>  <C c="warning" b>~</C>  <C c="fg">.claude/hooks/pre-push.sh</C>{`                `}<Tag kind="warn">병합</Tag>{`  `}<C c="dim">+8 / −2</C></div>
      <div>  <C c="info" b>=</C>  <C c="dim">.moai/manifest.yaml                       내 변경 보존</C></div>
      <div>  <C c="info" b>=</C>  <C c="dim">CLAUDE.md                                 내 변경 보존</C></div>
      {"\n"}
      <Rule label="안전 장치" n={88} />
      <div>  <C c="success" b>✓</C>  자동 백업 → <C c="info">.moai-backups/2026-04-25_14-42/</C></div>
      <div>  <C c="success" b>✓</C>  Constitution 가드 활성 — 보호 영역 변경 차단</div>
      <div>  <C c="success" b>✓</C>  Dry-run 결과와 동일 (재현 가능)</div>
      {"\n"}
      <Box width={88} title="진행" accent>
        <div>{`│  `}<C c="success" b>✓</C>  파일 백업                  <C c="dim">5 files · 18 KB · 0.04s</C></div>
        <div>{`│  `}<C c="success" b>✓</C>  새 스킬 추가                <C c="dim">spec-lint · 248 lines</C></div>
        <div>{`│  `}<C c="success" b>✓</C>  spec.md 3-way merge         <C c="dim">conflicts: 0</C></div>
        <div>{`│  `}<C c="success" b>✓</C>  pre-push.sh 3-way merge     <C c="dim">conflicts: 0</C></div>
        <div>{`│  `}<C c="success" b>✓</C>  manifest 갱신               <C c="dim">v3.2.1 → v3.2.4</C></div>
        <div>{`│  `}<C c="success" b>✓</C>  Constitution 재검증         <C c="dim">CX 7원칙 통과</C></div>
      </Box>
      {"\n"}
      <div>  <C c="success" b>업데이트 완료.</C> <C c="dim">  체크아웃 가능: </C><C c="accent" b>git diff HEAD</C></div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 11. moai worktree ───────────────────────────────────
function SceneWorktree() {
  const t = useTok();
  return (
    <Term title="moai — worktree" cols={92} footer={<><span>moai worktree list</span><span>3 active · base = main</span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai worktree list" />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14 }}>활성 worktree</div>
      {"\n"}
      <div style={{ color: t.dim }}>{`  BRANCH                      PATH                                         SPEC          STATUS`}</div>
      <div style={{ color: t.rule }}>{`  ──────────────────────────  ───────────────────────────────────────────  ────────────  ─────────`}</div>
      <div>  <C c="accent" b>★ feat/auth-flow</C>          <C c="fg">~/work/my-app.wt/auth-flow</C>{`         `}<Tag kind="primary">AUTH-001</Tag>{`     `}<C c="warning">+12 / −4</C></div>
      <div>    <C c="fg">fix/csrf-token</C>            <C c="fg">~/work/my-app.wt/csrf-token</C>{`        `}<Tag kind="primary">AUTH-001</Tag>{`     `}<C c="dim">clean</C></div>
      <div>    <C c="fg">spike/passkeys</C>            <C c="fg">~/work/my-app.wt/passkeys</C>{`          `}<Tag kind="info">SPIKE</Tag>{`         `}<C c="warning">+87 / −2</C></div>
      <div>    <C c="dim">main (base)</C>               <C c="dim">~/work/my-app</C>{`                       `}<C c="dim">—</C>{`             `}<C c="dim">clean</C></div>
      {"\n"}
      <Rule label="명령어" n={88} />
      <div>  <C c="accent" b>moai worktree new</C> <C c="fg">{`<branch>`}</C>{`         `}<C c="dim">새 worktree 생성 (SPEC 자동 연결)</C></div>
      <div>  <C c="accent" b>moai worktree switch</C> <C c="fg">{`<branch>`}</C>{`      `}<C c="dim">현재 worktree 전환</C></div>
      <div>  <C c="accent" b>moai worktree sync</C>{`               `}<C c="dim">base 브랜치 변경 사항 머지</C></div>
      <div>  <C c="accent" b>moai worktree done</C>{`               `}<C c="dim">PR 생성 후 정리</C></div>
      <div>  <C c="accent" b>moai worktree clean</C>{`              `}<C c="dim">병합된/오래된 worktree 일괄 삭제</C></div>
      <div>  <C c="accent" b>moai worktree recover</C>{`            `}<C c="dim">손상된 worktree 복구</C></div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 12. moai constitution check ─────────────────────────
function SceneConstitution() {
  const t = useTok();
  const law = (n, name, ok, note) => (
    <div>
      {"  "}{ok ? <C c="success" b>✓</C> : <C c="warning" b>!</C>}{"  "}
      <C c="dim">제{n}조</C>{"  "}
      <C c="fg" b>{name}</C>
      {note ? <span style={{ color: t.dim }}>  — {note}</span> : null}
    </div>
  );
  return (
    <Term title="moai — constitution check" cols={92} footer={<><span>moai constitution check</span><span>CX 7원칙 · 7/7 통과</span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai constitution check" />
      {"\n"}
      <Banner version="v3.2.4" />
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14, letterSpacing: "-0.025em" }}>
        모두의AI 헌법 — CX 7원칙 검증
      </div>
      <div style={{ color: t.dim }}>Collective eXperience — 베타 테스터와 함께 만드는 신뢰의 약속</div>
      {"\n"}
      <Rule label="결과" n={88} />
      {law("1", "포용 — 누구도 배제하지 않는다", true, "한국어 메시지 노출 검사 통과")}
      {law("2", "투명 — 매출·지표·실패 모두 공개", true, "telemetry는 로컬 전용 · 기본 익명")}
      {law("3", "동료 — 베타 테스터를 동료로", true, "기여 트래킹 활성")}
      {law("4", "검증 — 한국어 AI 큐레이션", true, "Constitution 가드 활성")}
      {law("5", "로컬 우선 — 데이터 주권", true, "외부 호출 0건 (이 세션)")}
      {law("6", "재현 가능 — 매 사이클 기록", true, "loop manifest sha 일치")}
      {law("7", "고객 만족 — 단 하나의 목표", true, "사용자 보고 0건 미해결")}
      {"\n"}
      <Box width={88} title="요약" accent>
        <div>{`│  `}<C c="success" b>● 7 / 7 통과</C>{`   `}<C c="dim">위반 0  ·  경고 0  ·  마지막 검사 14:42</C></div>
        <div>{`│  `}<C c="dim">이 프로젝트는 모두의AI 헌법을 준수합니다.</C></div>
        <div>{`│  `}<C c="accent" b>→</C> <C c="fg">moai constitution show --law 4</C>{` `}<C c="dim">로 조항 본문 보기</C></div>
      </Box>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 13. moai mx graph ───────────────────────────────────
function SceneMX() {
  const t = useTok();
  return (
    <Term title="moai — mx graph" cols={92} footer={<><span>moai mx graph internal/auth</span><span>412 anchors · 1873 edges</span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai mx graph internal/auth" />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14, letterSpacing: "-0.025em" }}>@MX 앵커 그래프{`  `}<C c="dim" style={{ fontWeight: 400, fontSize: 12 }}>· internal/auth</C></div>
      {"\n"}
      <div style={{ color: t.dim }}>{`        ┌─────────────────────┐`}</div>
      <div>{`        `}<C c="rule">│ </C><C c="accent" b>HandleGoogleCallback</C><C c="rule"> │</C>{`           `}<C c="dim">fan_in=4  fan_out=3</C></div>
      <div style={{ color: t.dim }}>{`        └──────────┬──────────┘`}</div>
      <div style={{ color: t.dim }}>{`                   │  `}<C c="info">@MX:CALLS</C></div>
      <div style={{ color: t.dim }}>{`         ┌─────────┼─────────┬─────────────┐`}</div>
      <div style={{ color: t.dim }}>{`         ▼         ▼         ▼             ▼`}</div>
      <div>{`   `}<C c="rule">┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐</C></div>
      <div>{`   `}<C c="rule">│</C> <C c="fg" b>exchToken</C> <C c="rule">│ │</C> <C c="fg" b>verifyState</C><C c="rule">│ │</C> <C c="fg" b>upsertUser</C><C c="rule">│ │</C> <C c="warning" b>VerifyCSRF</C>  <C c="rule">│</C></div>
      <div>{`   `}<C c="rule">└──────────┘ └──────────┘ └──────────┘ └──────────────┘</C></div>
      <div>{`     `}<C c="success">covered</C>{`     `}<C c="success">covered</C>{`     `}<C c="success">covered</C>{`        `}<C c="warning" b>missing</C></div>
      {"\n"}
      <Rule label="요약" n={88} />
      <div>  <C c="dim">노드</C>     <C c="fg" b>14</C>{`    `}<C c="dim">엣지</C> <C c="fg" b>23</C>{`    `}<C c="dim">사이클</C> <C c="fg">0</C>{`    `}<C c="dim">고립</C> <C c="fg">1</C></div>
      <div>  <C c="dim">테스트</C>   <C c="success" b>13/14 covered</C>{`    `}<C c="warning" b>!</C> <C c="fg">VerifyCSRFToken</C> <C c="dim">— 작성 중 (cycle 3)</C></div>
      <div>  <C c="dim">REASONS</C>  <C c="fg">28개 추론 자동 생성</C>{` `}<C c="dim">— 'fan_in &lt; 2' 항목은 정리 후보</C></div>
      {"\n"}
      <div>  <C c="accent" b>→</C> <C c="fg">moai mx query "fan_in &gt; 3"</C>{` `}<C c="dim">로 핫스팟 조회</C></div>
      <div>  <C c="accent" b>→</C> <C c="fg">moai mx export --json</C>{` `}<C c="dim">로 외부 시각화</C></div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 14. moai telemetry report ───────────────────────────
function SceneTelemetry() {
  const t = useTok();
  const bar = (label, value, of, color) => {
    const pct = Math.round((value / of) * 100);
    const filled = Math.round(pct * 0.36);
    const w = 36;
    return (
      <div>
        {"  "}<span style={{ color: t.dim, display: "inline-block", width: 22 }}>{label}</span>
        <span style={{ color: color }}>{"█".repeat(filled)}</span>
        <span style={{ color: t.rule }}>{"░".repeat(Math.max(0, w - filled))}</span>
        {"  "}
        <C c="fg" b>{value}</C>
        <C c="dim"> / {of}</C>
      </div>
    );
  };
  return (
    <Term title="moai — telemetry report" cols={92} footer={<><span>moai telemetry report --week</span><span>local-only · 익명</span></>}>
      <Prompt path="~/work/my-app" branch="main" cmd="moai telemetry report --week" />
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700, fontFamily: '"Pretendard", sans-serif', fontSize: 14, letterSpacing: "-0.025em" }}>주간 리포트{`  `}<C c="dim" style={{ fontWeight: 400, fontSize: 12 }}>· 2026-04-19 ~ 04-25 · 로컬 전용</C></div>
      {"\n"}
      <Rule label="활동" n={88} />
      {bar("loop", 24, 30, t.accent)}
      {bar("init", 1, 30, t.info)}
      {bar("cc", 47, 60, t.success)}
      {bar("update", 3, 30, t.warning)}
      {bar("worktree", 18, 30, t.info)}
      {bar("doctor", 6, 30, t.dim)}
      {"\n"}
      <Rule label="품질 지표 (이번 주)" n={88} />
      <div>  <C c="dim">테스트 통과율</C>{`           `}<C c="success" b>99.4%</C>{`   `}<C c="dim">218/219</C>{`   `}<C c="success">▲ 0.8%</C></div>
      <div>  <C c="dim">커버리지</C>{`                `}<C c="success" b>92.4%</C>{`   `}<C c="dim">목표 90%</C>{`   `}<C c="success">▲ 1.2%</C></div>
      <div>  <C c="dim">루프 평균 시간</C>{`           `}<C c="fg" b>3m 18s</C>{`   `}<C c="dim">중앙값 2m 51s</C>{`   `}<C c="success">▼ 22s</C></div>
      <div>  <C c="dim">RED→GREEN 성공률</C>{`         `}<C c="success" b>87%</C>{`     `}<C c="dim">42/48</C>{`   `}<C c="warning">▼ 3%</C></div>
      <div>  <C c="dim">CX 7원칙 위반</C>{`            `}<C c="success" b>0</C>{`       `}<C c="dim">7일 연속</C></div>
      {"\n"}
      <Rule label="외부 호출" n={88} />
      <div>  <C c="success" b>0</C>  외부 텔레메트리 전송{`  `}<C c="dim">— 모든 데이터는 ~/.moai/telemetry.db 에만 저장됩니다.</C></div>
      <div>  <C c="dim">  데이터 내보내기:</C> <C c="accent" b>moai telemetry export --json &gt; out.json</C></div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 15. error state ─────────────────────────────────────
function SceneError() {
  const t = useTok();
  return (
    <Term title="moai — error" cols={92} footer={<><span>moai update --project</span><span>exit code 1</span></>}>
      <Prompt path="~/work/my-app" branch="feat/auth-flow" dirty cmd="moai update --project" />
      {"\n"}
      <div style={{ color: t.danger, fontWeight: 700 }}>{`╔══════════════════════════════════════════════════════════════════════════════╗`}</div>
      <div style={{ color: t.danger, fontWeight: 700 }}>{`║  `}<C c="danger" b>오류</C>  Update aborted — uncommitted changes in protected files{`               ║`}</div>
      <div style={{ color: t.danger, fontWeight: 700 }}>{`╚══════════════════════════════════════════════════════════════════════════════╝`}</div>
      {"\n"}
      <div>  <C c="danger" b>✗</C>  <C c="fg" b>CLAUDE.md</C>{`             `}<C c="dim">3 라인 변경됨, 커밋되지 않음</C></div>
      <div>  <C c="danger" b>✗</C>  <C c="fg" b>.moai/manifest.yaml</C>{`   `}<C c="dim">버전 충돌 — 로컬 v3.2.1 / 원격 v3.2.4</C></div>
      {"\n"}
      <Box width={88} title="무엇이 일어났나" accent={false}>
        <div>{`│  `}<C c="dim">Update 명령은 Constitution 가드가 보호하는 파일을 안전하게 동기화하기 위해</C></div>
        <div>{`│  `}<C c="dim">먼저 깨끗한 작업 트리를 요구합니다. 현재 두 파일에 미커밋 변경이 있습니다.</C></div>
      </Box>
      {"\n"}
      <Box width={88} title="어떻게 풀까" accent>
        <div>{`│  `}<C c="accent" b>1.</C>  <C c="fg">변경 사항 커밋 또는 stash</C></div>
        <div>{`│       `}<C c="info" b>git stash --include-untracked</C></div>
        <div>{`│       `}<C c="dim">또는</C>{`  `}<C c="info" b>git commit -am "WIP"</C></div>
        <div>{`│`}</div>
        <div>{`│  `}<C c="accent" b>2.</C>  <C c="fg">강제로 백업 후 진행 (자동 백업 포함)</C></div>
        <div>{`│       `}<C c="info" b>moai update --project --force</C></div>
        <div>{`│`}</div>
        <div>{`│  `}<C c="accent" b>3.</C>  <C c="fg">먼저 변경 내용 확인</C></div>
        <div>{`│       `}<C c="info" b>moai update --project --dry-run</C></div>
      </Box>
      {"\n"}
      <div>  <C c="dim">코드: </C><Tag kind="err">E_UPDATE_DIRTY_TREE</Tag>{`   `}<C c="dim">자세히: </C><C c="info">moai doctor --explain E_UPDATE_DIRTY_TREE</C></div>
      <div>  <C c="dim">한국어 모드는 항상 켜져 있어요. 도움이 필요하면 </C><C c="accent" b>moai help</C><C c="dim">.</C></div>
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

// ── 16. statusline (single-line) ────────────────────────
function SceneStatusline() {
  const t = useTok();
  const Line = ({ label, children }) => (
    <div>
      <span style={{ color: t.dim, display: "inline-block", width: 90 }}>{label}</span>
      <span>{children}</span>
    </div>
  );
  return (
    <Term title="moai — statusline" cols={92} height={260} footer={<><span>moai statusline --watch</span><span>refresh 2s</span></>}>
      <Prompt path="~/work/my-app" branch="feat/auth-flow" dirty cmd="moai statusline --watch" />
      {"\n"}
      <div style={{ color: t.dim, fontFamily: '"Pretendard", sans-serif' }}>tmux · zsh · vim 등에 임베드 가능한 한 줄 상태선</div>
      {"\n"}
      <Rule label="기본" n={88} />
      <div style={{ background: t.panel, padding: "6px 10px", borderRadius: 6, border: `1px solid ${t.rule}` }}>
        <C c="accent" b>● MoAI</C>{`  `}<Tag kind="primary">AUTH-001</Tag>{`  `}<C c="fg" b>GREEN</C>{`  `}<C c="dim">cycle 3/6</C>{`  `}<C c="success">✓ 218</C>{`  `}<C c="success">92.4%</C>{`  `}<C c="warning">! 3</C>{`  `}<C c="dim">~/work/my-app</C>{`  `}<C c="info">feat/auth-flow ✗</C></div>
      {"\n"}
      <Rule label="조립" n={88} />
      {Line("프로젝트", <><C c="accent" b>● MoAI</C>{`  `}<C c="dim">SuperAgent 활성</C></>)}
      {Line("SPEC", <><Tag kind="primary">AUTH-001</Tag>{` `}<C c="fg">62%</C></>)}
      {Line("단계", <><C c="fg" b>GREEN</C>{` `}<C c="dim">RED → </C><C c="accent" b>GREEN</C><C c="dim"> → REFACTOR</C></>)}
      {Line("루프", <><C c="dim">cycle</C> <C c="fg" b>3 / 6</C>{`  `}<C c="dim">진행</C> <C c="accent">{`▰▰▰▱▱▱`}</C></>)}
      {Line("테스트", <><C c="success" b>✓ 218</C>{`  `}<C c="dim">12.4s</C></>)}
      {Line("커버리지", <><C c="success" b>92.4%</C></>)}
      {Line("Lint", <><C c="warning" b>! 3</C>{` `}<C c="dim">internal/auth/oauth.go</C></>)}
      {Line("Git", <><C c="info">feat/auth-flow</C>{` `}<C c="warning">✗</C>{` `}<C c="dim">+12 / −4</C></>)}
      {"\n"}
      <Rule label="테마" n={88} />
      <div>  <C c="accent" b>→</C> <C c="fg">moai statusline --theme=teal-dark</C></div>
      <div>  <C c="accent" b>→</C> <C c="fg">moai statusline --theme=teal-light</C></div>
      <div>  <C c="accent" b>→</C> <C c="fg">moai statusline --format='&#123;spec&#125; &#123;phase&#125; &#123;tests&#125;'</C></div>
    </Term>
  );
}

// ── 17. install.ps1 (Windows) ───────────────────────────
function SceneInstallPs() {
  const t = useTok();
  return (
    <Term title="PowerShell — install.ps1" cols={92} footer={<><span>install.ps1 · Windows 11</span><span>x64 · MSI-free</span></>}>
      <div><C c="info">PS C:\Users\yuna&gt;</C> <C c="fg">irm https://moai-adk.dev/install.ps1 | iex</C></div>
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700 }}>{`╔══════════════════════════════════════════════════════════════╗`}</div>
      <div style={{ color: t.accent, fontWeight: 700 }}>{`║          MoAI's Agentic Development Kit Installer            ║`}</div>
      <div style={{ color: t.accent, fontWeight: 700 }}>{`╚══════════════════════════════════════════════════════════════╝`}</div>
      {"\n"}
      <div><Tag kind="info">INFO</Tag>    Architecture detection (Layer 1 · Env Vars): AMD64</div>
      <div><Tag kind="ok">SUCCESS</Tag> Detected platform: <C c="fg" b>windows_amd64</C></div>
      <div><Tag kind="info">INFO</Tag>    Latest version: <C c="accent" b>v3.2.4</C></div>
      <div><Tag kind="info">INFO</Tag>    Downloading moai-adk_3.2.4_windows_amd64.zip…</div>
      <div style={{ color: t.dim }}>{`        ████████████████████████████████████████████  100%   19.1 MB`}</div>
      <div><Tag kind="ok">SUCCESS</Tag> Checksum verified  <C c="dim">sha256: 3e4b…77a1</C></div>
      <div><Tag kind="ok">SUCCESS</Tag> Extracted to <C c="fg">$env:LOCALAPPDATA\Programs\moai\</C></div>
      <div><Tag kind="info">INFO</Tag>    Adding to PATH (User scope)…</div>
      <div><Tag kind="ok">SUCCESS</Tag> PATH updated  <C c="dim">새 PowerShell 세션부터 적용됩니다.</C></div>
      {"\n"}
      <Box width={88} title="설치 완료" accent>
        <div>{`│  `}<C c="dim">바이너리</C>     <C c="fg">$env:LOCALAPPDATA\Programs\moai\moai.exe</C></div>
        <div>{`│  `}<C c="dim">버전</C>         <C c="accent" b>3.2.4</C></div>
        <div>{`│  `}<C c="dim">PowerShell</C>   <C c="fg">5.1+ · 7.4 권장</C></div>
        <div>{`│  `}<C c="dim">다음</C>         <C c="fg">moai init my-app</C></div>
      </Box>
      {"\n"}
      <div><C c="info">PS C:\Users\yuna&gt;</C> <span style={{ color: t.cursor, animation: "blink 1.05s steps(1) infinite" }}>▌</span></div>
    </Term>
  );
}

// ── 18. windows install.bat ─────────────────────────────
function SceneInstallBat() {
  const t = useTok();
  return (
    <Term title="cmd.exe — install.bat" cols={92} footer={<><span>install.bat (오프라인 풀백)</span><span>cmd.exe</span></>}>
      <div><C c="info">C:\Users\yuna&gt;</C> <C c="fg">install.bat</C></div>
      {"\n"}
      <div style={{ color: t.accent, fontWeight: 700 }}>{`+--------------------------------------------------------------+`}</div>
      <div style={{ color: t.accent, fontWeight: 700 }}>{`|       MoAI-ADK Installer (cmd.exe / 오프라인 모드)           |`}</div>
      <div style={{ color: t.accent, fontWeight: 700 }}>{`+--------------------------------------------------------------+`}</div>
      {"\n"}
      <div>[<C c="info" b>INFO</C>]    PowerShell이 차단된 환경 감지 — cmd.exe 폴백 사용</div>
      <div>[<C c="info" b>INFO</C>]    바이너리 위치   <C c="fg">.\bin\moai.exe</C></div>
      <div>[<C c="ok" b>OK</C>]      checksum.txt 검증 통과</div>
      <div>[<C c="info" b>INFO</C>]    %LOCALAPPDATA%\Programs\moai 로 복사</div>
      <div>[<C c="ok" b>OK</C>]      복사 완료 (3 파일, 18.4 MB)</div>
      <div>[<C c="info" b>INFO</C>]    PATH 등록 — setx 명령</div>
      <div>[<C c="ok" b>OK</C>]      설치 완료. 새 cmd 세션을 열어주세요.</div>
      {"\n"}
      <div><C c="dim">사용법:</C></div>
      <div>  <C c="accent" b>moai init my-app</C>     <C c="dim">새 프로젝트</C></div>
      <div>  <C c="accent" b>moai doctor</C>          <C c="dim">진단</C></div>
      <div>  <C c="accent" b>moai version</C>         <C c="dim">버전 확인</C></div>
      {"\n"}
      <div><C c="info">C:\Users\yuna&gt;</C> <span style={{ color: t.cursor, animation: "blink 1.05s steps(1) infinite" }}>▌</span></div>
    </Term>
  );
}

// ── 19. moai help (compact reference card) ──────────────
function SceneHelp() {
  const t = useTok();
  const G = ({ title, items }) => (
    <div style={{ marginBottom: 10 }}>
      <div style={{ color: t.accent, fontWeight: 700, letterSpacing: 0 }}>{title}</div>
      {items.map(([cmd, desc], i) => (
        <div key={i}>  <C c="fg" b>{cmd.padEnd(28)}</C>  <C c="dim">{desc}</C></div>
      ))}
    </div>
  );
  return (
    <Term title="moai — help (reference)" cols={92} footer={<><span>moai help</span><span>v3.2.4</span></>}>
      <Prompt path="~" branch="" cmd="moai help" />
      {"\n"}
      <div style={{ color: t.fg, fontFamily: '"Pretendard", sans-serif', fontSize: 14, fontWeight: 700, letterSpacing: "-0.025em" }}>
        moai-adk · 명령어 레퍼런스
      </div>
      <div style={{ color: t.dim, marginBottom: 6 }}>모든 명령은 <C c="accent" b>--help</C> 또는 <C c="accent" b>-h</C> 로 자세한 옵션을 볼 수 있습니다.</div>
      {"\n"}
      <G title="Launch" items={[
        ["moai cc",    "Claude Code 런처 (단일 모델)"],
        ["moai cg",    "Claude · GLM 콤보 런처"],
        ["moai glm",   "GLM 단일 모델 런처"],
      ]} />
      <G title="Project" items={[
        ["moai init [name]",   "프로젝트 초기화 (대화형 위저드)"],
        ["moai update",        "템플릿 / SKILL / Hook 동기화"],
        ["moai doctor",        "시스템 진단 (--fix, --verbose)"],
        ["moai status",        "현재 프로젝트 상태"],
        ["moai version",       "버전 정보"],
      ]} />
      <G title="SPEC & Loop" items={[
        ["moai spec view ID",     "SPEC 본문 보기"],
        ["moai spec lint",        "SPEC 형식 검사"],
        ["moai spec status",      "전체 SPEC 진행 현황"],
        ["moai loop start",       "자율 개발 루프 시작"],
        ["moai loop pause | resume | cancel", "루프 제어"],
      ]} />
      <G title="Git Worktree" items={[
        ["moai worktree new <branch>",     "새 worktree (SPEC 자동 연결)"],
        ["moai worktree list | switch | sync | done | clean | recover", "관리 명령"],
      ]} />
      <G title="Tools" items={[
        ["moai astgrep <pattern>",  "AST-grep 패턴 검색"],
        ["moai mx graph | query",   "@MX 앵커 / 의존성 그래프"],
        ["moai brain",              "Knowledge Graph & Memory"],
        ["moai constitution check", "CX 7원칙 검증"],
        ["moai harness run",        "테스트 하네스 / 품질 게이트"],
        ["moai telemetry report",   "로컬 텔레메트리 리포트"],
        ["moai statusline",         "tmux/vim 임베드용 상태선"],
        ["moai research",           "프로젝트 리서치 노트"],
        ["moai mcp lsp",            "LSP / MCP 진단"],
      ]} />
      <G title="GitHub" items={[
        ["moai github init",        "원격 저장소 + 보호 브랜치 자동 설정"],
        ["moai github auth",        "gh CLI 인증 동기화"],
        ["moai github status",      "PR / Actions 요약"],
      ]} />
      {"\n"}
      <Prompt cmd="" />
    </Term>
  );
}

Object.assign(window, {
  SceneInstallSh, SceneBanner, SceneInit, SceneDoctor, SceneStatus, SceneVersion,
  SceneCC, SceneLoop, SceneSpec, SceneUpdate, SceneWorktree, SceneConstitution,
  SceneMX, SceneTelemetry, SceneError, SceneStatusline, SceneInstallPs, SceneInstallBat,
  SceneHelp,
});
