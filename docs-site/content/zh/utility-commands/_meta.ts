import type { MetaRecord } from "nextra";

/**
 * 实用命令 - 自动化和反馈命令
 *
 * 新格式：/moai [subcommand]
 * 也支持旧格式（/moai:XXXX）
 */
const meta: MetaRecord = {
  index: { title: "实用命令", display: "hidden" },
  moai: "/moai",
  "moai-github": "/moai github",
  "moai-loop": "/moai loop",
  "moai-fix": "/moai fix",
  "moai-clean": "/moai clean",
  "moai-mx": "/moai mx",
  "moai-feedback": "/moai feedback",
};

export default meta;
