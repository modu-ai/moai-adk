---
title: プロジェクト管理ガイド
description: MoAI-ADKプロジェクトの初期化、設定、デプロイの完全ガイド
status: stable
---

# プロジェクト管理ガイド

MoAI-ADKプロジェクトのライフサイクル全体を管理する方法を学びます。初期化から設定、デプロイまでのすべての段階をカバーします。

## 🎯 プロジェクト管理の3つのステージ

### [1. プロジェクト初期化](init.md)
- `moai-adk init`コマンドで新規プロジェクトを作成
- プロジェクトテンプレートの選択
- 必要なファイル構造の自動生成

**主な生成項目**:
- `.moai/config.json` - プロジェクトメタデータ
- `.claude/` - Claude Code設定（agents, commands, skills, hooks）
- `pyproject.toml` - Pythonプロジェクト設定
- `pytest.ini` - テスト設定

### [2. 設定管理](config.md)
- `.moai/config.json`の詳細設定
- 言語とローカライゼーション設定
- 開発環境のカスタマイズ
- HookとAgentの設定

**コア設定項目**:
- プロジェクトメタデータ（名前、バージョン、説明）
- 言語とドメイン設定
- Gitワークフロー設定
- レポート生成ポリシー

### [3. デプロイ戦略](deploy.md)
- ローカル開発環境のセットアップ
- Dockerコンテナ化
- クラウドプラットフォームへのデプロイ（Vercel、Railway、AWS）
- CI/CDパイプラインの構築
- モニタリングとログ

## 📊 プロジェクト構造

```
my-awesome-project/
├── .moai/              # MoAI-ADKメタデータ
│   ├── config.json     # プロジェクト設定
│   ├── docs/           # 自動生成ドキュメント
│   └── reports/        # 分析とレポート
├── .claude/            # Claude Code設定
│   ├── agents/         # サブエージェントのカスタマイズ
│   ├── commands/       # スラッシュコマンド
│   ├── skills/         # プロジェクト固有のSkills
│   └── hooks/          # 自動化Hooks
├── src/                # ソースコード
├── tests/              # テストコード
├── docs/               # プロジェクトドキュメント
└── pyproject.toml      # Pythonプロジェクト設定
```

## 🔄 Alfredとの統合

プロジェクト管理はAlfred SuperAgentと完璧に統合されています:

- `/alfred:0-project` - プロジェクト設定の最適化
- `/alfred:1-plan` - 要件SPECの作成
- `/alfred:2-run` - TDD実装
- `/alfred:3-sync` - ドキュメント同期とデプロイ

[完全なAlfredワークフローガイド](../alfred/index.md)

## 📋 チェックリスト

プロジェクトをセットアップする際に確認すべき項目:

- [ ] プロジェクト初期化完了（`moai-adk init`）
- [ ] `.moai/config.json`のレビューとカスタマイズ
- [ ] Gitワークフロー設定の確認
- [ ] 開発環境のセットアップ完了
- [ ] CI/CDパイプラインの設定（オプション）
- [ ] デプロイ戦略の決定

## 🚀 次のステップ

- [プロジェクト初期化: init.md](init.md)
- [設定管理: config.md](config.md)
- [デプロイ戦略: deploy.md](deploy.md)
- [Alfred 0-project: 設定の最適化](../alfred/index.md)

---

**詳細情報**: プロジェクト管理はMoAI-ADKワークフローの基礎です。適切な設定から始めることで、開発生産性が大幅に向上します。
