// Mock data for MoAI-ADK v3 console
window.MoaiData = (() => {
  const SPECS = [
    { id: "SPEC-V3-CLI-001", title: "Permission bubble model with multi-source provenance", domain: "CLI", status: "in-progress", priority: "critical", phase: "run", owner: "manager-cycle", coverage: 73, ears: 8, mxTags: 4, updated: "2시간 전" },
    { id: "SPEC-V3-AGT-002", title: "manager-cycle 통합 (DDD + TDD 병합)", domain: "AGT", status: "in-progress", priority: "high", phase: "run", owner: "manager-cycle", coverage: 91, ears: 12, mxTags: 2, updated: "27분 전" },
    { id: "SPEC-V3-AGT-003", title: "builder-platform 통합 (agent/skill/plugin)", domain: "AGT", status: "review", priority: "high", phase: "sync", owner: "evaluator-active", coverage: 100, ears: 6, mxTags: 0, updated: "1시간 전" },
    { id: "SPEC-V3-HOOKS-001", title: "JSON 응답 프로토콜 마이그레이션", domain: "HOOKS", status: "blocked", priority: "high", phase: "plan", owner: "manager-strategy", coverage: 12, ears: 9, mxTags: 7, updated: "5시간 전" },
    { id: "SPEC-V3-HOOKS-002", title: "subagentStop tmux pane cleanup", domain: "HOOKS", status: "in-progress", priority: "critical", phase: "run", owner: "expert-debug", coverage: 64, ears: 4, mxTags: 1, updated: "12분 전" },
    { id: "SPEC-V3-HOOKS-003", title: "하드코딩 절대 경로 제거", domain: "HOOKS", status: "done", priority: "critical", phase: "sync", owner: "manager-cycle", coverage: 100, ears: 3, mxTags: 0, updated: "어제" },
    { id: "SPEC-V3-MEM-001", title: "Sprint Contract 영속화 + 평가자 메모리 분리", domain: "MEM", status: "in-progress", priority: "high", phase: "run", owner: "evaluator-active", coverage: 58, ears: 11, mxTags: 3, updated: "44분 전" },
    { id: "SPEC-V3-SCH-001", title: "Typed state schema + checkpoint 프로토콜", domain: "SCH", status: "review", priority: "high", phase: "sync", owner: "manager-strategy", coverage: 96, ears: 14, mxTags: 1, updated: "2시간 전" },
    { id: "SPEC-V3-SKL-002", title: "Skill 표면 축소 (48 → 24)", domain: "SKL", status: "in-progress", priority: "medium", phase: "plan", owner: "manager-project", coverage: 35, ears: 7, mxTags: 5, updated: "3시간 전" },
    { id: "SPEC-V3-CLN-001", title: "Constitutional sprawl 정리", domain: "CLN", status: "todo", priority: "medium", phase: "plan", owner: "—", coverage: 0, ears: 5, mxTags: 0, updated: "내일 시작" },
    { id: "SPEC-V3-OUT-001", title: "@MX 태그 ACI 표준화", domain: "OUT", status: "in-progress", priority: "medium", phase: "run", owner: "expert-refactoring", coverage: 47, ears: 8, mxTags: 6, updated: "1일 전" },
    { id: "SPEC-V3-PLG-001", title: "샌드박스 기본 적용 — Bubblewrap/Seatbelt", domain: "PLG", status: "in-progress", priority: "critical", phase: "run", owner: "expert-security", coverage: 22, ears: 16, mxTags: 9, updated: "30분 전" },
  ];

  const EARS_SAMPLE = [
    { id: "REQ-001", type: "ubiquitous", text: { pre: "시스템은", kw: "항상", post: "권한 결정의 출처(source)를 권한 응답 객체에 포함하여 반환해야 한다." }, status: "pass" },
    { id: "REQ-002", type: "event-driven", text: { pre: "에이전트가 위험 등급(risk≥medium) 도구를 호출", kw: "WHEN", post: ", 시스템은 부모 터미널에 권한 승인을 bubble 모드로 요청해야 한다." }, status: "pass" },
    { id: "REQ-003", type: "state-driven", text: { pre: "permission_mode가 'bypassPermissions'", kw: "WHILE", post: " 상태일 때, 시스템은 모든 도구 호출을 감사 로그(audit.jsonl)에 기록해야 한다." }, status: "pass" },
    { id: "REQ-004", type: "unwanted", text: { pre: "권한 응답이 5초 이내 도착하지 않으면", kw: "IF/THEN", post: ", 시스템은 도구 호출을 거부하고 BlockerReport를 생성해야 한다." }, status: "running" },
    { id: "REQ-005", type: "optional", text: { pre: "사전 승인 목록(allowlist)이 활성화된 경우", kw: "WHERE", post: ", 시스템은 매칭되는 도구 호출에 대해 bubble을 생략할 수 있다." }, status: "pending" },
    { id: "REQ-006", type: "complex", text: { pre: "워크트리 분리 모드에서, 에이전트가 부모 컨텍스트에 쓰기 시도하면", kw: "WHEN", post: " 시스템은 거부 + BlockerReport + 부모 터미널 알림을 모두 생성해야 한다." }, status: "fail" },
    { id: "REQ-007", type: "ubiquitous", text: { pre: "시스템은", kw: "항상", post: " sandbox 호출 wrapper가 호스트 파일시스템 외부 쓰기를 차단함을 검증해야 한다." }, status: "pass" },
    { id: "REQ-008", type: "event-driven", text: { pre: "사용자가 명시적으로 'plan' 모드를 활성화", kw: "WHEN", post: " 했을 때, 모든 implementer 에이전트는 read-only로 동작해야 한다." }, status: "pass" },
  ];

  const AGENTS = [
    { id: "manager-spec", kind: "manager", role: "SPEC 작성·검증", isolation: "session", effort: "high", status: "live", touches: 3, lastRun: "방금 전", desc: "EARS 요구사항 작성, Given/When/Then 수용 기준 검증, SPEC 헌법성 검사." },
    { id: "manager-cycle", kind: "manager", role: "DDD/TDD 통합 사이클", isolation: "worktree", effort: "high", status: "live", touches: 8, lastRun: "12초 전", desc: "DDD ANALYZE→PRESERVE→IMPROVE / TDD RED→GREEN→REFACTOR 통합. 사이클 타입 파라미터로 분기." },
    { id: "manager-strategy", kind: "manager", role: "DAG 합성 + 라우팅", isolation: "session", effort: "high", status: "idle", touches: 0, lastRun: "8분 전", desc: "/moai plan → 의존성 DAG. reads/writes 명시 + 동시성 분석." },
    { id: "manager-project", kind: "manager", role: ".moai/project 관리", isolation: "session", effort: "medium", status: "idle", touches: 0, lastRun: "2시간 전", desc: "product/structure/tech.md 만 관리. CLI 모드는 분리됨." },
    { id: "manager-quality", kind: "manager", role: "TRUST 5 게이트", isolation: "session", effort: "high", status: "live", touches: 1, lastRun: "방금 전", desc: "테스트·린트·타입·커버리지·ast 검사. 의심형 평가자 스탠스." },
    { id: "manager-design", kind: "manager", role: "디자인 파이프라인", isolation: "session", effort: "high", status: "idle", touches: 0, lastRun: "—", desc: "/moai design 1차 진입점. 도메인 디자인 스킬로 라우팅." },
    { id: "manager-git", kind: "manager", role: "git 워크플로우", isolation: "session", effort: "medium", status: "idle", touches: 0, lastRun: "1시간 전", desc: "브랜치, 커밋, PR. 8 키워드 라우팅으로 정밀화." },
    { id: "expert-backend", kind: "expert", role: "백엔드 구현", isolation: "worktree", effort: "high", status: "live", touches: 4, lastRun: "30초 전", desc: "Go/Python/Node 구현. --deepthink performance 모드 흡수." },
    { id: "expert-frontend", kind: "expert", role: "프론트엔드 구현", isolation: "worktree", effort: "high", status: "idle", touches: 0, lastRun: "어제", desc: "React/Vue 구현. Pencil MCP 분리됨." },
    { id: "expert-refactoring", kind: "expert", role: "구조 개선", isolation: "worktree", effort: "high", status: "live", touches: 2, lastRun: "1분 전", desc: "ast-grep 기반 패턴 마이그레이션. @MX 태그 동기화." },
    { id: "expert-security", kind: "expert", role: "보안 감사", isolation: "session", effort: "xhigh", status: "warn", touches: 1, lastRun: "5분 전", desc: "OWASP Top 10 Agentic Apps. 샌드박스 invocation 검증." },
    { id: "expert-devops", kind: "expert", role: "DevOps", isolation: "session", effort: "high", status: "idle", touches: 0, lastRun: "어제", desc: "CI/CD, container, deployment." },
    { id: "builder-platform", kind: "builder", role: "에이전트·스킬·플러그인", isolation: "worktree", effort: "high", status: "idle", touches: 0, lastRun: "12시간 전", desc: "통합 빌더. artifact 파라미터로 분기 (3개 빌더 합병)." },
    { id: "evaluator-active", kind: "evaluator", role: "GAN 평가자", isolation: "session", effort: "xhigh", status: "live", touches: 0, lastRun: "방금 전", desc: "iteration 별 fresh 컨텍스트. Sprint Contract만 영속화." },
    { id: "plan-auditor", kind: "evaluator", role: "계획 감사", isolation: "session", effort: "xhigh", status: "idle", touches: 0, lastRun: "3시간 전", desc: "/moai plan 출력 회의주의 검증. 누락된 의존성 탐지." },
    { id: "researcher", kind: "researcher", role: "리서치 + 실험", isolation: "worktree", effort: "high", status: "idle", touches: 0, lastRun: "이틀 전", desc: "외부 도구 / 논문 조사. 실험은 워크트리 격리." },
  ];

  const HOOKS = [
    { event: "PreToolUse", state: "active", protocol: "json", calls: 3421, latency: "8ms", desc: "도구 입력 변형, 권한 결정 bubble" },
    { event: "PostToolUse", state: "active", protocol: "json", calls: 3398, latency: "12ms", desc: "@MX 태그 주입, 관찰값 기록" },
    { event: "UserPromptSubmit", state: "active", protocol: "exit", calls: 142, latency: "4ms", desc: "프롬프트 라우팅 힌트" },
    { event: "SubagentStop", state: "active", protocol: "json", calls: 87, latency: "22ms", desc: "tmux pane cleanup (P-H02 수정 적용)" },
    { event: "SessionStart", state: "active", protocol: "exit", calls: 4, latency: "61ms", desc: "memory hydration, CLAUDE.md 검증" },
    { event: "SessionEnd", state: "active", protocol: "exit", calls: 4, latency: "18ms", desc: "checkpoint flush" },
    { event: "Notification", state: "stub", protocol: "—", calls: 0, latency: "—", desc: "logging-only no-op (P-H01)" },
    { event: "ConfigChange", state: "stub", protocol: "—", calls: 0, latency: "—", desc: "yaml reload 미구현 (P-H15)" },
    { event: "FileChanged", state: "stub", protocol: "—", calls: 0, latency: "—", desc: "MX 재스캔 미구현 (P-H17)" },
    { event: "InstructionsLoaded", state: "stub", protocol: "—", calls: 0, latency: "—", desc: "40k char 검증 미구현" },
    { event: "Setup", state: "orphan", protocol: "—", calls: 0, latency: "—", desc: "wrapper 없음 — Go 핸들러 고아 (P-H03)" },
    { event: "PostToolFailure", state: "stub", protocol: "—", calls: 0, latency: "—", desc: "logging-only" },
  ];

  const SKILLS = [
    { name: "moai-foundation-thinking", group: "foundation", level: 2, size: "11.8KB", refs: 18, desc: "사고 트리거. philosopher + workflow-thinking 흡수." },
    { name: "moai-foundation-core", group: "foundation", level: 1, size: "6.2KB", refs: 22, desc: "TRUST 5 + 헌법 SSOT 참조." },
    { name: "moai-foundation-cc", group: "foundation", level: 2, size: "9.1KB", refs: 11, desc: "Claude Code 런타임 가이드." },
    { name: "moai-workflow-spec", group: "workflow", level: 3, size: "18.2KB", refs: 14, desc: "EARS 작성, Given/When/Then 패턴." },
    { name: "moai-workflow-research", group: "workflow", level: 2, size: "9.4KB", refs: 6, desc: "외부 리서치 런북. researcher 에이전트 흡수 후보." },
    { name: "moai-workflow-jit-docs", group: "workflow", level: 2, size: "12.7KB", refs: 8, desc: "JIT 문서 생성. docs-generation 흡수." },
    { name: "moai-workflow-pencil-integration", group: "workflow", level: 2, size: "8.4KB", refs: 4, desc: "Pencil MCP → 코드 파이프라인." },
    { name: "moai-cmd-fix", group: "cmd", level: 1, size: "5.3KB", refs: 3, desc: "Agentless 픽서. 단일 에이전트 파이프라인." },
    { name: "moai-cmd-coverage", group: "cmd", level: 1, size: "4.1KB", refs: 2, desc: "커버리지 갭 채움. 결정적 변환." },
    { name: "moai-cmd-loop", group: "cmd", level: 2, size: "7.8KB", refs: 5, desc: "Ralph 루프. fresh context iteration." },
    { name: "moai-domain-backend", group: "domain", level: 2, size: "10.2KB", refs: 12, desc: "백엔드 라우터. 22→12 키워드 축소." },
    { name: "moai-domain-frontend", group: "domain", level: 2, size: "9.8KB", refs: 9, desc: "프론트 라우터. NOT for: 디자인 작업." },
    { name: "moai-design-system", group: "design", level: 2, size: "13.6KB", refs: 6, desc: "design-craft + domain-uiux 흡수." },
    { name: "moai-tool-figma", group: "tool", level: 1, size: "4.9KB", refs: 2, desc: "Figma MCP only. 분리됨." },
    { name: "moai-tool-pencil", group: "tool", level: 1, size: "5.1KB", refs: 3, desc: "Pencil MCP only." },
    { name: "moai-platform-vercel", group: "platform", level: 1, size: "3.7KB", refs: 2, desc: "Vercel only — auth0/clerk/firebase 분리." },
  ];

  const RULES = [
    { name: "moai-constitution.md", path: ".claude/rules/moai/core/", lines: 266, kind: "FROZEN", desc: "TRUST 5 + 16-언어 중립성 + AskUserQuestion 규칙" },
    { name: "design-constitution.md", path: ".claude/rules/moai/design/", lines: 404, kind: "FROZEN", desc: "디자인 헌법 v3.3.0. 평가자 메모리 분리 amend" },
    { name: "agent-common-protocol.md", path: ".claude/rules/moai/core/", lines: 157, kind: "EVOLVABLE", desc: "공통 에이전트 프로토콜. AskUserQuestion 차단" },
    { name: "coding-standards.md", path: ".claude/rules/moai/core/", lines: 122, kind: "EVOLVABLE", desc: "[HARD] Go 하드코딩 금지" },
    { name: "worktree-integration.md", path: ".claude/rules/moai/workflow/", lines: 303, kind: "EVOLVABLE", desc: "워크트리 격리. team-protocol 흡수" },
    { name: "spec-workflow.md", path: ".claude/rules/moai/workflow/", lines: 217, kind: "EVOLVABLE", desc: "Plan-Run-Sync 흐름. workflow-modes 흡수" },
  ];

  const TRUST = [
    { letter: "T", name: "Tests Green", status: "ok", value: "147/147", note: "단위·통합·E2E 모두 통과. Mutation score 0.84." },
    { letter: "R", name: "Requirements Traceable", status: "ok", value: "100%", note: "EARS ↔ 코드 ↔ 테스트 매핑 완전." },
    { letter: "U", name: "Unified Style", status: "warn", value: "3 위반", note: "ruff: 2 / golangci: 1. 자동 픽서 가능." },
    { letter: "S", name: "Secured", status: "ok", value: "0 critical", note: "샌드박스 활성. OWASP A1, A8 검증 완료." },
    { letter: "T", name: "Trackable Trail", status: "ok", value: "@MX 87 tags", note: "모든 변경 git + @MX 추적 가능." },
  ];

  const PERMISSIONS = [
    { tool: "Read", scope: "전 경로", source: "builtin", risk: "low", mode: "default" },
    { tool: "Write", scope: ".moai/, src/, tests/", source: "project", risk: "medium", mode: "acceptEdits" },
    { tool: "Bash", scope: "go test, ruff, npm test", source: "project", risk: "medium", mode: "default" },
    { tool: "Bash (network)", scope: "github.com, registry.npmjs.org, pypi.org", source: "project", risk: "high", mode: "bubble" },
    { tool: "Bash (rm)", scope: "—", source: "policy", risk: "critical", mode: "deny" },
    { tool: "WebFetch", scope: "*.anthropic.com", source: "user", risk: "low", mode: "default" },
    { tool: "Agent (fork)", scope: "워크트리 격리", source: "builtin", risk: "medium", mode: "bubble" },
  ];

  // Build a streaming-style log seed
  const LOG_SEED = [
    ["10:42:01", "info", "manager-cycle", "▶ /moai run SPEC-V3-CLI-001 started — phase: run, isolation: worktree"],
    ["10:42:02", "dbg",  "harness",       "harness=standard | budget=12000 tokens | retries=3"],
    ["10:42:02", "info", "context",       "loaded SPEC document (8 EARS, 4 @MX) — 2,142 tokens"],
    ["10:42:04", "info", "expert-backend","writing internal/permission/bubble.go (worktree=wt-cli-001)"],
    ["10:42:09", "ok",   "expert-backend","✓ permission stack with provenance — 4 sources merged"],
    ["10:42:10", "info", "manager-quality","TRUST 5 gate: starting"],
    ["10:42:13", "ok",   "trust",         "T  tests green   147/147 ✓"],
    ["10:42:14", "ok",   "trust",         "R  reqs traceable 8/8 EARS ↔ tests ✓"],
    ["10:42:15", "warn", "trust",         "U  style          3 violations (ruff: 2, golangci: 1) — autofixable"],
    ["10:42:15", "ok",   "trust",         "S  secured        0 critical ✓"],
    ["10:42:16", "ok",   "trust",         "T  trackable      @MX 87 tags coherent ✓"],
    ["10:42:17", "info", "evaluator-active","starting fresh-context judgment (iteration 3/∞)"],
    ["10:42:21", "warn", "evaluator-active","REQ-006 — worktree write boundary not enforced in error path"],
    ["10:42:22", "info", "manager-cycle", "blocker_report.md written. surfacing to AskUserQuestion."],
    ["10:42:23", "dbg",  "checkpoint",    ".moai/state/checkpoint-cli-001-iter3.yaml flushed (3.4 KB)"],
  ];

  // 24h sparkline data (small numbers)
  const SPARKS = {
    runs:     [42,38,40,55,60,52,48,61,73,68,64,71,82,79,77,84,90,86,88,92,96,101,98,103],
    coverage: [78,79,79,80,80,81,82,82,83,82,83,84,84,84,85,85,86,86,86,87,87,88,88,89],
    failures: [3,4,3,5,4,4,3,2,2,3,3,2,2,1,2,1,1,2,1,1,0,1,0,0],
    cost:     [1.2,1.4,1.3,1.6,1.8,1.7,1.6,1.9,2.1,2.0,2.0,2.2,2.4,2.3,2.5,2.6,2.7,2.6,2.8,3.0,3.1,3.2,3.3,3.4],
  };

  return { SPECS, EARS_SAMPLE, AGENTS, HOOKS, SKILLS, RULES, TRUST, PERMISSIONS, LOG_SEED, SPARKS };
})();
