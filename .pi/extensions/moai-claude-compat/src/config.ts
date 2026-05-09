import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join, relative, resolve } from "node:path";
import {
  CLAUDE_RULES_SOURCE_PATH,
  EXPECTED_MOAI_RULE_COUNT,
  LANGUAGE_CONFIG_PATH,
  PI_RULES_SOURCE_PATH,
  QUALITY_CONFIG_PATH,
  WORKFLOW_CONFIG_PATH,
} from "./constants.ts";
import { COMPAT_MANIFEST_FILES, getDefaultOutputStyleSource, loadRuntimeManifest } from "./runtime-config.ts";

export interface MoaiCompatConfig {
  conversationLanguage: string;
  codeCommentsLanguage: string;
  qualityMode: string;
  sourceMapPath: string;
  workflowConfigPath: string;
  outputStyle: OutputStyleConfig;
  rules: MoaiRulesConfig;
}

export interface OutputStyleConfig {
  name: string;
  sourcePath: string;
  instruction: string;
  loaded: boolean;
  sanitized: boolean;
  enforcement: "prompt-level";
  error?: string;
}

export interface MoaiRulesConfig {
  claudeSourcePath: string;
  piSourcePath: string;
  expectedCount: number;
  piCount: number;
  claudeCount: number;
  sourceMapRegistered: boolean;
  relativePathParity: boolean;
  missingInPi: string[];
  extraInPi: string[];
  guidance: string;
  enforcement: "prompt-level-triggered";
  loaded: boolean;
  error?: string;
}

function readIfExists(path: string): string {
  const abs = resolve(process.cwd(), path);
  return existsSync(abs) ? readFileSync(abs, "utf8") : "";
}

function pathExists(path: string): boolean {
  return existsSync(resolve(process.cwd(), path));
}

function findYamlScalar(text: string, keys: string[], fallback: string): string {
  for (const key of keys) {
    const match = text.match(new RegExp(`^\\s*${key}\\s*:\\s*[\"']?([^\"'\\n#]+)`, "m"));
    if (match?.[1]) return match[1].trim();
  }
  return fallback;
}

export function loadMoaiCompatConfig(): MoaiCompatConfig {
  const languageYaml = readIfExists(LANGUAGE_CONFIG_PATH);
  const qualityYaml = readIfExists(QUALITY_CONFIG_PATH);

  return {
    conversationLanguage: findYamlScalar(languageYaml, ["conversation_language", "user_responses", "user"], "ko"),
    codeCommentsLanguage: findYamlScalar(languageYaml, ["code_comments", "comments"], "ko"),
    qualityMode: findYamlScalar(qualityYaml, ["mode", "development_mode"], "tdd"),
    sourceMapPath: COMPAT_MANIFEST_FILES.sourceMap,
    workflowConfigPath: WORKFLOW_CONFIG_PATH,
    outputStyle: loadOutputStyleConfig(),
    rules: loadMoaiRulesConfig(),
  };
}

function loadOutputStyleConfig(): OutputStyleConfig {
  const fallback: OutputStyleConfig = {
    name: "moai",
    sourcePath: "",
    instruction: "MoAI output style: use concise Markdown, user's conversation language, no user-facing XML tags.",
    loaded: false,
    sanitized: true,
    enforcement: "prompt-level",
    error: "output style config not loaded",
  };

  try {
    const manifest = loadRuntimeManifest();
    const name = manifest.outputStyles.config.default || "moai";
    const sourcePath = getDefaultOutputStyleSource(manifest);
    if (!sourcePath) {
      return { ...fallback, name, error: `missing source for style ${name}` };
    }
    if (!pathExists(sourcePath)) {
      return { ...fallback, name, sourcePath, error: `missing output style source ${sourcePath}` };
    }

    const raw = readIfExists(sourcePath);
    const sanitized = sanitizeOutputStyleMarkdown(raw);
    return {
      name,
      sourcePath,
      instruction: buildOutputStyleInstruction(name, sanitized),
      loaded: true,
      sanitized: raw !== sanitized,
      enforcement: "prompt-level",
    };
  } catch (error) {
    return { ...fallback, error: error instanceof Error ? error.message : String(error) };
  }
}


function sanitizeOutputStyleMarkdown(markdown: string): string {
  return markdown
    .replace(/<moai>\s*(DONE|COMPLETE)\s*<\/moai>/gi, "MoAI DONE")
    .replace(/<\/?[A-Za-z][^>\n]*>/g, (tag) => `\`${tag}\``)
    .replace(/\bXML tags are reserved for internal agent-to-agent data transfer only\./g, "XML-like tags are reserved for internal data transfer only; never show them in user-facing output.");
}

function buildOutputStyleInstruction(name: string, markdown: string): string {
  const body = stripFrontmatter(markdown).trim();
  return [
    `MoAI output style '${name}' is active at prompt level (not hard post-generation enforcement).`,
    "Follow the style guidance below when responding to users. If any style example conflicts with Pi policy, Pi policy wins: Markdown only, no user-facing XML tags.",
    body,
  ].join("\n\n");
}

function stripFrontmatter(text: string): string {
  return text.replace(/^---\n[\s\S]*?\n---\n?/, "");
}

function walkMarkdownFiles(root: string): string[] {
  const absRoot = resolve(process.cwd(), root);
  if (!existsSync(absRoot)) return [];
  const out: string[] = [];
  const walk = (dir: string) => {
    for (const entry of readdirSync(dir, { withFileTypes: true })) {
      const path = join(dir, entry.name);
      if (entry.isDirectory()) walk(path);
      else if (entry.isFile() && entry.name.endsWith(".md")) out.push(relative(absRoot, path).replace(/\\/g, "/"));
    }
  };
  walk(absRoot);
  return out.sort();
}

function readSourceMapRulesPath(): string {
  try {
    const manifest = loadRuntimeManifest();
    return typeof manifest.sourceMap.config.sources?.rules === "string" ? manifest.sourceMap.config.sources.rules : "";
  } catch {
    return "";
  }
}

function loadMoaiRulesConfig(): MoaiRulesConfig {
  const fallback: MoaiRulesConfig = {
    claudeSourcePath: CLAUDE_RULES_SOURCE_PATH,
    piSourcePath: PI_RULES_SOURCE_PATH,
    expectedCount: EXPECTED_MOAI_RULE_COUNT,
    piCount: 0,
    claudeCount: 0,
    sourceMapRegistered: false,
    relativePathParity: false,
    missingInPi: [],
    extraInPi: [],
    guidance: buildMoaiRulesGuidance(PI_RULES_SOURCE_PATH),
    enforcement: "prompt-level-triggered",
    loaded: false,
    error: "rules config not loaded",
  };

  try {
    const piFiles = walkMarkdownFiles(PI_RULES_SOURCE_PATH);
    const claudeFiles = walkMarkdownFiles(CLAUDE_RULES_SOURCE_PATH);
    const sourceMapRulesPath = readSourceMapRulesPath();
    const normalizedSourceMapRules = sourceMapRulesPath.replace(/^\.\//, ".pi/");
    const sourceMapRegistered = normalizedSourceMapRules === PI_RULES_SOURCE_PATH;
    const piSet = new Set(piFiles);
    const claudeSet = new Set(claudeFiles);
    const missingInPi = claudeFiles.filter((file) => !piSet.has(file));
    const extraInPi = piFiles.filter((file) => !claudeSet.has(file));
    const relativePathParity = claudeFiles.length > 0 && missingInPi.length === 0 && extraInPi.length === 0;
    const expectedCount = loadRuntimeManifest().sourceMap.config.expectedCounts?.rules ?? EXPECTED_MOAI_RULE_COUNT;
    const loaded = piFiles.length === expectedCount && sourceMapRegistered && relativePathParity;

    return {
      ...fallback,
      expectedCount,
      piCount: piFiles.length,
      claudeCount: claudeFiles.length,
      sourceMapRegistered,
      relativePathParity,
      missingInPi,
      extraInPi,
      guidance: buildMoaiRulesGuidance(PI_RULES_SOURCE_PATH),
      loaded,
      error: loaded ? undefined : "rules source count, source-map registration, or relative path parity is incomplete",
    };
  } catch (error) {
    return { ...fallback, error: error instanceof Error ? error.message : String(error) };
  }
}

function buildMoaiRulesGuidance(rulesPath: string): string {
  return [
    `MoAI rules guidance is active at prompt level for MoAI-related inputs. Use the pi-local rules snapshot at ${rulesPath}; do not rely on .claude/rules at runtime.`,
    "Apply the compact MoAI hard rules: answer in the user's conversation language, use Markdown, never display XML tags in user-facing responses, and route complex work through appropriate agents/teams.",
    "For development tasks: explain the approach before non-trivial code changes, decompose changes touching 3+ files, write a reproduction test before bug fixes, and provide post-implementation risks plus suggested tests.",
    "For SPEC/MoAI workflows: consult relevant core/workflow/language rule files from the pi-local rules snapshot before acting, especially workflow/spec-workflow.md and core/moai-constitution.md.",
  ].join("\n");
}

export function outputStyleStatus(config: MoaiCompatConfig): string[] {
  return [
    config.outputStyle.loaded
      ? `ok: output style '${config.outputStyle.name}' loaded from ${config.outputStyle.sourcePath}`
      : `missing: output style '${config.outputStyle.name}' not loaded${config.outputStyle.error ? ` (${config.outputStyle.error})` : ""}`,
    config.outputStyle.sanitized
      ? "ok: output style XML-like user-facing markers sanitized"
      : "info: output style required no XML marker sanitization",
    `partial: output style enforcement is ${config.outputStyle.enforcement}; no hard post-generation output filter is installed`,
  ];
}

export function rulesStatus(config: MoaiCompatConfig): string[] {
  const missingDetail = config.rules.missingInPi.length ? ` missing=${config.rules.missingInPi.slice(0, 5).join(",")}` : "";
  const extraDetail = config.rules.extraInPi.length ? ` extra=${config.rules.extraInPi.slice(0, 5).join(",")}` : "";
  return [
    config.rules.piCount === config.rules.expectedCount
      ? `ok: rules source count ${config.rules.piCount}/${config.rules.expectedCount}`
      : `missing: rules source count ${config.rules.piCount}/${config.rules.expectedCount}`,
    config.rules.sourceMapRegistered
      ? `ok: rules source-map registered at ${config.rules.piSourcePath}`
      : "missing: rules source-map registration for pi-local rules snapshot",
    config.rules.relativePathParity
      ? `ok: rules relative path parity claude=${config.rules.claudeCount} pi=${config.rules.piCount}`
      : `partial: rules relative path parity incomplete claude=${config.rules.claudeCount} pi=${config.rules.piCount}${missingDetail}${extraDetail}`,
    `partial: rules enforcement is ${config.rules.enforcement}; no hard runtime rules engine is installed`,
  ];
}

export function buildCoreInstruction(config: MoaiCompatConfig): string {
  return [
    "MoAI pi compatibility layer is active.",
    `User-facing responses must use conversation_language=${config.conversationLanguage}.`,
    "Use Markdown for user-facing output and do not display XML tags.",
    "Use pi-local source snapshots under .pi/generated/source/**.",
    "Prefer pi packages over custom implementation. Use moai-claude-compat only for schema conversion, glue, and MoAI-specific policy enforcement.",
    "Claude permissionMode is intentionally excluded from parity. Preserve it as metadata only; enforce allow/ask/deny guardrails separately.",
    "Subagents must not ask users directly; escalate decisions to the parent session.",
  ].join("\n");
}
