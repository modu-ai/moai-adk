import type { MetaRecord } from "nextra";

/**
 * 工作流命令 - 4 阶段开发周期命令
 *
 * 新格式：/moai [subcommand]
 * 也支持旧格式（/moai:X-YYYY）
 */
const meta: MetaRecord = {
  index: { title: "工作流命令", display: "hidden" },
  "moai-project": "/moai project",
  "moai-plan": "/moai plan",
  "moai-run": "/moai run",
  "moai-sync": "/moai sync",
};

export default meta;
