#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs"
import { resolve } from "node:path"

const DEFAULT_PERMISSIONS = {
  ask: ["Bash(rm*)", "Bash(sudo*)", "Bash(chmod*)", "Bash(chown*)", "Read(.env*)"],
  deny: [
    "Read/Write/Edit/Grep/Glob(./secrets/**)",
    "Read/Write/Edit/Grep/Glob(~/.ssh/**)",
    "Read/Write/Edit/Grep/Glob(~/.aws/**)",
    "Read/Write/Edit/Grep/Glob(~/.config/gcloud/**)",
    "Bash(git push --force*)",
    "Bash(git push -f*)",
    "Bash(git push --force-with-lease*)",
    "Bash(git reset --hard*)",
    "Bash(git clean -fd*)",
    "Bash(git clean -fdx*)",
    "Bash(git rebase -i*)",
    "Bash(chmod 777*)",
    "Bash(chmod -R 777*)",
    "Bash(DROP DATABASE*)",
    "Bash(DROP TABLE*)",
    "Bash(TRUNCATE*)",
    "Bash(DELETE FROM*)",
    "Bash(redis-cli FLUSHALL*)",
    "Bash(redis-cli FLUSHDB*)",
    "Bash(psql -c DROP*)",
    "Bash(mysql -e DROP*)",
    "Bash(curl|wget pipe-to-shell*)",
  ],
}

const payload = JSON.parse(await readStdin())
const toolName = normalizeToolName(String(payload.tool_name ?? payload.toolName ?? payload.name ?? ""))
const toolArgs = isRecord(payload.tool_args)
  ? payload.tool_args
  : isRecord(payload.tool_input)
    ? payload.tool_input
    : isRecord(payload.input)
      ? payload.input
      : {}
const permissions = loadPermissions()

const denyRule = findMatchingRule(permissions.deny ?? [], toolName, toolArgs)
if (denyRule) deny(`MoAI pi guard denied ${toolName}: matched permissions.json deny rule ${denyRule}`)

const askRule = findMatchingRule(permissions.ask ?? [], toolName, toolArgs)
if (askRule) ask(`MoAI pi guard requires approval for ${toolName}: matched permissions.json ask rule ${askRule}`)

function loadPermissions() {
  const path = resolve(process.cwd(), ".pi/claude-compat/permissions.json")
  try {
    if (!existsSync(path)) return DEFAULT_PERMISSIONS
    const parsed = JSON.parse(readFileSync(path, "utf8"))
    return {
      ...DEFAULT_PERMISSIONS,
      ...parsed,
      ask: Array.isArray(parsed.ask) ? parsed.ask.filter((rule) => typeof rule === "string") : DEFAULT_PERMISSIONS.ask,
      deny: Array.isArray(parsed.deny) ? parsed.deny.filter((rule) => typeof rule === "string") : DEFAULT_PERMISSIONS.deny,
    }
  } catch {
    return DEFAULT_PERMISSIONS
  }
}

function findMatchingRule(rules, toolName, toolArgs) {
  for (const rule of rules) {
    if (ruleMatches(rule, toolName, toolArgs)) return rule
  }
  return ""
}

function ruleMatches(rule, toolName, toolArgs) {
  const parsed = parseRule(rule)
  if (!parsed || !parsed.tools.includes(toolName)) return false

  if (toolName === "bash") {
    const command = String(toolArgs.command ?? "").replace(/\s+/g, " ").trim()
    return bashPatternMatches(parsed.pattern, command)
  }

  const filePath = getToolPath(toolArgs)
  return pathPatternMatches(parsed.pattern, filePath)
}

function parseRule(rule) {
  const match = String(rule).match(/^([^()]+)\((.*)\)$/)
  if (!match) return undefined
  return {
    tools: match[1].split("/").map(normalizeToolName).filter(Boolean),
    pattern: match[2].trim(),
  }
}

function normalizeToolName(name) {
  const lower = String(name).toLowerCase()
  if (lower === "multi_edit" || lower === "multiedit") return "edit"
  return lower
}

function getToolPath(toolArgs) {
  return String(
    toolArgs.path
      ?? toolArgs.filePath
      ?? toolArgs.file_path
      ?? toolArgs.file
      ?? toolArgs.pattern
      ?? ""
  )
}

function bashPatternMatches(pattern, command) {
  const trimmed = String(pattern).trim()
  if (!trimmed) return false
  if (/^curl\|wget pipe-to-shell/i.test(trimmed)) return /(curl|wget)[^|;&]*\|\s*(sh|bash)(\s|$)/i.test(command)

  const source = wildcardToRegex(trimmed).replace(/\\ /g, "\\s+")
  return new RegExp(`(^|[;&|]\\s*)${source}`, "i").test(command)
}

function pathPatternMatches(pattern, filePath) {
  const normalizedPath = normalizePath(filePath)
  const normalizedPattern = normalizePath(pattern)
  if (!normalizedPattern) return false

  if (normalizedPattern === ".env*") {
    const base = normalizedPath.split("/").pop() ?? ""
    return base === ".env" || base.startsWith(".env.")
  }

  const candidatePatterns = [normalizedPattern]
  if (normalizedPattern.startsWith("~/")) candidatePatterns.push(normalizedPattern.slice(2))

  for (const candidatePattern of candidatePatterns) {
    const basePattern = candidatePattern.endsWith("/**") ? candidatePattern.slice(0, -3) : ""
    if (basePattern && (normalizedPath === basePattern || normalizedPath.startsWith(`${basePattern}/`) || normalizedPath.includes(`/${basePattern}/`))) return true
    const regex = new RegExp(`^${wildcardToRegex(candidatePattern)}$`, "i")
    if (regex.test(normalizedPath)) return true
  }

  return false
}

function wildcardToRegex(pattern) {
  return String(pattern)
    .replace(/[.+^${}()|[\]\\]/g, "\\$&")
    .replace(/\*\*/g, ".*")
    .replace(/\*/g, "[^/]*")
}

function normalizePath(filePath) {
  return String(filePath ?? "").replaceAll("\\", "/").replace(/^\.\//, "")
}

function deny(message) {
  console.error(message)
  process.exit(2)
}

function ask(message) {
  // pi-yaml-hooks headless/non-interactive policy: request-like guardrails fail closed.
  console.error(message)
  process.exit(2)
}

function isRecord(value) {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

async function readStdin() {
  const chunks = []
  for await (const chunk of process.stdin) chunks.push(chunk)
  return Buffer.concat(chunks).toString("utf8")
}
