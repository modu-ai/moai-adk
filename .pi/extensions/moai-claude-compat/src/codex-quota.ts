import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";

// Adapted from @kmiyh/pi-codex-plan-limits (MIT) for MoAI native footer integration.
const POLL_INTERVAL_MS = 60_000;
const MIN_EVENT_REFRESH_MS = 15_000;
const STALE_THRESHOLD_MS = 15 * 60_000;
const OPENAI_CODEX_PROVIDER = "openai-codex";
const DEFAULT_CHATGPT_BASE_URL = "https://chatgpt.com/backend-api/";

type UsageWindow = {
  label: string;
  usedPercent: number;
  resetsAtMs?: number;
};

type LimitsSnapshot = {
  source: "live" | "cached";
  capturedAtMs: number;
  planType?: string;
  primary?: UsageWindow;
  secondary?: UsageWindow;
  stale: boolean;
  error?: string;
};

type PiOpenAICodexOAuthCredential = {
  type: "oauth";
  access?: string;
  refresh?: string;
  expires?: number;
  accountId?: string;
};

type UsagePayloadWindow = {
  used_percent?: number;
  limit_window_seconds?: number;
  reset_at?: number;
};

type UsagePayload = {
  plan_type?: string;
  rate_limit?: {
    primary_window?: UsagePayloadWindow | null;
    secondary_window?: UsagePayloadWindow | null;
  } | null;
};

let latestSnapshot: LimitsSnapshot | undefined;
let refreshInFlight: Promise<void> | undefined;
let refreshInFlightKey = "";
let refreshGeneration = 0;
let pollTimer: ReturnType<typeof setInterval> | undefined;
let activeCtx: ExtensionContext | undefined;
let lastRefreshStartedAt = 0;
let shutdownRequested = false;

export function registerCodexQuota(pi: ExtensionAPI, onUpdate: (ctx: ExtensionContext) => void): void {
  async function refresh(ctx: ExtensionContext, options?: { force?: boolean; notify?: boolean }): Promise<void> {
    activeCtx = ctx;
    if (!shouldShowForModel(ctx)) {
      latestSnapshot = undefined;
      refreshGeneration++;
      onUpdate(ctx);
      return;
    }

    const now = Date.now();
    const requestKey = modelContextKey(ctx);
    if (!options?.force && now - lastRefreshStartedAt < 2_000) return refreshInFlight;
    if (!options?.force && refreshInFlight && refreshInFlightKey === requestKey) return refreshInFlight;

    const generation = ++refreshGeneration;
    lastRefreshStartedAt = now;
    refreshInFlightKey = requestKey;
    const currentRefresh = (async () => {
      let nextSnapshot: LimitsSnapshot | undefined;
      try {
        nextSnapshot = await loadBestSnapshot(ctx, latestSnapshot);
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        nextSnapshot = latestSnapshot
          ? { ...latestSnapshot, source: "cached", stale: true, error: message }
          : { source: "cached", capturedAtMs: Date.now(), stale: true, error: message };
      } finally {
        if (refreshInFlight === currentRefresh) refreshInFlight = undefined;
        const stillCurrent = !shutdownRequested
          && generation === refreshGeneration
          && activeCtx === ctx
          && modelContextKey(ctx) === requestKey
          && shouldShowForModel(ctx);
        if (stillCurrent) {
          latestSnapshot = nextSnapshot;
          onUpdate(ctx);
        }
      }
    })();
    refreshInFlight = currentRefresh;

    if (options?.notify) {
      void currentRefresh.then(() => {
        const text = latestSnapshot ? buildNotification(latestSnapshot) : "Codex limits unavailable";
        ctx.ui.notify(text, latestSnapshot?.stale ? "warning" : "info");
      });
    }

    return currentRefresh;
  }

  function refreshInBackground(ctx: ExtensionContext, options?: { force?: boolean; notify?: boolean }): void {
    void refresh(ctx, options);
  }

  function startPolling(ctx: ExtensionContext): void {
    activeCtx = ctx;
    stopPolling();
    pollTimer = setInterval(() => {
      if (activeCtx) refreshInBackground(activeCtx);
    }, POLL_INTERVAL_MS);
  }

  function stopPolling(): void {
    if (!pollTimer) return;
    clearInterval(pollTimer);
    pollTimer = undefined;
  }

  function refreshIfDue(ctx: ExtensionContext): void {
    activeCtx = ctx;
    if (!shouldShowForModel(ctx)) {
      latestSnapshot = undefined;
      onUpdate(ctx);
      return;
    }
    if (Date.now() - lastRefreshStartedAt < MIN_EVENT_REFRESH_MS) {
      onUpdate(ctx);
      return;
    }
    refreshInBackground(ctx);
  }

  pi.on("session_start", async (_event, ctx) => {
    shutdownRequested = false;
    activeCtx = ctx;
    startPolling(ctx);
    refreshInBackground(ctx, { force: true });
  });

  pi.on("model_select", async (_event, ctx) => {
    activeCtx = ctx;
    refreshInBackground(ctx, { force: true });
  });

  pi.on("turn_end", async (_event, ctx) => {
    refreshIfDue(ctx);
  });

  pi.on("session_shutdown", async () => {
    shutdownRequested = true;
    refreshGeneration++;
    stopPolling();
    activeCtx = undefined;
  });
}

export function getCodexQuotaFooterText(width: number): string | undefined {
  if (!latestSnapshot || (!latestSnapshot.primary && !latestSnapshot.secondary)) return undefined;
  const text = buildLimitsText(latestSnapshot, width);
  if (!text) return undefined;
  return latestSnapshot.stale || latestSnapshot.source === "cached" ? `${text} (cached)` : text;
}

export function hasActiveCodexQuotaContext(): boolean {
  return Boolean(activeCtx && shouldShowForModel(activeCtx));
}

async function loadBestSnapshot(ctx: ExtensionContext, previousSnapshot: LimitsSnapshot | undefined): Promise<LimitsSnapshot> {
  try {
    return await fetchLiveSnapshotFromPiAuth(ctx);
  } catch (error) {
    if (!previousSnapshot) throw error;
    const message = error instanceof Error ? error.message : String(error);
    return {
      ...previousSnapshot,
      source: "cached",
      stale: Date.now() - previousSnapshot.capturedAtMs > STALE_THRESHOLD_MS,
      error: message,
    };
  }
}

async function fetchLiveSnapshotFromPiAuth(ctx: ExtensionContext): Promise<LimitsSnapshot> {
  if (!ctx.model || ctx.model.provider !== OPENAI_CODEX_PROVIDER) {
    throw new Error("Active Pi model is not an OpenAI Codex subscription model");
  }

  const authResult = await ctx.modelRegistry.getApiKeyAndHeaders(ctx.model);
  if (!authResult.ok) throw new Error(authResult.error);

  const credential = ctx.modelRegistry.authStorage.get(OPENAI_CODEX_PROVIDER) as PiOpenAICodexOAuthCredential | undefined;
  const accessToken = credential?.access;
  const accountId = credential?.accountId;
  if (!accessToken || !accountId) {
    throw new Error("Missing Pi OpenAI Codex OAuth credentials. Run /login and select OpenAI Codex.");
  }

  const baseUrl = ensureTrailingSlash(process.env.PI_CODEX_CHATGPT_BASE_URL ?? DEFAULT_CHATGPT_BASE_URL);
  const usageUrl = new URL("wham/usage", baseUrl).toString();
  const response = await fetch(usageUrl, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "chatgpt-account-id": accountId,
      "Content-Type": "application/json",
    },
    signal: AbortSignal.timeout(15_000),
  });

  if (!response.ok) {
    const body = await safeReadText(response);
    throw new Error(`Usage request failed (${response.status}): ${truncateInline(body, 200)}`);
  }

  const payload = (await response.json()) as UsagePayload;
  const snapshot: LimitsSnapshot = {
    source: "live",
    capturedAtMs: Date.now(),
    planType: normalizePlanType(payload.plan_type),
    primary: mapWindow(payload.rate_limit?.primary_window, "5h"),
    secondary: mapWindow(payload.rate_limit?.secondary_window, "Weekly"),
    stale: false,
  };

  if (!snapshot.primary && !snapshot.secondary) {
    throw new Error("Usage response did not contain 5h/weekly windows");
  }

  return snapshot;
}

function shouldShowForModel(ctx: ExtensionContext): boolean {
  return Boolean(ctx.hasUI && ctx.model?.provider === OPENAI_CODEX_PROVIDER && ctx.modelRegistry.isUsingOAuth(ctx.model));
}

function modelContextKey(ctx: ExtensionContext): string {
  const model = ctx.model;
  const oauth = model ? ctx.modelRegistry.isUsingOAuth(model) : false;
  return `${model?.provider ?? "none"}:${model?.id ?? "none"}:${oauth ? "oauth" : "api-key"}`;
}

function mapWindow(window: UsagePayloadWindow | null | undefined, fallbackLabel: string): UsageWindow | undefined {
  if (!window) return undefined;
  const usedPercent = sanitizeNumber(window.used_percent);
  const resetsAtMs = secondsToMs(window.reset_at);
  const windowMinutes = secondsToMinutes(window.limit_window_seconds);
  if (usedPercent === undefined && resetsAtMs === undefined && windowMinutes === undefined) return undefined;
  return {
    label: labelForWindow(windowMinutes, fallbackLabel),
    usedPercent: clamp(usedPercent ?? 0, 0, 100),
    resetsAtMs,
  };
}

function buildLimitsText(snapshot: LimitsSnapshot, width: number): string {
  const compact = width < 110;
  const barWidth = compact ? 10 : 10;
  const windows = [
    snapshot.primary ? formatNativeWindow("5H:", snapshot.primary.usedPercent, snapshot.primary.resetsAtMs, barWidth, "relative") : undefined,
    snapshot.secondary ? formatNativeWindow("7D:", snapshot.secondary.usedPercent, snapshot.secondary.resetsAtMs, barWidth, "absolute") : undefined,
  ].filter(Boolean);
  return windows.join(" │ ");
}

function formatNativeWindow(label: string, usedPercent: number, resetsAtMs: number | undefined, barWidth: number, _resetMode: "relative" | "absolute"): string {
  const pct = Math.round(clamp(usedPercent, 0, 100));
  const resetText = resetsAtMs ? formatResetDuration(resetsAtMs) : "";
  const base = `${label} ${batteryIcon(pct)} ${renderNativeBar(pct, barWidth)} ${formatPercent(pct)}`;
  return resetText ? `${base} (${resetText})` : base;
}

function renderNativeBar(usedPercent: number, width: number): string {
  const filled = Math.max(0, Math.min(width, Math.round((clamp(usedPercent, 0, 100) / 100) * width)));
  return `${"█".repeat(filled)}${"░".repeat(width - filled)}`;
}

function batteryIcon(usedPercent: number): string {
  return usedPercent > 70 ? "🪫" : "🔋";
}

function formatResetDuration(timestampMs: number, nowMs = Date.now()): string {
  const remaining = timestampMs - nowMs;
  if (remaining <= 0) return "";
  const totalMinutes = Math.max(1, Math.round(remaining / 60_000));
  const days = Math.floor(totalMinutes / 1_440);
  const hours = Math.floor((totalMinutes % 1_440) / 60);
  const minutes = totalMinutes % 60;
  if (days > 0) return `${days}d ${hours}h ${minutes}m`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

export function formatResetDurationForTest(timestampMs: number, nowMs: number): string {
  return formatResetDuration(timestampMs, nowMs);
}

function buildNotification(snapshot: LimitsSnapshot): string {
  const source = snapshot.source === "live" ? "live" : "cached";
  const parts: string[] = [];
  if (snapshot.primary) parts.push(`${snapshot.primary.label} ${formatPercent(100 - snapshot.primary.usedPercent)} left`);
  if (snapshot.secondary) parts.push(`${snapshot.secondary.label} ${formatPercent(100 - snapshot.secondary.usedPercent)} left`);
  return parts.length > 0 ? `Codex limits refreshed (${source}): ${parts.join(" · ")}` : `Codex limits refreshed (${source})`;
}

function formatPercent(value: number): string {
  return `${Math.round(value)}%`;
}

function labelForWindow(windowMinutes: number | undefined, fallbackLabel: string): string {
  if (windowMinutes === 300) return "5h";
  if (windowMinutes === 10_080) return "Weekly";
  if (windowMinutes === undefined) return fallbackLabel;
  if (windowMinutes >= 60 && windowMinutes % 60 === 0) return `${windowMinutes / 60}h`;
  return `${windowMinutes}m`;
}

function sanitizeNumber(value: number | undefined): number | undefined {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}

function secondsToMinutes(value: number | undefined): number | undefined {
  const seconds = sanitizeNumber(value);
  if (seconds === undefined || seconds <= 0) return undefined;
  return Math.ceil(seconds / 60);
}

function secondsToMs(value: number | undefined): number | undefined {
  const seconds = sanitizeNumber(value);
  if (seconds === undefined || seconds <= 0) return undefined;
  return seconds * 1000;
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

function normalizePlanType(value: string | undefined): string | undefined {
  return value?.replace(/_/g, " ");
}

function ensureTrailingSlash(value: string): string {
  return value.endsWith("/") ? value : `${value}/`;
}

function truncateInline(value: string, limit: number): string {
  const normalized = value.replace(/\s+/g, " ").trim();
  return normalized.length <= limit ? normalized : `${normalized.slice(0, limit - 1)}…`;
}

async function safeReadText(response: Response): Promise<string> {
  try {
    return await response.text();
  } catch {
    return "";
  }
}
