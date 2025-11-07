# 翻译贡献指南

感谢您为MoAI-ADK文档翻译做出贡献！本指南将指导您如何开始翻译工作。

## 📊 当前翻译状态

要查看翻译进度，请参考[翻译状态面板](../translation-status.md)。

## 🌍 支持的语言

当前支持的语言:

- **English (en)** - 英语
- **Japanese (ja)** - 日语
- **Chinese (zh)** - 中文 (简体)

## 🚀 快速入门

### 1. 选择要翻译的文件

查看[翻译状态面板](../translation-status.md)中的缺失文件列表。

会显示每种语言需要翻译的文件。

### 2. 理解文件结构

```
docs/src/
├── index.md                    # 韩语 (默认)
├── getting-started/
│   ├── installation.md
│   └── quick-start.md
├── en/                         # 英语翻译
│   ├── getting-started/
│   │   ├── installation.md
│   │   └── quick-start.md
├── ja/                         # 日语翻译
│   └── ...
└── zh/                         # 中文翻译
    └── ...
```

**核心原则**:
- 韩语原文: `docs/src/` 根目录及子目录
- 翻译版本: `docs/src/{语言代码}/` 下保持相同的目录结构

### 3. 开始翻译工作

#### 方法A: 使用GitHub Web UI

1. 在GitHub中找到要翻译的文件
2. 点击 "Edit" 按钮
3. 编写翻译内容
4. 点击 "Propose changes"
5. 创建Pull Request

#### 方法B: 使用本地环境

```bash
# 1. Fork并clone仓库
git clone https://github.com/YOUR_USERNAME/moai-adk.git
cd moai-adk

# 2. 创建翻译分支
git checkout -b translate-ja-getting-started

# 3. 创建翻译文件
# 例如: 将docs/src/getting-started/installation.md翻译成日语
mkdir -p docs/src/ja/getting-started
cp docs/src/getting-started/installation.md docs/src/ja/getting-started/installation.md

# 4. 翻译文件 (用编辑器打开并工作)

# 5. 确认更改
python docs/scripts/check_translation_status.py

# 6. Commit并Push
git add docs/src/ja/
git commit -m "docs: Add Japanese translation for installation guide"
git push origin translate-ja-getting-started

# 7. 在GitHub上创建Pull Request
```

## 📝 翻译指南

### 术语统一

主要技术术语尽可能保持原文，必要时翻译后在括号中标注原文。

| 韩语 | English | Japanese | Chinese |
|--------|---------|----------|---------|
| SPEC | SPEC | SPEC | SPEC |
| TAG | TAG | TAG | TAG |
| Alfred | Alfred | Alfred | Alfred |
| 테스트 주도 개발 | Test-Driven Development (TDD) | テスト駆動開発 (TDD) | 测试驱动开发 (TDD) |
| 요구사항 | Requirements | 要件 | 需求 |
| 구현 | Implementation | 実装 | 实现 |

### 文体

- 保持**礼貌和专业的语气**
- **使用第二人称**: "当신"(韩语), "you"(英语), "あなた"(日语), "您"(中文)
- 使用**清晰简洁的表达**

### 代码块

代码示例不翻译，保持原样:

```python
# Keep code as-is (do not translate comments in code blocks)
def hello_world():
    print("Hello, World!")
```

### 链接与引用

- **内部链接**: 如有翻译页面，更改为相应语言的路径
  ```markdown
  <!-- Korean -->
  [설치 가이드](getting-started/installation.md)

  <!-- English -->
  [Installation Guide](../en/getting-started/installation.md)
  ```

- **外部链接**: 如可能，更改为相应语言版本的链接

## ✅ 质量检查清单

翻译完成后请检查以下事项:

- [ ] **文件结构**: 保持与韩语原文相同的目录结构
- [ ] **文件名**: 使用与原文相同的文件名
- [ ] **Markdown语法**: 标题、链接、代码块等语法无误
- [ ] **术语统一**: 主要术语保持一致翻译
- [ ] **代码保持**: 代码示例保持原样
- [ ] **链接验证**: 内部/外部链接正常工作
- [ ] **本地构建测试**: 使用`mkdocs serve`确认渲染

## 🔍 测试您的翻译

在本地确认翻译结果:

```bash
# 1. 安装Documentation依赖
cd docs
pip install -r requirements.txt

# 2. 运行MkDocs开发服务器
mkdocs serve

# 3. 在浏览器中确认
# http://localhost:8000
```

## 🤝 审查流程

1. **创建PR**: 翻译完成后提交PR
2. **自动验证**: CI/CD自动执行语法和链接验证
3. **审查**: 维护者或语言审查者进行审查
4. **修改请求**: 必要时反映反馈
5. **合并**: 批准后合并到main分支

## 📧 联系方式

如有问题或需要帮助:

- **GitHub Issues**: [moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
- **GitHub Discussions**: [moai-adk/discussions](https://github.com/modu-ai/moai-adk/discussions)

## 🎖️ 贡献者

感谢为翻译做出贡献的人们:

- 贡献者列表可在[Contributors](https://github.com/modu-ai/moai-adk/graphs/contributors)查看。

---

**谢谢!** 您的贡献让MoAI-ADK能够惠及更多用户。🌏
