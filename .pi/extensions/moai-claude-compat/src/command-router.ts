import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { ExtensionAPI, ExtensionCommandContext } from "@earendil-works/pi-coding-agent";
import { buildAuditReport, buildDoctorReport } from "./doctor.ts";
import { inferPhaseFromCommand, setMoaiWidget, updateMoaiStatus } from "./statusline.ts";
import type { MoaiCompatConfig } from "./config.ts";
import { writeConvertedAgents } from "./agent-converter.ts";
import { notifyMoai } from "./notification-adapter.ts";
import { loadRuntimeManifest, resolveCompatSource, type RuntimeManifest } from "./runtime-config.ts";

const PI_PROMPTS_PATH = ".pi/prompts";

const FALLBACK_MOAI_SUBCOMMANDS = [
  "plan", "run", "sync", "project", "fix", "loop", "mx", "review",
  "clean", "codemaps", "coverage", "e2e", "feedback", "context", "gate", "security",
] as const;

const MOAI_SUBCOMMAND_ALIASES: Record<string, string> = {
  spec: "plan",
  impl: "run",
  docs: "sync",
  pr: "sync",
  fb: "feedback",
  bug: "feedback",
  issue: "feedback",
  "code-review": "review",
  "dead-code": "clean",
  cov: "coverage",
  "e2e-test": "e2e",
  ctx: "context",
  memory: "context",
  check: "gate",
  "pre-commit": "gate",
  audit: "security",
  sec: "security",
};

const SPEC_ID_PATTERN = /SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*/i;

function unique(values: Iterable<string>): string[] {
  return [...new Set([...values].map((value) => value.trim()).filter(Boolean))];
}

function configuredMoaiSubcommands(runtime: RuntimeManifest): string[] {
  return unique([
    ...FALLBACK_MOAI_SUBCOMMANDS,
    ...(runtime.commandMap.config.moaiSubcommands ?? []),
    ...Object.keys(runtime.workflowMap.config.workflows ?? {}),
  ]).sort();
}

function normalizeSubcommand(value: string, runtime: RuntimeManifest): string | undefined {
  const subcommands = configuredMoaiSubcommands(runtime);
  if (subcommands.includes(value)) return value;
  const alias = MOAI_SUBCOMMAND_ALIASES[value];
  return alias && subcommands.includes(alias) ? alias : undefined;
}

function stripFrontmatter(text: string): string {
  return text.replace(/^---\n[\s\S]*?\n---\n?/, "").trim();
}

function readPromptBody(path: string): string | undefined {
  const abs = resolve(process.cwd(), path);
  if (!existsSync(abs)) return undefined;
  return stripFrontmatter(readFileSync(abs, "utf8"));
}

function applyArguments(prompt: string, args: string): string {
  return prompt.replaceAll("$ARGUMENTS", args).replaceAll("$@", args).trim();
}

function buildMoaiSkillInvocation(args: string): string {
  return `Use Skill("moai") with arguments: ${args}`.trim();
}

function promptPathForMoaiSubcommand(subcommand: string, runtime: RuntimeManifest): string | undefined {
  const promptPath = `${PI_PROMPTS_PATH}/moai-${subcommand}.md`;
  if (existsSync(resolve(process.cwd(), promptPath))) return promptPath;

  const workflowSource = runtime.workflowMap.config.workflows?.[subcommand]?.source;
  return typeof workflowSource === "string" ? resolveCompatSource(runtime.files.workflowMap, workflowSource, runtime.cwd) : undefined;
}

function promptPathForCommand(commandName: string, runtime: RuntimeManifest): string | undefined {
  const promptPath = `${PI_PROMPTS_PATH}/${commandName}.md`;
  if (existsSync(resolve(process.cwd(), promptPath))) return promptPath;

  const source = runtime.commandMap.config.commands?.[commandName]?.source;
  return typeof source === "string" ? resolveCompatSource(runtime.files.commandMap, source, runtime.cwd) : undefined;
}

function buildMoaiPrompt(subcommand: string, args: string, runtime = loadRuntimeManifest()): string {
  const normalized = subcommand || "auto";
  const skillArgs = normalized === "auto" ? args.trim() : `${normalized}${args ? ` ${args}` : ""}`.trim();
  const promptPath = normalized === "auto" ? undefined : promptPathForMoaiSubcommand(normalized, runtime);
  const promptBody = promptPath ? readPromptBody(promptPath) : undefined;
  return applyArguments(promptBody ?? buildMoaiSkillInvocation(skillArgs), args);
}

function buildCommandPrompt(commandName: string, args: string, runtime = loadRuntimeManifest()): string {
  const promptPath = promptPathForCommand(commandName, runtime);
  return applyArguments((promptPath ? readPromptBody(promptPath) : undefined) ?? `Arguments provided: $ARGUMENTS`, args);
}

async function dispatchMoai(pi: ExtensionAPI, args: string, ctx: ExtensionCommandContext, config: MoaiCompatConfig) {
  const runtime = loadRuntimeManifest();
  const [first = "", ...rest] = args.trim().split(/\s+/).filter(Boolean);
  const canonical = normalizeSubcommand(first, runtime);
  const subcommand = canonical ?? "auto";
  const remaining = canonical ? rest.join(" ") : args.trim();
  updateMoaiStatus(ctx, config, { phase: inferPhaseFromCommand(subcommand), specId: remaining.match(SPEC_ID_PATTERN)?.[0] });
  pi.sendUserMessage(buildMoaiPrompt(subcommand, remaining, runtime));
}

function registerMappedCommand(pi: ExtensionAPI, commandName: string, config: MoaiCompatConfig, runtime: RuntimeManifest) {
  if (commandName === "moai") return;
  pi.registerCommand(commandName, {
    description: `Run MoAI mapped command '${commandName}' from pi compat manifest`,
    handler: async (args, ctx) => {
      updateMoaiStatus(ctx, config, { phase: inferPhaseFromCommand(commandName) });
      pi.sendUserMessage(buildCommandPrompt(commandName, args, runtime));
    },
  });
}

export function registerCommands(pi: ExtensionAPI, config: MoaiCompatConfig) {
  const runtime = loadRuntimeManifest();
  const subcommands = configuredMoaiSubcommands(runtime);

  pi.registerCommand("moai", {
    description: "Route to the MoAI Claude harness through pi compatibility layer",
    getArgumentCompletions: (prefix) => {
      const candidates = [...subcommands, ...Object.keys(MOAI_SUBCOMMAND_ALIASES)].sort();
      const matches = candidates.filter((s) => s.startsWith(prefix));
      return matches.length ? matches.map((value) => ({ value, label: value })) : null;
    },
    handler: async (args, ctx) => dispatchMoai(pi, args, ctx, config),
  });

  for (const subcommand of subcommands) {
    pi.registerCommand(`moai-${subcommand}`, {
      description: `Shortcut for /moai ${subcommand}`,
      handler: async (args, ctx) => dispatchMoai(pi, `${subcommand} ${args}`.trim(), ctx, config),
    });
  }

  for (const commandName of unique(["github", "release", ...Object.keys(runtime.commandMap.config.commands ?? {})])) {
    registerMappedCommand(pi, commandName, config, runtime);
  }

  pi.registerCommand("moai-pi-doctor", {
    description: "Check MoAI pi overlay/package status",
    handler: async (_args, ctx) => {
      const lines = buildDoctorReport();
      setMoaiWidget(ctx, lines);
      await notifyMoai(ctx, lines.join("\n"), "info", { source: "moai-pi-doctor", command: "moai-pi-doctor" });
    },
  });

  pi.registerCommand("moai-pi-audit", {
    description: "Show MoAI pi parity audit checklist",
    handler: async (_args, ctx) => {
      const lines = buildAuditReport();
      setMoaiWidget(ctx, lines);
      await notifyMoai(ctx, lines.join("\n"), "info", { source: "moai-pi-audit", command: "moai-pi-audit" });
    },
  });

  pi.registerCommand("moai-pi-generate-agents", {
    description: "Generate normalized pi-side agent metadata from pi-local agent snapshot without modifying upstream sources",
    handler: async (_args, ctx) => {
      const agents = writeConvertedAgents();
      const lines = ["MoAI pi agent generation", `generated: ${agents.length} files under .pi/generated/agents`, "permissionMode: preserved as metadata-only"];
      setMoaiWidget(ctx, lines);
      await notifyMoai(ctx, lines.join("\n"), "info", { source: "moai-pi-generate-agents", command: "moai-pi-generate-agents" });
    },
  });

  pi.registerCommand("moai-pi-validate", {
    description: "Run MoAI pi validation report",
    handler: async (_args, ctx) => {
      const lines = ["MoAI pi validation", ...buildAuditReport().slice(1)];
      setMoaiWidget(ctx, lines);
      await notifyMoai(ctx, lines.join("\n"), "info", { source: "moai-pi-validate", command: "moai-pi-validate" });
    },
  });
}
