import { getPackagePriority, loadRuntimeManifest } from "./runtime-config.ts";

export interface PackageConflictFinding {
  level: "ok" | "warn";
  message: string;
}

export type PackageSpec = string | { source?: string };

function packageSpecSource(spec: PackageSpec): string {
  return typeof spec === "string" ? spec : spec.source ?? "";
}

function normalizePackageName(spec: PackageSpec): string {
  let s = packageSpecSource(spec).replace(/^npm:/, "").replace(/^git:/, "");
  s = s.split("#")[0].split("?")[0];
  if (s.startsWith("@")) {
    const parts = s.split("@");
    return parts.length > 2 ? `@${parts[1]}` : s;
  }
  return s.split("@")[0];
}

export function normalizePackageSpecs(specs: PackageSpec[] = []): string[] {
  return specs.map(normalizePackageName).filter(Boolean);
}

function packageCandidates(role: string): string[] {
  return getPackagePriority(loadRuntimeManifest(), role);
}

function roleFinding(role: string, label: string, specs: PackageSpec[]): PackageConflictFinding {
  const names = normalizePackageSpecs(specs);
  const candidates = packageCandidates(role).map(normalizePackageName);
  const active = candidates.filter((candidate) => names.includes(candidate));
  const nativeQuota = role === "quotaFooter" && candidates.includes("moai-claude-compat-native-codex-quota");
  if (nativeQuota && active.length === 0) return { level: "ok", message: `${label} active: moai-claude-compat-native-codex-quota` };
  if (active.length === 0) return { level: "warn", message: `${label} package not active; candidates=${candidates.join(", ") || "none"}` };
  if (active.length > 1) return { level: "warn", message: `Multiple ${label} packages active: ${active.join(", ")}` };
  return { level: "ok", message: `${label} package active: ${active[0]}` };
}

export function analyzePackageConflicts(specs: PackageSpec[] = []): PackageConflictFinding[] {
  const names = normalizePackageSpecs(specs);
  const findings: PackageConflictFinding[] = [];
  const has = (name: string) => names.includes(normalizePackageName(name));
  const packageMap = loadRuntimeManifest().packageMap.config;

  findings.push(roleFinding("agentTeams", "Agent Teams backend", specs));
  findings.push(roleFinding("quotaFooter", "Quota footer", specs));

  const permissions = packageCandidates("permissions");
  const guardrails = packageCandidates("guardrails");
  const hooks = packageCandidates("hooks");
  const activePolicyPackages = [...new Set([...permissions, ...guardrails, ...hooks]
    .map(normalizePackageName)
    .filter((candidate) => names.includes(candidate)))];

  if (activePolicyPackages.length > 1) {
    findings.push({ level: "warn", message: `Multiple policy/guardrail packages active: ${activePolicyPackages.join(", ")}` });
  } else if (activePolicyPackages.length === 1) {
    findings.push({ level: "ok", message: `Policy/guardrail package active: ${activePolicyPackages[0]}` });
  } else {
    findings.push({ level: "warn", message: "No configured policy/guardrail package active" });
  }

  if (has("pi-yaml-hooks") && has("@aliou/pi-guardrails")) {
    findings.push({ level: "warn", message: "pi-yaml-hooks and @aliou/pi-guardrails may both confirm/block dangerous commands; verify ordering" });
  }
  if (has("pi-yaml-hooks")) {
    findings.push({ level: "warn", message: "pi-yaml-hooks@2026.5.8 has known peer dependency risk; validate hook loading after session restart" });
  }
  if (has("@tmustier/pi-agent-teams") && has("pi-subagents")) {
    findings.push({ level: "ok", message: "pi-agent-teams and pi-subagents serve distinct roles: teams backend vs Agent/subagent tool" });
  }
  if (has("context-mode") && has("pi-yaml-hooks")) {
    findings.push({ level: "ok", message: "context-mode active as skills-only package; extension hooks disabled to avoid lifecycle overlap with MoAI/pi-yaml-hooks" });
  }
  if (has("pi-mcp-adapter")) {
    findings.push({ level: "warn", message: "pi-mcp-adapter may overlap with Pi MCP gateway; prefer one MCP route per workflow if duplicate tools appear" });
  }
  if (specs.length === 0) {
    findings.push({ level: "ok", message: "Active packages empty by design for skeleton mode" });
  }

  for (const [policy, description] of Object.entries(packageMap.conflictPolicy ?? {})) {
    findings.push({ level: "ok", message: `Configured package conflict policy ${policy}: ${description}` });
  }

  return findings;
}

export function formatFindings(findings: PackageConflictFinding[]): string[] {
  return findings.map((f) => `${f.level === "ok" ? "ok" : "warn(non-blocking)"}: ${f.message}`);
}
