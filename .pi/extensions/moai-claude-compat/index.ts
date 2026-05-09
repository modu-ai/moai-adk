import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { buildCoreInstruction, loadMoaiCompatConfig, type MoaiRulesConfig } from "./src/config.ts";
import { registerCommands } from "./src/command-router.ts";
import { EXTENSION_ID, PI_RULES_SOURCE_PATH } from "./src/constants.ts";
import { NON_BLOCKING_HOOK_BRIDGE_POLICY, runMoaiHook } from "./src/hook-bridge.ts";
import { registerCodexQuota } from "./src/codex-quota.ts";
import { notifyMoai, type MoaiNotificationContext } from "./src/notification-adapter.ts";
import { registerMoaiGlassNotifications } from "./src/pi-notify-glass.ts";
import { updateMoaiStatus } from "./src/statusline.ts";
import { buildSkillTriggerHints } from "./src/trigger-indexer.ts";

function isMoaiRelatedInput(text: string): boolean {
  return /(^|\s)\/moai\b/i.test(text)
    || /\bSPEC-[A-Za-z0-9_-]+\b/i.test(text)
    || /\bmoai\b/i.test(text)
    || text.includes("모아이");
}

function buildCompactRulesInstruction(rules: MoaiRulesConfig): string {
  return rules.loaded
    ? rules.guidance
    : [rules.guidance, `Rules status warning: ${rules.error ?? "rules snapshot incomplete"}`].join("\n");
}

interface AutoRouteConfig {
  enabled: boolean;
  mode: "all-natural-language" | "moai-keyword-only";
  excludePrefixes: string[];
}

function loadAutoRouteConfig(): AutoRouteConfig {
  const fallback: AutoRouteConfig = { enabled: false, mode: "all-natural-language", excludePrefixes: ["/", "!", "!!"] };
  try {
    const path = resolve(process.cwd(), ".pi/settings.json");
    if (!existsSync(path)) return fallback;
    const parsed = JSON.parse(readFileSync(path, "utf8")) as { moaiCompat?: { autoRoute?: Partial<AutoRouteConfig> } };
    const autoRoute = parsed.moaiCompat?.autoRoute;
    return {
      enabled: autoRoute?.enabled === true,
      mode: autoRoute?.mode === "moai-keyword-only" ? "moai-keyword-only" : "all-natural-language",
      excludePrefixes: Array.isArray(autoRoute?.excludePrefixes)
        ? autoRoute.excludePrefixes.filter((value): value is string => typeof value === "string")
        : fallback.excludePrefixes,
    };
  } catch {
    return fallback;
  }
}

function shouldAutoRouteToMoai(text: string, source: unknown): boolean {
  const trimmed = text.trim();
  if (!trimmed) return false;
  if (source === "extension") return false;
  if (/^Use\s+Skill\(["']moai["']\)/i.test(trimmed)) return false;
  const autoRoute = loadAutoRouteConfig();
  if (!autoRoute.enabled) return false;
  if (autoRoute.excludePrefixes.some((prefix) => prefix && trimmed.startsWith(prefix))) return false;
  if (autoRoute.mode === "moai-keyword-only" && !isMoaiRelatedInput(trimmed)) return false;
  return true;
}

function buildMoaiAutoRoutePrompt(text: string): string {
  return `Use Skill("moai") with arguments: ${text.trim()}`;
}

export default function moaiClaudeCompat(pi: ExtensionAPI) {
  const config = loadMoaiCompatConfig();
  const coreInstruction = buildCoreInstruction(config);
  const outputStyleInstruction = config.outputStyle.instruction;
  const basePromptInjection = [
    "MoAI pi compatibility runtime instructions:",
    coreInstruction,
    outputStyleInstruction,
  ].filter(Boolean).join("\n\n");
  let pendingPromptContext: { prompt: string; context: string } | undefined;

  registerCommands(pi, config);
  registerMoaiGlassNotifications(pi);

  async function invokeHook(eventName: string, payload: unknown, ctx?: MoaiNotificationContext) {
    // Compatibility hooks intentionally do not block Pi tool execution here.
    // Security-sensitive blocking is handled by pi-yaml-hooks tool.before.* guardrails.
    const result = await runMoaiHook(eventName, payload);
    if (!result.ok) {
      const detail = (result.stderr || result.stdout || `exit ${result.exitCode ?? "unknown"}`).trim();
      await notifyMoai(ctx, `MoAI hook '${eventName}' failed non-blocking: ${detail}`, "warning", {
        source: "hook-bridge",
        failedHookEvent: eventName,
      });
    }
    return result;
  }

  pi.on("session_start", async (event, ctx) => {
    await invokeHook("session-start", { hook_event_name: "SessionStart", event, cwd: ctx.cwd }, ctx);
    updateMoaiStatus(ctx, config, { phase: "idle" });
    await notifyMoai(ctx, "MoAI pi compatibility layer loaded", "info", {
      source: "moai-claude-compat",
      reason: event.reason,
    });
    pi.appendEntry(`${EXTENSION_ID}:loaded`, {
      conversationLanguage: config.conversationLanguage,
      qualityMode: config.qualityMode,
      permissionMode: "excluded-by-design",
      hookBridgePolicy: NON_BLOCKING_HOOK_BRIDGE_POLICY,
      outputStyle: {
        name: config.outputStyle.name,
        loaded: config.outputStyle.loaded,
        sanitized: config.outputStyle.sanitized,
        enforcement: config.outputStyle.enforcement,
      },
      rules: {
        loaded: config.rules.loaded,
        piCount: config.rules.piCount,
        expectedCount: config.rules.expectedCount,
        sourceMapRegistered: config.rules.sourceMapRegistered,
        relativePathParity: config.rules.relativePathParity,
        enforcement: config.rules.enforcement,
      },
    });
  });

  pi.on("session_shutdown", async (event, ctx) => {
    await invokeHook("session-end", { hook_event_name: "SessionEnd", event, cwd: ctx.cwd }, ctx);
  });

  pi.on("session_before_compact", async (event, ctx) => {
    await invokeHook("compact", { hook_event_name: "PreCompact", event, cwd: ctx.cwd }, ctx);
  });

  pi.on("turn_start", async (_event, ctx) => {
    updateMoaiStatus(ctx, config);
  });

  pi.on("turn_end", async (_event, ctx) => {
    updateMoaiStatus(ctx, config);
  });

  function buildPromptHints(text: string, hookStdout = ""): string {
    const lower = text.toLowerCase();
    const hints: string[] = [];
    if (lower.includes("--deepthink")) hints.push("Deepthink requested: prefer sequential-thinking MCP when available.");
    if (isMoaiRelatedInput(text)) hints.push(buildCompactRulesInstruction(config.rules));
    if (lower.includes("spec-") || lower.includes("/moai run")) hints.push(`SPEC workflow context likely required: read .moai/specs and pi-local MoAI workflow rules at ${PI_RULES_SOURCE_PATH}.`);
    if (lower.includes("permissionmode")) hints.push("Reminder: Claude permissionMode is excluded by design in pi parity.");
    hints.push(...buildSkillTriggerHints(text));
    if (hookStdout.trim()) hints.push(`MoAI user-prompt hook output: ${hookStdout.trim()}`);
    return hints.filter(Boolean).join("\n");
  }

  pi.on("input", async (event, ctx) => {
    const text = typeof event.text === "string" ? event.text : typeof event.input === "string" ? event.input : "";
    if (!text.trim()) return { action: "continue" };

    const hookResult = await invokeHook("user-prompt-submit", { hook_event_name: "UserPromptSubmit", prompt: text, event, cwd: ctx.cwd }, ctx);
    const routedText = shouldAutoRouteToMoai(text, event.source) ? buildMoaiAutoRoutePrompt(text) : text;
    pendingPromptContext = { prompt: routedText, context: buildPromptHints(text, hookResult.stdout) };
    return routedText === text ? { action: "continue" } : { action: "transform", text: routedText };
  });

  pi.on("before_agent_start", async (event) => {
    const text = typeof event.prompt === "string" ? event.prompt : "";
    const promptContext = pendingPromptContext?.prompt === text
      ? pendingPromptContext.context
      : buildPromptHints(text);
    if (pendingPromptContext?.prompt === text) pendingPromptContext = undefined;

    return {
      systemPrompt: [event.systemPrompt, basePromptInjection, promptContext]
        .filter(Boolean)
        .join("\n\n"),
    };
  });

  pi.on("tool_call", async (event, ctx) => {
    await invokeHook("pre-tool", { hook_event_name: "PreToolUse", tool_name: event.toolName, tool_input: event.input, event, cwd: ctx.cwd }, ctx);
  });

  pi.on("tool_result", async (event, ctx) => {
    const payload = {
      hook_event_name: "PostToolUse",
      tool_name: event.toolName,
      tool_input: event.input,
      tool_response: {
        content: event.content,
        details: event.details,
        isError: event.isError,
      },
      event,
      cwd: ctx.cwd,
    };
    await invokeHook("post-tool", payload, ctx);
    if (event.isError) {
      await invokeHook("post-tool-failure", { ...payload, hook_event_name: "PostToolUseFailure" }, ctx);
    }
  });

  pi.on("agent_end", async (event, ctx) => {
    await invokeHook("stop", { hook_event_name: "Stop", event, cwd: ctx.cwd }, ctx);
  });

  registerCodexQuota(pi, (ctx) => updateMoaiStatus(ctx, config));

}
