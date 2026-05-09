import { spawn } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { PI_HOOKS_SOURCE_PATH } from "./constants.ts";
import { COMPAT_MANIFEST_FILES } from "./runtime-config.ts";

const HOOK_EVENTS_CONFIG_PATH = COMPAT_MANIFEST_FILES.hookEvents;

export interface HookBridgeResult {
  ok: boolean;
  exitCode: number | null;
  stdout: string;
  stderr: string;
  skipped?: string;
}

type HookRuntimeState = "extension-connected" | "package-backed-adapter-available" | "package-backed-bridge-missing" | "adapter-needed" | "intentionally-excluded";

export interface HookRuntimeClassification {
  state: HookRuntimeState;
  detail: string;
}

interface HookEventsConfig {
  mapping?: Record<string, unknown>;
  runtimeBridge?: {
    connected?: unknown;
    statusOnly?: unknown;
    failurePolicy?: unknown;
  };
}

const DEFAULT_CLAUDE_EVENT_TO_BRIDGE_EVENT: Record<string, string> = {
  SessionStart: "session-start",
  PreCompact: "compact",
  SessionEnd: "session-end",
  PreToolUse: "pre-tool",
  PostToolUse: "post-tool",
  PostToolUseFailure: "post-tool-failure",
  Stop: "stop",
  AgentHook: "agent-hook",
  SubagentStart: "subagent-start",
  SubagentStop: "subagent-stop",
  Notification: "notification",
  UserPromptSubmit: "user-prompt-submit",
  PermissionRequest: "permission-request",
  TeammateIdle: "teammate-idle",
  TaskCompleted: "task-completed",
  WorktreeCreate: "worktree-create",
  WorktreeRemove: "worktree-remove",
};

const DEFAULT_CONNECTED_CLAUDE_EVENTS = [
  "SessionStart",
  "PreCompact",
  "SessionEnd",
  "UserPromptSubmit",
  "PreToolUse",
  "PostToolUse",
  "PostToolUseFailure",
  "Stop",
  "Notification",
];

const DEFAULT_HOOK_RUNTIME_CLASSIFICATION: Record<string, HookRuntimeClassification> = {
  "session-start": { state: "extension-connected", detail: "mapped from Pi session_start" },
  compact: { state: "extension-connected", detail: "mapped from Pi session_before_compact" },
  "session-end": { state: "extension-connected", detail: "mapped from Pi session_shutdown" },
  "pre-tool": { state: "extension-connected", detail: "mapped from Pi tool_call" },
  "post-tool": { state: "extension-connected", detail: "mapped from Pi tool_result" },
  stop: { state: "extension-connected", detail: "mapped from Pi agent_end for the main session" },
  "post-tool-failure": { state: "extension-connected", detail: "derived from Pi tool_result when isError is true" },
  "user-prompt-submit": { state: "extension-connected", detail: "mapped from Pi input" },
  "agent-hook": { state: "adapter-needed", detail: "Claude agent frontmatter hooks need a Pi agent metadata execution adapter" },
  "subagent-start": { state: "package-backed-bridge-missing", detail: "Pi subagent/team packages own worker lifecycle; no compat extension event is exposed for safe direct mapping" },
  "subagent-stop": { state: "package-backed-bridge-missing", detail: "Pi subagent/team packages own worker lifecycle; main-session agent_end is not a safe SubagentStop equivalent" },
  notification: { state: "extension-connected", detail: "mapped for MoAI compat internal notifyMoai calls; Pi-global notification interception is not available" },
  "permission-request": { state: "intentionally-excluded", detail: "Claude permissionMode parity is excluded by design; pi-yaml-hooks guardrails replace it" },
  "teammate-idle": { state: "package-backed-adapter-available", detail: "pi-agent-teams optional idle hooks can call project-local MoAI adapter scripts when explicitly enabled" },
  "task-completed": { state: "package-backed-adapter-available", detail: "pi-agent-teams optional task_completed hooks can call project-local MoAI adapter scripts when explicitly enabled" },
  "worktree-create": { state: "package-backed-bridge-missing", detail: "pi-agent-teams supports worktree creation internally, but exposes no compat extension hook for this lifecycle" },
  "worktree-remove": { state: "package-backed-bridge-missing", detail: "pi-agent-teams cleans worktrees internally, but exposes no compat extension hook for this lifecycle" },
};

function readHookEventsConfig(path = HOOK_EVENTS_CONFIG_PATH): HookEventsConfig | undefined {
  try {
    const abs = resolve(process.cwd(), path);
    if (!existsSync(abs)) return undefined;
    return JSON.parse(readFileSync(abs, "utf8")) as HookEventsConfig;
  } catch {
    return undefined;
  }
}

function asStringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
}

function toBridgeEventName(claudeEvent: string): string {
  return DEFAULT_CLAUDE_EVENT_TO_BRIDGE_EVENT[claudeEvent]
    ?? claudeEvent.replace(/([a-z0-9])([A-Z])/g, "$1-$2").replace(/_/g, "-").toLowerCase();
}

function loadClaudeEventToBridgeEvent(config = readHookEventsConfig()): Record<string, string> {
  return {
    ...DEFAULT_CLAUDE_EVENT_TO_BRIDGE_EVENT,
    ...Object.fromEntries(
      Object.keys(config?.mapping ?? {}).map((claudeEvent) => [claudeEvent, toBridgeEventName(claudeEvent)]),
    ),
  };
}

function bridgeEventDetailFromConfig(claudeEvent: string, config: HookEventsConfig | undefined): string | undefined {
  const detail = config?.mapping?.[claudeEvent];
  return typeof detail === "string" ? detail : undefined;
}

function loadConnectedBridgeEvents(config = readHookEventsConfig(), eventMap = loadClaudeEventToBridgeEvent(config)): string[] {
  const connectedClaudeEvents = asStringArray(config?.runtimeBridge?.connected);
  const source = connectedClaudeEvents.length ? connectedClaudeEvents : DEFAULT_CONNECTED_CLAUDE_EVENTS;
  return source.map((event) => eventMap[event] ?? toBridgeEventName(event));
}

function loadHookScriptByEvent(config = readHookEventsConfig(), eventMap = loadClaudeEventToBridgeEvent(config)): Record<string, string> {
  const bridgeEvents = new Set(Object.values(eventMap));
  for (const bridgeEvent of Object.keys(DEFAULT_HOOK_RUNTIME_CLASSIFICATION)) bridgeEvents.add(bridgeEvent);
  return Object.fromEntries(
    [...bridgeEvents].sort().map((bridgeEvent) => [bridgeEvent, `${PI_HOOKS_SOURCE_PATH}/handle-${bridgeEvent}.sh`]),
  );
}

function loadHookRuntimeClassification(config = readHookEventsConfig(), eventMap = loadClaudeEventToBridgeEvent(config)): Record<string, HookRuntimeClassification> {
  const connected = new Set(loadConnectedBridgeEvents(config, eventMap));
  const statusOnlyClaudeEvents = asStringArray(config?.runtimeBridge?.statusOnly);
  const out: Record<string, HookRuntimeClassification> = { ...DEFAULT_HOOK_RUNTIME_CLASSIFICATION };

  for (const [claudeEvent, bridgeEvent] of Object.entries(eventMap)) {
    const configDetail = bridgeEventDetailFromConfig(claudeEvent, config);
    if (connected.has(bridgeEvent)) {
      out[bridgeEvent] = {
        state: "extension-connected",
        detail: configDetail ? `mapped by ${HOOK_EVENTS_CONFIG_PATH}: ${configDetail}` : out[bridgeEvent]?.detail ?? "mapped by hook-events config",
      };
    }
  }

  for (const claudeEvent of statusOnlyClaudeEvents) {
    const bridgeEvent = eventMap[claudeEvent] ?? toBridgeEventName(claudeEvent);
    if (connected.has(bridgeEvent)) continue;
    const existing = out[bridgeEvent];
    const configDetail = bridgeEventDetailFromConfig(claudeEvent, config);
    out[bridgeEvent] = existing && existing.state !== "extension-connected"
      ? { ...existing, detail: configDetail ? `${existing.detail}; hook-events statusOnly: ${configDetail}` : existing.detail }
      : {
          state: "adapter-needed",
          detail: configDetail ? `hook-events statusOnly: ${configDetail}` : "listed as statusOnly in hook-events config",
        };
  }

  return out;
}

const hookEventsConfig = readHookEventsConfig();
const claudeEventToBridgeEvent = loadClaudeEventToBridgeEvent(hookEventsConfig);

export const HOOK_SCRIPT_BY_EVENT: Record<string, string> = loadHookScriptByEvent(hookEventsConfig, claudeEventToBridgeEvent);
export const CONNECTED_PI_HOOK_EVENTS = loadConnectedBridgeEvents(hookEventsConfig, claudeEventToBridgeEvent);
export const HOOK_RUNTIME_CLASSIFICATION: Record<string, HookRuntimeClassification> = loadHookRuntimeClassification(hookEventsConfig, claudeEventToBridgeEvent);

export const NON_BLOCKING_HOOK_BRIDGE_POLICY =
  (typeof hookEventsConfig?.runtimeBridge?.failurePolicy === "string" ? hookEventsConfig.runtimeBridge.failurePolicy : undefined) ??
  "extension hook bridge is non-blocking compatibility telemetry; blocking guardrails are enforced by pi-yaml-hooks tool.before.* policies";

export function unsupportedHookEvents(): string[] {
  const connected = new Set<string>(CONNECTED_PI_HOOK_EVENTS);
  return Object.keys(HOOK_SCRIPT_BY_EVENT).filter((event) => !connected.has(event));
}

export function hasHookScript(eventName: string): boolean {
  const script = HOOK_SCRIPT_BY_EVENT[eventName];
  return Boolean(script && existsSync(resolve(process.cwd(), script)));
}

export async function runMoaiHook(eventName: string, payload: unknown, timeoutMs = 10000): Promise<HookBridgeResult> {
  const script = HOOK_SCRIPT_BY_EVENT[eventName];
  if (!script) return { ok: true, exitCode: 0, stdout: "", stderr: "", skipped: `no hook script mapped for ${eventName}` };
  const abs = resolve(process.cwd(), script);
  if (!existsSync(abs)) return { ok: true, exitCode: 0, stdout: "", stderr: "", skipped: `hook script missing: ${script}` };

  return new Promise((resolveResult) => {
    const child = spawn("bash", [abs], {
      cwd: process.cwd(),
      env: { ...process.env, CLAUDE_PROJECT_DIR: process.cwd() },
      stdio: ["pipe", "pipe", "pipe"],
    });
    let stdout = "";
    let stderr = "";
    const timer = setTimeout(() => child.kill("SIGTERM"), timeoutMs);
    child.stdout.on("data", (chunk) => (stdout += String(chunk)));
    child.stderr.on("data", (chunk) => (stderr += String(chunk)));
    child.on("close", (exitCode) => {
      clearTimeout(timer);
      resolveResult({ ok: exitCode === 0, exitCode, stdout, stderr });
    });
    child.stdin.end(`${JSON.stringify(payload)}\n`);
  });
}

export function hookBridgeStatus(): string {
  const mapped = Object.keys(HOOK_SCRIPT_BY_EVENT);
  const present = mapped.filter(hasHookScript).length;
  const source = hookEventsConfig ? HOOK_EVENTS_CONFIG_PATH : "built-in fallback";
  if (present !== mapped.length) {
    return `missing: hook bridge script files present ${present}/${mapped.length} (mapping source: ${source})`;
  }
  return `info: hook bridge script files present ${present}/${mapped.length} (mapping source: ${source}; runtime wiring is partial)`;
}

export function hookRuntimeClassificationStatus(): string[] {
  const mapped = Object.keys(HOOK_SCRIPT_BY_EVENT);
  const byState = (state: HookRuntimeState) => mapped.filter((event) => HOOK_RUNTIME_CLASSIFICATION[event]?.state === state);
  const connected = byState("extension-connected");
  const packageBackedAvailable = byState("package-backed-adapter-available");
  const packageBackedMissing = byState("package-backed-bridge-missing");
  const adapterNeeded = byState("adapter-needed");
  const excluded = byState("intentionally-excluded");
  const notConnected = mapped.length - connected.length;
  return [
    `partial: extension-connected hook events ${connected.length}/${mapped.length}; ${notConnected} are package-backed, adapter-needed, or intentionally excluded`,
    packageBackedAvailable.length
      ? `partial: package-backed adapter-available hook events ${packageBackedAvailable.join(", ")} (requires explicit Agent Teams hook env activation)`
      : "ok: no package-backed hook adapters are waiting for activation",
    packageBackedMissing.length
      ? `partial: package-backed bridge-missing hook events ${packageBackedMissing.join(", ")} (Pi packages provide the lifecycle, but no safe compat extension event/adapter is installed yet)`
      : "ok: no package-backed hook bridges are missing",
    adapterNeeded.length
      ? `partial: adapter-needed hook events ${adapterNeeded.join(", ")} (compat extension needs explicit runtime adapters)`
      : "ok: no adapter-needed hook events",
    excluded.length
      ? `info: intentionally excluded hook events ${excluded.join(", ")} (replaced by Pi policy/guardrail mechanisms)`
      : "ok: no intentionally excluded hook events",
  ];
}

export function hookRuntimeClassificationDetailStatus(): string[] {
  return Object.entries(HOOK_RUNTIME_CLASSIFICATION)
    .filter(([, classification]) => classification.state !== "extension-connected")
    .map(([event, classification]) => `info: hook ${event} ${classification.state} - ${classification.detail}`);
}

export function hookBridgeParityStatus(): string[] {
  const mapped = Object.keys(HOOK_SCRIPT_BY_EVENT);
  const present = mapped.filter(hasHookScript).length;
  return [
    hookBridgeStatus(),
    ...hookRuntimeClassificationStatus(),
    ...hookRuntimeClassificationDetailStatus(),
    `info: ${NON_BLOCKING_HOOK_BRIDGE_POLICY}`,
    present === mapped.length
      ? "ok: all mapped hook scripts exist"
      : `missing: hook scripts present ${present}/${mapped.length}`,
  ];
}
