---
title: 迁移指南
description: 将现有.agency/项目转换为新设计系统
weight: 60
draft: false
---

# 迁移指南

根据SPEC-AGENCY-ABSORB-001,`/agency`命令已合并到`/moai design`中。拥有现有`.agency/`目录的项目可以通过**迁移**转换到新系统。

## 何时迁移

如果以下任何条件适用,请迁移:

- `.agency/`目录存在
- 使用现有agency学习/观测
- 使用旧代理(agency-copywriter、agency-designer等)

**迁移后:**

- `.agency/` → `.agency.archived/`(备份)
- `.moai/project/brand/`(新建)
- `.moai/config/sections/design.yaml`(新建)
- 前learnings可合并

## 运行迁移

### 步骤1: 前检查

```bash
# 验证.agency/存在
ls -la .agency/
# brand-voice.md
# visual-identity.md
# learnings/
# observations/
```

### 步骤2: 干运行(可选)

预览迁移结果:

```bash
moai migrate agency --dry-run
```

### 步骤3: 执行迁移

```bash
moai migrate agency
```

**执行(6个阶段):**

1. **验证** — 检查`.agency/`存在、磁盘空间
2. **Staging** — 复制到临时目录
3. **上下文转移** — 复制brand文件到`.moai/project/brand/`
4. **配置合并** — 将learnings/observations合并到`.moai/config/`
5. **Learning转移** — 将启发式转换为新结构
6. **Atomic Swap** — 备份到`.agency.archived/`,完成

**成功时:**
```
迁移完成 [TX-abc123def456]

转移文件: 47个
  ✓ .moai/project/brand/ 3个
  ✓ .moai/config/sections/design.yaml已创建
  ✓ .moai/research/配置已合并

备份: .agency.archived/

后续:
  /moai design
```

## 迁移选项

### --force选项

覆盖现有目标目录:

```bash
moai migrate agency --force
```

**警告:** 覆盖`.moai/project/brand/`(如果存在)。事先备份。

### --resume选项

恢复中断的迁移:

```bash
# SIGINT后暂停迁移
moai migrate agency --resume TX-abc123def456
```

检查点文件: `~/.moai/.migrate-tx-<txID>.json`

## 迁移错误代码

| 错误代码 | 原因 | 解决方案 |
|---|---|---|
| `MIGRATE_NO_SOURCE` | `.agency/`缺失 | 检查现有agency目录 |
| `MIGRATE_TARGET_EXISTS` | `.moai/project/brand/`存在 | 使用`--force` |
| `MIGRATE_ARCHIVE_EXISTS` | `.agency.archived/`存在 | 删除或移动旧备份 |
| `MIGRATE_DISK_FULL` | 磁盘空间不足 | 释放空间(最少100MB) |
| `MIGRATE_MERGE_CONFLICT` | tech-preferences.md冲突 | 备份`.moai/project/tech.md`,重试 |
| `MIGRATE_INTERRUPT` | SIGINT/SIGTERM接收 | 使用`--resume`继续 |
| `MIGRATE_CHECKPOINT_CORRUPT` | 检查点文件损坏 | 删除`~/.moai/.migrate-tx-*.json`,重试 |

## 迁移结果

### 生成的文件结构

```
.moai/
├── project/
│   └── brand/
│       ├── brand-voice.md
│       ├── visual-identity.md
│       └── target-audience.md
├── config/
│   └── sections/
│       └── design.yaml
└── research/
    ├── learnings/
    └── observations/

.agency.archived/
├── brand-voice.md
├── visual-identity.md
├── learnings/
└── observations/
```

### Learning合并

现有`.agency/learnings/`项目:
- 转换为新结构
- 合并到`.moai/research/learnings/`
- 标记MIGRATED标签

## 回滚

迁移后恢复前状态:

### 选项1: 从备份还原

```bash
# 从备份还原
mv .agency.archived .agency

# 删除新建文件
rm -rf .moai/project/brand
rm .moai/config/sections/design.yaml
```

### 选项2: Git还原

如果迁移生成了git commit:

```bash
git log --oneline | grep migrate
# abc1234 chore: migrate agency to moai design system

git revert abc1234
```

## 迁移后

1. **验证品牌背景**
   ```bash
   cat .moai/project/brand/brand-voice.md
   ```

2. **启动新设计工作流**
   ```
   /moai design
   ```

3. **检查现有learnings**
   ```bash
   ls .moai/research/learnings/
   ```

4. **可选: 删除旧备份**
   ```bash
   rm -rf .agency.archived
   ```

## 检查迁移状态

查询迁移结果:

```bash
# 查看迁移日志
cat ~/.moai/.migrate-tx-abc123def456.json

# 或检查状态
moai status design
```

## SIGINT/SIGTERM处理

迁移中断时:

**按Ctrl+C:**
```
迁移中断 [TX-abc123def456]
已完成阶段: validation, staging, context-transfer
未完成阶段: config-merge, learning-transfer, atomic-swap

恢复:
  moai migrate agency --resume TX-abc123def456
```

## FAQ

### Q: 迁移后能使用/agency命令吗?

**A:** 不能。`/agency`不再支持。使用`/moai design`。

### Q: 能多次运行迁移吗?

**A:** 首次完成后`.agency/`变为`.agency.archived/`。第二次运行失败。使用`--force`覆盖。

### Q: Learnings会丢失吗?

**A:** 不会。所有learnings/observations合并到`.moai/research/`。备份也保存在`.agency.archived/`。

### Q: 迁移中网络断开?

**A:** 完成的阶段已保存。网络恢复后用`--resume`继续。
