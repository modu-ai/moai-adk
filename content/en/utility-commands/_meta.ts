import type { MetaRecord } from "nextra";

/**
 * Utility Commands - Automation and feedback commands
 *
 * New format: /moai [subcommand]
 * Legacy format (/moai:XXXX) also supported
 */
const meta: MetaRecord = {
  index: { title: "Utility Commands", display: "hidden" },
  moai: "/moai",
  "moai-github": "/moai github",
  "moai-loop": "/moai loop",
  "moai-fix": "/moai fix",
  "moai-clean": "/moai clean",
  "moai-mx": "/moai mx",
  "moai-feedback": "/moai feedback",
};

export default meta;
