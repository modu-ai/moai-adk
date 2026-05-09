export const EXTENSION_ID = "moai-claude-compat";
export const STATUS_ID = "moai-status";
export const WIDGET_ID = "moai-dashboard";

export const PI_ROOT_PATH = ".pi";
export const SOURCE_MAP_PATH = ".pi/claude-compat/source-map.json";
export const TOOL_ALIASES_PATH = ".pi/claude-compat/tool-aliases.json";
export const OUTPUT_STYLES_CONFIG_PATH = ".pi/claude-compat/output-styles.json";
export const HOOK_EVENTS_CONFIG_PATH = ".pi/claude-compat/hook-events.json";
export const CLAUDE_RULES_SOURCE_PATH = ".claude/rules/moai";
export const EXPECTED_MOAI_RULE_COUNT = 46;
export const PI_SOURCE_ROOT = ".pi/generated/source";
export const PI_SKILLS_SOURCE_PATH = `${PI_SOURCE_ROOT}/skills`;
export const PI_COMMANDS_SOURCE_PATH = `${PI_SOURCE_ROOT}/commands`;
export const PI_AGENTS_SOURCE_PATH = `${PI_SOURCE_ROOT}/agents/moai`;
export const PI_RULES_SOURCE_PATH = `${PI_SOURCE_ROOT}/rules/moai`;
export const PI_HOOKS_SOURCE_PATH = `${PI_SOURCE_ROOT}/hooks/moai`;
export const PI_OUTPUT_STYLES_SOURCE_PATH = `${PI_SOURCE_ROOT}/output-styles/moai`;
export const PI_CLAUDE_MD_SOURCE_PATH = `${PI_SOURCE_ROOT}/CLAUDE.md`;
export const PI_SETTINGS_SOURCE_PATH = `${PI_SOURCE_ROOT}/settings/claude-settings.json`;
export const PI_MOAI_CONFIG_SOURCE_PATH = `${PI_SOURCE_ROOT}/moai-config/sections`;
export const LANGUAGE_CONFIG_PATH = `${PI_MOAI_CONFIG_SOURCE_PATH}/language.yaml`;
export const QUALITY_CONFIG_PATH = `${PI_MOAI_CONFIG_SOURCE_PATH}/quality.yaml`;
export const WORKFLOW_CONFIG_PATH = `${PI_MOAI_CONFIG_SOURCE_PATH}/workflow.yaml`;

export const TEAM_BACKEND_PRIORITY = [
  "@tmustier/pi-agent-teams",
  "pi-teams",
  "pi-crew",
] as const;

export const QUOTA_FOOTER_PRIORITY = [
  "moai-claude-compat-native-codex-quota",
  "pi-chatgpt-limit",
  "pi-usage",
] as const;
