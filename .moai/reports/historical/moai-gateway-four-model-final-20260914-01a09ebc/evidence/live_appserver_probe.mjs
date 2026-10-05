import { spawn } from "node:child_process";
import readline from "node:readline";

const child = spawn(
  "/Users/goos/.local/bin/codex",
  [
    "app-server",
    "--stdio",
    "-c", "approval_policy=\"never\"",
    "-c", "sandbox_mode=\"read-only\"",
    "-c", "web_search=\"disabled\"",
    "-c", "features.shell_tool=false",
    "-c", "features.codex_hooks=false",
    "-c", "features.plugins=false",
    "-c", "features.apps=false",
    "-c", "features.multi_agent=false",
  ],
  { cwd: process.cwd(), env: process.env, stdio: ["pipe", "pipe", "pipe"] },
);

const replies = new Map();
const observations = { stderr: [], protocolErrors: [] };
let nextId = 1;

function send(message) {
  child.stdin.write(`${JSON.stringify(message)}\n`);
}

function call(method, params = {}) {
  const id = nextId++;
  send({ id, method, params });
  return new Promise((resolve, reject) => replies.set(String(id), { resolve, reject }));
}

readline.createInterface({ input: child.stdout }).on("line", (line) => {
  let message;
  try {
    message = JSON.parse(line);
  } catch {
    observations.protocolErrors.push("non-json stdout");
    return;
  }
  if (message.id === undefined || message.method) return;
  const pending = replies.get(String(message.id));
  if (!pending) return;
  replies.delete(String(message.id));
  if (message.error) pending.reject(new Error(`RPC ${message.error.code}`));
  else pending.resolve(message.result);
});

readline.createInterface({ input: child.stderr }).on("line", (line) => {
  observations.stderr.push(line.replace(/[A-Za-z0-9_-]{32,}/g, "[redacted]"));
});

const timer = setTimeout(() => {
  child.kill("SIGTERM");
  console.error("probe timed out");
  process.exitCode = 124;
}, 20_000);

try {
  const initialized = await call("initialize", {
    clientInfo: { name: "moai-gateway-readonly-probe", version: "1" },
    capabilities: { experimentalApi: true },
  });
  send({ method: "initialized", params: {} });
  const [account, catalog, capabilities] = await Promise.all([
    call("account/read", { refreshToken: false }),
    call("model/list", { limit: 100, includeHidden: true }),
    call("modelProvider/capabilities/read", {}),
  ]);
  const allow = new Set(["gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-6-astra"]);
  const selected = (catalog.data ?? [])
    .filter((model) => allow.has(model.id) || allow.has(model.model))
    .map((model) => ({
      id: model.id,
      model: model.model,
      displayName: model.displayName,
      description: model.description,
      hidden: model.hidden,
      isDefault: model.isDefault,
      defaultReasoningEffort: model.defaultReasoningEffort,
      supportedReasoningEfforts: model.supportedReasoningEfforts,
      inputModalities: model.inputModalities,
      multiAgentVersion: model.multiAgentVersion,
      supportsPersonality: model.supportsPersonality,
      serviceTiers: model.serviceTiers,
    }));
  console.log(JSON.stringify({
    probe: "codex-app-server-read-only",
    codexVersion: "0.154.0",
    initialized: Boolean(initialized?.userAgent),
    account: account?.account
      ? { type: account.account.type, planType: account.account.planType }
      : null,
    requiresOpenAIAuth: account?.requiresOpenaiAuth,
    catalogCount: catalog.data?.length ?? 0,
    requestedModels: [...allow],
    selected,
    capabilities,
    nextCursor: catalog.nextCursor ?? null,
    protocolErrors: observations.protocolErrors,
    stderrLineCount: observations.stderr.length,
  }, null, 2));
} catch (error) {
  console.error(error.message);
  process.exitCode = 1;
} finally {
  clearTimeout(timer);
  child.stdin.end();
  child.kill("SIGTERM");
}
