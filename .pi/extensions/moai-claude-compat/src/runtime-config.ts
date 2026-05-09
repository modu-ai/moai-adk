import { existsSync, readFileSync } from "node:fs";
import { dirname, isAbsolute, relative, resolve } from "node:path";

export const COMPAT_ROOT = ".pi/claude-compat";

export const COMPAT_MANIFEST_FILES = {
  commandMap: `${COMPAT_ROOT}/command-map.json`,
  hookEvents: `${COMPAT_ROOT}/hook-events.json`,
  outputStyles: `${COMPAT_ROOT}/output-styles.json`,
  packageMap: `${COMPAT_ROOT}/package-map.json`,
  permissions: `${COMPAT_ROOT}/permissions.json`,
  sourceMap: `${COMPAT_ROOT}/source-map.json`,
  toolAliases: `${COMPAT_ROOT}/tool-aliases.json`,
  workflowMap: `${COMPAT_ROOT}/workflow-map.json`,
} as const;

type ManifestKey = keyof typeof COMPAT_MANIFEST_FILES;

export interface CommandMapConfig {
  version: number;
  commands: Record<string, { source?: string; router?: string }>;
  moaiSubcommands: string[];
}

export interface WorkflowMapConfig {
  version: number;
  workflows: Record<string, { source?: string; modeSource?: string; primaryAgents?: string[] }>;
  teamBackendPriority: string[];
  tddDdd: string;
}

export interface ToolAliasesConfig {
  version: number;
  aliases: Record<string, string>;
  notes: string[];
}

export interface PermissionsConfig {
  version: number;
  policy: {
    permissionMode: string;
    allowAskDeny: string;
    nonInteractiveAskDefault: string;
  };
  ask: string[];
  deny: string[];
  guardrailPackages: string[];
}

export interface HookEventsConfig {
  version: number;
  mapping: Record<string, string>;
  runtimeBridge: {
    connected: string[];
    statusOnly: string[];
    failurePolicy: string;
  };
  primaryPackage: string;
  fallbackPackage: string;
}

export interface OutputStylesConfig {
  version: number;
  default: string;
  styles: Record<string, { source?: string }>;
  runtime: {
    mechanism: string;
    userFacingXml: string;
    internalCompletionMarker: string;
  };
}

export interface SourceMapConfig {
  version: number;
  generatedAt: string;
  policy: Record<string, unknown>;
  sources: Record<string, string>;
  generatedRoot: string;
  compatRoot: string;
  expectedCounts: Record<string, number>;
  packagePriority: Record<string, string[]>;
}

export interface PackageMapConfig {
  version: number;
  primary: Record<string, string>;
  fallbacks: Record<string, string[]>;
  conflictPolicy: Record<string, string>;
}

export interface LoadedManifestFile<T> {
  key: ManifestKey;
  path: string;
  absPath: string;
  exists: boolean;
  loaded: boolean;
  config: T;
  fallbackUsed: boolean;
  error?: string;
}

export interface RuntimeManifest {
  cwd: string;
  files: typeof COMPAT_MANIFEST_FILES;
  commandMap: LoadedManifestFile<CommandMapConfig>;
  hookEvents: LoadedManifestFile<HookEventsConfig>;
  outputStyles: LoadedManifestFile<OutputStylesConfig>;
  packageMap: LoadedManifestFile<PackageMapConfig>;
  permissions: LoadedManifestFile<PermissionsConfig>;
  sourceMap: LoadedManifestFile<SourceMapConfig>;
  toolAliases: LoadedManifestFile<ToolAliasesConfig>;
  workflowMap: LoadedManifestFile<WorkflowMapConfig>;
}

const DEFAULT_COMMAND_MAP: CommandMapConfig = {
  version: 1,
  commands: {
    moai: { source: "./generated/source/skills/moai/SKILL.md", router: "moai-claude-compat/src/command-router.ts" },
    github: { source: "./generated/source/commands/98-github.md" },
    release: { source: "./generated/source/commands/99-release.md" },
  },
  moaiSubcommands: [
    "brain",
    "plan",
    "run",
    "sync",
    "design",
    "db",
    "project",
    "feedback",
    "fix",
    "loop",
    "mx",
    "review",
    "clean",
    "codemaps",
    "coverage",
    "e2e",
    "gate",
    "security",
  ],
};

const DEFAULT_WORKFLOW_MAP: WorkflowMapConfig = {
  version: 1,
  workflows: {
    brain: { source: "./generated/source/skills/moai/workflows/brain.md", primaryAgents: ["manager-brain"] },
    plan: { source: "./generated/source/skills/moai/workflows/plan.md", primaryAgents: ["manager-spec", "manager-strategy"] },
    run: {
      source: "./generated/source/skills/moai/workflows/run.md",
      modeSource: "./generated/source/moai-config/sections/quality.yaml",
      primaryAgents: ["manager-tdd", "manager-ddd"],
    },
    sync: {
      source: "./generated/source/skills/moai/workflows/sync.md",
      primaryAgents: ["manager-docs", "manager-quality", "manager-git"],
    },
    design: { source: "./generated/source/skills/moai/workflows/design.md" },
    db: { source: "./generated/source/skills/moai/workflows/db.md" },
    project: { source: "./generated/source/skills/moai/workflows/project.md", primaryAgents: ["manager-project", "manager-docs"] },
    feedback: { source: "./generated/source/skills/moai/workflows/feedback.md", primaryAgents: ["manager-quality"] },
    fix: { source: "./generated/source/skills/moai/workflows/fix.md", primaryAgents: ["expert-debug"] },
    loop: { source: "./generated/source/skills/moai/workflows/loop.md", primaryAgents: ["expert-debug", "expert-testing"] },
    mx: { source: "./generated/source/skills/moai/workflows/mx.md" },
    review: { source: "./generated/source/skills/moai/workflows/review.md", primaryAgents: ["manager-quality", "expert-security"] },
    clean: { source: "./generated/source/skills/moai/workflows/clean.md", primaryAgents: ["expert-refactoring", "expert-testing"] },
    codemaps: { source: "./generated/source/skills/moai/workflows/codemaps.md", primaryAgents: ["manager-docs"] },
    coverage: { source: "./generated/source/skills/moai/workflows/coverage.md", primaryAgents: ["expert-testing"] },
    e2e: { source: "./generated/source/skills/moai/workflows/e2e.md", primaryAgents: ["expert-testing", "expert-frontend"] },
    gate: { source: "./generated/source/skills/moai/workflows/gate.md" },
    security: { source: "./generated/source/skills/moai/workflows/security.md", primaryAgents: ["expert-security"] },
  },
  teamBackendPriority: ["@tmustier/pi-agent-teams", "pi-teams", "pi-crew"],
  tddDdd: "read .pi/generated/source/moai-config/sections/quality.yaml",
};

const DEFAULT_TOOL_ALIASES: ToolAliasesConfig = {
  version: 1,
  aliases: {
    Read: "read",
    Write: "write",
    Edit: "edit",
    MultiEdit: "edit",
    Grep: "bash:rg",
    Glob: "bash:find",
    Bash: "bash",
    TodoWrite: "@juicesharp/rpiv-todo",
    Agent: "pi-subagents:subagent",
    Skill: "pi skills/read",
    AskUserQuestion: "@juicesharp/rpiv-ask-user-question",
    WebSearch: "pi-web-access:web_search",
    WebFetch: "pi-web-access:fetch_content",
    TeamCreate: "@tmustier/pi-agent-teams",
    SendMessage: "@tmustier/pi-agent-teams",
    TaskCreate: "@tmustier/pi-agent-teams or pi-crew task state",
    TaskUpdate: "@tmustier/pi-agent-teams or pi-crew task state",
    TaskList: "@tmustier/pi-agent-teams or pi-crew task state",
    TaskGet: "@tmustier/pi-agent-teams or pi-crew task state",
    TeamDelete: "@tmustier/pi-agent-teams",
  },
  notes: [
    "MultiEdit maps to one pi edit call with multiple edits.",
    "Subagents cannot ask the user directly; parent must bridge AskUserQuestion.",
    "permissionMode is intentionally not mapped.",
  ],
};

const DEFAULT_PERMISSIONS: PermissionsConfig = {
  version: 1,
  policy: {
    permissionMode: "excluded-by-design",
    allowAskDeny: "semantic-equivalent guardrails",
    nonInteractiveAskDefault: "block",
  },
  ask: ["Bash(rm*)", "Bash(sudo*)", "Bash(chmod*)", "Bash(chown*)", "Read(.env*)"],
  deny: [
    "Read/Write/Edit/Grep/Glob(./secrets/**)",
    "Read/Write/Edit/Grep/Glob(~/.ssh/**)",
    "Read/Write/Edit/Grep/Glob(~/.aws/**)",
    "Read/Write/Edit/Grep/Glob(~/.config/gcloud/**)",
    "Bash(git push --force*)",
    "Bash(git push -f*)",
    "Bash(git push --force-with-lease*)",
    "Bash(git reset --hard*)",
    "Bash(git clean -fd*)",
    "Bash(git clean -fdx*)",
    "Bash(git rebase -i*)",
    "Bash(chmod 777*)",
    "Bash(chmod -R 777*)",
    "Bash(DROP DATABASE*)",
    "Bash(DROP TABLE*)",
    "Bash(TRUNCATE*)",
    "Bash(DELETE FROM*)",
    "Bash(redis-cli FLUSHALL*)",
    "Bash(redis-cli FLUSHDB*)",
    "Bash(psql -c DROP*)",
    "Bash(mysql -e DROP*)",
    "Bash(curl|wget pipe-to-shell*)",
  ],
  guardrailPackages: ["@gotgenes/pi-permission-system", "@aliou/pi-guardrails", "pi-yaml-hooks"],
};

const DEFAULT_HOOK_EVENTS: HookEventsConfig = {
  version: 1,
  mapping: {
    SessionStart: "session_start",
    PreCompact: "session_before_compact",
    PostCompact: "session_after_compact (mapped script present; no native pi lifecycle event)",
    SessionEnd: "session_shutdown",
    PreToolUse: "tool_call",
    PostToolUse: "tool_result",
    PostToolUseFailure: "tool_result:isError",
    Stop: "agent_end",
    StopFailure: "agent_end failure wrapper (mapped script present; no native pi lifecycle event)",
    SubagentStart: "team/subagent wrapper start (mapped script present; runtime event depends on team backend)",
    SubagentStop: "team/subagent wrapper end (mapped script present; runtime event depends on team backend)",
    AgentHook: "agent hook wrapper (mapped script present; no native pi lifecycle event)",
    Notification: "ui.notify/custom event (mapped script present; no native global interception)",
    UserPromptSubmit: "input",
    PermissionRequest: "guardrail confirm (mapped script present; native ask/deny handled by guardrails)",
    TeammateIdle: "agent-teams lifecycle or fallback state (mapped script present; runtime event depends on team backend)",
    TaskCreated: "agent-teams task create event or fallback state (mapped script present; runtime event depends on team backend)",
    TaskCompleted: "agent-teams task event or fallback state (mapped script present; runtime event depends on team backend)",
    WorktreeCreate: "worktree wrapper (mapped script present; runtime event depends on worktree backend)",
    WorktreeRemove: "worktree wrapper (mapped script present; runtime event depends on worktree backend)",
    ConfigChange: "configuration change wrapper (mapped script present; no native pi lifecycle event)",
    CwdChanged: "cwd changed wrapper (mapped script present; no native pi lifecycle event)",
    Elicitation: "elicitation request wrapper (mapped script present; AskUserQuestion package owns runtime UI)",
    ElicitationResult: "elicitation result wrapper (mapped script present; AskUserQuestion package owns runtime UI)",
    FileChanged: "file changed wrapper (mapped script present; no native pi lifecycle event)",
    InstructionsLoaded: "instructions loaded wrapper (mapped script present; no native pi lifecycle event)",
  },
  runtimeBridge: {
    connected: ["SessionStart", "PreCompact", "SessionEnd", "UserPromptSubmit", "PreToolUse", "PostToolUse", "PostToolUseFailure", "Stop", "Notification"],
    statusOnly: ["PostCompact", "StopFailure", "SubagentStart", "SubagentStop", "AgentHook", "PermissionRequest", "TeammateIdle", "TaskCreated", "TaskCompleted", "WorktreeCreate", "WorktreeRemove", "ConfigChange", "CwdChanged", "Elicitation", "ElicitationResult", "FileChanged", "InstructionsLoaded"],
    failurePolicy: "non-blocking warning",
  },
  primaryPackage: "pi-yaml-hooks",
  fallbackPackage: "pi-autohooks",
};

const DEFAULT_OUTPUT_STYLES: OutputStylesConfig = {
  version: 1,
  default: "moai",
  styles: {
    moai: { source: "./generated/source/output-styles/moai/moai.md" },
    einstein: { source: "./generated/source/output-styles/moai/einstein.md" },
  },
  runtime: {
    mechanism: "prompt-injection + ctx.ui.setStatus/widget",
    userFacingXml: "forbidden",
    internalCompletionMarker: "extension-state-only",
  },
};

const DEFAULT_SOURCE_MAP: SourceMapConfig = {
  version: 1,
  generatedAt: "1970-01-01T00:00:00.000Z",
  policy: {
    permissionMode: "excluded-by-design; record as metadata only; do not enforce runtime semantics",
    packageFirst: true,
  },
  sources: {
    claudeSettings: "./generated/source/settings/claude-settings.json",
    commands: "./generated/source/commands",
    agents: "./generated/source/agents/moai",
    skills: "./generated/source/skills",
    rules: "./generated/source/rules/moai",
    outputStyles: "./generated/source/output-styles/moai",
    hooks: "./generated/source/hooks/moai",
    moaiConfig: "./generated/source/moai-config/sections",
  },
  generatedRoot: "./generated",
  compatRoot: "./claude-compat",
  expectedCounts: { agents: 25, skills: 38, outputStyles: 2, rules: 46, hooks: 28, commands: 19 },
  packagePriority: {
    agentTeams: ["@tmustier/pi-agent-teams", "pi-teams", "pi-crew"],
    hooks: ["pi-yaml-hooks", "pi-autohooks", "moai-claude-compat"],
    permissions: ["@gotgenes/pi-permission-system", "@aliou/pi-guardrails", "pi-yaml-hooks", "moai-claude-compat"],
  },
};

const DEFAULT_PACKAGE_MAP: PackageMapConfig = {
  version: 1,
  primary: {
    subagents: "pi-subagents",
    agentTeams: "@tmustier/pi-agent-teams",
    askUserQuestion: "@juicesharp/rpiv-ask-user-question",
    todo: "@juicesharp/rpiv-todo",
    hooks: "pi-yaml-hooks",
    permissions: "@gotgenes/pi-permission-system",
    guardrails: "@aliou/pi-guardrails",
    worktrees: "@zenobius/pi-worktrees",
    mcp: "pi-mcp-adapter",
    web: "pi-web-access",
    documents: "pi-docparser",
    context: "context-mode",
    planning: "@ifi/pi-plan",
    triggerPreload: "pi-prompt-template-model",
    quotaFooter: "moai-claude-compat-native-codex-quota",
  },
  fallbacks: {
    agentTeams: ["pi-teams", "pi-crew"],
    quotaFooter: ["pi-chatgpt-limit", "pi-usage"],
    memory: ["pi-total-recall", "claude-recall"],
    ui: ["pi-claude-style-tools", "@plannotator/pi-extension"],
  },
  conflictPolicy: {
    footerQuota: "MoAI native quota footer owns Codex quota display; do not also enable @kmiyh/pi-codex-plan-limits, pi-chatgpt-limit, pi-codexbar",
    agentTeams: "select exactly one primary team backend per run; pi-subagents is tracked separately for Agent tool delegation, not as a team backend fallback",
    contextHooks: "context-mode and pi-yaml-hooks can both be active, but hook/guidance ordering must be validated after restart",
    mcp: "prefer one MCP route per workflow if pi-mcp-adapter and the native Pi MCP gateway both expose similar tools",
    permissionMode: "ignore runtime semantics by design",
  },
};

const FALLBACKS = {
  commandMap: DEFAULT_COMMAND_MAP,
  hookEvents: DEFAULT_HOOK_EVENTS,
  outputStyles: DEFAULT_OUTPUT_STYLES,
  packageMap: DEFAULT_PACKAGE_MAP,
  permissions: DEFAULT_PERMISSIONS,
  sourceMap: DEFAULT_SOURCE_MAP,
  toolAliases: DEFAULT_TOOL_ALIASES,
  workflowMap: DEFAULT_WORKFLOW_MAP,
};

let cachedManifest: RuntimeManifest | undefined;
let cachedCwd = "";

export function projectAbsPath(path: string, cwd = process.cwd()): string {
  return isAbsolute(path) ? path : resolve(cwd, path);
}

export function toProjectRelativePath(absPath: string, cwd = process.cwd()): string {
  const rel = relative(cwd, absPath).replace(/\\/g, "/");
  return rel.startsWith("..") ? absPath : rel || ".";
}

export function readTextIfExists(path: string, cwd = process.cwd()): string {
  const abs = projectAbsPath(path, cwd);
  return existsSync(abs) ? readFileSync(abs, "utf8") : "";
}

export function pathExists(path: string, cwd = process.cwd()): boolean {
  return existsSync(projectAbsPath(path, cwd));
}

export function resolveCompatSource(configPath: string, source: string | undefined, cwd = process.cwd()): string {
  if (!source?.trim()) return "";
  if (isAbsolute(source)) return source;

  const normalized = source.replace(/^\.\//, "");
  const piRelative = resolve(cwd, ".pi", normalized);
  if (existsSync(piRelative) || normalized.startsWith("generated/") || normalized.startsWith("prompts/") || normalized.startsWith("extensions/")) {
    return toProjectRelativePath(piRelative, cwd);
  }

  const configRelative = resolve(cwd, dirname(configPath), source);
  if (existsSync(configRelative)) return toProjectRelativePath(configRelative, cwd);

  return toProjectRelativePath(piRelative, cwd);
}

export function resolveSourceMapPath(sourceMap: SourceMapConfig, sourceKey: string, cwd = process.cwd()): string {
  return resolveCompatSource(COMPAT_MANIFEST_FILES.sourceMap, sourceMap.sources[sourceKey], cwd);
}

export function getDefaultOutputStyleSource(manifest: RuntimeManifest): string {
  const outputStyles = manifest.outputStyles.config;
  return resolveCompatSource(COMPAT_MANIFEST_FILES.outputStyles, outputStyles.styles[outputStyles.default]?.source, manifest.cwd);
}

export function getToolAlias(manifest: RuntimeManifest, claudeToolName: string): string | undefined {
  return manifest.toolAliases.config.aliases[claudeToolName];
}

export function getPackagePriority(manifest: RuntimeManifest, key: string): string[] {
  const sourceMapPriority = manifest.sourceMap.config.packagePriority[key];
  if (Array.isArray(sourceMapPriority) && sourceMapPriority.length) return sourceMapPriority;
  const primary = manifest.packageMap.config.primary[key];
  const fallbacks = manifest.packageMap.config.fallbacks[key] ?? [];
  return primary ? [primary, ...fallbacks] : fallbacks;
}

export function getPermissionPolicy(manifest: RuntimeManifest): PermissionsConfig["policy"] {
  return manifest.permissions.config.policy;
}

export function getHookMapping(manifest: RuntimeManifest, claudeEventName: string): string | undefined {
  return manifest.hookEvents.config.mapping[claudeEventName];
}

export function loadRuntimeManifest(cwd = process.cwd(), options: { useCache?: boolean } = {}): RuntimeManifest {
  if (options.useCache !== false && cachedManifest && cachedCwd === cwd) return cachedManifest;

  const manifest: RuntimeManifest = {
    cwd,
    files: COMPAT_MANIFEST_FILES,
    commandMap: loadManifestFile("commandMap", cwd),
    hookEvents: loadManifestFile("hookEvents", cwd),
    outputStyles: loadManifestFile("outputStyles", cwd),
    packageMap: loadManifestFile("packageMap", cwd),
    permissions: loadManifestFile("permissions", cwd),
    sourceMap: loadManifestFile("sourceMap", cwd),
    toolAliases: loadManifestFile("toolAliases", cwd),
    workflowMap: loadManifestFile("workflowMap", cwd),
  };

  cachedManifest = manifest;
  cachedCwd = cwd;
  return manifest;
}

export function clearRuntimeManifestCache(): void {
  cachedManifest = undefined;
  cachedCwd = "";
}

export function runtimeManifestStatus(manifest = loadRuntimeManifest()): string[] {
  const keys = Object.keys(COMPAT_MANIFEST_FILES) as ManifestKey[];
  return keys.map((key) => {
    const file = manifest[key];
    if (file.loaded && !file.fallbackUsed) return `ok: ${file.path} runtime manifest loaded`;
    if (file.exists) return `partial: ${file.path} invalid; fallback active${file.error ? ` (${file.error})` : ""}`;
    return `missing: ${file.path}; fallback active`;
  });
}

function loadManifestFile<K extends ManifestKey>(key: K, cwd: string): LoadedManifestFile<(typeof FALLBACKS)[K]> {
  const path = COMPAT_MANIFEST_FILES[key];
  const absPath = projectAbsPath(path, cwd);
  const fallback = FALLBACKS[key];

  if (!existsSync(absPath)) {
    return { key, path, absPath, exists: false, loaded: false, config: fallback, fallbackUsed: true, error: `missing ${path}` };
  }

  try {
    const parsed = JSON.parse(readFileSync(absPath, "utf8"));
    if (!isRecord(parsed)) throw new Error("manifest root must be an object");
    return { key, path, absPath, exists: true, loaded: true, config: mergeFallback(fallback, parsed), fallbackUsed: false };
  } catch (error) {
    return {
      key,
      path,
      absPath,
      exists: true,
      loaded: false,
      config: fallback,
      fallbackUsed: true,
      error: error instanceof Error ? error.message : String(error),
    };
  }
}

function mergeFallback<T>(fallback: T, parsed: unknown): T {
  if (!isRecord(fallback) || !isRecord(parsed)) return parsed as T;
  const out: Record<string, unknown> = { ...fallback };
  for (const [key, value] of Object.entries(parsed)) {
    const fallbackValue = (fallback as Record<string, unknown>)[key];
    out[key] = isRecord(fallbackValue) && isRecord(value) ? mergeFallback(fallbackValue, value) : value;
  }
  return out as T;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
