import { runMoaiHook } from "./hook-bridge.ts";

export type MoaiNotificationLevel = "info" | "warning" | "error" | "success";

export interface MoaiNotificationContext {
  cwd?: string;
  ui?: {
    notify?: (message: string, level?: MoaiNotificationLevel) => void;
  };
}

export interface MoaiNotificationPayload {
  [key: string]: unknown;
}

export async function notifyMoai(
  ctx: MoaiNotificationContext | undefined,
  message: string,
  level: MoaiNotificationLevel = "info",
  payload: MoaiNotificationPayload = {},
): Promise<void> {
  try {
    await runMoaiHook("notification", {
      ...payload,
      hook_event_name: "Notification",
      message,
      level,
      cwd: ctx?.cwd ?? process.cwd(),
    });
  } catch {
    // Notification hooks are telemetry-only; never block the actual Pi notification.
  }

  ctx?.ui?.notify?.(message, level);
}
