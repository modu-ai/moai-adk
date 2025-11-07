# 扩展和自定义指南

如何根据项目需求自定义 MoAI-ADK。

## 🎯 可扩展区域

1. **Custom Skills**：添加新的领域技能
2. **Custom Agents**：创建专用代理
3. **Custom Hooks**：项目特定的自动化钩子
4. **Custom Commands**：扩展 Alfred 命令

## 🛠️ 创建 Custom Skills

### Skill 结构

```
.claude/skills/
└── custom-skill/
    ├── index.md           # Skill 文档
    ├── examples.md        # 使用示例
    ├── reference.md       # API 规范
    └── templates/         # 提示词模板
```

### Skill 编写示例

```markdown
# moai-custom-mlops

机器学习流水线构建和部署技能

## 用途
- ML 模型训练
- 模型评估和验证
- 生产环境部署

## 包含内容
- MLflow 集成
- Kubeflow 部署
- 模型服务模式
```

### Skill 注册

```bash
# 添加 Skill 元数据
# .moai/config.json:
{
  "custom_skills": {
    "moai-custom-mlops": {
      "version": "1.0",
      "author": "team",
      "enabled": true
    }
  }
}
```

## 👥 创建 Custom Agents

### Agent 结构

```
.claude/agents/
└── custom-agent/
    ├── agent.py          # Agent 实现
    ├── prompts.md        # 提示词
    └── tools.json        # 工具列表
```

### Agent 示例

```python
# .claude/agents/ml-expert/agent.py

class MLExpert:
    """机器学习专家代理"""

    def __init__(self):
        self.skills = [
            "moai-domain-ml",
            "moai-lang-python"
        ]

    def analyze_data(self, dataset):
        """数据分析"""
        # 分析逻辑
        pass

    def train_model(self, data, params):
        """模型训练"""
        # 训练逻辑
        pass

    def evaluate(self, model, test_data):
        """模型评估"""
        # 评估逻辑
        pass
```

### 启用 Agent

```bash
# 添加自定义代理
# .moai/config.json:
{
  "custom_agents": {
    "ml-expert": {
      "enabled": true,
      "activation_keywords": ["machine learning", "mlops", "model"]
    }
  }
}
```

## 🔧 创建 Custom Hooks

### Hook 结构

```bash
.claude/hooks/
├── custom_pre_tool.sh        # Pre-tool Hook
└── custom_post_tool.sh       # Post-tool Hook
```

### Hook 示例

```bash
#!/bin/bash
# .claude/hooks/custom_post_tool.sh

# 创建 Python 文件后自动格式化
if [[ "$TOOL_NAME" == "Write" && "$FILE_PATH" == *.py ]]; then
    black "$FILE_PATH"
    ruff check "$FILE_PATH"
fi

# Git 提交后自动推送
if [[ "$TOOL_NAME" == "Bash" && "$COMMAND" == *"git commit"* ]]; then
    echo "Auto-pushing after commit..."
    git push
fi
```

### Hook 注册

```json
{
  "hooks": {
    "custom_post_tool": ".claude/hooks/custom_post_tool.sh"
  }
}
```

## 📝 创建 Custom Commands

### Command 文件结构

```
.claude/commands/
└── custom-deploy.md
```

### Command 文件示例

```markdown
# /custom-deploy

部署自动化命令

此命令执行以下操作：
1. 执行构建
2. 执行测试
3. 生产环境部署

## 使用方法

/custom-deploy [environment] [version]

### 示例

/custom-deploy production v1.0.0
```

### Command 执行

```
用户: /custom-deploy production

Alfred:
1. 执行构建
2. 验证测试
3. 确认部署
4. 设置监控

完成！
```

## 🔄 集成点

### 与 Alfred 集成

```python
# Alfred 激活自定义代理
if "mlops" in spec.keywords:
    activate(ml_expert)  # Custom agent
    activate(backend_expert)  # Built-in agent
```

### 与 Skills 集成

```python
# 加载自定义 skill
Skill("moai-custom-mlops")
Skill("moai-domain-backend")
```

## 📈 扩展示例：CI/CD 自动化

### 目标

从开发到部署的自动化流水线

### 实现

```bash
# Custom hook: .claude/hooks/custom_post_tool.sh

# Git 提交后自动 CI/CD
if [[ "$COMMAND" == *"git commit"* ]]; then
    echo "🚀 启动 CI/CD 流水线..."

    # 1. 构建
    docker build -t app:latest .

    # 2. 测试
    docker run app:latest pytest

    # 3. 部署
    kubectl apply -f k8s/

    echo "✅ 部署完成"
fi
```

## 🎯 最佳实践

### 创建 Custom Skill 时

```
✅ 应该：
- 编写清晰文档
- 提供多样示例
- 设计为可与其他 Skill 组合
- 版本管理（语义化版本）

❌ 不应该：
- 修改 Alfred 核心逻辑
- 忽视安全漏洞
- 硬编码路径/值
- 无测试代码
```

### 创建 Custom Agent 时

```
✅ 应该：
- 专注于特定领域
- 设计为与现有代理协作
- 定义明确的激活条件
- 包含错误处理

❌ 不应该：
- 承担过多职责
- 与 Alfred 功能重复
- 创建循环引用
- 硬编码配置
```

______________________________________________________________________

**下一步**：[安全高级指南](security.md) 或 [性能优化](performance.md)
