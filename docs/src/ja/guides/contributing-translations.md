# 翻訳貢献ガイド

MoAI-ADKドキュメントの翻訳にご協力いただき、ありがとうございます!このガイドでは、翻訳作業を開始するための手順を説明します。

## 📊 現在の翻訳状況

現在の翻訳状況を確認するには、[翻訳状況ダッシュボード](../translation-status.md)をご覧ください。

## 🌍 サポート言語

現在サポートされている言語:

- **英語 (en)** - English
- **日本語 (ja)** - Japanese
- **中国語 (zh)** - Chinese (Simplified)

## 🚀 クイックスタート

### 1. 翻訳するファイルを選択

[翻訳状況ダッシュボード](../translation-status.md)で不足しているファイルのリストを確認してください。

翻訳が必要なファイルは、言語ごとに表示されています。

### 2. ファイル構造を理解

```
docs/src/
├── index.md                    # 韓国語 (デフォルト)
├── getting-started/
│   ├── installation.md
│   └── quick-start.md
├── en/                         # 英語翻訳
│   ├── getting-started/
│   │   ├── installation.md
│   │   └── quick-start.md
├── ja/                         # 日本語翻訳
│   └── ...
└── zh/                         # 中国語翻訳
    └── ...
```

**基本原則**:
- 韓国語原文: `docs/src/` のルートとサブディレクトリ
- 翻訳: `docs/src/{言語コード}/` 配下に同一のディレクトリ構造を維持

### 3. 翻訳作業を開始

#### 方法A: GitHub Web UIを使用

1. GitHubで翻訳するファイルを検索
2. 「Edit」ボタンをクリック
3. 翻訳を記述
4. 「Propose changes」をクリック
5. Pull Requestを作成

#### 方法B: ローカル環境を使用

```bash
# 1. リポジトリをフォークしてクローン
git clone https://github.com/YOUR_USERNAME/moai-adk.git
cd moai-adk

# 2. 翻訳ブランチを作成
git checkout -b translate-ja-getting-started

# 3. 翻訳ファイルを作成
# 例: docs/src/getting-started/installation.md を日本語に翻訳
mkdir -p docs/src/ja/getting-started
cp docs/src/getting-started/installation.md docs/src/ja/getting-started/installation.md

# 4. ファイルを翻訳 (エディタで開く)

# 5. 変更を確認
python docs/scripts/check_translation_status.py

# 6. コミットしてプッシュ
git add docs/src/ja/
git commit -m "docs: Add Japanese translation for installation guide"
git push origin translate-ja-getting-started

# 7. GitHubでPull Requestを作成
```

## 📝 翻訳ガイドライン

### 用語の一貫性

技術用語は可能な限り原語のまま使用するか、翻訳して原語を括弧内に追加してください。

| 韓国語 | 英語 | 日本語 | 中国語 |
|--------|---------|----------|---------|
| SPEC | SPEC | SPEC | SPEC |
| TAG | TAG | TAG | TAG |
| Alfred | Alfred | Alfred | Alfred |
| 테스트 주도 개발 | Test-Driven Development (TDD) | テスト駆動開発 (TDD) | 测试驱动开发 (TDD) |
| 요구사항 | Requirements | 要件 | 需求 |
| 구현 | Implementation | 実装 | 实现 |

### 文体

- **丁寧でプロフェッショナルな口調**を維持
- **二人称**を使用: 「you」(英語)、「당신」(韓国語)、「あなた」(日本語)、「您」(中国語)
- **明確で簡潔な表現**を使用

### コードブロック

コード例はそのまま翻訳せずに保持:

```python
# コードはそのまま保持 (コードブロック内のコメントは翻訳しない)
def hello_world():
    print("Hello, World!")
```

### リンクと参照

- **内部リンク**: 翻訳ページが存在する場合は翻訳言語のパスに変更
  ```markdown
  <!-- 韓国語 -->
  [설치 가이드](getting-started/installation.md)

  <!-- 日本語 -->
  [インストールガイド](../ja/getting-started/installation.md)
  ```

- **外部リンク**: 言語固有のバージョンがある場合は変更

## ✅ 品質チェックリスト

翻訳完了後、以下を確認してください:

- [ ] **ファイル構造**: 韓国語原文と同一のディレクトリ構造を維持
- [ ] **ファイル名**: 原文と同じファイル名を使用
- [ ] **Markdown構文**: 見出し、リンク、コードブロックなどにエラーがない
- [ ] **用語の一貫性**: 重要な用語が一貫して翻訳されている
- [ ] **コードの保持**: コード例がそのまま保持されている
- [ ] **リンクの検証**: 内部/外部リンクが正しく機能する
- [ ] **ローカルビルドテスト**: `mkdocs serve` でレンダリングを確認

## 🔍 翻訳のテスト

ローカルで翻訳結果を確認する方法:

```bash
# 1. ドキュメント依存関係をインストール
cd docs
pip install -r requirements.txt

# 2. MkDocs開発サーバーを起動
mkdocs serve

# 3. ブラウザで確認
# http://localhost:8000
```

## 🤝 レビュープロセス

1. **Pull Requestを作成**: 翻訳完了後にPRを提出
2. **自動検証**: CI/CDが構文とリンクを自動検証
3. **レビュー**: メンテナーまたは言語固有のレビュアーが確認
4. **修正依頼**: 必要に応じてフィードバックを適用
5. **マージ**: 承認後にmainブランチにマージ

## 📧 お問い合わせ

質問や支援が必要な場合:

- **GitHub Issues**: [moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
- **GitHub Discussions**: [moai-adk/discussions](https://github.com/modu-ai/moai-adk/discussions)

## 🎖️ 貢献者

翻訳への貢献者:

- 貢献者リストは[Contributors](https://github.com/modu-ai/moai-adk/graphs/contributors)でご覧いただけます。

---

**ありがとうございます!** あなたの貢献により、MoAI-ADKは世界中のより多くのユーザーに届けられます。🌏
