import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { basename, join } from "node:path";
import { createHash } from "node:crypto";
import {
  asString,
  asStringArray,
  parseMarkdownFrontmatter,
} from "./frontmatter.ts";
import {
  PI_AGENTS_SOURCE_PATH,
  PI_SKILLS_SOURCE_PATH,
} from "./constants.ts";
import { loadRuntimeManifest } from "./runtime-config.ts";

export interface ConvertedAgent {
  name: string;
  description: string;
  model: string;
  thinking: string;
  tools: string[];
  skills: string[];
  sourcePath: string;
  sourceHash: string;
  generatedAt: string;
  body: string;
  compatibility: {
    permissionMode: string;
    permissionModePolicy: "metadata-only";
    originalModel: string;
    originalMaxTurns: string;
    originalMemory: string;
    originalTools: string[];
    toolAliases: Record<string, string>;
    hasAgentHooks: boolean;
  };
}

const MODEL_MAP: Record<string, { model: string; thinking: string }> = {
  opus: { model: "inherit", thinking: "high" },
  sonnet: { model: "inherit", thinking: "medium" },
  haiku: { model: "inherit", thinking: "low" },
  inherit: { model: "inherit", thinking: "medium" },
};

const FALLBACK_TOOL_ALIASES: Record<string, string> = {
  Read: "read",
  Write: "write",
  Edit: "edit",
  MultiEdit: "edit",
  Grep: "bash:rg",
  Glob: "bash:find",
  Bash: "bash",
  TodoWrite: "@juicesharp/rpiv-todo",
  Agent: "pi-subagents:subagent",
  Skill: "pi skills/read",
  AskUserQuestion: "@juicesharp/rpiv-ask-user-question",
  WebSearch: "pi-web-access:web_search",
  WebFetch: "pi-web-access:fetch_content",
  TeamCreate: "@tmustier/pi-agent-teams",
  SendMessage: "@tmustier/pi-agent-teams",
  TaskCreate: "@tmustier/pi-agent-teams or pi-crew task state",
  TaskUpdate: "@tmustier/pi-agent-teams or pi-crew task state",
  TaskList: "@tmustier/pi-agent-teams or pi-crew task state",
  TaskGet: "@tmustier/pi-agent-teams or pi-crew task state",
  TeamDelete: "@tmustier/pi-agent-teams",
};

const DIRECT_RUNTIME_TOOLS = new Set([
  "ask_user_question",
  "bash",
  "code_search",
  "document_parse",
  "edit",
  "fetch_content",
  "get_search_content",
  "mcp",
  "read",
  "subagent",
  "teams",
  "web_search",
  "write",
]);

function hash(text: string): string {
  return createHash("sha256").update(text).digest("hex").slice(0, 16);
}

function loadToolAliases(): Record<string, string> {
  try {
    return { ...FALLBACK_TOOL_ALIASES, ...loadRuntimeManifest().toolAliases.config.aliases };
  } catch {
    return FALLBACK_TOOL_ALIASES;
  }
}

function runtimeToolsFromAlias(alias: string): string[] {
  const normalized = alias.trim();
  if (!normalized) return [];
  if (DIRECT_RUNTIME_TOOLS.has(normalized)) return [normalized];
  if (normalized.startsWith("bash:")) return ["bash"];
  if (normalized.startsWith("pi-web-access:")) {
    const tool = normalized.split(":", 2)[1];
    return DIRECT_RUNTIME_TOOLS.has(tool) ? [tool] : [];
  }
  if (normalized === "pi-subagents:subagent") return ["subagent"];
  if (normalized === "pi skills/read") return ["read"];
  if (normalized === "@juicesharp/rpiv-ask-user-question") {
    return ["ask_user_question"];
  }
  if (normalized === "@tmustier/pi-agent-teams") return ["teams"];

  // No dedicated Pi TodoWrite runtime tool is activated in this compat layer.
  // Preserve TodoWrite via compatibility metadata, but do not emit a fake tool.
  if (normalized === "@juicesharp/rpiv-todo") return [];

  if (normalized.includes(" or ")) {
    return normalized.split(" or ").flatMap((part) => runtimeToolsFromAlias(part));
  }

  return [];
}

function normalizeTools(
  value: unknown,
  aliases = loadToolAliases(),
): string[] {
  const tools = new Set<string>();
  for (const raw of asStringArray(value)) {
    const name = raw.trim();
    if (!name) continue;
    if (name.startsWith("mcp__")) {
      tools.add("mcp");
      continue;
    }
    const alias = aliases[name] ?? name.toLowerCase();
    for (const mapped of runtimeToolsFromAlias(alias)) tools.add(mapped);
  }
  return [...tools].sort();
}

function collectToolAliases(
  originalTools: string[],
  aliases = loadToolAliases(),
): Record<string, string> {
  const result: Record<string, string> = {};
  for (const tool of originalTools) {
    if (tool.startsWith("mcp__")) {
      result[tool] = "mcp gateway";
    } else {
      result[tool] = aliases[tool] ?? tool.toLowerCase();
    }
  }
  return result;
}

function normalizeRuntimeReferences(body: string): string {
  return body;
}

function hasTopLevelField(text: string, field: string): boolean {
  const escaped = field.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  return new RegExp(`^${escaped}:`, "m").test(text.slice(4, text.indexOf("\n---", 4)));
}

function yamlQuoted(value: string): string {
  return JSON.stringify(value.replace(/\s+/g, " ").trim());
}

function renderAgentMarkdown(agent: ConvertedAgent): string {
  const description = yamlQuoted(
    agent.description || `Generated MoAI agent ${agent.name}`,
  );
  const tools = agent.tools.join(", ");
  const skillsLine = agent.skills.length
    ? `skills: ${agent.skills.join(", ")}\n`
    : "";
  const skillHints = agent.skills
    .map(
      (skill) =>
        `- Read skill '${skill}' from ${PI_SKILLS_SOURCE_PATH} when relevant.`,
    )
    .join("\n");
  const modelLine = agent.model === "inherit" ? "" : `model: ${agent.model}\n`;
  const aliasHints = Object.entries(agent.compatibility.toolAliases)
    .map(([source, alias]) => `${source} -> ${alias}`)
    .join("; ");
  return `---\nname: ${agent.name}\ndescription: ${description}\n${modelLine}thinking: ${agent.thinking}\ntools: ${tools}\n${skillsLine}systemPromptMode: replace\ninheritProjectContext: true\ninheritSkills: false\n---\n\n# Generated MoAI pi agent: ${agent.name}\n\nSource: ${agent.sourcePath}\nSource hash: ${agent.sourceHash}\nGenerated: ${agent.generatedAt}\n\nCompatibility metadata:\n\n- Runtime model: parent session default (model field omitted for inherit)\n- Original model tier: ${agent.compatibility.originalModel}\n- Original maxTurns: ${agent.compatibility.originalMaxTurns}\n- Original memory scope: ${agent.compatibility.originalMemory}\n- Original permissionMode: ${agent.compatibility.permissionMode}\n- permissionMode policy: metadata-only, excluded-by-design\n- Original Claude tools: ${agent.compatibility.originalTools.join(", ") || "none"}\n- Tool alias policy: ${aliasHints || "none"}\n- Original agent-local hooks: ${agent.compatibility.hasAgentHooks ? "preserved in source snapshot; Pi runtime uses project hook bridge/global pi-yaml-hooks" : "none"}\n\nPi compatibility notes:\n\n- Runtime reference files are resolved from .pi/generated/source/**.\n- Runtime tools are resolved from .pi/claude-compat/tool-aliases.json and emitted only when Pi has a matching callable tool.\n- Claude MCP tool names such as mcp__context7__* and mcp__sequential-thinking__* are used through Pi's mcp gateway tool.\n- Subagents escalate user decisions to the parent session.\n- If a referenced Claude tool is unavailable in pi, use the mapped package/tool or report the gap.\n\nSkill preload hints:\n\n${skillHints || "- No explicit skills listed."}\n\n---\n\n${agent.body}\n`;
}

export function convertAgents(
  sourceDir = PI_AGENTS_SOURCE_PATH,
): ConvertedAgent[] {
  if (!existsSync(sourceDir)) return [];
  return readdirSync(sourceDir)
    .filter((name) => name.endsWith(".md"))
    .sort()
    .map((file) => {
      const sourcePath = join(sourceDir, file);
      const text = readFileSync(sourcePath, "utf8");
      const parsed = parseMarkdownFrontmatter(text);
      const rawModel = asString(parsed.frontmatter.model, "inherit");
      const originalTools = asStringArray(parsed.frontmatter.tools);
      const toolAliases = collectToolAliases(originalTools);
      const modelInfo = MODEL_MAP[rawModel] ?? {
        model: rawModel,
        thinking: "medium",
      };
      return {
        name: asString(parsed.frontmatter.name, basename(file, ".md")),
        description: asString(parsed.frontmatter.description, ""),
        model: modelInfo.model,
        thinking: modelInfo.thinking,
        tools: normalizeTools(parsed.frontmatter.tools),
        skills: asStringArray(parsed.frontmatter.skills),
        sourcePath,
        sourceHash: hash(text),
        generatedAt: new Date().toISOString(),
        body: normalizeRuntimeReferences(parsed.body),
        compatibility: {
          permissionMode: asString(
            parsed.frontmatter.permissionMode,
            "unspecified",
          ),
          permissionModePolicy: "metadata-only",
          originalModel: rawModel,
          originalMaxTurns: asString(parsed.frontmatter.maxTurns, "unspecified"),
          originalMemory: asString(parsed.frontmatter.memory, "unspecified"),
          originalTools,
          toolAliases,
          hasAgentHooks: hasTopLevelField(text, "hooks"),
        },
      } satisfies ConvertedAgent;
    });
}

export function writeConvertedAgents(
  jsonDir = ".pi/generated/agents",
  markdownDir = ".pi/agents/moai",
): ConvertedAgent[] {
  const agents = convertAgents();
  mkdirSync(jsonDir, { recursive: true });
  mkdirSync(markdownDir, { recursive: true });
  for (const agent of agents) {
    writeFileSync(
      join(jsonDir, `${agent.name}.json`),
      `${JSON.stringify({ ...agent, body: undefined }, null, 2)}\n`,
      "utf8",
    );
    writeFileSync(
      join(markdownDir, `${agent.name}.md`),
      renderAgentMarkdown(agent),
      "utf8",
    );
  }
  return agents;
}

export function getAgentConversionStatus(expected = 25): string[] {
  const agents = convertAgents();
  const generatedJsonDir = ".pi/generated/agents";
  const generatedMarkdownDir = ".pi/agents/moai";
  const generatedJson = existsSync(generatedJsonDir)
    ? readdirSync(generatedJsonDir).filter((f) => f.endsWith(".json")).length
    : 0;
  const generatedMarkdown = existsSync(generatedMarkdownDir)
    ? readdirSync(generatedMarkdownDir).filter((f) => f.endsWith(".md")).length
    : 0;
  const latestSource = agents.reduce(
    (latest, agent) => Math.max(latest, statSync(agent.sourcePath).mtimeMs),
    0,
  );
  const sourceStatus = agents.length === expected ? "ok" : "warn(non-blocking)";
  const jsonStatus = generatedJson === expected ? "ok" : "pending";
  const markdownStatus = generatedMarkdown === expected ? "ok" : "pending";
  return [
    `${sourceStatus}: agent source count ${agents.length}/${expected}${latestSource ? ` latest=${new Date(latestSource).toISOString()}` : ""}`,
    `${jsonStatus}: generated agent JSON artifacts ${generatedJson}/${expected}`,
    `${markdownStatus}: generated pi-subagents markdown artifacts ${generatedMarkdown}/${expected}`,
    "ok: permissionMode preserved as metadata-only in converted agents",
  ];
}
