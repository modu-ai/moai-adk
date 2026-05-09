#!/usr/bin/env node
import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const EVENT_MAP = {
  idle: {
    moaiEvent: "teammate-idle",
    hookEventName: "TeammateIdle",
    script: "handle-teammate-idle.sh",
  },
  task_completed: {
    moaiEvent: "task-completed",
    hookEventName: "TaskCompleted",
    script: "handle-task-completed.sh",
  },
};

export async function runAdapter(expectedEvent) {
  if (process.env.MOAI_PI_TEAMS_HOOK_BRIDGE !== "1") {
    console.log("MoAI Pi Agent Teams hook bridge disabled; set MOAI_PI_TEAMS_HOOK_BRIDGE=1 to enable.");
    return 0;
  }

  const context = parseTeamsContext();
  const actualEvent = context.event || process.env.PI_TEAMS_HOOK_EVENT || expectedEvent;
  if (actualEvent !== expectedEvent) {
    console.error(`MoAI Pi Agent Teams hook bridge event mismatch: expected ${expectedEvent}, got ${actualEvent}`);
    return 1;
  }

  const mapping = EVENT_MAP[expectedEvent];
  if (!mapping) {
    console.error(`Unsupported Agent Teams hook event for MoAI bridge: ${expectedEvent}`);
    return 1;
  }

  const projectRoot = findProjectRoot();
  const hookScript = path.join(projectRoot, ".pi", "generated", "source", "hooks", "moai", mapping.script);
  if (!existsSync(hookScript)) {
    console.error(`MoAI hook wrapper not found: ${hookScript}`);
    return 1;
  }

  const payload = buildMoaiPayload(mapping.hookEventName, mapping.moaiEvent, context, projectRoot);
  const result = await runMoaiWrapper(hookScript, payload, projectRoot);
  if (result.stdout.trim()) process.stdout.write(result.stdout);
  if (result.stderr.trim()) process.stderr.write(result.stderr);
  return result.exitCode ?? 1;
}

function parseTeamsContext() {
  const raw = process.env.PI_TEAMS_HOOK_CONTEXT_JSON;
  if (!raw) return { version: Number(process.env.PI_TEAMS_HOOK_CONTEXT_VERSION || 0), event: process.env.PI_TEAMS_HOOK_EVENT };
  try {
    return JSON.parse(raw);
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    throw new Error(`Invalid PI_TEAMS_HOOK_CONTEXT_JSON: ${msg}`);
  }
}

function buildMoaiPayload(hookEventName, moaiEvent, context, projectRoot) {
  const team = isRecord(context.team) ? context.team : {};
  const task = isRecord(context.task) ? context.task : null;
  const teammateName = typeof context.member === "string" ? context.member : process.env.PI_TEAMS_MEMBER || "";
  const teamName = typeof team.id === "string" ? team.id : process.env.PI_TEAMS_TEAM_ID || "";

  return {
    session_id: teamName ? `pi-team-${teamName}` : "pi-agent-teams",
    cwd: projectRoot,
    project_dir: projectRoot,
    hook_event_name: hookEventName,
    team_name: teamName,
    teammate_name: teammateName,
    agent_id: teammateName,
    task_id: stringValue(task?.id ?? process.env.PI_TEAMS_TASK_ID),
    task_subject: stringValue(task?.subject ?? process.env.PI_TEAMS_TASK_SUBJECT),
    task_description: stringValue(task?.description),
    pi_compat: {
      source: "@tmustier/pi-agent-teams",
      bridge: "moai-claude-compat",
      moai_event: moaiEvent,
      teams_hook_event: context.event || process.env.PI_TEAMS_HOOK_EVENT || "",
      teams_hook_context_version: context.version ?? process.env.PI_TEAMS_HOOK_CONTEXT_VERSION ?? null,
    },
    pi_teams_hook_context: context,
  };
}

function stringValue(value) {
  return typeof value === "string" ? value : value == null ? "" : String(value);
}

function isRecord(value) {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function findProjectRoot() {
  const candidates = [process.env.CLAUDE_PROJECT_DIR, process.cwd(), path.dirname(fileURLToPath(import.meta.url))]
    .filter((value) => typeof value === "string" && value.trim())
    .map((value) => path.resolve(value));

  for (const start of candidates) {
    const found = searchUpward(start);
    if (found) return found;
  }

  return process.cwd();
}

function searchUpward(start) {
  let current = start;
  while (true) {
    if (existsSync(path.join(current, ".pi", "generated", "source", "hooks", "moai"))) return current;
    const parent = path.dirname(current);
    if (parent === current) return null;
    current = parent;
  }
}

function runMoaiWrapper(scriptPath, payload, projectRoot) {
  return new Promise((resolve) => {
    const child = spawn("bash", [scriptPath], {
      cwd: projectRoot,
      env: { ...process.env, CLAUDE_PROJECT_DIR: projectRoot },
      stdio: ["pipe", "pipe", "pipe"],
    });
    let stdout = "";
    let stderr = "";
    child.stdout.on("data", (chunk) => (stdout += String(chunk)));
    child.stderr.on("data", (chunk) => (stderr += String(chunk)));
    child.on("close", (exitCode) => resolve({ exitCode, stdout, stderr }));
    child.on("error", (err) => resolve({ exitCode: null, stdout, stderr: `${stderr}${err.message}\n` }));
    child.stdin.end(`${JSON.stringify(payload)}\n`);
  });
}
