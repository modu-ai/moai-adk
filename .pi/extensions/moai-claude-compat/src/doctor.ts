import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { PI_AGENTS_SOURCE_PATH, PI_HOOKS_SOURCE_PATH, PI_RULES_SOURCE_PATH, PI_SKILLS_SOURCE_PATH, SOURCE_MAP_PATH } from "./constants.ts";
import { getAgentConversionStatus } from "./agent-converter.ts";
import { analyzePackageConflicts, formatFindings, normalizePackageSpecs, type PackageSpec } from "./package-conflicts.ts";
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

interface MoaiSettings {
  packages?: PackageSpec[];
  moaiCompat?: {
    defaultPackages?: string[];
    securityAdvisories?: Array<{
      package?: string;
      severity?: string;
      via?: string;
      mitigation?: string;
    }>;
  };
}

function readMoaiSettings(): MoaiSettings | null {
  return readJson(".pi/settings.json") as MoaiSettings | null;
}

function configuredPackages(): PackageSpec[] {
  return readMoaiSettings()?.packages ?? [];
}

function packageSetupReport(settings: MoaiSettings | null): string[] {
  const configured = settings?.packages ?? [];
  const defaults = settings?.moaiCompat?.defaultPackages ?? [];
  const configuredNames = new Set(normalizePackageSpecs(configured));
  const runtimeOnly = new Set(["moai-claude-compat", "pi-notify-glass.ts"]);
  const missing = normalizePackageSpecs(defaults)
    .filter((name) => !runtimeOnly.has(name))
    .filter((name) => !configuredNames.has(name));

  return [
    missing.length
      ? `missing: MoAI default pi packages not active ${missing.join(", ")}`
      : `ok: MoAI default pi package set active (${configured.length} package specs)`,
    defaults.length
      ? `ok: MoAI default package manifest present (${defaults.length} entries; runtime-only entries excluded from package activation checks)`
      : "missing: MoAI default package manifest absent",
  ];
}

function securityAdvisoryReport(settings: MoaiSettings | null): string[] {
  const advisories = settings?.moaiCompat?.securityAdvisories ?? [];
  if (advisories.length === 0) return ["ok: no active MoAI package security advisories recorded"];
  return advisories.map((advisory) => {
    const pkg = advisory.package ?? "unknown package";
    const severity = advisory.severity ?? "unknown severity";
    const via = advisory.via ? ` via ${advisory.via}` : "";
    const mitigation = advisory.mitigation ? `; mitigation: ${advisory.mitigation}` : "";
    return `warn(non-blocking): security advisory recorded for ${pkg} (${severity})${via}${mitigation}`;
  });
}

const CLAUDE_HOOK_EVENT_TO_BRIDGE_EVENT: Record<string, string> = {
  SessionStart: "session-start",
  PreCompact: "compact",
  PostCompact: "post-compact",
  SessionEnd: "session-end",
  PreToolUse: "pre-tool",
  PostToolUse: "post-tool",
  Stop: "stop",
  StopFailure: "stop-failure",
  SubagentStop: "subagent-stop",
  PostToolUseFailure: "post-tool-failure",
  Notification: "notification",
  SubagentStart: "subagent-start",
  UserPromptSubmit: "user-prompt-submit",
  PermissionRequest: "permission-request",
  TeammateIdle: "teammate-idle",
  TaskCreated: "task-created",
  TaskCompleted: "task-completed",
  WorktreeCreate: "worktree-create",
  WorktreeRemove: "worktree-remove",
  ConfigChange: "config-change",
  CwdChanged: "cwd-changed",
  Elicitation: "elicitation",
  ElicitationResult: "elicitation-result",
  FileChanged: "file-changed",
  InstructionsLoaded: "instructions-loaded",
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
  const settings = readMoaiSettings();
  const packages = settings?.packages ?? [];
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
    ...packageSetupReport(settings),
    ...packageFindings,
    ...securityAdvisoryReport(settings),
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
