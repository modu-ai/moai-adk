---
title: 快速入门
description: 5分钟内创建第一个 MoAI-ADK 项目并体验 AI 驱动的 TDD 开发流程
---

# 快速入门

本指南将帮助您在 5 分钟内创建第一个 MoAI-ADK 项目并体验完整的 AI 驱动开发流程。

## 前置要求

开始之前，请确保：

- <span class="material-icons">check_circle</span> 已安装 [MoAI-ADK](installation.md)
- <span class="material-icons">check_circle</span> 已安装 [Claude Code](installation.md#claude-code-设置)
- <span class="material-icons">check_circle</span> 有基本的 Python 和 Git 知识

---

## 5 分钟快速流程

### 步骤 <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span>：创建项目（1 分钟）

```bash
# 创建新项目
moai-adk init hello-world

# 进入项目目录
cd hello-world
```

**创建了什么？**

```
hello-world/
├── .moai/              <span class="material-icons">check_circle</span> Alfred 配置
├── .claude/            <span class="material-icons">check_circle</span> Claude Code 自动化
└── CLAUDE.md           <span class="material-icons">check_circle</span> 项目指南
```

### 步骤 <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span>：验证安装（30 秒）

```bash
# 运行系统诊断
moai-adk doctor
```

**预期输出**：

```
<span class="material-icons">check_circle</span> Python 3.13.0
<span class="material-icons">check_circle</span> uv 0.5.1
<span class="material-icons">check_circle</span> .moai/ directory initialized
<span class="material-icons">check_circle</span> .claude/ directory ready
<span class="material-icons">check_circle</span> 16 agents configured
<span class="material-icons">check_circle</span> 74 skills loaded
<span class="material-icons">check_circle</span> 5 hooks active
```

### 步骤 <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span>：启动 Claude Code（30 秒）

```bash
# 启动 Claude Code
claude
```

### 步骤 <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_4</span>：初始化项目（2 分钟）

在 Claude Code 中运行：

```
/alfred:0-project
```

Alfred 会询问几个问题：

```
Q1: 项目名称？
A: hello-world

Q2: 项目目标？
A: 学习 MoAI-ADK

Q3: 主要开发语言？
A: python

Q4: 模式？
A: personal
```

**完成后会看到**：

```
<span class="material-icons">check_circle</span> 项目初始化完成
<span class="material-icons">check_circle</span> 配置保存到 .moai/config.json
<span class="material-icons">check_circle</span> 在 .moai/project/ 中创建文档
<span class="material-icons">check_circle</span> Alfred 完成技能推荐

下一步: /alfred:1-plan "第一个功能说明"
```

### 步骤 <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_5</span>：创建第一个功能（1 分钟）

继续在 Claude Code 中：

```
/alfred:1-plan "Hello World API - GET /hello 端点返回问候语"
```

Alfred 会自动：
- 创建 SPEC 文档
- 分配 SPEC ID（如 HELLO-001）
- 生成功能分支

---

## 第一次实践：Hello World API

现在让我们完整体验 MoAI-ADK 的核心工作流程。

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_one</span> 规划：创建 SPEC（2 分钟）

```bash
/alfred:1-plan "创建 Hello World API，接收名字参数返回个性化问候语"
```

**Alfred 创建的内容**：

```
<span class="material-icons">check_circle</span> SPEC ID: HELLO-001
<span class="material-icons">check_circle</span> 文件: .moai/specs/SPEC-HELLO-001/spec.md
<span class="material-icons">check_circle</span> 分支: feature/SPEC-HELLO-001
```

**查看生成的 SPEC**：

```bash
cat .moai/specs/SPEC-HELLO-001/spec.md
```

内容示例：

```yaml
---
id: HELLO-001
version: 0.0.1
status: draft
priority: high
---
# `@SPEC:EX-HELLO-001: Hello World API

## Ubiquitous Requirements
- 系统必须提供 HTTP GET /hello 端点

## Event-driven Requirements
- 当提供查询参数 name 时，必须返回 "Hello, {name}!"
- 当没有 name 时，必须返回 "Hello, World!"

## Constraints
- name 必须限制在最多 50 字符
- 响应必须是 JSON 格式
```

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_two</span> 运行：TDD 实现（5 分钟）

```bash
/alfred:2-run HELLO-001
```

Alfred 会自动执行完整的 TDD 流程：

#### 🔴 RED 阶段：失败的测试

创建 `tests/test_hello.py`：

```python
# `@TEST:EX-HELLO-002 | SPEC: SPEC-HELLO-001.md

import pytest
from fastapi.testclient import TestClient
from src.hello.api import app

client = TestClient(app)

def test_hello_with_name_should_return_personalized_greeting():
    """当提供 name 时，必须返回 "Hello, {name}!" """
    response = client.get("/hello?name=张三")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, 张三!"}

def test_hello_without_name_should_return_default_greeting():
    """当没有 name 时，必须返回 "Hello, World!" """
    response = client.get("/hello")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, World!"}
```

运行测试（预期失败）：
```bash
pytest tests/test_hello.py -v
# <span class="material-icons">cancel</span> FAILED - No module named 'fastapi'
```

#### 🟢 GREEN 阶段：最小实现

创建 `src/hello/api.py`：

```python
# `@CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI

app = FastAPI()

@app.get("/hello")
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - Hello 端点"""
    return {"message": f"Hello, {name}!"}
```

安装依赖并运行测试：
```bash
uv add fastapi pytest
pytest tests/test_hello.py -v
# <span class="material-icons">check_circle</span> PASSED - 所有测试通过
```

#### <span class="material-icons">recycling</span> REFACTOR 阶段：代码改进

添加验证逻辑：
```python
from fastapi import FastAPI, HTTPException

app = FastAPI()

@app.get("/hello")
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - 带验证的 Hello 端点"""
    if len(name) > 50:
        raise HTTPException(status_code=400, detail="Name too long (max 50 chars)")
    return {"message": f"Hello, {name}!"}
```

添加边界测试：
```python
def test_hello_with_long_name_should_return_400():
    """name 超过 50 字符时必须返回 400 错误"""
    long_name = "a" * 51
    response = client.get(f"/hello?name={long_name}")
    assert response.status_code == 400
    assert response.json()["detail"] == "Name too long (max 50 chars)"
```

最终测试验证：
```bash
pytest tests/test_hello.py -v
# <span class="material-icons">check_circle</span> PASSED - 所有测试通过，包括边界测试
```

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_3</span> 同步：文档自动生成（1 分钟）

```bash
/alfred:3-sync
```

**Alfred 自动完成**：

```
<span class="material-icons">check_circle</span> docs/api/hello.md - API 文档生成
<span class="material-icons">check_circle</span> README.md - API 使用方法添加
<span class="material-icons">check_circle</span> CHANGELOG.md - v0.1.0 发布说明添加
<span class="material-icons">check_circle</span> TAG 链验证 - 所有 @TAG 确认
```

**查看生成的 API 文档**：

```bash
cat docs/api/hello.md
```

### <span class="material-icons" style="font-size: 1em; vertical-align: middle;">looks_4</span> 验证：TAG 链追踪（1 分钟）

检查完整的 TAG 链：
```bash
rg '@(SPEC|TEST|CODE|DOC):HELLO-001' -n
```

**输出**：
```
.moai/specs/SPEC-HELLO-001/spec.md:7:# `@SPEC:EX-HELLO-001: Hello World API
tests/test_hello.py:3:# `@TEST:EX-HELLO-002 | SPEC: SPEC-HELLO-001.md
src/hello/api.py:3:# `@CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md
docs/api/hello.md:24:- `@SPEC:EX-HELLO-001
```

**意义**：需求 → 测试 → 实现 → 文档完美连接！

---

## 🎉 5 分钟后您获得了什么？

### 完整的项目结构

```
hello-world/
├── .moai/
│   ├── specs/SPEC-HELLO-001/
│   │   ├── spec.md              ← 需求文档
│   │   └── plan.md              ← 实现计划
│   ├── project/
│   │   ├── product.md           ← 产品愿景
│   │   ├── structure.md         ← 项目结构
│   │   └── tech.md              ← 技术栈
│   └── config.json              ← 项目配置
├── .claude/
│   ├── agents/                  ← 16个 AI 代理
│   ├── commands/                ← Alfred 命令
│   ├── skills/                  ← 74个专业技能
│   └── hooks/                   ← 自动化钩子
├── tests/
│   └── test_hello.py            ← 测试代码（100% 覆盖率）
├── src/
│   └── hello/
│       ├── api.py               ← API 实现
│       └── __init__.py
├── docs/
│   └── api/
│       └── hello.md             ← 自动生成的 API 文档
├── README.md                    ← 自动更新的项目说明
├── CHANGELOG.md                 ← 版本变更记录
├── CLAUDE.md                    ← Alfred 工作指南
└── pyproject.toml               ← Python 项目配置
```

### Git 历史记录

```bash
git log --oneline
```

输出：
```
a1b2c3d <span class="material-icons">check_circle</span> sync(HELLO-001): update docs and changelog
b2c3d4e <span class="material-icons">recycling</span> refactor(HELLO-001): add name length validation
c3d4e5f 🟢 feat(HELLO-001): implement hello API
d4e5f6g 🔴 test(HELLO-001): add failing hello API tests
e5f6g7h 🌿 Create feature/SPEC-HELLO-001 branch
f6g7h8i 📋 Initial project setup
```

### 核心体验

- <span class="material-icons">check_circle</span> **SPEC-First**：用 EARS 格式明确定义需求
- <span class="material-icons">check_circle</span> **TDD 流程**：RED → GREEN → REFACTOR 完整体验
- <span class="material-icons">check_circle</span> **自动化**：文档与代码同步生成
- <span class="material-icons">check_circle</span> **可追踪性**：@TAG 系统连接所有开发产物
- <span class="material-icons">check_circle</span> **质量保证**：测试覆盖率 100%，代码质量验证

---

## 运行您的第一个 API

### 启动服务器

```bash
# 安装 FastAPI（如果还没有）
uv add fastapi uvicorn

# 启动开发服务器
uvicorn src.hello.api:app --reload
```

### 测试 API

```bash
# 测试默认问候
curl http://localhost:8000/hello
# 返回: {"message": "Hello, World!"}

# 测试个性化问候
curl "http://localhost:8000/hello?name=张三"
# 返回: {"message": "Hello, 张三!"}

# 测试边界情况
curl "http://localhost:8000/hello?name=$(python -c 'print("a"*51)')"
# 返回: {"detail": "Name too long (max 50 chars)"}
```

### 查看自动文档

```bash
# 浏览器访问
open http://localhost:8000/docs
```

您将看到 FastAPI 自动生成的交互式 API 文档！

---

## 下一步

### 继续学习

1. **深入核心概念**：阅读 [核心概念指南](../guides/concepts.md)
2. **探索 Alfred 命令**：学习 [Alfred 命令详解](../guides/alfred/)
3. **理解 TDD 流程**：查看 [TDD 指南](../guides/tdd/)
4. **掌握 SPEC 编写**：阅读 [SPEC 指南](../guides/specs/)

### 实践建议

1. **创建下一个功能**：
   ```bash
   /alfred:1-plan "用户管理 API - 注册、登录、个人信息"
   ```

2. **尝试不同项目类型**：
   ```bash
   # Web 应用
   moai-adk init my-webapp --template web

   # CLI 工具
   moai-adk init my-cli --template cli

   # 数据分析项目
   moai-adk init my-analysis --template data
   ```

3. **探索高级功能**：
   - [多语言支持](../advanced/i18n.md)
   - [性能优化](../advanced/performance.md)
   - [安全最佳实践](../advanced/security.md)

### 加入社区

- **GitHub**: [moai-adk](https://github.com/modu-ai/moai-adk)
- **讨论**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **问题反馈**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)

---

## 常见问题

### Q: 可以在现有项目中使用 MoAI-ADK 吗？

A: 可以！在现有项目目录中运行：
```bash
moai-adk init .
```

### Q: 如何切换到不同的编程语言？

A: 在 `/alfred:0-project` 时选择相应的语言，或手动修改 `.moai/config.json`：
```json
{
  "language": {
    "primary": "typescript"
  }
}
```

### Q: 团队协作时如何保持一致性？

A: 所有团队成员使用相同的 `/alfred` 命令，TAG 系统确保代码追踪性。参考 [团队协作指南](../guides/team-collaboration.md)。

### Q: 如何自定义 Alfred 的行为？

A: 编辑 `.claude/agents/` 和 `.claude/skills/` 目录中的文件。详见 [自定义指南](../advanced/customization.md)。

---

**恭喜！您已经成功完成了 MoAI-ADK 的快速入门。现在您拥有了一个完整的、文档化的、测试覆盖的 API 项目，体验了 AI 驱动的现代化开发流程。**

继续探索，发现更多可能！<span class="material-icons">rocket_launch</span>