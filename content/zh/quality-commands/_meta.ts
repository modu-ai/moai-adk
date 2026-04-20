import type { MetaRecord } from "nextra";

/**
 * 质量命令 - 代码审查、覆盖率、E2E、架构分析
 *
 * 新格式：/moai [subcommand]
 * 也支持旧格式（/moai:XXXX）
 */
const meta: MetaRecord = {
  index: { title: "质量命令", display: "hidden" },
  "moai-review": "/moai review",
  "moai-coverage": "/moai coverage",
  "moai-e2e": "/moai e2e",
  "moai-codemaps": "/moai codemaps",
};

export default meta;
