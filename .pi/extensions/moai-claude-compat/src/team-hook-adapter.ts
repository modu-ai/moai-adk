import { existsSync } from "node:fs";
import { resolve } from "node:path";

export const TEAM_HOOK_ADAPTER_DIR = ".pi/extensions/moai-claude-compat/hooks/teams";

export const TEAM_HOOK_ADAPTER_SCRIPTS: Record<string, string> = {
  idle: `${TEAM_HOOK_ADAPTER_DIR}/on_idle.mjs`,
  task_completed: `${TEAM_HOOK_ADAPTER_DIR}/on_task_completed.mjs`,
};

export const TEAM_HOOK_ADAPTER_FEATURE_FLAG = "MOAI_PI_TEAMS_HOOK_BRIDGE";

export const TEAM_HOOK_ADAPTER_ACTIVATION =
  "set PI_TEAMS_HOOKS_ENABLED=1, MOAI_PI_TEAMS_HOOK_BRIDGE=1, and PI_TEAMS_HOOKS_DIR to the absolute adapter directory";

function isEnabled(value: string | undefined): boolean {
  return value === "1" || value?.toLowerCase() === "true";
}

export function teamHookAdapterDirAbs(): string {
  return resolve(process.cwd(), TEAM_HOOK_ADAPTER_DIR);
}

export function teamHookAdapterScriptsPresent(): { present: number; total: number; missing: string[] } {
  const entries = Object.values(TEAM_HOOK_ADAPTER_SCRIPTS);
  const missing = entries.filter((script) => !existsSync(resolve(process.cwd(), script)));
  return { present: entries.length - missing.length, total: entries.length, missing };
}

export function teamHookAdapterStatus(): string[] {
  const scripts = teamHookAdapterScriptsPresent();
  const adapterDir = teamHookAdapterDirAbs();
  const configuredDir = process.env.PI_TEAMS_HOOKS_DIR ? resolve(process.cwd(), process.env.PI_TEAMS_HOOKS_DIR) : "";
  const hooksEnabled = isEnabled(process.env.PI_TEAMS_HOOKS_ENABLED);
  const bridgeEnabled = isEnabled(process.env[TEAM_HOOK_ADAPTER_FEATURE_FLAG]);
  const dirSelected = configuredDir === adapterDir;

  return [
    scripts.present === scripts.total
      ? `ok: Agent Teams MoAI hook adapter scripts present ${scripts.present}/${scripts.total}`
      : `missing: Agent Teams MoAI hook adapter scripts present ${scripts.present}/${scripts.total} (${scripts.missing.join(", ")})`,
    hooksEnabled && bridgeEnabled && dirSelected
      ? "ok: Agent Teams MoAI hook bridge active for idle/task_completed"
      : `partial: Agent Teams MoAI hook bridge available but inactive; ${TEAM_HOOK_ADAPTER_ACTIVATION}`,
    `info: Agent Teams MoAI hook adapter dir ${adapterDir}`,
  ];
}
