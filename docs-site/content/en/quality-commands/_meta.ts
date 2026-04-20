import type { MetaRecord } from "nextra";

/**
 * Quality Commands - Code review, coverage, E2E, architecture analysis
 *
 * New format: /moai [subcommand]
 * Legacy format (/moai:XXXX) also supported
 */
const meta: MetaRecord = {
  index: { title: "Quality Commands", display: "hidden" },
  "moai-review": "/moai review",
  "moai-coverage": "/moai coverage",
  "moai-e2e": "/moai e2e",
  "moai-codemaps": "/moai codemaps",
};

export default meta;
