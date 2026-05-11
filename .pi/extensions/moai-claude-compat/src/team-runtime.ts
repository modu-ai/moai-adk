import { existsSync, readFileSync, readdirSync } from "node:fs";
import { join, resolve } from "node:path";
import type { ClaudeTeamApi } from "./team-schema.ts";

export interface NormalizedTeamCall {
  api: ClaudeTeamApi;
  params: Record<string, unknown>;
}

export interface TmustierTeamsToolCall {
  tool: "teams";
  params: Record<string, unknown>;
}

export function toTmustierTeamsToolCall(call: NormalizedTeamCall): TmustierTeamsToolCall {
  const p = call.params;
  switch (call.api) {
    case "TeamCreate":
      return {
        tool: "teams",
        params: {
          action: "delegate",
          teammates: p.teammates,
          tasks: p.tasks,
          contextMode: p.contextMode ?? "branch",
          workspaceMode: p.workspaceMode ?? "worktree",
          model: p.model,
          thinking: p.thinking,
          planRequired: p.planRequired,
        },
      };
    case "SendMessage":
      return {
        tool: "teams",
        params: p.name
          ? { action: p.urgent ? "message_steer" : "message_dm", name: p.name, message: p.message, urgent: p.urgent }
          : { action: "message_broadcast", message: p.message, urgent: p.urgent },
      };
    case "TaskCreate":
      return { tool: "teams", params: { action: "delegate", tasks: p.tasks ?? [{ text: p.text, assignee: p.assignee }] } };
    case "TaskUpdate":
      if (p.status) return { tool: "teams", params: { action: "task_set_status", taskId: p.taskId, status: p.status } };
      if (p.assignee) return { tool: "teams", params: { action: "task_assign", taskId: p.taskId, assignee: p.assignee } };
      return { tool: "teams", params: { action: "task_unassign", taskId: p.taskId } };
    case "TaskList":
      return { tool: "teams", params: { action: "member_status" } };
    case "TaskGet":
      return { tool: "teams", params: { action: "task_dep_ls", taskId: p.taskId } };
    case "TeamDelete":
      return { tool: "teams", params: { action: "team_done", all: p.force ?? true } };
  }
}

export const TEAM_MOAI_PROFILE_MAPPINGS = {
  plan: {
    researcher: "scout",
    analyst: "manager-spec",
    architect: "manager-strategy",
  },
  run: {
    "backend-dev": "expert-backend",
    "frontend-dev": "expert-frontend",
    tester: "expert-testing",
    quality: "manager-quality",
    reviewer: "manager-quality",
  },
  review: {
    "security-reviewer": "expert-security",
    "perf-reviewer": "expert-performance",
    "quality-reviewer": "manager-quality",
    "ux-reviewer": "expert-frontend",
  },
} as const;

const RUNTIME_FACING_SCAN_ROOTS = [
  ".pi/generated/source/skills",
  ".pi/generated/source/rules",
  ".pi/prompts",
] as const;

const MISSING_PSEUDO_AGENTS = ["team-reader", "team-validator", "team-coder", "team-tester", "team-designer"] as const;
const COMPAT_BUILTIN_PROFILES = new Set(["scout"]);

function collectMarkdownFiles(path: string, cwd: string): string[] {
  const abs = resolve(cwd, path);
  if (!existsSync(abs)) return [];
  const entries = readdirSync(abs, { withFileTypes: true });
  return entries.flatMap((entry) => {
    const child = join(path, entry.name);
    if (entry.isDirectory()) return collectMarkdownFiles(child, cwd);
    return entry.isFile() && entry.name.endsWith(".md") ? [child] : [];
  });
}

function readIfExists(path: string, cwd: string): string {
  const abs = resolve(cwd, path);
  return existsSync(abs) ? readFileSync(abs, "utf8") : "";
}

function runtimeFacingDocsText(cwd: string): string {
  return RUNTIME_FACING_SCAN_ROOTS
    .flatMap((root) => collectMarkdownFiles(root, cwd))
    .map((path) => readIfExists(path, cwd))
    .join("\n");
}

function moaiProfileExists(profile: string, cwd: string): boolean {
  if (COMPAT_BUILTIN_PROFILES.has(profile)) return true;
  return existsSync(resolve(cwd, ".pi/agents/moai", `${profile}.md`))
    || existsSync(resolve(cwd, ".pi/generated/source/agents/moai", `${profile}.md`));
}

export function teamMoaiProfileMappingStatus(cwd = process.cwd()): string[] {
  const docsText = runtimeFacingDocsText(cwd);
  const pseudoAgents = MISSING_PSEUDO_AGENTS.filter((name) => docsText.includes(name));
  const mappings = Object.entries(TEAM_MOAI_PROFILE_MAPPINGS)
    .flatMap(([phase, roles]) => Object.entries(roles).map(([role, profile]) => `${phase}.${role}->${profile}`));
  const missingProfiles = mappings
    .map((mapping) => mapping.split("->", 2) as [string, string])
    .filter(([, profile]) => !moaiProfileExists(profile, cwd))
    .map(([role, profile]) => `${role}->${profile}`);

  return [
    pseudoAgents.length
      ? `missing: runtime-facing docs reference removed pseudo-agents ${pseudoAgents.join(", ")}`
      : `ok: runtime-facing docs avoid removed team pseudo-agents (${MISSING_PSEUDO_AGENTS.length} checked)`,
    `ok: team MoAI profile mappings ${mappings.join(", ")}`,
    missingProfiles.length
      ? `missing: team MoAI profile files ${missingProfiles.join(", ")}`
      : "ok: team MoAI profiles resolve to project agents or approved Pi compat builtins",
    "ok: team runtime keeps general-purpose teammates for team capability; MoAI identity is injected via profile adoption prompts",
  ];
}

export function teamRuntimeStatus(): string {
  return "ok: @tmustier/pi-agent-teams normalized Team API mapping implemented for teams tool; live teammate invocation pending";
}
