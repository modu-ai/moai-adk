import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { PI_AGENTS_SOURCE_PATH, PI_HOOKS_SOURCE_PATH, PI_RULES_SOURCE_PATH, PI_SKILLS_SOURCE_PATH, SOURCE_MAP_PATH } from "./constants.ts";
import { getAgentConversionStatus } from "./agent-converter.ts";
import { analyzePackageConflicts, formatFindings } from "./package-conflicts.ts";
import { formatTeamSchemaReport } from "./team-schema.ts";
import { getSkillIndexStatus } from "./trigger-indexer.ts";
import {
  CONNECTED_PI_HOOK_EVENTS,
  HOOK_RUNTIME_CLASSIFICATION,
  HOOK_SCRIPT_BY_EVENT,
  hookBridgeParityStatus,
  hookBridgeStatus,
} from "./hook-bridge.ts";
import { teamRuntimeStatus } from "./team-runtime.ts";
import { teamHookAdapterStatus } from "./team-hook-adapter.ts";
import { loadMoaiCompatConfig, outputStyleStatus, rulesStatus } from "./config.ts";
import { runtimeManifestStatus } from "./runtime-config.ts";

function exists(path: string): boolean {
  return existsSync(resolve(process.cwd(), path));
}

function readJson(path: string): unknown {
  try {
    return JSON.parse(readFileSync(resolve(process.cwd(), path), "utf8"));
  } catch {
    return null;
  }
}

function status(label: string, ok: boolean, detail = ""): string {
  return `${ok ? "ok" : "missing"}: ${label}${detail ? ` ${detail}` : ""}`;
}

function configuredPackages(): string[] {
  const settings = readJson(".pi/settings.json") as { packages?: string[] } | null;
  return settings?.packages ?? [];
}

const CLAUDE_HOOK_EVENT_TO_BRIDGE_EVENT: Record<string, string> = {
  SessionStart: "session-start",
  PreCompact: "compact",
  SessionEnd: "session-end",
  PreToolUse: "pre-tool",
  PostToolUse: "post-tool",
  Stop: "stop",
  SubagentStop: "subagent-stop",
  PostToolUseFailure: "post-tool-failure",
  Notification: "notification",
  SubagentStart: "subagent-start",
  UserPromptSubmit: "user-prompt-submit",
  PermissionRequest: "permission-request",
  TeammateIdle: "teammate-idle",
  TaskCompleted: "task-completed",
};

function claudeSettingsHookParityReport(): string[] {
  const settings = readJson(".claude/settings.json") as { hooks?: Record<string, unknown> } | null;
  const claudeEvents = Object.keys(settings?.hooks ?? {}).sort();
  const connected = new Set<string>(CONNECTED_PI_HOOK_EVENTS);
  const bridgeEvents = new Set<string>(Object.keys(HOOK_SCRIPT_BY_EVENT));
  const connectedClaudeEvents = claudeEvents.filter((event) => connected.has(CLAUDE_HOOK_EVENT_TO_BRIDGE_EVENT[event] ?? ""));
  const unconnectedClaudeEvents = claudeEvents.filter((event) => !connected.has(CLAUDE_HOOK_EVENT_TO_BRIDGE_EVENT[event] ?? ""));
  const classifiedUnconnected = unconnectedClaudeEvents.map((event) => {
    const bridgeEvent = CLAUDE_HOOK_EVENT_TO_BRIDGE_EVENT[event] ?? "";
    const classification = HOOK_RUNTIME_CLASSIFICATION[bridgeEvent];
    return classification ? `${event}=${classification.state}` : `${event}=unmapped`;
  });
  const extraBridgeEvents = [...bridgeEvents]
    .filter((event) => !Object.values(CLAUDE_HOOK_EVENT_TO_BRIDGE_EVENT).includes(event))
    .sort();

  return [
    claudeEvents.length
      ? `partial: claude settings hook events extension-connected ${connectedClaudeEvents.length}/${claudeEvents.length}`
      : "missing: claude settings hook events unavailable",
    classifiedUnconnected.length
      ? `partial: claude settings hook events not extension-connected ${classifiedUnconnected.join(", ")}`
      : "ok: all claude settings hook events are extension-connected",
    extraBridgeEvents.length
      ? `info: extra mapped hook scripts not present in claude settings ${extraBridgeEvents.join(", ")}`
      : "ok: no extra bridge-only hook scripts",
  ];
}

function hookParityReport(): string[] {
  return [
    "MoAI hook parity",
    status("claude hook source", exists(".claude/hooks/moai")),
    status("pi hook source snapshot", exists(PI_HOOKS_SOURCE_PATH)),
    status("pi-yaml-hooks root config", exists(".pi/hooks.yaml")),
    status("pi-yaml-hooks guardrail config", exists(".pi/extensions/moai-claude-compat/hooks/yaml-hooks.yaml")),
    status("pi pre-tool guardrail policy", exists(".pi/extensions/moai-claude-compat/hooks/pre-tool-policy.mjs")),
    "partial: pi-yaml-hooks runtime trust/load is not proven by file presence; verify session hook banner or pi-yaml-hooks diagnostics",
    ...hookBridgeParityStatus(),
    ...claudeSettingsHookParityReport(),
  ];
}

export function buildDoctorReport(): string[] {
  const packages = configuredPackages();
  let config: ReturnType<typeof loadMoaiCompatConfig> | null = null;
  let configError = "";
  try {
    config = loadMoaiCompatConfig();
  } catch (error) {
    configError = error instanceof Error ? error.message : String(error);
  }
  const packageFindings = formatFindings(analyzePackageConflicts(packages));

  return [
    "MoAI pi doctor",
    status("source-map", exists(SOURCE_MAP_PATH)),
    status("settings", exists(".pi/settings.json")),
    status("extension", exists(".pi/extensions/moai-claude-compat/index.ts")),
    status("pi-local skills source", exists(PI_SKILLS_SOURCE_PATH)),
    status("pi-local agents source", exists(PI_AGENTS_SOURCE_PATH)),
    status("pi-local rules source", exists(PI_RULES_SOURCE_PATH)),
    status("pi-local hooks source", exists(PI_HOOKS_SOURCE_PATH)),
    "ok: runtime prompts/code use pi-local snapshots",
    ...runtimeManifestStatus(),
    "ok: permissionMode excluded-by-design; metadata only",
    getSkillIndexStatus(),
    ...getAgentConversionStatus(),
    ...(config ? outputStyleStatus(config) : [`missing: output style status unavailable (${configError})`]),
    ...(config ? rulesStatus(config) : [`missing: rules status unavailable (${configError})`]),
    hookBridgeStatus(),
    ...hookParityReport(),
    teamRuntimeStatus(),
    ...teamHookAdapterStatus(),
    ...packageFindings,
  ];
}

export function buildAuditReport(): string[] {
  return [
    "MoAI pi audit",
    ...buildDoctorReport().slice(1),
    ...formatTeamSchemaReport(),
    ...hookParityReport(),
    teamRuntimeStatus(),
    ...teamHookAdapterStatus(),
    "pending: invoke actual teams tool after @tmustier/pi-agent-teams package activation",
    "partial: pi-yaml-hooks is the blocking guardrail; extension hook bridge is non-blocking compatibility telemetry",
    "pending: Codex/GPT quota footer validation with openai-codex OAuth login",
  ];
}
