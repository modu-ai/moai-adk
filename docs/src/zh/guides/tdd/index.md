---
title: TDD 开发指南
description: 测试驱动开发完整指南 - 通过 RED、GREEN、REFACTOR 循环编写稳定代码
status: stable
---

# TDD (Test-Driven Development) 开发指南

**TDD (测试驱动开发)** 是 MoAI-ADK 的核心原则。本指南将教您如何通过 RED-GREEN-REFACTOR 循环实现测试优先开发。

## 📚 什么是 TDD？

测试驱动开发按以下顺序进行：

1. **RED**：编写失败的测试
2. **GREEN**：编写最小代码使测试通过
3. **REFACTOR**：改进代码质量

通过重复这个循环，编写满足需求的稳定代码。

## 🎯 各阶段指南

### [RED 阶段](red.md)
- 编写失败的测试
- 测试用例设计
- 边界值和异常处理

### [GREEN 阶段](green.md)
- 最小实现（YAGNI 原则）
- 快速通过测试
- 性能与功能平衡

### [REFACTOR 阶段](refactor.md)
- 代码整理和优化
- 应用 SOLID 原则
- 提升可读性

## 🔄 与 Alfred 一起使用 TDD

Alfred SuperAgent 自动化 TDD 循环：

- `/alfred:2-run SPEC-ID`：自动执行 RED-GREEN-REFACTOR
- 每个阶段自动验证
- Git 提交自动化

[使用 Alfred 工作流程开始 TDD](../alfred/2-run.md)

## 📊 TDD 的优势

| 项目 | 效果 |
|------|------|
| **测试覆盖率** | 自动达到 87%+ |
| **早期发现缺陷** | 开发中检测到 95% 以上 |
| **重构安全性** | 通过测试提供完美保护 |
| **文档化** | 测试本身就是可执行文档 |
| **设计改进** | 自动形成可测试设计 |

## 🚀 下一步

- [RED：编写失败的测试](red.md)
- [GREEN：最小实现通过](green.md)
- [REFACTOR：改进代码](refactor.md)
- [Alfred 2-run 工作流程](../alfred/2-run.md)

---

**了解更多**：MoAI-ADK 的 TDD 原则是 SPEC-First 开发哲学的核心。定义 SPEC 后使用 TDD 实现，就能完成完全满足需求的代码。
