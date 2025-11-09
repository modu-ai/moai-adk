# トラブルシューティングガイド

MoAI-ADK使用中に発生する問題の解決方法です。

## 問題別ソリューション検索

### インストールおよび初期化問題

- [インストールエラー](https://adk.mo.ai.kr/troubleshooting/installation)
- [初期化失敗](https://adk.mo.ai.kr/troubleshooting/initialization)
- [環境設定](https://adk.mo.ai.kr/troubleshooting/environment)

### Alfredコマンド問題

- [コマンド認識不可](https://adk.mo.ai.kr/troubleshooting/command-not-found)
- [SPEC作成失敗](https://adk.mo.ai.kr/troubleshooting/spec-creation)
- [TDDサイクルエラー](https://adk.mo.ai.kr/troubleshooting/tdd-errors)

### 開発およびビルド問題

- [テスト失敗](https://adk.mo.ai.kr/troubleshooting/test-failures)
- [依存関係エラー](https://adk.mo.ai.kr/troubleshooting/dependency-errors)
- [ビルドエラー](https://adk.mo.ai.kr/troubleshooting/build-errors)

### Gitおよびデプロイ問題

- [Git競合](https://adk.mo.ai.kr/troubleshooting/git-conflicts)
- [デプロイ失敗](https://adk.mo.ai.kr/troubleshooting/deployment-errors)
- [CI/CD問題](https://adk.mo.ai.kr/troubleshooting/cicd-issues)

______________________________________________________________________

## よくある質問（FAQ）

### 基本使用方法

**Q: MoAI-ADKを初めて始めるにはどうすればよいですか？** A: [クイックスタートガイド](../getting-started/quick-start.md)を参照してください。3分で基本設定を完了できます。

**Q: SPEC-Firstとは何ですか？** A: [基本概念](../getting-started/concepts.md)で詳しく説明しています。簡単に言うと、コードを書く前に仕様書を先に書く方法です。

**Q: Alfredはどのような役割を果たしますか？** A: [Alfredワークフロー](../guides/alfred/index.md)で確認してください。Alfredは19人のAI専門家チームを調整するスーパーエージェントです。

### TDD関連

**Q: TDDのRED-GREEN-REFACTORとは何ですか？** A: [TDDガイド](../guides/tdd/index.md)で各段階を詳しく説明しています。

**Q: テストカバレッジはどのくらい必要ですか？** A: MoAI-ADKは**85%以上のテストカバレッジ**を推奨しています。

### TAGシステム

**Q: @TAGシステムはなぜ必要ですか？** A: [TAGシステム](../guides/specs/tags.md)を通じて、SPEC、TEST、CODE、DOCをすべて接続し、完全な追跡可能性を提供します。

______________________________________________________________________

## 一般的なエラーメッセージ

### "Command not found: /alfred:1-plan"

**原因**: Claude CodeがAlfredコマンドを認識しない

**解決策**:

```bash
# 1. Claude Code再起動
exit
claude

# 2. ディレクトリ確認
ls .claude/commands/

# 3. 設定リフレッシュ
/alfred:0-project
```

### "SPEC file not found"

**原因**: SPECファイルが正しい場所に作成されていない

**解決策**:

```bash
# プロジェクトステータス確認
moai-adk doctor

# .moai/ディレクトリ権限確認
ls -la .moai/

# 再初期化
rm -rf .moai
/alfred:0-project
```

### "Test coverage below 85%"

**原因**: テストカバレッジが不足

**解決策**:

```bash
# 現在のカバレッジ確認
pytest --cov=src tests/

# 不足しているテスト追加
# tests/test_*.pyにテストケース追加

# 再度実行
pytest --cov=src tests/
```

______________________________________________________________________

## システム診断

### 診断ツール実行

```bash
# 全体システムステータス確認
moai-adk doctor

# 詳細出力
moai-adk doctor --verbose
```

### 確認される項目

- Pythonバージョンおよび依存関係
- Git設定および権限
- .moai/ディレクトリ構造
- .claude/設定ファイル
- 必須ツールインストール有無

______________________________________________________________________

## 追加のヘルプ

### コミュニティ

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) - 質問してアイデアを共有
- [Issue Tracker](https://github.com/modu-ai/moai-adk/issues) - バグ報告

### ドキュメント

- [オンラインドキュメント](https://adk.mo.ai.kr) - 最新情報
- [ローカルドキュメント](../index.md) - オフライン参照

### フィードバック

```bash
# 問題報告（GitHub Issue自動生成）
/alfred:9-feedback
```

______________________________________________________________________

**役に立ちましたか？** [より多くの質問](https://github.com/modu-ai/moai-adk/discussions)を歓迎します！



