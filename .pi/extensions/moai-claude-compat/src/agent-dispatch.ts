import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { loadRuntimeManifest, resolveCompatSource, type RuntimeManifest } from "./runtime-config.ts";

export type WorkflowDispatchMode = "skill" | "subagent" | "direct";

export interface WorkflowDispatchEntry {
  source?: string;
  modeSource?: string;
  primaryAgents?: string[];
  dispatchMode?: WorkflowDispatchMode;
  entryAgent?: string;
  agentAliases?: Record<string, string>;
}

export interface SubagentWorkflowDispatch {
  kind: "subagent";
  subcommand: string;
  agent: string;
  task: string;
  prompt: string;
  reason: string;
}

export interface FallbackWorkflowDispatch {
  kind: "fallback";
  subcommand: string;
  reason: string;
  dispatchMode?: WorkflowDispatchMode;
  selectedAgent?: string;
}

export type WorkflowDispatch = SubagentWorkflowDispatch | FallbackWorkflowDispatch;

const PROJECT_QUALITY_PATH = ".moai/config/sections/quality.yaml";
const GENERATED_QUALITY_PATH = ".pi/generated/source/moai-config/sections/quality.yaml";
const PROJECT_AGENT_DIR = ".pi/agents/moai";
const SOURCE_AGENT_DIR = ".pi/generated/source/agents/moai";

function readIfExists(path: string, cwd: string): string {
  const abs = resolve(cwd, path);
  return existsSync(abs) ? readFileSync(abs, "utf8") : "";
}

export function agentExists(agent: string, cwd = process.cwd()): boolean {
  if (!agent.trim()) return false;
  return existsSync(resolve(cwd, PROJECT_AGENT_DIR, `${agent}.md`))
    || existsSync(resolve(cwd, SOURCE_AGENT_DIR, `${agent}.md`));
}

export function resolveDevelopmentMode(cwd = process.cwd()): "ddd" | "tdd" {
  const text = readIfExists(PROJECT_QUALITY_PATH, cwd) || readIfExists(GENERATED_QUALITY_PATH, cwd);
  const match = text.match(/^\s*development_mode\s*:\s*([A-Za-z_-]+)\s*$/m);
  return match?.[1]?.toLowerCase() === "ddd" ? "ddd" : "tdd";
}

function firstExistingAgent(candidates: string[], cwd: string): string | undefined {
  return candidates.find((agent) => agentExists(agent, cwd));
}

function workflowSourcePath(subcommand: string, entry: WorkflowDispatchEntry | undefined, runtime: RuntimeManifest): string {
  if (!entry?.source) return "";
  return resolveCompatSource(runtime.files.workflowMap, entry.source, runtime.cwd) || `.pi/generated/source/skills/moai/workflows/${subcommand}.md`;
}

function selectEntryAgent(subcommand: string, entry: WorkflowDispatchEntry | undefined, cwd: string): string | undefined {
  if (subcommand === "run") return resolveDevelopmentMode(cwd) === "ddd" ? "manager-ddd" : "manager-tdd";
  return entry?.entryAgent ?? firstExistingAgent(entry?.primaryAgents ?? [], cwd) ?? entry?.primaryAgents?.[0];
}

export function explicitTeamModeRequested(args: string): boolean {
  return /(?:^|\s)--team(?:\s|$)/.test(args)
    || /(?:^|\s)--mode(?:=|\s+)team(?:\s|$)/.test(args);
}

export function reviewSpecialistPhaseRequested(args: string): boolean {
  return /(?:^|\s)--(?:security|design|critique)(?:[=\s]|$)/.test(args);
}

function buildSubagentTask(subcommand: string, args: string, agent: string, sourcePath: string): string {
  const argLine = args.trim() ? `Arguments: ${args.trim()}` : "Arguments: none";
  const sourceLine = sourcePath ? `Workflow source: ${sourcePath}` : "Workflow source: configured MoAI workflow manifest";
  return [
    `Execute the MoAI '${subcommand}' workflow as the '${agent}' project agent.`,
    argLine,
    sourceLine,
    "Before acting, read the workflow source above and follow its phase sequence, guards, and verification requirements where applicable.",
    "Use the project-local Pi compatibility runtime and MoAI rules from .pi/generated/source/**.",
    "Stay within this selected specialist role; do not silently replace parent MoAI orchestration.",
    "If the workflow requires parent-only AskUserQuestion, additional subagents, team orchestration, or cross-phase coordination outside this agent's role, stop and return an ORCHESTRATION_BLOCKER report for the parent session.",
    "Do not ask the user directly. If user input is required, return the blocker report with the exact decision needed.",
    "Return a focused result summary, changed files if any, and verification evidence.",
  ].join("\n");
}

export function buildSubagentDelegationPrompt(dispatch: SubagentWorkflowDispatch): string {
  const params = {
    agent: dispatch.agent,
    task: dispatch.task,
    context: "fork",
    agentScope: "project",
  };

  return [
    "MoAI Pi staged agent dispatch is active for a safe single-entry workflow.",
    "",
    "First action: call the Pi `subagent` tool with exactly these parameters:",
    JSON.stringify(params, null, 2),
    "",
    "Do not execute this workflow directly in the parent session before the subagent returns.",
    "The subagent task requires it to read the workflow source first, follow the workflow phase sequence where applicable, and return ORCHESTRATION_BLOCKER if parent-only AskUserQuestion, extra subagents, or team orchestration is needed.",
    "After the subagent returns, summarize the result in the user's conversation language, Korean when configured.",
    "If the subagent reports a blocker requiring user input or broader orchestration, bridge that decision from the parent session instead of letting the subagent ask the user directly.",
    "",
    `Dispatch reason: ${dispatch.reason}`,
  ].join("\n");
}

export function resolveWorkflowAgentDispatch(
  subcommand: string,
  args = "",
  runtime = loadRuntimeManifest(),
): WorkflowDispatch {
  const workflow = runtime.workflowMap.config.workflows?.[subcommand] as WorkflowDispatchEntry | undefined;
  if (!workflow) return { kind: "fallback", subcommand, reason: "workflow not found in runtime manifest" };

  if (explicitTeamModeRequested(args)) {
    return { kind: "fallback", subcommand, reason: "explicit team mode requested; use existing Skill prompt route for parent team orchestration", dispatchMode: "skill" };
  }

  if (subcommand === "review" && reviewSpecialistPhaseRequested(args)) {
    return { kind: "fallback", subcommand, reason: "review specialist flag requested; use existing Skill prompt route for conditional expert-security/expert-frontend phases", dispatchMode: "skill" };
  }

  const mode = workflow.dispatchMode ?? "skill";
  if (mode !== "subagent") {
    return { kind: "fallback", subcommand, reason: `dispatchMode=${mode}; use existing Skill prompt route`, dispatchMode: mode };
  }

  const selectedAgent = selectEntryAgent(subcommand, workflow, runtime.cwd);
  if (!selectedAgent) {
    return { kind: "fallback", subcommand, reason: "no entry agent configured", dispatchMode: mode };
  }

  const mappedAgent = workflow.agentAliases?.[selectedAgent] ?? selectedAgent;
  if (!agentExists(mappedAgent, runtime.cwd)) {
    return {
      kind: "fallback",
      subcommand,
      reason: `entry agent '${mappedAgent}' is missing from ${PROJECT_AGENT_DIR} and ${SOURCE_AGENT_DIR}`,
      dispatchMode: mode,
      selectedAgent: mappedAgent,
    };
  }

  const sourcePath = workflowSourcePath(subcommand, workflow, runtime);
  const task = buildSubagentTask(subcommand, args, mappedAgent, sourcePath);
  const dispatch: SubagentWorkflowDispatch = {
    kind: "subagent",
    subcommand,
    agent: mappedAgent,
    task,
    prompt: "",
    reason: subcommand === "run" ? `run workflow selected ${mappedAgent} from development_mode` : `workflow dispatch selected ${mappedAgent}`,
  };
  return { ...dispatch, prompt: buildSubagentDelegationPrompt(dispatch) };
}

export function workflowDispatchStatus(runtime = loadRuntimeManifest()): string[] {
  const entries = Object.entries(runtime.workflowMap.config.workflows ?? {}) as Array<[string, WorkflowDispatchEntry]>;
  const checked = entries.map(([subcommand]) => resolveWorkflowAgentDispatch(subcommand, "", runtime));
  const subagents = checked.filter((item) => item.kind === "subagent") as SubagentWorkflowDispatch[];
  const direct = checked.filter((item) => item.kind === "fallback" && item.dispatchMode === "direct");
  const skill = checked.filter((item) => item.kind === "fallback" && item.dispatchMode !== "direct");
  const missing = checked.filter((item) => item.kind === "fallback" && item.selectedAgent);

  return [
    `ok: workflow agent dispatch subagent coverage ${subagents.length}/${entries.length}${subagents.length ? ` (${subagents.map((item) => `${item.subcommand}->${item.agent}`).join(", ")})` : ""}`,
    direct.length ? `ok: workflow direct/skill fallback ${direct.map((item) => item.subcommand).join(", ")}` : "ok: no direct-only workflows configured",
    skill.length ? `info: workflow skill fallback ${skill.map((item) => `${item.subcommand}(${item.reason})`).join(", ")}` : "ok: no skill fallback workflows configured",
    missing.length ? `missing: workflow dispatch entry agents ${missing.map((item) => `${item.subcommand}->${item.selectedAgent}`).join(", ")}` : "ok: workflow dispatch entry agents exist",
  ];
}
