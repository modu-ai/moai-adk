import type { MetaRecord } from "nextra";

/**
 * 워크플로우 명령어 - 4단계 개발 사이클 명령어
 *
 * 새로운 형식: /moai [subcommand]
 * 레거시 형식(/moai:X-YYYY)도 호환됨
 */
const meta: MetaRecord = {
  index: { title: "워크플로우 명령어", display: "hidden" },
  "moai-project": "/moai project",
  "moai-plan": "/moai plan",
  "moai-run": "/moai run",
  "moai-sync": "/moai sync",
};

export default meta;
