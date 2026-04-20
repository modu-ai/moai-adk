import type { MetaRecord } from "nextra";

/**
 * Workflow Commands - 4-Phase development cycle commands
 *
 * New format: /moai [subcommand]
 * Legacy format (/moai:X-YYYY) also supported
 */
const meta: MetaRecord = {
  index: { title: "Workflow Commands", display: "hidden" },
  "moai-project": "/moai project",
  "moai-plan": "/moai plan",
  "moai-run": "/moai run",
  "moai-sync": "/moai sync",
};

export default meta;
