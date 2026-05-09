/**
 * MoAI built-in Glass notifications for Pi.
 *
 * This embeds the local pi-notify-glass behavior into the MoAI compatibility
 * extension so MoAI users do not need to install a separate pi-notify package
 * or global ~/.pi extension.
 */

import { execFile } from "node:child_process";
import { existsSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

const DISABLE_ENV = "MOAI_NOTIFY_GLASS_DISABLE";
const GLOBAL_GLASS_EXTENSION = join(homedir(), ".pi", "agent", "extensions", "pi-notify-glass.ts");

function escapeAppleScriptString(text: string): string {
  return text.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}

function sendNotification(title: string, body: string): void {
  const safeTitle = escapeAppleScriptString(title);
  const safeBody = escapeAppleScriptString(body);
  const script = `display notification "${safeBody}" with title "${safeTitle}" sound name "Glass"`;

  execFile("osascript", ["-e", script], (err) => {
    if (err) console.error("MoAI Glass notification failed:", err);
  });
}

function truncate(text: string, maxLen = 50): string {
  return text.length > maxLen ? `${text.slice(0, maxLen)}...` : text;
}

function summarizeAskUserQuestionInput(input: unknown): string {
  const record = input && typeof input === "object" ? input as { questions?: unknown } : {};
  const questions = Array.isArray(record.questions) ? record.questions : [];
  const first = questions[0] && typeof questions[0] === "object"
    ? questions[0] as { question?: unknown }
    : undefined;
  const firstQuestion = typeof first?.question === "string"
    ? first.question
    : "답변이 필요한 선택지가 열렸습니다.";
  const suffix = questions.length > 1 ? ` 외 ${questions.length - 1}개` : "";
  return truncate(`${firstQuestion}${suffix}`, 100);
}

function messageText(message: unknown): string {
  const record = message && typeof message === "object" ? message as { content?: unknown } : {};
  const content = record.content;
  if (typeof content === "string") return content;
  if (!Array.isArray(content)) return "";
  return content.map((part: unknown) => {
    const partRecord = part && typeof part === "object" ? part as { text?: unknown } : {};
    return typeof partRecord.text === "string" ? partRecord.text : "";
  }).join("\n");
}

function lastUserInput(messages: unknown[]): string {
  const userMessages = messages.filter((message: unknown) => {
    const record = message && typeof message === "object" ? message as { role?: unknown } : {};
    return record.role === "user";
  });
  return messageText(userMessages[userMessages.length - 1]).trim();
}

function isTeamInternalTurn(messages: unknown[]): boolean {
  const input = lastUserInput(messages);
  return input.startsWith("You are teammate '") || input.startsWith("You have received comrade message(s):");
}

function endedWithAgentError(messages: unknown[]): boolean {
  const assistantMessages = messages.filter((message: unknown) => {
    const record = message && typeof message === "object" ? message as { role?: unknown } : {};
    return record.role === "assistant";
  });
  const lastAssistant = assistantMessages[assistantMessages.length - 1] as { stopReason?: unknown } | undefined;
  return lastAssistant?.stopReason === "error" || lastAssistant?.stopReason === "aborted";
}

function isSubagentNotify(messages: unknown[]): boolean {
  return messages.some((message: unknown) => {
    const record = message && typeof message === "object" ? message as { customType?: unknown } : {};
    return record.customType === "subagent-notify" || JSON.stringify(message).includes("subagent-notify");
  });
}

function lastAssistantToolCallCount(messages: unknown[]): number {
  const assistantMessages = messages.filter((message: unknown) => {
    const record = message && typeof message === "object" ? message as { role?: unknown } : {};
    return record.role === "assistant";
  });
  const lastAssistant = assistantMessages[assistantMessages.length - 1] as { content?: unknown } | undefined;
  const content = Array.isArray(lastAssistant?.content) ? lastAssistant.content : [];
  return content.filter((part: unknown) => {
    const record = part && typeof part === "object" ? part as { type?: unknown } : {};
    return record.type === "toolCall";
  }).length;
}

function shouldEnableGlassNotifications(): boolean {
  if (process.env[DISABLE_ENV] === "1" || process.env[DISABLE_ENV] === "true") return false;
  if (process.env.PI_SUBAGENT_CHILD === "1" || process.env.PI_TEAMS_WORKER === "1") return false;
  if (process.platform !== "darwin") return false;

  // If the user's global pi-notify-glass extension is already installed, let it
  // own notifications to avoid double alerts in existing setups. New MoAI users
  // without that file still get built-in notifications automatically.
  if (existsSync(GLOBAL_GLASS_EXTENSION)) return false;

  return true;
}

export function registerMoaiGlassNotifications(pi: ExtensionAPI): void {
  if (!shouldEnableGlassNotifications()) return;

  pi.on("tool_call", async (event) => {
    if (event.toolName !== "ask_user_question") return;
    sendNotification("MoAI 질문 대기 중", summarizeAskUserQuestionInput(event.input));
  });

  pi.on("agent_end", async (event) => {
    const messages = Array.isArray(event.messages) ? event.messages : [];

    if (isSubagentNotify(messages)) return;
    if (isTeamInternalTurn(messages)) return;
    if (endedWithAgentError(messages)) return;

    const userInput = lastUserInput(messages);
    const toolCallCount = lastAssistantToolCallCount(messages);
    let body = truncate(userInput, 80);

    if (toolCallCount > 0) {
      body += ` | ${toolCallCount} tool${toolCallCount > 1 ? "s" : ""}`;
    }

    sendNotification("MoAI 완료", body);
  });
}
