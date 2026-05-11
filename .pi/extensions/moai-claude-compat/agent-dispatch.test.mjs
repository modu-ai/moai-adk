import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { loadRuntimeManifest } from "./src/runtime-config.ts";
import { resolveWorkflowAgentDispatch } from "./src/agent-dispatch.ts";

function setupProject(agents = []) {
  const cwd = mkdtempSync(join(tmpdir(), "moai-agent-dispatch-"));
  mkdirSync(join(cwd, ".pi/agents/moai"), { recursive: true });
  mkdirSync(join(cwd, ".pi/generated/source/agents/moai"), { recursive: true });
  for (const agent of agents) {
    writeFileSync(join(cwd, ".pi/agents/moai", `${agent}.md`), `---\nname: ${agent}\n---\n# ${agent}\n`, "utf8");
  }
  return cwd;
}

function writeWorkflowMap(cwd, workflows) {
  mkdirSync(join(cwd, ".pi/claude-compat"), { recursive: true });
  writeFileSync(join(cwd, ".pi/claude-compat/workflow-map.json"), `${JSON.stringify({
    version: 1,
    workflows,
    teamBackendPriority: ["@tmustier/pi-agent-teams"],
    tddDdd: "test",
  }, null, 2)}\n`, "utf8");
}

function runtime(cwd) {
  return loadRuntimeManifest(cwd, { useCache: false });
}

{
  const cwd = setupProject(["manager-brain"]);
  const dispatch = resolveWorkflowAgentDispatch("brain", "new idea", runtime(cwd));
  assert.equal(dispatch.kind, "subagent");
  assert.equal(dispatch.agent, "manager-brain");
  assert.match(dispatch.prompt, /"agent": "manager-brain"/);
  assert.match(dispatch.prompt, /"context": "fork"/);
  assert.match(dispatch.prompt, /"agentScope": "project"/);
  assert.match(dispatch.task, /read the workflow source/i);
  assert.match(dispatch.task, /ORCHESTRATION_BLOCKER/);
}

{
  const cwd = setupProject(["manager-quality"]);
  const feedback = resolveWorkflowAgentDispatch("feedback", "bug report", runtime(cwd));
  assert.equal(feedback.kind, "subagent");
  assert.equal(feedback.agent, "manager-quality");

  const review = resolveWorkflowAgentDispatch("review", "--staged", runtime(cwd));
  assert.equal(review.kind, "subagent");
  assert.equal(review.agent, "manager-quality");

  for (const flag of ["--security", "--design", "--critique"]) {
    const specialistReview = resolveWorkflowAgentDispatch("review", flag, runtime(cwd));
    assert.equal(specialistReview.kind, "fallback");
    assert.equal(specialistReview.dispatchMode, "skill");
    assert.match(specialistReview.reason, /specialist flag requested/);
  }

  const teamReview = resolveWorkflowAgentDispatch("review", "--team", runtime(cwd));
  assert.equal(teamReview.kind, "fallback");
  assert.equal(teamReview.dispatchMode, "skill");
  assert.match(teamReview.reason, /team mode requested/);

  const spacedModeTeamReview = resolveWorkflowAgentDispatch("review", "--mode team", runtime(cwd));
  assert.equal(spacedModeTeamReview.kind, "fallback");
  assert.equal(spacedModeTeamReview.dispatchMode, "skill");
  assert.match(spacedModeTeamReview.reason, /team mode requested/);

  const equalsModeTeamReview = resolveWorkflowAgentDispatch("review", "--mode=team", runtime(cwd));
  assert.equal(equalsModeTeamReview.kind, "fallback");
  assert.equal(equalsModeTeamReview.dispatchMode, "skill");
  assert.match(equalsModeTeamReview.reason, /team mode requested/);
}

{
  const cwd = setupProject(["manager-spec", "manager-tdd", "manager-docs", "manager-project"]);
  for (const command of ["plan", "run", "sync", "project"]) {
    const dispatch = resolveWorkflowAgentDispatch(command, "args", runtime(cwd));
    assert.equal(dispatch.kind, "fallback");
    assert.equal(dispatch.dispatchMode, "skill");
  }
}

{
  const cwd = setupProject([]);
  const dispatch = resolveWorkflowAgentDispatch("gate", "--fix", runtime(cwd));
  assert.equal(dispatch.kind, "fallback");
  assert.equal(dispatch.dispatchMode, "direct");
}

{
  const cwd = setupProject([]);
  const dispatch = resolveWorkflowAgentDispatch("brain", "missing agent", runtime(cwd));
  assert.equal(dispatch.kind, "fallback");
  assert.equal(dispatch.selectedAgent, "manager-brain");
  assert.match(dispatch.reason, /missing/);
}

{
  const cwd = setupProject(["manager-tdd", "manager-ddd"]);
  writeWorkflowMap(cwd, {
    run: {
      source: "./generated/source/skills/moai/workflows/run.md",
      modeSource: "./generated/source/moai-config/sections/quality.yaml",
      dispatchMode: "subagent",
      primaryAgents: ["manager-tdd", "manager-ddd"],
    },
  });
  const tdd = resolveWorkflowAgentDispatch("run", "SPEC-TEST-001", runtime(cwd));
  assert.equal(tdd.kind, "subagent");
  assert.equal(tdd.agent, "manager-tdd");

  mkdirSync(join(cwd, ".moai/config/sections"), { recursive: true });
  writeFileSync(join(cwd, ".moai/config/sections/quality.yaml"), "constitution:\n  development_mode: ddd\n", "utf8");
  const ddd = resolveWorkflowAgentDispatch("run", "SPEC-TEST-002", runtime(cwd));
  assert.equal(ddd.kind, "subagent");
  assert.equal(ddd.agent, "manager-ddd");
}

{
  const testDir = dirname(fileURLToPath(import.meta.url));
  const repoRoot = resolve(testDir, "../../..");
  const manifest = JSON.parse(readFileSync(join(repoRoot, ".pi/claude-compat/workflow-map.json"), "utf8"));
  const realRuntime = loadRuntimeManifest(repoRoot, { useCache: false });
  assert.equal(manifest.workflows.brain.dispatchMode, "subagent");
  assert.equal(realRuntime.workflowMap.config.workflows.brain.dispatchMode, "subagent");
  assert.equal(manifest.workflows.gate.dispatchMode, "direct");
  assert.equal(realRuntime.workflowMap.config.workflows.gate.dispatchMode, "direct");
  for (const command of ["plan", "run", "sync", "project"]) {
    assert.equal(manifest.workflows[command].dispatchMode, "skill");
    assert.equal(realRuntime.workflowMap.config.workflows[command].dispatchMode, "skill");
  }
}

console.log("agent-dispatch tests passed");
