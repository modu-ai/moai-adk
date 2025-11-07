# 参与贡献

感谢您对MoAI-ADK项目的关注。您的贡献将使项目更加强大。

## 贡献方式

### Bug报告

发现了Bug? 请通过GitHub Issues告诉我们:

1. **搜索**: 检查是否已有人报告该Bug
2. **详细信息**: 包括重现步骤、环境信息、截图
3. **最小化示例**: 提供展示问题的最小代码

### 功能建议

如果您有新功能的想法:

1. **问题定义**: 清楚说明要解决什么问题
2. **解决方案**: 分析提议解决方案的优缺点
3. **替代方案**: 考虑其他可能的解决方案

### 代码贡献

如果您想直接贡献代码:

1. **Fork**: Fork本仓库
2. **分支**: 按功能创建分支 (`git checkout -b feature/amazing-feature`)
3. **提交**: 提交更改 (`git commit -m 'Add amazing feature'`)
4. **推送**: 推送分支 (`git push origin feature/amazing-feature`)
5. **PR**: 创建Pull Request

## 开发环境设置

### 1. 克隆仓库

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### 2. 安装开发环境

```bash
# 使用UV (推荐)
uv sync

# 或使用pip
pip install -e ".[dev]"
```

### 3. 运行测试

```bash
pytest

# 包含覆盖率
pytest --cov=moai_adk
```

### 4. 代码风格检查

```bash
# 代码格式化
black .
ruff check .
ruff format .

# 类型检查
mypy .
```

## 文档贡献

改进文档也是重要的贡献:

- **修正错别字**: 修正发现的错别字或语法错误
- **翻译**: 将文档翻译成其他语言
- **添加示例**: 添加更多使用示例
- **改进说明**: 更清楚地解释复杂概念

## 编码标准

- **PEP 8**: 遵守Python风格指南
- **类型提示**: 所有函数包含类型提示
- **文档字符串**: 所有模块和函数包含文档字符串
- **测试**: 新功能必须包含测试代码

## 社区

- **GitHub Discussions**: 问题与讨论
- **Issues**: Bug报告与功能建议
- **Pull Requests**: 代码审查与协作

## 许可证

贡献的代码将在项目的MIT许可证下发布。
