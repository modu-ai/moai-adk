# 拡張とカスタマイズガイド

MoAI-ADKをプロジェクトに合わせてカスタマイズする方法。

## 拡張可能な領域

1. **Custom Skills**: 新しいドメインスキルの追加
2. **Custom Agents**: 専門化されたエージェントの作成
3. **Custom Hooks**: プロジェクト固有の自動化Hook
4. **Custom Commands**: Alfredコマンドの拡張

## Custom Skillsの作成

### Skill構造

```
.claude/skills/
└── custom-skill/
    ├── index.md           # Skillドキュメント
    ├── examples.md        # 使用例
    ├── reference.md       # API仕様
    └── templates/         # プロンプトテンプレート
```

### Skill作成例

```markdown
# moai-custom-mlops

機械学習パイプライン構築およびデプロイスキル

## 用途
- MLモデル学習
- モデル評価と検証
- プロダクションデプロイ

## 内容
- MLflow統合
- Kubeflowデプロイ
- モデルサービングパターン
```

### Skill登録

```bash
# Skillメタデータ追加
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

## Custom Agentsの作成

### Agent構造

```
.claude/agents/
└── custom-agent/
    ├── agent.py          # Agent実装
    ├── prompts.md        # プロンプト
    └── tools.json        # ツールリスト
```

### Agent例

```python
# .claude/agents/ml-expert/agent.py

class MLExpert:
    """機械学習専門家エージェント"""

    def __init__(self):
        self.skills = [
            "moai-domain-ml",
            "moai-lang-python"
        ]

    def analyze_data(self, dataset):
        """データ分析"""
        # 分析ロジック
        pass

    def train_model(self, data, params):
        """モデル学習"""
        # 学習ロジック
        pass

    def evaluate(self, model, test_data):
        """モデル評価"""
        # 評価ロジック
        pass
```

### Agent有効化

```bash
# カスタムエージェント追加
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

## Custom Hooksの作成

### Hook構造

```bash
.claude/hooks/
├── custom_pre_tool.sh        # Pre-tool Hook
└── custom_post_tool.sh       # Post-tool Hook
```

### Hook例

```bash
#!/bin/bash
# .claude/hooks/custom_post_tool.sh

# すべてのPythonファイル作成後自動フォーマット
if [[ "$TOOL_NAME" == "Write" && "$FILE_PATH" == *.py ]]; then
    black "$FILE_PATH"
    ruff check "$FILE_PATH"
fi

# Gitコミット後自動プッシュ
if [[ "$TOOL_NAME" == "Bash" && "$COMMAND" == *"git commit"* ]]; then
    echo "コミット後自動プッシュ..."
    git push
fi
```

### Hook登録

```json
{
  "hooks": {
    "custom_post_tool": ".claude/hooks/custom_post_tool.sh"
  }
}
```

## Custom Commandsの作成

### Commandファイル構造

```
.claude/commands/
└── custom-deploy.md
```

### Commandファイル例

```markdown
# /custom-deploy

デプロイ自動化コマンド

このコマンドは以下を実行します：
1. ビルド実行
2. テスト実行
3. プロダクションデプロイ

## 使用法

/custom-deploy [environment] [version]

### 例

/custom-deploy production v1.0.0
```

### Command実行

```
ユーザー: /custom-deploy production

Alfred:
1. ビルド実行
2. テスト検証
3. デプロイ確認
4. モニタリング設定

完了！
```

## 統合ポイント

### Alfredとの統合

```python
# Alfredがカスタムエージェントを有効化
if "mlops" in spec.keywords:
    activate(ml_expert)  # カスタムエージェント
    activate(backend_expert)  # 組み込みエージェント
```

### Skillsとの統合

```python
# カスタムスキルをロード
Skill("moai-custom-mlops")
Skill("moai-domain-backend")
```

## 拡張例: CI/CD自動化

### 目標

開発からデプロイまでの自動化パイプライン

### 実装

```bash
# カスタムHook: .claude/hooks/custom_post_tool.sh

# Gitコミット後自動CI/CD
if [[ "$COMMAND" == *"git commit"* ]]; then
    echo "🚀 CI/CDパイプライン開始..."

    # 1. ビルド
    docker build -t app:latest .

    # 2. テスト
    docker run app:latest pytest

    # 3. デプロイ
    kubectl apply -f k8s/

    echo "✅ デプロイ完了"
fi
```

## ベストプラクティス

### Custom Skill作成時

```
✅ DO:
- 明確なドキュメント作成
- 多様な例の提供
- 他のSkillと組み合わせ可能な設計
- バージョン管理（semantic versioning）

❌ DON'T:
- Alfredのコアロジックの変更
- セキュリティ脆弱性の無視
- ハードコードされたパス/値
- テストのないコード
```

### Custom Agent作成時

```
✅ DO:
- 特定のドメインに集中
- 既存のエージェントと協力する設計
- 明確な有効化条件の定義
- エラー処理の包含

❌ DON'T:
- 責任を負いすぎる
- Alfredと重複する機能
- 循環参照の作成
- ハードコードされた設定
```

______________________________________________________________________

**次**: [セキュリティ上級ガイド](security.md)または[パフォーマンス最適化](performance.md)



