import { existsSync } from "node:fs";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import type { ExtensionContext } from "@earendil-works/pi-coding-agent";
import { getCodexQuotaFooterText } from "./codex-quota.ts";
import { STATUS_ID, WIDGET_ID } from "./constants.ts";
import type { MoaiCompatConfig } from "./config.ts";

export interface MoaiStatusState {
  phase: "idle" | "plan" | "run" | "sync" | "review" | "gate";
  specId?: string;
  teamMode?: string;
  worktree?: string;
  quotaProvider?: string;
  statusText?: string;
  statusLines?: string[];
}

const state: MoaiStatusState = { phase: "idle" };

export function updateMoaiStatus(ctx: ExtensionContext, config: MoaiCompatConfig, patch: Partial<MoaiStatusState> = {}) {
  Object.assign(state, patch);
  state.statusLines = buildClaudeLikeStatusLines(ctx);
  state.statusText = state.statusLines[0] ?? buildFallbackStatus(config);

  // Keep the status provider populated for other footer implementations, but render
  // MoAI's own native footer when TUI is available.
  ctx.ui.setStatus(STATUS_ID, state.statusText);
  if (ctx.hasUI) setMoaiNativeFooter(ctx, config);
}

function setMoaiNativeFooter(ctx: ExtensionContext, config: MoaiCompatConfig): void {
  ctx.ui.setFooter((_tui, theme, footerData) => ({
    invalidate() {},
    render(width: number) {
      const lines = buildMoaiFooterLines(ctx, config, width, footerData);
      return lines.map((line, index) => theme.fg(index === 0 ? "accent" : "dim", line));
    },
  }));
}

function buildMoaiFooterLines(
  ctx: ExtensionContext,
  config: MoaiCompatConfig,
  width: number,
  footerData: {
    getGitBranch(): string | null;
    getExtensionStatuses(): ReadonlyMap<string, string>;
    getAvailableProviderCount(): number;
  },
): string[] {
  const safeWidth = safeFooterWidth(width);
  const claudeLines = state.statusLines?.length ? state.statusLines : [];
  const quota = getCodexQuotaFooterText(safeWidth);
  const context = buildContextWindowText(ctx);
  if (claudeLines.length > 0) {
    return composeMoaiNativeFooterLines(claudeLines, quota, safeWidth, context);
  }

  const status = state.statusText ?? buildFallbackStatus(config);
  const right = buildModelText(ctx, footerData);
  const bars = [context, quota].filter(Boolean).join(" │ ");
  const lines = [
    truncateVisual(status, safeWidth),
    fitLine(bars, right, safeWidth),
    truncateVisual(buildLocationText(ctx, footerData), safeWidth),
  ].filter(Boolean);

  const statusLine = Array.from(footerData.getExtensionStatuses().entries())
    .filter(([key]) => key !== STATUS_ID)
    .map(([, text]) => sanitizeFooterText(text))
    .filter(Boolean)
    .join(" ");
  if (statusLine) lines.push(truncateVisual(statusLine, safeWidth));

  return lines;
}

function buildClaudeLikeStatusLines(ctx: ExtensionContext): string[] {
  const cwd = ctx.cwd || process.cwd();
  const payload = JSON.stringify({
    cwd,
    workspace: {
      current_dir: cwd,
      project_dir: cwd,
    },
    model: normalizeModel(ctx.model),
    output_style: { name: "MoAI" },
  });

  const script = join(cwd, ".moai", "status_line.sh");
  const candidates = existsSync(script)
    ? [{ command: script, args: [] as string[] }, { command: "moai", args: ["statusline"] }]
    : [{ command: "moai", args: ["statusline"] }];

  for (const candidate of candidates) {
    const result = spawnSync(candidate.command, candidate.args, {
      cwd,
      input: payload,
      encoding: "utf8",
      timeout: 1_500,
      env: {
        ...process.env,
        CLAUDE_PROJECT_DIR: cwd,
        DEBUG_STATUSLINE: process.env.DEBUG_STATUSLINE ?? "0",
        HOME: process.env.MOAI_PI_STATUSLINE_HOME ?? "/tmp/moai-pi-statusline-no-home",
        MOAI_NO_COLOR: "1",
        NO_COLOR: "1",
      },
    });
    if (result.error || result.status !== 0) continue;

    const lines = normalizeClaudeStatusLines(result.stdout);
    if (lines.length > 0) return lines;
  }

  return [];
}

function normalizeModel(model: unknown): { id: string; name: string; display_name: string } {
  const record = typeof model === "object" && model !== null ? model as Record<string, unknown> : {};
  const id = String(record.id ?? record.name ?? "pi");
  const displayName = String(record.displayName ?? record.display_name ?? record.name ?? record.id ?? "Pi");
  return { id, name: id, display_name: displayName };
}

function normalizeClaudeStatusLines(output: string): string[] {
  return output
    .split(/\r?\n/)
    .map((line) => stripClaudeQuotaSegments(sanitizeFooterText(line)))
    .filter(Boolean);
}

function stripClaudeQuotaSegments(line: string): string {
  const parts = line
    .split("│")
    .map((part) => sanitizeFooterText(part))
    .filter(Boolean)
    .filter((part) => !isClaudeQuotaSegment(part));
  return parts.join(" │ ");
}

function isClaudeQuotaSegment(part: string): boolean {
  return /^(5H|5h|7D|7d):/.test(part);
}

function composeMoaiNativeFooterLines(claudeLines: string[], quota: string | undefined, width: number, context?: string): string[] {
  const lines = claudeLines.map((line) => truncateVisual(line, width)).filter(Boolean);
  const contextText = context ? sanitizeFooterText(context) : "";
  const quotaText = quota ? sanitizeFooterText(quota) : "";
  if (!contextText && !quotaText) return lines;

  const cwIndex = lines.findIndex((line) => /(^|\b)CW:/.test(line));
  if (cwIndex >= 0) {
    lines[cwIndex] = truncateVisual([lines[cwIndex], quotaText].filter(Boolean).join(" │ "), width);
    return lines;
  }

  const barLine = [contextText, quotaText].filter(Boolean).join(" │ ");
  const insertAt = Math.min(1, lines.length);
  return [...lines.slice(0, insertAt), truncateVisual(barLine, width), ...lines.slice(insertAt)];
}

export function normalizeClaudeStatusLinesForTest(output: string): string[] {
  return normalizeClaudeStatusLines(output);
}

export function composeMoaiNativeFooterLinesForTest(claudeLines: string[], quota: string | undefined, width: number, context?: string): string[] {
  return composeMoaiNativeFooterLines(claudeLines, quota, width, context);
}

function buildFallbackStatus(config: MoaiCompatConfig): string {
  return [
    `MoAI:${state.phase}`,
    state.specId,
    `quality:${config.qualityMode}`,
    state.teamMode ? `team:${state.teamMode}` : undefined,
    state.worktree ? `wt:${state.worktree}` : undefined,
  ].filter(Boolean).join(" ");
}

function buildModelText(
  ctx: ExtensionContext,
  footerData: { getAvailableProviderCount(): number },
): string {
  const modelName = ctx.model?.id || "no-model";
  if (footerData.getAvailableProviderCount() > 1 && ctx.model?.provider) return `(${ctx.model.provider}) ${modelName}`;
  return modelName;
}

function buildContextWindowText(ctx: ExtensionContext): string {
  const usage = ctx.getContextUsage();
  const pct = usage?.percent === null || usage?.percent === undefined
    ? 0
    : Math.max(0, Math.min(100, Math.round(usage.percent)));
  return `CW: ${pct > 70 ? "🪫" : "🔋"} ${renderMoaiBar(pct, 10)} ${pct}%`;
}

function renderMoaiBar(percent: number, width: number): string {
  const filled = Math.max(0, Math.min(width, Math.round((percent / 100) * width)));
  return `${"█".repeat(filled)}${"░".repeat(width - filled)}`;
}

function buildLocationText(
  ctx: ExtensionContext,
  footerData: { getGitBranch(): string | null },
): string {
  let pwd = ctx.cwd || process.cwd();
  const home = process.env.HOME || process.env.USERPROFILE;
  if (home && pwd.startsWith(home)) pwd = `~${pwd.slice(home.length)}`;

  const branch = footerData.getGitBranch();
  const sessionName = ctx.sessionManager.getSessionName();
  return [branch ? `${pwd} (${branch})` : pwd, sessionName].filter(Boolean).join(" • ");
}

function fitLine(left: string, right: string, width: number): string {
  const cleanLeft = sanitizeFooterText(left);
  const cleanRight = sanitizeFooterText(right);
  if (!cleanRight) return truncateVisual(cleanLeft, width);

  const leftWidth = visibleWidth(cleanLeft);
  const rightWidth = visibleWidth(cleanRight);
  if (leftWidth + 2 + rightWidth <= width) {
    return `${cleanLeft}${" ".repeat(Math.max(1, width - leftWidth - rightWidth))}${cleanRight}`;
  }

  const availableForLeft = width - rightWidth - 1;
  if (availableForLeft > 8) return `${truncateVisual(cleanLeft, availableForLeft)} ${cleanRight}`;
  return truncateVisual(cleanLeft, width);
}

function sanitizeFooterText(text: string): string {
  return text.replace(/[\r\n\t]/g, " ").replace(/ +/g, " ").trim();
}

function visibleWidth(text: string): number {
  return Array.from(stripAnsi(text)).reduce((width, char) => width + charWidth(char), 0);
}

function charWidth(char: string): number {
  const code = char.codePointAt(0) ?? 0;
  if (code === 0) return 0;
  if (code < 32 || (code >= 0x7f && code < 0xa0)) return 0;
  if (
    (code >= 0x1100 && code <= 0x115f) ||
    (code >= 0x2329 && code <= 0x232a) ||
    (code >= 0x2500 && code <= 0x259f) ||
    (code >= 0x2e80 && code <= 0xa4cf) ||
    (code >= 0xac00 && code <= 0xd7a3) ||
    (code >= 0xf900 && code <= 0xfaff) ||
    (code >= 0xfe10 && code <= 0xfe19) ||
    (code >= 0xfe30 && code <= 0xfe6f) ||
    (code >= 0xff00 && code <= 0xff60) ||
    (code >= 0xffe0 && code <= 0xffe6) ||
    (code >= 0x1f000 && code <= 0x1faff)
  ) return 2;
  return 1;
}

function safeFooterWidth(width: number): number {
  return Math.max(20, width - 8);
}

function truncateVisual(text: string, width: number): string {
  const clean = sanitizeFooterText(stripAnsi(text));
  if (width <= 0) return "";
  if (visibleWidth(clean) <= width) return clean;
  if (width <= 1) return "…".slice(0, width);

  let out = "";
  let used = 0;
  for (const char of Array.from(clean)) {
    const next = charWidth(char);
    if (used + next > width - 1) break;
    out += char;
    used += next;
  }
  return `${out}…`;
}

function stripAnsi(text: string): string {
  return text.replace(/\x1B\[[0-?]*[ -/]*[@-~]/g, "");
}

export function setMoaiWidget(ctx: ExtensionContext, lines?: string[]) {
  ctx.ui.setWidget(WIDGET_ID, lines && lines.length > 0 ? lines : undefined);
}

export function inferPhaseFromCommand(commandName: string): MoaiStatusState["phase"] {
  if (commandName.includes("plan")) return "plan";
  if (commandName.includes("run")) return "run";
  if (commandName.includes("sync")) return "sync";
  if (commandName.includes("review")) return "review";
  if (commandName.includes("gate") || commandName.includes("coverage") || commandName.includes("e2e")) return "gate";
  return "idle";
}
