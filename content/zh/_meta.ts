import type { MetaRecord } from "nextra";

/**
 * MoAI-ADK 文档 - 6 个部分结构
 *
 * 1. 开始使用 - 安装、基本设置、快速入门
 * 2. 核心概念 - 什么是 MoAI-ADK、SPEC、DDD、TRUST 5
 * 3. 工作流命令 - /moai project ~ /moai sync
 * 4. 实用命令 - /moai、loop、fix、feedback
 * 5. 高级 - 技能、代理、构建器、钩子、设置
 * 6. Git 工作树 - 完整的工作树 CLI 指南
 */
const meta: MetaRecord = {
  "getting-started": "开始使用",
  "core-concepts": "核心概念",
  "workflow-commands": "工作流命令",
  "utility-commands": "实用命令",
  "quality-commands": "质量命令",
  agency: "Agency",
  advanced: "高级",
  worktree: "Git 工作树",
};

export default meta;
